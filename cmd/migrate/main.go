package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "modernc.org/sqlite"
)

// XML structures matching legacy formats
type DatarootAnlieferer struct {
	XMLName xml.Name `xml:"dataroot"`
	Items   []struct {
		AnliefererHerkunft string `xml:"AnliefererHerkunft"`
	} `xml:"tblAnlieferer"`
}

type DatarootAnlieferungOrt struct {
	XMLName xml.Name `xml:"dataroot"`
	Items   []struct {
		AnlieferungsOrt string `xml:"AnlieferungsOrt"`
	} `xml:"tblAnlieferungOrt"`
}

type DatarootMaterialarten struct {
	XMLName xml.Name `xml:"dataroot"`
	Items   []struct {
		Material   string  `xml:"Material"`
		Nettopreis float64 `xml:"Nettopreis"`
		Mwst       float64 `xml:"Mwst"`
		Einheit    string  `xml:"Einheit"`
	} `xml:"tblMaterialarten"`
}

type DatarootFirmenliste struct {
	XMLName xml.Name `xml:"dataroot"`
	Items   []struct {
		Kundenkürzel    string `xml:"Kundenkürzel"`
		Kundenname      string `xml:"Kundenname"`
		Namenserweitung string `xml:"Namenserweitung"`
		Privat          int    `xml:"Privat_x0020__x003F_"`
		Brief           int    `xml:"Brief_x003F_"`
		Adresse         string `xml:"Adresse"`
		Postleitzahl    string `xml:"Postleitzahl"`
		Ort             string `xml:"Ort"`
		Telefon         string `xml:"Telefonnummer"`
		Fax             string `xml:"Faxnummer"`
		EmailAllgemein  string `xml:"E-Mail_Allgemein"`
		EmailRechnung   string `xml:"E-Mail_Rechnung"`
		Kontaktperson   string `xml:"Kontaktperson"`
		Iban            string `xml:"IBAN"`
		Bic             string `xml:"BIC"`
		Anmerkungen     string `xml:"Anmerkungen"`
	} `xml:"tblFirmenliste"`
}

type DatarootPreisliste struct {
	XMLName xml.Name `xml:"dataroot"`
	Items   []struct {
		Kundenkürzel string  `xml:"Kundenkürzel"`
		Material     string  `xml:"Material"`
		Nettopreis   float64 `xml:"Nettopreis"`
	} `xml:"tblPreisliste"`
}

type DatarootWiegezettel struct {
	XMLName xml.Name `xml:"dataroot"`
	Items   []struct {
		LfdNr          int     `xml:"Lfd-Nr"`
		Kundenkürzel   string  `xml:"Kundenkürzel"`
		Datum          string  `xml:"Datum"`
		Gewicht        float64 `xml:"Gewicht"`
		Material       string  `xml:"Material"`
		AnlieferungOrt string  `xml:"AnlieferungOrt"`
		Anlieferer     string  `xml:"Anlieferer"`
		Referenz       string  `xml:"Referenz"`
	} `xml:"tblWiegezettel"`
}

func main() {
	dbPath := "faktura.db"
	dataDir := filepath.Join("docs", "migration", "data")

	log.Printf("Starting data migration to %s...", dbPath)

	// Remove existing db to perform fresh migration
	os.Remove(dbPath)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Initialize tables using schema.sql
	schemaBytes, err := os.ReadFile(filepath.Join("internal", "db", "schema.sql"))
	if err != nil {
		log.Fatalf("Error reading schema.sql: %v", err)
	}

	_, err = db.Exec(string(schemaBytes))
	if err != nil {
		log.Fatalf("Error executing schema: %v", err)
	}
	log.Println("Database schema successfully initialized.")

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Error starting transaction: %v", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			log.Printf("Rollback error: %v", rollbackErr)
		}
	}()

	// 1. Migrate Anlieferer
	log.Println("Migrating Anlieferer...")
	anliefererFile, err := os.Open(filepath.Join(dataDir, "tblAnlieferer.xml"))
	if err != nil {
		log.Fatalf("Error opening Anlieferer XML: %v", err)
	}
	defer anliefererFile.Close()

	var datarootAnlieferer DatarootAnlieferer
	err = xml.NewDecoder(anliefererFile).Decode(&datarootAnlieferer)
	if err != nil {
		log.Fatalf("Error decoding Anlieferer XML: %v", err)
	}

	for _, item := range datarootAnlieferer.Items {
		if item.AnliefererHerkunft == "" {
			continue
		}
		_, err = tx.Exec("INSERT OR IGNORE INTO anlieferer (anlieferer_herkunft) VALUES (?)", item.AnliefererHerkunft)
		if err != nil {
			log.Fatalf("Error inserting Anlieferer: %v", err)
		}
	}

	// 2. Migrate Anlieferungsort
	log.Println("Migrating Anlieferungsorte...")
	ortFile, err := os.Open(filepath.Join(dataDir, "tblAnlieferungOrt.xml"))
	if err != nil {
		log.Fatalf("Error opening AnlieferungOrt XML: %v", err)
	}
	defer ortFile.Close()

	var datarootOrt DatarootAnlieferungOrt
	err = xml.NewDecoder(ortFile).Decode(&datarootOrt)
	if err != nil {
		log.Fatalf("Error decoding AnlieferungOrt XML: %v", err)
	}

	for _, item := range datarootOrt.Items {
		if item.AnlieferungsOrt == "" {
			continue
		}
		_, err = tx.Exec("INSERT OR IGNORE INTO anlieferungsorte (anlieferungs_ort) VALUES (?)", item.AnlieferungsOrt)
		if err != nil {
			log.Fatalf("Error inserting Anlieferungsort: %v", err)
		}
	}

	// 3. Migrate Materialarten
	log.Println("Migrating Materialarten...")
	matFile, err := os.Open(filepath.Join(dataDir, "tblMaterialarten.xml"))
	if err != nil {
		log.Fatalf("Error opening Materialarten XML: %v", err)
	}
	defer matFile.Close()

	var datarootMat DatarootMaterialarten
	err = xml.NewDecoder(matFile).Decode(&datarootMat)
	if err != nil {
		log.Fatalf("Error decoding Materialarten XML: %v", err)
	}

	materialMap := make(map[string]int64)
	materialPriceMap := make(map[int64]float64)

	for _, item := range datarootMat.Items {
		if item.Material == "" {
			continue
		}
		// Convert Mwst fraction (e.g. 0.19) to percentage (19.00)
		mwstSatz := item.Mwst * 100.0
		if mwstSatz == 0.0 {
			mwstSatz = 19.0 // default
		}
		var res sql.Result
		res, err = tx.Exec("INSERT INTO materialarten (materialname, standard_nettopreis, mwst_satz, einheit) VALUES (?, ?, ?, ?)",
			item.Material, item.Nettopreis, mwstSatz, item.Einheit)
		if err != nil {
			log.Fatalf("Error inserting Materialart: %v", err)
		}
		id, errLast := res.LastInsertId()
		if errLast != nil {
			log.Fatalf("Error getting last insert ID: %v", errLast)
		}
		materialMap[item.Material] = id
		materialPriceMap[id] = item.Nettopreis
	}

	// 4. Migrate Kunden (tblFirmenliste)
	log.Println("Migrating Kunden...")
	kundenFile, err := os.Open(filepath.Join(dataDir, "tblFirmenliste.xml"))
	if err != nil {
		log.Fatalf("Error opening Firmenliste XML: %v", err)
	}
	defer kundenFile.Close()

	var datarootKunden DatarootFirmenliste
	err = xml.NewDecoder(kundenFile).Decode(&datarootKunden)
	if err != nil {
		log.Fatalf("Error decoding Firmenliste XML: %v", err)
	}

	// Extract and sort short-codes alphabetically to assign stable IDs
	var shortcuts []string
	shortcutMap := make(map[string]int) // index in Dataroot
	for idx, item := range datarootKunden.Items {
		if item.Kundenkürzel == "" {
			continue
		}
		shortcuts = append(shortcuts, item.Kundenkürzel)
		shortcutMap[item.Kundenkürzel] = idx
	}
	sort.Strings(shortcuts)

	customerMapping := make(map[string]int) // shortcut -> Kundennummer
	startNum := 10001

	for _, kuerzel := range shortcuts {
		idx := shortcutMap[kuerzel]
		item := datarootKunden.Items[idx]

		kundenNummer := startNum
		customerMapping[kuerzel] = kundenNummer
		startNum++

		_, err = tx.Exec(`
			INSERT INTO kunden (
				kundennummer, kundenname, namenserweiterung, ist_privat, brief_versand,
				strasse, plz, ort, landescode, telefon, fax, email_allgemein, email_rechnung,
				kontaktperson, iban, bic, ust_id_nr, steuernummer, leitweg_id, zahlungsziel_tage, zahlungsart
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			kundenNummer, item.Kundenname, item.Namenserweitung, item.Privat == 1, item.Brief == 1,
			item.Adresse, item.Postleitzahl, item.Ort, "DE", item.Telefon, item.Fax,
			item.EmailAllgemein, item.EmailRechnung, item.Kontaktperson, item.Iban, item.Bic,
			nil, nil, nil, 14, "Ueberweisung", // ZUGFeRD specific defaults
		)
		if err != nil {
			log.Fatalf("Error inserting Kunde %s (%d): %v", kuerzel, kundenNummer, err)
		}
	}

	// 5. Migrate Preislisten
	log.Println("Migrating Preislisten (Sonderpreise)...")
	preisFile, err := os.Open(filepath.Join(dataDir, "tblPreisliste.xml"))
	if err != nil {
		log.Fatalf("Error opening Preisliste XML: %v", err)
	}
	defer preisFile.Close()

	var datarootPreis DatarootPreisliste
	err = xml.NewDecoder(preisFile).Decode(&datarootPreis)
	if err != nil {
		log.Fatalf("Error decoding Preisliste XML: %v", err)
	}

	for _, item := range datarootPreis.Items {
		if item.Kundenkürzel == "" || item.Material == "" {
			continue
		}
		kNummer, ok1 := customerMapping[item.Kundenkürzel]
		matId, ok2 := materialMap[item.Material]

		if !ok1 || !ok2 {
			// Skip unresolvable records
			continue
		}

		// Deduplicate: only insert if it differs from the standard price
		stdPrice := materialPriceMap[matId]
		if item.Nettopreis == stdPrice {
			continue
		}

		_, err = tx.Exec("INSERT OR REPLACE INTO preislisten (kundennummer, material_id, sonder_nettopreis) VALUES (?, ?, ?)",
			kNummer, matId, item.Nettopreis)
		if err != nil {
			log.Fatalf("Error inserting Preisliste: %v", err)
		}
	}

	// 6. Migrate Wiegezettel
	log.Println("Migrating Wiegezettel...")
	wiegeFile, err := os.Open(filepath.Join(dataDir, "tblWiegezettel.xml"))
	if err != nil {
		log.Fatalf("Error opening Wiegezettel XML: %v", err)
	}
	defer wiegeFile.Close()

	var datarootWiege DatarootWiegezettel
	dec := xml.NewDecoder(wiegeFile)
	// Access XML exports can be large. Let's decode in one block since the file size is ~218KB (completely fine for decode).
	err = dec.Decode(&datarootWiege)
	if err != nil {
		log.Fatalf("Error decoding Wiegezettel XML: %v", err)
	}

	for _, item := range datarootWiege.Items {
		if item.LfdNr == 0 || item.Kundenkürzel == "" || item.Material == "" {
			continue
		}
		kNummer, ok1 := customerMapping[item.Kundenkürzel]
		matId, ok2 := materialMap[item.Material]

		if !ok1 || !ok2 {
			continue
		}

		// Parse Datum (Access format 2026-01-05T00:00:00)
		parsedTime, errTime := time.Parse("2006-01-02T15:04:05", item.Datum)
		if errTime != nil {
			log.Printf("Warning: error parsing date '%s' for Wiegezettel %d, using current time. Error: %v", item.Datum, item.LfdNr, errTime)
			parsedTime = time.Now()
		}

		// Ensure lookups exist
		if item.Anlieferer != "" {
			_, err = tx.Exec("INSERT OR IGNORE INTO anlieferer (anlieferer_herkunft) VALUES (?)", item.Anlieferer)
			if err != nil {
				log.Fatalf("Error inserting Anlieferer: %v", err)
			}
		}
		if item.AnlieferungOrt != "" {
			_, err = tx.Exec("INSERT OR IGNORE INTO anlieferungsorte (anlieferungs_ort) VALUES (?)", item.AnlieferungOrt)
			if err != nil {
				log.Fatalf("Error inserting Anlieferungsort: %v", err)
			}
		}

		_, err = tx.Exec(`
			INSERT OR REPLACE INTO wiegezettel (
				wiegezettel_id, kundennummer, datum, gewicht, material_id, anlieferungsort, anlieferer, referenz
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			item.LfdNr, kNummer, parsedTime, item.Gewicht, matId, item.AnlieferungOrt, item.Anlieferer, item.Referenz,
		)
		if err != nil {
			log.Fatalf("Error inserting Wiegezettel: %v", err)
		}
	}

	// 7. Initialize EigeneStammdaten using sample invoice values
	log.Println("Initializing EigeneStammdaten...")
	_, err = tx.Exec(`
		INSERT INTO eigene_stammdaten (
			id, firmenname, inhaber, strasse, plz, ort, landescode, email, ust_id_nr, steuernummer, bankname, iban, bic
		) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"Kompostwerk Büttelborn GmbH", "T. Geipman", "Auf der Hardt/An der B 42", "64572", "Büttelborn", "DE",
		"faktura@aws-service.com", "DE5089000048161405", "12/345/67890", "Sparkasse Musterstadt",
		"DE58508900000048161405", "GENODEF1VBD",
	)
	if err != nil {
		log.Fatalf("Error inserting EigeneStammdaten: %v", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Fatalf("Error committing migration: %v", err)
	}

	log.Println("Data migration completed successfully!")

	// Print some stats
	var count int
	if err = db.QueryRow("SELECT COUNT(*) FROM kunden").Scan(&count); err == nil {
		fmt.Printf("Migrated Customers: %d\n", count)
	}
	if err = db.QueryRow("SELECT COUNT(*) FROM materialarten").Scan(&count); err == nil {
		fmt.Printf("Migrated Material types: %d\n", count)
	}
	if err = db.QueryRow("SELECT COUNT(*) FROM preislisten").Scan(&count); err == nil {
		fmt.Printf("Migrated Price list overrides: %d\n", count)
	}
	if err = db.QueryRow("SELECT COUNT(*) FROM wiegezettel").Scan(&count); err == nil {
		fmt.Printf("Migrated Weighing slips: %d\n", count)
	}
}
