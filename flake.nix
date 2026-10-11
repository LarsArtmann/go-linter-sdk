{
  description = "go-linter-sdk — Shared scaffolding for LarsArtmann Go linters";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      nixpkgs,
      flake-parts,
      treefmt-nix,
      ...
    }:
    let
      inherit (nixpkgs) lib;
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      # Inline systems: nixpkgs 26.11 dropped x86_64-darwin, which github:nix-systems/default still lists.
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      imports = [
        treefmt-nix.flakeModule
      ];

      perSystem =
        {
          pkgs,
          ...
        }:
        let
          goPkg = pkgs.go_1_27;

          mkApp = name: description: script: {
            type = "app";
            program = "${
              pkgs.writeShellApplication {
                inherit name;
                runtimeInputs = [
                  goPkg
                  pkgs.golangci-lint
                  pkgs.trash-cli
                ];
                text = script;
              }
            }/bin/${name}";
            meta = {
              inherit description;
              mainProgram = name;
              homepage = "https://github.com/larsartmann/go-linter-sdk";
              license = lib.licenses.mit;
              platforms = lib.platforms.unix;
              maintainers = [
                {
                  name = "Lars Artmann";
                  github = "LarsArtmann";
                }
              ];
            };
          };
        in
        {
          treefmt = {
            projectRootFile = "go.mod";
            programs = {
              gofumpt.enable = true;
              # goimports shells out to `go` at runtime; unless it sees a Go at
              # least as new as go.mod's floor, the toolchain auto-switch tries
              # to download a newer Go and fails in the network-less check
              # sandbox. Pin it to the same toolchain as the rest of the flake.
              goimports = {
                enable = true;
                package = pkgs.gotools.override { go = goPkg; };
              };
              golines = {
                enable = true;
                maxLength = 120;
              };
              nixfmt.enable = true;
            };
          };

          devShells.default = pkgs.mkShell {
            packages = [
              goPkg
              pkgs.golangci-lint
              pkgs.gofumpt
              pkgs.golines
              pkgs.gopls
              pkgs.gotools
              pkgs.trash-cli
            ];

            shellHook = ''
              echo "go-linter-sdk dev shell — $(go version)"
            '';
          };

          devShells.ci = pkgs.mkShellNoCC {
            packages = [
              goPkg
              pkgs.golangci-lint
            ];
          };

          apps = {
            test = mkApp "test" "Run all tests" ''
              go test ./... -count=1 "$@"
            '';

            test-race = mkApp "test-race" "Run all tests with race detector" ''
              go test ./... -race -count=1 "$@"
            '';

            bench = mkApp "bench" "Run benchmarks" ''
              go test ./... -bench=. -benchmem "$@"
            '';

            build = mkApp "build" "Build all packages" ''
              go build ./...
            '';

            vet = mkApp "vet" "Run go vet" ''
              go vet ./...
            '';

            lint = mkApp "lint" "Run golangci-lint" ''
              golangci-lint run ./...
            '';

            coverage = mkApp "coverage" "Run tests with coverage report" ''
              mkdir -p reports
              go test ./... -coverprofile=reports/coverage.out -covermode=atomic "$@"
              go tool cover -func=reports/coverage.out
            '';

            clean = mkApp "clean" "Clean build and test artifacts" ''
              # reports/coverage.out is the canonical path; the bare coverage.out
              # sweep removes artifacts from before the path was aligned.
              trash-put reports/coverage.out coverage.out 2>/dev/null || true
              go clean -testcache
            '';
          };
        };
    };
}
