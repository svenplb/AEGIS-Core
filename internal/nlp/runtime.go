//go:build nlp

package nlp

import (
	"os"

	ort "github.com/yalue/onnxruntime_go"
)

// InitRuntime initializes the ONNX Runtime shared library.
// Must be called once before creating any NERSession.
// Set ONNX_LIB_PATH to override the default library search path.
func InitRuntime() error {
	if p := os.Getenv("ONNX_LIB_PATH"); p != "" {
		ort.SetSharedLibraryPath(p)
	}
	return ort.InitializeEnvironment()
}

// ShutdownRuntime releases ONNX Runtime resources.
// Call after all sessions are closed.
func ShutdownRuntime() error {
	return ort.DestroyEnvironment()
}
