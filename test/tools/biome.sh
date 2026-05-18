section "biome"
FIXTURES="$SCRIPT_DIR/fixtures/biome"
if has_tool biome; then
    mkdir -p /tmp/t-biome && cd /tmp/t-biome && no_config
    cp "$FIXTURES/biome.json" .
    cp "$FIXTURES/web.ts" .
    cp "$FIXTURES/api.ts" .
    cp "$FIXTURES/clean.ts" .

    OUT=$(biome check --reporter=json web.ts api.ts 2>/dev/null | prettyout-biome || true)
    check "errors: shows category"     "$OUT" "issue"
    check "errors: rule count format"  "$OUT" " ("
    check "errors: summary separator"  "$OUT" " · "

    OUT=$(biome check --reporter=json clean.ts 2>/dev/null | prettyout-biome || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    with_config biome group_by file
    OUT=$(biome check --reporter=json web.ts api.ts 2>/dev/null | prettyout-biome || true)
    check "group_by:file: shows filename" "$OUT" "web.ts"
    no_config

    with_config biome colors false
    OUT=$(biome check --reporter=json web.ts api.ts 2>/dev/null | prettyout-biome || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config

    with_config biome colors false
    OUT=$(biome check --reporter=json web.ts api.ts 2>/dev/null | prettyout-biome || true)
    check "severity prefix present" "$OUT" "[ERROR]"
    no_config
else
    skip "biome"
fi
