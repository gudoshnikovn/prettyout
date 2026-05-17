#!/usr/bin/env bash
# prettyout plugin integration test runner
# Usage: ./test/run.sh [tool ...] — omit args to run all tools
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/helpers.sh"

FILTER=("$@")

if [ "${#FILTER[@]}" -gt 0 ]; then
    printf '\033[2mRunning: %s\033[0m\n' "${FILTER[*]}"
fi

for tool_file in "$SCRIPT_DIR/tools/"*.sh; do
    tool_name=$(basename "$tool_file" .sh)
    if run_section "$tool_name"; then
        # shellcheck source=/dev/null
        source "$tool_file"
    fi
done

printf '\n\033[1m══ Results ══\033[0m\n'
printf '\033[1;32mPASS: %d\033[0m  ' "$PASS"
printf '\033[1;31mFAIL: %d\033[0m  ' "$FAIL"
printf '\033[2mSKIP: %d\033[0m\n'   "$SKIP"
[ "$FAIL" -eq 0 ]
