package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/svenplb/aegis-core/internal/config"
	"github.com/svenplb/aegis-core/internal/redactor"
	"github.com/svenplb/aegis-core/internal/restorer"
	"github.com/svenplb/aegis-core/internal/scanner"
)

const version = "0.1.0"

// maxRequestBody is the maximum allowed request body size (1 MB).
const maxRequestBody int64 = 1 << 20

// scanRequest is the JSON shape for /api/scan and /api/redact.
type scanRequest struct {
	Text string `json:"text"`
}

// scanResponse is the JSON shape returned by /api/scan.
type scanResponse struct {
	Entities       []scanner.Entity `json:"entities"`
	ProcessingTime int64            `json:"processing_time_ms"`
}

// restoreRequest is the JSON shape for /api/restore.
type restoreRequest struct {
	Text     string            `json:"text"`
	Mappings []redactor.Mapping `json:"mappings"`
}

// restoreResponse is the JSON shape returned by /api/restore.
type restoreResponse struct {
	Text string `json:"text"`
}

// healthResponse is the JSON shape returned by /health.
type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// errorResponse is the JSON shape for error replies.
type errorResponse struct {
	Error string `json:"error"`
}

// corsMiddleware wraps a handler to add CORS headers and handle OPTIONS preflight.
func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigin := os.Getenv("AEGIS_CORS_ORIGINS")
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// writeJSON marshals v to JSON and writes it to w with the appropriate Content-Type.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// newMux creates the HTTP mux with all routes registered.
// Exported for use in tests.
func newMux(sc *scanner.CompositeScanner) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/scan", handleScan(sc))
	mux.HandleFunc("/api/redact", handleRedact(sc))
	mux.HandleFunc("/api/restore", handleRestore())

	return mux
}

// handleHealth returns the health check response.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, healthResponse{
		Status:  "ok",
		Version: version,
	})
}

// handleScan returns a handler that scans text for PII entities.
func handleScan(sc *scanner.CompositeScanner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)

		var req scanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Text == "" {
			writeError(w, http.StatusBadRequest, "text field is required")
			return
		}

		start := time.Now()
		entities := sc.Scan(req.Text)
		elapsed := time.Since(start).Milliseconds()

		writeJSON(w, http.StatusOK, scanResponse{
			Entities:       entities,
			ProcessingTime: elapsed,
		})
	}
}

// handleRedact returns a handler that scans and redacts text.
func handleRedact(sc *scanner.CompositeScanner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)

		var req scanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Text == "" {
			writeError(w, http.StatusBadRequest, "text field is required")
			return
		}

		entities := sc.Scan(req.Text)
		result := redactor.Redact(req.Text, entities)

		writeJSON(w, http.StatusOK, result)
	}
}

// handleRestore returns a handler that restores redacted tokens.
func handleRestore() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)

		var req restoreRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Text == "" {
			writeError(w, http.StatusBadRequest, "text field is required")
			return
		}

		restored := restorer.Restore(req.Text, req.Mappings)

		writeJSON(w, http.StatusOK, restoreResponse{Text: restored})
	}
}

func main() {
	portFlag := flag.Int("port", 0, "server port (default 9090, overrides AEGIS_SERVER_PORT)")
	configFlag := flag.String("config", "", "path to config.yaml (optional)")
	flag.Parse()

	// Determine port: flag > env > default.
	port := 9090
	if envPort := os.Getenv("AEGIS_SERVER_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}
	if *portFlag != 0 {
		port = *portFlag
	}

	// Load allowlist from config if provided.
	var allowlist []*regexp.Regexp
	if *configFlag != "" {
		cfg, err := config.Load(*configFlag)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
		for _, pattern := range cfg.Scanner.Allowlist {
			re, err := regexp.Compile(pattern)
			if err != nil {
				log.Fatalf("invalid allowlist pattern %q: %v", pattern, err)
			}
			allowlist = append(allowlist, re)
		}
	}

	// Create scanner once at startup (thread-safe for concurrent use).
	sc := scanner.DefaultScanner(allowlist)

	mux := newMux(sc)
	handler := corsMiddleware(mux)

	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("aegis-server %s starting on port %d", version, port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("server stopped")
}
