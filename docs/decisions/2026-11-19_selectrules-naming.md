# ADR 0004 — `SelectRules` naming (the plan's `FilterRulesRules` rejected)

**Status:** Accepted (2026-11-19)
**Context:** Post-v0.5.0 Pareto plan task P10 needed a `[]Rule` variant of
`FilterRules` (which takes `[]RuleFunc`) so opt-in rules — whose `OptIn`
wrapper type is unexported — can participate in `--enable`/`--disable` set
logic. The plan's literal name for it was `FilterRulesRules`.

## Decision

**Name it `SelectRules(all []Rule, enable, disable map[string]bool) []Rule`.**

`FilterRulesRules` was rejected at the naming gate: it reads like a typo
(the doubled `Rules` is a stutter, and to a fresh reader the distinction
"rules-rules vs plain rules" carries no information). `SelectRules` states
what the function does — select a subset — while the signature carries the
type distinction. `Filter` was already taken by the finding-level gate
(`filter.go`), and `SelectRules` mirrors it cleanly: filter/select choose
rules; Filter chooses findings.

## Consequences

- The rule-choice pair reads as `FilterRules` (`[]RuleFunc`, convenience
  path) and `SelectRules` (`[]Rule`, OptIn-safe path); godoc cross-links
  both ways.
- The registration gate keys off `IsEnabledByDefault()` instead of hardcoded
  rule IDs (the cli-triage pattern).
- Recorded in CHANGELOG [0.6.0] and AGENTS.md; this ADR makes it durable
  against doc rewrites (f49 harvest of the 2026-11-19 status report §f).

## Revisit trigger

Only if a third rule-choice surface appears and the select/filter pair stops
covering the vocabulary.
