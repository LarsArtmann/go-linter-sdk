# go-finding/toolsdk utilization review — go-linter-sdk

- **Reviewed:** 2026-11-19 by fleet audit Group A (sub-agent; orchestrated session)
- **Library:** github.com/larsartmann/go-finding/toolsdk (sub-module, independently tagged)
- **Sub-modules considered:** root, toolsdk, pipeline, analysis, cmd/go-finding (go-finding repo); consumer is single-module and imports only the root module
- **Version reviewed:** consumer does not pin it (absent from go.mod/go.sum); library latest tag `toolsdk/v1.16.0` (2026-10-11); library repo HEAD `498498f` 2026-11-18 (`git describe`: `cmd/go-finding/v1.15.1-87-g498498f`); point-in-time snapshot
- **Verdict:** NOT-ADOPTED — deliberate and, at the library layer, correct: self-registration is a *linter-binary* concern, and the SDK's documented single-dep policy forbids importing go-finding sub-modules. The one thing the SDK owes this module is guidance, and today that is a single comment line
- **Score:** no score (not adopted) — 1 of 2 applicable rows green in the mini-inventory below

## What the library offers

Verified from source at /home/lars/projects/go-finding/toolsdk (latest tag v1.16.0; local HEAD adds only `RaceEnabled` (race_on.go) and lint annotations — `git diff --stat toolsdk/v1.16.0..HEAD -- toolsdk` = 8 files, 66+/23-, no API change).

**Mechanism (confirmed from source):** a tool repo declares its capabilities as one `Spec` and calls `toolsdk.Register(Spec{...})` in a package-level var / `init()`; BuildFlow discovers it via a blank import plus `toolsdk.All()` and converts it one-way into its internal tool domain (doc.go:1-10, 29-40; registry.go:19, 31, 54-58). The package compiles against go-finding types only — no BuildFlow import (design constraint, doc.go:47-49; go.mod requires just `github.com/larsartmann/go-finding v1.15.1`).

| Capability | Library evidence (file:line in go-finding repo) |
| --- | --- |
| Self-registration: `Register(Spec) Spec`, process-global `All()`, eager validation panics | toolsdk/registry.go:31, 54-58, 89-118 |
| `Spec`: Name, Description, Trigger, DependsOn, ModuleFanOut, SwitchCases, Inputs, **Outputs**, Options, **Detect `finding.Detector`**, Repair `Repairer`, HealthCheck | toolsdk/spec.go:16-93 (`Detect finding.Detector` at spec.go:84) |
| `Trigger` + constructors `OnGoFiles`/`OnGoModule`/`OnFiles`/`AnyLanguage`, `ValidLanguages`, `ErrInvalidTrigger` (validated at Register since v1.16.0) | toolsdk/spec.go:97-173; toolsdk/triggers.go:7, 17, 27, 37 |
| Per-run knobs: `Option`/`OptionKind`/`OptionValues`, `WithOptions`/`OptionsFromContext`, typed `String/Int/BoolFromContext`, `Spec.ValidateOptions` | toolsdk/options.go:29, 44, 100-109, 123, 142-156 |
| Repair contract: `Repairer`/`RepairerFunc`, structured `RepairResult` (ModifiedFiles, Facts), dry-run flag `WithDryRun`/`DryRunFromContext` | toolsdk/spec.go:197, 224, 240; toolsdk/dryrun.go:10, 17 |
| Conditional repair paths: `SwitchCase` (Suffix/FindingRuleIDs/ExcludedRuleIDs) | toolsdk/spec.go:181-193 |
| Test isolation for the process-global registry: `ResetForTest`, `SnapshotForTest`, `RestoreForTest`; `EnsureContext` | toolsdk/registry.go:62-87, 123-129 |
| DAG ordering declarations (`DependsOn`, `Inputs`/`Outputs` producer edges, `ModuleFanOut`) | toolsdk/spec.go:27-45, 60-71 |

Fleet reach (unreleased, local HEAD only — go-finding `docs/ecosystem.md:150-168`, code-search verified 2026-11-08): **14 of 23** go-finding consumer repos already pin toolsdk v1.16.0, including **branching-flow** — the very linter this SDK's `examples/naked-return-guard` was ported from. Linter binaries registering as BuildFlow providers is the established fleet pattern, not an edge case.

## What go-linter-sdk uses today

Zero imports — `rg -n 'toolsdk' --type go` in the consumer returns nothing, and `go list -m all` resolves only `github.com/larsartmann/go-finding v1.15.1`. What exists is two accurate pointer references:

- registry.go:234-237 — `DetectorFromRegistry` doc: "adapts a registry to the canonical finding.Detector interface, so a linter plugs into BuildFlow's DAG with zero glue (via the tool-sdk Spec.Detect field)". **Verified accurate**: `toolsdk.Spec.Detect` is typed `finding.Detector` (toolsdk/spec.go:84), and `DetectorFromRegistry` returns exactly that type (registry.go:242-262), including the `finding.WorkingDirFromContext` read that `Spec.Detect`'s contract requires (spec.go:83).
- README.md:127-134 — "Plug into BuildFlow's DAG" shows `detector := linter.DetectorFromRegistry(registry, "my-linter")` then `// Pass detector to toolsdk.Spec{ Detect: detector, ... }`. Accurate but comment-only: no runnable example, no mention that `Register` also demands `Name`, a trimmed `Description`, and a valid `Trigger` (registry.go:89-118 panics otherwise).

## Capability inventory & gap analysis

| Capability | Status | Evidence (consumer) | Evidence (library) | Impact |
| --- | --- | --- | --- | ---|
| Spec.Detect compatibility (zero-glue handoff) | **Fully Leveraged** (by design, without importing) | registry.go:242-262 returns `finding.Detector` with the right context contract | toolsdk/spec.go:83-84 | An SDK-based linter binary registers as a BuildFlow provider with `Detect: linter.DetectorFromRegistry(r, name)` — no adapter code on either side |
| Self-registration mechanism (`Register`/`All`/init) | Not Applicable | the SDK is a library, not a tool; registering *itself* would hijack every consumer's tool identity (Register is process-global, registry.go:19) | toolsdk/registry.go:31-58 | Correctly left to linter binaries |
| Trigger/Options/DryRun/Repairer/HealthCheck surface | Not Applicable | SDK has no per-run knobs, no repair surface (see go-finding.md row 11), no external deps to health-check | toolsdk/spec.go:88-92, 97-173; options.go; dryrun.go | Becomes relevant only if the SDK grows a fix/repair surface |
| Consumer guidance (how an SDK linter registers) | Partially Used | one doc-comment parenthetical (registry.go:236) + one README comment line (README.md:131) | doc.go:25-45 shows the full pattern | Medium: the missing 10 lines of boilerplate (`var Provider = toolsdk.Register(toolsdk.Spec{Name, Description, Trigger: OnGoFiles(), Inputs, Detect})`) are exactly what every fleet linter must currently re-derive; branching-flow already did it by hand |

## Version currency

Not pinned, so currency is about the target should adoption-in-docs happen: `toolsdk/v1.16.0` (2026-10-11) is the latest tag — Trigger validation at Register, typed option accessors, `RepairResult` provenance, `Spec.Outputs` (toolsdk/CHANGELOG.md `1.16.0`). Local HEAD adds no API beyond `RaceEnabled`. The module tracks the same Go 1.27 floor and the same go-finding v1.15.1 root as the consumer, so a doc snippet pinned to v1.16.0 semantics will not rot quickly.

## The fleet argument

Owner's concern: "I have MANY projects to work on, and limited time and resources; If all my project re-implement/reintegrate everything I will only fix bugs all day long and none of my projects can benefit from collective improvements."

toolsdk is the fleet's collective answer to "how does a tool join BuildFlow": one declarative Spec, one init-time Register, zero BuildFlow imports. The SDK already did its half — `DetectorFromRegistry` produces the exact `finding.Detector` the Spec wants — but the other half (the registration recipe) lives only in a comment. Every linter that adopts this SDK and wants BuildFlow membership must re-derive that recipe; 14 fleet repos have already been through it. A ~15-line README/ADOPTION snippet written once here would be inherited by all six direct consumers and the ~60 indirect importers. That is the difference between "compatible by accident of types" and "collective improvement by documentation".

## Ranked opportunities (impact × ease)

1. **Document the registration recipe** — extend README.md's "Plug into BuildFlow's DAG" with the complete, copy-pasteable snippet: `var Provider = toolsdk.Register(toolsdk.Spec{Name: "my-linter", Description: "...", Trigger: toolsdk.OnGoFiles(), Inputs: []string{"**/*.go"}, Detect: linter.DetectorFromRegistry(registry, "my-linter")})` plus the blank-import note. Docs-only, no go.mod change, no policy change. — routing: TODO_LIST row "Add BuildFlow-provider registration recipe to README/ADOPTION (toolsdk v1.16.0)" / none
2. **Decide the runnable-example question explicitly** — an `examples/buildflow-provider` would require `toolsdk` in the root module's go.mod (examples/ are packages of this module), which the single-dep policy currently forbids (AGENTS.md:64-69). Either record the decision as "docs-only, permanent" or carve an exception. — routing: none (decision, not code)
3. **If the SDK grows a fix surface** (go-finding.md opportunity 3), model it on toolsdk's `Repairer` + `WithDryRun`/`DryRunFromContext` contract so SDK rules can graduate to BuildFlow repair steps without a second design pass. — routing: ROADMAP line "Rule-level fixes should target the toolsdk Repairer/dry-run contract"

## Why (not) adopted / flip trigger

Deliberate non-adoption, and structurally sound: (a) a library must not self-register a process-global tool identity its consumers never chose; (b) the SDK's single-dep policy (go-finding root only) is written down in AGENTS.md:64-69 and reinforced by the layering decision (AGENTS.md:188-196) — importing toolsdk would add a second, independently-tagged module to go.mod and couple this repo's releases to toolsdk's cadence; (c) toolsdk itself is designed so tools (binaries), not libraries, register. Integration correctly belongs to the consumer CLIs.

**Flip triggers:** (1) two or more SDK-based linters carry hand-copied identical Spec boilerplate → ship the docs recipe (opportunity 1); (2) the repo wants a compiled example → the single-dep policy must be consciously amended first (opportunity 2); (3) the SDK adds a repair surface → adopt toolsdk's Repairer/dry-run contract as the target shape (opportunity 3). Until any of these, non-adoption is the right state.

## How to verify

```bash
cd /home/lars/projects/go-linter-sdk
rg -n 'toolsdk' --type go; echo "exit=$?"
# exit=1 (no matches) — toolsdk is not imported anywhere
GOWORK=off go list -m all | grep finding
# github.com/larsartmann/go-finding v1.15.1   (only the root module)

git -C /home/lars/projects/go-finding tag -l 'toolsdk/*' | sort -V | tail -3
# toolsdk/v1.15.0
# toolsdk/v1.16.0
git -C /home/lars/projects/go-finding diff --stat toolsdk/v1.16.0..HEAD -- toolsdk | tail -1
# 8 files changed, 66 insertions(+), 23 deletions(-)   (RaceEnabled + lint annotations only)

# the zero-glue claim, both sides:
sed -n '84p' /home/lars/projects/go-finding/toolsdk/spec.go
#	Detect finding.Detector
sed -n '234,237p' /home/lars/projects/go-linter-sdk/registry.go   # the "tool-sdk Spec.Detect field" doc comment — accurate
rg -n 'toolsdk' /home/lars/projects/go-linter-sdk/README.md       # README.md:131 — the one-line recipe pointer
rg -n 'toolsdk v1.16.0 ×14|branching-flow|go-linter-sdk ' /home/lars/projects/go-finding/docs/ecosystem.md | head -4
# 153: ...toolsdk v1.16.0 ×14,   158: | branching-flow | v1.15.1 | — | — | v1.16.0 |
# 168: | go-linter-sdk | v1.15.1 | — | — | — |        (unreleased local HEAD doc)
```
