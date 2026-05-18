section "basedpyright"
FIXTURES="$SCRIPT_DIR/fixtures/basedpyright"
if has_tool basedpyright; then
    mkdir -p /tmp/t-pyright && cd /tmp/t-pyright && no_config
    cp "$FIXTURES/errors.py" .
    cp "$FIXTURES/clean.py" .
    cp "$FIXTURES/syntax_error.py" .
    cp "$FIXTURES/models.py" .
    cp "$FIXTURES/views.py" .

    OUT=$(basedpyright --outputjson errors.py 2>/dev/null | prettyout-basedpyright || true)
    check "errors: shows severity"     "$OUT" "[ERROR]"
    check "errors: shows rule code"    "$OUT" "report"
    check "errors: shows issue count"  "$OUT" "issue"

    OUT=$(basedpyright --outputjson clean.py 2>/dev/null | prettyout-basedpyright || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    OUT=$(basedpyright --outputjson syntax_error.py 2>/dev/null | prettyout-basedpyright || true)
    check_absent "syntax error: no dup line numbers" "$OUT" "line 1, 1"

    OUT=$(basedpyright --outputjson models.py views.py 2>/dev/null | prettyout-basedpyright || true)
    check "multi-file: shows filename"  "$OUT" "models.py"
    check "multi-file: shows severity"  "$OUT" "[ERROR]"
    check "multi-file: shows rule code" "$OUT" "report"

    with_config basedpyright group_by file
    OUT=$(basedpyright --outputjson errors.py 2>/dev/null | prettyout-basedpyright || true)
    check "group_by:file: shows filename" "$OUT" "errors.py"

    no_config
else
    skip "basedpyright"
fi
