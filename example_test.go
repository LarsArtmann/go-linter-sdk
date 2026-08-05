package linter_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

// ExampleRegistry_Run demonstrates the standalone execution path: register
// rules, run them against a directory, and read the aggregated report.
func ExampleRegistry_Run() {
	registry := linter.NewRegistry()

	registry.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          "demo-rule",
			Name:        "demo rule",
			Description: "emits a single finding for demonstration",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
		},
		Run: func(_ context.Context, _ string) ([]finding.Finding, error) {
			return []finding.Finding{
				finding.NewBuilder("demo-rule", "example", "found a style issue",
					finding.SeverityWarning,
					finding.Pos(finding.FilePath("demo.go"), 1, 1)).
					MustBuild(),
			}, nil
		},
	})

	report, _ := registry.Run(context.Background(), ".")

	fmt.Println("findings:", report.Len())
	// Output: findings: 1
}

// ExampleDetectorsFromRegistry demonstrates the go-finding/pipeline execution
// path: each rule becomes its own Detector, enabling per-rule parallelism,
// timeouts, and error isolation.
func ExampleDetectorsFromRegistry() {
	registry := linter.NewRegistry()

	registry.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          "pipe-rule",
			Name:        "pipeline rule",
			Description: "demonstrates the per-rule detector path",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityInfo,
		},
		Run: func(_ context.Context, _ string) ([]finding.Finding, error) {
			return []finding.Finding{
				finding.NewBuilder("pipe-rule", "example", "pipeline finding",
					finding.SeverityInfo,
					finding.Pos(finding.FilePath("main.go"), 10, 1)).
					MustBuild(),
			}, nil
		},
	})

	detectors := linter.DetectorsFromRegistry(registry)

	fmt.Println("detectors:", len(detectors))
	fmt.Println("name:", detectors[0].Name())

	findings, _ := detectors[0].Detect(context.Background())

	fmt.Println("findings:", len(findings))
	// Output:
	// detectors: 1
	// name: pipe-rule
	// findings: 1
}

// ExampleOptIn demonstrates creating a rule that is disabled by default — it
// only runs when a consumer explicitly enables it (e.g. via --enable <id>).
func ExampleOptIn() {
	rule := linter.OptIn(linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          "experimental",
			Name:        "experimental check",
			Description: "noisy; enable explicitly when needed",
			Cat:         linter.CategoryDesign,
			Sev:         finding.SeverityWarning,
		},
		Run: func(_ context.Context, _ string) ([]finding.Finding, error) { return nil, nil },
	})

	fmt.Println("enabled by default:", rule.IsEnabledByDefault())
	fmt.Println("id:", rule.ID())
	// Output:
	// enabled by default: false
	// id: experimental
}

// ExampleRegistry_Run_continueOnError demonstrates the continue-on-error
// failure policy: all rules run regardless of individual failures, partial
// findings are collected, and errors are joined.
func ExampleRegistry_Run_continueOnError() {
	registry := linter.NewRegistry()

	registry.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          "good",
			Name:        "good rule",
			Description: "always succeeds",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
		},
		Run: func(_ context.Context, _ string) ([]finding.Finding, error) {
			return []finding.Finding{
				finding.NewBuilder("good", "example", "partial result",
					finding.SeverityWarning,
					finding.Pos(finding.FilePath("a.go"), 1, 1)).
					MustBuild(),
			}, nil
		},
	})

	registry.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          "bad",
			Name:        "bad rule",
			Description: "always fails",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityWarning,
		},
		Run: func(_ context.Context, _ string) ([]finding.Finding, error) {
			return nil, errors.New("simulated failure")
		},
	})

	report, err := registry.Run(context.Background(), ".", linter.ContinueOnError())

	fmt.Println("findings:", report.Len())
	fmt.Println("has error:", err != nil)
	// Output:
	// findings: 1
	// has error: true
}
