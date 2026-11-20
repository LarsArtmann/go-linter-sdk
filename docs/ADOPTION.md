# Adoption

Who uses `go-linter-sdk`, how, and whether the SDK kept its core promise:
**rules emit `finding.Finding` directly — no custom Issue/Violation converter
layer.** Verified by source inspection on 2026-10-11 (see "Refresh procedure"
below to re-verify).

Consumers pin `go-linter-sdk` per their own `go.mod`; this doc intentionally
duplicates no version numbers (the doc-drift guard, `scripts/check-doc-versions.sh`,
keeps living docs version-pointered).

## Consumers

| Repository                             | Integration pattern                                                                                                                                                              | Converter layer                                                                     | Push state (2026-11-19)                                              |
| -------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| `go-humanize-linter`                   | 11 rules as `linter.RuleFunc` in `linter.Registry`; CLI + golangci plugin                                                                                                        | None — born native                                                                  | Local only (bump commits by its daemon; unpushed)                    |
| `cmdguard` (`cmdguard/lint`)           | Rules as `linter.RuleFunc`; BuildFlow provider via `go-finding/toolsdk` `Spec`                                                                                                   | None — born native                                                                  | Local only (bump commits by its daemon; unpushed)                    |
| `project-meta` (`project-meta/linter`) | `linter.RuleFunc` rules; `linter.Registry`; CLI exit via `linter.ExitCodeByConfidence`                                                                                           | None — born native                                                                  | Local only (bump commits by its daemon; unpushed)                    |
| `BuildFlow` (`tools/providers`)        | Runs sibling registries in-process (`Registry.Run`); deleted its `PrioritizedViolation` hierarchy (~4,737 LOC) and 10 converter files in favor of `finding.Finding`              | Deleted (see `BuildFlow/CHANGELOG.md`)                                              | Local only (bump commits by its daemon; unpushed; filed #112)        |
| `go-structure-linter` (pilot)          | SDK registry/detectors under a rule-level adapter (`internal/rules/sdk_adapter.go`); local `types.Issue` is a **type alias** of `finding.Finding`, so no data translation exists | Converter struct deleted; adapter kept **by design** (ADR-003)                      | Local only (bump commits by its daemon; unpushed)                    |
| `samber-linter`                        | Rules are `go/analysis` analyzers (golangci plugin contract); SDK supplies exit-code policy (`linter.ExitCodeByConfidence`)                                                      | Never had one; `analysis.Diagnostic → finding.Finding` mapping is documented design | Local only (README truth fix + bump + #5 fix + lint cures unpushed; #5 FIXED 2026-11-20) |

Push-state column added 2026-11-19 (harvest of status-report F14.8): all
consumer-repo work from the v0.5.0/v0.6.0 waves stays unpushed until owner
question g-1 is answered; this column is re-verified at every refresh
(see "Refresh procedure" step 5).

## Verdicts on the core claim

- **Born-native consumers (3):** built directly on `linter.RuleFunc` +
  `finding.Finding`; nothing to delete, nothing found.
- **Migrating consumers (2):** `BuildFlow` physically deleted its violation
  hierarchy and converter package; `go-structure-linter` deleted its custom
  `Issue` struct (now an alias) and per-finding translation.
- **Different-shape consumer (1):** `samber-linter` adopts the SDK's exit-code
  contract while keeping `go/analysis` as its rule interface — documented in
  its planning docs as deliberate.

No inspected repository retains a custom intermediate Issue/Violation struct
that is translated field-by-field into findings. The claim holds.

## What "adopted" means here

A repository counts as a consumer when its `go.mod` requires
`github.com/larsartmann/go-linter-sdk` and its code imports the SDK (registry,
rules, detectors, or exit-code helpers). Version numbers live in each
repository's `go.mod`.

## Metrics

Tracked snapshots; update alongside the refresh procedure.

**2026-10-11:**

- Direct consumers (go.mod `require` + code import): **6** (table above).
- Ecosystem `go-finding` version distribution: of 95 `go.mod` files referencing
  `go-finding` across `~/projects` — 60 on the current release, 31 one minor
  behind, 4 scattered on older minors. The SDK itself tracks the current
  release (see `go.mod`).
- pkg.go.dev "Imported by" for this module: **3** public importers shown
  (`cmdguard/lint`, `go-humanize-linter`, `samber-linter/internal/driver`) —
  the rest of the adoption is private or unindexed; the go.mod scan is the
  more complete signal.

**2026-10-15 (bump wave):**

- All six direct consumers at **v0.4.0**. Four were already bumped by their
  own sessions (go-structure-linter, samber-linter, go-humanize-linter, and
  all three BuildFlow modules); this session bumped `cmdguard/lint` and
  `project-meta/linter` (`go get` + tidy with go.work renamed during the
  write; build + `go test` + `-race` + `go mod verify` green; re-verified
  via go.mod grep after).
- Additive check done at source: the v0.3.1→v0.4.0 exported-symbol delta is
  tests/benchmarks plus three new `Category` constants — no removals, no
  renames.
- samber-linter master carries one red test (`TestGoWorkMultiModule` in
  internal/driver: HW-1 fixture expectation not firing); **bisect-proven
  pre-existing** — fails identically at v0.3.1; restored to v0.4.0 after.
- BuildFlow local gates were skipped deliberately: its working tree carried
  an uncommitted local `replace` plus an in-flight submodule bump from a
  concurrent session, so local runs would have measured the replace target,
  not the tag. Its CI is the verification layer.
- pkg.go.dev "Imported by": **3** (unchanged, re-verified 2026-10-15).

**2026-11-19 (v0.5.0 bump wave + fleet sweep):**

- All eight consumer module locations bumped to **v0.5.0** via
  `go get` + `go build ./...`, then `go test ./...`: go-structure-linter,
  samber-linter, go-humanize-linter, cmdguard/lint, project-meta/linter,
  BuildFlow/execution, BuildFlow/tools, BuildFlow root (indirect follow).
- **Correction to the 2026-10-15 snapshot:** project-meta/linter and all
  three BuildFlow modules had silently drifted back (v0.4.0/v0.4.1 mix
  found pre-wave); the wave brought every location to the current release.
  Lesson: bump waves decay — re-verify via go.mod grep each time (the
  refresh procedure below already says so; this is the recorded instance).
- **Test gates, honestly recorded:** green on go-structure-linter,
  go-humanize-linter, cmdguard/lint, project-meta/linter,
  BuildFlow/execution. Three failures, all **bisect-proven pre-existing**
  (fail identically with the bump stashed): samber-linter
  `TestAnalyzeEnv_GOWORKOffEscapesWorkspaceShadow` (pkg/sdk, workspace-shadow
  environment sensitivity) + `TestReadmeLatestReleaseMatchesGitTag` (that
  repo's own release-flow drift: README says previous release, tag already
  newer); BuildFlow/tools `TestSyncGoWorkFile_RealGoWorkUse_HonorsIgnoreDirs`
  (gomod sync fixture). BuildFlow root additionally fails its own audit
  gates (em-dash + oversized `discovery/doctor/checks/binary_freshness.go`)
  — files from a concurrent session, untouched here. None are SDK-caused;
  they belong to their repos' sessions.
- **Channel sweep (plan M14):** zero direct references to
  `DetectorV2`/`AddDetectResult`/`CoverageReporter`, `Relativize`, or
  `WithEdits` in any consumer — expected: v0.5.0 shipped the channels the
  same day. Baseline evidence + AFTER targets:
  `docs/review/libs/FLEET-SWEEP-BASELINE.md`.
- pkg.go.dev "Imported by": **3** (re-verified 2026-11-19).

**2026-11-19 later (v0.6.0 bump wave, LOCAL ONLY — pushes owner-gated):**

- **Second recorded decay instance, same day:** BuildFlow's three modules
  (execution, tools, root-indirect) were found back at **v0.4.1** when this
  wave re-grepped — the "all eight at v0.5.0" line above was wrong for
  BuildFlow within hours of being written (a concurrent BuildFlow session or
  its daemon reverted/reset the module files). The lesson compounds: never
  trust a snapshot; grep go.mod at wave time.
- All eight module locations now at **v0.6.0**: go-structure-linter,
  samber-linter, go-humanize-linter, cmdguard/lint, project-meta/linter
  (root + linter), BuildFlow/execution, BuildFlow/tools, BuildFlow root
  (indirect). `go get @v0.6.0` per module.
- **Gates:** green on go-structure-linter, go-humanize-linter, cmdguard/lint,
  project-meta/linter, BuildFlow/execution + build on all three BuildFlow
  modules. Known-failure inventory unchanged and still not SDK-caused:
  samber-linter `TestReadmeLatestReleaseMatchesGitTag` (their release-flow
  drift); BuildFlow/tools `providers` suite is **flaky** (failed once,
  passed on identical bytes immediately after — test-isolation issue in that
  repo, routed to its tracker, not a v0.6.0 regression).
- **Push state: all local.** No commits authored by this session in consumer
  repos (their daemons auto-committed the go.mod/go.sum changes; nothing
  pushed by this session). Owner question g-1 governs pushes.
- pkg.go.dev: v0.6.0 indexed (proxy `go list -m @v0.6.0` verified; page
  render checked post-release).
- **2026-11-19 addendum — decay #4 and the clean-run record:** a concurrent
  session clobbered the uncommitted V2-bridge edits (~20:49); everything was
  re-applied and the daemon committed it (`e18a22479`). After the re-apply,
  the BuildFlow/tools FULL providers suite passed on a clean run
  (golangci/gomod/npm/providers all `ok`; providers 52s) — the three flaky
  failures from the first run did not reproduce (test-isolation class,
  tracked as BuildFlow#112).
- **2026-11-19 teaching-debt sweep — clean:** four production
  `finding.NewBuilder` call sites exist across consumers and all are
  deliberate rather than debt: cmdguard/lint `newFinding(meta, …)`
  (init-cycle avoidance, documented in-code), go-humanize-linter
  `makeSuppressionVerificationFinding` (synthetic H0SUP rule), samber-linter
  driver.go (driver-layer stamping with per-rule severity/confidence
  lookup), project-meta/linter `metadataMissingFinding` (single-constant
  rule). BuildFlow `modules/*` construct findings directly — pre-SDK
  checkers at a different integration layer, out of scope for the
  `rule.NewFinding` teaching. No fixes required; re-check on consumer
  refreshes.
- **2026-11-20 third wave — v0.6.1:** all 8 module locations bumped to
  v0.6.1 the day of the release, gated per repo (build + full/targeted
  tests, GOWORK=off): go-structure-linter (build), samber-linter (6/6
  packages), go-humanize-linter (4/4), cmdguard/lint, project-meta/linter,
  BuildFlow/execution + tools (providers suite green, 83s) + root.
  samber-linter#5 and BuildFlow#112 were root-caused and FIXED locally in
  the same pass (see the TODO_LIST 2026-11-20 entry for the mechanisms).
  Push state: all local, per standing policy.

## Refresh procedure

1. For each row above: `grep -rn "go-linter-sdk" <repo> --include="go.mod"`.
2. Re-run the converter check: search the consumer for `type (Issue|Violation)`
   structs and `toFinding`-style translation functions; a finding is only a
   problem if a _data_ converter reappears (rule-level adapters documented as
   deliberate, like `go-structure-linter` ADR-003, are acceptable).
3. Update the table and this header's verification date.
4. Record new consumers via the module proxy "Imported by" count on
   pkg.go.dev plus the sibling-repo scan used in the 2026-10-11 session.
5. Re-verify the Push-state column per repo (`git status` + whether the
   bump commits are pushed); update the column's date stamp.

## Related

Point-in-time utilization reviews of every first-party library this SDK
touches (including this SDK's own self-audit) live in
[`docs/review/libs/`](review/libs/index.html) — start at the dashboard,
`index.html`; each report records the exact version reviewed and its
re-review trigger.
