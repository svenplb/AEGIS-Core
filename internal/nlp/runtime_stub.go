//go:build !nlp

package nlp

// InitRuntime is a no-op when built without the nlp tag.
func InitRuntime() error { return nil }

// ShutdownRuntime is a no-op when built without the nlp tag.
func ShutdownRuntime() error { return nil }
