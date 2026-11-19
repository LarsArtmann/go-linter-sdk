package linter_test

import (
	"slices"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

func TestNewRegistryFromRules(t *testing.T) {
	t.Parallel()

	rules := []linter.Rule{makeRule("r1", nil), makeRule("r2", nil)}

	r := linter.NewRegistryFromRules(rules, linter.WithToolName("bulk"))

	all := r.All()
	if len(all) != 2 || all[0].ID() != "r1" || all[1].ID() != "r2" {
		t.Fatalf("expected rules in registration order, got %d rules", len(all))
	}

	if !r.Has("r1") || !r.Has("r2") {
		t.Error("expected Get/Has to work on the constructed registry")
	}

	rule, ok := r.Get("r1")
	if !ok {
		t.Fatal("expected r1 present")
	}

	extracted, ok := rule.(linter.RuleFunc)
	if !ok {
		t.Fatalf("expected RuleFunc, got %T", rule)
	}

	if extracted.Meta.ToolName != "bulk" {
		t.Errorf("expected WithToolName auto-stamp to apply, got %q", extracted.Meta.ToolName)
	}
}

func TestNewRegistryFromRules_Empty(t *testing.T) {
	t.Parallel()

	r := linter.NewRegistryFromRules(nil)

	if len(r.All()) != 0 {
		t.Fatal("expected empty registry for nil rules")
	}
}

func TestNewRegistryFromRules_DuplicatePanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate rule ID")
		}
	}()

	linter.NewRegistryFromRules([]linter.Rule{makeRule("dup", nil), makeRule("dup", nil)})
}

func TestAllCategories(t *testing.T) {
	t.Parallel()

	cats := linter.AllCategories()
	if len(cats) != 11 {
		t.Fatalf("expected 11 recommended categories, got %d", len(cats))
	}

	if !slices.Contains(cats, linter.CategoryDesign) || !slices.Contains(cats, linter.CategoryConfiguration) {
		t.Error("expected built-in categories present")
	}

	cats[0] = "mutated"

	if linter.AllCategories()[0] == "mutated" {
		t.Error("AllCategories must return a fresh copy, not shared state")
	}
}

func TestExitCodeFromFindings(t *testing.T) {
	t.Parallel()

	if code := linter.ExitCodeFromFindings(nil); code != 0 {
		t.Errorf("expected exit 0 for nil findings, got %d", code)
	}

	if code := linter.ExitCodeFromFindings([]finding.Finding{}); code != 0 {
		t.Errorf("expected exit 0 for empty findings, got %d", code)
	}

	f := makeRule("x", nil).NewFinding("issue", finding.Pos(finding.FilePath("a.go"), 1, 1)).MustBuild()
	if code := linter.ExitCodeFromFindings([]finding.Finding{f}); code != 1 {
		t.Errorf("expected exit 1 for findings, got %d", code)
	}
}
