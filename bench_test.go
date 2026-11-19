package linter_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// benchRegistry builds a registry with n unique rules.
func benchRegistry(b *testing.B, n int) *linter.Registry {
	b.Helper()

	reg := linter.NewRegistry()

	for i := range n {
		reg.Register(linter.RuleFunc{
			Meta: linter.RuleMeta{
				ID:          "bench-rule-" + string(rune('a'+i%26)) + "-" + strconv.Itoa(i),
				Name:        "bench rule",
				Description: "benchmark rule",
				Cat:         linter.CategoryStyle,
				Sev:         finding.SeverityInfo,
			},
			Run: func(_ context.Context, _ string) ([]finding.Finding, error) { return nil, nil },
		})
	}

	return reg
}

func BenchmarkRegistry_Get(b *testing.B) {
	for _, ruleCount := range []int{100, 500, 1000} {
		b.Run("rules="+strconv.Itoa(ruleCount), func(b *testing.B) {
			reg := benchRegistry(b, ruleCount)

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				if _, ok := reg.Get("bench-rule-a-0"); !ok {
					b.Fatal("expected hit")
				}
			}
		})
	}
}

func BenchmarkRegistry_Has(b *testing.B) {
	for _, ruleCount := range []int{100, 500, 1000} {
		b.Run("rules="+strconv.Itoa(ruleCount), func(b *testing.B) {
			reg := benchRegistry(b, ruleCount)

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				reg.Has("bench-rule-z-" + strconv.Itoa(ruleCount-1))
			}
		})
	}
}

func BenchmarkRegistry_Deregister(b *testing.B) {
	for _, ruleCount := range []int{100, 500, 1000} {
		b.Run("rules="+strconv.Itoa(ruleCount), func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				b.StopTimer()
				reg := benchRegistry(b, ruleCount)
				b.StartTimer()

				if !reg.Deregister("bench-rule-a-0") {
					b.Fatal("expected deregister to remove the rule")
				}
			}
		})
	}
}

func BenchmarkRegistry_Run_Coverage(b *testing.B) {
	reg := linter.NewRegistry()

	for i := range 10 {
		reg.Register(coverageRule{RuleFunc: makeRule("bench-run-"+strconv.Itoa(i), nil), files: 50})
	}

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		report, err := reg.Run(ctx, ".")
		if err != nil {
			b.Fatal(err)
		}

		if report.Summary.FilesScanned != 500 {
			b.Fatalf("expected 500 files scanned, got %d", report.Summary.FilesScanned)
		}
	}
}

func benchRuleSet(b *testing.B, n int) *linter.RuleSet {
	b.Helper()

	rules := make([]linter.Rule, n)

	for i := range n {
		rules[i] = linter.RuleFunc{
			Meta: linter.RuleMeta{
				ID:          "bench-rule-" + string(rune('a'+i%26)) + "-" + strconv.Itoa(i),
				Name:        "bench rule",
				Description: "benchmark rule",
				Cat:         linter.CategoryStyle,
				Sev:         finding.SeverityInfo,
			},
			Run: func(_ context.Context, _ string) ([]finding.Finding, error) { return nil, nil },
		}
	}

	return linter.NewRuleSet(rules)
}

func BenchmarkRuleSet_Get(b *testing.B) {
	for _, ruleCount := range []int{100, 500, 1000} {
		b.Run("rules="+strconv.Itoa(ruleCount), func(b *testing.B) {
			set := benchRuleSet(b, ruleCount)

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				if _, ok := set.Get("bench-rule-a-0"); !ok {
					b.Fatal("expected hit")
				}
			}
		})
	}
}

// benchExecRules returns n unique no-op rules shared by the end-to-end
// execution benchmarks below.
func benchExecRules(b *testing.B, n int) []linter.Rule {
	b.Helper()

	rules := make([]linter.Rule, n)

	for i := range n {
		rules[i] = linter.RuleFunc{
			Meta: linter.RuleMeta{
				ID:          "exec-rule-" + strconv.Itoa(i),
				Name:        "exec rule",
				Description: "execution benchmark rule",
				Cat:         linter.CategoryStyle,
				Sev:         finding.SeverityInfo,
			},
			Run: func(_ context.Context, _ string) ([]finding.Finding, error) { return nil, nil },
		}
	}

	return rules
}

// BenchmarkRuleSet_Check and BenchmarkRegistry_Run quantify the end-to-end
// execution claim over the same 50-rule workload: RuleSet.Check is the
// mutex-free findings-only path; Registry.Run additionally acquires read
// locks per rule, wraps results, and builds a Report with coverage evidence.
func BenchmarkRuleSet_Check(b *testing.B) {
	set := linter.NewRuleSet(benchExecRules(b, 50))
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		findings, err := set.Check(ctx, ".")
		if err != nil {
			b.Fatal(err)
		}

		if len(findings) != 0 {
			b.Fatalf("expected 0 findings, got %d", len(findings))
		}
	}
}

func benchFindings(b *testing.B, n int) []finding.Finding {
	b.Helper()

	out := make([]finding.Finding, n)

	for i := range n {
		sev := finding.SeverityInfo
		if i%2 == 0 {
			sev = finding.SeverityError
		}

		out[i] = finding.NewBuilder("filter-bench", "bench", "bench finding", sev,
			finding.Pos(finding.FilePath("a.go"), i+1, 1)).
			WithCategory(finding.Category(linter.CategoryStyle)).
			MustBuild()
	}

	return out
}

func BenchmarkFilter_Apply(b *testing.B) {
	total := 1000

	findings := benchFindings(b, total)

	gate := linter.Filter{MinSeverity: finding.SeverityError}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		kept := gate.Apply(findings)
		if len(kept) != total/2 {
			b.Fatalf("expected %d kept, got %d", total/2, len(kept))
		}
	}
}
