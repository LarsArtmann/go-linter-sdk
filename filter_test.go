package linter_test

import (
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// filteredFixture builds one finding per (severity, category) combination
// used by the Filter tests.
type filterFixture struct {
	name string
	perm map[string][]finding.Finding // ruleID -> findings, keyed by fixture name
}

func filterFinding(sev finding.Severity, cat linter.Category) finding.Finding {
	builder := finding.NewBuilder(
		"filter-fixture", "linter", "filter fixture", sev,
		finding.Pos(finding.FilePath("a.go"), 1, 1),
	)
	if cat != "" {
		builder = builder.WithCategory(finding.Category(cat))
	}

	return builder.MustBuild()
}

func TestFilter_ZeroValueKeepsEverything(t *testing.T) {
	t.Parallel()

	findings := []finding.Finding{
		filterFinding(finding.SeverityInfo, ""),
		filterFinding(finding.SeverityCritical, linter.CategorySecurity),
	}

	f := linter.Filter{}
	got := f.Apply(findings)

	if len(got) != 2 {
		t.Fatalf("expected zero-value Filter to keep all findings, got %d", len(got))
	}
}

func TestFilter_MinSeverity(t *testing.T) {
	t.Parallel()

	findings := []finding.Finding{
		filterFinding(finding.SeverityInfo, linter.CategoryStyle),
		filterFinding(finding.SeverityWarning, linter.CategoryStyle),
		filterFinding(finding.SeverityError, linter.CategoryStyle),
		filterFinding(finding.SeverityCritical, linter.CategoryStyle),
	}

	f := linter.Filter{MinSeverity: finding.SeverityError}
	got := f.Apply(findings)

	if len(got) != 2 {
		t.Fatalf("expected 2 findings at or above error, got %d", len(got))
	}

	for _, fd := range got {
		if fd.Severity.Compare(finding.SeverityError) < 0 {
			t.Errorf("kept finding below threshold: %s", fd.Severity)
		}
	}
}

func TestFilter_Categories(t *testing.T) {
	t.Parallel()

	findings := []finding.Finding{
		filterFinding(finding.SeverityWarning, linter.CategoryStyle),
		filterFinding(finding.SeverityWarning, linter.CategorySecurity),
		filterFinding(finding.SeverityWarning, linter.CategoryDesign),
	}

	f := linter.Filter{Categories: map[linter.Category]bool{
		linter.CategorySecurity: true,
		linter.CategoryDesign:   true,
	}}
	got := f.Apply(findings)

	if len(got) != 2 {
		t.Fatalf("expected 2 findings in allowed categories, got %d", len(got))
	}
}

func TestFilter_Combined(t *testing.T) {
	t.Parallel()

	findings := []finding.Finding{
		filterFinding(finding.SeverityError, linter.CategoryStyle),
		filterFinding(finding.SeverityError, linter.CategorySecurity),
		filterFinding(finding.SeverityWarning, linter.CategorySecurity),
	}

	f := linter.Filter{
		MinSeverity: finding.SeverityError,
		Categories:  map[linter.Category]bool{linter.CategorySecurity: true},
	}
	got := f.Apply(findings)

	if len(got) != 1 {
		t.Fatalf("expected 1 finding (error+security), got %d", len(got))
	}

	if got[0].Category != finding.CategorySecurity {
		t.Errorf("expected security category, got %s", got[0].Category)
	}
}

func TestFilter_NilSliceYieldsNil(t *testing.T) {
	t.Parallel()

	f := linter.Filter{MinSeverity: finding.SeverityWarning}
	if got := f.Apply(nil); got != nil {
		t.Fatalf("expected nil for nil input, got %v", got)
	}
}

func TestFilter_DoesNotMutateInput(t *testing.T) {
	t.Parallel()

	findings := []finding.Finding{
		filterFinding(finding.SeverityInfo, linter.CategoryStyle),
		filterFinding(finding.SeverityError, linter.CategorySecurity),
	}

	f := linter.Filter{MinSeverity: finding.SeverityError}
	_ = f.Apply(findings)

	if len(findings) != 2 {
		t.Fatalf("input mutated: %d findings left", len(findings))
	}
}

func TestFilter_KeepPredicate(t *testing.T) {
	t.Parallel()

	f := linter.Filter{MinSeverity: finding.SeverityWarning}

	if !f.Keep(filterFinding(finding.SeverityWarning, "")) {
		t.Error("expected Keep true at threshold")
	}

	if f.Keep(filterFinding(finding.SeverityInfo, linter.CategorySecurity)) {
		t.Error("expected Keep false below threshold")
	}
}
