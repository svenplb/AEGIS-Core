//go:build !nlp

package scanner

// NERScanner is a stub when built without the nlp tag.
type NERScanner struct{}

// NewNERScanner returns nil when built without the nlp tag.
func NewNERScanner(_ any, _ float64) *NERScanner {
	return nil
}

// Scan is a no-op when built without the nlp tag.
func (ns *NERScanner) Scan(text string) []Entity {
	return nil
}
