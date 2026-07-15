package zugferd

import (
	"testing"
	"time"
)

func TestGenerateXML(t *testing.T) {
	inv := &InvoiceData{
		Rechnungsnummer:        "RE-2026-0001",
		Rechnungsdatum:         time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
		Faelligkeitsdatum:      time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC),
		LeistungszeitraumStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		LeistungszeitraumEnde:  time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),

		VerkaeuferName:      "Kompostwerk Büttelborn GmbH",
		VerkaeuferStrasse:   "Auf der Hardt/An der B 42",
		VerkaeuferPlz:       "64572",
		VerkaeuferOrt:       "Büttelborn",
		VerkaeuferLand:      "DE",
		VerkaeuferUstId:     "DE5089000048161405",
		VerkaeuferSteuerNum: "12/345/67890",
		VerkaeuferIban:      "DE58508900000048161405",
		VerkaeuferBic:       "GENODEF1VBD",

		KundeName:    "AWS Abfall-Wirtschafts-Service GmbH",
		KundeStrasse: "Auf der Hardt/An der B 42",
		KundePlz:     "64572",
		KundeOrt:     "Büttelborn",
		KundeLand:    "DE",
		KundeUstId:   "DE987654321",

		Lines: []InvoiceLineData{
			{
				PosNum:      1,
				Material:    "Grünabfall (AS 20 02 01)",
				Menge:       9.04,
				Einheit:     "t",
				Einzelpreis: 19.00,
				Netto:       171.76,
				MwstSatz:    19.00,
			},
			{
				PosNum:      2,
				Material:    "Feinkompost 0/12",
				Menge:       40.32,
				Einheit:     "t",
				Einzelpreis: 8.00,
				Netto:       322.56,
				MwstSatz:    7.00,
			},
		},
	}

	xmlBytes, err := GenerateXML(inv)
	if err != nil {
		t.Fatalf("Failed to generate ZUGFeRD XML: %v", err)
	}

	if len(xmlBytes) == 0 {
		t.Fatal("Generated empty XML byte array")
	}

	// Verify it's a valid XML document by checking key elements
	xmlStr := string(xmlBytes)

	expectedElements := []string{
		"<rsm:CrossIndustryInvoice",
		"<ram:GuidelineSpecifiedDocumentContextParameter>",
		"<ram:ID>urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:basic</ram:ID>",
		"<ram:ID>RE-2026-0001</ram:ID>",
		"<ram:TypeCode>380</ram:TypeCode>",
		"<ram:Name>Kompostwerk Büttelborn GmbH</ram:Name>",
		"<ram:Name>AWS Abfall-Wirtschafts-Service GmbH</ram:Name>",
		"<ram:ID schemeID=\"VA\">DE5089000048161405</ram:ID>",
		"<ram:IBANID>DE58508900000048161405</ram:IBANID>",
		"<ram:LineTotalAmount>494.32</ram:LineTotalAmount>", // 171.76 + 322.56
	}

	for _, elem := range expectedElements {
		if !contains(xmlStr, elem) {
			t.Errorf("Expected XML to contain '%s', but it did not.\nXML Output:\n%s", elem, xmlStr)
		}
	}
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && func() bool {
		for i := 0; i <= len(str)-len(substr); i++ {
			if str[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()
}
