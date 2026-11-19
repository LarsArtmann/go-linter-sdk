# ADR 0003 — RuleFunc generics: declined

**Status:** Declined (2026-11-19, plan task M20 / F20.1–F20.2)

**Context:** ROADMAP theme 1 asked whether `RuleFunc` should use generics
for type-safe rule definitions. The compile-checked sketch
(`TypedRule[State]` with a closure adapter back to `RuleFunc`) is in the
session log; the conclusion is recorded here so the question stays answered.

## What generics could buy

A `TypedRule[State any]` would give `Run` a typed state parameter instead of
a closure over local variables:

```go
type TypedRule[State any] struct {
	Meta linter.RuleMeta
	Run  func(ctx context.Context, dir string, state State) ([]finding.Finding, error)
}
```

## Why it is declined

1. **No type unsafety exists to fix.** Today's rules carry state via
   closures (`var rule linter.RuleFunc` + captured locals), which the
   compiler already type-checks completely. The sketch compiles only by
   erasing the type parameter at the `Rule` boundary (`adaptRule` returns a
   plain `RuleFunc`) — the registry stores `[]Rule`, so the type parameter
   buys nothing past construction time.
2. **Two parallel rule types** (`RuleFunc` and `TypedRule`) means every
   helper (`OptIn`, `FilterRules`, `NewFinding`, auto-stamping, docs,
   examples) needs a generic twin or a bridging doc. The SDK's one-job API
   (`RuleFunc` reads like data) is its main teaching asset.
3. **Consumer churn for zero observable gain:** six direct consumers and the
   examples would face a migration whose end state compiles, runs, and looks
   the same as today's closures.
4. **ireturn/`exhaustruct` lint posture** is built around the concrete
   `RuleFunc`; a generic twin reopens every one of those decisions.

## Reopen trigger

A concrete consumer need where closure-carried state demonstrably hurts —
e.g. a linter with many rules sharing one heavy parsed context AND a
requirement to construct rules from data tables where closures are awkward.
Even then, the lighter fix is a constructor returning `RuleFunc` closures
over the shared context, which stays available today without any API move.
