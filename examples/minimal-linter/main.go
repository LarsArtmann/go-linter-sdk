// Command minimal-linter demonstrates the full go-linter-sdk lifecycle: define
// a rule that emits finding.Finding directly, register it, run the registry,
// print findings, and exit with the ecosystem exit code.
//
// The rule checks for the presence of README.md in the target directory. Run:
//
//	GOEXPERIMENT=jsonv2 go run ./examples/minimal-linter [dir]
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	registry := linter.NewRegistry()

	registry.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{ //nolint:exhaustruct // example
			ID:          "require-readme",
			Name:        "require README.md",
			Description: "every project should ship a README.md at its root",
			Cat:         linter.CategoryStructure,
			Sev:         finding.SeverityWarning,
		},
		Run: func(_ context.Context, d string) ([]finding.Finding, error) {
			if _, err := os.Stat(filepath.Join(d, "README.md")); err == nil {
				return nil, nil
			}

			return []finding.Finding{
				finding.NewBuilder(
					"require-readme", "minimal-linter",
					"README.md is missing",
					finding.SeverityWarning,
					finding.Pos(finding.FilePath("README.md"), 0, 0),
				).
					WithConfidence(finding.ConfidenceHigh).
					WithFixStrategy(finding.FixStrategySuggest).
					WithSuggestion("Create a README.md describing the project.").
					MustBuild(),
			}, nil
		},
	})

	report, err := registry.Run(context.Background(), dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "minimal-linter: %v\n", err)

		os.Exit(2)
	}

	for _, f := range report.ActiveFindings() {
		fmt.Printf("%s:%d: %s\n", f.Position.File, f.Position.Line, f.Message)
	}

	os.Exit(linter.ExitCodeFromReport(report))
}
