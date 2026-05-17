section "eslint"
FIXTURES="$SCRIPT_DIR/fixtures/eslint"
if has_tool eslint; then
    mkdir -p /tmp/t-eslint && cd /tmp/t-eslint && no_config
    cp "$FIXTURES/eslint.config.mjs" .
    cp "$FIXTURES/errors.js" .
    cp "$FIXTURES/clean.js" .

    OUT=$(eslint --format=json errors.js 2>/dev/null | prettyout-eslint || true)
    check "errors: shows rule"  "$OUT" "issue"
    check "errors: shows file"  "$OUT" "errors.js"

    OUT=$(eslint --format=json clean.js 2>/dev/null | prettyout-eslint || true)
    check "clean: 0 issues" "$OUT" "0 issues"
else
    skip "eslint"
fi
