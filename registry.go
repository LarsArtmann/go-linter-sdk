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
// startup, not silently shadow at runtime. Panics if the rule's ID is empty.
func (r *Registry) Register(rule Rule) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := rule.ID()
	if id == "" {
		panic("linter: rule with empty ID cannot be registered")
	}

	for _, existing := range r.rules {
		if existing.ID() == id {
			panic("linter: duplicate rule ID " + id)
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

// wrapRuleError ensures err is a *RuleError tagged with the rule ID. If err
// is already a *RuleError (e.g. produced by RuleFunc.Check), it passes through
// unchanged to avoid double-wrapping.
func wrapRuleError(ruleID string, err error) error {
	if _, ok := errors.AsType[*RuleError](err); ok {
		return err
	}

	return NewRuleError(ruleID, err)
}

// Run executes every rule against dir, aggregating findings. This is the
// standalone execution path (the linter's own CLI). The BuildFlow integration
// path uses DetectorFromRegistry instead.
//
// If a rule fails, the returned error is a *RuleError identifying the rule.
func (r *Registry) Run(ctx context.Context, dir string) (*finding.Report, error) {
	toolName := finding.ToolName("linter")
	report := finding.NewReport(finding.ToolInfo{
		Name:    string(toolName),
		Version: "",
	})

	for _, rule := range r.All() {
		findings, err := rule.Check(ctx, dir)
		if err != nil {
			return nil, wrapRuleError(rule.ID(), err)
		}

		report.AddFindings(findings)
	}

	report.ComputeSummary()

	return report, nil
}

// DetectorFromRegistry adapts a registry to the canonical finding.Detector
// interface, so a linter plugs into BuildFlow's DAG with zero glue (via the
// tool-sdk Spec.Detect field). The working directory is read from ctx via
// finding.WorkingDirFromContext.
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
func DetectorsFromRegistry(registry *Registry, toolName string) []finding.Detector {
	all := registry.All()
	detectors := make([]finding.Detector, 0, len(all))

	for _, rule := range all {
		rule := rule // capture for closure

		detectors = append(detectors, finding.NamedDetectorFunc(rule.ID(), func(ctx context.Context) ([]finding.Finding, error) {
			dir := finding.WorkingDirFromContext(ctx)
			if dir == "" {
				dir = "."
			}

			findings, err := rule.Check(ctx, dir)
			if err != nil {
				return nil, wrapRuleError(rule.ID(), err)
			}

			return findings, nil
		}))
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
