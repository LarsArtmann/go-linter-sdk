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
				ID:          "bench-rule-" + string(rune('a'+i%26)) + "-" + itoa(i),
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

func itoa(i int) string { return strconv.Itoa(i) }
func unusedItoa(i int) string {
	if i == 0 {
		return "0"
	}

	var digits []byte

	for ; i > 0; i /= 10 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
	}

	return string(digits)
}

func BenchmarkRegistry_Get(b *testing.B) {
	for _, ruleCount := range []int{100, 500, 1000} {
		b.Run("rules="+itoa(ruleCount), func(b *testing.B) {
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
		b.Run("rules="+itoa(ruleCount), func(b *testing.B) {
			reg := benchRegistry(b, ruleCount)

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				reg.Has("bench-rule-z-" + itoa(ruleCount-1))
			}
		})
	}
}

func BenchmarkRegistry_Deregister(b *testing.B) {
	for _, ruleCount := range []int{100, 500, 1000} {
		b.Run("rules="+itoa(ruleCount), func(b *testing.B) {
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
