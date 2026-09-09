// Command oops-redundant-guard is a pilot port of erraudit's
// oops_redundant_guard rule, rewritten to use go-linter-sdk. It validates the
// converter-deletion claim for the error-handling domain:
//
//	Original (erraudit): pkg/oopspattern detection core (206 LOC) + the
//	go/analysis autofixer pkg/lint/oopsfix (102 LOC) + the violation detector
//	path in internal/ast + a Violation-to-finding bridge layer.
//
//	Ported (go-linter-sdk, self-contained): one RuleFunc whose Run closure
//	emits finding.Finding directly. No Violation type, no bridge.
//
// Detection logic is a faithful port: flag
//
//	if err != nil {
//	    return oops.Wrap(err)
//	}
//	return nil
//
// patterns. They are redundant because oops.Wrap/Wrapf return nil for nil
// input. The rule ID `oops_redundant_guard`, severity (warning), confidence
// (0.85), message, and suggestion text match the original for suppression and
// output compatibility.
//
// Run:
//
//	GOEXPERIMENT=jsonv2 go run ./examples/oops-redundant-guard [dir]
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
	ruleID   = "oops_redundant_guard" // original erraudit ErrorType, for suppression compatibility
	toolName = "erraudit"
	wrapName = "Wrap" // canonical oops method, fallback when the call shape is unexpected
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
			Name:        "redundant nil guard before oops.Wrap",
			Description: "detects `if err != nil { return oops.Wrap(err) }; return nil` guards that oops.Wrap's nil-for-nil contract makes redundant",
			Cat:         linter.CategoryErrorHandling,
			Sev:         finding.SeverityWarning,
			ToolName:    toolName,
		},
		Run: func(_ context.Context, target string) ([]finding.Finding, error) {
			return checkRedundantGuards(target)
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

// redundantGuardMatch describes a detected redundant nil guard:
//
//	if errName != nil { return oops.Wrap(errName) }
//	return nil
type redundantGuardMatch struct {
	ifStmt  *ast.IfStmt
	wrapRet *ast.ReturnStmt
	errName string
}

// checkRedundantGuards walks dir's Go files and emits one finding per match.
func checkRedundantGuards(dir string) ([]finding.Finding, error) {
	findings := make([]finding.Finding, 0)

	fset := token.NewFileSet()

	walkErr := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
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

		file, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return fmt.Errorf("parse %s: %w", path, parseErr)
		}

		ast.Inspect(file, func(n ast.Node) bool {
			block, ok := n.(*ast.BlockStmt)
			if !ok {
				return true
			}

			for _, m := range findRedundantGuards(block) {
				line := fset.Position(m.ifStmt.Pos()).Line
				findings = append(findings, newFinding(path, line, m))
			}

			return true
		})

		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walk %s: %w", dir, walkErr)
	}

	return findings, nil
}

// findRedundantGuards is the faithful port of erraudit's
// oopspattern.FindRedundantGuards: scan a block for
// `if errName != nil { return oops.Wrap*(errName) }` followed by `return nil`.
func findRedundantGuards(block *ast.BlockStmt) []redundantGuardMatch {
	var matches []redundantGuardMatch

	for i := 0; i+1 < len(block.List); i++ {
		ifStmt, ok := block.List[i].(*ast.IfStmt)
		if !ok {
			continue
		}

		errName := guardErrorName(ifStmt.Cond)
		if errName == "" || ifStmt.Else != nil {
			continue
		}

		ret := singleReturnFromBody(ifStmt.Body)
		if ret == nil || !isOopsWrapReturn(ret, errName) {
			continue
		}

		nextRet, ok := block.List[i+1].(*ast.ReturnStmt)
		if !ok || !isNilReturn(nextRet) {
			continue
		}

		matches = append(matches, redundantGuardMatch{
			ifStmt:  ifStmt,
			wrapRet: ret,
			errName: errName,
		})
	}

	return matches
}

// guardErrorName extracts the error variable name from `errName != nil`
// (either operand order). Returns "" for non nil-checks on error-like names.
func guardErrorName(cond ast.Expr) string {
	bin, ok := cond.(*ast.BinaryExpr)
	if !ok || bin.Op.String() != "!=" {
		return ""
	}

	if ident, ok := bin.X.(*ast.Ident); ok {
		if isNilIdent(bin.Y) && looksLikeErrorName(ident.Name) {
			return ident.Name
		}
	}

	if ident, ok := bin.Y.(*ast.Ident); ok {
		if isNilIdent(bin.X) && looksLikeErrorName(ident.Name) {
			return ident.Name
		}
	}

	return ""
}

// isNilIdent reports whether expr is the bare identifier `nil`.
func isNilIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)

	return ok && ident.Name == "nil"
}

// looksLikeErrorName reports whether name looks like an error variable
// (substring matching catches `statErr`, `myErr`, etc.).
func looksLikeErrorName(name string) bool {
	lower := strings.ToLower(name)

	return strings.Contains(lower, "err")
}

// singleReturnFromBody returns the sole ReturnStmt in a block, or nil.
func singleReturnFromBody(body *ast.BlockStmt) *ast.ReturnStmt {
	if body == nil || len(body.List) != 1 {
		return nil
	}

	ret, ok := body.List[0].(*ast.ReturnStmt)
	if !ok {
		return nil
	}

	return ret
}

// isOopsWrapReturn checks whether ret is `return oops.Wrap(errName)` or
// `return oops.WrapfN(errName, ...)` where the first argument matches errName.
func isOopsWrapReturn(ret *ast.ReturnStmt, errName string) bool {
	if len(ret.Results) != 1 {
		return false
	}

	call, ok := ret.Results[0].(*ast.CallExpr)
	if !ok {
		return false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok || pkgIdent.Name != "oops" {
		return false
	}

	if !isOopsWrapMethod(sel.Sel.Name) || len(call.Args) == 0 {
		return false
	}

	firstArg, ok := call.Args[0].(*ast.Ident)

	return ok && firstArg.Name == errName
}

// isOopsWrapMethod reports whether method is an oops.Wrap-family method name:
// Wrap, Wrapf, or Wrap/Wrapf followed by digits.
func isOopsWrapMethod(method string) bool {
	if method == "Wrap" || method == "Wrapf" {
		return true
	}

	// Check Wrapf before Wrap: "Wrapf2" has prefix "Wrap" but the remainder
	// "f2" is not all digits. Checking the longer prefix first avoids this.
	if strings.HasPrefix(method, "Wrapf") {
		return isWrapWithDigits(method, len("Wrapf"))
	}

	if strings.HasPrefix(method, "Wrap") {
		return isWrapWithDigits(method, len("Wrap"))
	}

	return false
}

// isWrapWithDigits reports whether the remainder after prefixLen is non-empty
// and all ASCII digits (the WrapN/WrapfN numbered family).
func isWrapWithDigits(method string, prefixLen int) bool {
	rest := method[prefixLen:]

	if rest == "" {
		return false
	}

	for _, c := range rest {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

// isNilReturn reports whether ret is `return nil`.
func isNilReturn(ret *ast.ReturnStmt) bool {
	if len(ret.Results) != 1 {
		return false
	}

	ident, ok := ret.Results[0].(*ast.Ident)

	return ok && ident.Name == "nil"
}

// wrapMethodName returns the oops method name from a wrap return statement,
// defaulting to "Wrap" when the shape is unexpected (mirrors erraudit).
func wrapMethodName(ret *ast.ReturnStmt) string {
	if len(ret.Results) == 0 {
		return wrapName
	}

	call, ok := ret.Results[0].(*ast.CallExpr)
	if !ok {
		return wrapName
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return wrapName
	}

	return sel.Sel.Name
}

// newFinding builds the finding the original violation detector produced:
// same ID, severity (warning), confidence (0.85), message, and suggestion.
func newFinding(path string, line int, match redundantGuardMatch) finding.Finding {
	method := wrapMethodName(match.wrapRet)

	message := fmt.Sprintf(
		"Redundant nil guard before oops.%s(%s) — Wrap/Wrapf return nil for nil input",
		method, match.errName,
	)

	suggestion := fmt.Sprintf(
		"Replace the if-block and `return nil` with a single `return oops.%s(%s)`",
		method, match.errName,
	)

	return finding.NewBuilder(
		ruleID, toolName, message,
		finding.SeverityWarning,
		finding.Pos(finding.FilePath(path), line, 0),
	).
		WithConfidence(0.85).
		WithFixStrategy(finding.FixStrategySuggest).
		WithSuggestion(suggestion).
		WithMetadata(map[string]string{
			"filePath": path,
			"line":     strconv.Itoa(line),
			"errName":  match.errName,
			"method":   method,
		}).
		MustBuild()
}
