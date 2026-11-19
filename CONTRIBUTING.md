# Contributing

Thanks for your interest in contributing to `go-linter-sdk`!

## Prerequisites

This project is part of the LarsArtmann Go ecosystem and follows the same
toolchain conventions as its sibling [`go-finding`](https://github.com/larsartmann/go-finding).

**You need [Nix](https://nixos.org/) with flakes enabled.** All build, test,
and lint commands run inside the Nix dev shell, which pins the Go 1.27
toolchain the module requires (see `go.mod` for the exact floor).

`go-finding` is resolved from the public module proxy as a published tag (see
`go.mod` for the pinned version — do not duplicate version numbers in prose).
No local `replace` directive and no sibling checkout are needed. For local
cross-repo development against an uncommitted `go-finding`, add a temporary
`go.work` or `replace` line — neither is committed.

## Development Setup

Enter the dev shell — it pins the Go toolchain and makes the `reports/`
directory buildflow expects:

```sh
nix develop
```

Then use the flake apps for every common task:

| Command               | What it does                                |
| --------------------- | ------------------------------------------- |
| `nix run .#test`      | Run all tests (`go test ./... -count=1`)    |
| `nix run .#test-race` | Run all tests with the race detector        |
| `nix run .#bench`     | Run benchmarks                              |
| `nix run .#build`     | Build all packages                          |
| `nix run .#vet`       | Run `go vet`                                |
| `nix run .#lint`      | Run `golangci-lint`                         |
| `nix run .#coverage`  | Run tests with coverage and print a summary |
| `nix flake check`     | Validate the flake and run treefmt checks   |

To format code (gofumpt + goimports + golines@120 + nixfmt):

```sh
nix fmt
```

### Doc-drift guard

`go.mod` is the single source of truth for dependency and Go versions. Living
docs (AGENTS.md, README.md, CONTRIBUTING.md, FEATURES.md, ROADMAP.md,
TODO_LIST.md, SUPPORT.md, SECURITY.md, docs/DOMAIN_LANGUAGE.md,
docs/migration-guide.md) must not hardcode versions that disagree with it.
CI enforces this with the `doc-drift` job; run it locally before pushing:

```sh
scripts/check-doc-versions.sh
```

It fails when a doc pins a dependency version that is not `go.mod`'s (or a
repo tag), or when the Go pins in `flake.nix`, `ci.yml`, and `.golangci.yml`
fall out of lockstep with `go.mod`. For a genuinely historical version
reference, append the marker `docguard:allow` to that line.

To run the full BuildFlow suite (mirrors CI):

```sh
nix develop --command buildflow
```

## How to Contribute

1. Fork the repository and create a feature branch.
2. Make your changes in `nix develop`, keeping code formatted with `nix fmt`.
3. Ensure `nix run .#test`, `nix run .#lint`, `nix flake check`, and
   `scripts/check-doc-versions.sh` all pass.
4. Update `CHANGELOG.md` under the `[Unreleased]` section.
5. Submit a pull request describing the change and its motivation.

## Examples directory

The `examples/` directory contains standalone CLI programs that demonstrate
the SDK's usage. Each example is a `package main` in its own subdirectory
(e.g., `examples/minimal-linter/`).

Conventions:

- Examples are part of the same Go module — they import
  `github.com/larsartmann/go-linter-sdk` directly.
- `.golangci.yml` excludes `examples/` from `depguard`, `forbidigo`, and `mnd`
  because example CLI code legitimately uses `fmt.Println`, exit codes, and
  unrestricted imports.
- Each example should have integration tests in `examples_integration_test.go`
  (root package) that build the binary, run it on clean/dirty temp dirs, and
  assert exit codes.
- Run examples with `go run ./examples/<name> [dir]`.

## Reporting Issues

Please use [GitHub Issues](https://github.com/larsartmann/go-linter-sdk/issues)
to report bugs or request features.

## Docs gates (docs-health battery)

This repo is judged by `docs-gates run` (the docs-health battery: freshness,
annotate, routing, cadence, plus the doc-drift guard). Its repo-specific
knobs live in `.docs-gates.json`; a repo without one gets the fleet defaults.
This repo pins one knob:

```json
{ "audit_max": 0 }
```

`audit_max` is the freshness ratchet: how many unharvested §f items the
newest status report may carry before the gate goes red. `0` means strict —
every harvestable item must be routed into TODO_LIST/ROADMAP the day it is
written. Never raise the pin; the ratchet only ratchets down. (History: it
was 33 while ancient archived reports carried unharvested items, then
drained to 0 on 2026-11-19.) The other knobs (corpus paths, thresholds,
skip_gates) keep fleet defaults — see `docs-gates --help`.

## Version-pinned references

Docs that name dependency or tool versions must point at `go.mod` (the single
source of truth) rather than hardcoding versions — `scripts/check-doc-versions.sh`
(CI job `doc-drift`) enforces this on the living docs. Point-in-time snapshot
reports (status, reviews, plans) may stamp the version they reviewed; the
library-utilization reviews additionally record a "version reviewed" line and a
re-review trigger — follow that convention when adding a report under
`docs/review/libs/`.
