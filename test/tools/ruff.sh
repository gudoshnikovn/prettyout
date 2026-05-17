section "ruff"
FIXTURES="$SCRIPT_DIR/fixtures/ruff"
if has_tool ruff; then
    mkdir -p /tmp/t-ruff && cd /tmp/t-ruff && no_config
    cp "$FIXTURES/errors.py" .
    cp "$FIXTURES/clean.py" .
    cp "$FIXTURES/empty.py" .

    OUT=$(ruff check --output-format=json errors.py clean.py empty.py 2>/dev/null | prettyout-ruff || true)
    check "errors: shows rule code"   "$OUT" "F401"
    check "errors: shows issue count" "$OUT" "issue"
    check "errors: shows files count" "$OUT" "file"

    OUT=$(ruff check --output-format=json clean.py empty.py 2>/dev/null | prettyout-ruff || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    with_config ruff group_by file
    OUT=$(ruff check --output-format=json errors.py 2>/dev/null | prettyout-ruff || true)
    check "group_by:file: shows filename" "$OUT" "errors.py"

    with_config ruff colors false
    OUT=$(ruff check --output-format=json errors.py 2>/dev/null | prettyout-ruff || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    check "fixable hint present" "$OUT" "fixable with --fix"

    with_config ruff max_message_length 10
    OUT=$(ruff check --output-format=json errors.py 2>/dev/null | prettyout-ruff || true)
    check "max_message_length: message truncated" "$OUT" "..."

    no_config
else
    skip "ruff"
fi
