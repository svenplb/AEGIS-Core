package scanner

import (
	"strings"
	"testing"
)

func TestEnhanceScores_BoostsWithKeyword(t *testing.T) {
	entities := []Entity{
		{Start: 5, End: 17, Type: "PERSON", Text: "Max Müller", Score: 0.6},
	}
	text := "Name Max Müller hier"

	result := EnhanceScores(entities, text, 0.15, 50)

	if result[0].Score <= 0.6 {
		t.Errorf("expected score boost, got %f", result[0].Score)
	}
	if result[0].Score > 1.0 {
		t.Errorf("score should not exceed 1.0, got %f", result[0].Score)
	}
}

func TestEnhanceScores_NoBoostWithoutKeyword(t *testing.T) {
	entities := []Entity{
		{Start: 0, End: 11, Type: "PERSON", Text: "Max Müller", Score: 0.6},
	}
	text := "Max Müller geht spazieren"

	result := EnhanceScores(entities, text, 0.15, 50)

	if result[0].Score != 0.6 {
		t.Errorf("expected no boost, got %f", result[0].Score)
	}
}

func TestEnhanceScores_CapsAt1(t *testing.T) {
	entities := []Entity{
		{Start: 5, End: 17, Type: "PERSON", Text: "Max Müller", Score: 0.95},
	}
	text := "Name Max Müller hier"

	result := EnhanceScores(entities, text, 0.15, 50)

	if result[0].Score != 1.0 {
		t.Errorf("score should cap at 1.0, got %f", result[0].Score)
	}
}

func TestEnhanceScores_MultipleEntityTypes(t *testing.T) {
	text := "Firma Acme Corp, Adresse Hauptstraße 1"
	entities := []Entity{
		{Start: 6, End: 15, Type: "ORG", Text: "Acme Corp", Score: 0.5},
		{Start: 25, End: len(text), Type: "ADDRESS", Text: "Hauptstraße 1", Score: 0.5},
	}

	result := EnhanceScores(entities, text, 0.2, 50)

	if result[0].Score <= 0.5 {
		t.Errorf("ORG should be boosted by 'Firma', got %f", result[0].Score)
	}
	if result[1].Score <= 0.5 {
		t.Errorf("ADDRESS should be boosted by 'Adresse', got %f", result[1].Score)
	}
}

func TestEnhanceScores_UnknownTypeIgnored(t *testing.T) {
	entities := []Entity{
		{Start: 0, End: 5, Type: "CREDIT_CARD", Text: "12345", Score: 0.7},
	}
	text := "12345 some text"

	result := EnhanceScores(entities, text, 0.15, 50)

	if result[0].Score != 0.7 {
		t.Errorf("unknown type should not be boosted, got %f", result[0].Score)
	}
}

func TestEnhanceScores_ZeroBoostFactor(t *testing.T) {
	entities := []Entity{
		{Start: 5, End: 17, Type: "PERSON", Text: "Max Müller", Score: 0.6},
	}
	text := "Name Max Müller hier"

	result := EnhanceScores(entities, text, 0, 50)
	if result[0].Score != 0.6 {
		t.Errorf("zero boost should not change score, got %f", result[0].Score)
	}
}

func TestEnhanceScores_ZeroWindowSize(t *testing.T) {
	entities := []Entity{
		{Start: 5, End: 17, Type: "PERSON", Text: "Max Müller", Score: 0.6},
	}
	text := "Name Max Müller hier"

	result := EnhanceScores(entities, text, 0.15, 0)
	if result[0].Score != 0.6 {
		t.Errorf("zero window should not change score, got %f", result[0].Score)
	}
}

func TestEnhanceScores_EmptyEntities(t *testing.T) {
	result := EnhanceScores(nil, "some text", 0.15, 50)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestEnhanceScores_WindowBoundary(t *testing.T) {
	// Keyword "name" is far away from the entity — outside window.
	padding := strings.Repeat("x", 100)
	text := "name " + padding + "Max Müller"
	start := 5 + len(padding)
	entities := []Entity{
		{Start: start, End: start + 11, Type: "PERSON", Text: "Max Müller", Score: 0.6},
	}

	result := EnhanceScores(entities, text, 0.15, 10)
	if result[0].Score != 0.6 {
		t.Errorf("keyword outside window should not boost, got %f", result[0].Score)
	}
}

func TestEnhanceScores_MultilingualKeywords(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"German", "Vorname Max Müller"},
		{"English", "customer Max Müller"},
		{"French", "monsieur Max Müller"},
		{"Spanish", "señor Max Müller"},
		{"Italian", "signor Max Müller"},
		{"Dutch", "meneer Max Müller"},
		{"Polish", "pan Max Müller"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Find "Max Müller" in the text.
			idx := strings.Index(tt.text, "Max")
			entities := []Entity{
				{Start: idx, End: idx + len("Max Müller"), Type: "PERSON", Text: "Max Müller", Score: 0.5},
			}
			result := EnhanceScores(entities, tt.text, 0.15, 50)
			if result[0].Score <= 0.5 {
				t.Errorf("%s keyword should boost score, got %f", tt.name, result[0].Score)
			}
		})
	}
}

func TestAlignToRuneStart(t *testing.T) {
	s := "Héllo" // é is 2 bytes (0xC3 0xA9)
	// byte 0: H, byte 1: 0xC3 (start of é), byte 2: 0xA9 (continuation)
	if got := alignToRuneStart(s, 2); got != 1 {
		t.Errorf("alignToRuneStart mid-rune = %d, want 1", got)
	}
	if got := alignToRuneStart(s, 0); got != 0 {
		t.Errorf("alignToRuneStart(0) = %d, want 0", got)
	}
	if got := alignToRuneStart(s, len(s)); got != len(s) {
		t.Errorf("alignToRuneStart(len) = %d, want %d", got, len(s))
	}
}

func TestAlignToRuneEnd(t *testing.T) {
	s := "Héllo"
	// byte 2 is continuation of é — should skip to byte 3
	if got := alignToRuneEnd(s, 2); got != 3 {
		t.Errorf("alignToRuneEnd mid-rune = %d, want 3", got)
	}
	if got := alignToRuneEnd(s, len(s)); got != len(s) {
		t.Errorf("alignToRuneEnd(len) = %d, want %d", got, len(s))
	}
}
