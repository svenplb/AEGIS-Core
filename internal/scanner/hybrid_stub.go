//go:build !nlp

package scanner

import "regexp"

// HybridScanner falls back to DefaultScanner when built without the nlp tag.
// The ner parameter is ignored.
func HybridScanner(allowlist []*regexp.Regexp, ner *NERScanner) *CompositeScanner {
	return DefaultScanner(allowlist)
}
