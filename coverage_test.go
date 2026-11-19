package linter_test

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// coverageRule wraps a RuleFunc with scan-coverage evidence, implementing the
// optional linter.CoverageReporter interface. Embedding keeps all Rule
// methods; the registered value satisfies both interfaces.
type coverageRule struct {
	linter.RuleFunc

	files   int
	skipped []string
}

func (c coverageRule) Coverage(_ context.Context, _ string) (int, []string) {
	return c.files, c.skipped
}

func newCoverageRule(id string, files int, skipped []string, findings []finding.Finding) coverageRule {
	return coverageRule{
		RuleFunc: makeRule(id, findings),
		files:    files,
		skipped:  skipped,
	}
}

func absFinding(dir, rel string, line int) finding.Finding {
	rule := makeRule("abs-fixture", nil)

	return rule.NewFinding("abs fixture", finding.Pos(finding.FilePath(filepath.Join(dir, rel)), line, 1)).MustBuild()
}

func TestRegistry_Run_RecordsCoverage(t *testing.T) {
	t.Parallel()

	f := makeRule("plain", nil).NewFinding("plain issue", finding.Pos(finding.FilePath("a.go"), 1, 1)).MustBuild()

	r := linter.NewRegistry()
	r.Register(newCoverageRule("cov", 3, []string{"nested/mod"}, nil))
	r.Register(makeRule("plain", []finding.Finding{f}))

	report, err := r.Run(context.Background(), ".")
	if err != nil {
		t.Fatal(err)
	}

	if report.Summary.FilesScanned != 3 {
		t.Errorf("expected FilesScanned 3, got %d", report.Summary.FilesScanned)
	}

	if !slices.Equal(report.Summary.SkippedModules, []string{"nested/mod"}) {
		t.Errorf("expected SkippedModules [nested/mod], got %v", report.Summary.SkippedModules)
	}

	if report.Len() != 1 {
		t.Errorf("expected 1 finding, got %d", report.Len())
	}
}

func TestRegistry_Run_CoverageAccumulatesAcrossRules(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(newCoverageRule("cov-a", 2, nil, nil))
	r.Register(newCoverageRule("cov-b", 5, nil, nil))

	report, err := r.Run(context.Background(), ".")
	if err != nil {
		t.Fatal(err)
	}

	if report.Summary.FilesScanned != 7 {
		t.Errorf("expected FilesScanned 7 (2+5), got %d", report.Summary.FilesScanned)
	}
}

func TestRegistry_Run_PlainRulesLeaveCoverageZero(t *testing.T) {
	t.Parallel()

	f := makeRule("plain", nil).NewFinding("issue", finding.Pos(finding.FilePath("a.go"), 1, 1)).MustBuild()

	r := linter.NewRegistry()
	r.Register(makeRule("plain", []finding.Finding{f}))

	report, err := r.Run(context.Background(), ".")
	if err != nil {
		t.Fatal(err)
	}

	if report.Summary.FilesScanned != 0 {
		t.Errorf("expected honest zero FilesScanned, got %d", report.Summary.FilesScanned)
	}

	if report.Len() != 1 {
		t.Errorf("expected findings to survive, got %d", report.Len())
	}
}

func TestRegistry_Run_SkippedModulesUnioned(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(newCoverageRule("cov-a", 1, []string{"vendor-x"}, nil))
	r.Register(newCoverageRule("cov-b", 1, []string{"vendor-x", "vendor-y"}, nil))

	report, err := r.Run(context.Background(), ".")
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"vendor-x", "vendor-y"}
	if !slices.Equal(report.Summary.SkippedModules, want) {
		t.Errorf("expected SkippedModules %v, got %v", want, report.Summary.SkippedModules)
	}
}

func TestRegistry_Run_RelativizesAbsolutePaths(t *testing.T) {
	t.Parallel()

	dir := "/golsdk-test-root"
	f := absFinding(dir, "a/b.go", 3)

	r := linter.NewRegistry(linter.WithToolName("rel-lint"))
	r.Register(makeRule("rel-rule", []finding.Finding{f}))

	report, err := r.Run(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}

	got := ""
	for fd := range report.All() {
		got = string(fd.Position.File)
	}

	if got != "a/b.go" {
		t.Fatalf("expected relativized path a/b.go, got %q", got)
	}

	if strings.Contains(got, dir) {
		t.Errorf("expected no host path in finding, got %q", got)
	}
}

func TestRegistry_Run_RelativePathsPassThrough(t *testing.T) {
	t.Parallel()

	rule := makeRule("rel-rule", nil)
	f := rule.NewFinding("issue", finding.Pos(finding.FilePath("a/b.go"), 3, 1)).MustBuild()

	r := linter.NewRegistry()
	r.Register(makeRule("rel-rule", []finding.Finding{f}))

	report, err := r.Run(context.Background(), ".")
	if err != nil {
		t.Fatal(err)
	}

	for fd := range report.All() {
		if string(fd.Position.File) != "a/b.go" {
			t.Errorf("expected path unchanged, got %q", fd.Position.File)
		}
	}
}

func TestRelativize_DoesNotMutateInput(t *testing.T) {
	t.Parallel()

	rule := makeRule("copy-rule", nil)
	f := rule.NewFinding("issue", finding.Pos(finding.FilePath("/golsdk-test-root/a.go"), 1, 1)).MustBuild()
	in := []finding.Finding{f}

	_ = finding.Relativize(in, "/golsdk-test-root")

	if string(in[0].Position.File) != "/golsdk-test-root/a.go" {
		t.Errorf("input mutated: got %q", in[0].Position.File)
	}
}

func TestDetectorFromRegistry_DetectV2_CoverageAndRelativize(t *testing.T) {
	t.Parallel()

	dir := "/golsdk-test-root"
	f := absFinding(dir, "x.go", 2)

	r := linter.NewRegistry()
	r.Register(newCoverageRule("cov-a", 4, []string{"skipped-mod"}, nil))
	r.Register(newCoverageRule("cov-b", 6, []string{"skipped-mod"}, []finding.Finding{f}))

	detector := linter.DetectorFromRegistry(r, "agg-lint")

	v2, ok := detector.(finding.DetectorV2)
	if !ok {
		t.Fatal("expected DetectorFromRegistry result to implement finding.DetectorV2")
	}

	res, err := v2.DetectV2(finding.WithWorkingDir(context.Background(), dir))
	if err != nil {
		t.Fatal(err)
	}

	if res.FilesScanned != 10 {
		t.Errorf("expected FilesScanned 10 (4+6), got %d", res.FilesScanned)
	}

	if !slices.Equal(res.SkippedModules, []string{"skipped-mod"}) {
		t.Errorf("expected unioned SkippedModules, got %v", res.SkippedModules)
	}

	if len(res.Findings) != 1 || string(res.Findings[0].Position.File) != "x.go" {
		t.Errorf("expected relativized finding, got %+v", res.Findings)
	}
}

func TestDetectorFromRegistry_DetectMatchesDetectV2Findings(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(newCoverageRule("cov-a", 1, nil, nil))

	detector := linter.DetectorFromRegistry(r, "agg-lint")

	v2, ok := detector.(finding.DetectorV2)
	if !ok {
		t.Fatal("expected DetectorV2")
	}

	v2Res, err := v2.DetectV2(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	v1Findings, err := detector.Detect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(v1Findings) != len(v2Res.Findings) {
		t.Errorf("Detect and DetectV2 disagree: %d vs %d findings", len(v1Findings), len(v2Res.Findings))
	}
}

func TestDetectorsFromRegistry_DetectV2_PerRuleCoverage(t *testing.T) {
	t.Parallel()

	dir := "/golsdk-test-root"
	f := absFinding(dir, "y.go", 5)

	r := linter.NewRegistry()
	r.Register(newCoverageRule("cov-a", 4, []string{"mod-a"}, nil))
	r.Register(newCoverageRule("cov-b", 6, nil, []finding.Finding{f}))

	detectors := linter.DetectorsFromRegistry(r)
	if len(detectors) != 2 {
		t.Fatalf("expected 2 detectors, got %d", len(detectors))
	}

	byName := map[string]finding.DetectorV2{}

	for _, det := range detectors {
		v2, ok := det.(finding.DetectorV2)
		if !ok {
			t.Fatalf("detector %q does not implement DetectorV2", det.Name())
		}

		byName[det.Name()] = v2
	}

	resA, err := byName["cov-a"].DetectV2(finding.WithWorkingDir(context.Background(), dir))
	if err != nil {
		t.Fatal(err)
	}

	if resA.FilesScanned != 4 || !slices.Equal(resA.SkippedModules, []string{"mod-a"}) {
		t.Errorf("cov-a coverage wrong: %d %v", resA.FilesScanned, resA.SkippedModules)
	}

	resB, err := byName["cov-b"].DetectV2(finding.WithWorkingDir(context.Background(), dir))
	if err != nil {
		t.Fatal(err)
	}

	if resB.FilesScanned != 6 {
		t.Errorf("cov-b expected coverage 6, got %d", resB.FilesScanned)
	}

	if len(resB.Findings) != 1 || string(resB.Findings[0].Position.File) != "y.go" {
		t.Errorf("cov-b findings not relativized: %+v", resB.Findings)
	}
}

func TestRegistry_MalformedCategoryPanics(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()

	defer func() {
		rec := recover()
		if rec == nil {
			t.Fatal("expected panic on malformed Category")
		}

		err, ok := rec.(error)
		if !ok {
			t.Fatalf("panic value is not an error: %v", rec)
		}

		if !errors.Is(err, linter.ErrInvalidCategory) {
			t.Errorf("expected ErrInvalidCategory, got: %v", err)
		}
	}()

	r.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          "bad-cat",
			Name:        "bad-cat",
			Description: "bad cat",
			Cat:         linter.Category("Bogus Cat"),
		},
		Run: func(_ context.Context, _ string) ([]finding.Finding, error) { return nil, nil },
	})
}

func TestRuleMeta_Validate_InvalidCategory(t *testing.T) {
	t.Parallel()

	meta := linter.RuleMeta{ID: "x", Name: "X", Description: "desc", Cat: linter.Category("SOME_CATEGORY")}

	err := meta.Validate()
	if err == nil {
		t.Fatal("expected error for malformed category")
	}

	if !errors.Is(err, linter.ErrInvalidCategory) {
		t.Errorf("expected ErrInvalidCategory, got: %v", err)
	}
}

func TestRuleMeta_Validate_CustomCategoryAccepted(t *testing.T) {
	t.Parallel()

	for _, cat := range []linter.Category{linter.Category("api"), linter.Category("boilerplate"), linter.Category("design")} {
		meta := linter.RuleMeta{ID: "x", Name: "X", Description: "desc", Cat: cat}
		if err := meta.Validate(); err != nil {
			t.Errorf("expected custom category %q to validate, got: %v", cat, err)
		}
	}
}

func TestDetectorFromRegistry_DetectV2_ErrorWrapsRuleError(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(failingRule("boom"))

	detector := linter.DetectorFromRegistry(r, "agg-lint")

	v2, ok := detector.(finding.DetectorV2)
	if !ok {
		t.Fatal("expected DetectorV2")
	}

	_, err := v2.DetectV2(context.Background())
	if err == nil {
		t.Fatal("expected error from failing rule")
	}

	ruleErr, ok := errors.AsType[*linter.RuleError](err)
	if !ok {
		t.Fatalf("expected *RuleError, got %T: %v", err, err)
	}

	if ruleErr.RuleID != "boom" {
		t.Errorf("expected rule ID boom, got %q", ruleErr.RuleID)
	}

	if !errors.Is(err, linter.ErrRuleFailed) {
		t.Errorf("expected ErrRuleFailed sentinel, got: %v", err)
	}

	v1Findings, detectErr := detector.Detect(context.Background())
	if detectErr == nil {
		t.Fatal("expected Detect to propagate the same error")
	}

	if len(v1Findings) != 0 {
		t.Errorf("expected no findings on error, got %d", len(v1Findings))
	}
}

func TestDetectorsFromRegistry_DetectV2_ErrorWrapsRuleError(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(failingRule("boom-per-rule"))

	detectors := linter.DetectorsFromRegistry(r)
	if len(detectors) != 1 {
		t.Fatalf("expected 1 detector, got %d", len(detectors))
	}

	v2, ok := detectors[0].(finding.DetectorV2)
	if !ok {
		t.Fatal("expected DetectorV2")
	}

	_, err := v2.DetectV2(context.Background())
	if err == nil {
		t.Fatal("expected error from failing rule")
	}

	ruleErr, ok := errors.AsType[*linter.RuleError](err)
	if !ok {
		t.Fatalf("expected *RuleError, got %T: %v", err, err)
	}

	if ruleErr.RuleID != "boom-per-rule" {
		t.Errorf("expected rule ID boom-per-rule, got %q", ruleErr.RuleID)
	}
}

func TestRegistry_Run_ContinueOnError_KeepsCoverageFromPassingRules(t *testing.T) {
	t.Parallel()

	f := makeRule("cov-ok", nil).NewFinding("kept issue", finding.Pos(finding.FilePath("kept.go"), 1, 1)).MustBuild()

	r := linter.NewRegistry()
	r.Register(failingRule("boom"))
	r.Register(newCoverageRule("cov-ok", 7, []string{"mod-x"}, []finding.Finding{f}))

	report, err := r.Run(context.Background(), ".", linter.ContinueOnError())
	if err == nil {
		t.Fatal("expected joined error from the failing rule")
	}

	if !errors.Is(err, linter.ErrRuleFailed) {
		t.Errorf("expected ErrRuleFailed in joined error, got: %v", err)
	}

	if report == nil {
		t.Fatal("expected partial report despite the failure")
	}

	if report.Summary.FilesScanned != 7 {
		t.Errorf("expected coverage from the passing rule (7), got %d", report.Summary.FilesScanned)
	}

	if !slices.Equal(report.Summary.SkippedModules, []string{"mod-x"}) {
		t.Errorf("expected SkippedModules [mod-x], got %v", report.Summary.SkippedModules)
	}

	if report.Len() != 1 {
		t.Errorf("expected the passing rule's finding to survive, got %d findings", report.Len())
	}
}

// preWrappedRule is a custom Rule (not RuleFunc) whose Check returns an
// already-wrapped *RuleError, exercising wrapRuleError's pass-through branch.
type preWrappedRule struct{}

func (preWrappedRule) ID() string                 { return "pre-wrapped" }
func (preWrappedRule) Name() string               { return "pre-wrapped" }
func (preWrappedRule) Description() string        { return "returns a wrapped error" }
func (preWrappedRule) Category() linter.Category  { return linter.CategoryDesign }
func (preWrappedRule) Severity() finding.Severity { return finding.SeverityWarning }
func (preWrappedRule) IsEnabledByDefault() bool   { return true }
func (preWrappedRule) Check(context.Context, string) ([]finding.Finding, error) {
	return nil, linter.NewRuleError("pre-wrapped", errSentinel)
}

func TestRegistry_Run_PreWrappedRuleErrorNotDoubleWrapped(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(preWrappedRule{})

	_, err := r.Run(context.Background(), ".")
	if err == nil {
		t.Fatal("expected the rule's error")
	}

	ruleErrs := linter.RuleErrors(err)
	if len(ruleErrs) != 1 {
		t.Fatalf("expected exactly one RuleError (no double wrap), got %d", len(ruleErrs))
	}

	if ruleErrs[0].RuleID != "pre-wrapped" {
		t.Errorf("expected rule ID pre-wrapped, got %q", ruleErrs[0].RuleID)
	}
}

func TestRuleErrors_SingleChainReachingNil(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("layer: %w", errSentinel)
	if linter.RuleErrors(wrapped) != nil {
		t.Error("expected nil for a chain with no RuleError leaves")
	}

	if linter.RuleErrors(nil) != nil {
		t.Error("expected nil for nil input")
	}
}

func TestDetectorFromRegistry_NameReturnsToolName(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()

	detector := linter.DetectorFromRegistry(r, "named-tool")
	if detector.Name() != "named-tool" {
		t.Errorf("expected detector name named-tool, got %q", detector.Name())
	}
}

func TestRuleFunc_SeverityReturnsMetaSev(t *testing.T) {
	t.Parallel()

	rule := makeRule("sev-rule", nil)
	if rule.Severity() != finding.SeverityWarning {
		t.Errorf("expected SeverityWarning from meta, got %v", rule.Severity())
	}
}

// rawErrorRule is a custom Rule whose Check returns a plain (unwrapped)
// error, exercising wrapRuleError's re-wrap branch for non-RuleFunc rules.
type rawErrorRule struct{}

func (rawErrorRule) ID() string                 { return "raw-error" }
func (rawErrorRule) Name() string               { return "raw-error" }
func (rawErrorRule) Description() string        { return "returns a raw error" }
func (rawErrorRule) Category() linter.Category  { return linter.CategoryDesign }
func (rawErrorRule) Severity() finding.Severity { return finding.SeverityWarning }
func (rawErrorRule) IsEnabledByDefault() bool   { return true }
func (rawErrorRule) Check(context.Context, string) ([]finding.Finding, error) {
	return nil, errSentinel
}

func TestRegistry_Run_RawRuleErrorGetsWrapped(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(rawErrorRule{})

	_, err := r.Run(context.Background(), ".")
	if err == nil {
		t.Fatal("expected the rule's error")
	}

	ruleErrs := linter.RuleErrors(err)
	if len(ruleErrs) != 1 {
		t.Fatalf("expected exactly one RuleError, got %d", len(ruleErrs))
	}

	if ruleErrs[0].RuleID != "raw-error" {
		t.Errorf("expected rule ID raw-error, got %q", ruleErrs[0].RuleID)
	}
}

type nilUnwrapError struct{}

func (nilUnwrapError) Error() string { return "wrapper with nil inner" }
func (nilUnwrapError) Unwrap() error { return nil }

func TestRuleErrors_NilUnwrapLeafTerminates(t *testing.T) {
	t.Parallel()

	if linter.RuleErrors(nilUnwrapError{}) != nil {
		t.Error("expected nil for a chain ending in a nil unwrap")
	}
}
