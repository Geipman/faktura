# Migration: Access DB Export

This directory contains the legacy MS Access database exports (schemas and data) to help import structures and analyze data rules for the new Faktura application.

## Directory Structure

* **`schema/`** ([schema/](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/docs/migration/schema/)): Store legacy database schemas and definitions here (e.g. `.xsd` schema files).
* **`data/`** ([data/](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/docs/migration/data/)): Store legacy database records and export data here (e.g. `.xml` data files).

## Intent
AI agents and developers will parse files in these directories to map database entities (Kunden, Wiegezettel, Rechnungen, etc.) and reconstruct historical billing calculations.
