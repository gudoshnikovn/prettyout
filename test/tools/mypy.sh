section "mypy"
FIXTURES="$SCRIPT_DIR/fixtures/mypy"
if has_tool mypy; then
    mkdir -p /tmp/t-mypy && cd /tmp/t-mypy && no_config
    cp "$FIXTURES/models.py" .
    cp "$FIXTURES/views.py" .
    cp "$FIXTURES/clean.py" .

    OUT=$(mypy --output=json models.py views.py 2>/dev/null | prettyout-mypy || true)
    check "errors: shows rule code"    "$OUT" "assignment"
    check "errors: shows file"         "$OUT" "models.py"
    check "errors: shows issue count"  "$OUT" "issue"
    check "errors: shows summary"      "$OUT" "rules"
    check "errors: rule count format"  "$OUT" " ("
    check "errors: summary separator"  "$OUT" " · "

    OUT=$(mypy --output=json clean.py 2>/dev/null | prettyout-mypy || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    with_config mypy group_by file
    OUT=$(mypy --output=json models.py views.py 2>/dev/null | prettyout-mypy || true)
    check "group_by:file: shows filename" "$OUT" "models.py"
    no_config

    with_config mypy group_by rule
    OUT=$(mypy --output=json models.py views.py 2>/dev/null | prettyout-mypy || true)
    check "group_by:rule: collapses lines" "$OUT" "lines 1, 3"
    no_config

    with_config mypy colors false
    OUT=$(mypy --output=json models.py views.py 2>/dev/null | prettyout-mypy || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config

    with_config mypy max_message_length 15
    OUT=$(mypy --output=json models.py views.py 2>/dev/null | prettyout-mypy || true)
    check "max_message_length: message truncated" "$OUT" "..."
    no_config

    with_config mypy colors false
    OUT=$(mypy --output=json models.py views.py 2>/dev/null | prettyout-mypy || true)
    check "severity prefix present" "$OUT" "[ERROR]"
    no_config
else
    skip "mypy"
fi
