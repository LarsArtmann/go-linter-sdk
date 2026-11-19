# linter-autoconfigure-sdk utilization review — go-linter-sdk

- **Reviewed:** 2026-11-19 by fleet audit Group B (sub-agent; orchestrated session)
- **Library:** github.com/larsartmann/linter-autoconfigure-sdk
- **Sub-modules:** single module at repo root (`go.mod` line 1; no nested modules); consumer is also single-module
- **Version reviewed:** library repo `v0.8.1-5-gd42d4ff`, HEAD `d42d4ff557979968059f3e4f6e3c2e91e61009db` 2026-11-15; consumer pin: none (not in go.mod); point-in-time snapshot
- **Verdict:** NOT-ADOPTED — go-linter-sdk builds linters, not auto-configurers; this repo already receives the library's value *through* its downstream tool (`golangci-lint-auto-configure`) as an external dev tool, and importing the library would 5× the SDK's dependency footprint in its published API
- **Score:** no score (not adopted)

## What the library offers

A shared foundation for building **linter auto-configuration tools** (verified from source, not from the name):

| Capability | Evidence (library repo) |
| --- | --- |
| Config file I/O: read, typed JSON load/parse, canonical deterministic marshal, atomic idempotent save | `autoconfigure.go:99` `ReadConfig`, `autoconfigure.go:109` `LoadJSON[T]`, `autoconfigure.go:128` `ParseJSON[T]`, `autoconfigure.go:155` `MarshalJSONIndented`, `autoconfigure.go:181/213` `SaveJSONBytes`/`SaveJSON` (atomic + crash-durable via go-atomic-write); YAML deliberately out of scope (FEATURES.md "Config I/O") |
| Typed config errors (`ConfigError` + `Op` enum, errors.Is/As chain traversal) | `autoconfigure.go:42` `Op`, `autoconfigure.go:64` `ConfigError` |
| Config-issue → finding emission (priority→Severity, suggestion→FixStrategySuggest) | `autoconfigure.go:247` `ConfigIssue`, `autoconfigure.go:288` `FindingFromIssue`, `autoconfigure.go:336` `FindingsFromIssues` |
| Config diff engine (maps/sets/blobs, deterministic path-sorted `Change` model, `Summary`/`FormatDiff`) | `diff.go:43` `DiffMaps`, `diff.go:71` `DiffSets`, `diff.go:103` `DiffBlobs`, `diff.go:126/147` `Summary`/`FormatDiff` |
| Multi-candidate config discovery | `autoconfigure.go:500` `FirstExisting` |
| BuildFlow provider adapter: spec closures → canonical `toolsdk.Spec` (`toolsdk.Register`-ready) | `autoconfigure.go:360` `ProviderSpec`, `autoconfigure.go:437` `ProviderFromSpec` |
| Bootstrap lifecycle: derives missing-only Detect, never-overwrite Repair, advisory drift HealthCheck from a generic spec | `bootstrap.go:51` `BootstrapSpec[T]`, `bootstrap.go:128` `BootstrapProviderFromSpec` |
| `jsondeterminism` analyzer + vettool (flags bare json/v2 Marshal of non-deterministic types) | `determinism/determinism.go` (`NewAnalyzer`), `cmd/jsondeterminism` (singlechecker main); FEATURES.md "Determinism enforcement" |
| Existing consumers: `oxlint-auto-configure` (bootstrap lifecycle, config I/O, discovery, diff) and `golangci-lint-auto-configure` (health findings via `FindingFromIssue`, diff vocabulary, jsondeterminism gate) | README.md:259-272 "Consumers" section |

Its dependency reality (library `go.mod`): **4 direct deps** — `go-atomic-write v0.6.2`, `go-finding v1.15.1`, `go-finding/toolsdk v1.16.0`, `golang.org/x/tools v0.52.0` — plus indirect `go-error-family`, `xxhash`, `gofrs/flock`, `x/mod`, `x/sync`, `x/sys`.

## Relevance to go-linter-sdk

**The problem it solves is real for this repo — and already solved, at the tool layer.**

- This repo's `.golangci.yml` (127 enabled linters, extensive tuned settings) is **not hand-rolled in isolation**: commit `56b667e` (2026-11-02) is titled `chore(config): converge golangci-lint config via golangci-lint-auto-configure` (2-line change to `.golangci.yml`). The repo also ran `buildflow -s go-version-auto-configure --format finding` in the 2026-10-11 session (`docs/status/2026-10-11_07-11_buildflow-green-go127-and-jsonv2-cleanup.md:29,170`). So the consumer of `linter-autoconfigure-sdk` that matters here — `golangci-lint-auto-configure` — already operates on this repo as an **external tool**, and the value arrives without a single line of Go coupling.
- **Domain mismatch:** go-linter-sdk is scaffolding for *linters* (Rule/Registry/exit codes; `rule.go:95` `Rule` interface). linter-autoconfigure-sdk is scaffolding for *auto-configurers* (config round-trip, drift, bootstrap). None of the SDK's six consumers (go-structure-linter, samber-linter, go-humanize-linter, cmdguard/lint, project-meta/linter, BuildFlow/execution) is an auto-configurer, so the SDK re-exporting its types would serve zero current consumers.
- **Cost as a code dep:** the SDK's published API is single-dep by explicit decision — AGENTS.md:187-193: "its single-dep policy (`go-finding` only) would be destroyed by cmdguard's transitive cobra/fang/lipgloss/samber-do/koanf tree". Importing linter-autoconfigure-sdk would add `go-atomic-write` + `go-finding/toolsdk` + `golang.org/x/tools` (and flock/xxhash indirectly) to the *published* tree — a strictly worse violation than the one the policy already rejects, to gain nothing the SDK's consumers need.
- **Cost/benefit as tooling:** the `cmd/jsondeterminism` vettool has **no target here** — this repo's own Go code contains zero `json.Marshal` call sites (verified: `rg 'json\.Marshal' --glob '*.go'` → no matches). The downstream converge tool is already used manually; wiring it into flake.nix/CI would be adopting `golangci-lint-auto-configure` (a different repo), not this library.

## Analysis

The right integration point for this library is **the auto-configurer tools themselves** (where it already lives, with two migrated consumers) and, marginally, **this repo's manual tooling workflow** (where it already touches us via commit `56b667e`). It is not the SDK's published API and not the SDK's CI. Adopting it as a code dep would be the classic wrong-layer move: pulling a sibling domain SDK into a lower layer so a hypothetical future feature is one import away, while every one of the six real consumers pays the dependency cost forever. The library's own ROADMAP frames its success criterion as auto-configurer migration ("first real consumer migrates"), which has happened — twice — without go-linter-sdk's involvement.

The only genuine gap found: the converge workflow is tribal knowledge (visible only in git history and a status report). That's a documentation opportunity, not an import.

## The fleet argument

Owner: "I have MANY projects to work on, and limited time and resources; If all my project re-implement/reintegrate everything I will only fix bugs all day long and none of my projects can benefit from collective improvements."

This library is the fleet argument **working as intended** — for a different segment of the fleet. Two auto-configurers (`oxlint-auto-configure`, `golangci-lint-auto-configure`) deleted their duplicated plumbing into it (README.md:22-45 "Why?" table), and this repo already free-rides on the result: every future improvement to config convergence lands in `.golangci.yml` maintenance here without any integration work. Re-implementation pressure on go-linter-sdk itself: **zero** — the SDK has no config-file round-trip code to deduplicate. Forcing the import would invert the benefit: coupling without deleted code, i.e. the maintenance tax with none of the saving.

## Ranked opportunities (impact × ease)

1. **Document the `.golangci.yml` converge workflow** — a short AGENTS.md/CONTRIBUTING note that config convergence runs via `golangci-lint-auto-configure` (evidence: commit `56b667e`), so the practice survives the session that introduced it — impact M, ease S — routing: TODO_LIST row "Document the golangci-lint-auto-configure converge workflow (AGENTS.md tooling section)"
2. **Watch the ROADMAP's `configerr` sub-package extraction idea** (extracting `ConfigError`/`Op` without the domain types) — if it ships, a future config-audit feature could import far less — impact L (today), ease n/a — routing: none (other repo's roadmap)

## Flip trigger

go-linter-sdk grows a **config-audit surface**: e.g. a shipped Rule/helper that lints consumers' config files (`.golangci.yml` staleness, missing linters), or the fleet starts a third auto-configurer that wants the SDK's finding vocabulary directly. At that point `ConfigIssue`/`FindingFromIssue` (`autoconfigure.go:247,288`) become the shared plumbing worth importing — conditional on (a) an explicit AGENTS.md amendment to the single-dep policy accepting `go-atomic-write` + `toolsdk` + `x/tools` in the published tree, and (b) the `configerr` extraction idea having been decided (the lighter import path is strictly preferred).

## How to verify

```bash
cd /home/lars/projects/linter-autoconfigure-sdk && git describe --tags
# v0.8.1-5-gd42d4ff
sed -n '247p;288p;360p;437p' autoconfigure.go
# type ConfigIssue struct {
# func FindingFromIssue(toolName finding.ToolName, issue ConfigIssue) (finding.Finding, error) {
# type ProviderSpec struct {
# func ProviderFromSpec(spec ProviderSpec) (toolsdk.Spec, error) {
cd /home/lars/projects/go-linter-sdk && rg -n 'autoconfigure' go.mod go.sum || echo "autoconfigure: not a dependency"
# autoconfigure: not a dependency
rg -c 'json\.Marshal' --glob '*.go' . || echo "no json.Marshal call sites in consumer SDK"
# no json.Marshal call sites in consumer SDK
git log --oneline -i --grep='auto-configure'
# 56b667e chore(config): converge golangci-lint config via golangci-lint-auto-configure
```
