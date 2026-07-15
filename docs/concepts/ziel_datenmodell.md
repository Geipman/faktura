# Konzeptdokument: Ziel-Datenmodell für Faktura

Dieses Dokument beschreibt das konzeptionelle Ziel-Datenmodell für das neue Go + SQLite-basierte Faktura-System. Es baut auf den Strukturen der alten Access-Datenbank auf, führt jedoch wesentliche Verbesserungen bezüglich Standardisierung, Rechnungsarchivierung und ZUGFeRD-Konformität ein.

Alle Bezeichnungen (Tabellen und Spalten) wurden vollständig ins Deutsche übersetzt.

---

## 1. Übersicht der Tabellen im Ziel-Datenmodell

Das Modell besteht aus den folgenden Tabellen:

1. **`Kunde`** (Stammdaten der Rechnungsempfänger, erweitert um ZUGFeRD-Felder)
2. **`Materialart`** (Produktkatalog inkl. Standardpreisen und USt-Sätzen)
3. **`Preisliste`** (Sonderkonditionen pro Kunde und Materialart)
4. **`Wiegezettel`** (Buchungen der Waage mit Zuordnung zu Kunden und Rechnungen)
5. **`Rechnung`** (Kopfdaten der archivierten Rechnungen)
6. **`Rechnungsposition`** (Aggregierte Abrechnungszeilen einer Rechnung)
7. **`EigeneStammdaten`** (Unternehmensprofil des Rechnungsausstellers für ZUGFeRD)
8. **`Anlieferer`** (Lookup-Tabelle für Herkunft/Überbringer)
9. **`Anlieferungsort`** (Lookup-Tabelle für Übergabeorte)

---

## 2. Detailbeschreibung der Tabellen & Felder

### 2.1 EigeneStammdaten (Neu)
Speichert die Daten unseres eigenen Unternehmens (Rechnungsaussteller). Diese Daten werden für den ZUGFeRD-XML-Header und den PDF-Briefkopf benötigt. Es gibt in dieser Tabelle genau einen Datensatz.

| Feldname | Datentyp | Nullable | Beschreibung |
| :--- | :--- | :--- | :--- |
| `Id` | `INTEGER` | Nein | Primärschlüssel (Wert: 1). |
| `Firmenname` | `TEXT(100)` | Nein | Name des Kompostwerks. |
| `Inhaber` | `TEXT(100)` | Nein | Vor- und Nachname des Inhabers. |
| `Strasse` | `TEXT(100)` | Nein | Straße und Hausnummer. |
| `Plz` | `TEXT(10)` | Nein | Postleitzahl. |
| `Ort` | `TEXT(50)` | Nein | Ort. |
| `Landescode` | `TEXT(2)` | Nein | Standard: "DE". |
| `Telefon` | `TEXT(30)` | Ja | Telefonnummer. |
| `Email` | `TEXT(100)` | Nein | Allgemeine E-Mail-Adresse. |
| `Webseite` | `TEXT(100)` | Ja | Homepage-URL. |
| `UStIdNr` | `TEXT(14)` | Nein | Umsatzsteuer-Identifikationsnummer (z. B. "DE123456789"). |
| `Steuernummer` | `TEXT(30)` | Nein | Steuernummer beim Finanzamt. |
| `Bankname` | `TEXT(100)` | Nein | Name des Kreditinstituts. |
| `Iban` | `TEXT(34)` | Nein | Eigene IBAN. |
| `Bic` | `TEXT(11)` | Nein | Eigene BIC. |
| `Handelsregister` | `TEXT(50)` | Ja | Registernummer (falls vorhanden). |
| `Amtsgericht` | `TEXT(50)` | Ja | Ort des zuständigen Registergerichts. |

---

### 2.2 Kunde
Enthält alle Rechnungsempfänger. Das bisherige `Kundenkürzel` wird durch eine numerische, 5-stellige `Kundennummer` ersetzt (Nummernkreis startet bei 10001).

| Feldname | Datentyp | Nullable | Beschreibung |
| :--- | :--- | :--- | :--- |
| `Kundennummer` | `INTEGER` | Nein | Primärschlüssel. 5-stellig (ab 10001). |
| `Kundenname` | `TEXT(100)` | Nein | Hauptname (Firma / Name). |
| `Namenserweiterung`| `TEXT(100)` | Ja | Zweite Namenszeile (z. B. "z. H. Herr Müller" oder GaLaBau-Abteilung). |
| `IstPrivat` | `BOOLEAN` | Nein | Default: `FALSE`. True = Privatperson, False = Firma. |
| `BriefVersand` | `BOOLEAN` | Nein | Default: `FALSE`. True = Rechnung per Post senden. |
| `Strasse` | `TEXT(100)` | Ja | Straße und Hausnummer. |
| `Plz` | `TEXT(10)` | Ja | Postleitzahl. |
| `Ort` | `TEXT(50)` | Ja | Ort. |
| `Landescode` | `TEXT(2)` | Nein | Default: "DE" (ISO 3166-1 alpha-2, wichtig für ZUGFeRD). |
| `Telefon` | `TEXT(30)` | Ja | Telefonnummer. |
| `Fax` | `TEXT(30)` | Ja | Faxnummer. |
| `EmailAllgemein` | `TEXT(100)` | Ja | Allgemeine E-Mail-Adresse. |
| `EmailRechnung` | `TEXT(255)` | Ja | E-Mail-Adresse speziell für den E-Rechnungsversand. |
| `Kontaktperson` | `TEXT(100)` | Ja | Ansprechpartner. |
| `Iban` | `TEXT(34)` | Ja | IBAN des Kunden (für Lastschrift / Rückbuchungen). |
| `Bic` | `TEXT(11)` | Ja | BIC des Kunden. |
| `UStIdNr` | `TEXT(14)` | Ja | Umsatzsteuer-Identifikationsnummer des Kunden (Pflicht bei B2B). |
| `Steuernummer` | `TEXT(30)` | Ja | Lokale Steuernummer des Kunden. |
| `LeitwegId` | `TEXT(50)` | Ja | Leitweg-ID des Kunden (wichtig für Behörden / XRechnung). |
| `ZahlungszielTage` | `INTEGER` | Nein | Default: 14. Fälligkeit in Tagen nach Rechnungsstellung. |
| `Zahlungsart` | `TEXT(30)` | Nein | Default: "Ueberweisung". Mögliche Werte: "Ueberweisung", "Lastschrift", "Bar". |
| `Anmerkungen` | `TEXT` | Ja | Notizen zum Kunden. |

---

### 2.3 Materialart
Der Produktkatalog des Kompostwerks.

| Feldname | Datentyp | Nullable | Beschreibung |
| :--- | :--- | :--- | :--- |
| `MaterialId` | `INTEGER` | Nein | Primärschlüssel (Auto-Increment). |
| `Materialname` | `TEXT(100)` | Nein | Name/Bezeichnung des Materials (z. B. "Grüngut"). |
| `StandardNettopreis` | `DECIMAL(10,2)`| Nein | Standard-Verkaufspreis pro Einheit (z. B. pro Tonne). |
| `MwstSatz` | `DECIMAL(5,2)` | Nein | Steuersatz in Prozent (z. B. 19.00 oder 7.00). |
| `Einheit` | `TEXT(10)` | Nein | Standard: "t" (weitere: "m3", "Stk"). |

---

### 2.4 Preisliste
Definiert kundenspezifische Sonderpreise, die den Standardpreis aus `Materialart` überschreiben.

| Feldname | Datentyp | Nullable | Beschreibung |
| :--- | :--- | :--- | :--- |
| `Kundennummer` | `INTEGER` | Nein | Primärschlüssel (Teil 1) & Foreign Key -> `Kunde.Kundennummer`. |
| `MaterialId` | `INTEGER` | Nein | Primärschlüssel (Teil 2) & Foreign Key -> `Materialart.MaterialId`. |
| `SonderNettopreis` | `DECIMAL(10,2)`| Nein | Kundenspezifischer Nettopreis pro Einheit. |

---

### 2.5 Wiegezettel
Repräsentiert die Wiegevorgänge. Um zu wissen, welche Wiegezettel bereits abgerechnet wurden (und über welche Rechnungsposition), erhält die Tabelle einen Fremdschlüssel auf die Rechnungsposition.

| Feldname | Datentyp | Nullable | Beschreibung |
| :--- | :--- | :--- | :--- |
| `WiegezettelId` | `INTEGER` | Nein | Primärschlüssel. Entspricht der historischen `Lfd-Nr` oder wird neu vergeben. |
| `Kundennummer` | `INTEGER` | Nein | Foreign Key -> `Kunde.Kundennummer`. |
| `Datum` | `DATETIME` | Nein | Datum und Uhrzeit des Wiegevorgangs. |
| `Gewicht` | `DECIMAL(10,3)`| Nein | Netto-Gewicht der Ladung in Tonnen (t). |
| `MaterialId` | `INTEGER` | Nein | Foreign Key -> `Materialart.MaterialId`. |
| `Anlieferungsort` | `TEXT(50)` | Nein | Übergabeort (z. B. "Biebesheim"). |
| `Anlieferer` | `TEXT(50)` | Nein | Wer hat angeliefert (z. B. "Kunde", "Dritte"). |
| `Referenz` | `TEXT(255)` | Ja | Kundenreferenz (z. B. Kennzeichen, Bestellnummer). |
| `RechnungspositionId` | `INTEGER` | Ja | Foreign Key -> `Rechnungsposition.RechnungspositionId`. Ist `NULL`, solange der Zettel noch nicht abgerechnet wurde. |

---

### 2.6 Rechnung (Neu)
Kopfdaten einer erstellten und archivierten Rechnung.

| Feldname | Datentyp | Nullable | Beschreibung |
| :--- | :--- | :--- | :--- |
| `RechnungId` | `INTEGER` | Nein | Primärschlüssel (Auto-Increment). |
| `Rechnungsnummer` | `TEXT(30)` | Nein | Eindeutige, fortlaufende Rechnungsnummer (z. B. "RE-2026-0001"). |
| `Kundennummer` | `INTEGER` | Nein | Foreign Key -> `Kunde.Kundennummer`. |
| `Rechnungsdatum` | `DATE` | Nein | Tag der Rechnungsstellung. |
| `Faelligkeitsdatum` | `DATE` | Nein | Errechnet aus Rechnungsdatum + `Kunde.ZahlungszielTage`. |
| `LeistungszeitraumStart` | `DATE` | Nein | Beginn des Abrechnungszeitraums. |
| `LeistungszeitraumEnde` | `DATE` | Nein | Ende des Abrechnungszeitraums. |
| `NettoSumme` | `DECIMAL(10,2)`| Nein | Summe aller Nettobeträge. |
| `SteuerSumme` | `DECIMAL(10,2)`| Nein | Gesamte ausgewiesene Umsatzsteuer. |
| `BruttoSumme` | `DECIMAL(10,2)`| Nein | Zahlbetrag (`NettoSumme` + `SteuerSumme`). |
| `Zahlungsstatus` | `TEXT(20)` | Nein | Default: "Offen". Werte: "Offen", "Bezahlt", "Storniert". |
| `ZugferdXml` | `TEXT` | Nein | Archivierter XML-Inhalt nach dem ZUGFeRD-Standard. |
| `PdfDateipfad` | `TEXT` | Ja | Pfad zur erstellten Rechnungs-PDF (inkl. eingebettetem ZUGFeRD-XML). |

---

### 2.7 Rechnungsposition (Neu)
Die Positionen einer Rechnung. Jede Position fasst alle Wiegezettel des **gleichen Materials** im Abrechnungszeitraum zusammen.

| Feldname | Datentyp | Nullable | Beschreibung |
| :--- | :--- | :--- | :--- |
| `RechnungspositionId`| `INTEGER` | Nein | Primärschlüssel (Auto-Increment). |
| `RechnungId` | `INTEGER` | Nein | Foreign Key -> `Rechnung.RechnungId` (Kaskadierendes Löschen). |
| `MaterialId` | `INTEGER` | Nein | Foreign Key -> `Materialart.MaterialId`. |
| `Positionsnummer` | `INTEGER` | Nein | Fortlaufende Nummer der Position auf der Rechnung (1, 2, 3...). |
| `Gesamtgewicht` | `DECIMAL(10,3)`| Nein | Summiertes Gewicht aller zugeordneten Wiegezettel. |
| `Einheit` | `TEXT(10)` | Nein | Einheit (Kopie aus `Materialart.Einheit`, meist "t"). |
| `EinzelpreisNetto` | `DECIMAL(10,2)`| Nein | Der angewendete Nettopreis (Sonderpreis oder Standardpreis). |
| `GesamtpreisNetto` | `DECIMAL(10,2)`| Nein | `Gesamtgewicht` $\times$ `EinzelpreisNetto` (kaufmännisch gerundet). |
| `MwstSatz` | `DECIMAL(5,2)` | Nein | Kopie des Steuersatzes zum Zeitpunkt der Rechnungsstellung. |
| `Steuerbetrag` | `DECIMAL(10,2)`| Nein | Berechneter Steuerbetrag für diese Position. |

---

### 2.8 Hilfstabellen: Anlieferer & Anlieferungsort
Einfache Lookup-Tabellen zur Datenkonsistenz an der Waage (wie im Altsystem).

* **`Anlieferer`**:
  * `AnliefererHerkunft` (`TEXT(50)`, Primärschlüssel) - z. B. "Kunde", "Dritte".
* **`Anlieferungsort`**:
  * `AnlieferungsOrt` (`TEXT(50)`, Primärschlüssel) - z. B. "Biebesheim".

---

## 3. Logik: Abrechnung & Aggregation

### 3.1 Aggregationsprozess (Wiegezettel -> Rechnungspositionen)
Wenn eine Rechnung für einen Kunden über einen bestimmten Zeitraum (z. B. 01.07.2026 bis 31.07.2026) erstellt wird, läuft folgende Logik ab:

1. **Selektion offener Wiegezettel**:
   Wähle alle Wiegezettel aus, bei denen gilt:
   * `Wiegezettel.Kundennummer == gewaehlterKunde`
   * `Wiegezettel.Datum` liegt im Abrechnungszeitraum.
   * `Wiegezettel.RechnungspositionId IS NULL` (Zettel wurde noch nicht abgerechnet).
2. **Gruppierung nach Material**:
   Gruppiere diese Wiegezettel nach der `MaterialId`.
3. **Erstellung der Rechnungspositionen**:
   Für jede Gruppe (gleiches Material):
   * Summiere das Gewicht: $\text{Gesamtgewicht} = \sum \text{Gewicht}$.
   * Ermittle den Preis:
     1. Suche in `Preisliste` nach `(Kundennummer, MaterialId)`.
     2. Falls vorhanden, nutze `Preisliste.SonderNettopreis` als `EinzelpreisNetto`.
     3. Falls nicht vorhanden, nutze `Materialart.StandardNettopreis` als `EinzelpreisNetto`.
   * Berechne: $\text{GesamtpreisNetto} = \text{Gesamtgewicht} \times \text{EinzelpreisNetto}$ (kaufmännisch auf 2 Nachkommastellen gerundet).
   * Berechne: $\text{Steuerbetrag} = \text{GesamtpreisNetto} \times (\text{Materialart.MwstSatz} / 100)$ (kaufmännisch gerundet).
   * Erzeuge einen Datensatz in `Rechnungsposition` und merke dir die generierte `RechnungspositionId`.
4. **Verknüpfung aktualisieren**:
   Aktualisiere bei allen in Schritt 3 einbezogenen Wiegezetteln das Feld `RechnungspositionId` mit der neu erzeugten ID. Damit sind die Zettel gesperrt und als "abgerechnet" markiert.

---

## 4. Migrationsregeln (Altdaten -> Zieldaten)

Da die Altdaten migriert werden müssen, legen wir folgende Regeln fest:

1. **Kundennummer-Zuweisung**:
   * Sortiere die alten Kunden aus `tblFirmenliste` alphabetisch nach `Kundenkürzel`.
   * Vergib fortlaufende Kundennummern beginnend mit `10001` (z. B. "HAUG" -> `10001`, "MUELLER" -> `10002` usw.).
   * Erstelle eine temporäre Mapping-Tabelle (`Kuerzel_Zu_Nummer`), um die Verknüpfungen in den Preislisten und Wiegezetteln während der Migration anzupassen.
2. **Standardisierung**:
   * `Privat_x0020__x003F_` wird zu `IstPrivat`.
   * `Brief_x003F_` wird zu `BriefVersand`.
   * Telefon- und Faxnummern-Masken aus Access werden bereinigt (nur Zahlen und standardisierte Trennzeichen speichern).
3. **Wiegezettel-Migration**:
   * Die Spalte `Lfd-Nr` wird als `WiegezettelId` übernommen.
   * `Kundenkürzel` wird über das Mapping in die entsprechende `Kundennummer` aufgelöst.
   * Da in den Altdaten keine Rechnungen existierten, bleibt `RechnungspositionId` für alle migrierten Wiegezettel initial `NULL` (können also theoretisch neu abgerechnet werden, oder wir markieren sie optional als "historisch abgerechnet").
4. **Bereinigung der Preisliste (Ansatz A - Vererbung)**:
   * Importiere aus der alten `tblPreisliste` nur diejenigen Einträge, bei denen der Netto-Sonderpreis des Kunden vom `StandardNettopreis` des jeweiligen Materials in `tblMaterialarten` abweicht.
   * Einträge mit identischen Preisen werden beim Import verworfen. Dadurch bleibt die neue Tabelle `Preisliste` minimal klein und übersichtlich. Wenn kein Eintrag existiert, greift das System automatisch auf den Standardpreis zurück.
