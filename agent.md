# Agent.md - Welcome AI Assistant!

You are working on the **Faktura** project, a web-based invoicing application for a composting facility (Kompostierungsanlage), written in **Go** and using **SQLite**, **templ**, **HTMX**, and **TailwindCSS v4**.

This repository is designed for collaborative development between humans and different AI agents (e.g. Antigravity, OpenAI Codex).

## Project Overview

- **Business Domain**: Invoicing (Fakturierung) for products sold by weight.
- **Key Concepts**: Customer data (Kunden), Price lists (Preislisten), Weighing slips (Wiegezettel), and Invoice generation (Rechnungserstellung).
- **Core Standard**: Electronic invoices must conform to the German **ZUGFeRD** standard.
- **Architecture**: A single-binary Go application. Standard Go Project Layout with `cmd/` for entry points and `internal/` for private business logic, database operations, server endpoints, and HTML templates.

## Developer & Agent Guidelines

Before making any changes or suggesting plans, you **MUST** read and adhere to:
* **[Agent Rules](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/docs/guidelines/agent_rules.md)**: Coding standards, database patterns, testing requirements, and template security.
* **[Skills & Tooling](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/skills.md)**: Available development tasks and make targets.

## Repository Layout

```
.
├── .github/workflows/       # CI/CD pipelines
├── .vscode/                 # Editor/IDE configurations
├── assets/                  # Frontend assets (Tailwind input CSS, images, etc.)
│   └── css/
│       └── input.css        # Tailwind directive
├── cmd/
│   └── faktura/             # Main executable and entry point
├── docs/                    # Technical documentation & architecture records
│   ├── architecture/adr/    # Architecture Decision Records (ADRs)
│   ├── guidelines/          # Coding and style guidelines
│   └── requirements/        # Product specifications and feature lists
├── internal/                # Private application packages
│   ├── db/                  # SQLite connection and migration code
│   ├── server/              # HTTP routing, server setup, middleware
│   └── templates/           # templ HTML view components
├── Makefile                 # Task runner for development commands
├── go.mod                   # Go module definition
├── go.sum                   # Go dependencies checksums
├── agent.md                 # This file (AI orientation)
└── skills.md                # Development tasks and commands for KIs
```

## How to Proceed
1. If you are starting a new feature or task, review existing [Architecture Decision Records (ADRs)](file:////wsl.localhost/Ubuntu-24.04/home/geipman/projects/faktura/docs/architecture/adr/).
2. Create an implementation plan before writing complex code, and discuss it with the user.
3. Write high-coverage tests for any domain and invoicing logic.
4. Run `make lint` and `make test` before submitting changes or claiming completion.
