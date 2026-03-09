//go:build !nlp

package main

import (
	"regexp"

	"github.com/svenplb/aegis-core/internal/config"
	"github.com/svenplb/aegis-core/internal/scanner"
)

// initScanner creates a regex-only scanner when built without the nlp tag.
func initScanner(_ *config.Config, allowlist []*regexp.Regexp) (*scanner.CompositeScanner, func()) {
	return scanner.DefaultScanner(allowlist), func() {}
}
