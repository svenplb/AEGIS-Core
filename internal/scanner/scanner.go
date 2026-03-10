package scanner

import (
	"regexp"
	"sort"

	"golang.org/x/text/unicode/norm"
)

// Scanner detects PII entities in text.
type Scanner interface {
	Scan(text string) []Entity
}

// RegexScanner wraps a single compiled regex for one entity type.
type RegexScanner struct {
	re         *regexp.Regexp
	entityType string
	score      float64
	// validate is an optional function that post-validates a match.
	// If non-nil, only matches where validate returns true are kept.
	validate func(match string) bool
	// contextValidate is an optional function that validates a match
	// against the surrounding text. Receives full text and byte offsets.
	contextValidate func(fullText string, start, end int) bool
	// extractGroup specifies which submatch group to use as the entity text.
	// 0 means the full match, 1+ means the corresponding capture group.
	extractGroup int
}

// RegexScannerOption configures a RegexScanner.
type RegexScannerOption func(*RegexScanner)

// WithValidator adds a post-match validation function.
func WithValidator(fn func(string) bool) RegexScannerOption {
	return func(rs *RegexScanner) { rs.validate = fn }
}

// WithContextValidator adds a context-aware validation function.
// It receives the full text and the byte offsets of the match.
func WithContextValidator(fn func(fullText string, start, end int) bool) RegexScannerOption {
	return func(rs *RegexScanner) { rs.contextValidate = fn }
}

// WithExtractGroup sets which submatch group to use as the entity.
func WithExtractGroup(group int) RegexScannerOption {
	return func(rs *RegexScanner) { rs.extractGroup = group }
}

// NewRegexScanner creates a scanner from a compiled regex.
func NewRegexScanner(re *regexp.Regexp, entityType string, score float64, opts ...RegexScannerOption) *RegexScanner {
	rs := &RegexScanner{re: re, entityType: entityType, score: score}
	for _, opt := range opts {
		opt(rs)
	}
	return rs
}

// Scan finds all matches in text and returns entities with byte offsets.
func (rs *RegexScanner) Scan(text string) []Entity {
	if rs.extractGroup > 0 {
		return rs.scanWithGroups(text)
	}

	indices := rs.re.FindAllStringIndex(text, -1)
	entities := make([]Entity, 0, len(indices))
	for _, loc := range indices {
		matched := text[loc[0]:loc[1]]
		if rs.validate != nil && !rs.validate(matched) {
			continue
		}
		if rs.contextValidate != nil && !rs.contextValidate(text, loc[0], loc[1]) {
			continue
		}
		entities = append(entities, Entity{
			Start:    loc[0],
			End:      loc[1],
			Type:     rs.entityType,
			Text:     matched,
			Score:    rs.score,
			Detector: "regex",
		})
	}
	return entities
}

func (rs *RegexScanner) scanWithGroups(text string) []Entity {
	matches := rs.re.FindAllStringSubmatchIndex(text, -1)
	entities := make([]Entity, 0, len(matches))
	for _, loc := range matches {
		g := rs.extractGroup
		if g*2+1 >= len(loc) || loc[g*2] < 0 {
			continue
		}
		start := loc[g*2]
		end := loc[g*2+1]
		matched := text[start:end]
		if rs.validate != nil && !rs.validate(matched) {
			continue
		}
		if rs.contextValidate != nil && !rs.contextValidate(text, start, end) {
			continue
		}
		entities = append(entities, Entity{
			Start:    start,
			End:      end,
			Type:     rs.entityType,
			Text:     matched,
			Score:    rs.score,
			Detector: "regex",
		})
	}
	return entities
}

// CompositeScanner runs multiple scanners and merges/deduplicates results.
type CompositeScanner struct {
	scanners  []Scanner
	allowlist []*regexp.Regexp
}

// NewCompositeScanner creates a scanner that runs all provided scanners.
func NewCompositeScanner(scanners []Scanner, allowlist []*regexp.Regexp) *CompositeScanner {
	return &CompositeScanner{scanners: scanners, allowlist: allowlist}
}

// Scan runs all child scanners, merges results, deduplicates overlapping
// entities (keeping the longer match), filters by allowlist, and sorts by Start.
func (cs *CompositeScanner) Scan(text string) []Entity {
	// NFC normalize before scanning.
	text = norm.NFC.String(text)

	var all []Entity
	for _, s := range cs.scanners {
		all = append(all, s.Scan(text)...)
	}

	// Sort by Start, then by length descending (longer match first),
	// then by score descending (higher confidence wins ties).
	sort.Slice(all, func(i, j int) bool {
		if all[i].Start != all[j].Start {
			return all[i].Start < all[j].Start
		}
		li := all[i].End - all[i].Start
		lj := all[j].End - all[j].Start
		if li != lj {
			return li > lj
		}
		return all[i].Score > all[j].Score
	})

	// Deduplicate: keep longer match when overlapping.
	deduped := make([]Entity, 0, len(all))
	lastEnd := -1
	for _, e := range all {
		if e.Start < lastEnd {
			// Overlaps with a previous (longer or equal) entity — skip.
			continue
		}
		deduped = append(deduped, e)
		if e.End > lastEnd {
			lastEnd = e.End
		}
	}

	// Allowlist filter: drop entities matching any allowlist pattern.
	if len(cs.allowlist) > 0 {
		filtered := make([]Entity, 0, len(deduped))
		for _, e := range deduped {
			allowed := false
			for _, al := range cs.allowlist {
				if al.MatchString(e.Text) {
					allowed = true
					break
				}
			}
			if !allowed {
				filtered = append(filtered, e)
			}
		}
		return filtered
	}

	return deduped
}

// DefaultScanner returns a CompositeScanner with all built-in patterns.
func DefaultScanner(allowlist []*regexp.Regexp) *CompositeScanner {
	scanners := BuiltinScanners()
	return NewCompositeScanner(scanners, allowlist)
}
