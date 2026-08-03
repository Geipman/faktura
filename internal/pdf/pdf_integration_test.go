package pdf

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Geipman/faktura/internal/zugferd"
)

func TestValidateZUGFeRDPDF(t *testing.T) {
	// Find project root to locate bin/mustang-cli
	// Since tests are run in internal/pdf, the project root is "../../"
	mustangPath := filepath.Join("..", "..", "bin", "mustang-cli")

	// If run from workspace root, adjust path
	if _, err := os.Stat("bin/mustang-cli"); err == nil {
		mustangPath = "bin/mustang-cli"
	}

	// Check if mustang-cli wrapper is available
	if _, err := os.Stat(mustangPath); os.IsNotExist(err) {
		t.Skip("mustang-cli not found, skipping integration validation test")
	}

	// 1. Prepare realistic dummy data
	inv := &zugferd.InvoiceData{
		Rechnungsnummer:        "RE-2026-9999",
		Rechnungsdatum:         time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC),
		Faelligkeitsdatum:      time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC),
		LeistungszeitraumStart: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		LeistungszeitraumEnde:  time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC),

		VerkaeuferName:      "Kompostwerk Test GmbH",
		VerkaeuferStrasse:   "Musterstraße 42",
		VerkaeuferPlz:       "12345",
		VerkaeuferOrt:       "Musterstadt",
		VerkaeuferLand:      "DE",
		VerkaeuferUstId:     "DE123456789",
		VerkaeuferSteuerNum: "123/456/7890",
		VerkaeuferIban:      "DE89370400440532013000",
		VerkaeuferBic:       "GENODEF1VBD",

		KundeName:    "AWS Musterkunde GmbH",
		KundeStrasse: "Kundenstraße 10",
		KundePlz:     "54321",
		KundeOrt:     "Kundenstadt",
		KundeLand:    "DE",
		KundeUstId:   "DE987654321",

		Lines: []zugferd.InvoiceLineData{
			{
				PosNum:      1,
				Material:    "Grünabfall",
				Menge:       12.5,
				Einheit:     "t",
				Einzelpreis: 20.0,
				Netto:       250.0,
				MwstSatz:    19.0,
			},
			{
				PosNum:      2,
				Material:    "Bodenmischung",
				Menge:       5.0,
				Einheit:     "m3",
				Einzelpreis: 15.0,
				Netto:       75.0,
				MwstSatz:    7.0,
			},
		},
	}

	// 2. Generate ZUGFeRD XML
	xmlBytes, err := zugferd.GenerateXML(inv)
	if err != nil {
		t.Fatalf("Failed to generate ZUGFeRD XML: %v", err)
	}

	// 3. Generate PDF/A-3
	tmpDir := filepath.Join("..", "..", "tmp")
	if _, statErr := os.Stat(tmpDir); os.IsNotExist(statErr) {
		tmpDir = "tmp" // fall back if run from other cwd
	}
	if mkdirErr := os.MkdirAll(tmpDir, 0755); mkdirErr != nil {
		t.Fatalf("Failed to create tmp dir: %v", mkdirErr)
	}

	pdfPath := filepath.Join(tmpDir, "integration_test_invoice.pdf")
	defer os.Remove(pdfPath)

	err = GenerateInvoicePDF(inv, xmlBytes, pdfPath)
	if err != nil {
		t.Fatalf("Failed to generate ZUGFeRD PDF: %v", err)
	}

	// 4. Run Mustang-CLI to validate PDF
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, mustangPath, "--action", "validate", "--source", pdfPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		t.Fatalf("mustang-cli failed to run: %v\nstderr: %s\nstdout: %s", err, stderr.String(), stdout.String())
	}

	outputXML := stdout.String()

	// 5. Assert the validation result
	if strings.Contains(outputXML, "<error>") || strings.Contains(outputXML, "<severity>error</severity>") {
		t.Errorf("Validation failed for ZUGFeRD PDF! Validation report:\n%s", outputXML)
	} else {
		t.Logf("Success! ZUGFeRD validation report is clean:\n%s", outputXML)
	}
}
