package linter

import "github.com/larsartmann/go-finding"

// Filter selects findings by severity and category criteria. It is the
// finding-side counterpart to [FilterRules] (which selects rules by ID):
// use it when a consumer wants to gate on "error or above" or on a category
// subset without hand-rolling slices.
//
// The zero value keeps everything, so a partially configured Filter behaves
// as "no constraint" rather than "keep nothing":
//
//	gate := linter.Filter{MinSeverity: finding.SeverityError}
//	exit := linter.ExitCodeFromFindings(gate.Apply(findings))
//
// Apply never mutates its input and returns nil for nil input.
type Filter struct {
	// MinSeverity keeps findings at or above this severity. The zero value
	// (empty severity) keeps findings of every severity.
	MinSeverity finding.Severity

	// Categories, when non-empty, keeps only findings whose category is a
	// key mapped to true. Findings without a category are dropped while
	// this constraint is active.
	//
	// Type decision (recorded): a map rather than a slice because the
	// dominant caller shape is a set literal at the call site
	// (`Categories: map[Category]bool{linter.CategorySecurity: true}`),
	// membership checks are O(1) per finding, and callers sometimes want to
	// express "keep design but not style" without re-deriving a slice. The
	// trade-off is verbosity for consumers holding a []Category, who pay one
	// loop to build the map. A slice field would invert that cost onto every
	// Keep call, which runs per finding — the hotter path.
	Categories map[Category]bool
}

// Keep reports whether a single finding passes the filter. It is the
// predicate Apply applies per finding; use it directly for streaming or
// custom iteration.
func (f Filter) Keep(fd finding.Finding) bool {
	if f.MinSeverity != "" && fd.Severity.Compare(f.MinSeverity) < 0 {
		return false
	}

	if len(f.Categories) > 0 && !f.Categories[Category(fd.Category)] {
		return false
	}

	return true
}

// Apply returns the subset of findings that pass the filter, in input order.
// The input slice is never modified; nil yields nil.
func (f Filter) Apply(findings []finding.Finding) []finding.Finding {
	if findings == nil {
		return nil
	}

	out := make([]finding.Finding, 0, len(findings))

	for _, fd := range findings {
		if f.Keep(fd) {
			out = append(out, fd)
		}
	}

	return out
}
