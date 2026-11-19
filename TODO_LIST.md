# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, see `ROADMAP.md`.
> Items are ranked by impact. Status is verified, not assumed.

## Status legend

| Status      | Meaning                                                     |
| ----------- | ----------------------------------------------------------- |
| TODO        | Not started. Needs doing.                                   |
| IN_PROGRESS | Actively being worked on.                                   |
| BLOCKED     | Cannot proceed, external dependency or decision needed.     |
| `DONE`      | Completed. Remove from this list and log in `CHANGELOG.md`. |

## Open work

<!-- Numbering gap at #1, #2, #4: those items were completed and removed
     (1: v0.3.1 release, 2: secret deletion, 4: dprint decision). Numbers are
     never reused — historical reports reference them. -->

### Consumer adoption (the reason this SDK exists)

None open. (#10 converter-deletion verification and #11 the durable adoption
doc both completed 2026-10-11: all six consumers verified at source level,
evidence and refresh procedure in `docs/ADOPTION.md`; logged in the v0.4.0
CHANGELOG entry.)

### Documentation & drift

None open. (#12 the CI doc-drift guard and #13 the version-pointer refactor
both completed 2026-10-11: `scripts/check-doc-versions.sh` runs as a CI job
and a flake check; living docs point at `go.mod`. Logged in the v0.4.0
CHANGELOG entry.)

### Release

None open. (#19 pkg.go.dev verification completed 2026-10-15: the site serves
v0.4.0 as Latest with the frozen v0.4.0 README — verified mechanically from
the page's raw HTML (breadcrumb `go-linter-sdk@v0.4.0`, `Version: v0.4.0`,
`Published: Oct 11, 2026`); no "Request a version" escalation was needed.)

### Dependency hygiene

None open. (#15 benchmarks re-run 2026-10-11 — every measured op improved
20-30% vs the 2026-09-09 baseline, no regressions; baseline refreshed.
#16 dependabot sweep 2026-10-11: zero open PRs, nothing superseded.)

### Fleet library review follow-ups (2026-11-19, `docs/review/libs/`)

- **#23 BLOCKED — Revisit go-policy-dsl README companion-libs mention when it
  tags a stable v1** (its ROADMAP explicitly names this SDK as prospective
  secondary consumer; pre-v1 breaking changes make an earlier mention a
  liability). Evidence: `docs/review/libs/go-policy-dsl.md`.

Executed from this batch (2026-11-19): #20, #21, #22, #24, #25, #26, #27
shipped in or alongside v0.5.0 — see the [0.5.0] CHANGELOG entry and the
2026-11-19 ADOPTION snapshot; rows removed per the DONE legend above.

### From the v0.5.0 execution (2026-11-19, Pareto plan M-series)

- **#28 TODO — Teach fix DECLARATION in an example (ADR 0002 GO items).**
  Extend `examples/oops-redundant-guard` to attach `WithEdits` +
  `FixStrategyDirect` to its findings (~15 LOC, no application code) and add
  a README subsection pointing hosts at `finding.ApplySimpleFixes` and the
  toolsdk `Repairer` recipe. Design: `docs/decisions/2026-11-19_rule-fix-representation.md`.
- **#29 TODO — CoverageReporter adoption at consumers (fleet AFTER targets).**
  Each walking-rule linter implements `CoverageReporter` on its tree-walking
  rules so fleet dashboards get `filesScanned ≠ 0`; BuildFlow/execution
  consumes `DetectorV2` from the SDK adapters. Before-state and per-repo
  targets: `docs/review/libs/FLEET-SWEEP-BASELINE.md`.

### Quality hardening

None open. (#17 flake `checks` for `go test` / `go vet` / `golangci-lint` plus
the doc guard completed 2026-10-11 — `nix flake check` runs five hermetic
derivations. #18 the bump checklist is encoded in
`scripts/bump-checklist.sh` and referenced from CONTRIBUTING.md.)

---

<!-- Guidance for the builder:
  - Source of truth is the CODE. Verify each item before adding; many
    documented TODOs are already done.
  - DONE items are REMOVED, not kept. Log them in CHANGELOG.md.
  - If a task turns vague, move it to ROADMAP.md.
  - Deduplicate by semantic intent, not by text match.
  - Vague / long-term items belong in ROADMAP.md, not here.
-->

### Docs corpus debt (ratchet-pinned 2026-11-19)

- **#30 TODO — Drain the audit-all corpus debt (33 historical unharvested
  items across 27 archived reports, July–Oct sessions).** Each item needs an
  inline verdict (`done at <hash>` / NOT-DO / routed) per docs-health
  ANNOTATE mode. `.docs-gates.json` pins `audit_max: 33` — the ratchet only
  goes DOWN; every drain pass lowers it toward strict (-1). Evidence:
  `docs-gates run` audit-all replay.
