# Proof of Concept (POC) Walkthrough: Faktura Invoicing

Wir haben die Proof-of-Concept-Anwendung (POC) für das Fakturierungssystem erfolgreich implementiert. Im Folgenden finden Sie eine detaillierte Zusammenfassung der erstellten Komponenten sowie eine Anleitung zur manuellen Verifizierung.

---

## 1. Erledigte Arbeiten

### Phase 1: Datenbank-Migration
- **Datenbankschema**: Ziel-Schema in [schema.sql](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/internal/db/schema.sql) mit allen Tabellen, Fremdschlüsseln und Indizes angelegt.
- **Migrationstool**: Ein Go-Kommandozeilen-Tool in [main.go](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/cmd/migrate/main.go) implementiert, das:
  - Die MS Access XML-Exporte einliest.
  - Kunden alphabetisch sortiert und 5-stellige Kundennummern ab `10001` vergibt.
  - Sonderpreislisten importiert und Duplikate (Preise, die dem Standardpreis entsprechen) herausfiltert.
  - Historische Wiegezettel mit den neuen Kunden- und Material-IDs verknüpft.
  - Eigene Stammdaten auf Basis der Beispielrechnung (Rechnungsaussteller) anlegt.
- **Makefile**: Befehl `make migrate` hinzugefügt, um die Migration schnell ausführen zu können.

### Phase 2: ZUGFeRD 2.2 XML-Generierung
- Die Go-Strukturen und die Serialisierungs-Logik für das offizielle **ZUGFeRD 2.2 BASIC** XML-Format in [zugferd.go](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/internal/zugferd/zugferd.go) implementiert.
- Die Korrektheit des XML-Strukturaufbaus über einen Unit-Test in [zugferd_test.go](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/internal/zugferd/zugferd_test.go) abgesichert.

### Phase 3: ZUGFeRD PDF/A-3 Generierung
- Visuelle Darstellung (A4-Rechnungsbogen) in [pdf.go](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/internal/pdf/pdf.go) via `gopdf` und DejaVuSans-Schriftarten umgesetzt.
- Einbettung der Rechnungs-XML (`factur-x.xml`) als Anhang in die PDF-Struktur mittels der `pdfcpu`-API realisiert. Dies erzeugt eine konforme ZUGFeRD PDF/A-3-Rechnung direkt in Go.
- Die PDF-Generierung sowie die XML-Einbettung über einen Unit-Test in [pdf_test.go](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/internal/pdf/pdf_test.go) abgesichert.

### Phase 4: Web-Interface & Aggregationslogik
- Die `templ`-Rechnungsansichten (Übersichtsliste, Erstellungsformular und DIN A4-Druckvorschau im Browser) in [invoices.templ](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/internal/templates/invoices.templ) erstellt.
- Routen-Handler und Abrechnungslogik in [server.go](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/internal/server/server.go) implementiert:
  - Aggregation: Filtert offene Wiegezettel für einen Zeitraum und fasst sie nach Material zusammen.
  - Preisfindung: Ermittelt, ob Sonderkonditionen existieren, andernfalls greift der Standard-Materialpreis.
  - Runden: Führt kaufmännische Rundung auf Positionsebene durch und trennt Steuersätze nach 7 % und 19 % (VAT Matrix).

---

## 2. Anleitung zur manuellen Verifizierung

### Schritt 1: Datenbank befüllen & Migration starten
Führen Sie den folgenden Befehl im Projektverzeichnis aus, um die Datenbank aus den XML-Dateien neu aufzubauen:
```bash
make migrate
```
*Erwartete Ausgabe:*
```
Running data migration...
Database schema successfully initialized.
Migrating Anlieferer...
Migrating Anlieferungsorte...
Migrating Materialarten...
Migrating Kunden...
Migrating Preislisten (Sonderpreise)...
Migrating Wiegezettel...
Initializing EigeneStammdaten...
Data migration completed successfully!
Migrated Customers: 536
Migrated Material types: 124
Migrated Price list overrides: 2462
Migrated Weighing slips: 694
```

### Schritt 2: Anwendung starten
Starten Sie den Go-Entwicklungsserver:
```bash
make run
```
*Erwartete Ausgabe:*
```
Starting development server with air...
Starting server on :8080...
```

### Schritt 3: Rechnung erstellen
1. Öffnen Sie Ihren Browser unter `http://localhost:8080/rechnungen`.
2. Klicken Sie oben rechts auf **Neue Rechnung erstellen** (oder gehen Sie direkt auf `/rechnungen/neu`).
3. Wählen Sie als Empfänger **AWS (10001)** aus.
4. Setzen Sie den Abrechnungszeitraum auf **01.01.2026** bis **31.01.2026** (um die Wiegezettel der Beispielrechnung zu erfassen).
5. Klicken Sie auf **Rechnung generieren**.

### Schritt 4: Rechnung & ZUGFeRD-Dokument prüfen
1. Sie werden auf die Liste der Rechnungen zurückgeleitet. Klicken Sie bei der neuen Rechnung (z. B. `RE-2026-0001`) auf **Details**.
2. Prüfen Sie die Vorschau:
   - Die Anschriften von Empfänger und Aussteller entsprechen dem Beleg.
   - Grünabfall wurde korrekt aufsummiert (9,04 t, 14,20 t, 6,09 t und 29,72 t ergeben die Positionen im Leistungszeitraum).
   - Der Preis von 19,00 € pro Tonne für Grünabfall (Sonderkondition) wurde korrekt angewendet.
   - Netto, Steuern (19 % und 7 %) und Bruttobetrag stimmen exakt mit der Beispielrechnung überein.
3. Über den Button **PDF/A-3** können Sie das fertige Rechnungs-PDF herunterladen und in einem PDF-Reader oder E-Rechnungsprüfer öffnen, um den ZUGFeRD-XML-Anhang zu betrachten.
4. Über den Button **XML** können Sie das rohe ZUGFeRD-XML herunterladen und prüfen.
