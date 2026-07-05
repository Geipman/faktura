# ADR 0001: Project Setup & Technology Stack

## Status
Accepted

## Context
The legacy invoicing system runs on MS Access, which cannot handle modern e-invoicing requirements (such as the German ZUGFeRD/Factur-X XML format). We need a new web-based invoicing application that is highly reliable, easily testable, and simple to run and deploy. We are two software architects evaluating collaborative AI workflows.

## Decision
We decided on the following technologies and project structures:

1. **Backend Language**: Go (v1.25+). It provides excellent performance, rapid compilation, built-in concurrency model, static type safety, and easily compiles into a single deployable binary.
2. **Project Layout**: Standard Go Project Layout. Separation of code into `cmd/faktura` (app entrance) and `internal/` (app-specific domain logic, views, server).
3. **Database**: SQLite (local database file) using the CGO-free `modernc.org/sqlite` driver. This removes external DB dependencies for development, and allows pure Go cross-compilation and easy integration tests using in-memory databases (`:memory:`).
4. **View Template Engine**: `templ` (github.com/a-h/templ). Provides type-safe component-based HTML generation, compile-time error checking, and integrated formatting.
5. **Frontend Interactivity**: `HTMX`. Allows us to build a dynamic single-page-like experience using standard HTML attributes without the complexity of a full JavaScript SPA build pipeline.
6. **Styling**: `TailwindCSS v4`. Compiled via the standalone Tailwind CLI, eliminating the need for Node.js and NPM in the workspace.
7. **Task Automation**: `Makefile` for running builds, code generation, and test execution. `air` for hot-reloading code and templates in development.

## Consequences
* **Single Binary Release**: The Go binary can embed all assets, templates, and static CSS, producing a single executable.
* **Portable Code & CI**: By using CGO-free SQLite and standalone Tailwind CLI, compilation does not depend on GCC, Node.js, or complex runtime setups.
* **Rapid Developer Loop**: Running `make run` auto-compiles changes (code, template, and CSS) and updates the local server in seconds.
* **Strong Type Safety**: The contract between HTML views and Go handlers is checked at compile-time by `templ`.
