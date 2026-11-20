# go-policy-dsl utilization review — go-linter-sdk

- **Reviewed:** 2026-11-19 by fleet audit Group B (sub-agent; orchestrated session)
- **Library:** github.com/larsartmann/go-policy-dsl
- **Sub-modules:** single module at repo root (`go.mod` line 1; no nested modules); consumer is also single-module
- **Version reviewed:** library repo `v0.4.0`, HEAD `e791b456fdf58bfc8aa137660009fb419ae8f43e` 2026-11-02; consumer pin: none (not in go.mod); point-in-time snapshot
- **Verdict:** NOT-ADOPTED — a fluent, zero-dependency DSL for _library-governance_ policies; the right consumers are dependency-guard linters (cmdguard/lint, library-policy — the latter already shipped on it), not this SDK, whose Rule surface is code-lint rules and would gain a competing Severity/Category vocabulary
- **Score:** no score (not adopted)

## What the library offers

Policy-as-code: **fluent, compile-time-checked Go values** for declaring library governance policies — bans, requirements, companions. "Pure Go values, no YAML, no codegen, no runtime parsing" (README.md:7). Verified from source:

| Capability                                                                                                                       | Evidence (library repo)                                                                                                                                                                             |
| -------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Fluent ban-policy builder with typed defaults (SeverityCritical + CategorySecurity + ModeBan)                                    | `builder.go:26` `Ban(name) *Builder`; `Because/WithSeverity/WithCategory/DetectVia/VersionRange/Suggest/...` chain in `builder.go:33-106`                                                           |
| Immutable `PolicySpec` result value                                                                                              | `policy.go:168` `PolicySpec{Name, Reason, Severity, Category, Detection, Description, Alternatives, CVEs, VersionMin/Max, Companions, Mode}`                                                        |
| Version constraints via branded semver-lite type (no semver dep); inverted ranges rejected by `Validate()` with a concrete error | `version.go:18` `Version`, `version.go:49` `ParseVersion`, `policy.go:203` `ErrInvertedVersionRange` + `policy.go:211` `InvertedVersionRangeError`                                                  |
| Validated CVE identifiers (`CVE-YYYY-NNNN` branded string)                                                                       | `cve.go:14` `CVE`, `cve.go:26` `NewCVE`                                                                                                                                                             |
| Detection model with **no execution semantics** — patterns are opaque strings, fuzz-pinned                                       | `policy.go:86` `Detection{ImportPatterns, GoModPatterns, ExcludeIfContains, ExcludeIfTransitiveFrom, RequireIfContains}`; `FuzzBuilder_PatternsOpaque` (ROADMAP.md "matching / execution boundary") |
| Companion policies ("if you use X you must also use Y"), incl. companion-only mode that never bans                               | `builder.go:232` `Companion`, `policy.go:132` `CompanionSpec`, `policy.go:151` `Mode` (`ModeBan` / `ModeCompanionOnly`), `AsCompanionOnly()`                                                        |
| Replacement suggestions preserving per-entry reasons                                                                             | `builder.go:252` `NewReplacement`, `policy.go:122` `Replacement`, `Suggest`/`SuggestExplicit`/`WithAlternatives`                                                                                    |
| Own Severity (8 values) and Category (9 values) string enums — deliberately NOT the consumer's types                             | `policy.go:31` `Severity`, `policy.go:70` `Category`; README.md:162 "Severity is a string, not your consumer's type"                                                                                |
| **Zero dependencies** — stdlib only, go 1.27                                                                                     | library `go.mod`: module + `go 1.27`, no require blocks                                                                                                                                             |
| Shipped consumer: `library-policy` imports v0.3.0 and deleted its local `domain/policy/spec.go` copy                             | library ROADMAP.md: "First consumer shipped. MET (2026-08-08)"                                                                                                                                      |

Pre-v1 status matters: its ROADMAP states "Until v1.0.0: breaking changes are allowed in any `0.x` bump" — v0.4.0 reviewed here is three minors past the consumer's pinned v0.3.0.

## Relevance to go-linter-sdk

- **The library's own roadmap explicitly courts this repo**: ROADMAP.md:50-52 — "### `go-linter-sdk` adoption (secondary consumer) … `go-linter-sdk` may adopt this as its rule-declaration language. No code yet." So the question is not hypothetical; the DSL's author expects this evaluation. Verdict input: the fit is weaker than the roadmap hopes.
- **Domain mismatch.** The SDK's `Rule` (`rule.go:95-102`: ID/Name/Description/Category/Severity/IsEnabledByDefault/Check) declares **code-lint rules** about Go source in a directory. The DSL declares **library-governance policies** about imports and go.mod lines, with no detection engine (README.md:195: "You need a consumer (a CLI, a linter, an LSP server) that reads `PolicySpec` and implements the matching"). They are complementary layers, not the same layer: a _dependency-governance linter built on the SDK_ would be the joint product — that consumer is `cmdguard/lint`, which is "born native" on the SDK (`docs/ADOPTION.md:17`), not the SDK itself.
- **Vocabulary collision in the published API.** The SDK already publishes `linter.Category` (`rule.go:48`, 11 values) and leans on `finding.Severity`. The DSL publishes its own `policydsl.Severity`/`policydsl.Category`. Importing it would put two severity taxonomies inside the SDK's surface with no bridging helper on either side — every consumer would face the mapping (policydsl `critical`…`info` → finding.Severity) alone. That is added concept weight for six consumers, none of which declares library policies today.
- **Cost as a code dep.** Uniquely among the fleet's candidates, this one is **zero-dependency** — importing it would add zero transitive tree and zero module risk. But the single-dep policy is stated absolutely (AGENTS.md:187-193: "its single-dep policy (`go-finding only)`"), not "go-finding plus zero-tree libs"; adoption would still require an explicit AGENTS.md amendment, and would spend the policy's credibility on a problem the SDK does not have.
- **Cost as docs/tooling:** none available — the DSL is a library, not a tool; there is nothing to wire into flake.nix or CI.

## Analysis

The honest assessment: the DSL solves a real problem (stringly-typed YAML governance policies) for a repo that has that problem. This repo does not. The SDK contains no policy-declaration code to deduplicate; nothing in `.golangci.yml`, `rule.go`, `registry.go`, or the six consumers' integration notes mentions policies or bans. The fleet's dependency-governance consumers are the right evaluators — and one (`library-policy`) has already adopted it and deleted its local copy, which is the fleet argument working. A README "Companion libraries" row today would recommend a pre-v1 API (breaking changes allowed per its ROADMAP) for a concern only one consumer segment has — doc-drift risk now, benefit deferred. The right move is to decline the roadmap's invitation _explicitly_ (this report serves as that record) and revisit at v1.0.0.

## The fleet argument

Owner: "I have MANY projects to work on, and limited time and resources; If all my project re-implement/reintegrate everything I will only fix bugs all day long and none of my projects can benefit from collective improvements."

The collective-improvement event for this library has already happened where it belongs: `library-policy` adopted the DSL and deleted its local spec copy (library ROADMAP gate 1, MET 2026-08-08). The fleet wins when the _next_ governance consumer (cmdguard/lint, the golangci-plugin template sketched in the DSL's ROADMAP) imports the same typed `PolicySpec` — not when a lower-layer SDK re-exports it to consumers that don't govern dependencies. Adopting here would be re-integration without re-implementation savings: pure coupling cost, the exact trap the owner's concern describes in reverse.

## Ranked opportunities (impact × ease)

1. **Deferred companion-library mention** — once go-policy-dsl tags v1.0.0 (API stable, validation surface settled per its ROADMAP gates), add a README "Companion libraries" row: "Dependency-policy declaration — go-policy-dsl (for linters that enforce library governance)", optionally with a severity-bridging snippet — impact M (routes cmdguard/lint-style consumers to a typed DSL), ease S — routing: TODO_LIST row "Revisit go-policy-dsl README companion-libs mention when it tags v1.0.0"
2. **Direct cmdguard/lint evaluation** (outside this repo) — the born-native dependency-guard consumer is the natural second adopter; evidence for its fitness already exists in the DSL's behaviour suite — impact M, ease M — routing: none (other repo's TODO list)

## Flip trigger

A concrete SDK consumer needs typed ban/companion policy declarations: `cmdguard/lint` adopts the DSL for its rules, or the fleet builds the golangci-plugin template (sketched in the DSL's ROADMAP) on top of go-linter-sdk, or the SDK itself grows a curated dependency-governance rules package. Then adopt it as the declaration language — zero transitive deps make it the cheapest possible second dependency, but the adoption must amend AGENTS.md's single-dep policy explicitly (it is stated as `go-finding` only) and ship a documented `policydsl.Severity → finding.Severity` bridge so consumers never write the mapping by hand.

## How to verify

```bash
cd /home/lars/projects/go-policy-dsl && git describe --tags && cat go.mod
# v0.4.0
# module github.com/larsartmann/go-policy-dsl
#
# go 1.27
sed -n '168p' policy.go && sed -n '26p' builder.go
# type PolicySpec struct {
# func Ban(name string) *Builder {
sed -n '50,52p' ROADMAP.md
# ### `go-linter-sdk` adoption (secondary consumer)
#
# `go-linter-sdk` may adopt this as its rule-declaration language. No code yet.
cd /home/lars/projects/go-linter-sdk && rg -n 'policydsl|policy-dsl' go.mod go.sum *.go || echo "policy-dsl: not a dependency"
# policy-dsl: not a dependency
rg -n 'require github.com/larsartmann/go-finding v' go.mod
# 5:require github.com/larsartmann/go-finding v1.15.1
sed -n '95,102p' rule.go   # the SDK's Rule interface: code-lint rules, not library policies
# type Rule interface {
#   ID() string
#   Name() string
#   Description() string
#   Category() Category
#   Severity() finding.Severity
#   IsEnabledByDefault() bool
#   Check(ctx context.Context, dir string) ([]finding.Finding, error)
# }
```

> **Re-review trigger:** re-run this audit when the consumer pins a newer
> release of the reviewed library than the version stamped above, or when a
> follow-up task from the 2026-11-19 Pareto plan lands in this repo
> (footers standardized 2026-11-19, plan task M22/F22.1).
