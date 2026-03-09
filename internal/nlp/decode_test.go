package nlp

import (
	"math"
	"testing"
)

// makeLogits builds a flat logits array where for each token position,
// the label at targetLabel[i] gets a high score and all others get a low score.
func makeLogits(targetLabels []int, numLabels int) []float32 {
	logits := make([]float32, len(targetLabels)*numLabels)
	for i, label := range targetLabels {
		offset := i * numLabels
		for j := 0; j < numLabels; j++ {
			logits[offset+j] = -10.0 // low default
		}
		logits[offset+label] = 10.0 // high score for target label
	}
	return logits
}

func TestDecodeBIO_SinglePerson(t *testing.T) {
	// Simulate: [BOS] [B-PER] [I-PER] [O] [EOS]
	// Label indices: O=0, B-PER=1, I-PER=2
	targets := []int{0, 1, 2, 0, 0}
	logits := makeLogits(targets, NumLabels)

	labels, scores := DecodeBIO(logits, 5, NumLabels)

	want := []string{"O", "B-PER", "I-PER", "O", "O"}
	for i, l := range labels {
		if l != want[i] {
			t.Errorf("labels[%d] = %q, want %q", i, l, want[i])
		}
	}

	// Each score should be high (close to 1.0) since we gave 10.0 logit.
	for i, s := range scores {
		if s < 0.9 {
			t.Errorf("scores[%d] = %.4f, want > 0.9", i, s)
		}
	}
}

func TestDecodeBIO_MultipleEntities(t *testing.T) {
	// [BOS] [B-PER] [I-PER] [O] [B-LOC] [O] [EOS]
	// B-LOC=5, I-LOC=6
	targets := []int{0, 1, 2, 0, 5, 0, 0}
	logits := makeLogits(targets, NumLabels)

	labels, _ := DecodeBIO(logits, 7, NumLabels)

	want := []string{"O", "B-PER", "I-PER", "O", "B-LOC", "O", "O"}
	for i, l := range labels {
		if l != want[i] {
			t.Errorf("labels[%d] = %q, want %q", i, l, want[i])
		}
	}
}

func TestMergeBIOSpans_SinglePerson(t *testing.T) {
	// Simulate tokenized "Max Mueller":
	// [BOS] [▁Max] [▁Mu] [eller] [O...] [EOS]
	labels := []string{"O", "B-PER", "I-PER", "I-PER", "O", "O"}
	scores := []float64{0.99, 0.95, 0.90, 0.88, 0.99, 0.99}
	offsets := []TokenOffset{
		{0, 0},   // BOS
		{0, 3},   // "Max"
		{4, 6},   // "Mu"
		{6, 11},  // "eller" (Mueller chars 4-10)
		{0, 0},   // padding
		{0, 0},   // EOS
	}
	text := "Max Mueller"

	results := MergeBIOSpans(labels, scores, offsets, text)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	r := results[0]
	if r.Label != "PER" {
		t.Errorf("Label = %q, want %q", r.Label, "PER")
	}
	if r.Start != 0 {
		t.Errorf("Start = %d, want 0", r.Start)
	}
	if r.End != 11 {
		t.Errorf("End = %d, want 11", r.End)
	}
	if r.Text != "Max Mueller" {
		t.Errorf("Text = %q, want %q", r.Text, "Max Mueller")
	}

	expectedScore := (0.95 + 0.90 + 0.88) / 3.0
	if math.Abs(r.Score-expectedScore) > 0.001 {
		t.Errorf("Score = %.4f, want %.4f", r.Score, expectedScore)
	}
}

func TestMergeBIOSpans_TwoEntities(t *testing.T) {
	// "Max wohnt in Berlin"
	// [BOS] [B-PER] [O] [O] [B-LOC] [EOS]
	labels := []string{"O", "B-PER", "O", "O", "B-LOC", "O"}
	scores := []float64{0.99, 0.95, 0.99, 0.99, 0.92, 0.99}
	offsets := []TokenOffset{
		{0, 0},   // BOS
		{0, 3},   // "Max"
		{4, 9},   // "wohnt"
		{10, 12}, // "in"
		{13, 19}, // "Berlin"
		{0, 0},   // EOS
	}
	text := "Max wohnt in Berlin"

	results := MergeBIOSpans(labels, scores, offsets, text)

	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}

	if results[0].Label != "PER" || results[0].Text != "Max" {
		t.Errorf("result[0] = {%q, %q}, want {PER, Max}", results[0].Label, results[0].Text)
	}
	if results[1].Label != "LOC" || results[1].Text != "Berlin" {
		t.Errorf("result[1] = {%q, %q}, want {LOC, Berlin}", results[1].Label, results[1].Text)
	}
}

func TestMergeBIOSpans_NoEntities(t *testing.T) {
	labels := []string{"O", "O", "O", "O"}
	scores := []float64{0.99, 0.99, 0.99, 0.99}
	offsets := []TokenOffset{{0, 0}, {0, 5}, {6, 11}, {0, 0}}
	text := "hello world"

	results := MergeBIOSpans(labels, scores, offsets, text)

	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}

func TestMergeBIOSpans_EntityAtEnd(t *testing.T) {
	// Entity at the last content token before EOS.
	// [BOS] [O] [B-ORG] [I-ORG] [EOS]
	labels := []string{"O", "O", "B-ORG", "I-ORG", "O"}
	scores := []float64{0.99, 0.99, 0.91, 0.89, 0.99}
	offsets := []TokenOffset{
		{0, 0},
		{0, 5},   // "hello"
		{6, 10},  // "Goog"
		{10, 12}, // "le"
		{0, 0},
	}
	text := "hello Google"

	results := MergeBIOSpans(labels, scores, offsets, text)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Label != "ORG" || results[0].Text != "Google" {
		t.Errorf("result = {%q, %q}, want {ORG, Google}", results[0].Label, results[0].Text)
	}
}

func TestMergeBIOSpans_IWithoutB(t *testing.T) {
	// I-PER without a preceding B-PER should be ignored.
	labels := []string{"O", "I-PER", "O", "O"}
	scores := []float64{0.99, 0.90, 0.99, 0.99}
	offsets := []TokenOffset{{0, 0}, {0, 3}, {4, 9}, {0, 0}}
	text := "Max wohnt"

	results := MergeBIOSpans(labels, scores, offsets, text)

	if len(results) != 0 {
		t.Errorf("got %d results, want 0 (I- without B- should be ignored)", len(results))
	}
}

func TestMergeBIOSpans_DifferentTypeIAfterB(t *testing.T) {
	// B-PER followed by I-ORG — should emit PER and start ORG.
	labels := []string{"O", "B-PER", "I-ORG", "O", "O"}
	scores := []float64{0.99, 0.90, 0.85, 0.99, 0.99}
	offsets := []TokenOffset{
		{0, 0},
		{0, 3},   // "Max"
		{4, 9},   // "GmbH"
		{10, 15}, // other
		{0, 0},
	}
	text := "Max GmbH other"

	results := MergeBIOSpans(labels, scores, offsets, text)

	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Label != "PER" {
		t.Errorf("results[0].Label = %q, want PER", results[0].Label)
	}
	if results[1].Label != "ORG" {
		t.Errorf("results[1].Label = %q, want ORG", results[1].Label)
	}
}

func TestSoftmaxScore(t *testing.T) {
	// With one dominant logit, softmax should return ~1.0.
	logits := []float32{-10.0, 10.0, -10.0}
	score := SoftmaxScore(logits, 1)
	if score < 0.99 {
		t.Errorf("SoftmaxScore with dominant logit = %.6f, want > 0.99", score)
	}

	// With equal logits, softmax should return 1/n.
	equal := []float32{1.0, 1.0, 1.0}
	score = SoftmaxScore(equal, 0)
	expected := 1.0 / 3.0
	if math.Abs(score-expected) > 0.001 {
		t.Errorf("SoftmaxScore with equal logits = %.6f, want ~%.6f", score, expected)
	}

	// Empty logits.
	score = SoftmaxScore([]float32{}, 0)
	if score != 0 {
		t.Errorf("SoftmaxScore with empty logits = %.6f, want 0", score)
	}
}

func TestDecodeBIOAndMerge_Integration(t *testing.T) {
	// Full pipeline: logits → DecodeBIO → MergeBIOSpans.
	// Simulate "Max Mueller wohnt in Berlin"
	// Tokens: [BOS] [▁Max] [▁Mu] [eller] [▁wohnt] [▁in] [▁Berlin] [EOS]
	// Labels:  O     B-PER  I-PER  I-PER   O        O     B-LOC     O

	seqLen := 8
	targets := []int{
		0, // BOS → O
		1, // ▁Max → B-PER
		2, // ▁Mu → I-PER
		2, // eller → I-PER
		0, // ▁wohnt → O
		0, // ▁in → O
		5, // ▁Berlin → B-LOC
		0, // EOS → O
	}
	logits := makeLogits(targets, NumLabels)

	labels, scores := DecodeBIO(logits, seqLen, NumLabels)

	offsets := []TokenOffset{
		{0, 0},   // BOS
		{0, 3},   // "Max"
		{4, 6},   // "Mu"
		{6, 11},  // "eller"
		{12, 17}, // "wohnt"
		{18, 20}, // "in"
		{21, 27}, // "Berlin"
		{0, 0},   // EOS
	}
	text := "Max Mueller wohnt in Berlin"

	results := MergeBIOSpans(labels, scores, offsets, text)

	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}

	// First entity: "Max Mueller"
	if results[0].Label != "PER" {
		t.Errorf("results[0].Label = %q, want PER", results[0].Label)
	}
	if results[0].Text != "Max Mueller" {
		t.Errorf("results[0].Text = %q, want %q", results[0].Text, "Max Mueller")
	}
	if results[0].Start != 0 || results[0].End != 11 {
		t.Errorf("results[0] span = [%d,%d), want [0,11)", results[0].Start, results[0].End)
	}

	// Second entity: "Berlin"
	if results[1].Label != "LOC" {
		t.Errorf("results[1].Label = %q, want LOC", results[1].Label)
	}
	if results[1].Text != "Berlin" {
		t.Errorf("results[1].Text = %q, want %q", results[1].Text, "Berlin")
	}
	if results[1].Start != 21 || results[1].End != 27 {
		t.Errorf("results[1] span = [%d,%d), want [21,27)", results[1].Start, results[1].End)
	}
}
