package nlp

import (
	"path/filepath"
	"testing"
)

func loadTestTokenizer(t *testing.T) *Tokenizer {
	t.Helper()
	tok, err := LoadTokenizer(filepath.Join("testdata", "tokenizer.json"), 32)
	if err != nil {
		t.Fatalf("LoadTokenizer: %v", err)
	}
	return tok
}

func TestLoadTokenizer(t *testing.T) {
	tok := loadTestTokenizer(t)

	if tok.bosID != 0 {
		t.Errorf("bosID = %d, want 0", tok.bosID)
	}
	if tok.eosID != 2 {
		t.Errorf("eosID = %d, want 2", tok.eosID)
	}
	if tok.padID != 1 {
		t.Errorf("padID = %d, want 1", tok.padID)
	}
	if tok.unkID != 3 {
		t.Errorf("unkID = %d, want 3", tok.unkID)
	}
	if len(tok.vocab) == 0 {
		t.Fatal("vocab is empty")
	}
	if tok.replacer != "▁" {
		t.Errorf("replacer = %q, want %q", tok.replacer, "▁")
	}
}

func TestPreTokenize(t *testing.T) {
	tok := loadTestTokenizer(t)

	tests := []struct {
		input string
		words []string
		start []int // charStart per word
	}{
		{
			input: "Max Mueller",
			words: []string{"▁Max", "▁Mueller"},
			start: []int{0, 4},
		},
		{
			input: "  hello  world ",
			words: []string{"▁hello", "▁world"},
			start: []int{2, 9},
		},
		{
			input: "",
			words: nil,
			start: nil,
		},
		{
			input: "single",
			words: []string{"▁single"},
			start: []int{0},
		},
	}

	for _, tt := range tests {
		spans := tok.preTokenize(tt.input)

		if len(spans) != len(tt.words) {
			t.Errorf("preTokenize(%q): got %d spans, want %d", tt.input, len(spans), len(tt.words))
			continue
		}

		for i, span := range spans {
			if span.text != tt.words[i] {
				t.Errorf("preTokenize(%q)[%d].text = %q, want %q", tt.input, i, span.text, tt.words[i])
			}
			if span.charStart != tt.start[i] {
				t.Errorf("preTokenize(%q)[%d].charStart = %d, want %d", tt.input, i, span.charStart, tt.start[i])
			}
		}
	}
}

func TestEncodeProducesBOSAndEOS(t *testing.T) {
	tok := loadTestTokenizer(t)
	out := tok.Encode("Max")

	// First token must be BOS.
	if out.InputIDs[0] != int64(tok.bosID) {
		t.Errorf("InputIDs[0] = %d, want BOS=%d", out.InputIDs[0], tok.bosID)
	}

	// Find EOS position (last non-zero attention).
	eosPos := 0
	for i, m := range out.AttentionMask {
		if m == 1 {
			eosPos = i
		}
	}
	if out.InputIDs[eosPos] != int64(tok.eosID) {
		t.Errorf("InputIDs[%d] = %d, want EOS=%d", eosPos, out.InputIDs[eosPos], tok.eosID)
	}
}

func TestEncodePadding(t *testing.T) {
	tok := loadTestTokenizer(t)
	out := tok.Encode("Max")

	if len(out.InputIDs) != tok.maxLen {
		t.Errorf("len(InputIDs) = %d, want maxLen=%d", len(out.InputIDs), tok.maxLen)
	}
	if len(out.AttentionMask) != tok.maxLen {
		t.Errorf("len(AttentionMask) = %d, want maxLen=%d", len(out.AttentionMask), tok.maxLen)
	}

	// Padded positions should have attention_mask = 0.
	activeTokens := 0
	for _, m := range out.AttentionMask {
		if m == 1 {
			activeTokens++
		}
	}
	if activeTokens < 3 { // at least BOS + 1 token + EOS
		t.Errorf("activeTokens = %d, want >= 3", activeTokens)
	}
	if activeTokens >= tok.maxLen {
		t.Errorf("no padding applied, all %d tokens active", activeTokens)
	}
}

func TestEncodeOffsets(t *testing.T) {
	tok := loadTestTokenizer(t)
	out := tok.Encode("Max Mueller")

	// BOS offset should be {0,0}.
	if out.Offsets[0].Start != 0 || out.Offsets[0].End != 0 {
		t.Errorf("BOS offset = {%d,%d}, want {0,0}", out.Offsets[0].Start, out.Offsets[0].End)
	}

	// Content tokens should have non-trivial offsets.
	hasContent := false
	for i := 1; i < len(out.Offsets)-1; i++ {
		off := out.Offsets[i]
		if off.End > off.Start {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("no content token has Start < End")
	}
}

func TestEncodeUTF8(t *testing.T) {
	tok := loadTestTokenizer(t)

	// Test with multi-byte characters.
	// Even if they're not in vocab, tokenizer should not panic.
	out := tok.Encode("Müller Straße")

	if len(out.InputIDs) != tok.maxLen {
		t.Errorf("len(InputIDs) = %d, want maxLen=%d", len(out.InputIDs), tok.maxLen)
	}

	// Should have some active tokens.
	activeTokens := 0
	for _, m := range out.AttentionMask {
		if m == 1 {
			activeTokens++
		}
	}
	if activeTokens < 3 {
		t.Errorf("activeTokens = %d, want >= 3 for UTF-8 input", activeTokens)
	}
}

func TestTokenizeWordViterbi(t *testing.T) {
	tok := loadTestTokenizer(t)

	// "▁Max" should be tokenized as a single token since it's in the vocab.
	ids, offs := tok.tokenizeWord("▁Max", 0)
	if len(ids) == 0 {
		t.Fatal("tokenizeWord returned empty for ▁Max")
	}

	// The whole word "▁Max" is in the vocab, so it should be 1 token.
	// (Viterbi should find the single-token solution as optimal.)
	expected := tok.vocab["▁Max"]
	found := false
	for _, id := range ids {
		if id == expected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("tokenizeWord(▁Max) = %v, expected to contain vocab[▁Max]=%d", ids, expected)
	}

	// Check offset: should map back to the original text.
	if len(offs) > 0 && offs[0].End < offs[0].Start {
		t.Errorf("offset End (%d) < Start (%d)", offs[0].End, offs[0].Start)
	}
}

func TestCharToByteOffset(t *testing.T) {
	tests := []struct {
		text    string
		charIdx int
		want    int
	}{
		{"hello", 0, 0},
		{"hello", 5, 5},
		{"hello", 3, 3},
		{"Müller", 0, 0},  // M at byte 0
		{"Müller", 1, 1},  // ü at byte 1 (2 bytes wide)
		{"Müller", 2, 3},  // l at byte 3
		{"Müller", 6, 7},  // end: 6 chars = 7 bytes (ü=2 bytes, rest=1 each)
		{"日本語", 0, 0},    // 日 at byte 0
		{"日本語", 1, 3},    // 本 at byte 3
		{"日本語", 2, 6},    // 語 at byte 6
		{"日本語", 3, 9},    // end
	}

	for _, tt := range tests {
		got := CharToByteOffset(tt.text, tt.charIdx)
		if got != tt.want {
			t.Errorf("CharToByteOffset(%q, %d) = %d, want %d", tt.text, tt.charIdx, got, tt.want)
		}
	}
}

func TestDecodeTokenOffsets(t *testing.T) {
	text := "Müller in Berlin"
	results := []NERResult{
		{Label: "PER", Start: 0, End: 6, Text: "Müller"},   // char offsets
		{Label: "LOC", Start: 10, End: 16, Text: "Berlin"},  // char offsets
	}

	decoded := DecodeTokenOffsets(text, results)

	// "Müller" = 6 chars, 7 bytes (M=1 + ü=2 + l=1 + l=1 + e=1 + r=1).
	if decoded[0].Start != 0 {
		t.Errorf("decoded[0].Start = %d, want 0", decoded[0].Start)
	}
	if decoded[0].End != 7 {
		t.Errorf("decoded[0].End = %d, want 7", decoded[0].End)
	}

	// "Berlin" at char 10 = byte 11 (1 extra byte from ü).
	if decoded[1].Start != 11 {
		t.Errorf("decoded[1].Start = %d, want 11", decoded[1].Start)
	}
	if decoded[1].End != 17 { // "Berlin" = 6 bytes, starts at byte 11
		t.Errorf("decoded[1].End = %d, want 17", decoded[1].End)
	}
}

func TestBuildCharToByteMap(t *testing.T) {
	text := "aüb" // a(1 byte) + ü(2 bytes) + b(1 byte) = 4 bytes
	m := buildCharToByteMap(text)

	// 3 chars + 1 sentinel = 4 entries
	if len(m) != 4 {
		t.Fatalf("len(charToByte) = %d, want 4", len(m))
	}
	if m[0] != 0 {
		t.Errorf("m[0] = %d, want 0", m[0])
	}
	if m[1] != 1 { // ü starts at byte 1
		t.Errorf("m[1] = %d, want 1", m[1])
	}
	if m[2] != 3 { // b starts at byte 3
		t.Errorf("m[2] = %d, want 3", m[2])
	}
	if m[3] != 4 { // sentinel = total byte length
		t.Errorf("m[3] = %d, want 4", m[3])
	}
}
