#!/usr/bin/env bash
# fleet-sweep.sh — sweep the 6 consumer repos for SDK-channel adoption.
# 3 channels (DetectorV2/AddDetectResult/CoverageReporter, Relativize,
# WithEdits/TextEdit) × 6 consumer module roots, plus each module's SDK pin.
# Companion to docs/review/libs/FLEET-SWEEP-BASELINE.md (BEFORE) — run after
# each release and diff against the baseline / previous AFTER.
set -uo pipefail

CONSUMERS=(
	"$HOME/projects/go-structure-linter"
	"$HOME/projects/samber-linter"
	"$HOME/projects/go-humanize-linter"
	"$HOME/projects/cmdguard/lint"
	"$HOME/projects/project-meta/linter"
	"$HOME/projects/BuildFlow/execution"
	"$HOME/projects/BuildFlow/tools"
)

CHANNELS=('DetectorV2|AddDetectResult|CoverageReporter' 'Relativize' 'WithEdits|TextEdit')

for mod in "${CONSUMERS[@]}"; do
	echo "== $mod"
	pin=$(grep -h 'go-linter-sdk v' "$mod/go.mod" 2>/dev/null | head -1 | awk '{print $2}')
	echo "   pin: ${pin:-NONE}"
	for ch in "${CHANNELS[@]}"; do
		count=$(cd "$mod" 2>/dev/null && rg -c "$ch" --glob '*.go' . 2>/dev/null | awk -F: '{s+=$2} END {print s+0}')
		echo "   $ch: ${count:-0} refs"
	done
done
