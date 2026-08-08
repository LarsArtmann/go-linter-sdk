package linter_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/larsartmann/go-linter-sdk"
)

func TestNewRuleError_NilCause(t *testing.T) {
	t.Parallel()

	ruleErr := linter.NewRuleError("my-rule", nil)

	msg := ruleErr.Error()
	if !strings.Contains(msg, "my-rule") {
		t.Errorf("expected rule ID in message, got %q", msg)
	}

	if ruleErr.Unwrap() != nil {
		t.Errorf("expected Unwrap to return nil for nil cause")
	}

	if !errors.Is(ruleErr, linter.ErrRuleFailed) {
		t.Error("expected errors.Is(ruleErr, ErrRuleFailed) to be true")
	}
}

func TestRuleError_Is_Canceled(t *testing.T) {
	t.Parallel()

	ruleErr := linter.NewRuleError("ctx-rule", context.Canceled)

	if !errors.Is(ruleErr, context.Canceled) {
		t.Errorf("expected errors.Is(ruleErr, context.Canceled) to be true")
	}

	if !errors.Is(ruleErr, linter.ErrRuleFailed) {
		t.Error("expected errors.Is(ruleErr, ErrRuleFailed) to remain true")
	}
}

func TestRuleError_Is_DeadlineExceeded(t *testing.T) {
	t.Parallel()

	ruleErr := linter.NewRuleError("ctx-rule", context.DeadlineExceeded)

	if !errors.Is(ruleErr, context.DeadlineExceeded) {
		t.Errorf("expected errors.Is(ruleErr, context.DeadlineExceeded) to be true")
	}

	if !errors.Is(ruleErr, linter.ErrRuleFailed) {
		t.Error("expected errors.Is(ruleErr, ErrRuleFailed) to remain true")
	}
}

func FuzzNewRuleError(f *testing.F) {
	f.Add("test-rule", "something went wrong")
	f.Add("", "error with empty id")
	f.Add("nil-cause-rule", "")

	f.Fuzz(func(t *testing.T, ruleID, causeMsg string) {
		var cause error

		if causeMsg != "" {
			cause = errors.New(causeMsg)
		}

		ruleErr := linter.NewRuleError(ruleID, cause)

		// Error() must never panic, even with nil cause or empty ruleID.
		msg := ruleErr.Error()
		if msg == "" {
			t.Errorf("Error() returned empty string for ruleID=%q causeMsg=%q", ruleID, causeMsg)
		}

		// Every RuleError must match the ErrRuleFailed sentinel.
		if !errors.Is(ruleErr, linter.ErrRuleFailed) {
			t.Errorf("errors.Is(ruleErr, ErrRuleFailed) = false for ruleID=%q", ruleID)
		}

		// The cause must be reachable through the unwrap chain.
		if cause != nil && !errors.Is(ruleErr, cause) {
			t.Errorf("cause unreachable for ruleID=%q causeMsg=%q", ruleID, causeMsg)
		}
	})
}

func TestRuleErrors_Single(t *testing.T) {
	t.Parallel()

	ruleErr := linter.NewRuleError("my-rule", errors.New("boom"))
	result := linter.RuleErrors(ruleErr)

	if len(result) != 1 {
		t.Fatalf("expected 1 rule error, got %d", len(result))
	}

	if result[0].RuleID != "my-rule" {
		t.Errorf("expected rule ID 'my-rule', got %q", result[0].RuleID)
	}
}

func TestRuleErrors_JoinedMultiple(t *testing.T) {
	t.Parallel()

	err1 := linter.NewRuleError("rule-a", errors.New("fail-a"))
	err2 := linter.NewRuleError("rule-b", errors.New("fail-b"))
	joined := errors.Join(err1, err2)

	result := linter.RuleErrors(joined)

	if len(result) != 2 {
		t.Fatalf("expected 2 rule errors, got %d", len(result))
	}

	if result[0].RuleID != "rule-a" || result[1].RuleID != "rule-b" {
		t.Errorf("expected rule-a, rule-b; got %s, %s", result[0].RuleID, result[1].RuleID)
	}
}

func TestRuleErrors_JoinedMixed(t *testing.T) {
	t.Parallel()

	ruleErr := linter.NewRuleError("rule-x", errors.New("fail"))
	other := errors.New("unrelated error")
	joined := errors.Join(ruleErr, other)

	result := linter.RuleErrors(joined)

	if len(result) != 1 {
		t.Fatalf("expected 1 rule error from mixed join, got %d", len(result))
	}

	if result[0].RuleID != "rule-x" {
		t.Errorf("expected rule-x, got %q", result[0].RuleID)
	}
}

func TestRuleErrors_Nil(t *testing.T) {
	t.Parallel()

	if result := linter.RuleErrors(nil); result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}
}

func TestRuleErrors_NonRuleError(t *testing.T) {
	t.Parallel()

	if result := linter.RuleErrors(errors.New("plain error")); result != nil {
		t.Errorf("expected nil for non-RuleError, got %v", result)
	}
}

func TestRuleErrors_DeeplyNested(t *testing.T) {
	t.Parallel()

	ruleErr := linter.NewRuleError("deep-rule", errors.New("root cause"))
	wrapped := fmt.Errorf("outer: %w", fmt.Errorf("middle: %w", ruleErr))

	result := linter.RuleErrors(wrapped)

	if len(result) != 1 {
		t.Fatalf("expected 1 rule error through nested wrapping, got %d", len(result))
	}

	if result[0].RuleID != "deep-rule" {
		t.Errorf("expected 'deep-rule', got %q", result[0].RuleID)
	}
}

func TestErrMissingFields(t *testing.T) {
	t.Parallel()

	err := linter.RuleMeta{
		Name:        "X",
		Description: "desc",
		Cat:         linter.CategoryStyle,
	}.Validate()

	if !errors.Is(err, linter.ErrMissingFields) {
		t.Errorf("expected errors.Is(err, ErrMissingFields) to be true")
	}

	validErr := linter.RuleMeta{
		ID:          "x",
		Name:        "X",
		Description: "desc",
		Cat:         linter.CategoryStyle,
	}.Validate()

	if errors.Is(validErr, linter.ErrMissingFields) {
		t.Errorf("expected valid meta to not match ErrMissingFields")
	}
}
