section "eslint"
FIXTURES="$SCRIPT_DIR/fixtures/eslint"
if has_tool eslint; then
    mkdir -p /tmp/t-eslint && cd /tmp/t-eslint && no_config
    cp "$FIXTURES/eslint.config.mjs" .
    cp "$FIXTURES/web.js" .
    cp "$FIXTURES/api.js" .
    cp "$FIXTURES/clean.js" .

    OUT=$(eslint --format=json web.js api.js 2>/dev/null | prettyout-eslint || true)
    check "errors: shows rule"        "$OUT" "no-unused-vars"
    check "errors: shows issue count" "$OUT" "issue"
    check "errors: shows summary"     "$OUT" "issue"
    check "errors: summary separator" "$OUT" " · "
    check "errors: rule count format" "$OUT" " ("

    OUT=$(eslint --format=json clean.js 2>/dev/null | prettyout-eslint || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    with_config eslint group_by file
    OUT=$(eslint --format=json web.js api.js 2>/dev/null | prettyout-eslint || true)
    check "group_by:file: shows filename" "$OUT" "web.js"
    no_config

    with_config eslint colors false
    OUT=$(eslint --format=json web.js api.js 2>/dev/null | prettyout-eslint || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    check "severity prefix present" "$OUT" "[ERROR]"
    no_config
else
    skip "eslint"
fi
