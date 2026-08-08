# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md`.
> Items are ranked by impact. Status is verified, not assumed.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| TODO         | Not started. Needs doing.                                   |
| IN_PROGRESS  | Actively being worked on.                                   |
| BLOCKED      | Cannot proceed, external dependency or decision needed.     |
| DONE         | Completed. Remove from this list and log in `CHANGELOG.md`. |

## Open work

Harvested from `docs/status/2026-08-05_10-32` (TODO execution report, section
f) and `docs/status/2026-08-05_10-27` (doc-drift audit, section f). Each item
was verified against the codebase — items already shipped were dropped and
logged in `CHANGELOG.md`.

### Testing

| # | Task                                                    | Impact | Effort | Evidence                                                                  |
| --- | ------------------------------------------------------ | ------ | ------ | ------------------------------------------------------------------------- |
| 1   | Integration tests for both example binaries            | High   | 30m    | `examples/minimal-linter/main.go`, `examples/no-go-mod/main.go` — compile and run manually but have no automated test asserting exit codes. Prevents rot of the SDK's key validation evidence. (report 10-32 §f.11-12) |
| 2   | Fuzz `NewRuleError` with nil cause                     | Medium | 15m    | `errors.go` — `Error()` formats `e.Cause` with `%v`; a nil cause has no test. Verify it does not panic. (report 10-32 §f.13, 10-27 §f.38) |
| 3   | Test `errors.Is` propagation for context errors        | Medium | 15m    | `errors.go` — `RuleError.Unwrap()` returns `Cause`, but no test verifies `errors.Is(ruleErr, context.Canceled)` or `context.DeadlineExceeded` chains correctly. (report 10-32 §f.14-15) |
| 4   | Test `Deregister` during concurrent `Run`               | Low    | 20m    | `registry.go` — `Run` snapshots via `All()` before iterating; a rule deregistered after the snapshot but before execution still runs. Document and test this snapshot semantics. (report 10-32 §e.2, §f.16) |

### Code quality

| # | Task                                                    | Impact | Effort | Evidence                                                                  |
| --- | ------------------------------------------------------ | ------ | ------ | ------------------------------------------------------------------------- |
| 5   | Refactor `validateRuleIdentity` + `RuleMeta.Validate`  | Medium | 20m    | `rule.go` — both functions check the same four identity fields through different paths (interface methods vs struct fields). Eliminate the duplication while keeping the interface-level check for custom `Rule` implementations. (report 10-32 §e.4, §f.42) |

---

<!-- Guidance for the builder:
  - Source of truth is the CODE. Verify each item before adding; many
    documented TODOs are already done.
  - DONE items are REMOVED, not kept. Log them in CHANGELOG.md.
  - If a task turns vague, move it to ROADMAP.md.
  - Deduplicate by semantic intent, not by text match.
  - Vague / long-term items belong in ROADMAP.md, not here.
-->
