//go:build nlp

package main

import (
	"log"
	"regexp"

	"github.com/svenplb/aegis-core/internal/config"
	"github.com/svenplb/aegis-core/internal/nlp"
	"github.com/svenplb/aegis-core/internal/scanner"
)

// initScanner creates a scanner with optional NER support.
// Returns the scanner and a cleanup function.
func initScanner(cfg *config.Config, allowlist []*regexp.Regexp) (*scanner.CompositeScanner, func()) {
	if !cfg.NLP.Enabled {
		return scanner.DefaultScanner(allowlist), func() {}
	}

	// Initialize ONNX Runtime.
	if err := nlp.InitRuntime(); err != nil {
		log.Printf("warning: ONNX Runtime init failed: %v, falling back to regex only", err)
		return scanner.DefaultScanner(allowlist), func() {}
	}

	// Create NER session.
	session, err := nlp.NewNERSession(
		cfg.NLP.ModelDir,
		cfg.NLP.NERModel,
		cfg.NLP.NERTokenizer,
		cfg.NLP.MaxLength,
		cfg.NLP.NumThreads,
	)
	if err != nil {
		log.Printf("warning: NER model load failed: %v, falling back to regex only", err)
		return scanner.DefaultScanner(allowlist), func() {
			nlp.ShutdownRuntime()
		}
	}

	log.Printf("NLP/NER enabled (model_dir=%s, model=%s)", cfg.NLP.ModelDir, cfg.NLP.NERModel)

	nerScanner := scanner.NewNERScanner(session, cfg.NLP.MinScore)
	sc := scanner.HybridScanner(allowlist, nerScanner)

	cleanup := func() {
		session.Close()
		nlp.ShutdownRuntime()
	}

	return sc, cleanup
}
