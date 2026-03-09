package config

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Scanner: ScannerConfig{
			CustomPatterns: nil,
			Allowlist:      nil,
		},
		NLP: NLPConfig{
			Enabled:      false,
			ModelDir:     "models",
			NERModel:     "ner.onnx",
			NERTokenizer: "tokenizer.json",
			MaxLength:    512,
			NumThreads:   4,
			MinScore:     0.5,
		},
		Context: ContextConfig{
			BoostFactor: 0.15,
			WindowSize:  50,
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}
}
