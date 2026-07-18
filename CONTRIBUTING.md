# Contributing

Thanks for your interest in contributing to `go-linter-sdk`!

## Prerequisites

This project is part of the LarsArtmann Go ecosystem and follows the same
toolchain conventions as its sibling [`go-finding`](https://github.com/larsartmann/go-finding).

**You need [Nix](https://nixos.org/) with flakes enabled.** All build, test,
and lint commands run inside the Nix dev shell so that the experimental
`GOEXPERIMENT=jsonv2` flag is set consistently — `go-finding` (a transitive
dependency) uses `encoding/json/v2`, which fails to compile without it.

You also need the sibling repo checked out next to this one, because `go.mod`
uses a local `replace` directive until `go-finding` has a tagged release:

```
~/projects/
├── go-finding/      # github.com/larsartmann/go-finding
└── go-linter-sdk/   # this repo
```

## Development Setup

Enter the dev shell — this sets `GOEXPERIMENT=jsonv2` and makes the `reports/`
directory buildflow expects:

```sh
nix develop
```

Then use the flake apps for every common task:

| Command              | What it does                                  |
| -------------------- | --------------------------------------------- |
| `nix run .#test`     | Run all tests (`go test ./... -count=1`)      |
| `nix run .#test-race`| Run all tests with the race detector          |
| `nix run .#bench`    | Run benchmarks                                |
| `nix run .#build`    | Build all packages                            |
| `nix run .#vet`      | Run `go vet`                                  |
| `nix run .#lint`     | Run `golangci-lint`                           |
| `nix run .#coverage` | Run tests with coverage and print a summary   |
| `nix flake check`    | Validate the flake and run treefmt checks     |

To format code (gofumpt + goimports + golines@120 + nixfmt):

```sh
nix fmt
```

To run the full BuildFlow suite (mirrors CI):

```sh
nix develop --command buildflow
```

## Why `GOEXPERIMENT=jsonv2` is mandatory

`go-finding` imports `encoding/json/v2` and `encoding/json/jsontext`, which the
Go toolchain gates behind the `jsonv2` experiment. Without the flag, every
build, test, and lint in this repo fails with
"build constraints exclude all Go files in encoding/json/jsontext". The flake
sets it for you — never run plain `go build`/`go test` outside the dev shell.

## How to Contribute

1. Fork the repository and create a feature branch.
2. Make your changes in `nix develop`, keeping code formatted with `nix fmt`.
3. Ensure `nix run .#test`, `nix run .#lint`, and `nix flake check` all pass.
4. Update `CHANGELOG.md` under the `[Unreleased]` section.
5. Submit a pull request describing the change and its motivation.

## Reporting Issues

Please use [GitHub Issues](https://github.com/larsartmann/go-linter-sdk/issues)
to report bugs or request features.
