#!/usr/bin/env bash
# Shared helpers for prettyout integration tests.
# Sourced by run.sh before each tool file.

PASS=0
FAIL=0
SKIP=0

green() { printf '\033[1;32m%s\033[0m\n' "$*"; }
red()   { printf '\033[1;31m%s\033[0m\n' "$*"; }
dim()   { printf '\033[2m%s\033[0m\n'   "$*"; }

check() {
    local label="$1" output="$2" expect="$3"
    if echo "$output" | grep -qF "$expect"; then
        green "  PASS  $label"
        PASS=$((PASS+1))
    else
        red   "  FAIL  $label"
        dim   "        expected to find: $expect"
        dim   "        got: $(echo "$output" | head -3)"
        FAIL=$((FAIL+1))
    fi
}

check_absent() {
    local label="$1" output="$2" absent="$3"
    if echo "$output" | grep -qF "$absent"; then
        red   "  FAIL  $label"
        dim   "        expected NOT to find: $absent"
        FAIL=$((FAIL+1))
    else
        green "  PASS  $label"
        PASS=$((PASS+1))
    fi
}

has_tool() { command -v "$1" &>/dev/null; }

skip() {
    printf '\033[2m  SKIP  %s (not installed)\033[0m\n' "$*"
    SKIP=$((SKIP+1))
}

section() { printf '\n\033[1m══ %s ══\033[0m\n' "$*"; }

with_config() {
    local tool="$1" key="$2" val="$3"
    printf 'settings:\n  %s:\n    %s: %s\n' "$tool" "$key" "$val" > .prettyout.yaml
}

no_config() { rm -f .prettyout.yaml; }

# run_section returns 0 if the tool should run.
# FILTER is set by run.sh from its positional args; spaces normalised to hyphens.
run_section() {
    local name="${1// /-}"
    [ "${#FILTER[@]}" -eq 0 ] && return 0
    for f in "${FILTER[@]}"; do
        [ "${f// /-}" = "$name" ] && return 0
    done
    return 1
}
