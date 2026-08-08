package linter_test

import (
	"context"
	"errors"
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
