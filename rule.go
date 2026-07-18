// Package linter provides the shared scaffolding for LarsArtmann Go linters
// (branching-flow, hierarchical-errors, go-structure-linter, and future
// additions).
//
// Each of those linters independently reinvented the same three layers:
//   - A Rule interface (Name/Description/Severity/Check)
//   - A registry of rules
//   - A converter from their native domain type to finding.Finding
//
// The last layer — the converter — is the most expensive duplication:
// branching-flow ships 1,871 LOC of converters, hierarchical-errors 1,214 LOC,
// just to bridge their Violation/ErrorViolation types to finding.Finding.
// go-structure-linter eliminated this entirely by aliasing
// `type Issue = finding.Finding`.
//
// This SDK codifies the go-structure-linter pattern: a Rule emits finding.Finding
// directly, so there is NO converter layer. A linter that adopts this SDK ships
// a rules.go file and a main.go one-liner — the registry, detection, and exit
// codes are shared.
//
// Design constraint: depends only on go-finding (the ecosystem hub) so any
// linter — CLI, library, LSP server, golangci plugin — can adopt it.
package linter

import (
	"context"

	"github.com/larsartmann/go-finding"
)

// Category classifies what kind of issue a rule detects, for filtering and
// reporting. Maps to finding.Category at the finding boundary.
type Category string

const (
	CategoryDesign       Category = "design"       // design smells (coupling, cohesion)
	CategoryStructure    Category = "structure"    // file/package layout
	CategoryErrorHandling Category = "error-handling"
	CategoryCorrectness  Category = "correctness"
	CategoryStyle        Category = "style"
	CategoryPerformance  Category = "performance"
	CategorySecurity     Category = "security"
	CategoryConfiguration Category = "configuration"
)

// Rule is a single lint check. A linter is a collection of rules. Each rule
// declares its identity (Name/Description/Category/Severity) and a Check
// function that emits findings for a target directory.
//
// Rules emit finding.Finding DIRECTLY — there is no intermediate Violation or
// Issue type to convert. This is the key design decision that eliminates the
// thousands of LOC of converter code the existing linters carry. Use
// finding.NewBuilder(...) to construct each finding.
type Rule interface {
	Name() string
	Description() string
	Category() Category
	Severity() finding.Severity
	Check(ctx context.Context, dir string) ([]finding.Finding, error)
}

// RuleFunc adapts a function to the Rule interface, filling in the identity
// fields from the supplied RuleMeta. This is the common case: most rules are a
// metadata header plus a check closure.
type RuleFunc struct {
	Meta RuleMeta
	Run  func(ctx context.Context, dir string) ([]finding.Finding, error)
}

// RuleMeta is the declarative identity of a rule: name, description, category,
// severity. Supplied as a struct literal so rules read like data.
type RuleMeta struct {
	Name        string
	Description string
	Cat         Category
	Sev         finding.Severity
}

// Name implements Rule.
func (r RuleFunc) Name() string { return r.Meta.Name }

// Description implements Rule.
func (r RuleFunc) Description() string { return r.Meta.Description }

// Category implements Rule.
func (r RuleFunc) Category() Category { return r.Meta.Cat }

// Severity implements Rule.
func (r RuleFunc) Severity() finding.Severity { return r.Meta.Sev }

// Check implements Rule.
func (r RuleFunc) Check(ctx context.Context, dir string) ([]finding.Finding, error) {
	return r.Run(ctx, dir)
}
