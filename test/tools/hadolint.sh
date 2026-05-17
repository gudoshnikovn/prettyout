section "hadolint"
FIXTURES="$SCRIPT_DIR/fixtures/hadolint"
if has_tool hadolint; then
    mkdir -p /tmp/t-hadolint && cd /tmp/t-hadolint && no_config
    cp "$FIXTURES/Dockerfile.errors" .
    cp "$FIXTURES/Dockerfile.clean" .

    OUT=$(hadolint --format=json Dockerfile.errors 2>/dev/null | prettyout-hadolint || true)
    check "errors: shows DL code"      "$OUT" "DL"
    check "errors: shows file"         "$OUT" "Dockerfile"
    check "errors: rule count format"  "$OUT" " ("
    check "errors: summary separator"  "$OUT" " · "

    OUT=$(hadolint --format=json Dockerfile.clean 2>/dev/null | prettyout-hadolint || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    with_config hadolint group_by file
    OUT=$(hadolint --format=json Dockerfile.errors 2>/dev/null | prettyout-hadolint || true)
    check "group_by:file: shows filename" "$OUT" "Dockerfile"
    no_config

    with_config hadolint colors false
    OUT=$(hadolint --format=json Dockerfile.errors 2>/dev/null | prettyout-hadolint | cat || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config
else
    skip "hadolint"
fi
