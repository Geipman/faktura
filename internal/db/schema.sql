-- SQLite-Schema für Faktura

-- Lookup-Tabelle für Anlieferer
CREATE TABLE IF NOT EXISTS anlieferer (
    anlieferer_herkunft TEXT PRIMARY KEY
);

-- Lookup-Tabelle für Übergabeorte
CREATE TABLE IF NOT EXISTS anlieferungsorte (
    anlieferungs_ort TEXT PRIMARY KEY
);

-- Eigene Stammdaten des Rechnungsausstellers (nur 1 Datensatz)
CREATE TABLE IF NOT EXISTS eigene_stammdaten (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    firmenname TEXT NOT NULL,
    inhaber TEXT NOT NULL,
    strasse TEXT NOT NULL,
    plz TEXT NOT NULL,
    ort TEXT NOT NULL,
    landescode TEXT NOT NULL DEFAULT 'DE',
    telefon TEXT,
    email TEXT NOT NULL,
    webseite TEXT,
    ust_id_nr TEXT NOT NULL,
    steuernummer TEXT NOT NULL,
    bankname TEXT NOT NULL,
    iban TEXT NOT NULL,
    bic TEXT NOT NULL,
    handelsregister TEXT,
    amtsgericht TEXT
);

-- Kunden-Stammdaten (Rechnungsempfänger)
CREATE TABLE IF NOT EXISTS kunden (
    kundennummer INTEGER PRIMARY KEY CHECK (kundennummer >= 10001 AND kundennummer <= 99999),
    kundenname TEXT NOT NULL,
    namenserweiterung TEXT,
    ist_privat BOOLEAN NOT NULL DEFAULT 0,
    brief_versand BOOLEAN NOT NULL DEFAULT 0,
    strasse TEXT,
    plz TEXT,
    ort TEXT,
    landescode TEXT NOT NULL DEFAULT 'DE',
    telefon TEXT,
    fax TEXT,
    email_allgemein TEXT,
    email_rechnung TEXT,
    kontaktperson TEXT,
    iban TEXT,
    bic TEXT,
    ust_id_nr TEXT,
    steuernummer TEXT,
    leitweg_id TEXT,
    zahlungsziel_tage INTEGER NOT NULL DEFAULT 14,
    zahlungsart TEXT NOT NULL DEFAULT 'Ueberweisung',
    anmerkungen TEXT
);

-- Materialkatalog (Artikel)
CREATE TABLE IF NOT EXISTS materialarten (
    material_id INTEGER PRIMARY KEY AUTOINCREMENT,
    materialname TEXT NOT NULL UNIQUE,
    standard_nettopreis REAL NOT NULL DEFAULT 0.0,
    mwst_satz REAL NOT NULL DEFAULT 19.0,
    einheit TEXT NOT NULL DEFAULT 't'
);

-- Sonderpreislisten (Kundenspezifisch)
CREATE TABLE IF NOT EXISTS preislisten (
    kundennummer INTEGER NOT NULL,
    material_id INTEGER NOT NULL,
    sonder_nettopreis REAL NOT NULL,
    PRIMARY KEY (kundennummer, material_id),
    FOREIGN KEY (kundennummer) REFERENCES kunden(kundennummer) ON DELETE CASCADE,
    FOREIGN KEY (material_id) REFERENCES materialarten(material_id) ON DELETE CASCADE
);

-- Rechnungen (Archiv)
CREATE TABLE IF NOT EXISTS rechnungen (
    rechnung_id INTEGER PRIMARY KEY AUTOINCREMENT,
    rechnungsnummer TEXT NOT NULL UNIQUE,
    kundennummer INTEGER NOT NULL,
    rechnungsdatum DATE NOT NULL,
    faelligkeitsdatum DATE NOT NULL,
    leistungszeitraum_start DATE NOT NULL,
    leistungszeitraum_ende DATE NOT NULL,
    netto_summe REAL NOT NULL,
    steuer_summe REAL NOT NULL,
    brutto_summe REAL NOT NULL,
    zahlungsstatus TEXT NOT NULL DEFAULT 'Offen',
    zugferd_xml TEXT NOT NULL,
    pdf_dateipfad TEXT,
    FOREIGN KEY (kundennummer) REFERENCES kunden(kundennummer)
);

-- Rechnungspositionen
CREATE TABLE IF NOT EXISTS rechnungspositionen (
    rechnungsposition_id INTEGER PRIMARY KEY AUTOINCREMENT,
    rechnung_id INTEGER NOT NULL,
    material_id INTEGER NOT NULL,
    positionsnummer INTEGER NOT NULL,
    gesamtgewicht REAL NOT NULL,
    einheit TEXT NOT NULL,
    einzelpreis_netto REAL NOT NULL,
    gesamtpreis_netto REAL NOT NULL,
    mwst_satz REAL NOT NULL,
    steuerbetrag REAL NOT NULL,
    FOREIGN KEY (rechnung_id) REFERENCES rechnungen(rechnung_id) ON DELETE CASCADE,
    FOREIGN KEY (material_id) REFERENCES materialarten(material_id)
);

-- Wiegezettel
CREATE TABLE IF NOT EXISTS wiegezettel (
    wiegezettel_id INTEGER PRIMARY KEY, -- entspricht historischer Lfd-Nr
    kundennummer INTEGER NOT NULL,
    datum DATETIME NOT NULL,
    gewicht REAL NOT NULL,
    material_id INTEGER NOT NULL,
    anlieferungsort TEXT NOT NULL,
    anlieferer TEXT NOT NULL,
    referenz TEXT,
    rechnungsposition_id INTEGER,
    FOREIGN KEY (kundennummer) REFERENCES kunden(kundennummer),
    FOREIGN KEY (material_id) REFERENCES materialarten(material_id),
    FOREIGN KEY (anlieferungsort) REFERENCES anlieferungsorte(anlieferungs_ort),
    FOREIGN KEY (anlieferer) REFERENCES anlieferer(anlieferer_herkunft),
    FOREIGN KEY (rechnungsposition_id) REFERENCES rechnungspositionen(rechnungsposition_id) ON DELETE SET NULL
);

-- Indizes für performante Abfragen
CREATE INDEX IF NOT EXISTS idx_wiegezettel_kunde ON wiegezettel(kundennummer);
CREATE INDEX IF NOT EXISTS idx_wiegezettel_rechnungsposition ON wiegezettel(rechnungsposition_id);
CREATE INDEX IF NOT EXISTS idx_rechnungspositionen_rechnung ON rechnungspositionen(rechnung_id);
CREATE INDEX IF NOT EXISTS idx_rechnungen_kunde ON rechnungen(kundennummer);
