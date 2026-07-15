package zugferd

import (
	"encoding/xml"
	"fmt"
	"time"
)

// Amount formats float64 to exactly 2 decimal places in XML
type Amount float64

func (a Amount) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(fmt.Sprintf("%.2f", a), start)
}

// Quantity formats float64 to up to 4 decimal places in XML
type Quantity float64

func (q Quantity) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(fmt.Sprintf("%.4f", q), start)
}

// DateTime represents a nested ZUGFeRD date structure
type DateTime struct {
	DateTimeString DateTimeString `xml:"udt:DateTimeString"`
}

type DateTimeString struct {
	Value  string `xml:",chardata"`
	Format string `xml:"format,attr"`
}

// ZUGFeRD 2.2 BASIC XML Structures
type CrossIndustryInvoice struct {
	XMLName                     xml.Name                    `xml:"rsm:CrossIndustryInvoice"`
	XmlnsRsm                    string                      `xml:"xmlns:rsm,attr"`
	XmlnsRam                    string                      `xml:"xmlns:ram,attr"`
	XmlnsQdt                    string                      `xml:"xmlns:qdt,attr"`
	XmlnsUdt                    string                      `xml:"xmlns:udt,attr"`
	ExchangedDocumentContext    ExchangedDocumentContext    `xml:"rsm:ExchangedDocumentContext"`
	ExchangedDocument           ExchangedDocument           `xml:"rsm:ExchangedDocument"`
	SupplyChainTradeTransaction SupplyChainTradeTransaction `xml:"rsm:SupplyChainTradeTransaction"`
}

type ExchangedDocumentContext struct {
	GuidelineSpecifiedDocumentContextParameter GuidelineSpecifiedDocumentContextParameter `xml:"ram:GuidelineSpecifiedDocumentContextParameter"`
}

type GuidelineSpecifiedDocumentContextParameter struct {
	ID string `xml:"ram:ID"`
}

type ExchangedDocument struct {
	ID            string   `xml:"ram:ID"`
	TypeCode      string   `xml:"ram:TypeCode"`
	IssueDateTime DateTime `xml:"ram:IssueDateTime"`
}

type SupplyChainTradeTransaction struct {
	IncludedSupplyChainTradeLineItem []IncludedSupplyChainTradeLineItem `xml:"ram:IncludedSupplyChainTradeLineItem"`
	ApplicableHeaderTradeAgreement   ApplicableHeaderTradeAgreement     `xml:"ram:ApplicableHeaderTradeAgreement"`
	ApplicableHeaderTradeDelivery    ApplicableHeaderTradeDelivery      `xml:"ram:ApplicableHeaderTradeDelivery"`
	ApplicableHeaderTradeSettlement  ApplicableHeaderTradeSettlement    `xml:"ram:ApplicableHeaderTradeSettlement"`
}

type IncludedSupplyChainTradeLineItem struct {
	AssociatedDocumentLineDocument AssociatedDocumentLineDocument `xml:"ram:AssociatedDocumentLineDocument"`
	SpecifiedTradeProduct          SpecifiedTradeProduct          `xml:"ram:SpecifiedTradeProduct"`
	SpecifiedLineTradeAgreement    SpecifiedLineTradeAgreement    `xml:"ram:SpecifiedLineTradeAgreement"`
	SpecifiedLineTradeDelivery     SpecifiedLineTradeDelivery     `xml:"ram:SpecifiedLineTradeDelivery"`
	SpecifiedLineTradeSettlement   SpecifiedLineTradeSettlement   `xml:"ram:SpecifiedLineTradeSettlement"`
}

type AssociatedDocumentLineDocument struct {
	LineID string `xml:"ram:LineID"`
}

type SpecifiedTradeProduct struct {
	Name string `xml:"ram:Name"`
}

type SpecifiedLineTradeAgreement struct {
	NetPriceProductTradePrice NetPriceProductTradePrice `xml:"ram:NetPriceProductTradePrice"`
}

type NetPriceProductTradePrice struct {
	ChargeAmount Amount `xml:"ram:ChargeAmount"`
}

type SpecifiedLineTradeDelivery struct {
	BilledQuantity BilledQuantity `xml:"ram:BilledQuantity"`
}

type BilledQuantity struct {
	Value    Quantity `xml:",chardata"`
	UnitCode string   `xml:"unitCode,attr"`
}

type SpecifiedLineTradeSettlement struct {
	ApplicableTradeTax                            LineApplicableTradeTax                        `xml:"ram:ApplicableTradeTax"`
	SpecifiedTradeSettlementLineMonetarySummation SpecifiedTradeSettlementLineMonetarySummation `xml:"ram:SpecifiedTradeSettlementLineMonetarySummation"`
}

type SpecifiedTradeSettlementLineMonetarySummation struct {
	LineTotalAmount Amount `xml:"ram:LineTotalAmount"`
}

type ApplicableHeaderTradeAgreement struct {
	SellerTradeParty TradeParty `xml:"ram:SellerTradeParty"`
	BuyerTradeParty  TradeParty `xml:"ram:BuyerTradeParty"`
}

type TradeParty struct {
	Name                     string                     `xml:"ram:Name"`
	PostalTradeAddress       PostalTradeAddress         `xml:"ram:PostalTradeAddress"`
	SpecifiedTaxRegistration []SpecifiedTaxRegistration `xml:"ram:SpecifiedTaxRegistration,omitempty"`
}

type PostalTradeAddress struct {
	PostcodeCode string `xml:"ram:PostcodeCode"`
	LineOne      string `xml:"ram:LineOne,omitempty"`
	LineTwo      string `xml:"ram:LineTwo,omitempty"`
	CityName     string `xml:"ram:CityName"`
	CountryID    string `xml:"ram:CountryID"`
}

type SpecifiedTaxRegistration struct {
	ID TaxRegistrationID `xml:"ram:ID"`
}

type TaxRegistrationID struct {
	Value    string `xml:",chardata"`
	SchemeID string `xml:"schemeID,attr"`
}

type ApplicableHeaderTradeDelivery struct {
	ActualDeliverySupplyChainEvent ActualDeliverySupplyChainEvent `xml:"ram:ActualDeliverySupplyChainEvent"`
}

type ActualDeliverySupplyChainEvent struct {
	OccurrenceDateTime DateTime `xml:"ram:OccurrenceDateTime"`
}

type ApplicableHeaderTradeSettlement struct {
	PaymentReference                                string                                          `xml:"ram:PaymentReference,omitempty"`
	InvoiceCurrencyCode                             string                                          `xml:"ram:InvoiceCurrencyCode"`
	SpecifiedTradeSettlementPaymentMeans            *SpecifiedTradeSettlementPaymentMeans           `xml:"ram:SpecifiedTradeSettlementPaymentMeans,omitempty"`
	ApplicableTradeTax                              []ApplicableTradeTax                            `xml:"ram:ApplicableTradeTax"`
	SpecifiedTradePaymentTerms                      SpecifiedTradePaymentTerms                      `xml:"ram:SpecifiedTradePaymentTerms"`
	SpecifiedTradeSettlementHeaderMonetarySummation SpecifiedTradeSettlementHeaderMonetarySummation `xml:"ram:SpecifiedTradeSettlementHeaderMonetarySummation"`
}

type SpecifiedTradeSettlementPaymentMeans struct {
	TypeCode                                   string                                      `xml:"ram:TypeCode"`
	PayeePartyCreditorFinancialAccount         PayeePartyCreditorFinancialAccount          `xml:"ram:PayeePartyCreditorFinancialAccount"`
	PayeeSpecifiedCreditorFinancialInstitution *PayeeSpecifiedCreditorFinancialInstitution `xml:"ram:PayeeSpecifiedCreditorFinancialInstitution,omitempty"`
}

type PayeePartyCreditorFinancialAccount struct {
	IBANID string `xml:"ram:IBANID"`
}

type PayeeSpecifiedCreditorFinancialInstitution struct {
	BICID string `xml:"ram:BICID"`
}

type ApplicableTradeTax struct {
	CalculatedAmount      Amount `xml:"ram:CalculatedAmount"`
	TypeCode              string `xml:"ram:TypeCode"`
	BasisAmount           Amount `xml:"ram:BasisAmount"`
	CategoryCode          string `xml:"ram:CategoryCode"`
	RateApplicablePercent Amount `xml:"ram:RateApplicablePercent"`
}

type LineApplicableTradeTax struct {
	TypeCode              string `xml:"ram:TypeCode"`
	CategoryCode          string `xml:"ram:CategoryCode"`
	RateApplicablePercent Amount `xml:"ram:RateApplicablePercent"`
}

type SpecifiedTradePaymentTerms struct {
	Description     string    `xml:"ram:Description,omitempty"`
	DueDateDateTime *DateTime `xml:"ram:DueDateDateTime,omitempty"`
}

type SpecifiedTradeSettlementHeaderMonetarySummation struct {
	LineTotalAmount      Amount         `xml:"ram:LineTotalAmount"`
	ChargeTotalAmount    Amount         `xml:"ram:ChargeTotalAmount"`
	AllowanceTotalAmount Amount         `xml:"ram:AllowanceTotalAmount"`
	TaxBasisTotalAmount  Amount         `xml:"ram:TaxBasisTotalAmount"`
	TaxTotalAmount       TaxTotalAmount `xml:"ram:TaxTotalAmount"`
	GrandTotalAmount     Amount         `xml:"ram:GrandTotalAmount"`
	DuePayableAmount     Amount         `xml:"ram:DuePayableAmount"`
}

type TaxTotalAmount struct {
	Value      Amount `xml:",chardata"`
	CurrencyID string `xml:"currencyID,attr"`
}

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

// GenerateXML creates the ZUGFeRD 2.2 BASIC XML document
func GenerateXML(inv *InvoiceData) ([]byte, error) {
	// Group items to calculate VAT aggregates
	vatGroups := make(map[float64]*ApplicableTradeTax)
	for _, line := range inv.Lines {
		tax, ok := vatGroups[line.MwstSatz]
		if !ok {
			tax = &ApplicableTradeTax{
				TypeCode:              "VAT",
				CategoryCode:          "S",
				RateApplicablePercent: Amount(line.MwstSatz),
			}
			vatGroups[line.MwstSatz] = tax
		}
		tax.BasisAmount += Amount(line.Netto)
		tax.CalculatedAmount += Amount(line.Netto * (line.MwstSatz / 100.0))
	}

	var taxes []ApplicableTradeTax
	var lineTotal float64
	var taxTotal float64

	for _, tax := range vatGroups {
		// Round to 2 decimals
		tax.BasisAmount = Amount(float64(int(float64(tax.BasisAmount)*100+0.5)) / 100.0)
		tax.CalculatedAmount = Amount(float64(int(float64(tax.CalculatedAmount)*100+0.5)) / 100.0)
		taxes = append(taxes, *tax)

		lineTotal += float64(tax.BasisAmount)
		taxTotal += float64(tax.CalculatedAmount)
	}

	lineTotal = float64(int(lineTotal*100+0.5)) / 100.0
	taxTotal = float64(int(taxTotal*100+0.5)) / 100.0
	grandTotal := lineTotal + taxTotal

	// Construct XML objects
	rsm := CrossIndustryInvoice{
		XmlnsRsm: "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100",
		XmlnsRam: "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100",
		XmlnsQdt: "urn:un:unece:uncefact:data:standard:QualifiedDataType:100",
		XmlnsUdt: "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100",
		ExchangedDocumentContext: ExchangedDocumentContext{
			GuidelineSpecifiedDocumentContextParameter: GuidelineSpecifiedDocumentContextParameter{
				ID: "urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:basic",
			},
		},
		ExchangedDocument: ExchangedDocument{
			ID:       inv.Rechnungsnummer,
			TypeCode: "380", // Commercial invoice
			IssueDateTime: DateTime{
				DateTimeString: DateTimeString{
					Value:  inv.Rechnungsdatum.Format("20060102"),
					Format: "102",
				},
			},
		},
		SupplyChainTradeTransaction: SupplyChainTradeTransaction{
			ApplicableHeaderTradeAgreement: ApplicableHeaderTradeAgreement{
				SellerTradeParty: TradeParty{
					Name: inv.VerkaeuferName,
					PostalTradeAddress: PostalTradeAddress{
						PostcodeCode: inv.VerkaeuferPlz,
						LineOne:      inv.VerkaeuferStrasse,
						CityName:     inv.VerkaeuferOrt,
						CountryID:    inv.VerkaeuferLand,
					},
					SpecifiedTaxRegistration: []SpecifiedTaxRegistration{
						{ID: TaxRegistrationID{Value: inv.VerkaeuferUstId, SchemeID: "VA"}},
						{ID: TaxRegistrationID{Value: inv.VerkaeuferSteuerNum, SchemeID: "FC"}},
					},
				},
				BuyerTradeParty: TradeParty{
					Name: inv.KundeName,
					PostalTradeAddress: PostalTradeAddress{
						PostcodeCode: inv.KundePlz,
						LineOne:      inv.KundeStrasse,
						CityName:     inv.KundeOrt,
						CountryID:    inv.KundeLand,
					},
				},
			},
			ApplicableHeaderTradeDelivery: ApplicableHeaderTradeDelivery{
				ActualDeliverySupplyChainEvent: ActualDeliverySupplyChainEvent{
					OccurrenceDateTime: DateTime{
						DateTimeString: DateTimeString{
							Value:  inv.LeistungszeitraumEnde.Format("20060102"),
							Format: "102",
						},
					},
				},
			},
			ApplicableHeaderTradeSettlement: ApplicableHeaderTradeSettlement{
				PaymentReference:    inv.Rechnungsnummer,
				InvoiceCurrencyCode: "EUR",
				SpecifiedTradeSettlementPaymentMeans: &SpecifiedTradeSettlementPaymentMeans{
					TypeCode: "30", // Credit transfer (Ueberweisung)
					PayeePartyCreditorFinancialAccount: PayeePartyCreditorFinancialAccount{
						IBANID: inv.VerkaeuferIban,
					},
					// Omit BIC to comply with ZUGFeRD BASIC profile schema
					PayeeSpecifiedCreditorFinancialInstitution: nil,
				},
				ApplicableTradeTax: taxes,
				SpecifiedTradePaymentTerms: SpecifiedTradePaymentTerms{
					Description: "Zahlbar innerhalb 14 Tagen ohne Abzuege.",
					DueDateDateTime: &DateTime{
						DateTimeString: DateTimeString{
							Value:  inv.Faelligkeitsdatum.Format("20060102"),
							Format: "102",
						},
					},
				},
				SpecifiedTradeSettlementHeaderMonetarySummation: SpecifiedTradeSettlementHeaderMonetarySummation{
					LineTotalAmount:      Amount(lineTotal),
					ChargeTotalAmount:    Amount(0.0),
					AllowanceTotalAmount: Amount(0.0),
					TaxBasisTotalAmount:  Amount(lineTotal),
					TaxTotalAmount: TaxTotalAmount{
						Value:      Amount(taxTotal),
						CurrencyID: "EUR",
					},
					GrandTotalAmount: Amount(grandTotal),
					DuePayableAmount: Amount(grandTotal),
				},
			},
		},
	}

	// Add buyer tax registration if present
	if inv.KundeUstId != "" {
		rsm.SupplyChainTradeTransaction.ApplicableHeaderTradeAgreement.BuyerTradeParty.SpecifiedTaxRegistration = []SpecifiedTaxRegistration{
			{ID: TaxRegistrationID{Value: inv.KundeUstId, SchemeID: "VA"}},
		}
	}

	// Add lines
	for _, line := range inv.Lines {
		item := IncludedSupplyChainTradeLineItem{
			AssociatedDocumentLineDocument: AssociatedDocumentLineDocument{
				LineID: fmt.Sprintf("%d", line.PosNum),
			},
			SpecifiedTradeProduct: SpecifiedTradeProduct{
				Name: line.Material,
			},
			SpecifiedLineTradeAgreement: SpecifiedLineTradeAgreement{
				NetPriceProductTradePrice: NetPriceProductTradePrice{
					ChargeAmount: Amount(line.Einzelpreis),
				},
			},
			SpecifiedLineTradeDelivery: SpecifiedLineTradeDelivery{
				BilledQuantity: BilledQuantity{
					Value:    Quantity(line.Menge),
					UnitCode: MapUnitCode(line.Einheit),
				},
			},
			SpecifiedLineTradeSettlement: SpecifiedLineTradeSettlement{
				ApplicableTradeTax: LineApplicableTradeTax{
					TypeCode:              "VAT",
					CategoryCode:          "S",
					RateApplicablePercent: Amount(line.MwstSatz),
				},
				SpecifiedTradeSettlementLineMonetarySummation: SpecifiedTradeSettlementLineMonetarySummation{
					LineTotalAmount: Amount(line.Netto),
				},
			},
		}
		rsm.SupplyChainTradeTransaction.IncludedSupplyChainTradeLineItem = append(rsm.SupplyChainTradeTransaction.IncludedSupplyChainTradeLineItem, item)
	}

	// Marshal XML with header declaration
	xmlBytes, err := xml.MarshalIndent(rsm, "", "  ")
	if err != nil {
		return nil, err
	}

	header := []byte(xml.Header)
	return append(header, xmlBytes...), nil
}
