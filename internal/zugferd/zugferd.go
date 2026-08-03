package zugferd

import (
	"bytes"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/speedata/einvoice"
)

// InvoiceData holds aggregated data to generate ZUGFeRD XML
type InvoiceData struct {
	Rechnungsnummer        string
	Rechnungsdatum         time.Time
	Faelligkeitsdatum      time.Time
	LeistungszeitraumStart time.Time
	LeistungszeitraumEnde  time.Time

	// Eigene Stammdaten
	VerkaeuferName      string
	VerkaeuferStrasse   string
	VerkaeuferPlz       string
	VerkaeuferOrt       string
	VerkaeuferLand      string
	VerkaeuferUstId     string
	VerkaeuferSteuerNum string
	VerkaeuferIban      string
	VerkaeuferBic       string

	// Kunde
	KundeName        string
	KundeStrasse     string
	KundePlz         string
	KundeOrt         string
	KundeLand        string
	KundeUstId       string
	KundeZahlungsart string

	// Positionen
	Lines []InvoiceLineData
}

type InvoiceLineData struct {
	PosNum      int
	Material    string
	Menge       float64
	Einheit     string
	Einzelpreis float64
	Netto       float64
	MwstSatz    float64
}

// Helper to map UI/Access units to UN/ECE codes
func MapUnitCode(einheit string) string {
	switch einheit {
	case "t":
		return "TNE" // Metric Ton
	case "m3":
		return "MTQ" // Cubic Meter
	case "Stk":
		return "H87" // Piece
	default:
		return "C62" // Unit / One
	}
}

// GenerateXML creates the ZUGFeRD 2.2 BASIC XML document using speedata/einvoice
func GenerateXML(inv *InvoiceData) ([]byte, error) {
	ein := einvoice.Invoice{
		InvoiceNumber:   inv.Rechnungsnummer,
		InvoiceTypeCode: einvoice.CodeDocument(380), // Commercial invoice
		GuidelineSpecifiedDocumentContextParameter: einvoice.SpecFacturXBasic,
		InvoiceDate:         inv.Rechnungsdatum,
		InvoiceCurrencyCode: "EUR",
		OccurrenceDateTime:  inv.LeistungszeitraumEnde,
		SchemaType:          einvoice.CII,
		Seller: einvoice.Party{
			Name:              inv.VerkaeuferName,
			VATaxRegistration: inv.VerkaeuferUstId,
			FCTaxRegistration: inv.VerkaeuferSteuerNum,
			PostalAddress: &einvoice.PostalAddress{
				Line1:        inv.VerkaeuferStrasse,
				City:         inv.VerkaeuferOrt,
				PostcodeCode: inv.VerkaeuferPlz,
				CountryID:    inv.VerkaeuferLand,
			},
		},
		Buyer: einvoice.Party{
			Name:              inv.KundeName,
			VATaxRegistration: inv.KundeUstId,
			PostalAddress: &einvoice.PostalAddress{
				Line1:        inv.KundeStrasse,
				City:         inv.KundeOrt,
				PostcodeCode: inv.KundePlz,
				CountryID:    inv.KundeLand,
			},
		},
		PaymentMeans: []einvoice.PaymentMeans{
			{
				TypeCode:                               30, // Credit transfer (Ueberweisung)
				PayeePartyCreditorFinancialAccountIBAN: inv.VerkaeuferIban,
			},
		},
		SpecifiedTradePaymentTerms: []einvoice.SpecifiedTradePaymentTerms{
			{
				Description: "Zahlbar innerhalb 14 Tagen ohne Abzuege.",
				DueDate:     inv.Faelligkeitsdatum,
			},
		},
		BillingSpecifiedPeriodStart: inv.LeistungszeitraumStart,
		BillingSpecifiedPeriodEnd:   inv.LeistungszeitraumEnde,
	}

	for _, line := range inv.Lines {
		item := einvoice.InvoiceLine{
			LineID:                   fmt.Sprintf("%d", line.PosNum),
			ItemName:                 line.Material,
			BilledQuantity:           decimal.NewFromFloat(line.Menge),
			BilledQuantityUnit:       MapUnitCode(line.Einheit),
			NetPrice:                 decimal.NewFromFloat(line.Einzelpreis),
			TaxRateApplicablePercent: decimal.NewFromFloat(line.MwstSatz),
			TaxTypeCode:              "VAT",
			TaxCategoryCode:          "S",
			Total:                    decimal.NewFromFloat(line.Netto),
		}
		ein.InvoiceLines = append(ein.InvoiceLines, item)
	}

	// Calculate taxes and totals
	ein.UpdateApplicableTradeTax(nil)
	ein.UpdateTotals()

	var buf bytes.Buffer
	if err := ein.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
