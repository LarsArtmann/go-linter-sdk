> [!NOTE]
> This filing was drafted by GLM via Crush from an AI-run investigation, not at my request.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[2026-11-19T17:xx +01:00]`

**TL;DR:** `go test ./providers/` in BuildFlow/tools fails intermittently without any code change — observed FAIL on a full-suite run and PASS on the immediate rerun with identical bytes (and the individually-named failing test always passes alone). The ask: run the suite to failure with `-v`, identify which test leaks state (the once-observed candidate is `TestFilterUnmanagedFormatterRecommendations_DropsUnregistered`), and fix the isolation so CI stops flapping.

## Problem

The `providers` test suite flakes at full-suite scale. Observed pattern on 2026-11-19 (Go 1.27.3, module pinned to `go-linter-sdk v0.6.0` — the dependency is unrelated, the failure predates and survives the bump):

| Run | Command | Result |
| --- | --- | --- |
| 1 | `go test ./...` (tools, full) | FAIL `tools/providers` (48.5s) |
| 2 | `go test ./providers/ -run TestFilterUnmanagedFormatterRecommendations_DropsUnregistered` | ok (0.01s) |
| 3 | `go test ./providers/` | FAIL (43.5s) |
| 4 | `go test ./providers/` (immediate rerun, identical bytes) | ok (36.6s) |
| 5-6 | `go test ./providers/ -count=2` | ok, ok (0 FAIL lines, 76s) |

The only named failure captured: `--- FAIL: TestFilterUnmanagedFormatterRecommendations_DropsUnregistered (0.00s)` in run 1's tail. Assertion output was not captured in that pass; runs 3-6 reproduced the suite-level flake (FAIL then PASS on identical bytes) without naming a different test.

## What this is not

- Not the dependency bump: the failure appeared with v0.4.1-pinned and v0.6.0-pinned trees alike.
- Not a broken test: the named test passes deterministically in isolation (`0.00s`), which points at cross-test state leakage (env var, shared cache dir, or golangci-lint binary path pollution between provider tests) rather than an assertion bug.

## Ask

A `-v` full-suite run until the flake lands (roughly 1-in-2 odds per the table above), then pin the leaking state from the failing test's log context. The suite gating CI makes the flap expensive — one red run per green, currently.

💘 Generated with Crush
