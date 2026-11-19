# go-error-family utilization review — go-linter-sdk

- **Reviewed:** 2026-11-19 (orchestrating session; follow-up to the seven-library fleet review — this is the forgotten 8th library: it is in this repo's `go.sum` but was not on the fleet list)
- **Library:** github.com/larsartmann/go-error-family (root module)
- **Sub-modules considered:** single root module (classification core is stable; `agent`, `diagnose`, `bridge` packages are experimental per `doc.go` "Stability" section)
- **Version reviewed:** v0.12.0 (indirect dependency via go-finding; `go mod why` → `go-linter-sdk → go-finding → go-error-family`); point-in-time snapshot
- **Verdict:** NOT-ADOPTED (deliberate, and doubly so: the SDK never imports it, and its own error vocabulary solves a different problem)
- **Score:** not scored — zero direct usage; the interesting question is whether the SDK *should* adopt it, and the answer is no today

## What the library offers

Read from the module-cache copy of v0.12.0
(`$(go env GOMODCACHE)/github.com/larsartmann/go-error-family@v0.12.0/`).

| Capability | Library evidence (file) |
| --- | --- |
| Four-interface error protocol (`Coded`, `Classified`, `Contextual`, `Retryable`) — "share the protocol, not the implementation" | `interfaces.go` (3 types), `doc.go:7-13` |
| Behavioral families: transient / infrastructure / rejection / conflict / corruption / orchestration | `family.go`, constructors in `constructors.go` (`NewTransient`, `WrapConflict`, …) |
| ~60 exported functions: constructors + `Wrap`/`Wrapf`/`WrapOnce` variants per family | `constructors.go` (16 `Wrap*`, 8 `New*`) |
| Classification registry: custom classifiers, templates, stdlib defaults | `registry.go`, `classify.go` (`RegisterClassifier`, `RegisterStdlibDefaults`) |
| Boundary helpers: `ExitCode`, HTTP handler, gRPC unary interceptor, log helpers | `http.go`, `handle.go`, `log.go` |
| Host-portable retryability query (`IsRetryable`) | `classify.go` |

## What go-linter-sdk uses today

**Nothing.** `rg -l 'go-error-family' *.go` → no matches; the dependency
appears only because `go-finding` (the single direct dependency) imports it.
The SDK's own error model is deliberately different in kind:

| SDK error concept | Lives at | Why go-error-family is not it |
| --- | --- | --- |
| `RuleError` + `ErrRuleFailed` | `errors.go` | Rule **attribution** (which rule failed), not behavioral classification (retry/no-retry) |
| `ErrMissingFields` / `ErrInvalidCategory` | `rule.go` | Registration-time programming errors — deterministic, never retryable |
| `ExitCodeFromReport` / `ExitCodeByConfidence` | `registry.go` | Maps **findings** to exit codes, not errors |

Rule failures are deterministic defects in a rule's closure — classifying them
into transient/infrastructure families would add a taxonomy with no consumer.
The library's own `doc.go` blesses this shape: "Projects with domain-specific
error shapes build their own struct and implement only the interfaces they
need."

## Gap grading

| Capability | Grade | Reason |
| --- | --- | --- |
| Error families / classification | Not Applicable | SDK errors are deterministic programming errors, not runtime behaviors |
| `Coded`/`Classified` protocol | Not Applicable | `*RuleError` already answers the only question consumers ask (which rule) via `errors.AsType` |
| `ExitCode` / HTTP / gRPC boundary helpers | Not Applicable | The SDK stops at findings/report/exit-codes by layering (see DOMAIN_LANGUAGE "What is NOT in this domain"); it owns no HTTP or RPC surface |
| Log helpers | Not Applicable | A library must not log (the consumer owns logging) |
| `IsRetryable` | Not Applicable | Nothing in the SDK retries; the pipeline module (a separate repo concern) owns retry policy |

## Ranked opportunities

None today. The single honest observation: if the parked `cmd/` CLI
(ROADMAP theme, plan M26) ever grows a process boundary with retryable
transports (HTTP endpoints, external tool invocations), that binary — not this
library — is where `HandleError`/`ExitCode` belong.

## Why (not) adopted / flip trigger

- **Adoption would violate the single-dep policy** (AGENTS.md layering
  decision, 2026-09-17): the SDK depends only on `go-finding`, and
  go-error-family already rides in transitively. A direct import would add
  surface without a consumer.
- **Flip trigger:** the `cmd/` CLI graduates from ROADMAP to TODO with a
  process boundary that needs family-based exit codes or retryable error
  handling — revisit at that point, adopting at the binary layer.

## How to verify

```bash
cd /home/lars/projects/go-linter-sdk
# zero direct references (exit 1 expected):
rg -l 'go-error-family' *.go examples/
# transitive-only edge:
go mod why github.com/larsartmann/go-error-family | head -4
# inventory source (read-only):
sed -n '1,40p' "$(go env GOMODCACHE)/github.com/larsartmann/go-error-family@v0.12.0/doc.go"
rg '^func [A-Z]' "$(go env GOMODCACHE)/github.com/larsartmann/go-error-family@v0.12.0" --glob '!*_test.go' --no-filename -o | sort -u | wc -l
```

— Review written 2026-11-19 by the orchestrating session. Point-in-time
snapshot; re-review trigger: the cmd/ CLI decision (M26) or the next
go-error-family minor adopted by go-finding.
