//go:build nlp

package scanner

import "regexp"

// HybridScanner creates a CompositeScanner that combines built-in regex
// scanners with an optional NER scanner. If ner is nil, falls back to
// regex-only detection (same as DefaultScanner).
func HybridScanner(allowlist []*regexp.Regexp, ner *NERScanner) *CompositeScanner {
	scanners := BuiltinScanners()
	if ner != nil {
		scanners = append(scanners, ner)
	}
	return NewCompositeScanner(scanners, allowlist)
}
