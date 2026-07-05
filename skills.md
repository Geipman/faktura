# Skills.md - Commands & Automation

This document outlines the standard tools, tasks, and shell commands required to build, test, and run the **Faktura** application. All developers and AI agents should use these commands to maintain consistency.

## Available Makefile Commands

We use a `Makefile` to unify development workflows. Always prefer executing these commands over raw terminal scripts.

| Command | Description |
| :--- | :--- |
| `make install-tools` | Downloads Tailwind CSS CLI and installs required Go tools (`templ`, `air`, `golangci-lint`) |
| `make generate` | Compiles `.templ` files to Go source code |
| `make css` | Builds the static/production Tailwind CSS bundle |
| `make run` | Starts the hot-reloading development server via `air` (handles autobuilds of Go and templates) |
| `make build` | Generates templates, builds Tailwind CSS, and compiles the production Go binary |
| `make test` | Runs all unit and integration tests |
| `make test-cover` | Runs tests and generates a test coverage report |
| `make lint` | Runs `golangci-lint` to check code quality |
| `make clean` | Removes compiled binaries, temporary database files, and caches |

## Manual Executions

If you need to run specific tools directly, use these paths and patterns:

### Go Toolchain
Go is installed at `/usr/local/go/bin/go`.
- Run tests in a package:
  `wsl /usr/local/go/bin/go test ./internal/server`
- Format Go files:
  `wsl /usr/local/go/bin/go fmt ./...`

### Templ Compiler
`templ` is located at `/home/geipman/go/bin/templ`.
- Compile templates in watch-mode (rebuild on changes):
  `wsl /home/geipman/go/bin/templ generate --watch`

### Tailwind CSS CLI
Tailwind CSS v4 is compiled via the standalone binary located at `./bin/tailwindcss`.
- Compile Tailwind CSS manually:
  `wsl ./bin/tailwindcss -i ./assets/css/input.css -o ./internal/server/static/css/output.css --watch`

### Live Reloading (Air)
`air` is located at `/home/geipman/go/bin/air`.
- Run development server with live reload:
  `wsl /home/geipman/go/bin/air`

### Linting (golangci-lint)
`golangci-lint` is located at `/home/geipman/go/bin/golangci-lint`.
- Run the linter:
  `wsl /home/geipman/go/bin/golangci-lint run`
