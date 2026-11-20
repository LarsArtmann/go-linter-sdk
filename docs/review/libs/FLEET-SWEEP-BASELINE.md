# Fleet sweep baseline — 3 missed channels × 6 consumers (2026-11-19)

> Plan task M14 (F14.1–F14.7), executed immediately after the v0.5.0 release
> (which ships the channels). Purpose: record the BEFORE state so the AFTER
> (consumers adopting CoverageReporter / Relativize / WithEdits) is
> measurable on the next sweep. Sweep method: `rg -c '<channel>' --glob
> '*.go' .` per consumer module, run at each module root.

## Baseline: all channels zero

| Consumer (module root)                      | SDK pin before wave | DetectorV2 / AddDetectResult / CoverageReporter | Relativize | WithEdits / TextEdit |
| ------------------------------------------- | ------------------- | ----------------------------------------------- | ---------- | -------------------- |
| go-structure-linter                         | v0.4.1              | 0 refs                                          | 0 refs     | 0 refs               |
| samber-linter                               | v0.4.1              | 0 refs                                          | 0 refs     | 0 refs               |
| go-humanize-linter                          | v0.4.1              | 0 refs                                          | 0 refs     | 0 refs               |
| cmdguard/lint                               | v0.4.1              | 0 refs                                          | 0 refs     | 0 refs               |
| project-meta/linter                         | v0.4.0              | 0 refs                                          | 0 refs     | 0 refs               |
| BuildFlow/execution (+tools, root indirect) | v0.4.0 / v0.4.1     | 0 refs                                          | 0 refs     | 0 refs               |

## Bump wave executed 2026-11-19 (F13.5)

All 8 module locations were bumped to `go-linter-sdk v0.5.0` with `go get` +
`go build ./...` green; `go test ./...` gates recorded in the session log.
Note: project-meta/linter and BuildFlow's three modules were still on
v0.4.0/v0.4.1 BEFORE this wave — the 2026-10-15 "all six at v0.4.0" snapshot
had silently gone stale (found and fixed here; ADOPTION refresh F13.6
records the corrected table).

**Correction (2026-11-19, second wave):** the "all 8 at v0.5.0" statement
above decayed within the day — BuildFlow was found back on v0.4.1 when the
v0.6.0 wave ran (its go.mod/go.sum had been reverted by local operations in
that repo). The wave-time discovery and the honest correction live in
docs/ADOPTION.md's "2026-11-19 later (v0.6.0 bump wave, LOCAL ONLY)"
snapshot; this historical section is left as-written per the annotate
convention (the correction travels with the newer record, not by rewriting
history). Reader lesson: an "all N at version X" claim is only true until
the next local mutation — re-grep go.mod before trusting it.

## What changes with v0.5.0 without consumer code

- **Relativize:** consumer reports via `Registry.Run` and both adapters now
  emit root-relative finding IDs automatically — the host-path leak
  (`/home/...` in IDs) is gone fleet-wide with zero consumer edits.
- **filesScanned:** still 0 until a consumer rule implements
  `CoverageReporter` — the channel is shipped; the evidence needs one small
  interface per walking rule.
- **Edits:** unchanged (rules opt in via `WithEdits`, hosts apply; ADR 0002).

## AFTER-target for the next sweep

1. Each walking-rule linter (go-structure-linter, samber-linter,
   go-humanize-linter, cmdguard/lint, project-meta/linter) implements
   `CoverageReporter` on its tree-walking rules → `filesScanned ≠ 0` in
   fleet dashboards.
2. BuildFlow/execution consumes `DetectorV2` from SDK adapters in its DAG
   conversion (it already runs `RunDetect`-shaped hosts).
3. Any rule with a mechanical fix attaches `WithEdits` (ADR 0002 GO items).

## AFTER sweep 1 — post-v0.6.0 (2026-11-19, scripts/fleet-sweep.sh)

| Consumer (module root) | SDK pin | DetectorV2 / AddDetectResult / CoverageReporter | Relativize | WithEdits / TextEdit |
| ---------------------- | ------- | ----------------------------------------------- | ---------- | -------------------- |
| go-structure-linter    | v0.6.0  | 0 refs                                          | 0 refs     | 0 refs               |
| samber-linter          | v0.6.0  | 0 refs                                          | 0 refs     | 0 refs               |
| go-humanize-linter     | v0.6.0  | 0 refs                                          | 0 refs     | 0 refs               |
| cmdguard/lint          | v0.6.0  | 0 refs                                          | 0 refs     | 0 refs               |
| project-meta/linter    | v0.6.0  | 0 refs                                          | 0 refs     | 0 refs               |
| BuildFlow/execution    | v0.6.0  | 0 refs                                          | 0 refs     | 0 refs               |
| BuildFlow/tools        | v0.6.0  | 0 refs                                          | 0 refs     | 0 refs               |

Reading: every location is on the current release; channel adoption is still
zero — that is TODO_LIST #29's consumer-side work (CoverageReporter on
walking rules, DetectorV2 in BuildFlow's DAG, WithEdits where rules know a
mechanical fix). Relativize needs no consumer code, so host-portable IDs are
already fleet-wide through `Registry.Run`/adapters. The sweep itself is now
mechanical: `bash scripts/fleet-sweep.sh`.

## AFTER sweep 2 — TODO #29 executed (2026-11-19, same day)

| Consumer (module root) | SDK pin | DetectorV2 / AddDetectResult / CoverageReporter | Relativize | WithEdits / TextEdit |
| ---------------------- | ------- | ----------------------------------------------- | ---------- | -------------------- |
| go-structure-linter    | v0.6.0  | **7 refs**                                      | 0 refs     | 0 refs               |
| samber-linter          | v0.6.0  | **6 refs**                                      | 0 refs     | 0 refs               |
| go-humanize-linter     | v0.6.0  | **6 refs**                                      | 0 refs     | 0 refs               |
| cmdguard/lint          | v0.6.0  | **7 refs**                                      | 0 refs     | 0 refs               |
| project-meta/linter    | v0.6.0  | **8 refs**                                      | 0 refs     | 0 refs               |
| BuildFlow/execution    | v0.6.0  | 0 refs                                          | 0 refs     | 0 refs               |
| BuildFlow/tools        | v0.6.0  | **3 refs**                                      | 0 refs     | 0 refs               |

Reading — the coverage channel is LIVE fleet-wide, first non-zero sweep:

- **All five walking-rule linters report CoverageReporter evidence**
  (TODO #29's primary target): go-structure-linter counts via a
  `countingProject` decorator at the SDK-adapter's project factory;
  samber-linter via the driver-level `scanCoverage` (parsed-file count
  across loaded packages); go-humanize-linter via a per-Check context sink
  through the shared walker (`checkFuncDecls`), registry wrapped with
  `coverageRule`; cmdguard/lint via a `treeWalkingRule` wrapper over the
  single `analyze()` walker (all 8 rules); project-meta/linter via honest
  per-scan counters on `metadataMissingCheck` (0 on every early return).
- **BuildFlow's V2 lane**: `dtool.DetectorV2FromFinding`
  (domain/tool) folds DetectResult scan evidence into FindingReport.Summary;
  `ToolFromSpec` (tools/providers) prefers the V2 lane when the SDK
  detector implements `gofinding.DetectorV2`. The 3 tools refs are that
  bridge; execution itself stays 0 by design — the DAG consumes the domain
  Detector abstraction, and the coverage flow enters at the
  tools/providers boundary (domain/tool is its own module, not swept).
- **Relativize still 0 refs**: correct — it needs no consumer code; the
  channel is active through the SDK's own Run/adapters.
- **WithEdits still 0 refs**: remaining adoption surface (rule-level fix
  declarations, ADR-0002 GO items) — future work, tracked in ROADMAP.
  **Eligibility sweep verdict (2026-11-19, follow-up session): zero eligible
  landing sites exist today.** go-humanize-linter's rules are function-level
  heuristics with type-dependent suggestions (Suggest is honest; edits would
  fabricate span/type precision the rules do not have); cmdguard/lint and
  samber-linter fixes are API-migration patterns across call sites (not
  single-span rewrites); go-structure-linter's direct fixes are whole-file
  create/delete operations, correctly carried as BeforeCode/AfterCode
  (TextEdit has no file-create/delete op); project-meta's fix creates a new
  file (FixEngine territory, not TextEdit). The one true inline-fix rule —
  cli-triage's redundant-nil-guard oops example — already carries WithEdits
  (ADR 0002 teaching, shipped with the v0.5.0-era #28 work). The channel
  stays 0-ref BY DESIGN until a consumer rule gains genuine inline-fixable
  detection; re-run this sweep when one lands.
- **Accumulation note**: go-finding's `AddDetectResult` SUMS per-rule
  counts, so an N-rule registry over F files reports N·F — the documented
  contract, pinned by cmdguard's and go-humanize's new tests.
- **Decay instance #4 recorded**: a concurrent BuildFlow session wiped the
  uncommitted V2-bridge edits mid-session (its own 20:39 race-triage status
  report proves the collision); re-applied and daemon-committed
  (e18a22479) within minutes. Shared-checkout lesson: gate AND commit fast.

## How to verify

```bash
for mod in ~/projects/go-structure-linter ~/projects/samber-linter \
           ~/projects/go-humanize-linter ~/projects/cmdguard/lint \
           ~/projects/project-meta/linter ~/projects/buildflow/execution; do
  echo "== $mod"; (cd $mod && rg -c 'DetectorV2|AddDetectResult|CoverageReporter|Relativize|WithEdits' --glob '*.go' . || echo 0 refs)
done
grep -n 'go-linter-sdk' ~/projects/{go-structure-linter,samber-linter,go-humanize-linter}/go.mod \
  ~/projects/cmdguard/lint/go.mod ~/projects/project-meta/linter/go.mod \
  ~/projects/buildflow/{execution,tools}/go.mod
```

— Snapshot 2026-11-19, orchestrating session. Point-in-time; living state in
docs/ADOPTION.md.

## AFTER sweep 3 — v0.6.1 wave (2026-11-20, follow-up session)

| Consumer (module root) | SDK pin | DetectorV2 / AddDetectResult / CoverageReporter | Relativize | WithEdits / TextEdit |
| ---------------------- | ------- | ----------------------------------------------- | ---------- | -------------------- |
| go-structure-linter    | v0.6.1  | **7 refs**                                      | 0 refs     | 0 refs               |
| samber-linter          | v0.6.1  | **6 refs**                                      | 0 refs     | 0 refs               |
| go-humanize-linter     | v0.6.1  | **6 refs**                                      | 0 refs     | 0 refs               |
| cmdguard/lint          | v0.6.1  | **7 refs**                                      | 0 refs     | 0 refs               |
| project-meta/linter    | v0.6.1  | **8 refs**                                      | 0 refs     | 0 refs               |
| BuildFlow/execution    | v0.6.1  | 0 refs                                          | 0 refs     | 0 refs               |
| BuildFlow/tools        | v0.6.1  | **3 refs**                                      | 0 refs     | 0 refs               |

Reading — pins uniform at v0.6.1 (third wave, gated green per repo:
build + full/targeted tests, GOWORK=off everywhere); channel counts stable
vs sweep 2, as expected — no behavior changed in the PATCH. WithEdits stays
0 by the eligibility verdict above (zero single-span mechanical fixes exist
fleet-wide; the oops example carries the teaching pattern). All bumps are
LOCAL; pushes remain each repo's own session's end-of-session duty.
