// Command cli-triage demonstrates the discovery and triage capabilities that
// no other example covers:
//
//   - FilterRules (--enable/--disable set logic)
//   - OptIn rules (disabled by default; registered only when explicitly enabled)
//   - ExitCodeByConfidence (tiered exit codes)
//
// Exit contract:
//
//	0 clean · 1 must-fix findings · 2 triage-only findings · 3 tool failure
//
// The tiers come from ExitCodeByConfidence: high-confidence findings gate at 1,
// low-confidence findings route to triage at 2. Exit 3 is reserved for linter
// execution errors, so CI can distinguish "findings below threshold" from
// "the tool crashed".
//
// The linter ships two rules: a default-enabled "long-func" check whose
// findings are high-confidence, and an opt-in "no-todos" check whose findings
// are deliberately low-confidence so they route to triage, never gate.
//
// Run:
//
//	go run ./examples/cli-triage [dir] [--enable=r1,r2] [--disable=r3] [--help]
//
// --enable/--disable accept comma-separated rule IDs. --help prints usage and
// exits 0.
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
	toolName = "cli-triage"

	exitClean       = 0
	exitMustFix     = 1
	exitTriage      = 2
	exitToolFailure = 3

	longFuncID    = "long-func"
	noTodosID     = "no-todos"
	longFuncLines = 40
)

var usage = `cli-triage — discovery and triage demo linter

Usage:
  cli-triage [dir] [flags]

Flags:
  --enable=r1,r2   run only these rules (comma-separated IDs)
  --disable=r1,r2  never run these rules (comma-separated IDs)
  -h, --help       show this help and exit

Rules:
  long-func  default on    flags functions longer than ` + strconv.Itoa(longFuncLines) + ` lines
  no-todos   opt-in        flags TODO comments (triage-only by confidence)

Exit codes:
  0 clean · 1 must-fix findings · 2 triage-only findings · 3 tool failure
`

func main() {
	var dir string

	enable := map[string]bool{}
	disable := map[string]bool{}

	for _, arg := range os.Args[1:] {
		switch {
		case arg == "--help" || arg == "-h":
			fmt.Print(usage)

			os.Exit(exitClean)
		case strings.HasPrefix(arg, "--enable="):
			for id := range strings.SplitSeq(strings.TrimPrefix(arg, "--enable="), ",") {
				enable[id] = true
			}
		case strings.HasPrefix(arg, "--disable="):
			for id := range strings.SplitSeq(strings.TrimPrefix(arg, "--disable="), ",") {
				disable[id] = true
			}
		default:
			dir = arg
		}
	}

	if dir == "" {
		dir = "."
	}

	registry := linter.NewRegistry(linter.WithToolName(toolName))

	// SelectRules works on []Rule, so the opt-in rule rides through already
	// wrapped in OptIn. OptIn rules carry IsEnabledByDefault() == false, so
	// the gate below is generic: skip any opt-in rule that was not explicitly
	// enabled, no per-rule ID checks needed.
	for _, rule := range linter.SelectRules(allRules(), enable, disable) {
		if !rule.IsEnabledByDefault() && !enable[rule.ID()] {
			continue
		}

		registry.Register(rule)
	}

	report, err := registry.Run(context.Background(), dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", toolName, err)

		os.Exit(exitToolFailure)
	}

	for _, f := range report.ActiveFindings() {
		fmt.Printf("%s:%d: [%s] %s\n", f.Position.File, f.Position.Line, f.Rule, f.Message)
	}

	code := linter.ExitCodeByConfidence(report, finding.ConfidenceHigh)
	fmt.Fprintf(os.Stderr, "%s: %d finding(s), exit %d (0 clean, 1 must fix, 2 triage, 3 tool failure)\n",
		toolName, report.Len(), code)

	os.Exit(code)
}

// allRules returns every rule the linter ships, in registration order, as the
// general Rule interface: the opt-in rule is already wrapped in OptIn, so
// SelectRules and the registration gate can key off IsEnabledByDefault.
func allRules() []linter.Rule {
	var longFunc linter.RuleFunc

	longFunc = linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          longFuncID,
			Name:        "long function",
			Description: "flags functions whose body spans more than 40 lines",
			Cat:         linter.CategoryDesign,
			Sev:         finding.SeverityWarning,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			return findLongFuncs(longFunc, dir)
		},
	}

	var noTodos linter.RuleFunc

	noTodos = linter.RuleFunc{
		Meta: linter.RuleMeta{
			ID:          noTodosID,
			Name:        "TODO comment",
			Description: "flags TODO comments; noisy by design, so opt-in",
			Cat:         linter.CategoryDocumentation,
			Sev:         finding.SeverityInfo,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, dir string) ([]finding.Finding, error) {
			return findTODOs(noTodos, dir)
		},
	}

	return []linter.Rule{longFunc, linter.OptIn(noTodos)}
}

// findLongFuncs walks dir's Go files and flags oversized function bodies.
// Findings are high-confidence: ExitCodeByConfidence routes them to "must fix".
func findLongFuncs(rule linter.RuleFunc, dir string) ([]finding.Finding, error) {
	return walkGoFiles(dir, func(fset *token.FileSet, file *ast.File, path string) []finding.Finding {
		var out []finding.Finding

		ast.Inspect(file, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok || funcDecl.Body == nil {
				return true
			}

			start := fset.Position(funcDecl.Body.Pos()).Line
			length := fset.Position(funcDecl.Body.End()).Line - start

			if length <= longFuncLines {
				return true
			}

			out = append(out, rule.NewFinding(
				fmt.Sprintf("%s() spans %d lines; split it up", funcDecl.Name.Name, length),
				finding.Pos(finding.FilePath(path), start, 0),
			).
				WithConfidence(finding.ConfidenceHigh).
				WithFixStrategy(finding.FixStrategySuggest).
				WithSuggestion(fmt.Sprintf("Extract blocks of %s() into named functions", funcDecl.Name.Name)).
				MustBuild())

			return true
		})

		return out
	})
}

// findTODOs flags TODO comments in Go files. Findings are deliberately
// low-confidence so ExitCodeByConfidence routes them to triage, not gate.
func findTODOs(rule linter.RuleFunc, dir string) ([]finding.Finding, error) {
	return walkGoFiles(dir, func(fset *token.FileSet, file *ast.File, path string) []finding.Finding {
		var out []finding.Finding

		for _, comment := range file.Comments {
			text := comment.Text()
			if !strings.Contains(text, "TODO") {
				continue
			}

			pos := fset.Position(comment.Pos())

			out = append(out, rule.NewFinding(
				"TODO comment found: "+firstLine(text),
				finding.Pos(finding.FilePath(path), pos.Line, 0),
			).
				WithConfidence(finding.ConfidenceLow).
				WithFixStrategy(finding.FixStrategyNone).
				MustBuild())
		}

		return out
	})
}

// walkGoFiles parses every non-test, non-vendor Go file under dir and calls
// visit per parsed file.
func walkGoFiles(
	dir string,
	visit func(*token.FileSet, *ast.File, string) []finding.Finding,
) ([]finding.Finding, error) {
	var findings []finding.Finding

	fset := token.NewFileSet()

	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if entry.Name() == "vendor" && path != dir {
				return filepath.SkipAll
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if parseErr != nil {
			return fmt.Errorf("parse %s: %w", path, parseErr)
		}

		findings = append(findings, visit(fset, file, path)...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", dir, err)
	}

	return findings, nil
}

func firstLine(text string) string {
	if head, _, found := strings.Cut(text, "\n"); found {
		return strings.TrimSpace(head)
	}

	return strings.TrimSpace(text)
}
