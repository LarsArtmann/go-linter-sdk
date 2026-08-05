package linter

import (
	"context"
	"errors"
	"sync"

	"github.com/larsartmann/go-finding"
)

// Registry holds a set of rules. A linter registers all its rules (typically
// in a rules.go via init() or a constructor) and the registry drives both
// standalone execution and BuildFlow integration.
type Registry struct {
	mu    sync.RWMutex
	rules []Rule
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		mu:    sync.RWMutex{},
		rules: []Rule{},
	}
}

// Register adds a rule. Panics if a rule with the same ID is already
// registered — duplicate IDs are a programming error that should surface at
// startup, not silently shadow at runtime. Panics if any required identity
// field (ID, Name, Description, Category) is empty.
func (r *Registry) Register(rule Rule) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := validateRuleIdentity(rule); err != nil {
		panic(err)
	}

	ruleID := rule.ID()
	for _, existing := range r.rules {
		if existing.ID() == ruleID {
			panic("linter: duplicate rule ID " + ruleID)
		}
	}

	r.rules = append(r.rules, rule)
}

// All returns every registered rule, in registration order.
func (r *Registry) All() []Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Rule, len(r.rules))
	copy(out, r.rules)

	return out
}

// Get returns the rule with the given ID and true, or nil and false if no rule
// with that ID is registered. Lookup is by the stable ID() (not the display
// Name()), matching Register's deduplication key.
func (r *Registry) Get(id string) (Rule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rule := range r.rules {
		if rule.ID() == id {
			return rule, true
		}
	}

	return nil, false
}

// Has reports whether a rule with the given ID is registered.
func (r *Registry) Has(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rule := range r.rules {
		if rule.ID() == id {
			return true
		}
	}

	return false
}

// Deregister removes the rule with the given ID. Returns true if a rule was
// removed, false if no rule with that ID was registered. Safe to call
// concurrently with Run — Run snapshots the rule list via All() before
// iterating.
func (r *Registry) Deregister(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, rule := range r.rules {
		if rule.ID() == id {
			r.rules = append(r.rules[:i], r.rules[i+1:]...)
			return true
		}
	}

	return false
}

// wrapRuleError ensures err is a *RuleError tagged with the rule ID. If err
// is already a *RuleError (e.g. produced by RuleFunc.Check), it passes through
// unchanged to avoid double-wrapping.
func wrapRuleError(ruleID string, err error) error {
	if _, ok := errors.AsType[*RuleError](err); ok {
		return err
	}

	return NewRuleError(ruleID, err)
}

// RunOption configures a Registry.Run invocation.
type RunOption func(*runConfig)

type runConfig struct {
	continueOnError bool
}

// ContinueOnError configures Run to continue executing remaining rules after
// one fails, collecting partial findings and joining all errors. The returned
// report contains findings from every rule that completed (including partial
// findings from failed rules); the returned error is the join of all rule
// failures, each individually wrapped as a *RuleError.
//
// Without this option (the default), Run fails fast: the first rule error
// aborts the run and returns (nil, err).
func ContinueOnError() RunOption {
	return func(c *runConfig) { c.continueOnError = true }
}

// Run executes every rule against dir, aggregating findings. This is the
// standalone execution path (the linter's own CLI). The BuildFlow integration
// path uses DetectorFromRegistry instead.
//
// By default Run fails fast: the first rule error aborts the run and returns
// (nil, err). Pass ContinueOnError() to run all rules regardless of
// individual failures, collecting partial findings and joining all errors:
//
//	report, err := registry.Run(ctx, dir, linter.ContinueOnError())
//
// In continue-on-error mode the returned report is always non-nil (it contains
// findings from every rule that succeeded) and err is the join of all failures
// (nil if every rule succeeded).
func (r *Registry) Run(ctx context.Context, dir string, opts ...RunOption) (*finding.Report, error) {
	cfg := runConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	toolName := finding.ToolName("linter")
	report := finding.NewReport(finding.ToolInfo{
		Name:    string(toolName),
		Version: "",
	})

	var errs []error

	for _, rule := range r.All() {
		findings, err := rule.Check(ctx, dir)
		if err != nil {
			wrapped := wrapRuleError(rule.ID(), err)
			if !cfg.continueOnError {
				return nil, wrapped
			}
			errs = append(errs, wrapped)
		}

		report.AddFindings(findings)
	}

	report.ComputeSummary()

	if len(errs) > 0 {
		return report, errors.Join(errs...)
	}

	return report, nil
}

// DetectorFromRegistry adapts a registry to the canonical finding.Detector
// interface, so a linter plugs into BuildFlow's DAG with zero glue (via the
// tool-sdk Spec.Detect field). The working directory is read from ctx via
// finding.WorkingDirFromContext.
//
// For go-finding/pipeline integration with per-rule parallelism, use
// DetectorsFromRegistry instead — it returns one detector per rule rather
// than collapsing all rules into a single opaque detector.
func DetectorFromRegistry(registry *Registry, toolName string) finding.Detector {
	return finding.NamedDetectorFunc(toolName, func(ctx context.Context) ([]finding.Finding, error) {
		dir := finding.WorkingDirFromContext(ctx)
		if dir == "" {
			dir = "."
		}

		var all []finding.Finding

		for _, rule := range registry.All() {
			findings, err := rule.Check(ctx, dir)
			if err != nil {
				return nil, wrapRuleError(rule.ID(), err)
			}

			all = append(all, findings...)
		}

		return all, nil
	})
}

// DetectorsFromRegistry returns one finding.Detector per registered rule,
// so a pipeline (go-finding/pipeline) can run them with per-detector
// parallelism, timeouts, and error isolation. Each detector is named after the
// rule's ID for pipeline metric attribution.
//
// Unlike DetectorFromRegistry, which collapses all rules into a single opaque
// detector, this function preserves per-rule granularity. The working directory
// is read from ctx via finding.WorkingDirFromContext.
//
// pipeline.Detector is a type alias for finding.Detector, so the returned slice
// plugs directly into pipeline.New(config, rootDir, detectors...).
func DetectorsFromRegistry(registry *Registry) []finding.Detector {
	all := registry.All()
	detectors := make([]finding.Detector, 0, len(all))

	for _, rule := range all {
		detectorName := rule.ID()

		detectors = append(detectors, finding.NamedDetectorFunc(
			detectorName,
			func(ctx context.Context) ([]finding.Finding, error) {
				dir := finding.WorkingDirFromContext(ctx)
				if dir == "" {
					dir = "."
				}

				findings, err := rule.Check(ctx, dir)
				if err != nil {
					return nil, wrapRuleError(rule.ID(), err)
				}

				return findings, nil
			},
		))
	}

	return detectors
}

// ExitCodeFromReport returns the process exit code for a lint run: 0 when
// there are no findings, 1 otherwise. Critical-severity findings could map to
// a different code, but the ecosystem convention is binary (clean / not clean).
func ExitCodeFromReport(report *finding.Report) int {
	if report == nil || report.Len() == 0 {
		return 0
	}

	return 1
}
