// Command no-go-mod is a pilot port of go-structure-linter's NoGoModRule,
// rewritten to use go-linter-sdk. It validates the converter-deletion claim:
// a rule emits finding.Finding directly via rule.NewFinding(...), with no
// intermediate Violation/Issue type and no converter layer.
//
// Original (go-structure-linter, ~40 LOC + LinterRule base + types.Project):
//
//	func (r *NoGoModRule) CheckWithProject(proj types.Project) []types.Issue {
//	    if !proj.FileExists("go.mod") {
//	        issues = append(issues, types.Issue{Rule: r.Name(), ...})
//	    }
//	    return issues
//	}
//
// Ported (go-linter-sdk, self-contained):
//
//	Run: func(ctx, dir) ([]finding.Finding, error) { ... }
//
// Run:
//
//	go run ./examples/no-go-mod [dir]
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

const (
	ruleID   = "no-go-mod"
	toolName = "no-go-mod"
)

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	registry := linter.NewRegistry(linter.WithToolName(toolName))

	var goModRule linter.RuleFunc

	goModRule = linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          ruleID,
			Name:        "missing go.mod",
			Description: "ensures go.mod exists for Go module management",
			Cat:         linter.CategoryStructure,
			Sev:         finding.SeverityCritical,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, d string) ([]finding.Finding, error) {
			if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
				return nil, nil
			}

			return []finding.Finding{
				goModRule.NewFinding(
					"Missing go.mod file. Go modules are mandatory for dependency management since Go 1.16.",
					finding.Pos(finding.FilePath("go.mod"), 0, 0),
				).
					WithConfidence(finding.ConfidenceHigh).
					WithFixStrategy(finding.FixStrategySuggest).
					WithSuggestion("Run 'go mod init <module-path>' to create go.mod.").
					MustBuild(),
			}, nil
		},
	}

	registry.Register(goModRule)

	report, err := registry.Run(context.Background(), dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", toolName, err)

		os.Exit(2)
	}

	for _, f := range report.ActiveFindings() {
		fmt.Printf("%s: %s\n", f.Position.File, f.Message)
	}

	os.Exit(linter.ExitCodeFromReport(report))
}
