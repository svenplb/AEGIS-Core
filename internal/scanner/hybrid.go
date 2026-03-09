//go:build nlp

package scanner

import "regexp"

// HybridScanner creates a CompositeScanner that combines built-in regex
// scanners with an optional NER scanner. If ner is nil, falls back to
// regex-only detection (same as DefaultScanner).
//
// After merging, NER-only LOC/ORG entities are filtered out unless a
// PERSON entity exists nearby, reducing false positives in news-like text.
func HybridScanner(allowlist []*regexp.Regexp, ner *NERScanner) *CompositeScanner {
	scanners := BuiltinScanners()
	if ner != nil {
		scanners = append(scanners, &contextFilteredNER{ner: ner})
	}
	return NewCompositeScanner(scanners, allowlist)
}

// contextFilteredNER wraps NERScanner and filters out standalone LOC/ORG
// entities that appear without any PERSON entity in the same text.
// A location or organization is only PII when associated with a person.
type contextFilteredNER struct {
	ner *NERScanner
}

func (cf *contextFilteredNER) Scan(text string) []Entity {
	entities := cf.ner.Scan(text)
	if len(entities) == 0 {
		return entities
	}

	// Check if any PERSON entity was found by NER.
	hasPerson := false
	for _, e := range entities {
		if e.Type == "PERSON" {
			hasPerson = true
			break
		}
	}

	// If no PERSON, drop NER-only LOC and ORG (they are likely not PII).
	if !hasPerson {
		filtered := make([]Entity, 0, len(entities))
		for _, e := range entities {
			if e.Type == "ADDRESS" || e.Type == "ORG" {
				continue
			}
			filtered = append(filtered, e)
		}
		return filtered
	}

	return entities
}
