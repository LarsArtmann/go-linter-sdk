# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md`.
> Items are ranked by impact. Status is verified, not assumed.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                                   |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                   |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed.     |
| 🟢 `DONE`        | Completed. Remove from this list and log in `CHANGELOG.md`. |

## Open work

All items from the previous backlog have been completed and logged in
`CHANGELOG.md`. The registry API now includes `Get`/`Has`/`Deregister`,
`RuleMeta.Validate`, configurable `Run` failure policy (`ContinueOnError`),
testable examples, and two example consumers (`examples/minimal-linter`,
`examples/no-go-mod`).

When a `ROADMAP.md` idea is refined into a bounded task, it moves here.

---

<!-- Guidance for the builder:
  - Source of truth is the CODE. Verify each item before adding; many
    documented TODOs are already done (dropped here: LICENSE, .golangci.yml,
    CONTRIBUTING.md, reports/.gitkeep, the `vet` app, .editorconfig, the flake
    `self` arg, and several tests).
  - DONE items are REMOVED, not kept. Log them in CHANGELOG.md.
  - If a task turns vague, move it to ROADMAP.md.
  - Deduplicate by semantic intent, not by text match.
-->
