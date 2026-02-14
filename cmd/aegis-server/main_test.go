package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/svenplb/aegis-core/internal/scanner"
)

// newTestServer creates a test HTTP server with the full mux and CORS middleware.
func newTestServer() *httptest.Server {
	sc := scanner.DefaultScanner(nil)
	mux := newMux(sc)
	handler := corsMiddleware(mux)
	return httptest.NewServer(handler)
}

func TestHealthEndpoint(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body healthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", body.Status)
	}
	if body.Version != "0.1.0" {
		t.Errorf("expected version '0.1.0', got %q", body.Version)
	}
}

func TestScanEndpoint(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	payload := `{"text": "Contact Thomas at thomas@example.com"}`
	resp, err := http.Post(ts.URL+"/api/scan", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body scanResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(body.Entities) == 0 {
		t.Fatal("expected at least one entity, got none")
	}

	// Verify at least one EMAIL entity was found.
	foundEmail := false
	for _, e := range body.Entities {
		if e.Type == "EMAIL" {
			foundEmail = true
			break
		}
	}
	if !foundEmail {
		t.Error("expected an EMAIL entity in scan results")
	}
}

func TestRedactEndpoint(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	payload := `{"text": "Contact Thomas at thomas@example.com"}`
	resp, err := http.Post(ts.URL+"/api/redact", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	sanitized, ok := body["sanitized_text"].(string)
	if !ok {
		t.Fatal("expected sanitized_text field in response")
	}

	// The sanitized text should contain replacement tokens (brackets).
	if sanitized == "Contact Thomas at thomas@example.com" {
		t.Error("sanitized_text should differ from original when PII is present")
	}

	// Verify mappings are present.
	mappings, ok := body["mappings"].([]any)
	if !ok || len(mappings) == 0 {
		t.Error("expected non-empty mappings in redact response")
	}
}

func TestRestoreEndpoint(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	payload := `{
		"text": "Contact [PERSON_1] at [EMAIL_1]",
		"mappings": [
			{"token": "[PERSON_1]", "original": "Thomas", "type": "PERSON"},
			{"token": "[EMAIL_1]", "original": "thomas@example.com", "type": "EMAIL"}
		]
	}`
	resp, err := http.Post(ts.URL+"/api/restore", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body restoreResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expected := "Contact Thomas at thomas@example.com"
	if body.Text != expected {
		t.Errorf("expected %q, got %q", expected, body.Text)
	}
}

func TestScanMethodNotAllowed(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/scan")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", resp.StatusCode)
	}
}

func TestEmptyBody(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Test with empty text field.
	payload := `{"text": ""}`
	resp, err := http.Post(ts.URL+"/api/scan", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}

	var body errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if body.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestCORSHeaders(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	origin := resp.Header.Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin '*', got %q", origin)
	}

	methods := resp.Header.Get("Access-Control-Allow-Methods")
	if methods != "GET, POST, OPTIONS" {
		t.Errorf("expected Access-Control-Allow-Methods 'GET, POST, OPTIONS', got %q", methods)
	}

	headers := resp.Header.Get("Access-Control-Allow-Headers")
	if headers != "Content-Type" {
		t.Errorf("expected Access-Control-Allow-Headers 'Content-Type', got %q", headers)
	}
}

func TestCORSCustomOrigin(t *testing.T) {
	// Set custom origin via env var
	t.Setenv("AEGIS_CORS_ORIGINS", "https://app.example.com")

	ts := newTestServer()
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	origin := resp.Header.Get("Access-Control-Allow-Origin")
	if origin != "https://app.example.com" {
		t.Errorf("expected Access-Control-Allow-Origin 'https://app.example.com', got %q", origin)
	}
}

func TestOptionsPreflightRequest(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	req, err := http.NewRequest(http.MethodOptions, ts.URL+"/api/scan", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}

	origin := resp.Header.Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin '*', got %q", origin)
	}

	methods := resp.Header.Get("Access-Control-Allow-Methods")
	if methods != "GET, POST, OPTIONS" {
		t.Errorf("expected Access-Control-Allow-Methods 'GET, POST, OPTIONS', got %q", methods)
	}

	headers := resp.Header.Get("Access-Control-Allow-Headers")
	if headers != "Content-Type" {
		t.Errorf("expected Access-Control-Allow-Headers 'Content-Type', got %q", headers)
	}
}
