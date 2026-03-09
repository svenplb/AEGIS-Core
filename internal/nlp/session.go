//go:build nlp

package nlp

import (
	"fmt"
	"path/filepath"

	ort "github.com/yalue/onnxruntime_go"
)

// NERSession wraps an ONNX model for Named Entity Recognition.
type NERSession struct {
	tokenizer *Tokenizer
	session   *ort.DynamicAdvancedSession
	maxLen    int
}

// NewNERSession loads the ONNX NER model and tokenizer.
// modelDir is the directory containing model files.
// nerModel is the ONNX model filename (e.g. "ner.onnx").
// nerTokenizer is the tokenizer filename (e.g. "tokenizer.json").
func NewNERSession(modelDir, nerModel, nerTokenizer string, maxLen, numThreads int) (*NERSession, error) {
	tokPath := filepath.Join(modelDir, nerTokenizer)
	tok, err := LoadTokenizer(tokPath, maxLen)
	if err != nil {
		return nil, fmt.Errorf("nlp: load tokenizer: %w", err)
	}

	modelPath := filepath.Join(modelDir, nerModel)

	opts, err := ort.NewSessionOptions()
	if err != nil {
		return nil, fmt.Errorf("nlp: session options: %w", err)
	}
	defer opts.Destroy()

	if err := opts.SetIntraOpNumThreads(numThreads); err != nil {
		return nil, fmt.Errorf("nlp: set threads: %w", err)
	}

	inputNames := []string{"input_ids", "attention_mask"}
	outputNames := []string{"logits"}

	session, err := ort.NewDynamicAdvancedSession(modelPath, inputNames, outputNames, opts)
	if err != nil {
		return nil, fmt.Errorf("nlp: create session from %s: %w", modelPath, err)
	}

	return &NERSession{
		tokenizer: tok,
		session:   session,
		maxLen:    maxLen,
	}, nil
}

// Predict runs NER inference on the given text.
// Returns entities with character-based offsets.
func (s *NERSession) Predict(text string) ([]NERResult, error) {
	if text == "" {
		return nil, nil
	}

	// Tokenize.
	encoded := s.tokenizer.Encode(text)

	// Create input tensors.
	shape := ort.NewShape(1, int64(s.maxLen))

	inputIDs, err := ort.NewTensor(shape, encoded.InputIDs)
	if err != nil {
		return nil, fmt.Errorf("nlp: create input_ids tensor: %w", err)
	}
	defer inputIDs.Destroy()

	attentionMask, err := ort.NewTensor(shape, encoded.AttentionMask)
	if err != nil {
		return nil, fmt.Errorf("nlp: create attention_mask tensor: %w", err)
	}
	defer attentionMask.Destroy()

	// Create output tensor: shape [1, seq_len, num_labels].
	outputShape := ort.NewShape(1, int64(s.maxLen), int64(NumLabels))
	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		return nil, fmt.Errorf("nlp: create output tensor: %w", err)
	}
	defer outputTensor.Destroy()

	// Run inference.
	err = s.session.Run(
		[]ort.Value{inputIDs, attentionMask},
		[]ort.Value{outputTensor},
	)
	if err != nil {
		return nil, fmt.Errorf("nlp: inference: %w", err)
	}

	logits := outputTensor.GetData()

	// Count active tokens (non-padding).
	seqLen := 0
	for _, m := range encoded.AttentionMask {
		if m == 1 {
			seqLen++
		}
	}

	// Decode BIO labels.
	labels, scores := DecodeBIO(logits, seqLen, NumLabels)

	// Merge BIO spans into entities.
	results := MergeBIOSpans(labels, scores, encoded.Offsets, text)

	return results, nil
}

// Close releases the ONNX session resources.
func (s *NERSession) Close() error {
	if s.session != nil {
		return s.session.Destroy()
	}
	return nil
}
