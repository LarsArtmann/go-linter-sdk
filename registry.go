package linter

import (
	"context"
	"errors"
	"slices"
	"sync"

	"github.com/larsartmann/go-finding"
)

// Registry holds a set of rules. A linter registers all its rules (typically
// in a rules.go via init() or a constructor) and the registry drives both
// standalone execution and BuildFlow integration.
type Registry struct {
	mu       sync.RWMutex
	rules    []Rule
	byID     map[string]Rule
	toolName finding.ToolName
}

// RegistryOption configures a Registry at construction time.
type RegistryOption func(*Registry)

// WithToolName sets the tool name that the Registry stamps onto findings and
// reports. When set, Register auto-fills [RuleMeta.ToolName] on every
// [RuleFunc] (and [OptIn] rule) that does not already specify one, and Run
// uses it for the [finding.ToolInfo] in the aggregated report.
//
// Without this option, the tool name defaults to "linter".
func WithToolName(name string) RegistryOption {
	return func(r *Registry) { r.toolName = finding.ToolName(name) }
}

// NewRegistry creates an empty registry. Pass [WithToolName] to set the tool
// name that flows into finding identity and report metadata:
//
//	r := linter.NewRegistry(linter.WithToolName("my-linter"))
func NewRegistry(opts ...RegistryOption) *Registry {
	r := &Registry{ //nolint:exhaustruct_v5 // toolName is set via options
		mu:    sync.RWMutex{},
		rules: []Rule{},
		byID:  map[string]Rule{},
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// NewRegistryFromRules constructs a registry from a pre-built rule slice,
// applying the same validation as Register (panics on duplicate IDs or empty
// identity fields) and the same option handling:
//
//	r := linter.NewRegistryFromRules(rules, linter.WithToolName("my-linter"))
//
// It is the bulk constructor for consumers that assemble rules before
// registry construction (e.g. after [FilterRules]); per-rule [Register] is
// equivalent, just noisier.
func NewRegistryFromRules(rules []Rule, opts ...RegistryOption) *Registry {
	r := NewRegistry(opts...)

	for _, rule := range rules {
		r.Register(rule)
	}

	return r
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
	if _, exists := r.byID[ruleID]; exists {
		panic("linter: duplicate rule ID " + ruleID)
	}

	// Auto-stamp the registry's tool name onto rules that do not set their own.
	if r.toolName != "" {
		switch concrete := rule.(type) {
		case RuleFunc:
			if concrete.Meta.ToolName == "" {
				concrete.Meta.ToolName = r.toolName
				rule = concrete
			}
		case optInRule:
			if concrete.Meta.ToolName == "" {
				concrete.Meta.ToolName = r.toolName
				rule = concrete
			}
		}
	}

	r.rules = append(r.rules, rule)
	r.byID[ruleID] = rule
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
// Name()), matching Register's deduplication key. O(1) via the ID index.
func (r *Registry) Get(id string) (Rule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rule, ok := r.byID[id]

	return rule, ok
}

// Has reports whether a rule with the given ID is registered. O(1) via the
// ID index — safe to call in hot loops.
func (r *Registry) Has(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.byID[id]

	return ok
}

// Deregister removes the rule with the given ID. Returns true if a rule was
// removed, false if no rule with that ID was registered. Safe to call
// concurrently with Run — Run takes a snapshot of the rule list via All()
// before iterating, so a rule in the snapshot still executes even if it is
// deregistered mid-run.
//
// The removal itself is O(n) in the rule count (the ordered rules slice must
// shift to preserve registration order); only Get and Has are O(1).
func (r *Registry) Deregister(ruleID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, rule := range r.rules {
		if rule.ID() == ruleID {
			r.rules = append(r.rules[:i], r.rules[i+1:]...)
			delete(r.byID, ruleID)

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
// failures, each individually wrapped as a *RuleError. Use RuleErrors(err) to
// enumerate every individual *RuleError from the joined result.
//
// Without this option (the default), Run fails fast: the first rule error
// aborts the run and returns (nil, err).
func ContinueOnError() RunOption {
	return func(c *runConfig) { c.continueOnError = true }
}

// Run executes every rule against dir, aggregating findings and coverage
// evidence. This is the standalone execution path (the linter's own CLI). The
// BuildFlow integration path uses DetectorFromRegistry instead.
//
// Rules implementing the optional [CoverageReporter] interface contribute
// their scan evidence (Summary.FilesScanned / Summary.SkippedModules) via
// go-finding's DetectResult channel; rules that do not contribute findings
// only, with coverage zeroed. Findings are normalized through
// finding.Relativize against dir, so finding IDs never embed host paths and
// are identical whether dir was absolute or relative.
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
	cfg := runConfig{continueOnError: false}
	for _, opt := range opts {
		opt(&cfg)
	}

	toolName := r.toolName
	if toolName == "" {
		toolName = "linter"
	}

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

		res := finding.DetectResult{
			Findings:       finding.Relativize(findings, dir),
			FilesScanned:   0,
			SkippedModules: nil,
		}
		if cov, ok := rule.(CoverageReporter); ok {
			res.FilesScanned, res.SkippedModules = cov.Coverage(ctx, dir)
		}

		report.AddDetectResult(res)
	}

	report.ComputeSummary()

	if len(errs) > 0 {
		return report, errors.Join(errs...)
	}

	return report, nil
}

// workingDir resolves the target directory from ctx, defaulting to "." when
// no working directory was supplied.
func workingDir(ctx context.Context) string {
	dir := finding.WorkingDirFromContext(ctx)
	if dir == "" {
		dir = "."
	}

	return dir
}

// DetectorFromRegistry adapts a registry to the canonical finding.Detector
// interface, so a linter plugs into BuildFlow's DAG with zero glue (via the
// Spec.Detect field of the go-finding/toolsdk module). The working directory
// is read from ctx via finding.WorkingDirFromContext.
//
// The returned value also implements finding.DetectorV2: hosts that use
// finding.RunDetect receive findings plus coverage evidence (rules
// implementing [CoverageReporter] contribute their scan counts), and findings
// are normalized through finding.Relativize against the working directory so
// IDs match the standalone Registry.Run lane. Plain Detect callers keep
// working untouched.
//
// For go-finding/pipeline integration with per-rule parallelism, use
// DetectorsFromRegistry instead — it returns one detector per rule rather
// than collapsing all rules into a single opaque detector.
func DetectorFromRegistry(registry *Registry, toolName string) finding.Detector {
	return &registryDetector{registry: registry, name: toolName}
}

// registryDetector adapts a whole Registry as one Detector/DetectorV2.
type registryDetector struct {
	registry *Registry
	name     string
}

func (d *registryDetector) Name() string { return d.name }

// Detect implements finding.Detector, returning the run's findings.
func (d *registryDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
	res, err := d.DetectV2(ctx)
	if err != nil {
		return nil, err
	}

	return res.Findings, nil
}

// DetectV2 implements finding.DetectorV2, returning findings plus the summed
// coverage evidence of every CoverageReporter rule. Detect and DetectV2 agree
// on the findings they return for the same run.
func (d *registryDetector) DetectV2(ctx context.Context) (finding.DetectResult, error) {
	dir := workingDir(ctx)

	var res finding.DetectResult

	for _, rule := range d.registry.All() {
		findings, err := rule.Check(ctx, dir)
		if err != nil {
			return finding.DetectResult{}, wrapRuleError(rule.ID(), err)
		}

		res.Findings = append(res.Findings, finding.Relativize(findings, dir)...)

		if cov, ok := rule.(CoverageReporter); ok {
			files, skipped := cov.Coverage(ctx, dir)
			res.FilesScanned += files

			for _, module := range skipped {
				if !slices.Contains(res.SkippedModules, module) {
					res.SkippedModules = append(res.SkippedModules, module)
				}
			}
		}
	}

	return res, nil
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
// Every returned detector also implements finding.DetectorV2 (coverage
// evidence from [CoverageReporter] rules, Relativize-normalized findings),
// matching DetectorFromRegistry's contract.
//
// pipeline.Detector is a type alias for finding.Detector, so the returned slice
// plugs directly into pipeline.New(config, rootDir, detectors...).
func DetectorsFromRegistry(registry *Registry) []finding.Detector {
	all := registry.All()
	detectors := make([]finding.Detector, 0, len(all))

	for _, rule := range all {
		detectors = append(detectors, &ruleDetector{rule: rule})
	}

	return detectors
}

// ruleDetector adapts a single Rule as one Detector/DetectorV2.
type ruleDetector struct {
	rule Rule
}

func (d *ruleDetector) Name() string { return d.rule.ID() }

// Detect implements finding.Detector, returning the rule's findings.
func (d *ruleDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
	res, err := d.DetectV2(ctx)
	if err != nil {
		return nil, err
	}

	return res.Findings, nil
}

// DetectV2 implements finding.DetectorV2 for a single rule: findings plus
// that rule's coverage evidence, if it implements [CoverageReporter].
func (d *ruleDetector) DetectV2(ctx context.Context) (finding.DetectResult, error) {
	dir := workingDir(ctx)

	findings, err := d.rule.Check(ctx, dir)
	if err != nil {
		return finding.DetectResult{}, wrapRuleError(d.rule.ID(), err)
	}

	res := finding.DetectResult{
		Findings:       finding.Relativize(findings, dir),
		FilesScanned:   0,
		SkippedModules: nil,
	}
	if cov, ok := d.rule.(CoverageReporter); ok {
		res.FilesScanned, res.SkippedModules = cov.Coverage(ctx, dir)
	}

	return res, nil
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

// ExitCodeFromFindings is [ExitCodeFromReport] for callers that hold a plain
// finding slice and never build a Report: 0 when findings is nil or empty,
// 1 otherwise. Same binary convention, no aggregation step.
func ExitCodeFromFindings(findings []finding.Finding) int {
	if len(findings) == 0 {
		return 0
	}

	return 1
}

// FilterRules returns the subset of rules that should run, given the enable
// and disable sets. When enable is non-empty, only those rules are included
// (minus any also disabled). When enable is empty, all rules except disabled
// ones are included. When both are empty, all rules are returned unchanged.
//
// This is the standard --enable/--disable filtering logic shared by CLI and
// plugin entry points:
//
//	rules := linter.FilterRules(humanizelint.AllRules(), enableSet, disableSet)
//	for _, rule := range rules {
//		registry.Register(rule)
//	}
func FilterRules(all []RuleFunc, enable, disable map[string]bool) []RuleFunc {
	if len(enable) == 0 && len(disable) == 0 {
		return all
	}

	var filtered []RuleFunc

	for _, rule := range all {
		if disable[rule.Meta.ID] {
			continue
		}

		if len(enable) > 0 && !enable[rule.Meta.ID] {
			continue
		}

		filtered = append(filtered, rule)
	}

	return filtered
}

// ExitCodeByConfidence returns a tiered exit code based on finding confidence:
//   - 0 when the report is nil or has no findings (clean).
//   - 1 when at least one finding is at or above threshold (must fix).
//   - 2 when findings exist but all are below threshold (triage).
//
// This lets CI distinguish "please review" from "must fix". For example,
// exit non-zero only for high-confidence findings:
//
//	code := linter.ExitCodeByConfidence(report, finding.ConfidenceHigh)
//	os.Exit(code)
func ExitCodeByConfidence(report *finding.Report, threshold finding.Confidence) int {
	if report == nil || report.Len() == 0 {
		return 0
	}

	for f := range report.All() {
		if f.Confidence.Compare(threshold) >= 0 {
			return 1
		}
	}

	return 2
}
