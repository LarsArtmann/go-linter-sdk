# go-compliance utilization review — go-linter-sdk

- **Reviewed:** 2026-11-19 by fleet audit Group C (sub-agent; orchestrated session)
- **Library:** github.com/larsartmann/go-compliance (+ sub-module `github.com/larsartmann/go-compliance/hub`)
- **Sub-modules:** two modules — core at repo root (`go.mod`: deps `go-branded-id v0.7.1`, `go-error-family v0.12.0`, ginkgo/gomega) and hub (`hub/go.mod`: ~30 direct deps — templ, datastar, sqlite, daemon, cmdguard, templ-components...; still carries `replace github.com/larsartmann/go-compliance => ../` at `hub/go.mod:135`, commented "Dev-only until core is tagged"). Consumer is single-module; go-compliance is deliberately split the way go-finding splits `pipeline/`/`toolsdk/`.
- **Version reviewed:** library repo **no tags** (`git tag -l | wc -l` → 0); HEAD moving during the audit itself: brief said `2a62a68`, session observed `1861917` then `ec61ad8` (2026-11-19, "chore: auto-commit N changed file(s) (heuristic)"); consumer pin: none (not in go.mod); point-in-time snapshot of an actively-churning pre-release tree
- **Verdict:** NOT-ADOPTED — and **not reverse-adopted either**: go-compliance consumes frozen v1 JSON evidence documents, not Go lint APIs, so it neither should be an SDK dependency nor needs to import the SDK; the SDK already feeds it transitively through BuildFlow
- **Score:** no score (not adopted)

## What the library offers

Verified from source (`go doc .`, exported-symbol sweep, README, FEATURES.md, `docs/guides/integrate.md`):

Core module (`/home/lars/projects/go-compliance/*.go`):

| Capability | Evidence | What it actually does |
| --- | --- | --- |
| Compliance state model | `state.go:13` (`State`), `state.go:112` (`DeriveState`) | Four states: `compliant` / `failing` / `stale` / `never-covered`; window semantics — one failed step anywhere in the window keeps the repo `failing` |
| Frozen v1 wire contract | `v1.go:14` (`SchemaVersionV1`), `v1.go:32`/`78`/`91`/`115` | `V1Report`/`MarshalV1`/`DecodeV1Report`/`BuildV1Report`; byte-faithful to BuildFlow's `compliance --json` producer, golden-tested both directions |
| Evidence fold | `fold.go:113` (`FoldEvidence`), `fold.go:196` (`CompleteWithDiscovery`) | Source-agnostic ensure/union fold over window health, latest runs, worst steps, **gate findings (`fold.go:56 RepoFindings`, `{step: count}`)** and **severity counts (`fold.go:63 RepoSeverities`)** |
| Source adapter interface | `source.go:12` (`Source`), `source.go:29` (`SourceFunc`) | Read-side port: ctx + window + step filter → `[]RepoEvidence` |
| BuildFlow adapters | `adapter.go:46` (`NewBuildFlowFileSource`), `adapter.go:190` (`NewBuildFlowCommandSource`), `adapter.go:131` (`V1ReportToEvidence`) | File-piped v1 documents and CLI-spawned `buildflow compliance --json` |
| Reporter SDK one-liner | `reporter.go:63` (`Record`), `reporter.go:73` (`RecordTo`), `reporter.go:23` (`RunResult`) | `compliance.Record(ctx, repo, RunResult{...})` writes to `/var/lib/overview/compliance.json` (env-overridable); `RunResult.GateFindings map[string]int` (`reporter.go:37`) and `FindingSeverities map[string]int` (`reporter.go:41`) — severities are dropped by file-backed writers (frozen v1 has no field) |
| Branded types + error family | `branded.go`, `errors.go` | `RepoName`/`SourceName` via go-branded-id; classified errors via go-error-family |

Hub module: realtime WebUI (Datastar over SSE), `compliance serve|report|table` CLI, SQLite evidence cache, fleet discovery client (`never-covered`), health probes, systemd/sd_notify hardening. FEATURES.md marks core (9 rows) and hub (13 rows) FULLY_FUNCTIONAL against the 2026-11-18 tree; README states "Status: built, pre-release."

**It is a compliance-posture/evidence hub — a *consumer of run evidence*, not a linter and not a linting library.**

## Relevance to go-linter-sdk

### Reverse direction (checked, per audit instructions)

`rg 'go-finding|go-linter-sdk' --type go` across `/home/lars/projects/go-compliance` finds **zero imports**. The only hits are a doc-comment example (`branded.go:9,12` — "e.g. 'go-finding'") and test fixtures using `"go-finding"` as a sample repo name. Neither go.mod requires either module.

And that is *correct*, not a gap: go-compliance's seam is the **frozen v1 JSON document**, deliberately schema-pinned and golden-tested, exactly so that producers and consumers never share Go types. Making go-compliance import go-linter-sdk would add a Go-coupling its design explicitly avoids.

### Forward direction (SDK → go-compliance): already integrated, transitively, for free

The compliance pipeline already runs on SDK output with zero SDK changes:

> linter rules (`linter.RuleFunc`) → `DetectorFromRegistry` (`/home/lars/projects/go-linter-sdk/registry.go:242`, "plugs into BuildFlow's DAG with zero glue") → BuildFlow step → `buildflow compliance --json` (v1 document) → `go-compliance` `FileSource` (`adapter.go:46`) → hub dashboard.

The SDK is upstream of go-compliance *through BuildFlow*, which is the integration point both sides designed for (`docs/guides/integrate.md`: "Zero-work: BuildFlow projects ... Nothing to do").

The only direct touchpoint that could ever involve SDK knowledge is the **non-BuildFlow path**: a standalone lint binary built on the SDK appending one reporter call to its `main`:

```go
// sketch of what a consumer linter's main() would do (NOT SDK code):
_ = compliance.Record(ctx, compliance.NewRepoName("my-linter"),
    compliance.RunResult{
        Verdict:        compliance.VerdictPassed, // or Failed, from ExitCodeFromReport
        GateFindings:   map[string]int{"lint": len(report.Findings)},
        FindingSeverities: severityRollup(report), // dropped by file writers; kept by hub cache
    })
```

That is a ~10-line snippet in a *consumer's* `main` (cf. `/home/lars/projects/go-linter-sdk/examples/minimal-linter/main.go`) or in go-compliance's own `examples/reporter/`. It is not SDK code, needs no SDK API change, and would cost the SDK its single-dep policy if it were anything more than documentation.

### What adoption would cost today

- Code dep: two new direct deps in the core path (go-branded-id, go-error-family) plus alignment churn with an untagged, daily-auto-committing tree — direct violation of `AGENTS.md:187-191` single-dep/layering policy.
- Docs recommendation today: would point consumers at a module that cannot be `go get`-resolved (no tags; hub still on a `replace ../` dev pin, with TODO_LIST row "Drop hub replace directive when core tags v0.1.0").

## Analysis

The right relationship between these two repos is the one that already exists: **go-compliance is a downstream fleet consumer of evidence that SDK-built linters (via BuildFlow) already produce.** Direction matters: compliance posture is an *aggregation* concern (windows, staleness, fleet discovery), the SDK is a *production* concern (rules → findings → exit codes). Mixing them means every compliance contract change version-bumps every linter, and every SDK release pressures the frozen v1 contract. Keeping the seam at JSON is what lets "v1 stays emittable forever" (ROADMAP) coexist with a fast-moving SDK.

CONSUMER-DIRECTION was considered and rejected on evidence: go-compliance should *not* import go-linter-sdk — it consumes documents, not types. The genuine consumer-direction candidate is one level down: **individual lint binaries** built on the SDK may become go-compliance reporters via the one-liner, once go-compliance is tag-resolvable.

## The fleet argument

Owner: "I have MANY projects to work on, and limited time and resources; If all my project re-implement/reintegrate everything I will only fix bugs all day long and none of my projects can benefit from collective improvements."

go-compliance exists precisely to end per-project compliance re-integration — and the SDK already benefits from it *without a single line of integration*: every improvement to SDK-based linters makes BuildFlow gates (and thus fleet compliance evidence) better automatically; every improvement to go-compliance's fold/staleness/dashboard reaches every SDK-built linter's reporting for free. That is the collective-improvement loop working through a frozen contract. The one way to *break* the loop is to weld the two repos together with a Go dependency, re-creating the coupling the v1 contract was designed to eliminate.

## Ranked opportunities (impact × ease)

1. (Blocked, docs-scope) Note the fleet-compliance reporting snippet for standalone SDK-built linters — routing: ROADMAP line under "Consumer adoption": "When go-compliance core tags v0.1.0 and a standalone (non-BuildFlow) linter wants fleet visibility, add a short docs note: `compliance.Record` with `GateFindings{"lint": len(report.Findings)}` derived from `ExitCodeFromReport` — docs-only, never a dependency." Not a TODO_LIST row: it is blocked on an external event (the tag), which is ROADMAP territory by the SDK's own definitions.
2. None today — no code, no tooling, no current docs change.

## Flip trigger

go-compliance core tags **v0.1.0** (proxy-resolvable; the hub's own TODO gates on exactly this) **and** at least one SDK-built linter runs *outside* BuildFlow but needs fleet visibility. Then the verdict flips to **ADOPT-AS-DOCS**: a "Fleet compliance reporting" snippet in the SDK README/ADOPTION showing the `Record` one-liner (with the honest caveat that `FindingSeverities` is dropped by file-backed writers, `reporter.go:38-41`). It never flips to ADOPT: the single-dep policy and the frozen-JSON seam both forbid a code dependency, tag or no tag.

## How to verify

```bash
git -C /home/lars/projects/go-compliance tag -l | wc -l && git -C /home/lars/projects/go-compliance log -1 --format='%h %ad %s' --date=short
# 0
# ec61ad8 2026-11-19 chore: auto-commit 1 changed file(s) (heuristic)
# (no tags; HEAD moved 2a62a68 -> 1861917 -> ec61ad8 during the audit session)

find /home/lars/projects/go-compliance -maxdepth 3 -name go.mod -not -path '*/vendor/*'
# /home/lars/projects/go-compliance/hub/go.mod
# /home/lars/projects/go-compliance/go.mod

rg -n 'go-finding|go-linter-sdk' /home/lars/projects/go-compliance --type go | grep -v '_test.go' || echo "NO NON-TEST MATCHES"
# /home/lars/projects/go-compliance/branded.go:9:// Name implements the branded-id naming interface ("Repo:go-finding").
# /home/lars/projects/go-compliance/branded.go:12:// RepoName identifies one repository in the fleet (e.g. "go-finding").
# (doc-comment examples only — zero Go imports of go-finding or go-linter-sdk)

grep -n 'replace' /home/lars/projects/go-compliance/hub/go.mod
# 133:// Dev-only until core is tagged: the release procedure drops this replace
# 135:replace github.com/larsartmann/go-compliance => ../

rg -n 'go-compliance' /home/lars/projects/go-linter-sdk -g '!docs/review/**' || echo "NO MATCHES"
# NO MATCHES (exit 1) — zero references in the consumer repo

sed -n '3,7p' /home/lars/projects/go-linter-sdk/go.mod
# go 1.27
#
# require github.com/larsartmann/go-finding v1.15.1
#
# require github.com/larsartmann/go-error-family v0.12.0 // indirect
```

> **Re-review trigger:** re-run this audit when the consumer pins a newer
> release of the reviewed library than the version stamped above, or when a
> follow-up task from the 2026-11-19 Pareto plan lands in this repo
> (footers standardized 2026-11-19, plan task M22/F22.1).
