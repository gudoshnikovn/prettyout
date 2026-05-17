section "semgrep"
FIXTURES_SEMGREP="$SCRIPT_DIR/fixtures/semgrep"
if has_tool semgrep; then
    rm -rf /tmp/t-semgrep && mkdir -p /tmp/t-semgrep && cd /tmp/t-semgrep && no_config
    cp "$FIXTURES_SEMGREP/errors.py" .
    cp "$FIXTURES_SEMGREP/clean.py" .
    cp "$FIXTURES_SEMGREP/semgrep-rules.yaml" .

    OUT=$(semgrep --config semgrep-rules.yaml --json errors.py 2>/dev/null | prettyout-semgrep || true)
    check "errors: shows severity prefix"  "$OUT" "[ERROR]"
    check "errors: shows rule name"        "$OUT" "shell-injection"
    check "errors: shows line number"      "$OUT" "line 3"
    check "errors: no 'lines N' singular"  "$OUT" "line 3"
    check_absent "errors: no 'lines 3' bug" "$OUT" "lines 3"
    check "errors: shows warn prefix"      "$OUT" "[WARN]"
    check "errors: plural lines for 2 hits" "$OUT" "lines"
    check "errors: shows issue count"      "$OUT" "3 issues"
    check "errors: summary separator"      "$OUT" " · "
    check "errors: shows rules count"      "$OUT" "2 rules"
    check "errors: shows file count"       "$OUT" "1 file"
    check "errors: path is relative"       "$OUT" "errors.py"
    check_absent "errors: no absolute path" "$OUT" "/tmp/t-semgrep/"

    OUT=$(semgrep --config semgrep-rules.yaml --json clean.py 2>/dev/null | prettyout-semgrep || true)
    check "clean: 0 issues" "$OUT" "0 issues · 0 rules · 0 files"

    with_config semgrep group_by file
    OUT=$(semgrep --config semgrep-rules.yaml --json errors.py 2>/dev/null | prettyout-semgrep || true)
    check "group_by:file: shows filename" "$OUT" "errors.py"
    check "group_by:file: shows line"     "$OUT" "line 3"
    no_config

    with_config semgrep colors false
    OUT=$(semgrep --config semgrep-rules.yaml --json errors.py 2>/dev/null | prettyout-semgrep | cat || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    check "colors:false: still shows severity" "$OUT" "[ERROR]"
    no_config

    OUT=$(echo -n '' | prettyout-semgrep 2>&1 || true)
    check "empty stdin: error message" "$OUT" "invalid JSON"

    EXITCODE=0
    echo 'not json' | prettyout-semgrep 2>/dev/null || EXITCODE=$?
    if [ "$EXITCODE" -eq 1 ]; then
        green "  PASS  invalid JSON: exit 1"
        PASS=$((PASS+1))
    else
        red   "  FAIL  invalid JSON: exit 1 (got exit=$EXITCODE)"
        FAIL=$((FAIL+1))
    fi
else
    skip "semgrep"
fi
