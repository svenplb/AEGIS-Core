package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/svenplb/aegis-core/internal/redactor"
	"github.com/svenplb/aegis-core/internal/restorer"
)

var testBinary string

func TestMain(m *testing.M) {
	// Build the binary for integration tests.
	dir, err := os.MkdirTemp("", "aegis-scan-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	bin := "aegis-scan"
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	testBinary = filepath.Join(dir, bin)
	cmd := exec.Command("go", "build", "-o", testBinary, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("failed to build test binary: " + err.Error())
	}

	os.Exit(m.Run())
}

func runBinary(args ...string) (string, int, error) {
	cmd := exec.Command(testBinary, args...)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
		err = nil // non-zero exit is expected, not an error
	} else if err != nil {
		return string(out), -1, err
	}
	return string(out), exitCode, nil
}

func runBinaryWithStdin(input string, args ...string) (string, int, error) {
	cmd := exec.Command(testBinary, args...)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
		err = nil
	} else if err != nil {
		return string(out), -1, err
	}
	return string(out), exitCode, nil
}

func samplesDir() string {
	// Find testdata relative to the repo root.
	// When running tests, the working directory is cmd/aegis-scan/.
	return filepath.Join("..", "..", "testdata", "samples")
}

func TestMedicalDE(t *testing.T) {
	out, code, err := runBinary("--file", filepath.Join(samplesDir(), "medical_de.txt"), "--json")
	if err != nil {
		t.Fatal(err)
	}
	if code != 1 {
		t.Errorf("exit code = %d, want 1 (findings detected)", code)
	}

	var result redactor.RedactResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}

	wantTypes := map[string]bool{
		"PERSON": false, "DATE": false, "PHONE": false,
		"EMAIL": false, "FINANCIAL": false, "ADDRESS": false, "IBAN": false,
	}
	for _, e := range result.Entities {
		wantTypes[e.Type] = true
	}
	for typ, found := range wantTypes {
		if !found {
			t.Errorf("expected entity type %s not found", typ)
		}
	}
}

func TestFinancialMixed(t *testing.T) {
	out, code, err := runBinary("--file", filepath.Join(samplesDir(), "financial_mixed.txt"), "--json")
	if err != nil {
		t.Fatal(err)
	}
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}

	var result redactor.RedactResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	wantTypes := map[string]bool{
		"EMAIL": false, "CREDIT_CARD": false, "IBAN": false,
		"FINANCIAL": false, "SECRET": false, "PHONE": false,
	}
	for _, e := range result.Entities {
		wantTypes[e.Type] = true
	}
	for typ, found := range wantTypes {
		if !found {
			t.Errorf("expected entity type %s not found", typ)
		}
	}
}

func TestClean(t *testing.T) {
	out, code, err := runBinary("--file", filepath.Join(samplesDir(), "clean.txt"), "--json")
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0 (no findings)", code)
	}

	var result redactor.RedactResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(result.Entities) != 0 {
		t.Errorf("expected 0 entities for clean text, got %d: %v", len(result.Entities), result.Entities)
	}
}

func TestMultilingual(t *testing.T) {
	out, code, err := runBinary("--file", filepath.Join(samplesDir(), "multilingual.txt"), "--json")
	if err != nil {
		t.Fatal(err)
	}
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}

	var result redactor.RedactResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	wantTypes := map[string]bool{
		"PERSON": false, "ADDRESS": false, "PHONE": false, "EMAIL": false,
	}
	for _, e := range result.Entities {
		wantTypes[e.Type] = true
	}
	for typ, found := range wantTypes {
		if !found {
			t.Errorf("expected entity type %s not found", typ)
		}
	}
}

func TestJSONOutputValid(t *testing.T) {
	out, _, err := runBinaryWithStdin("Herr Thomas Schmidt, +49 170 1234567", "--json")
	if err != nil {
		t.Fatal(err)
	}

	var result redactor.RedactResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", err, out)
	}

	if result.OriginalText == "" {
		t.Error("original_text is empty")
	}
	if result.SanitizedText == "" {
		t.Error("sanitized_text is empty")
	}
}

func TestStdinInput(t *testing.T) {
	out, code, err := runBinaryWithStdin("Frau Maria Müller, geboren am 01.01.2000", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}

	var result redactor.RedactResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(result.Entities) < 2 {
		t.Errorf("expected at least 2 entities (PERSON + DATE), got %d", len(result.Entities))
	}
}

func TestTextFlag(t *testing.T) {
	out, code, err := runBinary("--text", "Email me at test@example.com", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}

	var result redactor.RedactResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	found := false
	for _, e := range result.Entities {
		if e.Type == "EMAIL" {
			found = true
		}
	}
	if !found {
		t.Error("EMAIL entity not found")
	}
}

func TestRoundTrip(t *testing.T) {
	samples := []string{
		"medical_de.txt",
		"financial_mixed.txt",
		"multilingual.txt",
	}

	for _, sample := range samples {
		t.Run(sample, func(t *testing.T) {
			out, _, err := runBinary("--file", filepath.Join(samplesDir(), sample), "--json")
			if err != nil {
				t.Fatal(err)
			}

			var result redactor.RedactResult
			if err := json.Unmarshal([]byte(out), &result); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}

			// Restore: replace tokens back to originals.
			restored := restorer.Restore(result.SanitizedText, result.Mappings)

			if restored != result.OriginalText {
				t.Errorf("round-trip failed:\noriginal:  %q\nrestored:  %q", result.OriginalText, restored)
			}
		})
	}
}

func TestNoInputError(t *testing.T) {
	// Running without any input should produce exit code 2.
	cmd := exec.Command(testBinary)
	cmd.Stdin = nil // no stdin, not a pipe
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}
	_ = out

	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2 (error)", exitCode)
	}
}
