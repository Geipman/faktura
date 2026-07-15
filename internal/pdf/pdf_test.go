package pdf

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Geipman/faktura/internal/zugferd"
)

func TestGenerateInvoicePDF(t *testing.T) {
	// Prepare dummy data
	inv := &zugferd.InvoiceData{
		Rechnungsnummer:        "RE-TEST-0001",
		Rechnungsdatum:         time.Now(),
		Faelligkeitsdatum:      time.Now().Add(14 * 24 * time.Hour),
		LeistungszeitraumStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		LeistungszeitraumEnde:  time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),

		VerkaeuferName: "Test Kompostwerk GmbH",
		VerkaeuferIban: "DE58508900000048161405",
		VerkaeuferBic:  "GENODEF1VBD",

		KundeName: "Test Kunde GmbH",

		Lines: []zugferd.InvoiceLineData{
			{
				PosNum:      1,
				Material:    "Grünabfall (AS 20 02 01)",
				Menge:       10.0,
				Einheit:     "t",
				Einzelpreis: 19.0,
				Netto:       190.0,
				MwstSatz:    19.0,
			},
		},
	}

	xmlBytes := []byte("<test>ZUGFeRD XML stub</test>")

	tmpDir := "tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	outputPath := filepath.Join(tmpDir, "test_output.pdf")
	defer os.Remove(outputPath)

	// Execute generator
	err := GenerateInvoicePDF(inv, xmlBytes, outputPath)
	if err != nil {
		t.Fatalf("Failed to generate PDF/A-3 invoice: %v", err)
	}

	// Verify file is created
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Output file does not exist: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Output PDF file is empty")
	}
}
