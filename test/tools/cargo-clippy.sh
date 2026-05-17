section "cargo clippy"
FIXTURES_CLIPPY="$SCRIPT_DIR/fixtures/cargo-clippy"
if has_tool cargo; then
    rm -rf /tmp/t-cargo-errors && cp -r "$FIXTURES_CLIPPY/errors" /tmp/t-cargo-errors
    cd /tmp/t-cargo-errors && no_config

    OUT=$(cargo clippy --message-format=json 2>/dev/null | prettyout-cargo-clippy || true)
    check "errors: shows rule"          "$OUT" "clippy::"
    check "errors: shows issue count"   "$OUT" "issue"
    check "errors: summary separator"   "$OUT" " · "
    check "errors: rule count format"   "$OUT" " ("
    check "errors: primary span line"   "$OUT" "line 6"
    check "errors: artifact filtering (3 issues)" "$OUT" "3 issues"

    cd /tmp/t-cargo-errors && with_config cargo-clippy group_by file
    OUT=$(cargo clippy --message-format=json 2>/dev/null | prettyout-cargo-clippy || true)
    check "group_by:file: shows filename"  "$OUT" "src/main.rs"
    no_config

    cd /tmp/t-cargo-errors && with_config cargo-clippy colors false
    OUT=$(cargo clippy --message-format=json 2>/dev/null | prettyout-cargo-clippy || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config

    rm -rf /tmp/t-cargo-clean && cp -r "$FIXTURES_CLIPPY/clean" /tmp/t-cargo-clean
    cd /tmp/t-cargo-clean && no_config
    OUT=$(cargo clippy --message-format=json 2>/dev/null | prettyout-cargo-clippy || true)
    check "clean: 0 issues" "$OUT" "0 issues · 0 rules · 0 files"

    rm -rf /tmp/t-cargo-rustc && cp -r "$FIXTURES_CLIPPY/rustc-warning" /tmp/t-cargo-rustc
    cd /tmp/t-cargo-rustc && no_config
    OUT=$(cargo clippy --message-format=json 2>/dev/null | prettyout-cargo-clippy || true)
    check "rustc code: shown without clippy:: prefix" "$OUT" "unused_variables"
    check_absent "rustc code: no clippy:: prefix" "$OUT" "clippy::unused_variables"
    check "rustc code: issue count" "$OUT" "1 issue"
else
    skip "cargo"
fi
