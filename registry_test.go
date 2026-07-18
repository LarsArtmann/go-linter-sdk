package linter

import (
	"context"
	"errors"
	"testing"

	"github.com/larsartmann/go-finding"
)

func makeRule(name string, findings []finding.Finding) Rule {
	return RuleFunc{
		Meta: RuleMeta{Name: name, Description: name + " rule", Cat: CategoryDesign, Sev: finding.SeverityWarning},
		Run:  func(_ context.Context, _ string) ([]finding.Finding, error) { return findings, nil },
	}
}

func TestRegistry_RegisterAndAll(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	r.Register(makeRule("r1", nil))
	r.Register(makeRule("r2", nil))

	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(all))
	}
}

func TestRegistry_DuplicatePanics(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	r.Register(makeRule("dup", nil))

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate")
		}
	}()

	r.Register(makeRule("dup", nil))
}

func TestRegistry_Run_AggregatesFindings(t *testing.T) {
	t.Parallel()

	r := NewRegistry()

	f1 := finding.NewBuilder("rule-a", "test", "a", finding.SeverityWarning,
		finding.Pos(finding.FilePath("a.go"), 1, 1)).MustBuild()
	f2 := finding.NewBuilder("rule-b", "test", "b", finding.SeverityError,
		finding.Pos(finding.FilePath("b.go"), 2, 1)).MustBuild()

	r.Register(makeRule("rule-a", []finding.Finding{f1}))
	r.Register(makeRule("rule-b", []finding.Finding{f2}))

	report, err := r.Run(context.Background(), ".")
	if err != nil {
		t.Fatal(err)
	}

	if report.Len() != 2 {
		t.Fatalf("expected 2 findings, got %d", report.Len())
	}
}

func TestDetectorFromRegistry_ReadsWorkDirFromContext(t *testing.T) {
	t.Parallel()

	r := NewRegistry()

	seen := ""

	r.Register(RuleFunc{
		Meta: RuleMeta{Name: "ctx-rule"},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			seen = dir

			return nil, nil
		},
	})

	detector := DetectorFromRegistry(r, "test-linter")

	ctx := finding.WithWorkingDir(context.Background(), "/custom/dir")
	if _, err := detector.Detect(ctx); err != nil {
		t.Fatal(err)
	}

	if seen != "/custom/dir" {
		t.Errorf("expected /custom/dir, got %q", seen)
	}
}

func TestDetectorFromRegistry_DefaultDir(t *testing.T) {
	t.Parallel()

	r := NewRegistry()

	seen := ""

	r.Register(RuleFunc{
		Meta: RuleMeta{Name: "default-dir-rule"},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			seen = dir

			return nil, nil
		},
	})

	detector := DetectorFromRegistry(r, "test-linter")
	if _, err := detector.Detect(context.Background()); err != nil {
		t.Fatal(err)
	}

	if seen != "." {
		t.Errorf("expected '.', got %q", seen)
	}
}

func TestExitCodeFromReport(t *testing.T) {
	t.Parallel()

	if code := ExitCodeFromReport(nil); code != 0 {
		t.Errorf("nil report should be exit 0, got %d", code)
	}

	report := finding.NewReport(finding.ToolInfo{Name: "test"})
	if code := ExitCodeFromReport(report); code != 0 {
		t.Errorf("empty report should be exit 0, got %d", code)
	}

	f := finding.NewBuilder("r", "test", "m", finding.SeverityWarning,
		finding.Pos(finding.FilePath("a.go"), 1, 1)).MustBuild()
	report.AddFindings([]finding.Finding{f})

	if code := ExitCodeFromReport(report); code != 1 {
		t.Errorf("non-empty report should be exit 1, got %d", code)
	}
}

var errSentinel = errors.New("sentinel test error")

func failingRule(name string) Rule {
	return RuleFunc{
		Meta: RuleMeta{Name: name, Description: "fails", Cat: CategoryDesign, Sev: finding.SeverityWarning},
		Run:  func(_ context.Context, _ string) ([]finding.Finding, error) { return nil, errSentinel },
	}
}

func TestRegistry_Run_WrapsRuleError(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	r.Register(failingRule("boom"))

	_, err := r.Run(context.Background(), ".")

	var ruleErr *RuleError
	if !errors.As(err, &ruleErr) {
		t.Fatalf("expected *RuleError, got %T: %v", err, err)
	}

	if ruleErr.RuleName != "boom" {
		t.Errorf("expected rule name 'boom', got %q", ruleErr.RuleName)
	}

	if !errors.Is(ruleErr, errSentinel) {
		t.Errorf("expected cause to unwrap to errSentinel, got: %v", ruleErr.Cause)
	}

	if !errors.Is(err, ErrRuleFailed) {
		t.Errorf("expected errors.Is(err, ErrRuleFailed) to be true")
	}
}

func TestRegistry_Run_NoDoubleWrap(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	r.Register(failingRule("once"))

	_, err := r.Run(context.Background(), ".")

	var ruleErr *RuleError
	if !errors.As(err, &ruleErr) {
		t.Fatalf("expected *RuleError, got %T: %v", err, err)
	}

	// RuleFunc.Check wraps once; Registry.Run must not wrap again. A second
	// wrapping would make ruleErr.Cause itself a *RuleError, which is the
	// double-wrap bug this test guards against.
	var inner *RuleError
	if errors.As(ruleErr.Cause, &inner) {
		t.Fatalf("error double-wrapped: outer RuleName=%s, inner RuleName=%s",
			ruleErr.RuleName, inner.RuleName)
	}
}

func TestDetectorFromRegistry_WrapsRuleError(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	r.Register(failingRule("detect-boom"))

	detector := DetectorFromRegistry(r, "test-linter")

	_, err := detector.Detect(context.Background())

	var ruleErr *RuleError
	if !errors.As(err, &ruleErr) {
		t.Fatalf("expected *RuleError from detector, got %T: %v", err, err)
	}

	if ruleErr.RuleName != "detect-boom" {
		t.Errorf("expected rule name 'detect-boom', got %q", ruleErr.RuleName)
	}

	if !errors.Is(err, ErrRuleFailed) {
		t.Errorf("expected errors.Is(err, ErrRuleFailed) to be true")
	}
}
