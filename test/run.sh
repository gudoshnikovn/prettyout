#!/usr/bin/env bash
# prettyout plugin integration test runner
# Runs every tool with real JSON output and pipes through our plugin.
# Tests: errors, clean, empty, syntax error, group_by:file, colors:false.
set -euo pipefail

PASS=0
FAIL=0
SKIP=0

# ── Helpers ──────────────────────────────────────────────────────────────────

green() { printf '\033[1;32m%s\033[0m\n' "$*"; }
red()   { printf '\033[1;31m%s\033[0m\n' "$*"; }
dim()   { printf '\033[2m%s\033[0m\n'   "$*"; }

check() {
    local label="$1" output="$2" expect="$3"
    if echo "$output" | grep -qF "$expect"; then
        green "  PASS  $label"
        PASS=$((PASS+1))
    else
        red   "  FAIL  $label"
        dim   "        expected to find: $expect"
        dim   "        got: $(echo "$output" | head -3)"
        FAIL=$((FAIL+1))
    fi
}

check_absent() {
    local label="$1" output="$2" absent="$3"
    if echo "$output" | grep -qF "$absent"; then
        red   "  FAIL  $label"
        dim   "        expected NOT to find: $absent"
        FAIL=$((FAIL+1))
    else
        green "  PASS  $label"
        PASS=$((PASS+1))
    fi
}

has_tool() { command -v "$1" &>/dev/null; }

skip() {
    printf '\033[2m  SKIP  %s (not installed)\033[0m\n' "$*"
    SKIP=$((SKIP+1))
}

section() { printf '\n\033[1m══ %s ══\033[0m\n' "$*"; }

# Write a .prettyout.yaml in current dir and clean up after test block
with_config() {
    local tool="$1" key="$2" val="$3"
    printf 'settings:\n  %s:\n    %s: %s\n' "$tool" "$key" "$val" > .prettyout.yaml
}

no_config() { rm -f .prettyout.yaml; }

# ── ruff ─────────────────────────────────────────────────────────────────────
section "ruff"
if has_tool ruff; then
    mkdir -p /tmp/t-ruff && cd /tmp/t-ruff && no_config

    cat > errors.py << 'PY'
import os
import sys

def foo():
    try:
        return 1
    except Exception:
        raise ValueError("fail")
PY
    cat > clean.py << 'PY'
def add(a: int, b: int) -> int:
    return a + b
PY
    cat > empty.py << 'PY'
PY

    OUT=$(ruff check --output-format=json errors.py clean.py empty.py 2>/dev/null | prettyout-ruff || true)
    check "errors: shows rule code"   "$OUT" "F401"
    check "errors: shows issue count" "$OUT" "issue"
    check "errors: shows files count" "$OUT" "file"

    OUT=$(ruff check --output-format=json clean.py empty.py 2>/dev/null | prettyout-ruff || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    with_config ruff group_by file
    OUT=$(ruff check --output-format=json errors.py 2>/dev/null | prettyout-ruff || true)
    check "group_by:file: shows filename" "$OUT" "errors.py"

    with_config ruff colors false
    OUT=$(ruff check --output-format=json errors.py 2>/dev/null | prettyout-ruff || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['

    with_config ruff max_message_length 10
    OUT=$(ruff check --output-format=json errors.py 2>/dev/null | prettyout-ruff || true)
    check "max_message_length: message truncated" "$OUT" "..."

    no_config
else
    skip "ruff"
fi

# ── basedpyright ──────────────────────────────────────────────────────────────
section "basedpyright"
if has_tool basedpyright; then
    mkdir -p /tmp/t-pyright && cd /tmp/t-pyright && no_config

    cat > errors.py << 'PY'
x: int = "not an int"
def f() -> int:
    pass
PY
    cat > clean.py << 'PY'
def add(a: int, b: int) -> int:
    return a + b
PY
    cat > syntax_error.py << 'PY'
def broken(:
PY

    OUT=$(basedpyright --outputjson errors.py 2>/dev/null | prettyout-basedpyright || true)
    check "errors: shows severity"     "$OUT" "[ERROR]"
    check "errors: shows rule code"    "$OUT" "report"
    check "errors: shows issue count"  "$OUT" "issue"

    OUT=$(basedpyright --outputjson clean.py 2>/dev/null | prettyout-basedpyright || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    OUT=$(basedpyright --outputjson syntax_error.py 2>/dev/null | prettyout-basedpyright || true)
    check "syntax error: no duplicate lines" "$(echo "$OUT" | grep 'line 1, 1')" "" || true
    check_absent "syntax error: no dup line numbers" "$OUT" "line 1, 1"

    with_config basedpyright group_by file
    OUT=$(basedpyright --outputjson errors.py 2>/dev/null | prettyout-basedpyright || true)
    check "group_by:file: shows filename" "$OUT" "errors.py"

    no_config
else
    skip "basedpyright"
fi

# ── mypy ──────────────────────────────────────────────────────────────────────
section "mypy"
if has_tool mypy; then
    mkdir -p /tmp/t-mypy && cd /tmp/t-mypy && no_config

    cat > errors.py << 'PY'
x: int = "not an int"
result = x + "hello"
y: int = "also bad"
PY
    cat > clean.py << 'PY'
def add(a: int, b: int) -> int:
    return a + b
PY

    OUT=$(mypy --output=json errors.py 2>/dev/null | prettyout-mypy || true)
    check "errors: shows rule code"    "$OUT" "assignment"
    check "errors: shows file"         "$OUT" "errors.py"
    check "errors: shows issue count"  "$OUT" "issue"
    check "errors: shows summary"      "$OUT" "rules"
    check "errors: rule count format"  "$OUT" " ("
    check "errors: summary separator"  "$OUT" " · "

    OUT=$(mypy --output=json clean.py 2>/dev/null | prettyout-mypy || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    with_config mypy group_by file
    OUT=$(mypy --output=json errors.py 2>/dev/null | prettyout-mypy || true)
    check "group_by:file: shows filename" "$OUT" "errors.py"
    no_config

    with_config mypy group_by rule
    OUT=$(mypy --output=json errors.py 2>/dev/null | prettyout-mypy || true)
    check "group_by:rule: collapses lines" "$OUT" "lines 1, 3"
    no_config

    with_config mypy colors false
    OUT=$(mypy --output=json errors.py 2>/dev/null | prettyout-mypy || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config

    with_config mypy max_message_length 15
    OUT=$(mypy --output=json errors.py 2>/dev/null | prettyout-mypy || true)
    check "max_message_length: message truncated" "$OUT" "..."
    no_config
else
    skip "mypy"
fi

# ── bandit ────────────────────────────────────────────────────────────────────
section "bandit"
if has_tool bandit; then
    mkdir -p /tmp/t-bandit && cd /tmp/t-bandit && no_config

    cat > errors.py << 'PY'
import hashlib
import subprocess
hashlib.md5(b"data")
subprocess.call("ls", shell=True)
PY
    cat > clean.py << 'PY'
def add(a: int, b: int) -> int:
    return a + b
PY

    OUT=$(bandit -f json errors.py 2>/dev/null | prettyout-bandit || true)
    check "errors: shows test id"       "$OUT" "B"
    check "errors: shows file"          "$OUT" "errors.py"
    check "errors: rule count format"   "$OUT" " ("
    check "errors: summary separator"   "$OUT" " · "

    OUT=$(bandit -f json clean.py 2>/dev/null | prettyout-bandit || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    with_config bandit group_by file
    OUT=$(bandit -f json errors.py 2>/dev/null | prettyout-bandit || true)
    check "group_by:file: shows filename" "$OUT" "errors.py"
    no_config

    with_config bandit colors false
    OUT=$(bandit -f json errors.py 2>/dev/null | prettyout-bandit || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config
else
    skip "bandit"
fi

# ── pylint ────────────────────────────────────────────────────────────────────
section "pylint"
if has_tool pylint; then
    mkdir -p /tmp/t-pylint && cd /tmp/t-pylint && no_config

    cat > errors.py << 'PY'
import os
x = undefined_var
PY
    cat > clean.py << 'PY'
"""Clean module."""


def add(a, b):
    """Add two numbers."""
    return a + b
PY

    OUT=$(pylint --output-format=json errors.py 2>/dev/null | prettyout-pylint || true)
    check "errors: shows message-id" "$OUT" "/"
    check "errors: shows file"       "$OUT" "errors.py"
    check "errors: rule count format" "$OUT" " ("
    check "errors: summary separator" "$OUT" " · "

    # pylint on minimal clean file still warns — just check no crash
    OUT=$(pylint --output-format=json clean.py 2>/dev/null | prettyout-pylint || true)
    check "clean: no crash" "$OUT" "issue"

    with_config pylint group_by file
    OUT=$(pylint --output-format=json errors.py 2>/dev/null | prettyout-pylint || true)
    check "group_by:file: shows filename" "$OUT" "errors.py"
    no_config

    with_config pylint colors false
    OUT=$(pylint --output-format=json errors.py 2>/dev/null | prettyout-pylint || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config
else
    skip "pylint"
fi

# ── eslint ────────────────────────────────────────────────────────────────────
section "eslint"
if has_tool eslint; then
    mkdir -p /tmp/t-eslint && cd /tmp/t-eslint && no_config

    cat > eslint.config.mjs << 'JS'
export default [
  {
    rules: {
      "no-undef": "error",
      "no-unused-vars": "warn",
      "eqeqeq": "warn"
    }
  }
];
JS
    cat > errors.js << 'JS'
var x = 1
var y = undefined_var
x + y
JS
    cat > clean.js << 'JS'
const x = 1;
const y = 2;
const z = x + y;
export default z;
JS

    OUT=$(eslint --format=json errors.js 2>/dev/null | prettyout-eslint || true)
    check "errors: shows rule"  "$OUT" "issue"
    check "errors: shows file"  "$OUT" "errors.js"

    OUT=$(eslint --format=json clean.js 2>/dev/null | prettyout-eslint || true)
    check "clean: 0 issues" "$OUT" "0 issues"
else
    skip "eslint"
fi

# ── biome ─────────────────────────────────────────────────────────────────────
section "biome"
if has_tool biome; then
    mkdir -p /tmp/t-biome && cd /tmp/t-biome && no_config
    printf '{"linter":{"enabled":true,"rules":{"recommended":true}}}' > biome.json

    cat > errors.ts << 'TS'
var x = 1
if (x == "1") { console.log("bad") }
debugger;
TS
    # Use tab indentation and double quotes so biome formatter is satisfied
    printf 'const x: number = 1;\nif (x === 1) {\n\tconsole.log("good");\n}\n' > clean.ts

    OUT=$(biome check --reporter=json errors.ts 2>/dev/null | prettyout-biome || true)
    check "errors: shows category"     "$OUT" "issue"
    check "errors: rule count format"  "$OUT" " ("
    check "errors: summary separator"  "$OUT" " · "

    OUT=$(biome check --reporter=json clean.ts 2>/dev/null | prettyout-biome || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    with_config biome group_by file
    OUT=$(biome check --reporter=json errors.ts 2>/dev/null | prettyout-biome || true)
    check "group_by:file: shows filename" "$OUT" "errors.ts"
    no_config

    with_config biome colors false
    OUT=$(biome check --reporter=json errors.ts 2>/dev/null | prettyout-biome || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config
else
    skip "biome"
fi

# ── shellcheck ────────────────────────────────────────────────────────────────
section "shellcheck"
if has_tool shellcheck; then
    mkdir -p /tmp/t-shellcheck && cd /tmp/t-shellcheck && no_config

    cat > errors.sh << 'SH'
#!/bin/bash
x=`echo hello`
if [ $x == "hello" ]; then
    echo "yes"
fi
SH
    cat > clean.sh << 'SH'
#!/bin/bash
x="hello"
echo "$x"
SH

    OUT=$(shellcheck --format=json errors.sh 2>/dev/null | prettyout-shellcheck || true)
    check "errors: shows SC code"      "$OUT" "SC"
    check "errors: shows file"         "$OUT" "errors.sh"
    check "errors: rule count format"  "$OUT" " ("
    check "errors: summary separator"  "$OUT" " · "

    OUT=$(shellcheck --format=json clean.sh 2>/dev/null | prettyout-shellcheck || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    with_config shellcheck group_by file
    OUT=$(shellcheck --format=json errors.sh 2>/dev/null | prettyout-shellcheck || true)
    check "group_by:file: shows filename" "$OUT" "errors.sh"
    no_config

    with_config shellcheck colors false
    OUT=$(shellcheck --format=json errors.sh 2>/dev/null | prettyout-shellcheck | cat || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config
else
    skip "shellcheck"
fi

# ── hadolint ──────────────────────────────────────────────────────────────────
section "hadolint"
if has_tool hadolint; then
    mkdir -p /tmp/t-hadolint && cd /tmp/t-hadolint && no_config

    cat > Dockerfile.errors << 'DF'
FROM ubuntu
RUN apt-get install vim
RUN apt-get install curl
DF
    cat > Dockerfile.clean << 'DF'
FROM ubuntu:24.04
RUN apt-get update && apt-get install -y --no-install-recommends vim=2:9.1.0016-1ubuntu7.1 && rm -rf /var/lib/apt/lists/*
DF

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

# ── golangci-lint ─────────────────────────────────────────────────────────────
# Note: use --disable-all --enable=ineffassign to avoid the buildir/goanalysis_metalinter
# crash (exit 3, empty stdout) that occurs when golangci-lint's Go version mismatches
# the system Go. ineffassign is pure-AST and always works.
section "golangci-lint"
if has_tool golangci-lint; then
    mkdir -p /tmp/t-golangci && cd /tmp/t-golangci && no_config
    # Write go.mod explicitly at go 1.21 — golangci-lint 1.61 (built with go1.23) refuses
    # to run against a newer Go version detected from go.mod, so pin to 1.21.
    printf 'module golangci_test\ngo 1.21\n' > go.mod

    cat > main.go << 'GO'
package main

import "fmt"

func main() {
    x := 42
    x = 100
    fmt.Println(x)
}
GO

    OUT=$(golangci-lint run --out-format=json --disable-all --enable=ineffassign 2>/dev/null | prettyout-golangci || true)
    check "errors: shows linter name"   "$OUT" "ineffassign"
    check "errors: shows issue count"   "$OUT" "issue"
    check "errors: shows summary"       "$OUT" "rule"
    check "errors: summary separator"   "$OUT" " · "

    cat > main.go << 'GO'
package main

import "fmt"

func main() { fmt.Println("hello") }
GO
    OUT=$(golangci-lint run --out-format=json --disable-all --enable=ineffassign 2>/dev/null | prettyout-golangci || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    cat > main.go << 'GO'
package main

import "fmt"

func main() {
    x := 42
    x = 100
    fmt.Println(x)
}
GO
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

# ── cargo clippy ──────────────────────────────────────────────────────────────
section "cargo clippy"
FIXTURES_CLIPPY=/project/test/fixtures/cargo-clippy
if has_tool cargo; then
    # errors fixture — 3 clippy warnings (needless_return, useless_vec, vec_init_then_push)
    cp -r "$FIXTURES_CLIPPY/errors" /tmp/t-cargo-errors
    cd /tmp/t-cargo-errors && no_config

    OUT=$(cargo clippy --message-format=json 2>/dev/null | prettyout-cargo-clippy || true)
    check "errors: shows rule"          "$OUT" "clippy::"
    check "errors: shows issue count"   "$OUT" "issue"
    check "errors: summary separator"   "$OUT" " · "
    check "errors: rule count format"   "$OUT" " ("
    # needless_return is on line 6 — plugin must use primary span line_start
    check "errors: primary span line"   "$OUT" "line 6"
    # artifacts filtered: only 3 clippy warnings counted, not compiler-artifact/build-finished entries
    check "errors: artifact filtering (3 issues)" "$OUT" "3 issues"

    with_config cargo-clippy group_by file
    OUT=$(cargo clippy --message-format=json 2>/dev/null | prettyout-cargo-clippy || true)
    check "group_by:file: shows filename"  "$OUT" "src/main.rs"
    no_config

    with_config cargo-clippy colors false
    OUT=$(cargo clippy --message-format=json 2>/dev/null | prettyout-cargo-clippy || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config

    # clean fixture — no warnings
    cp -r "$FIXTURES_CLIPPY/clean" /tmp/t-cargo-clean
    cd /tmp/t-cargo-clean && no_config
    OUT=$(cargo clippy --message-format=json 2>/dev/null | prettyout-cargo-clippy || true)
    check "clean: 0 issues" "$OUT" "0 issues · 0 rules · 0 files"

    # rustc-warning fixture — unused_variables is a rustc code, not a clippy:: lint
    cp -r "$FIXTURES_CLIPPY/rustc-warning" /tmp/t-cargo-rustc
    cd /tmp/t-cargo-rustc && no_config
    OUT=$(cargo clippy --message-format=json 2>/dev/null | prettyout-cargo-clippy || true)
    check "rustc code: shown without clippy:: prefix" "$OUT" "unused_variables"
    check_absent "rustc code: no clippy:: prefix" "$OUT" "clippy::unused_variables"
else
    skip "cargo"
fi

# ── stylelint ──────────────────────────────────────────────────────────────────
section "stylelint"
if has_tool stylelint; then
    mkdir -p /tmp/t-stylelint && cd /tmp/t-stylelint && no_config

    cat > .stylelintrc.json << 'JSON'
{"extends": "stylelint-config-standard"}
JSON
    cat > errors.css << 'CSS'
a { color: #gggggg; FONT-SIZE: 12px }
CSS
    cat > clean.css << 'CSS'
a {
  color: #fff;
  font-size: 12px;
}
CSS

    # stylelint writes JSON to stderr; redirect stderr to stdin of plugin
    OUT=$(stylelint --formatter=json errors.css 2>&1 >/dev/null | prettyout-stylelint || true)
    check "errors: shows rule"        "$OUT" "declaration-property-value-no-unknown"
    check "errors: shows file"        "$OUT" "errors.css"
    check "errors: shows issue count" "$OUT" "issue"

    OUT=$(stylelint --formatter=json clean.css 2>&1 >/dev/null | prettyout-stylelint || true)
    check "clean: 0 issues" "$OUT" "0 issues"

    with_config stylelint group_by file
    OUT=$(stylelint --formatter=json errors.css 2>&1 >/dev/null | prettyout-stylelint || true)
    check "group_by:file: shows filename" "$OUT" "errors.css"
    no_config

    with_config stylelint colors false
    OUT=$(stylelint --formatter=json errors.css 2>&1 >/dev/null | prettyout-stylelint | cat || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config
else
    skip "stylelint"
fi

# ── trivy ─────────────────────────────────────────────────────────────────────
section "trivy"
if has_tool trivy; then
    mkdir -p /tmp/t-trivy && cd /tmp/t-trivy && no_config

    # Scan a small known-vulnerable image
    OUT=$(trivy image --format=json --quiet alpine:3.11 2>/dev/null | prettyout-trivy || true)
    check "image scan: no crash" "$OUT" "vulnerabilit"

    # Filesystem scan on clean dir
    mkdir -p /tmp/t-trivy/clean
    OUT=$(trivy fs --format=json --quiet /tmp/t-trivy/clean 2>/dev/null | prettyout-trivy || true)
    check "clean fs: no crash" "$OUT" ""
else
    skip "trivy"
fi

# ── Summary ───────────────────────────────────────────────────────────────────
printf '\n\033[1m══ Results ══\033[0m\n'
printf '\033[1;32mPASS: %d\033[0m  ' "$PASS"
printf '\033[1;31mFAIL: %d\033[0m  ' "$FAIL"
printf '\033[2mSKIP: %d\033[0m\n'   "$SKIP"

[ "$FAIL" -eq 0 ]
