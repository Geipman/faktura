package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Geipman/faktura/internal/zugferd"
	einvoicezugferd "github.com/dotwavehq/go-einvoice/zugferd"
	"github.com/signintech/gopdf"
)

// pdfDrawer implements the ErrWriter pattern to simplify sequential PDF drawing.
type pdfDrawer struct {
	pdf *gopdf.GoPdf
	err error
}

func (d *pdfDrawer) SetFont(family string, style string, size int) {
	if d.err != nil {
		return
	}
	d.err = d.pdf.SetFont(family, style, size)
}

func (d *pdfDrawer) SetX(x float64) {
	if d.err != nil {
		return
	}
	d.pdf.SetX(x)
}

func (d *pdfDrawer) SetY(y float64) {
	if d.err != nil {
		return
	}
	d.pdf.SetY(y)
}

func (d *pdfDrawer) Cell(w *gopdf.Rect, txt string) {
	if d.err != nil {
		return
	}
	d.err = d.pdf.Cell(w, txt)
}

func (d *pdfDrawer) CellWithOption(w *gopdf.Rect, txt string, opt gopdf.CellOption) {
	if d.err != nil {
		return
	}
	d.err = d.pdf.CellWithOption(w, txt, opt)
}

func (d *pdfDrawer) Line(x1 float64, y1 float64, x2 float64, y2 float64) {
	if d.err != nil {
		return
	}
	d.pdf.Line(x1, y1, x2, y2)
}

func (d *pdfDrawer) SetTextColor(r uint8, g uint8, b uint8) {
	if d.err != nil {
		return
	}
	d.pdf.SetTextColor(r, g, b)
}

func (d *pdfDrawer) SetLineWidth(width float64) {
	if d.err != nil {
		return
	}
	d.pdf.SetLineWidth(width)
}

// GenerateInvoicePDF creates the visual PDF invoice and embeds the ZUGFeRD XML file.
func GenerateInvoicePDF(inv *zugferd.InvoiceData, xmlBytes []byte, outputPath string) error {
	// Create temporary directories if they don't exist
	tmpDir := "tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	tempPDFPath := filepath.Join(tmpDir, fmt.Sprintf("temp_%s.pdf", inv.Rechnungsnummer))

	// Create visual PDF layout
	pdfDoc := gopdf.GoPdf{}
	pdfDoc.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdfDoc.AddPage()

	// Load DejaVuSans fonts
	fontPathRegular := filepath.Join("assets", "fonts", "DejaVuSans.ttf")
	fontPathBold := filepath.Join("assets", "fonts", "DejaVuSans-Bold.ttf")

	// Fallback for tests running in subdirectories
	if _, err := os.Stat(fontPathRegular); os.IsNotExist(err) {
		fontPathRegular = filepath.Join("..", "..", "assets", "fonts", "DejaVuSans.ttf")
	}
	if _, err := os.Stat(fontPathBold); os.IsNotExist(err) {
		fontPathBold = filepath.Join("..", "..", "assets", "fonts", "DejaVuSans-Bold.ttf")
	}

	if err := pdfDoc.AddTTFFont("DejaVuSans", fontPathRegular); err != nil {
		return fmt.Errorf("failed to load regular font (tried %s): %w", fontPathRegular, err)
	}
	if err := pdfDoc.AddTTFFont("DejaVuSans-Bold", fontPathBold); err != nil {
		return fmt.Errorf("failed to load bold font (tried %s): %w", fontPathBold, err)
	}

	// Initialize wrapper drawer
	drawer := &pdfDrawer{pdf: &pdfDoc}

	// 1. Draw Header Info (right-aligned metadata)
	drawer.SetFont("DejaVuSans", "", 9)
	drawer.SetTextColor(71, 85, 105) // Slate gray color

	metaX := 380.0
	drawer.SetX(metaX)
	drawer.SetY(70)
	drawer.Cell(nil, fmt.Sprintf("Datum:                  %s", inv.Rechnungsdatum.Format("02.01.2006")))
	drawer.SetX(metaX)
	drawer.SetY(85)
	drawer.Cell(nil, fmt.Sprintf("Rechnungs-Nr:    %s", inv.Rechnungsnummer))

	// 2. Draw Recipient Address (left-aligned)
	drawer.SetTextColor(15, 23, 42) // Dark slate
	drawer.SetFont("DejaVuSans-Bold", "", 10)
	drawer.SetX(50)
	drawer.SetY(130)
	drawer.Cell(nil, inv.KundeName)

	drawer.SetFont("DejaVuSans", "", 10)
	if inv.KundeStrasse != "" {
		drawer.SetX(50)
		drawer.SetY(145)
		drawer.Cell(nil, inv.KundeStrasse)
	}
	drawer.SetX(50)
	drawer.SetY(160)
	drawer.Cell(nil, fmt.Sprintf("%s %s", inv.KundePlz, inv.KundeOrt))

	// 3. Title "RECHNUNG" (centered)
	drawer.SetFont("DejaVuSans-Bold", "", 14)
	drawer.SetX(0)
	drawer.SetY(220)
	// Draw centered text manually (Page A4 width is 595 pt)
	drawer.CellWithOption(&gopdf.Rect{W: 595, H: 20}, "R E C H N U N G", gopdf.CellOption{Align: gopdf.Center})

	// 4. Intro text
	drawer.SetFont("DejaVuSans", "", 9)
	drawer.SetX(50)
	drawer.SetY(260)
	drawer.Cell(nil, "Sehr geehrte Damen und Herren,")

	drawer.SetX(50)
	drawer.SetY(280)
	drawer.Cell(nil, "für die Annahme pflanzlicher Abfälle zur Verwertung bzw. den Verkauf von Bodensubstraten")
	drawer.SetX(50)
	drawer.SetY(295)
	drawer.Cell(nil, "berechnen wir Ihnen nachfolgende Positionen gemäß der als Anlage beigefügten Wiegebelege.")

	// 5. Billing Period
	drawer.SetFont("DejaVuSans-Bold", "", 9)
	drawer.SetX(50)
	drawer.SetY(330)
	drawer.Cell(nil, fmt.Sprintf("Rechnungszeitraum:  %s %d", getGermanMonth(inv.LeistungszeitraumStart.Month()), inv.LeistungszeitraumStart.Year()))

	// 6. Positions Table Header
	drawer.SetFont("DejaVuSans-Bold", "", 9)
	drawer.SetY(360)

	drawer.SetX(50)
	drawer.Cell(nil, "Material")

	drawer.SetX(280)
	drawer.Cell(nil, "Menge")

	drawer.SetX(350)
	drawer.Cell(nil, "Preis pro Einh.")

	drawer.SetX(450)
	drawer.Cell(nil, "USt-Satz")

	drawer.SetX(510)
	drawer.Cell(nil, "Netto")

	// Table divider line
	drawer.SetLineWidth(0.5)
	drawer.Line(50, 375, 545, 375)

	// Draw rows
	y := 390.0
	var totalNet float64
	var totalTaxes float64

	// Track tax aggregation
	vatGroups := make(map[float64]float64) // rate -> net sum

	drawer.SetFont("DejaVuSans", "", 9)
	for _, line := range inv.Lines {
		drawer.SetX(50)
		drawer.SetY(y)
		drawer.Cell(nil, line.Material)

		drawer.SetX(280)
		drawer.Cell(nil, fmt.Sprintf("%.2f %s", line.Menge, line.Einheit))

		drawer.SetX(350)
		drawer.Cell(nil, fmt.Sprintf("%.2f €", line.Einzelpreis))

		drawer.SetX(450)
		drawer.Cell(nil, fmt.Sprintf("%.0f%%", line.MwstSatz))

		drawer.SetX(510)
		drawer.Cell(nil, fmt.Sprintf("%.2f €", line.Netto))

		totalNet += line.Netto
		vatGroups[line.MwstSatz] += line.Netto

		y += 20.0
	}

	// Divider line after items
	drawer.Line(50, y, 545, y)
	y += 20.0

	// 7. Taxes Matrix
	drawer.SetFont("DejaVuSans-Bold", "", 8)
	drawer.SetY(y)
	drawer.SetX(290)
	drawer.Cell(nil, "Netto (€)")
	drawer.SetX(360)
	drawer.Cell(nil, "MwSt. (%)")
	drawer.SetX(430)
	drawer.Cell(nil, "MwSt. (€)")
	drawer.SetX(500)
	drawer.Cell(nil, "Brutto (€)")

	drawer.SetFont("DejaVuSans", "", 8)

	taxRates := []float64{7.0, 19.0}
	for _, rate := range taxRates {
		y += 15.0
		netVal := vatGroups[rate]
		taxVal := netVal * (rate / 100.0)

		// Round to 2 decimals
		netVal = float64(int(netVal*100+0.5)) / 100.0
		taxVal = float64(int(taxVal*100+0.5)) / 100.0
		grossVal := netVal + taxVal
		totalTaxes += taxVal

		drawer.SetY(y)
		drawer.SetX(290)
		drawer.Cell(nil, fmt.Sprintf("%.2f €", netVal))

		drawer.SetX(360)
		drawer.Cell(nil, fmt.Sprintf("%.0f %%", rate))

		drawer.SetX(430)
		drawer.Cell(nil, fmt.Sprintf("%.2f €", taxVal))

		drawer.SetX(500)
		drawer.Cell(nil, fmt.Sprintf("%.2f €", grossVal))
	}

	y += 20.0
	drawer.Line(50, y, 545, y)
	y += 10.0

	// 8. Rechnungsbetrag Total
	totalNet = float64(int(totalNet*100+0.5)) / 100.0
	totalTaxes = float64(int(totalTaxes*100+0.5)) / 100.0
	grandTotal := totalNet + totalTaxes

	drawer.SetFont("DejaVuSans-Bold", "", 10)
	drawer.SetY(y)
	drawer.SetX(350)
	drawer.Cell(nil, "Rechnungsbetrag")
	drawer.SetX(480)
	drawer.Cell(nil, fmt.Sprintf("%.2f €", grandTotal))

	y += 15.0
	drawer.Line(50, y, 545, y)
	y += 2.0
	drawer.Line(50, y, 545, y) // double lines

	// 9. Payment footer (fixed bottom)
	drawer.SetFont("DejaVuSans", "", 9)
	drawer.SetX(50)
	drawer.SetY(730)
	drawer.Cell(nil, "Zahlbar innerhalb 14 Tagen ohne Abzüge.")

	drawer.SetX(50)
	drawer.SetY(745)
	drawer.Cell(nil, fmt.Sprintf("IBAN: %s  BIC: %s", inv.VerkaeuferIban, inv.VerkaeuferBic))

	// Check if any drawing error occurred
	if drawer.err != nil {
		return fmt.Errorf("error during PDF drawing: %w", drawer.err)
	}

	// Write visual PDF to temporary file
	if err := pdfDoc.WritePdf(tempPDFPath); err != nil {
		return fmt.Errorf("failed to write temporary PDF: %w", err)
	}
	defer os.Remove(tempPDFPath)

	// Embed the XML into the PDF as ZUGFeRD PDF/A-3
	err := einvoicezugferd.EmbedXML(tempPDFPath, xmlBytes, outputPath)
	if err != nil {
		return fmt.Errorf("failed to embed ZUGFeRD XML into PDF: %w", err)
	}

	return nil
}

// getGermanMonth converts Month to German name
func getGermanMonth(m time.Month) string {
	months := map[time.Month]string{
		time.January:   "Januar",
		time.February:  "Februar",
		time.March:     "März",
		time.April:     "April",
		time.May:       "Mai",
		time.June:      "Juni",
		time.July:      "Juli",
		time.August:    "August",
		time.September: "September",
		time.October:   "Oktober",
		time.November:  "November",
		time.December:  "Dezember",
	}
	return months[m]
}
