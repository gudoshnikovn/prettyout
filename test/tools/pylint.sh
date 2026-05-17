section "pylint"
FIXTURES="$SCRIPT_DIR/fixtures/pylint"
if has_tool pylint; then
    mkdir -p /tmp/t-pylint && cd /tmp/t-pylint && no_config
    cp "$FIXTURES/errors.py" .
    cp "$FIXTURES/clean.py" .

    OUT=$(pylint --output-format=json errors.py 2>/dev/null | prettyout-pylint || true)
    check "errors: shows message-id" "$OUT" "/"
    check "errors: shows file"       "$OUT" "errors.py"
    check "errors: rule count format" "$OUT" " ("
    check "errors: summary separator" "$OUT" " · "

    # pylint on minimal clean file still warns — just check no crash
    OUT=$(pylint --output-format=json clean.py 2>/dev/null | prettyout-pylint || true)
    check "clean: no crash" "$OUT" "issue"

    with_config pylint group_by file
    OUT=$(pylint --output-format=json errors.py 2>/dev/null | prettyout-pylint || true)
    check "group_by:file: shows filename" "$OUT" "errors.py"
    no_config

    with_config pylint colors false
    OUT=$(pylint --output-format=json errors.py 2>/dev/null | prettyout-pylint || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config
else
    skip "pylint"
fi
