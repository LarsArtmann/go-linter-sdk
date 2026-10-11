#!/usr/bin/env bash
# check-doc-versions.sh — doc-drift guard.
#
# Fails when a living doc hardcodes a dependency or Go-toolchain version that
# disagrees with go.mod, or when the toolchain pins in flake.nix /
# .github/workflows/ci.yml / .golangci.yml disagree with go.mod's minor.
# go.mod is the single source of truth; living docs must point at it, not
# duplicate it (this guard exists because prose duplicates rotted six docs in
# one day, see docs/status/2026-10-11_07-11_session-self-review-and-status.md).
#
# Rules
#   1. Any literal semantic version vX.Y.Z in a living doc must be either
#      - the go-finding or go-error-family version required by go.mod, or
#      - a tag of this repository (git tag -l; append-only release history).
#   2. Any Go version literal (go1.M.P, go 1.M, 1.M.P) in a living doc must
#      carry go.mod's minor version.
#   3. Toolchain pins must agree with go.mod's minor:
#      - flake.nix:            pkgs.go_1_M
#      - ci.yml:               GOTOOLCHAIN: go1.M.P and go-version: "1.M"
#      - .golangci.yml:        run.go: 1.M.P
#
# Escape hatch: append the marker `docguard:allow` to a line to exempt it.
# Use sparingly; every use is a promise to revisit that line on the next bump.
#
# Usage: scripts/check-doc-versions.sh   (from anywhere; exits 0 clean, 1 drift)

set -euo pipefail

repo="$(git rev-parse --show-toplevel)"
cd "$repo"

LIVING_DOCS=(
	AGENTS.md
	README.md
	CONTRIBUTING.md
	FEATURES.md
	ROADMAP.md
	TODO_LIST.md
	SUPPORT.md
	SECURITY.md
	docs/DOMAIN_LANGUAGE.md
	docs/migration-guide.md
)

fail=0

die() {
	echo "FAIL: $*" >&2
	fail=1
}

[ -f go.mod ] || die "go.mod not found"

go_minor="$(awk '/^go [0-9]/ {split($2, a, "."); print a[2]; exit}' go.mod)"
go_major="$(awk '/^go [0-9]/ {split($2, a, "."); print a[1]; exit}' go.mod)"
finding_ver="$(awk '$2 == "github.com/larsartmann/go-finding" {print $3; exit}' go.mod)"
errfam_ver="$(awk '$2 == "github.com/larsartmann/go-error-family" {print $3; exit}' go.mod)"

[ -n "$go_minor" ] || die "could not parse 'go' directive from go.mod"
[ -n "$finding_ver" ] || die "go-finding not found in go.mod require block"

allowed_versions() {
	git tag -l
	echo "$finding_ver"
	[ -n "$errfam_ver" ] && echo "$errfam_ver"
	return 0
}

allowed="$(allowed_versions)"

# Rule 1 + 2 over living prose.
for doc in "${LIVING_DOCS[@]}"; do
	[ -f "$doc" ] || continue
	while IFS= read -r line; do
		lnum="${line%%:*}"
		text="${line#*:}"
		case "$text" in *docguard:allow*) continue ;; esac
		for ver in $(printf '%s\n' "$text" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' || true); do
			if ! printf '%s\n' "$allowed" | grep -qxF "$ver"; then
				die "$doc:$lnum hardcodes $ver, which is neither a go.mod dependency ($finding_ver${errfam_ver:+, $errfam_ver}) nor a repo tag. Point at go.mod or update the reference."
			fi
		done
		for golit in $(printf '%s\n' "$text" | grep -oE '(go1?[- ])?[0-9]+\.[0-9]+\.[0-9]+|go 1\.[0-9]+|Go 1\.[0-9]+' || true); do
			case "$golit" in
			go1* | *go\ 1* | *Go\ 1*)
				minor="$(printf '%s' "$golit" | grep -oE '1\.[0-9]+' | head -1 | cut -d. -f2)"
				[ "$minor" = "$go_minor" ] || die "$doc:$lnum pins Go $golit but go.mod requires go $go_major.$go_minor. Update the reference or move it to a dated status report."
				;;
			esac
		done
	done < <(grep -nE 'v[0-9]+\.[0-9]+\.[0-9]+|go1?[0-9]*\.?[0-9]*\.[0-9]+|[Gg]o 1\.[0-9]+' "$doc" || true)
done

# Rule 3: toolchain pins must agree with go.mod's minor.
if grep -qE "go_1_[0-9]+" flake.nix; then
	pin="$(grep -oE 'go_1_[0-9]+' flake.nix | head -1 | cut -d_ -f3)"
	[ "$pin" = "$go_minor" ] || die "flake.nix pins go_1_$pin but go.mod requires go $go_major.$go_minor. Keep the three pins in lockstep (flake.nix, ci.yml, .golangci.yml)."
fi
if [ -f .github/workflows/ci.yml ]; then
	while IFS=: read -r _ goline; do
		minor="$(printf '%s' "$goline" | grep -oE 'go1\.[0-9]+' | cut -d. -f2)"
		[ "$minor" = "$go_minor" ] || die ".github/workflows/ci.yml GOTOOLCHAIN '$goline' disagrees with go.mod go $go_major.$go_minor."
	done < <(grep -nE 'GOTOOLCHAIN: *go1\.' .github/workflows/ci.yml || true)
	while IFS=: read -r _ goline; do
		minor="$(printf '%s' "$goline" | grep -oE '"1\.[0-9]+"' | tr -d '"' | cut -d. -f2)"
		[ "$minor" = "$go_minor" ] || die ".github/workflows/ci.yml go-version '$goline' disagrees with go.mod go $go_major.$go_minor."
	done < <(grep -nE 'go-version: *"1\.[0-9]+" *$' .github/workflows/ci.yml || true)
fi
if grep -qE '^  go: 1\.[0-9]+' .golangci.yml 2>/dev/null; then
	pin="$(grep -E '^  go: 1\.[0-9]+' .golangci.yml | grep -oE '1\.[0-9]+' | cut -d. -f2)"
	[ "$pin" = "$go_minor" ] || die ".golangci.yml run.go pins 1.$pin but go.mod requires go $go_major.$go_minor."
fi

if [ "$fail" -ne 0 ]; then
	echo
	echo "Doc drift detected. go.mod is the source of truth (go $go_major.$go_minor, $finding_ver${errfam_ver:+, $errfam_ver})." >&2
	echo "Fix: replace hardcoded versions in living docs with a pointer to go.mod, or correct the stale pin. Append 'docguard:allow' only for genuinely historical lines." >&2
	exit 1
fi
echo "doc versions OK (go $go_major.$go_minor, $finding_ver${errfam_ver:+, $errfam_ver})"
