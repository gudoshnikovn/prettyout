section "shellcheck"
FIXTURES="$SCRIPT_DIR/fixtures/shellcheck"
if has_tool shellcheck; then
    mkdir -p /tmp/t-shellcheck && cd /tmp/t-shellcheck && no_config
    cp "$FIXTURES/errors.sh" .
    cp "$FIXTURES/clean.sh" .

    OUT=$(shellcheck --format=json errors.sh 2>/dev/null | prettyout-shellcheck || true)
    check "errors: shows SC code"      "$OUT" "SC"
    check "errors: shows file"         "$OUT" "errors.sh"
    check "errors: rule count format"  "$OUT" " ("
    check "errors: summary separator"  "$OUT" " · "

    OUT=$(shellcheck --format=json clean.sh 2>/dev/null | prettyout-shellcheck || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    with_config shellcheck group_by file
    OUT=$(shellcheck --format=json errors.sh 2>/dev/null | prettyout-shellcheck || true)
    check "group_by:file: shows filename" "$OUT" "errors.sh"
    no_config

    with_config shellcheck colors false
    OUT=$(shellcheck --format=json errors.sh 2>/dev/null | prettyout-shellcheck | cat || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config
else
    skip "shellcheck"
fi
