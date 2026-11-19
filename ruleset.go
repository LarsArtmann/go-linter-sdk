package linter

import (
	"context"
	"slices"

	"github.com/larsartmann/go-finding"
)

// RuleSet is an immutable, order-preserving collection of rules with O(1)
// ID lookup. It is the mutex-free alternative to [Registry] for consumers
// that assemble a fixed rule set once and never mutate it: no locks, no
// Register/Deregister surface, and the same identity validation
// (panics on duplicate IDs or empty identity fields, matching [Registry]).
//
// RuleSet holds data, not execution. Run it by converting to a Registry
// ([RuleSet.ToRegistry]) or by feeding [Registry.Run] /
// [DetectorsFromRegistry] — the execution, coverage, and Relativize logic
// lives in exactly one place.
//
//	sel := linter.FilterRules(allRules, enableSet, disableSet)
//	set := linter.NewRuleSet(toRuleSlice(sel)) // []linter.Rule
type RuleSet struct {
	rules []Rule
	byID  map[string]Rule
}

// NewRuleSet builds a RuleSet from rules, in slice order. Panics on
// duplicate rule IDs or empty identity fields — the same programming-error
// contract as [Registry.Register], so a bad rule surfaces at construction
// regardless of which container it lands in. The input slice is copied;
// later mutations of the caller's slice do not affect the set.
func NewRuleSet(rules []Rule) *RuleSet {
	set := &RuleSet{
		rules: make([]Rule, 0, len(rules)),
		byID:  make(map[string]Rule, len(rules)),
	}

	for _, rule := range rules {
		if err := validateRuleIdentity(rule); err != nil {
			panic(err)
		}

		id := rule.ID()
		if _, exists := set.byID[id]; exists {
			panic("linter: duplicate rule ID " + id)
		}

		set.rules = append(set.rules, rule)
		set.byID[id] = rule
	}

	return set
}

// Rules returns a copy of the rules in construction order.
func (s *RuleSet) Rules() []Rule {
	out := make([]Rule, len(s.rules))
	copy(out, s.rules)

	return out
}

// Len returns the number of rules in the set.
func (s *RuleSet) Len() int {
	return len(s.rules)
}

// Get returns the rule with the given ID and true, or nil and false.
func (s *RuleSet) Get(id string) (Rule, bool) {
	rule, ok := s.byID[id]

	return rule, ok
}

// Has reports whether a rule with the given ID is in the set.
func (s *RuleSet) Has(id string) bool {
	_, ok := s.byID[id]

	return ok
}

// ToRegistry converts the set into a mutable [Registry], preserving rule
// order and applying the given options. The registry starts from the set's
// rules but diverges on mutation: Register/Deregister on the registry never
// affect the set. Each Check call still resolves identity through the rule
// itself, so sharing rules between a set and its registry is safe.
func (s *RuleSet) ToRegistry(opts ...RegistryOption) *Registry {
	return NewRegistryFromRules(slices.Clone(s.rules), opts...)
}

// Check runs every rule in the set against dir — the mutex-free execution
// path for immutable sets. It applies the same normalization as
// [Registry.Run]'s default (fail-fast): findings pass through
// finding.Relativize and the first rule error aborts, wrapped as
// [*RuleError]. Callers that need a *finding.Report with coverage evidence
// should use ToRegistry + Run instead.
func (s *RuleSet) Check(ctx context.Context, dir string) ([]finding.Finding, error) {
	var all []finding.Finding

	for _, rule := range s.rules {
		findings, err := rule.Check(ctx, dir)
		if err != nil {
			return nil, wrapRuleError(rule.ID(), err)
		}

		all = append(all, finding.Relativize(findings, dir)...)
	}

	return all, nil
}
