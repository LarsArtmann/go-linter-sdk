package linter

import (
	"errors"
	"fmt"
)

// ErrRuleFailed is the sentinel error for a rule execution failure.
// Use errors.Is(err, linter.ErrRuleFailed) to check whether a Registry.Run
// or DetectorFromRegistry call failed because a rule returned an error.
var ErrRuleFailed = errors.New("linter: rule failed")

// RuleError wraps an error from a specific rule, preserving the rule's stable
// ID so callers can identify which rule in a registry failed. Returned by
// Registry.Run, RuleFunc.Check, DetectorFromRegistry, and DetectorsFromRegistry
// when a rule's execution fails.
//
// Use errors.AsType (Go 1.26+) to extract the rule ID:
//
//	ruleErr, ok := errors.AsType[*linter.RuleError](err)
//	if ok {
//	    log.Printf("rule %s failed", ruleErr.RuleID)
//	}
//
// To check whether any rule failed (regardless of which), match the sentinel:
//
//	if errors.Is(err, linter.ErrRuleFailed) {
//	    // a rule returned an error
//	}
type RuleError struct {
	RuleID string
	Cause  error
}

// Error implements error.
func (e *RuleError) Error() string {
	return fmt.Sprintf("rule %q failed: %v", e.RuleID, e.Cause)
}

// Unwrap returns the underlying cause for errors.Unwrap / errors.Is.
func (e *RuleError) Unwrap() error {
	return e.Cause
}

// Is supports errors.Is against ErrRuleFailed.
func (*RuleError) Is(target error) bool {
	return target == ErrRuleFailed
}

// NewRuleError wraps cause as a RuleError for the rule identified by ruleID.
func NewRuleError(ruleID string, cause error) *RuleError {
	return &RuleError{
		RuleID: ruleID,
		Cause:  cause,
	}
}
