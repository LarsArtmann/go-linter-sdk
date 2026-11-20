# go-dag-app utilization review — go-linter-sdk

- **Reviewed:** 2026-11-19 by fleet audit Group C (sub-agent; orchestrated session)
- **Library:** github.com/larsartmann/go-dag-app
- **Sub-modules:** single module at repo root (`/home/lars/projects/go-dag-app/go.mod`, Go 1.27 directive, sole dep `github.com/samber/do/v2 v2.1.0` + one indirect). Consumer is single-module; layouts match, but that is the only thing that matches.
- **Version reviewed:** library repo `v0.1.1` (git describe; HEAD `d9934c8` 2026-11-02, tag sits on HEAD); consumer pin: none (not in go.mod); point-in-time snapshot
- **Verdict:** NOT-ADOPTED — a 3-function DI-container bootstrap for long-running `samber/do` apps; go-linter-sdk is a pure library with no injector to bootstrap, and adopting it would pull the exact `samber-do` transitive tree the SDK's layering decision explicitly refuses
- **Score:** no score (not adopted)

## What the library offers

Despite the name, go-dag-app is **not a DAG framework**. Verified from source (`go doc .` and `/home/lars/projects/go-dag-app/container.go`), the entire exported API is three functions and one constant — a shared bootstrap skeleton for apps that run on `samber/do` v2 DI containers:

| Capability                                                        | Evidence          | What it actually does                                                                                                                               |
| ----------------------------------------------------------------- | ----------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| Per-app debug env gate `IsDebugDIEnabled(appName)`                | `container.go:34` | True when `<APP>_DEBUG_DI` is `"1"`/`"true"` (case-insensitive, trimmed)                                                                            |
| Dependency-tree printer `PrintDependencyTree(injector)`           | `container.go:49` | Prints `do.ExplainInjector` output to stdout; nil-safe                                                                                              |
| Container bootstrap `NewContainer(appName) (do.Injector, func())` | `container.go:75` | Fresh `do.New()` injector + cleanup that prints the tree (when gated) then `injector.Shutdown()`                                                    |
| Package doc declaring scope                                       | `container.go:1`  | "Deep DAG features (dependency-graph compilation, retry, resume, audit-log diffing) stay in each app — This package is only the bootstrap skeleton" |

The README and FEATURES.md confirm the positioning: consumers are **BuildFlow** (`internal/di`) and **library-policy** (`di`), both long-running apps that previously copy-pasted this plumbing. FEATURES.md's adoption table shows both consumers delegate `IsDebugDIEnabled`/`PrintDependencyTree` but do **not** use `NewContainer` yet. CHANGELOG `[0.1.0] 2026-07-18`: "shared DI-container bootstrap extracted from BuildFlow and library-policy".

Also relevant: `/home/lars/projects/go-dag-app/LICENSE` is **PROPRIETARY** (the SDK is MIT).

## Relevance to go-linter-sdk

Near zero, on every axis:

1. **Nothing to bootstrap.** The SDK has no DI container, no injector, no shutdown sequence. Its entire runtime surface is `Registry.Run` → `Report` → `ExitCodeFromReport` inside a CLI process that lives for milliseconds. There is no place any of the three functions could plug in.
2. **The DAG overlap concern is a category error, and it is already resolved in the right direction.** The DAG orchestration in this fleet lives in BuildFlow, which **already imports go-linter-sdk** (`/home/lars/projects/go-linter-sdk/docs/ADOPTION.md`, BuildFlow row: runs sibling registries in-process via `Registry.Run`, deleted ~4,737 LOC of its own violation hierarchy). The SDK plugs into BuildFlow's DAG through `DetectorFromRegistry` (`/home/lars/projects/go-linter-sdk/registry.go:234-241`: "adapts a registry to the canonical finding.Detector interface, so a linter plugs into BuildFlow's DAG with zero glue"). go-dag-app is a _third-party utility consumed by BuildFlow inside BuildFlow's process_. The SDK taking a dependency on it would cross a boundary that composition already handles for free.
3. **Dependency math forbids it.** `/home/lars/projects/go-linter-sdk/AGENTS.md:187-191` (decided 2026-09-17): the single-dep policy "would be destroyed by cmdguard's transitive cobra/fang/lipgloss/**samber-do**/koanf tree." go-dag-app's one direct dependency _is_ `samber/do/v2`. Adopting go-dag-app imports precisely the transitive tree the layering decision names as disqualifying.
4. **License mismatch.** A proprietary dependency inside a public MIT module is a policy problem independent of capability.

## Analysis

The right integration point for anything go-dag-app offers is **BuildFlow** (and library-policy), where it already lives. For the SDK, the correct amount of go-dag-app is zero: no code dep, no tooling use (nothing in the SDK's dev loop runs a DI container — `nix run .#test`/`.#lint` are plain Go commands), and no docs recommendation (SDK consumers building CLI linters have no injector either; consumers building _services_ would meet go-dag-app through BuildFlow's own docs).

The only conceivable future hook: if the SDK ever ships an `examples/` entry for a long-running service (a `linterd`-style daemon with a DI container), `dagapp.NewContainer("LINTERD")` would be the fleet-idiomatic bootstrap for that example binary — dev-scope only, never the published API. That is hypothetical; nothing in the SDK's ROADMAP themes (API maturation, consumer adoption) points there.

## The fleet argument

Owner: "I have MANY projects to work on, and limited time and resources; If all my project re-implement/reintegrate everything I will only fix bugs all day long and none of my projects can benefit from collective improvements."

go-dag-app is the fleet argument **working correctly, elsewhere**: BuildFlow and library-policy stopped copy-pasting DI plumbing and now share it — collective improvement flows to both when the bootstrap changes. The same argument tells the SDK to stay out: if the SDK re-integrated this scaffold (or grew its own container helper), it would create a _second_ thing to maintain in a repo whose charter is "shared scaffolding for rules/findings", duplicating a seam BuildFlow already owns. Not adopting is not re-implementation — the SDK has nothing to re-implement. Every bug fixed in go-dag-app already reaches the SDK's most important consumer (BuildFlow) without the SDK lifting a finger.

## Ranked opportunities (impact × ease)

1. None — routing: none. No TODO_LIST row, no ROADMAP line. There is no bounded work item for the SDK here; even a docs mention would add drift surface for zero consumer value.

## Flip trigger

The SDK ships (or documents) a long-running daemon/service example that uses `samber/do` — i.e. `examples/` gains something like `linterd` with an injector and a shutdown sequence. Then adopt go-dag-app **in that example only** (`dagapp.NewContainer`, dev-scope, still never a published-API dependency). Secondary trigger: go-compliance/hub-style service consumers start asking the SDK docs "how do I wire a DI container?" — same answer, docs-scope.

## How to verify

```bash
git -C /home/lars/projects/go-dag-app describe --tags && git -C /home/lars/projects/go-dag-app log -1 --format='%h %ad %s' --date=short
# v0.1.1
# d9934c8 2026-11-02 chore(config): converge golangci-lint config via golangci-lint-auto-configure

find /home/lars/projects/go-dag-app -maxdepth 2 -name go.mod -not -path '*/vendor/*'
# /home/lars/projects/go-dag-app/go.mod        (single module)

grep -n '^func \|^const \|^// Package' /home/lars/projects/go-dag-app/container.go | head -8
# 1:// Package dagapp provides the shared DI-container bootstrap used by
# 29:const EnvSuffix = "_DEBUG_DI"
# 34:func IsDebugDIEnabled(appName string) bool {
# 49:func PrintDependencyTree(injector do.Injector) {
# 75:func NewContainer(appName string) (do.Injector, func()) {
# (entire exported API = 3 funcs + 1 const)

rg -n 'go-dag-app|dagapp' /home/lars/projects/go-linter-sdk -g '!docs/review/**' || echo "NO MATCHES"
# NO MATCHES (exit 1) — zero references in the consumer repo

grep require /home/lars/projects/go-linter-sdk/go.mod
# require github.com/larsartmann/go-finding v1.15.1
# require github.com/larsartmann/go-error-family v0.12.0 // indirect
# (single-dep policy intact)

grep -n -A4 'Layering vs sibling libs' /home/lars/projects/go-linter-sdk/AGENTS.md
# 187:- **Layering vs sibling libs (decided 2026-09-17).** The SDK deliberately does
# 188-  NOT depend on `cmdguard` (CLI framework) or `go-output` (renderer): it is a
# 189-  library that stops at findings/report/exit codes, and its single-dep policy
# 190-  (`go-finding` only) would be destroyed by cmdguard's transitive cobra/fang/
# 191-  lipgloss/samber-do/koanf tree. ...
```

> **Re-review trigger:** re-run this audit when the consumer pins a newer
> release of the reviewed library than the version stamped above, or when a
> follow-up task from the 2026-11-19 Pareto plan lands in this repo
> (footers standardized 2026-11-19, plan task M22/F22.1).
