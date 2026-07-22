package linter

import (
	"errors"
	"fmt"
)

// ErrRuleFailed is the sentinel error for a rule execution failure.
// Use errors.Is(err, linter.ErrRuleFailed) to check whether a Registry.Run
// or DetectorFromRegistry call failed because a rule returned an error.
var ErrRuleFailed = errors.New("linter: rule failed")

// RuleError wraps an error from a specific rule, preserving the rule name so
// callers can identify which rule in a registry failed. Returned by
// Registry.Run, RuleFunc.Check, and DetectorFromRegistry when a rule's
// execution fails.
//
// Use errors.As to extract the rule name:
//
//	var ruleErr *linter.RuleError
//	if errors.As(err, &ruleErr) {
//	    log.Printf("rule %s failed", ruleErr.RuleName)
//	}
type RuleError struct {
	RuleName string
	Cause    error
}

// Error implements error.
func (e *RuleError) Error() string {
	return fmt.Sprintf("rule %q failed: %v", e.RuleName, e.Cause)
}

// Unwrap returns the underlying cause for errors.Unwrap / errors.Is.
func (e *RuleError) Unwrap() error {
	return e.Cause
}

// Is supports errors.Is against ErrRuleFailed.
func (*RuleError) Is(target error) bool {
	return target == ErrRuleFailed
}

// NewRuleError wraps cause as a RuleError for the named rule.
func NewRuleError(ruleName string, cause error) *RuleError {
	return &RuleError{
		RuleName: ruleName,
		Cause:    cause,
	}
}
