# ADR 0001 — Registry.Register: panic vs error return

**Status:** Accepted (2026-09-09)
**Context:** Plan task Q9/Q10 of the 2026-09-09 execution plan. Register
currently panics on (a) duplicate rule ID and (b) empty required identity
fields (ID, Name, Description, Category). A question has been raised whether
Register should return an error instead, and whether changing it would be a
v0.4.0-breaking API movement.

## Arguments for panics (current behavior)

1. **Programming errors, not runtime conditions.** Duplicate IDs and missing
   identity fields are developer mistakes detectable at startup. Go's stdlib
   panics on exactly this class: `regexp.MustCompile`, `template.Must`,
   nil map writes. Fail-fast at registration beats a delayed error every
   caller must handle.
2. **The panics are unrecoverable by design.** If two rules share an ID,
   suppression matching and dedup are silently wrong — continuing is worse
   than crashing. An error return invites callers to log-and-continue, which
   produces a linter whose findings can't be suppressed reliably.
3. **Call-site ergonomics.** Rules register in `init()` or package-level
   constructors where error handling is boilerplate: every consumer would
   write `if err := reg.Register(r); err != nil { panic(err) }` — the panic
   just moves into user code.
4. **Escape hatch exists.** `RuleMeta.Validate()` returns the same errors as
   a value for consumers who want to validate before registering; tests can
   use it without triggering panics.

## Arguments for error returns

1. Testability: asserting a panic requires recover() ceremony; an error is a
   plain assert.
2. Composability: plugin systems that register user-supplied rules could
   reject bad rules gracefully instead of crashing the host.
3. Library convention: many modern Go libraries avoid panics entirely.

## Decision

**Keep panics.** Rationale: this registry is a startup-time composition
point for first-party rules, not a runtime ingestion point for untrusted
input. The failure modes panic guards against are exactly "the program is
wrong", and the ecosystem's Must* precedent matches. The plugin-composability
concern is hypothetical today (no active consumers register dynamic rule
sets); if a real consumer needs graceful rejection, add a separate
`TryRegister(rule) error` alongside the panic API — that would be additive,
not breaking, and would NOT justify a v0.4.0 on its own.

**Consequence for R1:** no API movement, so v0.4.0 remains gated on future
real API changes (per AGENTS.md user decisions, 2026-09-09).

## Revisit trigger

A consumer that registers rules from external input (plugin loading, config
files) where a panic would crash a long-running host process.
