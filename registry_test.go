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
			ID:          name,
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

func TestRegistry_Get(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(makeRule("r1", nil))
	r.Register(makeRule("r2", nil))

	rule, ok := r.Get("r1")
	if !ok {
		t.Fatal("expected to find r1")
	}

	if rule.ID() != "r1" {
		t.Errorf("expected rule ID 'r1', got %q", rule.ID())
	}

	if _, ok := r.Get("nonexistent"); ok {
		t.Error("expected Get to return false for unregistered ID")
	}
}

func TestRegistry_Has(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(makeRule("exists", nil))

	if !r.Has("exists") {
		t.Error("expected Has to return true for registered rule")
	}

	if r.Has("missing") {
		t.Error("expected Has to return false for unregistered rule")
	}
}

func TestRegistry_Deregister(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(makeRule("r1", nil))
	r.Register(makeRule("r2", nil))
	r.Register(makeRule("r3", nil))

	if !r.Deregister("r2") {
		t.Fatal("expected Deregister to return true for existing rule")
	}

	if r.Has("r2") {
		t.Error("expected r2 to be gone after Deregister")
	}

	if got := len(r.All()); got != 2 {
		t.Fatalf("expected 2 rules after Deregister, got %d", got)
	}

	remaining := r.All()
	if remaining[0].ID() != "r1" || remaining[1].ID() != "r3" {
		t.Errorf("expected r1,r3 after Deregister, got %s,%s",
			remaining[0].ID(), remaining[1].ID())
	}

	if r.Deregister("nonexistent") {
		t.Error("expected Deregister to return false for unregistered ID")
	}
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
		Meta: linter.RuleMeta{ID: "ctx-rule", Name: "ctx-rule", Description: "test", Cat: linter.CategoryStyle},
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
		Meta: linter.RuleMeta{
			ID:          "default-dir-rule",
			Name:        "default-dir-rule",
			Description: "test",
			Cat:         linter.CategoryStyle,
		},
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
			ID:          name,
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

	if ruleErr.RuleID != "boom" {
		t.Errorf("expected rule ID 'boom', got %q", ruleErr.RuleID)
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

func TestRegistry_Run_ContinueOnError_CollectsPartialAndErrors(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()

	f1 := finding.NewBuilder("ok-rule", "test", "found", finding.SeverityWarning,
		finding.Pos(finding.FilePath("a.go"), 1, 1)).MustBuild()

	r.Register(makeRule("ok-rule", []finding.Finding{f1}))
	r.Register(failingRule("bad-rule"))

	report, err := r.Run(context.Background(), ".", linter.ContinueOnError())

	if report == nil {
		t.Fatal("expected non-nil report in continue-on-error mode")
	}

	if report.Len() != 1 {
		t.Errorf("expected 1 finding from the successful rule, got %d", report.Len())
	}

	if err == nil {
		t.Fatal("expected error from failed rule in continue-on-error mode")
	}

	if !errors.Is(err, linter.ErrRuleFailed) {
		t.Error("expected errors.Is(err, ErrRuleFailed) to be true")
	}
}

func TestRegistry_Run_ContinueOnError_MultipleFailures(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()

	f1 := finding.NewBuilder("ok", "test", "ok", finding.SeverityInfo,
		finding.Pos(finding.FilePath("ok.go"), 1, 1)).MustBuild()

	r.Register(failingRule("fail-1"))
	r.Register(makeRule("ok", []finding.Finding{f1}))
	r.Register(failingRule("fail-2"))

	report, err := r.Run(context.Background(), ".", linter.ContinueOnError())

	if report == nil || report.Len() != 1 {
		t.Fatalf("expected 1 finding from successful rule, got %d", report.Len())
	}

	if err == nil {
		t.Fatal("expected joined error from two failed rules")
	}

	if !errors.Is(err, linter.ErrRuleFailed) {
		t.Error("expected errors.Is(err, ErrRuleFailed) to be true")
	}
}

func TestRegistry_Run_DefaultFailsFast(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()

	secondRan := false

	r.Register(failingRule("first-fail"))
	r.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          "second",
			Name:        "second",
			Description: "should not run",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
		},
		Run: func(_ context.Context, _ string) ([]finding.Finding, error) {
			secondRan = true

			return nil, nil
		},
	})

	report, err := r.Run(context.Background(), ".")

	if report != nil {
		t.Error("expected nil report on fail-fast")
	}

	if err == nil {
		t.Fatal("expected error on fail-fast")
	}

	if secondRan {
		t.Error("second rule should not have run in fail-fast mode")
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

	if ruleErr.RuleID != "detect-boom" {
		t.Errorf("expected rule ID 'detect-boom', got %q", ruleErr.RuleID)
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

	for b.Loop() {
		_ = r.All()
	}
}

func TestRuleFunc_IsEnabledByDefault(t *testing.T) {
	t.Parallel()

	r := makeRule("default-on", nil)
	if !r.IsEnabledByDefault() {
		t.Errorf("RuleFunc should be enabled by default")
	}
}

func TestOptIn_IsDisabledByDefault(t *testing.T) {
	t.Parallel()

	optIn := linter.OptIn(makeRule("opt-in", nil))
	if optIn.IsEnabledByDefault() {
		t.Errorf("OptIn rule should be disabled by default")
	}

	if optIn.Name() != "opt-in" {
		t.Errorf("OptIn should preserve rule name, got %q", optIn.Name())
	}
}

func TestRegistry_DuplicateIDPanics(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(makeRule("same-id", nil))

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate ID")
		}
	}()

	r.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          "same-id",
			Name:        "different-display-name",
			Description: "test",
			Cat:         linter.CategoryStyle,
		},
		Run: func(_ context.Context, _ string) ([]finding.Finding, error) { return nil, nil },
	})
}

func TestRegistry_EmptyIDPanics(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on empty ID")
		}
	}()

	r.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{Name: "has-name-but-no-id"},
		Run:  func(_ context.Context, _ string) ([]finding.Finding, error) { return nil, nil },
	})
}

func TestRegistry_EmptyDescriptionPanics(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on empty Description")
		}
	}()

	r.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{ID: "has-id", Name: "has-name", Cat: linter.CategoryStyle},
		Run:  func(_ context.Context, _ string) ([]finding.Finding, error) { return nil, nil },
	})
}

func TestRegistry_EmptyCategoryPanics(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on empty Category")
		}
	}()

	r.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{ID: "has-id", Name: "has-name", Description: "has-desc"},
		Run:  func(_ context.Context, _ string) ([]finding.Finding, error) { return nil, nil },
	})
}

func TestRuleMeta_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		meta    linter.RuleMeta
		wantErr bool
	}{
		{
			name:    "all fields set",
			meta:    linter.RuleMeta{ID: "x", Name: "X", Description: "desc", Cat: linter.CategoryStyle},
			wantErr: false,
		},
		{
			name:    "empty ID",
			meta:    linter.RuleMeta{Name: "X", Description: "desc", Cat: linter.CategoryStyle},
			wantErr: true,
		},
		{
			name:    "empty Name",
			meta:    linter.RuleMeta{ID: "x", Description: "desc", Cat: linter.CategoryStyle},
			wantErr: true,
		},
		{
			name:    "empty Description",
			meta:    linter.RuleMeta{ID: "x", Name: "X", Cat: linter.CategoryStyle},
			wantErr: true,
		},
		{
			name:    "empty Cat",
			meta:    linter.RuleMeta{ID: "x", Name: "X", Description: "desc"},
			wantErr: true,
		},
		{
			name:    "multiple missing",
			meta:    linter.RuleMeta{ID: "x"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.meta.Validate()
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestDetectorsFromRegistry_OneDetectorPerRule(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()

	f1 := finding.NewBuilder("id-a", "test", "a", finding.SeverityWarning,
		finding.Pos(finding.FilePath("a.go"), 1, 1)).MustBuild()
	f2 := finding.NewBuilder("id-b", "test", "b", finding.SeverityError,
		finding.Pos(finding.FilePath("b.go"), 2, 1)).MustBuild()

	r.Register(makeRule("id-a", []finding.Finding{f1}))
	r.Register(makeRule("id-b", []finding.Finding{f2}))

	detectors := linter.DetectorsFromRegistry(r)
	if len(detectors) != 2 {
		t.Fatalf("expected 2 detectors, got %d", len(detectors))
	}

	if detectors[0].Name() != "id-a" {
		t.Errorf("expected first detector name 'id-a', got %q", detectors[0].Name())
	}

	if detectors[1].Name() != "id-b" {
		t.Errorf("expected second detector name 'id-b', got %q", detectors[1].Name())
	}
}

func TestDetectorsFromRegistry_ReadsWorkDirFromContext(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()

	seen := ""

	r.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{ID: "dir-rule", Name: "dir-rule", Description: "test", Cat: linter.CategoryStyle},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			seen = dir

			return nil, nil
		},
	})

	detectors := linter.DetectorsFromRegistry(r)

	ctx := finding.WithWorkingDir(context.Background(), "/custom/dir")
	if _, err := detectors[0].Detect(ctx); err != nil {
		t.Fatal(err)
	}

	if seen != "/custom/dir" {
		t.Errorf("expected /custom/dir, got %q", seen)
	}
}

func TestDetectorsFromRegistry_WrapsRuleError(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()
	r.Register(failingRule("detect-multi-boom"))

	detectors := linter.DetectorsFromRegistry(r)

	_, err := detectors[0].Detect(context.Background())

	ruleErr, ok := errors.AsType[*linter.RuleError](err)
	if !ok {
		t.Fatalf("expected *RuleError from detector, got %T: %v", err, err)
	}

	if ruleErr.RuleID != "detect-multi-boom" {
		t.Errorf("expected rule ID 'detect-multi-boom', got %q", ruleErr.RuleID)
	}

	if !errors.Is(err, linter.ErrRuleFailed) {
		t.Errorf("expected errors.Is(err, ErrRuleFailed) to be true")
	}
}

func TestDetectorsFromRegistry_EmptyRegistry(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()

	detectors := linter.DetectorsFromRegistry(r)
	if len(detectors) != 0 {
		t.Fatalf("expected 0 detectors for empty registry, got %d", len(detectors))
	}
}

func BenchmarkRegistry_Run(b *testing.B) {
	r := linter.NewRegistry()

	for i := range 100 {
		r.Register(makeRule(strconv.Itoa(i), nil))
	}

	ctx := context.Background()

	for b.Loop() {
		_, _ = r.Run(ctx, ".")
	}
}

func TestWithToolName_AutoStampsRuleFunc(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry(linter.WithToolName("my-linter"))
	r.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          "r1",
			Name:        "r1",
			Description: "test",
			Cat:         linter.CategoryStyle,
		},
		Run: func(_ context.Context, _ string) ([]finding.Finding, error) { return nil, nil },
	})

	rules := r.All()

	extractedRule, ok := rules[0].(linter.RuleFunc)
	if !ok {
		t.Fatalf("expected RuleFunc, got %T", rules[0])
	}

	if extractedRule.Meta.ToolName != "my-linter" {
		t.Errorf("expected auto-stamped tool name 'my-linter', got %q", extractedRule.Meta.ToolName)
	}
}

func TestWithToolName_DoesNotOverrideExistingToolName(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry(linter.WithToolName("registry-tool"))
	r.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          "r1",
			Name:        "r1",
			Description: "test",
			Cat:         linter.CategoryStyle,
			ToolName:    "explicit-tool",
		},
		Run: func(_ context.Context, _ string) ([]finding.Finding, error) { return nil, nil },
	})

	rules := r.All()
	extractedRule, _ := rules[0].(linter.RuleFunc)

	if extractedRule.Meta.ToolName != "explicit-tool" {
		t.Errorf("expected 'explicit-tool' to be preserved, got %q", extractedRule.Meta.ToolName)
	}
}

func TestWithToolName_AutoStampsOptInRule(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry(linter.WithToolName("my-linter"))
	r.Register(linter.OptIn(linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          "r1",
			Name:        "r1",
			Description: "test",
			Cat:         linter.CategoryStyle,
		},
		Run: func(_ context.Context, _ string) ([]finding.Finding, error) { return nil, nil },
	}))

	rules := r.All()

	optIn, ok := rules[0].(interface{ IsEnabledByDefault() bool })
	if !ok || optIn.IsEnabledByDefault() {
		t.Fatalf("expected opt-in rule, got %T", rules[0])
	}
}

func TestWithToolName_DefaultIsLinter(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry()

	report, err := r.Run(context.Background(), ".")
	if err != nil {
		t.Fatal(err)
	}

	if report.Tool.Name != "linter" {
		t.Errorf("expected default tool name 'linter', got %q", report.Tool.Name)
	}
}

func TestWithToolName_AppliesToReport(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistry(linter.WithToolName("custom-linter"))

	report, err := r.Run(context.Background(), ".")
	if err != nil {
		t.Fatal(err)
	}

	if report.Tool.Name != "custom-linter" {
		t.Errorf("expected 'custom-linter', got %q", report.Tool.Name)
	}
}

func TestFilterRules_NoFilterReturnsAll(t *testing.T) {
	t.Parallel()

	all := []linter.RuleFunc{
		makeRule("a", nil),
		makeRule("b", nil),
		makeRule("c", nil),
	}

	got := linter.FilterRules(all, nil, nil)
	if len(got) != 3 {
		t.Fatalf("expected 3 rules with no filtering, got %d", len(got))
	}
}

func TestFilterRules_DisableRemovesRule(t *testing.T) {
	t.Parallel()

	all := []linter.RuleFunc{
		makeRule("a", nil),
		makeRule("b", nil),
		makeRule("c", nil),
	}

	disable := map[string]bool{"b": true}
	got := linter.FilterRules(all, nil, disable)

	if len(got) != 2 {
		t.Fatalf("expected 2 rules after disabling b, got %d", len(got))
	}

	for _, rule := range got {
		if rule.Meta.ID == "b" {
			t.Error("disabled rule 'b' should not be in filtered set")
		}
	}
}

func TestFilterRules_EnableWhitelistsRules(t *testing.T) {
	t.Parallel()

	all := []linter.RuleFunc{
		makeRule("a", nil),
		makeRule("b", nil),
		makeRule("c", nil),
	}

	enable := map[string]bool{"a": true, "c": true}
	got := linter.FilterRules(all, enable, nil)

	if len(got) != 2 {
		t.Fatalf("expected 2 rules when enabling a+c, got %d", len(got))
	}

	ids := []string{got[0].Meta.ID, got[1].Meta.ID}
	if ids[0] != "a" || ids[1] != "c" {
		t.Errorf("expected [a c], got %v", ids)
	}
}

func TestFilterRules_EnableAndDisableCombined(t *testing.T) {
	t.Parallel()

	all := []linter.RuleFunc{
		makeRule("a", nil),
		makeRule("b", nil),
		makeRule("c", nil),
	}

	enable := map[string]bool{"a": true, "b": true}
	disable := map[string]bool{"b": true}
	got := linter.FilterRules(all, enable, disable)

	if len(got) != 1 {
		t.Fatalf("expected 1 rule (a only, b disabled), got %d", len(got))
	}

	if got[0].Meta.ID != "a" {
		t.Errorf("expected rule 'a', got %q", got[0].Meta.ID)
	}
}

func TestExitCodeByConfidence_NilReport(t *testing.T) {
	t.Parallel()

	if code := linter.ExitCodeByConfidence(nil, finding.ConfidenceHigh); code != 0 {
		t.Errorf("nil report should be exit 0, got %d", code)
	}
}

func TestExitCodeByConfidence_EmptyReport(t *testing.T) {
	t.Parallel()

	report := finding.NewReport(finding.ToolInfo{Name: "test"})
	if code := linter.ExitCodeByConfidence(report, finding.ConfidenceHigh); code != 0 {
		t.Errorf("empty report should be exit 0, got %d", code)
	}
}

func TestExitCodeByConfidence_AboveThreshold(t *testing.T) {
	t.Parallel()

	report := finding.NewReport(finding.ToolInfo{Name: "test"})

	high := finding.NewBuilder("r", "test", "high-confidence",
		finding.SeverityWarning,
		finding.Pos(finding.FilePath("a.go"), 1, 1)).
		WithConfidence(finding.ConfidenceFull).
		MustBuild()
	report.AddFindings([]finding.Finding{high})

	if code := linter.ExitCodeByConfidence(report, finding.ConfidenceHigh); code != 1 {
		t.Errorf("report with full-confidence finding should be exit 1, got %d", code)
	}
}

func TestExitCodeByConfidence_BelowThreshold(t *testing.T) {
	t.Parallel()

	report := finding.NewReport(finding.ToolInfo{Name: "test"})

	low := finding.NewBuilder("r", "test", "low-confidence",
		finding.SeverityInfo,
		finding.Pos(finding.FilePath("a.go"), 1, 1)).
		WithConfidence(finding.ConfidenceLow).
		MustBuild()
	report.AddFindings([]finding.Finding{low})

	if code := linter.ExitCodeByConfidence(report, finding.ConfidenceHigh); code != 2 {
		t.Errorf("report with only low-confidence findings should be exit 2, got %d", code)
	}
}

func TestExitCodeByConfidence_MixedConfidence(t *testing.T) {
	t.Parallel()

	report := finding.NewReport(finding.ToolInfo{Name: "test"})

	low := finding.NewBuilder("r1", "test", "low",
		finding.SeverityInfo,
		finding.Pos(finding.FilePath("a.go"), 1, 1)).
		WithConfidence(finding.ConfidenceLow).
		MustBuild()
	high := finding.NewBuilder("r2", "test", "high",
		finding.SeverityWarning,
		finding.Pos(finding.FilePath("b.go"), 1, 1)).
		WithConfidence(finding.ConfidenceHigh).
		MustBuild()
	report.AddFindings([]finding.Finding{low, high})

	if code := linter.ExitCodeByConfidence(report, finding.ConfidenceHigh); code != 1 {
		t.Errorf("report with at least one high-confidence finding should be exit 1, got %d", code)
	}
}

func TestRuleFunc_NewFinding_PreFillsIdentity(t *testing.T) {
	t.Parallel()

	rule := linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          "H001",
			Name:        "test-rule",
			Description: "test description",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
			ToolName:    "go-humanize-linter",
		},
	}

	pos := finding.Pos(finding.FilePath("demo.go"), 10, 5)
	result := rule.NewFinding("manual byte formatting", pos).MustBuild()

	if result.Rule != "H001" {
		t.Errorf("expected rule 'H001', got %q", result.Rule)
	}

	if result.ToolName != "go-humanize-linter" {
		t.Errorf("expected tool 'go-humanize-linter', got %q", result.ToolName)
	}

	if result.Severity != finding.SeverityWarning {
		t.Errorf("expected severity warning, got %q", result.Severity)
	}

	if result.Category != finding.CategoryStyle {
		t.Errorf("expected category style, got %q", result.Category)
	}
}

func TestRuleFunc_NewFinding_DefaultsToLinterTool(t *testing.T) {
	t.Parallel()

	rule := linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          "R001",
			Name:        "test",
			Description: "test",
			Cat:         linter.CategoryDesign,
			Sev:         finding.SeverityInfo,
		},
	}

	result := rule.NewFinding("msg", finding.Pos(finding.FilePath("x.go"), 1, 1)).MustBuild()

	if result.ToolName != "linter" {
		t.Errorf("expected fallback tool 'linter', got %q", result.ToolName)
	}
}

func TestRuleFunc_NewFinding_AllowsChaining(t *testing.T) {
	t.Parallel()

	rule := linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          "H001",
			Name:        "test",
			Description: "test",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
			ToolName:    "my-linter",
		},
	}

	result := rule.NewFinding("found issue", finding.Pos(finding.FilePath("a.go"), 1, 1)).
		WithConfidence(finding.ConfidenceHigh).
		WithSuggestion("use humanize.Bytes").
		WithFixStrategy(finding.FixStrategySuggest).
		MustBuild()

	if result.Confidence != finding.ConfidenceHigh {
		t.Errorf("expected ConfidenceHigh, got %v", result.Confidence)
	}

	if result.Suggestion != "use humanize.Bytes" {
		t.Errorf("expected suggestion, got %q", result.Suggestion)
	}

	if result.FixStrategy != finding.FixStrategySuggest {
		t.Errorf("expected FixStrategySuggest, got %q", result.FixStrategy)
	}
}
