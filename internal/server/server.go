package server

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/Geipman/faktura/internal/templates"
)

// Server represents our HTTP server and dependencies.
type Server struct {
	db   *sql.DB
	mux  *http.ServeMux
	addr string
}

// NewServer creates a new Server instance.
func NewServer(addr string, db *sql.DB) *Server {
	s := &Server{
		db:   db,
		mux:  http.NewServeMux(),
		addr: addr,
	}
	s.routes()
	return s
}

// routes configures the HTTP routes.
func (s *Server) routes() {
	// Serve static files
	fs := http.FileServer(http.Dir("internal/server/static"))
	s.mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Main dashboard page
	s.mux.HandleFunc("/", s.handleIndex)
}

// handleIndex renders the landing/dashboard page.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Create components
	indexComp := templates.Index()
	layoutComp := templates.Layout("Dashboard", indexComp)

	// Render layout
	if err := layoutComp.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering index page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// LoggingMiddleware logs HTTP requests.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// Start runs the HTTP server.
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.addr)

	// Apply logging middleware
	loggedHandler := LoggingMiddleware(s.mux)

	srv := &http.Server{
		Addr:         s.addr,
		Handler:      loggedHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return srv.ListenAndServe()
}
