section "stylelint"
FIXTURES="$SCRIPT_DIR/fixtures/stylelint"
if has_tool stylelint; then
    mkdir -p /tmp/t-stylelint && cd /tmp/t-stylelint && no_config
    cp "$FIXTURES/.stylelintrc.json" .
    cp "$FIXTURES/errors.css" .
    cp "$FIXTURES/clean.css" .

    # stylelint writes JSON to stderr; redirect stderr to stdin of plugin
    OUT=$(stylelint --formatter=json errors.css 2>&1 >/dev/null | prettyout-stylelint || true)
    check "errors: shows rule"        "$OUT" "declaration-property-value-no-unknown"
    check "errors: shows file"        "$OUT" "errors.css"
    check "errors: shows issue count" "$OUT" "issue"

    OUT=$(stylelint --formatter=json clean.css 2>&1 >/dev/null | prettyout-stylelint || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    with_config stylelint group_by file
    OUT=$(stylelint --formatter=json errors.css 2>&1 >/dev/null | prettyout-stylelint || true)
    check "group_by:file: shows filename" "$OUT" "errors.css"
    no_config

    with_config stylelint colors false
    OUT=$(stylelint --formatter=json errors.css 2>&1 >/dev/null | prettyout-stylelint | cat || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config

    with_config stylelint colors false
    OUT=$(stylelint --formatter=json errors.css 2>&1 >/dev/null | prettyout-stylelint || true)
    check "severity prefix present" "$OUT" "[ERROR]"
    no_config
else
    skip "stylelint"
fi
