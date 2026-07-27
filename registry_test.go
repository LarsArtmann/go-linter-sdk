package linter_test

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// makeRule builds a rule that returns the supplied findings. Returns the
// concrete linter.RuleFunc (not the Rule interface) so ireturn has nothing to
// flag — test factories should hand back concrete types.
func makeRule(name string, findings []finding.Finding) linter.RuleFunc {
	return linter.RuleFunc{
		Meta: linter.RuleMeta{
			Name:        name,
			Description: name + " rule",
			Cat:         linter.CategoryDesign,
			Sev:         finding.SeverityWarning,
		},
		Run: func(_ context.Context, _ string) ([]finding.Finding, error) { return findings, nil },
	}
}

func TestRegistry_RegisterAndAll(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(makeRule("r1", nil))
	r.Register(makeRule("r2", nil))

	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(all))
	}
}

func TestRegistry_DuplicatePanics(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
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

	r := linter.NewRegistry()

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

	r := linter.NewRegistry()

	seen := ""

	r.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{Name: "ctx-rule"},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			seen = dir

			return nil, nil
		},
	})

	detector := linter.DetectorFromRegistry(r, "test-linter")

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

	r := linter.NewRegistry()

	seen := ""

	r.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{Name: "default-dir-rule"},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			seen = dir

			return nil, nil
		},
	})

	detector := linter.DetectorFromRegistry(r, "test-linter")
	if _, err := detector.Detect(context.Background()); err != nil {
		t.Fatal(err)
	}

	if seen != "." {
		t.Errorf("expected '.', got %q", seen)
	}
}

func TestExitCodeFromReport(t *testing.T) {
	t.Parallel()

	if code := linter.ExitCodeFromReport(nil); code != 0 {
		t.Errorf("nil report should be exit 0, got %d", code)
	}

	report := finding.NewReport(finding.ToolInfo{Name: "test"})
	if code := linter.ExitCodeFromReport(report); code != 0 {
		t.Errorf("empty report should be exit 0, got %d", code)
	}

	f := finding.NewBuilder("r", "test", "m", finding.SeverityWarning,
		finding.Pos(finding.FilePath("a.go"), 1, 1)).MustBuild()
	report.AddFindings([]finding.Finding{f})

	if code := linter.ExitCodeFromReport(report); code != 1 {
		t.Errorf("non-empty report should be exit 1, got %d", code)
	}
}

var errSentinel = errors.New("sentinel test error")

// failingRule builds a rule that always fails with errSentinel. Returns the
// concrete linter.RuleFunc so ireturn has nothing to flag.
func failingRule(name string) linter.RuleFunc {
	return linter.RuleFunc{
		Meta: linter.RuleMeta{
			Name:        name,
			Description: "fails",
			Cat:         linter.CategoryDesign,
			Sev:         finding.SeverityWarning,
		},
		Run: func(_ context.Context, _ string) ([]finding.Finding, error) { return nil, errSentinel },
	}
}

func TestRegistry_Run_WrapsRuleError(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(failingRule("boom"))

	_, err := r.Run(context.Background(), ".")

	ruleErr, ok := errors.AsType[*linter.RuleError](err)
	if !ok {
		t.Fatalf("expected *RuleError, got %T: %v", err, err)
	}

	if ruleErr.RuleName != "boom" {
		t.Errorf("expected rule name 'boom', got %q", ruleErr.RuleName)
	}

	if !errors.Is(ruleErr, errSentinel) {
		t.Errorf("expected cause to unwrap to errSentinel, got: %v", ruleErr.Cause)
	}

	if !errors.Is(err, linter.ErrRuleFailed) {
		t.Errorf("expected errors.Is(err, ErrRuleFailed) to be true")
	}
}

func TestRegistry_Run_NoDoubleWrap(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(failingRule("once"))

	_, err := r.Run(context.Background(), ".")

	ruleErr, ok := errors.AsType[*linter.RuleError](err)
	if !ok {
		t.Fatalf("expected *RuleError, got %T: %v", err, err)
	}

	// The cause must be the raw sentinel — a positive signal that the error was
	// wrapped exactly once. A double-wrap (Registry.Run re-wrapping an
	// already-wrapped *RuleError) would make ruleErr.Cause a *RuleError rather
	// than errSentinel. errors.Is is intentionally NOT used here: it matches any
	// error wrapping errSentinel down the chain, so it cannot distinguish one
	// wrap from two.
	if ruleErr.Cause != errSentinel { //nolint:errorlint // exact identity proves single-wrap; errors.Is would also match a re-wrapped cause
		t.Fatalf("expected cause to be raw errSentinel (wrapped exactly once), got %T: %v",
			ruleErr.Cause, ruleErr.Cause)
	}
}

func TestDetectorFromRegistry_WrapsRuleError(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(failingRule("detect-boom"))

	detector := linter.DetectorFromRegistry(r, "test-linter")

	_, err := detector.Detect(context.Background())

	ruleErr, ok := errors.AsType[*linter.RuleError](err)
	if !ok {
		t.Fatalf("expected *RuleError from detector, got %T: %v", err, err)
	}

	if ruleErr.RuleName != "detect-boom" {
		t.Errorf("expected rule name 'detect-boom', got %q", ruleErr.RuleName)
	}

	if !errors.Is(err, linter.ErrRuleFailed) {
		t.Errorf("expected errors.Is(err, ErrRuleFailed) to be true")
	}
}

// TestRegistry_ConcurrentReadWrite exercises the RWMutex: writers register
// uniquely-named rules while readers concurrently snapshot All() and run every
// rule. The race detector (go test -race) must report no data races. Without
// this test, -race passed vacuously because no goroutines touched the registry.
func TestRegistry_ConcurrentReadWrite(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()

	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	for i := range goroutines {
		name := strconv.Itoa(i)

		go func() {
			defer wg.Done()

			r.Register(makeRule(name, nil))
		}()
		go func() {
			defer wg.Done()

			_ = r.All()
		}()
		go func() {
			defer wg.Done()

			_, _ = r.Run(context.Background(), ".")
		}()
	}

	wg.Wait()

	if got := len(r.All()); got != goroutines {
		t.Fatalf("expected %d rules after concurrent registration, got %d", goroutines, got)
	}
}

func BenchmarkRegistry_Register(b *testing.B) {
	r := linter.NewRegistry()
	rules := make([]linter.RuleFunc, b.N)

	for i := range b.N {
		rules[i] = makeRule(strconv.Itoa(i), nil)
	}

	b.ResetTimer()

	for i := range b.N {
		r.Register(rules[i])
	}
}

func BenchmarkRegistry_All(b *testing.B) {
	r := linter.NewRegistry()

	for i := range 100 {
		r.Register(makeRule(strconv.Itoa(i), nil))
	}

	b.ResetTimer()

	for range b.N {
		_ = r.All()
	}
}

func BenchmarkRegistry_Run(b *testing.B) {
	r := linter.NewRegistry()

	for i := range 100 {
		r.Register(makeRule(strconv.Itoa(i), nil))
	}

	ctx := context.Background()

	b.ResetTimer()

	for range b.N {
		_, _ = r.Run(ctx, ".")
	}
}
