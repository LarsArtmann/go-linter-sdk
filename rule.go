// Package linter provides the shared scaffolding for LarsArtmann Go linters
// (branching-flow, erraudit, go-structure-linter, and future
// additions).
//
// Each of those linters independently reinvented the same three layers:
//   - A Rule interface (Name/Description/Severity/Check)
//   - A registry of rules
//   - A converter from their native domain type to finding.Finding
//
// The last layer — the converter — is the most expensive duplication:
// branching-flow ships 1,871 LOC of converters, erraudit 1,214 LOC,
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

// Category values classify what kind of issue a rule detects, for filtering
// and reporting. Each maps to finding.Category at the finding boundary.
//
// These are RECOMMENDED values, not a closed set. Define your own for
// domain-specific taxonomies:
//
//	const CategoryAPI Category = "api" // your domain category
const (
	CategoryDesign        Category = "design"    // design smells (coupling, cohesion)
	CategoryStructure     Category = "structure" // file/package layout
	CategoryErrorHandling Category = "error-handling"
	CategoryCorrectness   Category = "correctness"
	CategoryStyle         Category = "style"
	CategoryPerformance   Category = "performance"
	CategorySecurity      Category = "security"
	CategoryConfiguration Category = "configuration"
)

// Rule is a single lint check. A linter is a collection of rules. Each rule
// declares its identity (ID/Name/Description/Category/Severity) and a Check
// function that emits findings for a target directory.
//
// # Dual identity: ID vs Name
//
// Every rule has two identity fields with distinct stability contracts:
//
//   - ID() is the STABLE identifier. It never changes once published. Used for
//     registry deduplication, suppression matching (//tool:ignore ID), filter
//     config (--enable ID), and the finding.RuleName field. Think "G001",
//     "C001", "no-unused-vars".
//   - Name() is the DISPLAY name. Human-readable, can change across versions
//     without breaking suppression, filtering, or health-score computation.
//     Think "missing transaction commit", "unused variable".
//
// Rules emit finding.Finding DIRECTLY — there is no intermediate Violation or
// Issue type to convert. This is the key design decision that eliminates the
// thousands of LOC of converter code the existing linters carry. Use
// finding.NewBuilder(...) to construct each finding.
//
// # Building findings with Confidence and FixStrategy
//
// finding.Finding already carries Confidence (finding.ConfidenceLow/Medium/High)
// and FixStrategy (finding.FixStrategyNone/Suggest/Direct/AI) at the per-finding
// level. Set these on individual findings via finding.NewBuilder(...):
//
//	finding.NewBuilder(ruleID, toolName, msg, sev, pos).
//	    WithConfidence(finding.ConfidenceHigh).
//	    WithFixStrategy(finding.FixStrategyDirect).
//	    MustBuild()
//
// Per-finding values are strictly more expressive than rule-level defaults: a
// single rule can emit findings with different confidence levels and fix
// strategies depending on the matched pattern. RuleMeta.Sev is the default
// severity — individual findings can override it via the Builder.
type Rule interface {
	ID() string
	Name() string
	Description() string
	Category() Category
	Severity() finding.Severity
	IsEnabledByDefault() bool
	Check(ctx context.Context, dir string) ([]finding.Finding, error)
}

// RuleFunc adapts a function to the Rule interface, filling in the identity
// fields from the supplied RuleMeta. This is the common case: most rules are a
// metadata header plus a check closure.
type RuleFunc struct {
	Meta RuleMeta
	Run  func(ctx context.Context, dir string) ([]finding.Finding, error)
}

// RuleMeta is the declarative identity of a rule: ID, name, description,
// category, severity. Supplied as a struct literal so rules read like data.
//
// ID is REQUIRED — it is the stable identifier that never changes once
// published. It is used for registry deduplication, suppression matching,
// filter config, and the finding.RuleName field.
type RuleMeta struct {
	ID          string
	Name        string
	Description string
	Cat         Category
	Sev         finding.Severity
}

// optInRule is a RuleFunc with default-enabled=false. Created via OptIn().
// Embeds RuleFunc so all methods (ID/Name/Description/Category/Severity/Check)
// are inherited; only IsEnabledByDefault is overridden.
type optInRule struct {
	RuleFunc
}

func (optInRule) IsEnabledByDefault() bool { return false }

// OptIn returns a Rule that is disabled by default — it only runs when a
// consumer explicitly enables it (e.g. via --enable <id>). Use for rules
// that are noisy, experimental, or domain-specific:
//
//	r.Register(linter.OptIn(linter.RuleFunc{Meta: ..., Run: ...}))
func OptIn(rf RuleFunc) Rule { return optInRule{rf} }

// ID implements Rule. Returns the stable identifier from RuleMeta.ID.
func (r RuleFunc) ID() string { return r.Meta.ID }

// Name implements Rule.
func (r RuleFunc) Name() string { return r.Meta.Name }

// Description implements Rule.
func (r RuleFunc) Description() string { return r.Meta.Description }

// Category implements Rule.
func (r RuleFunc) Category() Category { return r.Meta.Cat }

// Severity implements Rule.
func (r RuleFunc) Severity() finding.Severity { return r.Meta.Sev }

// IsEnabledByDefault returns true for RuleFunc. Rules created via RuleFunc are
// enabled by default — they run unless a consumer explicitly disables them.
// For opt-in rules, use OptIn().
func (RuleFunc) IsEnabledByDefault() bool { return true }

// Check implements Rule. If Run returns an error, it is wrapped into a
// *RuleError carrying the rule's stable ID, so callers of Registry.Run and
// DetectorFromRegistry can identify which rule failed via errors.As.
func (r RuleFunc) Check(ctx context.Context, dir string) ([]finding.Finding, error) {
	findings, err := r.Run(ctx, dir)
	if err != nil {
		return findings, NewRuleError(r.Meta.ID, err)
	}

	return findings, nil
}
