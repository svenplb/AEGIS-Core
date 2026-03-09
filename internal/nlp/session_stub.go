//go:build !nlp

package nlp

// NERSession is a stub when built without the nlp tag.
type NERSession struct{}

// NewNERSession returns nil when built without the nlp tag.
func NewNERSession(modelDir, nerModel, nerTokenizer string, maxLen, numThreads int) (*NERSession, error) {
	return nil, nil
}

// Predict is a no-op when built without the nlp tag.
func (s *NERSession) Predict(text string) ([]NERResult, error) {
	return nil, nil
}

// Close is a no-op when built without the nlp tag.
func (s *NERSession) Close() error {
	return nil
}
