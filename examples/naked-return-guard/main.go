// Command naked-return-guard is a pilot port of branching-flow's
// pkg/nakedreturn rule, rewritten to use go-linter-sdk. It validates the
// converter-deletion claim for the branching-flow analyzer family:
//
//	Original (branching-flow): analyzer.go 106 LOC + types.go 17 LOC
//	+ to_findings.go converter 33 LOC + astutil dependencies — the Detection
//	type and DetectionsToFindings converter existed only to bridge to
//	finding.Finding.
//
//	Ported (go-linter-sdk, self-contained): one RuleFunc whose Run closure
//	emits finding.Finding directly. No Detection type, no converter.
//
// Detection logic is a faithful port: flag naked `return` statements
// (return with zero result expressions) inside functions that declare named
// results and whose body spans more than 10 lines. Test files and vendor
// trees are skipped, matching the original's astutil.ShouldSkipFile.
//
// Run:
//
//	GOEXPERIMENT=jsonv2 go run ./examples/naked-return-guard [dir]
package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-linter-sdk"
)

const (
	ruleID   = "NAKED_RETURN_GUARD" // original branching-flow rule ID, for suppression compatibility
	toolName = "branching-flow"
	maxLines = 10
)

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	registry := linter.NewRegistry()

	registry.Register(linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          ruleID,
			Name:        "naked return in long function",
			Description: "detects naked return statements in functions longer than 10 lines with named results",
			Cat:         linter.CategoryStyle,
			Sev:         finding.SeverityInfo,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, target string) ([]finding.Finding, error) {
			return checkNakedReturns(target)
		},
	})

	report, err := registry.Run(context.Background(), dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", ruleID, err)

		os.Exit(2)
	}

	for _, f := range report.ActiveFindings() {
		fmt.Printf("%s:%d: %s\n", f.Position.File, f.Position.Line, f.Message)
	}

	os.Exit(linter.ExitCodeFromReport(report))
}

// checkNakedReturns is the faithful port of branching-flow's nakedreturn
// Analyzer.AnalyzeFile loop plus DetectionsToFindings emission.
func checkNakedReturns(dir string) ([]finding.Finding, error) {
	findings := make([]finding.Finding, 0)

	fset := token.NewFileSet()

	if walkErr := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() == "vendor" && path != dir {
				return filepath.SkipAll
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return fmt.Errorf("parse %s: %w", path, parseErr)
		}

		findings = append(findings, scanFile(fset, file, path)...)

		return nil
	}); walkErr != nil {
		return nil, fmt.Errorf("walk %s: %w", dir, walkErr)
	}

	return findings, nil
}

// scanFile inspects one parsed file for naked returns in long functions.
func scanFile(fset *token.FileSet, file *ast.File, path string) []finding.Finding {
	findings := make([]finding.Finding, 0)

	ast.Inspect(file, func(n ast.Node) bool {
		fnDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if fnDecl.Body == nil || !hasNamedResults(fnDecl) {
			return false
		}

		bodyStart := fset.Position(fnDecl.Body.Pos()).Line
		bodyLen := fset.Position(fnDecl.Body.End()).Line - bodyStart

		if bodyLen <= maxLines {
			return false
		}

		ast.Inspect(fnDecl.Body, func(inner ast.Node) bool {
			retStmt, ok := inner.(*ast.ReturnStmt)
			if !ok || len(retStmt.Results) != 0 {
				return true
			}

			line := fset.Position(retStmt.Pos()).Line

			findings = append(findings, newFinding(path, line, fnDecl.Name.Name, bodyLen))

			return true
		})

		return false // body already inspected; Go has no nested FuncDecls
	})

	return findings
}

// newFinding builds the finding that DetectionsToFindings produced in the
// original: same ID, message, severity (Info), confidence (0.85), and
// metadata keys.
func newFinding(path string, line int, function string, functionLen int) finding.Finding {
	return finding.NewBuilder(
		ruleID, toolName,
		fmt.Sprintf("naked return in %s() (%d lines), use explicit return values", function, functionLen),
		finding.SeverityInfo,
		finding.Pos(finding.FilePath(path), line, 0),
	).
		WithConfidence(0.85).
		WithFixStrategy(finding.FixStrategySuggest).
		WithSuggestion(fmt.Sprintf("Replace naked return in %s() with explicit return values", function)).
		WithMetadata(map[string]string{
			"filePath":    path,
			"line":        strconv.Itoa(line),
			"function":    function,
			"functionLen": strconv.Itoa(functionLen),
		}).
		MustBuild()
}

// hasNamedResults reports whether the function declares named result values —
// the precondition that makes a naked return meaningful.
func hasNamedResults(fnDecl *ast.FuncDecl) bool {
	if fnDecl.Type.Results == nil {
		return false
	}

	for _, field := range fnDecl.Type.Results.List {
		if len(field.Names) > 0 {
			return true
		}
	}

	return false
}
