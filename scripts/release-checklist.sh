#!/usr/bin/env bash
# release-checklist.sh — enforce the release-time checks that kept getting
# skipped (FEATURES.md missed an entire release; a stale 98.7% coverage claim
# froze into a tag; dead test helpers shipped; vision verdicts were issued on
# pre-fix PNGs).
#
# Run BEFORE `git push` of a release commit:
#
#   scripts/release-checklist.sh
#
# Companion to scripts/bump-checklist.sh (which owns dependency bumps); this
# script owns the release-truth steps. Interactive steps print at the end.

set -uo pipefail

repo="$(git rev-parse --show-toplevel)"
cd "$repo" || exit 1

fail=0
pass() { echo "  OK  $1"; }
miss() {
	echo "  FAIL $1"
	fail=1
}
warn() { echo "  WARN $1"; }

echo "Release checklist for $(basename "$repo")"
echo

echo "[1/7] FEATURES.md covers the release's new exported symbols"
symbols=$(awk '/^## \[/{n++} n==1 && /^### Added/{f=1; next} n==1 && /^### /{f=0} f' CHANGELOG.md |
	grep -E '^- \*\*`' |
	sed -E 's/^- \*\*`([A-Za-z]+).*/\1/' |
	grep -E '^[A-Z]' |
	sort -u || true)
missing_features=""
if [ -z "$symbols" ]; then
	warn "no Added symbols parsed from the newest CHANGELOG entry (check manually)"
else
	for sym in $symbols; do
		if ! grep -q "$sym" FEATURES.md; then
			missing_features="$missing_features $sym"
		fi
	done
	if [ -z "$missing_features" ]; then
		pass "every CHANGELOG symbol appears in FEATURES.md"
	else
		miss "FEATURES.md missing:$missing_features"
	fi
fi

echo "[2/7] Coverage claim in README matches measured library coverage"
measured=$(go list ./... | grep -v /examples/ | tr '\n' ' ' |
	xargs go test -count=1 -coverprofile=/tmp/release-cov.out >/dev/null 2>&1 &&
	go tool cover -func=/tmp/release-cov.out | tail -1 |
	grep -oE '[0-9]+\.[0-9]%' | head -1)
claimed=$(grep -oE '[0-9]+(\.[0-9]+)?% library coverage' README.md | grep -oE '[0-9]+(\.[0-9]+)?%' | head -1)
if [ -z "$measured" ]; then
	miss "could not measure library coverage"
elif [ -z "$claimed" ]; then
	warn "no 'X.X% library coverage' claim found in README (nothing to drift)"
else
	if awk -v m="$measured" -v c="$claimed" 'BEGIN {
	if (m + 0 >= c + 0) exit 0; else exit 1
}'; then
		pass "README claims ${claimed}%, measured ${measured}%"
	else
		miss "README claims ${claimed}% but measured coverage is ${measured}% — update the claim"
	fi
fi

echo "[3/7] Dead code sweep over _test.go helpers (root + examples)"
list_test_files() {
	find . -name '*_test.go' -not -path './vendor/*' -print0 2>/dev/null
}
dead=$(list_test_files | xargs -0 grep -hoE '^(func|type) [a-z][A-Za-z0-9_]*' 2>/dev/null |
	awk '{print $2}' | sort -u | while IFS= read -r sym; do
	uses=$(list_test_files | xargs -0 grep -hcw "$sym" 2>/dev/null | awk '{s+=$1} END {print s+0}')
	if [ "$uses" -le 1 ]; then
		echo "$sym"
	fi
done)
if [ -z "$dead" ]; then
	pass "no single-mention test helpers/types (ghost scaffolding)"
else
	miss "unused test symbols:$(
		echo "$dead" | tr '\n' ' '
	)"
fi

echo "[4/7] Review-corpus render gate (sha256-bound vision verdict)"
html=docs/review/libs/index.html
if [ -f "$html" ]; then
	hash_line=$(grep -o 'render-gate: sha256 .*' "$html" | head -1)
	if [ -n "$hash_line" ]; then
		for pair in full:index-render-full.png top:index-render-top.png; do
			key="${pair%%:*}"
			file="${pair#*:}"
			expected=$(printf '%s' "$hash_line" | grep -oE "$key=[0-9a-f]{64}" | cut -d= -f2)
			if [ -z "$expected" ]; then
				warn "render-gate line missing the '$key=' hash"
				continue
			fi
			actual=$(sha256sum "docs/review/libs/assets/$file" | cut -d' ' -f1)
			if [ "$actual" = "$expected" ]; then
				pass "vision verdict matches $key bytes (sha256)"
			else
				miss "$key PNG bytes changed since the vision verdict (sha256 mismatch) — re-shoot and re-verify, then update the render-gate hash line"
			fi
		done
	else
		stale=$(find docs/review/libs/assets -name '*.png' ! -newer "$html" 2>/dev/null | head -3)
		if [ -z "$stale" ]; then
			pass "screenshots newer than the dashboard HTML"
		else
			warn "stale PNGs (older than index.html): re-shoot and vision-verify on the FINAL bytes before release"
		fi
		warn "no render-gate hash line in index.html — add '<!-- render-gate: sha256 full=... top=... -->' so the gate binds verdicts to bytes, not mtimes"
	fi
	if grep -q 'vision-checked' "$html"; then
		pass "vision verdict recorded in the dashboard"
	else
		warn "no vision-verdict line in index.html — record model + artifact paths (F5.3 pattern)"
	fi
else
	echo "  --  no review dashboard (skipped)"
fi

echo "[5/7] Standard gates (build, vet, race)"
if GOWORK=off go build ./...; then pass "go build ./..."; else miss "go build ./..."; fi
if GOWORK=off go vet ./...; then pass "go vet ./..."; else miss "go vet ./..."; fi
if GOWORK=off go test . -race -count=1 >/tmp/release-race.log 2>&1; then
	pass "go test -race"
else
	miss "go test -race (see /tmp/release-race.log)"
fi

echo "[6/7] Doc drift"
if scripts/check-doc-versions.sh; then pass "check-doc-versions"; else miss "check-doc-versions"; fi
if docs-gates run >/tmp/release-docgates.log 2>&1; then
	pass "docs-gates run"
else
	miss "docs-gates run (see /tmp/release-docgates.log)"
fi

echo "[7/7] Flake evaluates (version pins, treefmt, checks graph)"
if nix flake check --no-build >/tmp/release-flake.log 2>&1; then
	pass "nix flake check --no-build"
else
	miss "nix flake check --no-build (see /tmp/release-flake.log)"
fi

head_sha="$(git rev-parse --short HEAD)"

echo
if [ "$fail" -ne 0 ]; then
	echo "RELEASE CHECKLIST FAILED — fix the FAIL lines above before pushing." >&2
	exit 1
fi

echo "Release gates green. Remaining manual steps:"
echo "  1. Cut CHANGELOG: [Unreleased] -> [version] with narrative."
echo "  2. Commit, push, WAIT for CI green on the exact commit."
echo "  3. Tag ONLY on that green run (never before it finishes). The daemon"
echo "     may commit between the green run and the tag: at tag time, verify"
echo "     the green run's SHA is STILL HEAD (now: $head_sha) or pass the"
echo "     green SHA explicitly to git tag."
echo "  4. After tagging: proxy go list -m @version + pkg.go.dev @latest render."
