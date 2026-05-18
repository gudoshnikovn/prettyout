section "pylint"
FIXTURES="$SCRIPT_DIR/fixtures/pylint"
if has_tool pylint; then
    mkdir -p /tmp/t-pylint && cd /tmp/t-pylint && no_config
    cp "$FIXTURES/models.py" .
    cp "$FIXTURES/views.py" .
    cp "$FIXTURES/clean.py" .

    OUT=$(pylint --output-format=json2 models.py views.py 2>/dev/null | prettyout-pylint || true)
    check "errors: shows message-id"      "$OUT" "/"
    check "errors: shows file"            "$OUT" "models.py"
    check "errors: rule count format"     "$OUT" " ("
    check "errors: summary separator"     "$OUT" " · "
    check "errors: shows severity prefix" "$OUT" "[ERROR]"
    check "errors: shows rating"          "$OUT" "rated"
    check "errors: rating format"         "$OUT" "/10"

    OUT=$(pylint --output-format=json2 clean.py 2>/dev/null | prettyout-pylint || true)
    check "clean: no crash" "$OUT" "issue"

    with_config pylint group_by file
    OUT=$(pylint --output-format=json2 models.py views.py 2>/dev/null | prettyout-pylint || true)
    check "group_by:file: shows filename" "$OUT" "models.py"
    no_config

    with_config pylint colors false
    OUT=$(pylint --output-format=json2 models.py views.py 2>/dev/null | prettyout-pylint || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config
else
    skip "pylint"
fi
