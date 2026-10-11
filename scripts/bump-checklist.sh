#!/usr/bin/env bash
# bump-checklist.sh — enforce the dependency-bump checklist.
#
# Encodes the checklist that kept getting skipped (the go-finding v1.10 → v1.14
# window landed with no benchmark re-run and stale docs). Run this after ANY
# `go get` of a dependency or Go toolchain change, BEFORE pushing:
#
#   scripts/bump-checklist.sh
#
# Exits non-zero if any gate fails. CHANGELOG and tag steps are interactive
# reminders, not automatable checks.

set -uo pipefail

repo="$(git rev-parse --show-toplevel)"
cd "$repo"

fail=0
pass() { echo "  OK  $1"; }
miss() { echo "  FAIL $1"; fail=1; }

echo "Dependency-bump checklist for $(basename "$repo")"
echo

echo "[1/8] Build (GOWORK=off, clean env)"
GOWORK=off GOEXPERIMENT= go build ./... && pass "go build ./..." || miss "go build ./..."

echo "[2/8] Vet"
GOWORK=off GOEXPERIMENT= go vet ./... && pass "go vet ./..." || miss "go vet ./..."

echo "[3/8] Tests"
GOWORK=off GOEXPERIMENT= go test ./... -count=1 > /tmp/bump-test.log 2>&1 && pass "go test ./..." || { miss "go test ./... (see /tmp/bump-test.log)"; }

echo "[4/8] Race"
GOWORK=off GOEXPERIMENT= go test ./... -race -count=1 > /tmp/bump-race.log 2>&1 && pass "go test -race" || { miss "go test -race (see /tmp/bump-race.log)"; }

echo "[5/8] Lint"
if command -v golangci-lint > /dev/null; then
	GOWORK=off golangci-lint run --timeout=5m ./... && pass "golangci-lint" || miss "golangci-lint"
else
	miss "golangci-lint not on PATH (nix develop first)"
fi

echo "[6/8] Tidy drift (NEVER tidy with go.work active)"
cp go.mod /tmp/bump-go.mod.bak
cp go.sum /tmp/bump-go.sum.bak
GOWORK=off go mod tidy
if git diff --quiet go.mod go.sum; then
	pass "go.mod/go.sum are tidy"
else
	miss "go.mod/go.sum not tidy — commit the tidy diff (restored originals; run the tidy yourself and commit)"
	cp /tmp/bump-go.mod.bak go.mod
	cp /tmp/bump-go.sum.bak go.sum
fi

echo "[7/8] Doc drift"
scripts/check-doc-versions.sh && pass "check-doc-versions" || miss "check-doc-versions"

echo "[8/8] Benchmarks (manual comparison)"
echo "  RUN: GOWORK=off go test -bench . -benchmem -benchtime=1000x -run '^$' ."
echo "  DIFF against reports/benchmarks-baseline.txt; update the baseline if"
echo "  improved, investigate if regressed >2x. (Not asserted here — hardware.)"

echo
if [ "$fail" -ne 0 ]; then
	echo "CHECKLIST FAILED — fix the FAIL lines above before pushing." >&2
	exit 1
fi

echo "Mechanical gates green. Remaining manual steps:"
echo "  1. CHANGELOG.md: add/update the [Unreleased] entry for this bump."
echo "  2. Commit the go.mod/go.sum/docs diff together."
echo "  3. Tag ONLY after CI is green on the exact commit (never before the"
echo "     run finishes — see the v0.3.0 incident in AGENTS.md)."
echo "  4. After tagging: verify pkg.go.dev renders @latest."
