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

		// Split entities that span newlines into separate parts.
		for _, part := range splitAtNewlines(text, start, end) {
			entities = append(entities, Entity{
				Start:    part.start,
				End:      part.end,
				Type:     entityType,
				Text:     text[part.start:part.end],
				Score:    r.Score,
				Detector: "ner",
			})
		}
	}

	return entities
}

// span is a byte range within text.
type span struct{ start, end int }

// splitAtNewlines splits a byte range at newline characters, returning
// trimmed non-empty segments. This prevents NER entities from merging
// text across line boundaries (e.g. a name on one line and an address
// on the next).
func splitAtNewlines(text string, start, end int) []span {
	s := text[start:end]
	var spans []span
	segStart := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' || s[i] == '\r' {
			if segStart < i {
				// Trim trailing spaces from segment.
				segEnd := i
				for segEnd > segStart && s[segEnd-1] == ' ' {
					segEnd--
				}
				if segStart < segEnd {
					spans = append(spans, span{start + segStart, start + segEnd})
				}
			}
			// Skip past \r\n or \n.
			if s[i] == '\r' && i+1 < len(s) && s[i+1] == '\n' {
				i++
			}
			segStart = i + 1
			// Trim leading spaces from next segment.
			for segStart < len(s) && s[segStart] == ' ' {
				segStart++
			}
		}
	}
	// Last segment.
	if segStart < len(s) {
		segEnd := len(s)
		for segEnd > segStart && s[segEnd-1] == ' ' {
			segEnd--
		}
		if segStart < segEnd {
			spans = append(spans, span{start + segStart, start + segEnd})
		}
	}
	if len(spans) == 0 {
		return nil
	}
	return spans
}
