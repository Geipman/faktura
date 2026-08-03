package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Geipman/faktura/internal/pdf"
	"github.com/Geipman/faktura/internal/templates"
	"github.com/Geipman/faktura/internal/zugferd"
)

// Server represents our HTTP server and dependencies.
type Server struct {
	db   *sql.DB
	mux  *http.ServeMux
	addr string
}

// NewServer creates a new Server instance.
func NewServer(addr string, db *sql.DB) *Server {
	s := &Server{
		db:   db,
		mux:  http.NewServeMux(),
		addr: addr,
	}
	s.routes()
	return s
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	log.Printf("Starting server on %s...", s.addr)
	return http.ListenAndServe(s.addr, s.mux)
}

// routes configures the HTTP routes.
func (s *Server) routes() {
	// Serve static files (Tailwind CSS, static PDFs)
	fs := http.FileServer(http.Dir("internal/server/static"))
	s.mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("GET /rechnungen", s.handleInvoicesList)
	s.mux.HandleFunc("GET /rechnungen/neu", s.handleInvoiceNewForm)
	s.mux.HandleFunc("POST /rechnungen/neu", s.handleInvoiceCreate)
	s.mux.HandleFunc("GET /rechnungen/ansicht/{id}", s.handleInvoiceView)
	s.mux.HandleFunc("GET /rechnungen/xml/{id}", s.handleInvoiceXML)
	s.mux.HandleFunc("GET /rechnungen/pdf/{id}", s.handleInvoicePDF)
	s.mux.HandleFunc("POST /rechnungen/stornieren/{id}", s.handleInvoiceCancel)

	// Wiegezettel CRUD Routes
	s.mux.HandleFunc("GET /wiegezettel", s.handleWiegezettelList)
	s.mux.HandleFunc("GET /wiegezettel/tabelle", s.handleWiegezettelTable)
	s.mux.HandleFunc("GET /wiegezettel/neu", s.handleWiegezettelNewForm)
	s.mux.HandleFunc("POST /wiegezettel/neu", s.handleWiegezettelCreate)
	s.mux.HandleFunc("GET /wiegezettel/bearbeiten/{id}", s.handleWiegezettelEditForm)
	s.mux.HandleFunc("POST /wiegezettel/bearbeiten/{id}", s.handleWiegezettelUpdate)
	s.mux.HandleFunc("POST /wiegezettel/loeschen/{id}", s.handleWiegezettelDelete)
}

// handleIndex redirects to dashboard.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var stats templates.DashboardStats
	// 1. Offene Wiegezettel
	err := s.db.QueryRow("SELECT COUNT(*) FROM wiegezettel WHERE rechnungsposition_id IS NULL").Scan(&stats.OpenSlips)
	if err != nil {
		log.Printf("Error querying open slips: %v", err)
	}

	// 2. Rechnungen (Monat)
	err = s.db.QueryRow("SELECT COUNT(*) FROM rechnungen").Scan(&stats.TotalInvoices)
	if err != nil {
		log.Printf("Error querying total invoices: %v", err)
	}

	// 3. Aktive Kunden
	err = s.db.QueryRow("SELECT COUNT(*) FROM kunden").Scan(&stats.ActiveKunden)
	if err != nil {
		log.Printf("Error querying active customers: %v", err)
	}

	// 4. Last 5 Wiegezettel
	rows, err := s.db.Query(`
		SELECT w.wiegezettel_id, w.kundennummer, k.kundenname, w.datum, w.gewicht, 
		       w.material_id, m.materialname, m.einheit, w.anlieferungsort, w.anlieferer, 
		       COALESCE(w.referenz, '')
		FROM wiegezettel w
		JOIN kunden k ON w.kundennummer = k.kundennummer
		JOIN materialarten m ON w.material_id = m.material_id
		ORDER BY w.datum DESC, w.wiegezettel_id DESC
		LIMIT 5
	`)
	var recent []templates.WiegezettelRow
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var row templates.WiegezettelRow
			var dateStr string
			scanErr := rows.Scan(
				&row.ID, &row.Kundennummer, &row.Kundenname, &dateStr, &row.Gewicht,
				&row.MaterialID, &row.Materialname, &row.Einheit, &row.Anlieferungsort, &row.Anlieferer,
				&row.Referenz,
			)
			if scanErr == nil {
				var parsedTime time.Time
				var parseErr error
				layouts := []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05Z", "2006-01-02T15:04:05", "2006-01-02"}
				for _, layout := range layouts {
					if parsedTime, parseErr = time.Parse(layout, dateStr[:Min(len(dateStr), len(layout))]); parseErr == nil {
						break
					}
				}
				if parseErr == nil {
					row.Datum = parsedTime
				} else {
					row.Datum = time.Now()
				}
				recent = append(recent, row)
			}
		}
	} else {
		log.Printf("Error querying recent slips: %v", err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	indexComp := templates.Index(stats, recent)
	layoutComp := templates.Layout("Dashboard", indexComp)

	if err := layoutComp.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering index page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleInvoicesList lists all created invoices.
func (s *Server) handleInvoicesList(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(`
		SELECT r.rechnung_id, r.rechnungsnummer, r.kundennummer, k.kundenname, 
		       r.rechnungsdatum, r.leistungszeitraum_start, r.leistungszeitraum_ende, 
		       r.netto_summe, r.brutto_summe, r.zahlungsstatus
		FROM rechnungen r
		JOIN kunden k ON r.kundennummer = k.kundennummer
		ORDER BY r.rechnungsnummer DESC
	`)
	if err != nil {
		log.Printf("DB error querying invoices: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var invoices []templates.InvoiceHeader
	for rows.Next() {
		var inv templates.InvoiceHeader
		var dateStr, startStr, endStr string
		scanErr := rows.Scan(
			&inv.RechnungId, &inv.Rechnungsnummer, &inv.Kundennummer, &inv.Kundenname,
			&dateStr, &startStr, &endStr, &inv.NettoSumme, &inv.BruttoSumme, &inv.Zahlungsstatus,
		)
		if scanErr != nil {
			log.Printf("Scan error: %v", scanErr)
			continue
		}

		if t, parseErr := time.Parse("2006-01-02", dateStr[:10]); parseErr == nil {
			inv.Rechnungsdatum = t
		}
		if t, parseErr := time.Parse("2006-01-02", startStr[:10]); parseErr == nil {
			inv.LeistungszeitraumStart = t
		}
		if t, parseErr := time.Parse("2006-01-02", endStr[:10]); parseErr == nil {
			inv.LeistungszeitraumEnde = t
		}

		invoices = append(invoices, inv)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	comp := templates.Invoices(invoices)
	layout := templates.Layout("Rechnungen", comp)
	if err := layout.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering invoices list: %v", err)
	}
}

// handleInvoiceNewForm displays the create invoice form.
func (s *Server) handleInvoiceNewForm(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("SELECT kundennummer, kundenname FROM kunden ORDER BY kundenname ASC")
	if err != nil {
		log.Printf("DB error: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var kunden []templates.KundeLookup
	for rows.Next() {
		var k templates.KundeLookup
		if scanErr := rows.Scan(&k.Kundennummer, &k.Kundenname); scanErr == nil {
			kunden = append(kunden, k)
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	comp := templates.InvoiceNew(kunden)
	layout := templates.Layout("Neue Rechnung", comp)
	if err := layout.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering new invoice form: %v", err)
	}
}

// handleInvoiceCreate executes the pricing and aggregation and generates the ZUGFeRD invoice.
func (s *Server) handleInvoiceCreate(w http.ResponseWriter, r *http.Request) {
	kNummerStr := r.FormValue("kundennummer")
	startStr := r.FormValue("start_datum")
	endeStr := r.FormValue("ende_datum")

	kNummer, err := strconv.ParseInt(kNummerStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}
	ende, err := time.Parse("2006-01-02", endeStr)
	if err != nil {
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}

	// Add time to end date to make it inclusive (end of day)
	endeInclusive := time.Date(ende.Year(), ende.Month(), ende.Day(), 23, 59, 59, 0, ende.Location())

	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("DB transaction start error: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			log.Printf("Rollback error: %v", rollbackErr)
		}
	}()

	// 1. Fetch unbilled weighing slips
	rows, err := tx.Query(`
		SELECT wiegezettel_id, gewicht, material_id 
		FROM wiegezettel 
		WHERE kundennummer = ? AND datum >= ? AND datum <= ? AND rechnungsposition_id IS NULL
	`, kNummer, start, endeInclusive)
	if err != nil {
		log.Printf("Query weighing slips error: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ticket struct {
		id      int64
		gewicht float64
	}
	groups := make(map[int64][]ticket)
	var hasSlips bool

	for rows.Next() {
		var t ticket
		var matId int64
		if scanErr := rows.Scan(&t.id, &t.gewicht, &matId); scanErr == nil {
			groups[matId] = append(groups[matId], t)
			hasSlips = true
		}
	}

	if !hasSlips {
		// Redirect with error
		http.Redirect(w, r, "/rechnungen/neu?error=Keine+offenen+Wiegezettel+gefunden", http.StatusSeeOther)
		return
	}

	// 2. Fetch customer details
	var kundeName, kStrasse, kPlz, kOrt, kIban, kBic, kUstId, kZahlungsart string
	var kZahlzielTage int
	err = tx.QueryRow(`
		SELECT kundenname, COALESCE(strasse, ''), COALESCE(plz, ''), COALESCE(ort, ''), 
		       COALESCE(iban, ''), COALESCE(bic, ''), COALESCE(ust_id_nr, ''), 
		       zahlungsziel_tage, zahlungsart 
		FROM kunden WHERE kundennummer = ?
	`, kNummer).Scan(&kundeName, &kStrasse, &kPlz, &kOrt, &kIban, &kBic, &kUstId, &kZahlzielTage, &kZahlungsart)
	if err != nil {
		log.Printf("Kunden query error: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	// 3. Create Invoice Record
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM rechnungen").Scan(&count)
	if err != nil {
		log.Printf("Error counting invoices: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	rechnungsNummer := fmt.Sprintf("RE-%d-%04d", time.Now().Year(), count+1)
	rechnungsDatum := time.Now()
	faelligkeitsDatum := rechnungsDatum.AddDate(0, 0, kZahlzielTage)

	res, err := tx.Exec(`
		INSERT INTO rechnungen (
			rechnungsnummer, kundennummer, rechnungsdatum, faelligkeitsdatum, 
			leistungszeitraum_start, leistungszeitraum_ende, netto_summe, steuer_summe, brutto_summe, zugferd_xml
		) VALUES (?, ?, ?, ?, ?, ?, 0.0, 0.0, 0.0, '')`,
		rechnungsNummer, kNummer, rechnungsDatum.Format("2006-01-02"), faelligkeitsDatum.Format("2006-01-02"),
		start.Format("2006-01-02"), ende.Format("2006-01-02"),
	)
	if err != nil {
		log.Printf("Insert invoice error: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	invoiceId, err := res.LastInsertId()
	if err != nil {
		log.Printf("Failed to get invoice ID: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	// 4. Create Invoice Items and link Weighing Slips
	posNum := 1
	var totalNet, totalTax float64
	var lineData []zugferd.InvoiceLineData

	for matId, tickets := range groups {
		// Sum weights
		var totalWeight float64
		for _, t := range tickets {
			totalWeight += t.gewicht
		}

		// Resolve price (Sonderpreis override or standard fallback)
		var price float64
		err = tx.QueryRow("SELECT sonder_nettopreis FROM preislisten WHERE kundennummer = ? AND material_id = ?", kNummer, matId).Scan(&price)
		if err == sql.ErrNoRows {
			// fallback
			err = tx.QueryRow("SELECT standard_nettopreis FROM materialarten WHERE material_id = ?", matId).Scan(&price)
			if err != nil {
				log.Printf("Standard price query error: %v", err)
				http.Error(w, "DB Error", http.StatusInternalServerError)
				return
			}
		} else if err != nil {
			log.Printf("Custom price query error: %v", err)
			http.Error(w, "DB Error", http.StatusInternalServerError)
			return
		}

		// Fetch material metadata
		var matName, einheit string
		var mwstSatz float64
		err = tx.QueryRow("SELECT materialname, einheit, mwst_satz FROM materialarten WHERE material_id = ?", matId).Scan(&matName, &einheit, &mwstSatz)
		if err != nil {
			log.Printf("Material query error: %v", err)
			http.Error(w, "DB Error", http.StatusInternalServerError)
			return
		}

		lineNet := totalWeight * price
		lineNet = float64(int(lineNet*100+0.5)) / 100.0 // kaufmännisch runden
		lineTax := lineNet * (mwstSatz / 100.0)
		lineTax = float64(int(lineTax*100+0.5)) / 100.0

		// Insert invoice item
		var resItem sql.Result
		resItem, err = tx.Exec(`
			INSERT INTO rechnungspositionen (
				rechnung_id, material_id, positionsnummer, gesamtgewicht, einheit, 
				einzelpreis_netto, gesamtpreis_netto, mwst_satz, steuerbetrag
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			invoiceId, matId, posNum, totalWeight, einheit, price, lineNet, mwstSatz, lineTax,
		)
		if err != nil {
			log.Printf("Insert invoice item error: %v", err)
			http.Error(w, "DB Error", http.StatusInternalServerError)
			return
		}
		itemId, errLast := resItem.LastInsertId()
		if errLast != nil {
			log.Printf("Failed to get invoice item ID: %v", errLast)
			http.Error(w, "DB Error", http.StatusInternalServerError)
			return
		}

		// Link weighing slips
		for _, t := range tickets {
			_, err = tx.Exec("UPDATE wiegezettel SET rechnungsposition_id = ? WHERE wiegezettel_id = ?", itemId, t.id)
			if err != nil {
				log.Printf("Update wiegezettel link error: %v", err)
				http.Error(w, "DB Error", http.StatusInternalServerError)
				return
			}
		}

		totalNet += lineNet
		totalTax += lineTax

		lineData = append(lineData, zugferd.InvoiceLineData{
			PosNum:      posNum,
			Material:    matName,
			Menge:       totalWeight,
			Einheit:     einheit,
			Einzelpreis: price,
			Netto:       lineNet,
			MwstSatz:    mwstSatz,
		})

		posNum++
	}

	totalNet = float64(int(totalNet*100+0.5)) / 100.0
	totalTax = float64(int(totalTax*100+0.5)) / 100.0
	totalGross := totalNet + totalTax

	// 5. Fetch company details
	var ownName, ownInhaber, ownStrasse, ownPlz, ownOrt, ownLand, ownEmail, ownUstId, ownSteuer, ownBank, ownIban, ownBic string
	err = tx.QueryRow(`
		SELECT firmenname, inhaber, strasse, plz, ort, landescode, email, ust_id_nr, steuernummer, bankname, iban, bic 
		FROM eigene_stammdaten WHERE id = 1
	`).Scan(&ownName, &ownInhaber, &ownStrasse, &ownPlz, &ownOrt, &ownLand, &ownEmail, &ownUstId, &ownSteuer, &ownBank, &ownIban, &ownBic)
	if err != nil {
		log.Printf("Own company details query error: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	// 6. Generate ZUGFeRD XML
	invData := &zugferd.InvoiceData{
		Rechnungsnummer:        rechnungsNummer,
		Rechnungsdatum:         rechnungsDatum,
		Faelligkeitsdatum:      faelligkeitsDatum,
		LeistungszeitraumStart: start,
		LeistungszeitraumEnde:  ende,

		VerkaeuferName:      ownName,
		VerkaeuferStrasse:   ownStrasse,
		VerkaeuferPlz:       ownPlz,
		VerkaeuferOrt:       ownOrt,
		VerkaeuferLand:      ownLand,
		VerkaeuferUstId:     ownUstId,
		VerkaeuferSteuerNum: ownSteuer,
		VerkaeuferIban:      ownIban,
		VerkaeuferBic:       ownBic,

		KundeName:    kundeName,
		KundeStrasse: kStrasse,
		KundePlz:     kPlz,
		KundeOrt:     kOrt,
		KundeLand:    "DE",
		KundeUstId:   kUstId,

		Lines: lineData,
	}

	xmlBytes, err := zugferd.GenerateXML(invData)
	if err != nil {
		log.Printf("ZUGFeRD generation error: %v", err)
		http.Error(w, "ZUGFeRD Error", http.StatusInternalServerError)
		return
	}

	// 7. Generate PDF/A-3
	pdfDir := filepath.Join("internal", "server", "static", "rechnungen")
	if dirErr := os.MkdirAll(pdfDir, 0755); dirErr != nil {
		log.Printf("Failed to create PDF folder: %v", dirErr)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
	pdfFileName := fmt.Sprintf("%s.pdf", rechnungsNummer)
	pdfPath := filepath.Join(pdfDir, pdfFileName)

	err = pdf.GenerateInvoicePDF(invData, xmlBytes, pdfPath)
	if err != nil {
		log.Printf("PDF generation error: %v", err)
		http.Error(w, "PDF Error", http.StatusInternalServerError)
		return
	}

	// 8. Update Invoice totals & XML path
	staticPDFLink := "/static/rechnungen/" + pdfFileName
	_, err = tx.Exec(`
		UPDATE rechnungen 
		SET netto_summe = ?, steuer_summe = ?, brutto_summe = ?, zugferd_xml = ?, pdf_dateipfad = ? 
		WHERE rechnung_id = ?`,
		totalNet, totalTax, totalGross, string(xmlBytes), staticPDFLink, invoiceId,
	)
	if err != nil {
		log.Printf("Update invoice error: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Printf("Commit error: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/rechnungen", http.StatusSeeOther)
}

// handleInvoiceView renders the HTML invoice on screen.
func (s *Server) handleInvoiceView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	var inv templates.InvoiceDetails
	var dateStr, dueStr, startStr, endStr string
	var customerId int64

	err = s.db.QueryRow(`
		SELECT r.rechnungsnummer, r.rechnungsdatum, r.faelligkeitsdatum, 
		       r.leistungszeitraum_start, r.leistungszeitraum_ende,
		       r.netto_summe, r.steuer_summe, r.brutto_summe, r.zahlungsstatus,
		       r.kundennummer, k.kundenname, COALESCE(k.strasse, ''), COALESCE(k.plz, ''), COALESCE(k.ort, '')
		FROM rechnungen r
		JOIN kunden k ON r.kundennummer = k.kundennummer
		WHERE r.rechnung_id = ?
	`, id).Scan(
		&inv.Rechnungsnummer, &dateStr, &dueStr, &startStr, &endStr,
		&inv.NettoSumme, &inv.SteuerSumme, &inv.BruttoSumme, &inv.Zahlungsstatus,
		&customerId, &inv.Kundenname, &inv.KundeStrasse, &inv.KundePlz, &inv.KundeOrt,
	)
	if err != nil {
		log.Printf("Query invoice details error: %v", err)
		http.NotFound(w, r)
		return
	}

	if t, parseErr := time.Parse("2006-01-02", dateStr[:10]); parseErr == nil {
		inv.Rechnungsdatum = t
	}
	if t, parseErr := time.Parse("2006-01-02", dueStr[:10]); parseErr == nil {
		inv.Faelligkeitsdatum = t
	}
	if t, parseErr := time.Parse("2006-01-02", startStr[:10]); parseErr == nil {
		inv.LeistungszeitraumStart = t
	}
	if t, parseErr := time.Parse("2006-01-02", endStr[:10]); parseErr == nil {
		inv.LeistungszeitraumEnde = t
	}
	inv.Kundennummer = customerId

	// Fetch own company data
	err = s.db.QueryRow(`
		SELECT firmenname, strasse, plz, ort, iban, bic 
		FROM eigene_stammdaten WHERE id = 1
	`).Scan(&inv.VerkaeuferName, &inv.VerkaeuferStrasse, &inv.VerkaeuferPlz, &inv.VerkaeuferOrt, &inv.VerkaeuferIban, &inv.VerkaeuferBic)
	if err != nil {
		log.Printf("Query own company data error: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	// Fetch invoice items
	rows, err := s.db.Query(`
		SELECT rp.positionsnummer, m.materialname, rp.gesamtgewicht, rp.einheit, 
		       rp.einzelpreis_netto, rp.gesamtpreis_netto, rp.mwst_satz, rp.steuerbetrag
		FROM rechnungspositionen rp
		JOIN materialarten m ON rp.material_id = m.material_id
		WHERE rp.rechnung_id = ?
		ORDER BY rp.positionsnummer ASC
	`, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var line templates.InvoiceLineDetail
			scanErr := rows.Scan(
				&line.Positionsnummer, &line.Materialname, &line.Gesamtgewicht, &line.Einheit,
				&line.EinzelpreisNetto, &line.GesamtpreisNetto, &line.MwstSatz, &line.Steuerbetrag,
			)
			if scanErr == nil {
				inv.Lines = append(inv.Lines, line)
			}
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	comp := templates.InvoicePreview(&inv)
	layout := templates.Layout(fmt.Sprintf("Rechnung %s", inv.Rechnungsnummer), comp)
	if err := layout.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering invoice preview: %v", err)
	}
}

// handleInvoiceXML serves the raw ZUGFeRD XML download.
func (s *Server) handleInvoiceXML(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	var rechnungsnummer, xmlContent string
	err = s.db.QueryRow("SELECT rechnungsnummer, zugferd_xml FROM rechnungen WHERE rechnung_id = ?", id).Scan(&rechnungsnummer, &xmlContent)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=ZUGFeRD-%s.xml", rechnungsnummer))
	if _, writeErr := w.Write([]byte(xmlContent)); writeErr != nil {
		log.Printf("Failed to write XML response: %v", writeErr)
	}
}

// handleInvoicePDF triggers the PDF download.
func (s *Server) handleInvoicePDF(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	var rechnungsnummer, pdfPath string
	err = s.db.QueryRow("SELECT rechnungsnummer, pdf_dateipfad FROM rechnungen WHERE rechnung_id = ?", id).Scan(&rechnungsnummer, &pdfPath)
	if err != nil || pdfPath == "" {
		http.NotFound(w, r)
		return
	}

	// Resolve local path
	localPath := filepath.Join("internal", "server", "static", "rechnungen", fmt.Sprintf("%s.pdf", rechnungsnummer))

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", rechnungsnummer))
	http.ServeFile(w, r, localPath)
}

// handleInvoiceCancel deletes the invoice records, unlinks all associated weighing slips, and deletes the PDF.
func (s *Server) handleInvoiceCancel(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("DB transaction start error: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			log.Printf("Rollback error: %v", rollbackErr)
		}
	}()

	// 1. Get PDF path to delete later
	var pdfPath string
	err = tx.QueryRow("SELECT pdf_dateipfad FROM rechnungen WHERE rechnung_id = ?", id).Scan(&pdfPath)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Failed to select pdf_dateipfad: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	// 2. Reset wiegezettel references to open
	_, err = tx.Exec(`
		UPDATE wiegezettel 
		SET rechnungsposition_id = NULL 
		WHERE rechnungsposition_id IN (
			SELECT rechnungsposition_id FROM rechnungspositionen WHERE rechnung_id = ?
		)
	`, id)
	if err != nil {
		log.Printf("Failed to unlink weighing slips: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	// 3. Delete invoice positions
	_, err = tx.Exec("DELETE FROM rechnungspositionen WHERE rechnung_id = ?", id)
	if err != nil {
		log.Printf("Failed to delete positions: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	// 4. Delete invoice itself
	_, err = tx.Exec("DELETE FROM rechnungen WHERE rechnung_id = ?", id)
	if err != nil {
		log.Printf("Failed to delete invoice: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Printf("Commit error: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	// 5. Delete physical PDF file if present
	if pdfPath != "" {
		pdfFileName := filepath.Base(pdfPath)
		localPath := filepath.Join("internal", "server", "static", "rechnungen", pdfFileName)
		_ = os.Remove(localPath)
	}

	// Redirect using HTMX Header or standard HTTP redirect
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Redirect", "/rechnungen")
		w.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(w, r, "/rechnungen", http.StatusSeeOther)
	}
}

// queryWiegezettel is a helper that parses query parameters and executes the SQL query.
func (s *Server) queryWiegezettel(r *http.Request) ([]templates.WiegezettelRow, int64, int64, string, string, error) {
	kIDStr := r.FormValue("kunde_id")
	mIDStr := r.FormValue("material_id")
	status := r.FormValue("status")
	search := r.FormValue("search")

	var kID, mID int64
	if kIDStr != "" {
		if id, err := strconv.ParseInt(kIDStr, 10, 64); err == nil {
			kID = id
		}
	}
	if mIDStr != "" {
		if id, err := strconv.ParseInt(mIDStr, 10, 64); err == nil {
			mID = id
		}
	}
	if status == "" {
		status = "all"
	}

	query := `
		SELECT w.wiegezettel_id, w.kundennummer, k.kundenname, w.datum, w.gewicht, 
		       w.material_id, m.materialname, m.einheit, w.anlieferungsort, w.anlieferer, 
		       COALESCE(w.referenz, ''), w.rechnungsposition_id, r.rechnungsnummer
		FROM wiegezettel w
		JOIN kunden k ON w.kundennummer = k.kundennummer
		JOIN materialarten m ON w.material_id = m.material_id
		LEFT JOIN rechnungspositionen rp ON w.rechnungsposition_id = rp.rechnungsposition_id
		LEFT JOIN rechnungen r ON rp.rechnung_id = r.rechnung_id
		WHERE 1=1
	`
	var args []interface{}

	if kID > 0 {
		query += " AND w.kundennummer = ?"
		args = append(args, kID)
	}
	if mID > 0 {
		query += " AND w.material_id = ?"
		args = append(args, mID)
	}
	if status == "open" {
		query += " AND w.rechnungsposition_id IS NULL"
	} else if status == "billed" {
		query += " AND w.rechnungsposition_id IS NOT NULL"
	}
	if search != "" {
		query += " AND (w.wiegezettel_id = ? OR w.referenz LIKE ?)"
		searchInt, err := strconv.ParseInt(search, 10, 64)
		if err == nil {
			args = append(args, searchInt, "%"+search+"%")
		} else {
			args = append(args, -1, "%"+search+"%")
		}
	}

	query += " ORDER BY w.datum DESC, w.wiegezettel_id DESC LIMIT 100"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, 0, "", "", err
	}
	defer rows.Close()

	var result []templates.WiegezettelRow
	for rows.Next() {
		var row templates.WiegezettelRow
		var dateStr string
		var rpID sql.NullInt64
		var rNum sql.NullString
		scanErr := rows.Scan(
			&row.ID, &row.Kundennummer, &row.Kundenname, &dateStr, &row.Gewicht,
			&row.MaterialID, &row.Materialname, &row.Einheit, &row.Anlieferungsort, &row.Anlieferer,
			&row.Referenz, &rpID, &rNum,
		)
		if scanErr != nil {
			log.Printf("Scan error in queryWiegezettel: %v", scanErr)
			continue
		}

		var parsedTime time.Time
		var parseErr error
		layouts := []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05Z", "2006-01-02T15:04:05", "2006-01-02"}
		for _, layout := range layouts {
			if parsedTime, parseErr = time.Parse(layout, dateStr[:Min(len(dateStr), len(layout))]); parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			row.Datum = time.Now()
		} else {
			row.Datum = parsedTime
		}

		row.RechnungspositionID = rpID
		row.Rechnungsnummer = rNum
		result = append(result, row)
	}

	return result, kID, mID, status, search, nil
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *Server) handleWiegezettelList(w http.ResponseWriter, r *http.Request) {
	rows, kID, mID, status, search, err := s.queryWiegezettel(r)
	if err != nil {
		log.Printf("Error querying weighing slips: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	kunden, materialien, orte, anliefererList, err := s.fetchLookups()
	if err != nil {
		log.Printf("Lookup fetch error: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page := templates.WiegezettelPage(rows, kunden, materialien, orte, anliefererList, kID, mID, status, search)
	layout := templates.Layout("Wiegezettel", page)
	if err := layout.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering Wiegezettel page: %v", err)
	}
}

func (s *Server) handleWiegezettelTable(w http.ResponseWriter, r *http.Request) {
	rows, _, _, _, _, err := s.queryWiegezettel(r)
	if err != nil {
		log.Printf("Error querying weighing slips: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	comp := templates.WiegezettelTable(rows)
	if err := comp.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering Wiegezettel table: %v", err)
	}
}

func (s *Server) fetchLookups() ([]templates.KundeLookup, []templates.MaterialLookup, []string, []string, error) {
	kRows, err := s.db.Query("SELECT kundennummer, kundenname FROM kunden ORDER BY kundenname ASC")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer kRows.Close()
	var kunden []templates.KundeLookup
	for kRows.Next() {
		var k templates.KundeLookup
		if scanErr := kRows.Scan(&k.Kundennummer, &k.Kundenname); scanErr == nil {
			kunden = append(kunden, k)
		}
	}

	mRows, err := s.db.Query("SELECT material_id, materialname, einheit FROM materialarten ORDER BY materialname ASC")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer mRows.Close()
	var materialien []templates.MaterialLookup
	for mRows.Next() {
		var m templates.MaterialLookup
		if scanErr := mRows.Scan(&m.MaterialID, &m.Materialname, &m.Einheit); scanErr == nil {
			materialien = append(materialien, m)
		}
	}

	oRows, err := s.db.Query("SELECT anlieferungs_ort FROM anlieferungsorte ORDER BY anlieferungs_ort ASC")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer oRows.Close()
	var orte []string
	for oRows.Next() {
		var o string
		if scanErr := oRows.Scan(&o); scanErr == nil {
			orte = append(orte, o)
		}
	}

	aRows, err := s.db.Query("SELECT anlieferer_herkunft FROM anlieferer ORDER BY anlieferer_herkunft ASC")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer aRows.Close()
	var anliefererList []string
	for aRows.Next() {
		var a string
		if scanErr := aRows.Scan(&a); scanErr == nil {
			anliefererList = append(anliefererList, a)
		}
	}

	return kunden, materialien, orte, anliefererList, nil
}

func (s *Server) handleWiegezettelNewForm(w http.ResponseWriter, r *http.Request) {
	kunden, materialien, orte, anliefererList, err := s.fetchLookups()
	if err != nil {
		log.Printf("Lookup fetch error: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	var nextID int64
	err = s.db.QueryRow("SELECT COALESCE(MAX(wiegezettel_id), 0) + 1 FROM wiegezettel").Scan(&nextID)
	if err != nil {
		nextID = 1
	}

	var dummyRow templates.WiegezettelRow
	dummyRow.Datum = time.Now()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	form := templates.WiegezettelForm(&dummyRow, kunden, materialien, orte, anliefererList, false, nextID)
	layout := templates.Layout("Wiegezettel erfassen", form)
	if err := layout.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering new weighing slip form: %v", err)
	}
}

func (s *Server) handleWiegezettelCreate(w http.ResponseWriter, r *http.Request) {
	idStr := r.FormValue("wiegezettel_id")
	kNumStr := r.FormValue("kundennummer")
	datumStr := r.FormValue("datum")
	gewichtStr := r.FormValue("gewicht")
	materialIDStr := r.FormValue("material_id")
	ort := r.FormValue("anlieferungsort")
	anlieferer := r.FormValue("anlieferer")
	referenz := r.FormValue("referenz")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return
	}
	kNum, err := strconv.ParseInt(kNumStr, 10, 64)
	if err != nil {
		http.Error(w, "Ungültiger Kunde", http.StatusBadRequest)
		return
	}
	datum, err := time.Parse("2006-01-02T15:04", datumStr)
	if err != nil {
		http.Error(w, "Ungültiges Datum", http.StatusBadRequest)
		return
	}
	gewicht, err := strconv.ParseFloat(gewichtStr, 64)
	if err != nil {
		http.Error(w, "Ungültiges Gewicht", http.StatusBadRequest)
		return
	}
	materialID, err := strconv.ParseInt(materialIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Ungültiges Material", http.StatusBadRequest)
		return
	}

	_, err = s.db.Exec(`
		INSERT INTO wiegezettel (
			wiegezettel_id, kundennummer, datum, gewicht, material_id, anlieferungsort, anlieferer, referenz
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, kNum, datum.Format("2006-01-02 15:04:05"), gewicht, materialID, ort, anlieferer, referenz,
	)
	if err != nil {
		log.Printf("Insert Wiegezettel error: %v", err)
		http.Error(w, "Datenbankfehler beim Einfügen", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/wiegezettel", http.StatusSeeOther)
}

func (s *Server) handleWiegezettelEditForm(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	var row templates.WiegezettelRow
	var dateStr string
	var rpID sql.NullInt64
	err = s.db.QueryRow(`
		SELECT wiegezettel_id, kundennummer, datum, gewicht, material_id, 
		       anlieferungsort, anlieferer, COALESCE(referenz, ''), rechnungsposition_id
		FROM wiegezettel WHERE wiegezettel_id = ?
	`, id).Scan(
		&row.ID, &row.Kundennummer, &dateStr, &row.Gewicht, &row.MaterialID,
		&row.Anlieferungsort, &row.Anlieferer, &row.Referenz, &rpID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			log.Printf("Query error fetching Wiegezettel: %v", err)
			http.Error(w, "DB Error", http.StatusInternalServerError)
		}
		return
	}

	row.RechnungspositionID = rpID

	if rpID.Valid {
		http.Error(w, "Dieser Wiegezettel wurde bereits abgerechnet und kann nicht bearbeitet werden.", http.StatusForbidden)
		return
	}

	var parsedTime time.Time
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05Z", "2006-01-02T15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if parsedTime, err = time.Parse(layout, dateStr[:Min(len(dateStr), len(layout))]); err == nil {
			break
		}
	}
	if err != nil {
		row.Datum = time.Now()
	} else {
		row.Datum = parsedTime
	}

	kunden, materialien, orte, anliefererList, err := s.fetchLookups()
	if err != nil {
		log.Printf("Lookup fetch error: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	form := templates.WiegezettelForm(&row, kunden, materialien, orte, anliefererList, true, 0)
	layout := templates.Layout("Wiegezettel bearbeiten", form)
	if err := layout.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering edit weighing slip form: %v", err)
	}
}

func (s *Server) handleWiegezettelUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	var rpID sql.NullInt64
	err = s.db.QueryRow("SELECT rechnungsposition_id FROM wiegezettel WHERE wiegezettel_id = ?", id).Scan(&rpID)
	if err != nil {
		log.Printf("Query error check billed: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	if rpID.Valid {
		http.Error(w, "Dieser Wiegezettel wurde bereits abgerechnet und kann nicht bearbeitet werden.", http.StatusForbidden)
		return
	}

	kNumStr := r.FormValue("kundennummer")
	datumStr := r.FormValue("datum")
	gewichtStr := r.FormValue("gewicht")
	materialIDStr := r.FormValue("material_id")
	ort := r.FormValue("anlieferungsort")
	anlieferer := r.FormValue("anlieferer")
	referenz := r.FormValue("referenz")

	kNum, err := strconv.ParseInt(kNumStr, 10, 64)
	if err != nil {
		http.Error(w, "Ungültiger Kunde", http.StatusBadRequest)
		return
	}
	datum, err := time.Parse("2006-01-02T15:04", datumStr)
	if err != nil {
		http.Error(w, "Ungültiges Datum", http.StatusBadRequest)
		return
	}
	gewicht, err := strconv.ParseFloat(gewichtStr, 64)
	if err != nil {
		http.Error(w, "Ungültiges Gewicht", http.StatusBadRequest)
		return
	}
	materialID, err := strconv.ParseInt(materialIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Ungültiges Material", http.StatusBadRequest)
		return
	}

	_, err = s.db.Exec(`
		UPDATE wiegezettel 
		SET kundennummer = ?, datum = ?, gewicht = ?, material_id = ?, 
		    anlieferungsort = ?, anlieferer = ?, referenz = ?
		WHERE wiegezettel_id = ?`,
		kNum, datum.Format("2006-01-02 15:04:05"), gewicht, materialID, ort, anlieferer, referenz, id,
	)
	if err != nil {
		log.Printf("Update Wiegezettel error: %v", err)
		http.Error(w, "Datenbankfehler beim Aktualisieren", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/wiegezettel", http.StatusSeeOther)
}

func (s *Server) handleWiegezettelDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	var rpID sql.NullInt64
	err = s.db.QueryRow("SELECT rechnungsposition_id FROM wiegezettel WHERE wiegezettel_id = ?", id).Scan(&rpID)
	if err != nil {
		log.Printf("Query error check billed: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	if rpID.Valid {
		http.Error(w, "Dieser Wiegezettel wurde bereits abgerechnet und kann nicht gelöscht werden.", http.StatusForbidden)
		return
	}

	_, err = s.db.Exec("DELETE FROM wiegezettel WHERE wiegezettel_id = ?", id)
	if err != nil {
		log.Printf("Delete Wiegezettel error: %v", err)
		http.Error(w, "Datenbankfehler beim Löschen", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
