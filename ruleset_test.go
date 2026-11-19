package linter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

func TestRuleSet_ConstructionAndLookup(t *testing.T) {
	t.Parallel()

	r1 := makeRule("s1", nil)
	r2 := makeRule("s2", nil)

	set := linter.NewRuleSet([]linter.Rule{r1, r2})

	all := set.Rules()
	if set.Len() != 2 || len(all) != 2 || all[0].ID() != "s1" || all[1].ID() != "s2" {
		t.Fatalf("expected rules preserved in order, got %d", set.Len())
	}

	if got, ok := set.Get("s1"); !ok || got.ID() != "s1" {
		t.Errorf("expected Get to find s1, got %v %v", got, ok)
	}

	if set.Has("missing") {
		t.Error("expected Has false for unknown ID")
	}
}

func TestRuleSet_RulesReturnsCopy(t *testing.T) {
	t.Parallel()

	set := linter.NewRuleSet([]linter.Rule{makeRule("s1", nil)})

	rules := set.Rules()
	rules[0] = nil

	if set.Rules()[0] == nil {
		t.Fatal("Rules must return a copy; caller mutation leaked into the set")
	}
}

func TestRuleSet_InputSliceMutationSafe(t *testing.T) {
	t.Parallel()

	r1 := makeRule("s1", nil)
	in := []linter.Rule{r1}

	set := linter.NewRuleSet(in)

	in[0] = nil

	if !set.Has("s1") {
		t.Fatal("caller mutation of input slice corrupted the set")
	}
}

func TestRuleSet_DuplicatePanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate ID")
		}
	}()

	linter.NewRuleSet([]linter.Rule{makeRule("dup", nil), makeRule("dup", nil)})
}

func TestRuleSet_EmptyIdentityPanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on empty identity")
		}
	}()

	linter.NewRuleSet([]linter.Rule{linter.RuleFunc{
		Meta: linter.RuleMeta{ID: "", Name: "n", Description: "d", Cat: linter.CategoryStyle},
	}})
}

func TestRuleSet_Check_RelativizesAndOrders(t *testing.T) {
	t.Parallel()

	dir := "/golsdk-test-root"
	rule := makeRule("s1", nil)
	f := rule.NewFinding("issue", finding.Pos(finding.FilePath(dir+"/x.go"), 1, 1)).MustBuild()

	set := linter.NewRuleSet([]linter.Rule{makeRule("s1", []finding.Finding{f})})

	findings, err := set.Check(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(findings) != 1 || string(findings[0].Position.File) != "x.go" {
		t.Fatalf("expected relativized finding, got %+v", findings)
	}
}

func TestRuleSet_Check_WrapsRuleError(t *testing.T) {
	t.Parallel()

	set := linter.NewRuleSet([]linter.Rule{failingRule("boom")})

	findings, err := set.Check(context.Background(), ".")
	if err == nil {
		t.Fatal("expected wrapped rule error")
	}

	if findings != nil {
		t.Fatalf("expected nil findings on failure, got %v", findings)
	}

	ruleErr, found := errors.AsType[*linter.RuleError](err)
	if !found || ruleErr.RuleID != "boom" {
		t.Fatalf("expected RuleError for boom, got %v", err)
	}
}

func TestRuleSet_ToRegistry(t *testing.T) {
	t.Parallel()

	set := linter.NewRuleSet([]linter.Rule{makeRule("s1", nil), makeRule("s2", nil)})

	reg := set.ToRegistry(linter.WithToolName("from-set"))

	all := reg.All()
	if len(all) != 2 || all[0].ID() != "s1" || all[1].ID() != "s2" {
		t.Fatalf("expected registry to preserve order, got %d rules", len(all))
	}

	extracted, ok := all[0].(linter.RuleFunc)
	if !ok {
		t.Fatalf("expected RuleFunc, got %T", all[0])
	}

	if extracted.Meta.ToolName != "from-set" {
		t.Errorf("expected option to apply, got %q", extracted.Meta.ToolName)
	}

	reg.Register(makeRule("s3", nil))

	if set.Len() != 2 || set.Has("s3") {
		t.Error("registry mutation must not leak into the set")
	}
}
