# Note: use --disable-all --enable=ineffassign to avoid the buildir/goanalysis_metalinter
# crash (exit 3, empty stdout) when golangci-lint's Go version mismatches the system Go.
section "golangci-lint"
FIXTURES="$SCRIPT_DIR/fixtures/golangci-lint"
if has_tool golangci-lint; then
    rm -rf /tmp/t-golangci-errors && cp -r "$FIXTURES/errors" /tmp/t-golangci-errors
    cd /tmp/t-golangci-errors && no_config

    OUT=$(golangci-lint run --out-format=json --disable-all --enable=ineffassign 2>/dev/null | prettyout-golangci || true)
    check "errors: shows linter name"   "$OUT" "ineffassign"
    check "errors: shows issue count"   "$OUT" "issue"
    check "errors: shows summary"       "$OUT" "rule"
    check "errors: summary separator"   "$OUT" " · "

    rm -rf /tmp/t-golangci-clean && cp -r "$FIXTURES/clean" /tmp/t-golangci-clean
    cd /tmp/t-golangci-clean && no_config
    OUT=$(golangci-lint run --out-format=json --disable-all --enable=ineffassign 2>/dev/null | prettyout-golangci || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    cd /tmp/t-golangci-errors
    with_config golangci-lint group_by file
    OUT=$(golangci-lint run --out-format=json --disable-all --enable=ineffassign 2>/dev/null | prettyout-golangci || true)
    check "group_by:file: shows filename" "$OUT" "main.go"
    no_config

    with_config golangci-lint colors false
    OUT=$(golangci-lint run --out-format=json --disable-all --enable=ineffassign 2>/dev/null | prettyout-golangci | cat || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config
else
    skip "golangci-lint"
fi
