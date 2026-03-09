package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// CustomPattern defines a user-supplied regex pattern for PII detection.
type CustomPattern struct {
	Name    string  `yaml:"name"`
	Type    string  `yaml:"type"`
	Pattern string  `yaml:"pattern"`
	Score   float64 `yaml:"score"`
}

// ScannerConfig holds scanner-related settings.
type ScannerConfig struct {
	CustomPatterns []CustomPattern `yaml:"custom_patterns"`
	Allowlist      []string        `yaml:"allowlist"`
}

// NLPConfig holds NLP/NER-related settings.
type NLPConfig struct {
	Enabled      bool    `yaml:"enabled"`
	ModelDir     string  `yaml:"model_dir"`
	NERModel     string  `yaml:"ner_model"`
	NERTokenizer string  `yaml:"ner_tokenizer"`
	MaxLength    int     `yaml:"max_length"`
	NumThreads   int     `yaml:"num_threads"`
	MinScore     float64 `yaml:"min_score"`
}

// ContextConfig holds context enhancement settings.
type ContextConfig struct {
	BoostFactor float64 `yaml:"boost_factor"`
	WindowSize  int     `yaml:"window_size"`
}

// LoggingConfig holds logging-related settings.
type LoggingConfig struct {
	Level string `yaml:"level"`
}

// Config is the top-level aegis-core configuration.
type Config struct {
	Scanner ScannerConfig `yaml:"scanner"`
	NLP     NLPConfig     `yaml:"nlp"`
	Context ContextConfig `yaml:"context"`
	Logging LoggingConfig `yaml:"logging"`
}

// validLogLevels enumerates accepted log level strings.
var validLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

// Load reads a YAML configuration file from path and returns a Config.
// Missing optional fields are filled from DefaultConfig.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that every custom pattern regex compiles and that the
// log level is recognised.
func (c *Config) Validate() error {
	for i, cp := range c.Scanner.CustomPatterns {
		if _, err := regexp.Compile(cp.Pattern); err != nil {
			return fmt.Errorf("config: custom_patterns[%d] (%s): invalid regex: %w", i, cp.Name, err)
		}
	}

	for i, pattern := range c.Scanner.Allowlist {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("config: allowlist[%d]: invalid regex: %w", i, err)
		}
	}

	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("config: unknown log level %q (want debug|info|warn|error)", c.Logging.Level)
	}

	if c.NLP.MaxLength < 0 {
		return fmt.Errorf("config: nlp.max_length must be non-negative")
	}
	if c.NLP.NumThreads < 0 {
		return fmt.Errorf("config: nlp.num_threads must be non-negative")
	}
	if c.Context.BoostFactor < 0 {
		return fmt.Errorf("config: context.boost_factor must be non-negative")
	}
	if c.Context.WindowSize < 0 {
		return fmt.Errorf("config: context.window_size must be non-negative")
	}

	return nil
}
