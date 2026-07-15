# Implementierungsplan: Proof of Concept Faktura

Dieser Plan beschreibt das Vorgehen zur Implementierung der Proof-of-Concept-Anwendung (POC) für die Fakturierung mit Go, SQLite, `templ`, `HTMX` und Tailwind CSS. Die Anwendung wird die exportierten XML-Altdaten migrieren, Wiegezettel nach Kunde und Material aggregieren, Rechnungen erstellen und ein ZUGFeRD-konformes PDF/A-3 Dokument erzeugen.

---

## 1. Wichtige Entscheidungen zur Freigabe

Bitte prüfen Sie die folgenden Punkte:
> [!IMPORTANT]
> **ZUGFeRD XML-Profil**: Wir implementieren das offizielle **ZUGFeRD 2.2 BASIC** Profil. Dieses ist voll kompatibel mit den gesetzlichen Vorgaben der Richtlinie EN 16931 für B2B-Rechnungen und bildet das Ziel-Datenmodell sauber ab.
>
> **PDF/A-3 Generierung & Einbettung**: 
> 1. Wir erzeugen das Rechnungs-PDF direkt im Go-Backend mithilfe einer schlanken, CGO-freien PDF-Bibliothek (`github.com/signintech/gopdf`).
> 2. Wir nutzen die integrierten Funktionen von `gopdf` zur Dateianhang-Einbettung, um die generierte `factur-x.xml` direkt in die PDF-Struktur einzubetten. Dies erzeugt eine vollkonforme ZUGFeRD PDF/A-3-Rechnung direkt in Go ohne externe Betriebssystem-Abhängigkeiten.
>
> **Migrations-Skript**: Wir erstellen ein Go-Kommandozeilen-Tool (`cmd/migrate/main.go`), das die XML-Dateien aus `docs/migration/data/` ausliest, verarbeitet und in die neue SQLite-Datenbank schreibt. Dies kann über `make migrate` ausgeführt werden.

---

## 2. Vorgeschlagene Änderungen

### Phase 1: Datenbank-Schema & Migration

* **[internal/db/schema.sql](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/internal/db/schema.sql) [NEU]**
  * SQL-Skript zur Tabellendefinition (deutsche Tabellen/Spalten, Fremdschlüssel und Indizes).
* **[cmd/migrate/main.go](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/cmd/migrate/main.go) [NEU]**
  * Das Migrationsprogramm:
    1. Liest XML-Dateien ein (`tblAnlieferer`, `tblAnlieferungOrt`, `tblMaterialarten`, `tblFirmenliste`, `tblPreisliste`, `tblWiegezettel`).
    2. Ermittelt Kundenkürzel, sortiert diese alphabetisch und vergibt 5-stellige Kundennummern ab `10001`.
    3. Importiert die Preislisten-Sonderkonditionen (Filtert identische Standardpreise heraus).
    4. Migriert alle Wiegezettel mit den neuen IDs.
    5. Legt einen Standard-Stammdatensatz (`EigeneStammdaten`) für das Kompostwerk an (unter Verwendung der Daten des Rechnungsausstellers "AWS" aus der Beispielrechnung).

### Phase 2: ZUGFeRD XML-Generierung

* **[internal/zugferd/zugferd.go](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/internal/zugferd/zugferd.go) [NEU]**
  * Definiert die Go-Strukturen für das ZUGFeRD 2.2 XML-Format (BASIC Profile).
  * Implementiert die Marshalling-Logik, die aus den Rechnungs- und Kundendaten ein standardkonformes XML (`factur-x.xml`) erzeugt.

### Phase 3: PDF-Generierung & XML-Einbettung

* **[internal/pdf/pdf.go](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/internal/pdf/pdf.go) [NEU]**
  * Generiert das Briefpapier A4-Rechnungsdokument mithilfe von `gopdf`, das dem Layout von `Beispielrechnung.pdf` entspricht.
  * Bettet das ZUGFeRD-XML als Anhang (`factur-x.xml`) mit dem PDF-Beziehungstyp `Alternative` ein und speichert es als PDF/A-3-kompatible Datei.

### Phase 4: Web-Interface & Aggregationslogik

* **[internal/templates/invoices.templ](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/internal/templates/invoices.templ) [NEU]**
  * `templ`-Views für:
    * Rechnungsübersicht (`/rechnungen`).
    * Rechnungs-Erstellungsformular (`/rechnungen/neu`) mit Kunden- und Zeitraumauswahl.
    * Eine HTML-Vorschau der Rechnung im exakten DIN A4-Layout.
* **[internal/server/server.go](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/internal/server/server.go) [MODIFIZIEREN]**
  * Neue Routen und Abrechnungs-Controller:
    * `GET /rechnungen`: Auflistung.
    * `GET /rechnungen/neu`: Zeigt Erstellungsformular.
    * `POST /rechnungen/neu`: Führt Aggregationslogik aus (offene Wiegezettel filtern, nach Material gruppieren, Preisfindung durchführen, Datensätze schreiben, XML/PDF erzeugen).
    * `GET /rechnungen/ansicht/{id}`: HTML-Druckansicht.
    * `GET /rechnungen/xml/{id}`: Download der XML-Rechnung.
    * `GET /rechnungen/pdf/{id}`: Download der PDF/A-3-Rechnung.
* **[internal/templates/layout.templ](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/internal/templates/layout.templ) [MODIFIZIEREN]**
  * Verknüpft den Navigationslink "Rechnungen" mit der Route `/rechnungen`.
* **[Makefile](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/Makefile) [MODIFIZIEREN]**
  * Ergänzt den Befehl `make migrate` zur schnellen Kompilierung und Ausführung des Migrationswerkzeugs.

---

## 3. Verifikationsplan

### Automatisierte Tests
* Ausführen von `go test ./...` zur Verifizierung der XML-Parsing-Logik und der ZUGFeRD-Generierung.

### Manuelle Verifikation
1. `make migrate` ausführen: Konvertiert die XML-Dateien und befüllt die lokale `faktura.db`.
2. `make run` ausführen, um den Entwicklungsserver zu starten.
3. Im Browser unter `http://localhost:8080/rechnungen/neu` eine Rechnung für den Kunden "AWS" im Zeitraum Januar 2026 erstellen.
4. Prüfen, ob die generierte Rechnung:
   * Kundennummer 10001 (oder passendes Mapping) aufweist.
   * Wiegezettel korrekt nach Material summiert.
   * Inhaltlich und visuell exakt der Beispielrechnung entspricht.
   * Button für XML- und PDF-Download bereitstellt.
5. Heruntergeladene PDF in Adobe Acrobat oder einem E-Rechnungsprüfer öffnen, um den ZUGFeRD-XML-Anhang zu validieren.
