section "trivy"
FIXTURES_TRIVY="$SCRIPT_DIR/fixtures/trivy"
if has_tool trivy; then
    rm -rf /tmp/t-trivy && cp -r "$FIXTURES_TRIVY/django" /tmp/t-trivy
    cd /tmp/t-trivy && no_config

    OUT=$(trivy fs --format=json --quiet /tmp/t-trivy 2>/dev/null | prettyout-trivy || true)
    check "vuln scan: shows CRITICAL"       "$OUT" "CRITICAL"
    check "vuln scan: shows HIGH"           "$OUT" "HIGH"
    check "vuln scan: shows vuln count"     "$OUT" "vulnerabilities"
    check "vuln scan: shows CVE ID"         "$OUT" "CVE-"
    check "vuln scan: shows fix version"    "$OUT" "fix:"
    check "vuln scan: summary line"         "$OUT" "severity levels"

    CRITICAL_LINE=$(echo "$OUT" | grep -n "CRITICAL" | head -1 | cut -d: -f1)
    HIGH_LINE=$(echo "$OUT" | grep -n "HIGH" | head -1 | cut -d: -f1)
    if [ -n "$CRITICAL_LINE" ] && [ -n "$HIGH_LINE" ] && [ "$CRITICAL_LINE" -lt "$HIGH_LINE" ]; then
        green "  PASS  vuln scan: CRITICAL before HIGH"
        PASS=$((PASS+1))
    else
        red   "  FAIL  vuln scan: CRITICAL before HIGH"
        dim   "        CRITICAL_LINE=$CRITICAL_LINE HIGH_LINE=$HIGH_LINE"
        FAIL=$((FAIL+1))
    fi

    with_config trivy colors false
    OUT=$(trivy fs --format=json --quiet /tmp/t-trivy 2>/dev/null | prettyout-trivy | cat || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config

    mkdir -p /tmp/t-trivy-clean
    OUT=$(trivy fs --format=json --quiet /tmp/t-trivy-clean 2>/dev/null | prettyout-trivy || true)
    check "clean scan: 0 vulns message" "$OUT" "No vulnerabilities found."

    OUT=$(printf '{"Results":[{"Target":"test.txt","Vulnerabilities":null}]}' | prettyout-trivy || true)
    check "null Vulnerabilities: no crash" "$OUT" "No vulnerabilities found."

    OUT=$(printf '{"Results":[{"Target":"req.txt","Vulnerabilities":[{"VulnerabilityID":"CVE-2099-9999","PkgName":"somepkg","InstalledVersion":"1.0","FixedVersion":"","Severity":"HIGH"}]}]}' | prettyout-trivy || true)
    check "no fix available: shown for unfixed CVE" "$OUT" "no fix available"

    OUT=$(echo -n '' | prettyout-trivy 2>&1 || true)
    check "empty stdin: error message" "$OUT" "invalid JSON"

    EXITCODE=0
    echo 'not json' | prettyout-trivy 2>/dev/null || EXITCODE=$?
    if [ "$EXITCODE" -eq 1 ]; then
        green "  PASS  invalid JSON: exit 1"
        PASS=$((PASS+1))
    else
        red   "  FAIL  invalid JSON: exit 1 (got exit=$EXITCODE)"
        FAIL=$((FAIL+1))
    fi

    with_config trivy group_by file
    OUT=$(trivy fs --format=json --quiet /tmp/t-trivy 2>/dev/null | prettyout-trivy || true)
    check "group_by:file: shows target filename" "$OUT" "requirements.txt"
    no_config
else
    skip "trivy"
fi
