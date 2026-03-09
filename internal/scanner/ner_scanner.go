//go:build nlp

package scanner

import (
	"log"

	"github.com/svenplb/aegis-core/internal/nlp"
)

// nerLabelToEntityType maps NER labels to aegis entity types.
var nerLabelToEntityType = map[string]string{
	"PER":  "PERSON",
	"LOC":  "ADDRESS",
	"ORG":  "ORG",
	"DIS":  "MEDICAL",
	"TIME": "DATE",
}

// NERScanner implements the Scanner interface using an ONNX NER model.
type NERScanner struct {
	session    *nlp.NERSession
	minScore   float64
	labelMap   map[string]string
}

// NewNERScanner wraps a NERSession as a Scanner.
// Only entities with score >= minScore are returned.
func NewNERScanner(session *nlp.NERSession, minScore float64) *NERScanner {
	if session == nil {
		return nil
	}
	return &NERScanner{
		session:  session,
		minScore: minScore,
		labelMap: nerLabelToEntityType,
	}
}

// Scan runs NER inference and returns detected entities with byte offsets.
func (ns *NERScanner) Scan(text string) []Entity {
	if ns == nil || ns.session == nil {
		return nil
	}

	results, err := ns.session.Predict(text)
	if err != nil {
		log.Printf("ner_scanner: inference error: %v", err)
		return nil
	}

	// Convert character offsets to byte offsets.
	results = nlp.DecodeTokenOffsets(text, results)

	entities := make([]Entity, 0, len(results))
	for _, r := range results {
		if r.Score < ns.minScore {
			continue
		}

		entityType, ok := ns.labelMap[r.Label]
		if !ok {
			continue // skip unmapped labels
		}

		// Clamp offsets to valid byte range.
		start := r.Start
		end := r.End
		if start < 0 {
			start = 0
		}
		if end > len(text) {
			end = len(text)
		}
		if start >= end {
			continue
		}

		entities = append(entities, Entity{
			Start:    start,
			End:      end,
			Type:     entityType,
			Text:     text[start:end],
			Score:    r.Score,
			Detector: "ner",
		})
	}

	return entities
}
