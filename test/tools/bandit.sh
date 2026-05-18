section "bandit"
FIXTURES="$SCRIPT_DIR/fixtures/bandit"
if has_tool bandit; then
    mkdir -p /tmp/t-bandit && cd /tmp/t-bandit && no_config
    cp "$FIXTURES/web.py" .
    cp "$FIXTURES/api.py" .
    cp "$FIXTURES/errors.py" .
    cp "$FIXTURES/clean.py" .

    OUT=$(bandit -f json web.py api.py 2>/dev/null | prettyout-bandit || true)
    check "errors: shows test id"       "$OUT" "B"
    check "errors: shows file"          "$OUT" "web.py"
    check "errors: rule count format"   "$OUT" " ("
    check "errors: summary separator"   "$OUT" " · "

    OUT=$(bandit -f json clean.py 2>/dev/null | prettyout-bandit || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    with_config bandit group_by file
    OUT=$(bandit -f json web.py api.py 2>/dev/null | prettyout-bandit || true)
    check "group_by:file: shows filename" "$OUT" "web.py"
    no_config

    with_config bandit colors false
    OUT=$(bandit -f json web.py api.py 2>/dev/null | prettyout-bandit || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config
else
    skip "bandit"
fi
