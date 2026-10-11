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
| DONE        | Completed. Remove from this list and log in `CHANGELOG.md`. |

## Open work

<!-- Numbering gap at #1, #2, #4: those items were completed and removed
     (1: v0.3.1 release, 2: secret deletion, 4: dprint decision). Numbers are
     never reused — historical reports reference them. -->

### Publication (repo went public 2026-09-08)

None open. (2026-09-09: gitleaks full-history scan clean + sensitive-docs
sweep accepted with dated verdict; GitHub metadata set — description, topics
`go`/`linter`/`static-analysis`, homepage → pkg.go.dev.)

### Consumer adoption (the reason this SDK exists)

- **#10 — Verify converter-package deletion.** Six internal repos import the SDK
  (verified 2026-10-11: `go-structure-linter`, `samber-linter`,
  `go-humanize-linter`, `cmdguard/lint`, `project-meta/linter`,
  `BuildFlow/execution`), but importing is not migrating. Inspect each and
  record whether its bridge/converter package is gone. Impact: **Critical**.
- **#11 — Record the consumer list as a living adoption doc.** The
  dependency-graph scan is the only place this data lives today; capture a
  durable, refreshable version. Impact: Medium.

### Documentation & drift

- **#12 — Add a CI guard against doc/version drift.** Fail when a living doc's
  `go-finding`/Go version disagrees with `go.mod`. That failure class produced
  the 2026-10-11 drift (AGENTS/README/TODO/ROADMAP/DOMAIN_LANGUAGE all stale).
  Impact: High.
- **#13 — Convert remaining prose version references to "see go.mod".**
  AGENTS.md and README now point at `go.mod`; sweep the rest. Impact: Medium.

### Release

- **#14 — Cut the next release.** Five dependency/Go bumps and a flake fix have
  landed since `v0.3.1` with no tag; consumers pin an artifact that no longer
  matches master. Tag only after green CI. Impact: High.

### Dependency hygiene

- **#15 — Re-run benchmarks** after the v1.10.0 → v1.14.0 bump window against
  `reports/benchmarks-baseline.txt` (forgotten in the v1.10.0 session).
  Impact: Medium.
- **#16 — Sweep dependabot PRs** (`gh pr list`); close any superseded by the
  manual bumps. Impact: Low.

### Quality hardening

- **#17 — Add flake `checks` for `go test`, `go vet`, and `golangci-lint`**
  (today only treefmt is a check derivation). Impact: Medium.
- **#18 — Enforce the bump checklist** (build/test/race/lint/vet/flake-check +
  `GOWORK=off` tidy + benchmark + changelog) so bumps stop landing undocumented.
  Impact: Medium.

Current dependency state (verified 2026-10-11): `go-finding` v1.14.0 (latest
root release), `go-error-family` v0.11.0, Go 1.27.

---

<!-- Guidance for the builder:
  - Source of truth is the CODE. Verify each item before adding; many
    documented TODOs are already done.
  - DONE items are REMOVED, not kept. Log them in CHANGELOG.md.
  - If a task turns vague, move it to ROADMAP.md.
  - Deduplicate by semantic intent, not by text match.
  - Vague / long-term items belong in ROADMAP.md, not here.
-->
