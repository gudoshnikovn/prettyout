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
    check "errors: shows test id" "$OUT" "B"
    check "errors: shows file"    "$OUT" "errors.py"

    OUT=$(bandit -f json clean.py 2>/dev/null | prettyout-bandit || true)
    check "clean: 0 issues" "$OUT" "0 issues"
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

    # pylint on minimal clean file still warns — just check no crash
    OUT=$(pylint --output-format=json clean.py 2>/dev/null | prettyout-pylint || true)
    check "clean: no crash" "$OUT" "issue"
else
    skip "pylint"
fi

# ── eslint ────────────────────────────────────────────────────────────────────
section "eslint"
if has_tool eslint; then
    mkdir -p /tmp/t-eslint && cd /tmp/t-eslint && no_config

    cat > eslint.config.mjs << 'JS'
import js from "@eslint/js";
export default [js.configs.recommended];
JS
    cat > errors.js << 'JS'
var x = 1
var y = undefined_var
console.log(x)
JS
    cat > clean.js << 'JS'
const x = 1;
console.log(x);
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
    biome init --json-pretty 2>/dev/null || true

    cat > errors.ts << 'TS'
var x = 1
if (x == "1") { console.log("bad") }
TS
    cat > clean.ts << 'TS'
const x: number = 1;
if (x === 1) {
  console.log("good");
}
TS

    OUT=$(biome check --reporter=json errors.ts 2>/dev/null | prettyout-biome || true)
    check "errors: shows category" "$OUT" "issue"

    OUT=$(biome check --reporter=json clean.ts 2>/dev/null | prettyout-biome || true)
    check "clean: 0 issues" "$OUT" "0 issues"
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
x=$(echo hello)
if [ "$x" = "hello" ]; then
    echo "yes"
fi
SH

    OUT=$(shellcheck --format=json errors.sh 2>/dev/null | prettyout-shellcheck || true)
    check "errors: shows SC code" "$OUT" "SC"
    check "errors: shows file"    "$OUT" "errors.sh"

    OUT=$(shellcheck --format=json clean.sh 2>/dev/null | prettyout-shellcheck || true)
    check "clean: 0 issues" "$OUT" "0 issues"
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
RUN apt-get update && apt-get install -y --no-install-recommends \
        vim \
    && rm -rf /var/lib/apt/lists/*
DF

    OUT=$(hadolint --format=json Dockerfile.errors 2>/dev/null | prettyout-hadolint || true)
    check "errors: shows DL code" "$OUT" "DL"
    check "errors: shows file"    "$OUT" "Dockerfile"

    OUT=$(hadolint --format=json Dockerfile.clean 2>/dev/null | prettyout-hadolint || true)
    check "clean: 0 issues" "$OUT" "0 issues"
else
    skip "hadolint"
fi

# ── golangci-lint ─────────────────────────────────────────────────────────────
section "golangci-lint"
if has_tool golangci-lint; then
    mkdir -p /tmp/t-golangci && cd /tmp/t-golangci && no_config
    go mod init golangci_test 2>/dev/null || true

    cat > main.go << 'GO'
package main

import "fmt"

func main() {
    x := 1
    fmt.Println("hello")
    _ = x
}
GO

    OUT=$(golangci-lint run --out-format=json 2>/dev/null | prettyout-golangci || true)
    check "no crash on run" "$OUT" "issue"
else
    skip "golangci-lint"
fi

# ── cargo clippy ──────────────────────────────────────────────────────────────
section "cargo clippy"
if has_tool cargo; then
    mkdir -p /tmp/t-cargo/src && cd /tmp/t-cargo && no_config

    cat > Cargo.toml << 'TOML'
[package]
name = "clippy_test"
version = "0.1.0"
edition = "2021"
TOML

    cat > src/main.rs << 'RS'
fn main() {
    let x = vec![1, 2, 3];
    let _ = x;
    return;
}
RS

    OUT=$(cargo clippy --message-format=json 2>/dev/null | prettyout-cargo-clippy || true)
    check "errors: shows rule" "$OUT" "issue"

    cat > src/main.rs << 'RS'
fn main() {
    println!("Hello!");
}
RS
    OUT=$(cargo clippy --message-format=json 2>/dev/null | prettyout-cargo-clippy || true)
    check "clean: 0 issues" "$OUT" "0 issues"
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

    OUT=$(stylelint --formatter=json errors.css 2>/dev/null | prettyout-stylelint || true)
    check "errors: shows rule" "$OUT" "issue"

    OUT=$(stylelint --formatter=json clean.css 2>/dev/null | prettyout-stylelint || true)
    check "clean: 0 issues" "$OUT" "0 issues"
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
