# Agent Coding Guidelines

To maintain a professional, clean, and highly testable codebase, all AI agents (including Antigravity and Codex) and human developers must adhere to these rules.

---

## 1. General Go Principles

* **Error Handling**: 
  * Never discard errors using `_`.
  * Wrap errors to provide context when bubbles up: `fmt.Errorf("context message: %w", err)`.
  * For logging errors, use structured logging.
* **Concurrency**:
  * Only use goroutines when explicitly necessary (e.g., background jobs). Always handle panic recovery in goroutines.
* **Dependency Injection**:
  * Pass dependencies (like database connections, config, and clients) explicitly via constructors or function arguments. Do not use global variables for mutable state or database instances.
* **Naming Conventions**:
  * Follow standard Go styling (camelCase, package names as short, single-word, lowercase names without underscores).

---

## 2. Database (SQLite) Conventions

* **Driver**: Use `modernc.org/sqlite` (pure Go, CGO-free).
* **Connection Management**:
  * Always close rows (`defer rows.Close()`) immediately after checking query errors to avoid connection leaks.
  * Always use parameterized queries (`?` placeholder) to prevent SQL Injection.
* **Data Access Layer**:
  * Encapsulate SQLite queries inside a repository layer. Handlers should never write raw SQL queries directly.

---

## 3. Frontend & View (templ + HTMX + Tailwind)

* **Separation of Concerns**:
  * `.templ` files should contain only view structure and minimal display logic (loops, conditionals). Complex logic, calculations, and data fetching must live in the Go server/handler code.
* **HTMX Actions**:
  * Use standard HTMX attributes (e.g., `hx-get`, `hx-post`, `hx-target`, `hx-swap`) for interactivity.
  * Ensure every HTMX endpoint returns a valid HTML fragment (compiled via `templ`), NOT full pages.
* **Tailwind CSS**:
  * Do not write inline styles. Use Tailwind utility classes.
  * Avoid custom CSS files unless registering base fonts or Tailwind directives.
* **Security (CSRF & XSS)**:
  * `templ` automatically escapes variables to prevent XSS. Do not use `templ.Raw()` unless the input is fully sanitized and verified.
  * All modifying endpoints (POST, PUT, DELETE) must validate CSRF tokens.

---

## 4. Testing Requirements

* **Business Logic Target**:
  * **100% Code Coverage** is required for invoicing, price calculations, weighing slip calculations, and ZUGFeRD XML generation.
* **Structure**:
  * Use standard table-driven tests (`t.Run`) for complex input/output validation.
  * Place test files in the same directory as the code under test (e.g., `calculator_test.go` next to `calculator.go`).
  * Ensure tests are fast, isolation-friendly, and do not rely on local configuration files or external network endpoints. For database tests, use an in-memory SQLite database (`:memory:`).
