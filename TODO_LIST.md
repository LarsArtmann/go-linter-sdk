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

None open. (#29 CoverageReporter fleet adoption completed 2026-11-19: all
five walking-rule linters report coverage evidence — go-structure-linter,
samber-linter, go-humanize-linter, cmdguard/lint, project-meta/linter —
plus BuildFlow's V2 bridge (`dtool.DetectorV2FromFinding`, preferred by
`ToolFromSpec`); the first non-zero sweep is recorded as "AFTER sweep 2" in
`docs/review/libs/FLEET-SWEEP-BASELINE.md` and logged in the [Unreleased]
CHANGELOG entry. (#28 the WithEdits declaration teaching shipped
2026-11-19: oops example carries `WithEdits` + `FixStrategyDirect`, README
documents the three host-side lanes; row removed per the DONE legend.))

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

- **#31 QUESTION (BLOCKED on owner) — Commit/push the consumer bumps (v0.5.0
  AND v0.6.0 waves, both local-only).** All 8 consumer module locations are at
  **v0.6.1** and gate-green LOCALLY (2026-11-20 third wave, after the v0.6.1
  release; the earlier waves' history is in the ADOPTION snapshots). Nothing
  pushed by any session. Owner call: push all six repos now, or leave each
  repo to its own session's end-of-session push (the standing policy).
  Context: docs/ADOPTION.md snapshots; question g-1 of
  docs/status/2026-11-19_14-43. samber-linter#5 and BuildFlow#112 are FIXED
  locally (2026-11-20); their fixes ride each repo's own push.

### From the coverage-fleet-adoption session (2026-11-19 21:41 report §f)

None open. (#36 v0.6.1 released 2026-11-20: checklist 7/7, push, CI green
on `6cc2311`, tag on the green SHA, proxy `go list -m @v0.6.1` verified;
#40's ADOPTION/DOMAIN_LANGUAGE/ROADMAP edits landed, the AGENTS session
section followed the release. Both upstream flakes fixed: samber-linter#5
root-caused to ambient-GOWORK leakage into the loader child — the test now
pins the workspace explicitly — plus the master CI red streak cured
(paralleltest/t.Chdir, golines, wrapcheck ignore-sigs); BuildFlow#112's two
flakes root-caused from a `-v` failure capture — a breaker-vs-scheduler race
in the AlternatingRepoSmoke test (assertions relaxed to the documented
batching semantics) and a real global-mutation leak (`KeepsAllowed` mutates
FormatterRecommendationAllowlist while parallel siblings read it; mutator
made serial). Providers suite green twice consecutively post-fix.)


### From the v0.6.1-release session (2026-11-20 report §f)

- **#41 TODO — Record the pkg.go.dev "Imported by" trend** once propagation
  settles (baseline 3, set 2026-10-15): fetch the v0.6.1 page's count and
  note it next to the baseline in AGENTS' post-release-habit section.
  Deferred from the release because the bot-served page lacked the count.
- **#42 TODO — Capture the BuildFlow#112 flake evidence into that repo**
  before /tmp cleanup: `/tmp/providers-v.log` holds the `-v` failure capture
  (run 2, AlternatingRepoSmoke empty-vs-v1.2.3) that identified the breaker
  race; the repo's tracker note (or its TODO_LIST row for #112) should carry
  the mechanism summary. Its own session closes the issue on push.
- **#43 TODO (samber-linter, its session) — push master with the #5 fix +
  lint cures; verify CI green with f38's checklist.**

### Docs corpus debt (ratchet-pinned 2026-11-19)

None open. (#35's audit drain completed 2026-11-20: five inline verdicts
applied at the replay's live anchors (the pair the 21:41 report predicted
had moved — two `--format json` rows, the CoverageReporter README row, the
cli-triage README example row, and the pkg.go.dev homepage item, verified
already set via `gh repo view`), plus the two items the #35 TODO row itself
used to anchor (severity-tiered exit codes → done; the go-finding
DetectorV2 cross-link ask → routed to go-finding's TODO_LIST). docs-gates
5×PASS. The lesson stands in AGENTS: completing a routed item needs its
inline verdict in the SAME session. (#30 the audit-all drain completed
2026-11-19 in one pass: all 24 historical unharvested items carry inline
verdicts — done / NOT-DO superseded / routed; `.docs-gates.json` ratchet
lowered 33 → 0, so any future unharvested item gates red immediately.
#32 and #33, both routed from that drain, completed 2026-11-19: #32 landed
as the flake's eval-time `golangci-version-sync` assertion (caught live
action-pin-vs-flake drift on its first run) plus a shape rule in
check-doc-versions.sh; #33 landed as a shellcheck flake check + CI job with
all four scripts clean — logged in the [Unreleased] CHANGELOG entry.)
