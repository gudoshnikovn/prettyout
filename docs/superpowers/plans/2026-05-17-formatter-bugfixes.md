# Formatter Bug Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix 5 bugs across formatter plugins: severity prefix for 8 tools, ruff fixable count, hadolint singular/plural, pylint json2+rating, trivy group_by:file.

**Architecture:** Each formatter is a standalone `cmd/prettyout-<tool>/main.go`. Bugs are isolated to individual files. Test scripts are in `test/tools/<tool>.sh` and run inside Docker via `test/run.sh`. semgrep already has severity prefix — skip it.

**Tech Stack:** Go, Docker (`prettyout-test` image), bash test scripts in `test/tools/`.

---

## File Map

**Modify:**
- `cmd/prettyout-hadolint/main.go` — fix `lines 1` singular/plural + add severity prefix
- `cmd/prettyout-mypy/main.go` — add severity prefix
- `cmd/prettyout-shellcheck/main.go` — add severity prefix
- `cmd/prettyout-eslint/main.go` — add severity prefix
- `cmd/prettyout-stylelint/main.go` — add severity prefix
- `cmd/prettyout-biome/main.go` — add severity prefix
- `cmd/prettyout-cargo-clippy/main.go` — add severity prefix
- `cmd/prettyout-pylint/main.go` — add severity prefix + switch to json2 + show rating
- `cmd/prettyout-ruff/main.go` — add fixable count
- `cmd/prettyout-trivy/main.go` — implement group_by:file
- `internal/registry/builtin.yaml` — update pylint output_args to json2
- `test/tools/hadolint.sh` — add severity prefix check
- `test/tools/mypy.sh` — add severity prefix check
- `test/tools/shellcheck.sh` — add severity prefix check
- `test/tools/eslint.sh` — add severity prefix check
- `test/tools/stylelint.sh` — add severity prefix check
- `test/tools/biome.sh` — add severity prefix check
- `test/tools/cargo-clippy.sh` — add severity prefix check
- `test/tools/pylint.sh` — update to json2 + add rating + severity check
- `test/tools/ruff.sh` — add fixable check
- `test/tools/trivy.sh` — add group_by:file check

---

## Severity prefix pattern (reference)

All 8 tools use this same pattern in `formatByRule`. Add a `severityLabel` function and update the rule header `fmt.Printf`.

**Add this function** (exact values differ per tool — see each task):
```go
func severityLabel(sev string) string {
    switch sev {
    case "error", "fatal":
        return "ERROR"
    case "warning":
        return "WARN"
    default:
        return "INFO"
    }
}
```

**Change the rule header print** (currently varies per tool, always this shape):
```go
// BEFORE (no-color branch):
fmt.Printf("%s (%d) — %s\n", ruleCode, count, r.message)

// AFTER:
label := severityLabel(r.severity)
if cfg.Colors {
    fmt.Printf("%s[%s]%s %s (%d) — %s\n", col, label, reset, ruleCode, count, r.message)
} else {
    fmt.Printf("[%s] %s (%d) — %s\n", label, ruleCode, count, r.message)
}
```

---

## Task 1: hadolint — singular/plural fix + severity prefix

**Files:**
- Modify: `cmd/prettyout-hadolint/main.go`
- Modify: `test/tools/hadolint.sh`

- [ ] **Update test script to check for new behavior (currently failing)**

In `test/tools/hadolint.sh`, after the existing checks, add:
```bash
with_config hadolint colors false
OUT=$(hadolint --format=json Dockerfile.errors 2>/dev/null | prettyout-hadolint || true)
check "severity prefix present"  "$OUT" "[WARN]"
check "singular line label"       "$OUT" "line 1"
check_absent "no plural for single" "$OUT" "lines 1"
no_config
```

- [ ] **Run failing test to confirm it fails**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  go build -o /usr/local/bin/prettyout-hadolint /project/cmd/prettyout-hadolint/ 2>&1
  bash /project/test/tools/hadolint.sh
" 2>&1 | grep -E "PASS|FAIL|SKIP"
```

Expected: the three new checks FAIL.

- [ ] **Fix `formatByRule` in `cmd/prettyout-hadolint/main.go`**

Add `severityLabel` function (add before `format`):
```go
func severityLabel(level string) string {
    switch level {
    case "error":
        return "ERROR"
    case "warning":
        return "WARN"
    default:
        return "INFO"
    }
}
```

In `formatByRule`, replace the existing rule header block:
```go
		col := hadolintColor(r.severity, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		if cfg.Colors {
			fmt.Printf("%s%s%s (%d) — %s\n", col, rule, reset, count, r.message)
		} else {
			fmt.Printf("%s (%d) — %s\n", rule, count, r.message)
		}
```

Replace with:
```go
		col := hadolintColor(r.severity, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		label := severityLabel(r.severity)
		if cfg.Colors {
			fmt.Printf("%s[%s]%s %s (%d) — %s\n", col, label, reset, rule, count, r.message)
		} else {
			fmt.Printf("[%s] %s (%d) — %s\n", label, rule, count, r.message)
		}
```

Also fix the singular/plural bug in `formatByRule`. Find:
```go
			fmt.Printf("  - %s — lines %s\n", f, strings.Join(lineStrs, ", "))
```

Replace with:
```go
			lineWord := formatter.Plural(len(ls), "line", "lines")
			fmt.Printf("  - %s — %s %s\n", f, lineWord, strings.Join(lineStrs, ", "))
```

Remove `"strings"` from imports if it's now unused (check — `strings.Join` is still used, so keep it).

- [ ] **Build and run test**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  go build -o /usr/local/bin/prettyout-hadolint /project/cmd/prettyout-hadolint/ 2>&1
  bash /project/test/tools/hadolint.sh
" 2>&1 | grep -E "PASS|FAIL|SKIP"
```

Expected: all checks PASS.

- [ ] **Commit**

```bash
git add cmd/prettyout-hadolint/main.go test/tools/hadolint.sh
git commit -m "fix(hadolint): add severity prefix, fix singular line label"
```

---

## Task 2: mypy — severity prefix

**Files:**
- Modify: `cmd/prettyout-mypy/main.go`
- Modify: `test/tools/mypy.sh`

- [ ] **Update test script**

In `test/tools/mypy.sh`, after existing checks, add:
```bash
with_config mypy colors false
OUT=$(mypy --output=json errors.py 2>/dev/null | prettyout-mypy || true)
check "severity prefix present" "$OUT" "[ERROR]"
no_config
```

- [ ] **Run failing test**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  go build -o /usr/local/bin/prettyout-mypy /project/cmd/prettyout-mypy/ 2>&1
  bash /project/test/tools/mypy.sh
" 2>&1 | grep -E "PASS|FAIL|SKIP"
```

Expected: severity prefix check FAILS.

- [ ] **Fix `formatByRule` in `cmd/prettyout-mypy/main.go`**

Add `severityLabel` before `format`:
```go
func severityLabel(sev string) string {
    switch sev {
    case "error":
        return "ERROR"
    case "warning":
        return "WARN"
    default:
        return "INFO"
    }
}
```

In `formatByRule`, find the rule header print block:
```go
		col := formatter.SeverityColor(r.severity, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		if cfg.Colors {
			fmt.Printf("%s%s%s (%d) — %s\n", col, rule, reset, count, r.message)
		} else {
			fmt.Printf("%s (%d) — %s\n", rule, count, r.message)
		}
```

Replace with:
```go
		col := formatter.SeverityColor(r.severity, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		label := severityLabel(r.severity)
		if cfg.Colors {
			fmt.Printf("%s[%s]%s %s (%d) — %s\n", col, label, reset, rule, count, r.message)
		} else {
			fmt.Printf("[%s] %s (%d) — %s\n", label, rule, count, r.message)
		}
```

- [ ] **Build and run test**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  go build -o /usr/local/bin/prettyout-mypy /project/cmd/prettyout-mypy/ 2>&1
  bash /project/test/tools/mypy.sh
" 2>&1 | grep -E "PASS|FAIL|SKIP"
```

Expected: all PASS.

- [ ] **Commit**

```bash
git add cmd/prettyout-mypy/main.go test/tools/mypy.sh
git commit -m "fix(mypy): add severity prefix to rule headers"
```

---

## Task 3: shellcheck — severity prefix

**Files:**
- Modify: `cmd/prettyout-shellcheck/main.go`
- Modify: `test/tools/shellcheck.sh`

- [ ] **Update test script**

In `test/tools/shellcheck.sh`, add:
```bash
with_config shellcheck colors false
OUT=$(shellcheck --format=json errors.sh 2>/dev/null | prettyout-shellcheck || true)
check "severity prefix present" "$OUT" "[WARN]"
no_config
```

- [ ] **Run failing test**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  go build -o /usr/local/bin/prettyout-shellcheck /project/cmd/prettyout-shellcheck/ 2>&1
  bash /project/test/tools/shellcheck.sh
" 2>&1 | grep -E "PASS|FAIL|SKIP"
```

Expected: severity check FAILS.

- [ ] **Fix `cmd/prettyout-shellcheck/main.go`**

Add `severityLabel` function before `format`:
```go
func severityLabel(level string) string {
    switch level {
    case "error":
        return "ERROR"
    case "warning":
        return "WARN"
    default:
        return "INFO"
    }
}
```

In `formatByRule`, find the rule header print. It currently uses `shellcheckColor`. Replace the print block with:
```go
		col := shellcheckColor(r.severity, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		label := severityLabel(r.severity)
		if cfg.Colors {
			fmt.Printf("%s[%s]%s %s (%d) — %s\n", col, label, reset, rule, count, r.message)
		} else {
			fmt.Printf("[%s] %s (%d) — %s\n", label, rule, count, r.message)
		}
```

- [ ] **Build and run test**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  go build -o /usr/local/bin/prettyout-shellcheck /project/cmd/prettyout-shellcheck/ 2>&1
  bash /project/test/tools/shellcheck.sh
" 2>&1 | grep -E "PASS|FAIL|SKIP"
```

Expected: all PASS.

- [ ] **Commit**

```bash
git add cmd/prettyout-shellcheck/main.go test/tools/shellcheck.sh
git commit -m "fix(shellcheck): add severity prefix to rule headers"
```

---

## Task 4: eslint + stylelint — severity prefix

**Files:**
- Modify: `cmd/prettyout-eslint/main.go`
- Modify: `cmd/prettyout-stylelint/main.go`
- Modify: `test/tools/eslint.sh`
- Modify: `test/tools/stylelint.sh`

- [ ] **Update test scripts**

In `test/tools/eslint.sh`, add:
```bash
with_config eslint colors false
OUT=$(eslint --format=json errors.js 2>/dev/null | prettyout-eslint || true)
check "severity prefix present" "$OUT" "[ERROR]"
no_config
```

In `test/tools/stylelint.sh`, add:
```bash
with_config stylelint colors false
OUT=$(stylelint --formatter=json errors.css 2>&1 >/dev/null | prettyout-stylelint || true)
check "severity prefix present" "$OUT" "[ERROR]"
no_config
```

- [ ] **Fix `cmd/prettyout-eslint/main.go`**

eslint stores severity as string `"error"`/`"warning"` in `ruleEntry.severity` (set via `severityStr`). Add `severityLabel`:
```go
func severityLabel(sev string) string {
    switch sev {
    case "error":
        return "ERROR"
    default:
        return "WARN"
    }
}
```

In `formatByRule`, find the rule header print block and apply the same pattern as previous tasks:
```go
		label := severityLabel(r.severity)
		if cfg.Colors {
			fmt.Printf("%s[%s]%s %s (%d) — %s\n", col, label, reset, rule, count, r.message)
		} else {
			fmt.Printf("[%s] %s (%d) — %s\n", label, rule, count, r.message)
		}
```

- [ ] **Fix `cmd/prettyout-stylelint/main.go`**

stylelint stores `severity` as `"error"`/`"warning"` in `issueItem.severity`. Add `severityLabel`:
```go
func severityLabel(sev string) string {
    switch sev {
    case "error":
        return "ERROR"
    default:
        return "WARN"
    }
}
```

In `formatByRule`, apply the same header pattern. The rule entry has `r.severity` already set.

- [ ] **Build and run tests**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  go build -o /usr/local/bin/prettyout-eslint /project/cmd/prettyout-eslint/ 2>&1
  go build -o /usr/local/bin/prettyout-stylelint /project/cmd/prettyout-stylelint/ 2>&1
  bash /project/test/tools/eslint.sh
  bash /project/test/tools/stylelint.sh
" 2>&1 | grep -E "PASS|FAIL|SKIP"
```

Expected: all PASS.

- [ ] **Commit**

```bash
git add cmd/prettyout-eslint/main.go cmd/prettyout-stylelint/main.go test/tools/eslint.sh test/tools/stylelint.sh
git commit -m "fix(eslint,stylelint): add severity prefix to rule headers"
```

---

## Task 5: biome + cargo-clippy — severity prefix

**Files:**
- Modify: `cmd/prettyout-biome/main.go`
- Modify: `cmd/prettyout-cargo-clippy/main.go`
- Modify: `test/tools/biome.sh`
- Modify: `test/tools/cargo-clippy.sh`

- [ ] **Update test scripts**

In `test/tools/biome.sh`, add:
```bash
with_config biome colors false
OUT=$(biome check --reporter=json errors.ts 2>/dev/null | prettyout-biome || true)
check "severity prefix present" "$OUT" "[ERROR]"
no_config
```

In `test/tools/cargo-clippy.sh`, add:
```bash
with_config cargo-clippy colors false
OUT=$(cargo clippy --message-format=json 2>/dev/null | prettyout-cargo-clippy || true)
check "severity prefix present" "$OUT" "[WARN]"
no_config
```

- [ ] **Fix `cmd/prettyout-biome/main.go`**

biome stores `d.Severity` as `"error"`/`"warning"` (lowercase). Add `severityLabel`:
```go
func severityLabel(sev string) string {
    switch strings.ToLower(sev) {
    case "error", "fatal":
        return "ERROR"
    case "warning":
        return "WARN"
    default:
        return "INFO"
    }
}
```

Add `"strings"` to imports if not already present. Apply the same header pattern in `formatByRule` — biome's `ruleEntry` has `r.severity`.

The current biome color call uses `formatter.SeverityColor`. Keep it, add label around it:
```go
		col := formatter.SeverityColor(strings.ToLower(r.severity), cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		label := severityLabel(r.severity)
		if cfg.Colors {
			fmt.Printf("%s[%s]%s %s (%d) — %s\n", col, label, reset, rule, count, r.message)
		} else {
			fmt.Printf("[%s] %s (%d) — %s\n", label, rule, count, r.message)
		}
```

- [ ] **Fix `cmd/prettyout-cargo-clippy/main.go`**

cargo-clippy filters to `"compiler-message"` and stores level as `"error"`/`"warning"`. Add `severityLabel`:
```go
func severityLabel(level string) string {
    switch level {
    case "error":
        return "ERROR"
    case "warning":
        return "WARN"
    default:
        return "INFO"
    }
}
```

Apply the same header pattern in `formatByRule`. The `ruleEntry` has `r.severity` set from `cl.Message.Level`.

- [ ] **Build and run tests** (run from the errors fixture dir for cargo-clippy)

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  go build -o /usr/local/bin/prettyout-biome /project/cmd/prettyout-biome/ 2>&1
  go build -o /usr/local/bin/prettyout-cargo-clippy /project/cmd/prettyout-cargo-clippy/ 2>&1
  bash /project/test/tools/biome.sh
  bash /project/test/tools/cargo-clippy.sh
" 2>&1 | grep -E "PASS|FAIL|SKIP"
```

Expected: all PASS.

- [ ] **Commit**

```bash
git add cmd/prettyout-biome/main.go cmd/prettyout-cargo-clippy/main.go test/tools/biome.sh test/tools/cargo-clippy.sh
git commit -m "fix(biome,cargo-clippy): add severity prefix to rule headers"
```

---

## Task 6: pylint — severity prefix + json2 + rating

**Files:**
- Modify: `cmd/prettyout-pylint/main.go`
- Modify: `internal/registry/builtin.yaml`
- Modify: `test/tools/pylint.sh`

The json2 output wraps messages:
```json
{"messages": [...same fields as json...], "statistics": {"score": 0.0, ...}}
```

- [ ] **Update test script**

Replace the existing pylint test content in `test/tools/pylint.sh` with:
```bash
section "pylint"
FIXTURES="$SCRIPT_DIR/fixtures/pylint"
if has_tool pylint; then
    mkdir -p /tmp/t-pylint && cd /tmp/t-pylint && no_config
    cp "$FIXTURES/errors.py" .
    cp "$FIXTURES/clean.py" .

    OUT=$(pylint --output-format=json2 errors.py 2>/dev/null | prettyout-pylint || true)
    check "errors: shows message-id"      "$OUT" "/"
    check "errors: shows file"            "$OUT" "errors.py"
    check "errors: rule count format"     "$OUT" " ("
    check "errors: summary separator"     "$OUT" " · "
    check "errors: shows severity prefix" "$OUT" "[ERROR]"
    check "errors: shows rating"          "$OUT" "rated"
    check "errors: rating format"         "$OUT" "/10"

    OUT=$(pylint --output-format=json2 clean.py 2>/dev/null | prettyout-pylint || true)
    check "clean: no crash" "$OUT" "issue"

    with_config pylint group_by file
    OUT=$(pylint --output-format=json2 errors.py 2>/dev/null | prettyout-pylint || true)
    check "group_by:file: shows filename" "$OUT" "errors.py"
    no_config

    with_config pylint colors false
    OUT=$(pylint --output-format=json2 errors.py 2>/dev/null | prettyout-pylint || true)
    check_absent "colors:false: no ANSI codes" "$OUT" $'\033['
    no_config
else
    skip "pylint"
fi
```

- [ ] **Run failing tests**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  go build -o /usr/local/bin/prettyout-pylint /project/cmd/prettyout-pylint/ 2>&1
  bash /project/test/tools/pylint.sh
" 2>&1 | grep -E "PASS|FAIL|SKIP"
```

Expected: severity prefix, rating, json2 checks FAIL.

- [ ] **Update `cmd/prettyout-pylint/main.go`**

**1. Change the top-level struct for json2:**

Replace the `format` function signature and unmarshalling. At the top of the file, add a new wrapper type:
```go
type pylintJSON2 struct {
    Messages   []pylintMsg `json:"messages"`
    Statistics struct {
        Score float64 `json:"score"`
    } `json:"statistics"`
}
```

Change `format` to:
```go
func format(data []byte, cfg formatter.Config) error {
    var wrapper pylintJSON2
    if err := json.Unmarshal(data, &wrapper); err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }
    msgs := wrapper.Messages
    score := wrapper.Statistics.Score

    if cfg.GroupBy == "file" {
        return formatByFile(msgs, cfg)
    }
    return formatByRule(msgs, cfg, score)
}
```

**2. Add `severityLabel`:**
```go
func severityLabel(t string) string {
    switch t {
    case "error", "fatal":
        return "ERROR"
    case "warning":
        return "WARN"
    default:
        return "INFO"
    }
}
```

**3. Update `formatByRule` signature and rule header:**

Change signature to `func formatByRule(msgs []pylintMsg, cfg formatter.Config, score float64) error`.

In the rule header print block, replace:
```go
		if cfg.Colors {
			fmt.Printf("%s%s%s (%d) — %s\n", col, r.display, reset, count, r.message)
		} else {
			fmt.Printf("%s (%d) — %s\n", r.display, count, r.message)
		}
```

With:
```go
		label := severityLabel(msgType)
		if cfg.Colors {
			fmt.Printf("%s[%s]%s %s (%d) — %s\n", col, label, reset, r.display, count, r.message)
		} else {
			fmt.Printf("[%s] %s (%d) — %s\n", label, r.display, count, r.message)
		}
```

**4. Add rating after summary in `formatByRule`:**

After `fmt.Println(formatter.Summary(...))`, add:
```go
    fmt.Printf("  ↳ rated %.2f/10\n", score)
```

- [ ] **Update registry `internal/registry/builtin.yaml`**

Find the pylint entry:
```yaml
  pylint:
    plugin: prettyout-pylint
    output_args: [--output-format=json]
```

Change to:
```yaml
  pylint:
    plugin: prettyout-pylint
    output_args: [--output-format=json2]
```

- [ ] **Build and run tests**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  go build -o /usr/local/bin/prettyout-pylint /project/cmd/prettyout-pylint/ 2>&1
  bash /project/test/tools/pylint.sh
" 2>&1 | grep -E "PASS|FAIL|SKIP"
```

Expected: all PASS.

- [ ] **Commit**

```bash
git add cmd/prettyout-pylint/main.go internal/registry/builtin.yaml test/tools/pylint.sh
git commit -m "fix(pylint): add severity prefix, switch to json2, show rating"
```

---

## Task 7: ruff — fixable count

**Files:**
- Modify: `cmd/prettyout-ruff/main.go`
- Modify: `test/tools/ruff.sh`

ruff JSON: `"fix": null` means not fixable, `"fix": {...}` means fixable.

- [ ] **Update test script**

In `test/tools/ruff.sh`, after existing checks, add:
```bash
with_config ruff colors false
OUT=$(ruff check --output-format=json errors.py 2>/dev/null | prettyout-ruff || true)
check "fixable hint present" "$OUT" "fixable with --fix"
no_config
```

- [ ] **Run failing test**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  go build -o /usr/local/bin/prettyout-ruff /project/cmd/prettyout-ruff/ 2>&1
  bash /project/test/tools/ruff.sh
" 2>&1 | grep -E "PASS|FAIL|SKIP"
```

Expected: fixable hint check FAILS.

- [ ] **Update `cmd/prettyout-ruff/main.go`**

Add `Fix` field to the `issue` struct:
```go
type issue struct {
    Code        string      `json:"code"`
    Message     string      `json:"message"`
    Filename    string      `json:"filename"`
    Location    location    `json:"location"`
    EndLocation location    `json:"end_location"`
    Fix         interface{} `json:"fix"`
}
```

In `formatByRule`, count fixable issues and print the hint. After the `fmt.Println(formatter.Summary(...))` line, add:
```go
    fixable := 0
    for _, iss := range issues {
        if iss.Fix != nil {
            fixable++
        }
    }
    if fixable > 0 {
        fmt.Printf("  ↳ %d %s with --fix\n", fixable, formatter.Plural(fixable, "fixable", "fixable"))
    }
```

Note: `formatter.Plural` with same singular and plural just returns the word — use a plain string instead:
```go
    if fixable > 0 {
        fmt.Printf("  ↳ %d fixable with --fix\n", fixable)
    }
```

- [ ] **Build and run test**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  go build -o /usr/local/bin/prettyout-ruff /project/cmd/prettyout-ruff/ 2>&1
  bash /project/test/tools/ruff.sh
" 2>&1 | grep -E "PASS|FAIL|SKIP"
```

Expected: all PASS.

- [ ] **Commit**

```bash
git add cmd/prettyout-ruff/main.go test/tools/ruff.sh
git commit -m "fix(ruff): show fixable count after summary"
```

---

## Task 8: trivy — implement group_by:file

**Files:**
- Modify: `cmd/prettyout-trivy/main.go`
- Modify: `test/tools/trivy.sh`

Currently `cfg.GroupBy` is never read. `group_by:file` should group by `Target` (e.g. `requirements.txt`, `Cargo.lock`), showing vulns sorted by severity rank within each target.

- [ ] **Update test script**

In `test/tools/trivy.sh`, add:
```bash
with_config trivy group_by file
OUT=$(trivy fs --format=json --quiet . 2>/dev/null | prettyout-trivy || true)
check "group_by:file: shows target filename" "$OUT" "requirements.txt"
no_config
```

- [ ] **Run failing test**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  cp -r /project/test/fixtures/trivy/django /tmp/t-trivy && cd /tmp/t-trivy
  go build -o /usr/local/bin/prettyout-trivy /project/cmd/prettyout-trivy/ 2>&1
  bash /project/test/tools/trivy.sh
" 2>&1 | grep -E "PASS|FAIL|SKIP"
```

Expected: group_by:file check FAILS (currently both modes output identically).

- [ ] **Add `formatByFile` to `cmd/prettyout-trivy/main.go`**

In the `format` function, add the `GroupBy` dispatch (currently missing). Replace:
```go
func format(data []byte, cfg formatter.Config) error {
    var report trivyReport
    if err := json.Unmarshal(data, &report); err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }

    // Collect all vulns grouped by severity
    bySeverity := map[string][]trivyVuln{}
```

With:
```go
func format(data []byte, cfg formatter.Config) error {
    var report trivyReport
    if err := json.Unmarshal(data, &report); err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }

    if cfg.GroupBy == "file" {
        return formatByFile(report, cfg)
    }

    // Collect all vulns grouped by severity
    bySeverity := map[string][]trivyVuln{}
```

Add the `formatByFile` function at the end of the file:
```go
func formatByFile(report trivyReport, cfg formatter.Config) error {
    totalVulns := 0
    targets := 0

    for _, result := range report.Results {
        if len(result.Vulnerabilities) == 0 {
            continue
        }
        targets++
        vulns := result.Vulnerabilities
        totalVulns += len(vulns)

        fmt.Println(result.Target)

        // Sort by severity rank then CVE ID
        sort.Slice(vulns, func(i, j int) bool {
            ri := severityRank(vulns[i].Severity)
            rj := severityRank(vulns[j].Severity)
            if ri != rj {
                return ri < rj
            }
            return vulns[i].VulnerabilityID < vulns[j].VulnerabilityID
        })

        for _, v := range vulns {
            sev := v.Severity
            if sev == "" {
                sev = "UNKNOWN"
            }
            col := trivyColor(sev, cfg.Colors)
            reset := ""
            if cfg.Colors {
                reset = "\033[0m"
            }
            if v.FixedVersion != "" {
                fmt.Printf("  %s%-8s%s %s — %s %s → fix: %s\n", col, sev, reset, v.VulnerabilityID, v.PkgName, v.InstalledVersion, v.FixedVersion)
            } else {
                fmt.Printf("  %s%-8s%s %s — %s %s → no fix available\n", col, sev, reset, v.VulnerabilityID, v.PkgName, v.InstalledVersion)
            }
        }
        fmt.Println("────────────────────────────────────────────────")
    }

    if totalVulns == 0 {
        fmt.Println("No vulnerabilities found.")
        return nil
    }

    fmt.Printf("%d %s · %d %s\n",
        totalVulns, formatter.Plural(totalVulns, "vulnerability", "vulnerabilities"),
        targets, formatter.Plural(targets, "target", "targets"))
    return nil
}
```

- [ ] **Build and run test**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  cp -r /project/test/fixtures/trivy/django /tmp/t-trivy && cd /tmp/t-trivy
  go build -o /usr/local/bin/prettyout-trivy /project/cmd/prettyout-trivy/ 2>&1
  bash /project/test/tools/trivy.sh
" 2>&1 | grep -E "PASS|FAIL|SKIP"
```

Expected: all PASS.

- [ ] **Commit**

```bash
git add cmd/prettyout-trivy/main.go test/tools/trivy.sh
git commit -m "fix(trivy): implement group_by:file view grouped by target"
```

---

## Task 9: Full build and test suite

- [ ] **Build all binaries**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
go build ./...
```

Expected: no errors.

- [ ] **Run Go tests**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
go test ./...
```

Expected: all PASS.

- [ ] **Run full integration test suite in Docker**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash /project/test/run.sh 2>&1 | tail -20
```

Expected: all tools PASS, no FAIL lines.

---

## Task 10: Regenerate showcase docs for affected tools

**Files:** Update these docs with new real outputs:
- `docs/tools/ruff.md` — new `↳ N fixable` line
- `docs/tools/mypy.md` — `[ERROR]` prefix
- `docs/tools/eslint.md` — `[ERROR]`/`[WARN]` prefix
- `docs/tools/stylelint.md` — `[ERROR]` prefix
- `docs/tools/shellcheck.md` — `[WARN]` prefix
- `docs/tools/hadolint.md` — `[WARN]` prefix + `line 1` (not `lines 1`)
- `docs/tools/biome.md` — `[ERROR]` prefix
- `docs/tools/cargo-clippy.md` — `[WARN]` prefix
- `docs/tools/pylint.md` — `[ERROR]` prefix + `↳ rated 0.00/10`
- `docs/tools/trivy.md` — group_by:file now shows different output

For each tool, run the capture command and overwrite the file. Use the same Docker commands from the original showcase plan (`2026-05-17-showcase-docs.md`) — they are unchanged except pylint now uses `--output-format=json2`.

The pylint capture command changes to:
```bash
pylint --output-format=json2 errors.py 2>/dev/null | prettyout-pylint
```

Also update `docs/summary.md` comparison table: remove the caveat "mypy" from "Severity not shown" row — all 9 tools now show it.

- [ ] **Regenerate all 10 tool docs + update summary.md**

Run the Docker capture for each tool and rewrite the file. Commit all at once:

```bash
git add docs/tools/ docs/summary.md
git commit -m "docs: regenerate showcase docs with severity prefix and bug fixes"
```

---

## Self-Review

**Spec coverage:**
- [x] Severity prefix for 8 tools (mypy, shellcheck, eslint, stylelint, biome, cargo-clippy, pylint, hadolint) — Tasks 1–6
- [x] semgrep already has it — confirmed, skipped
- [x] ruff fixable count — Task 7
- [x] hadolint singular/plural — Task 1
- [x] pylint json2 + rating — Task 6
- [x] trivy group_by:file — Task 8
- [x] registry.yaml update for pylint — Task 6
- [x] test scripts updated for all changed tools — Tasks 1–8
- [x] showcase docs regenerated — Task 10

**No placeholders:** All code changes show exact Go code with line-level context. All test commands are runnable.

**Type consistency:** `pylintJSON2.Messages` is `[]pylintMsg` — same type used in `formatByRule`/`formatByFile`. `formatByRule` signature change to add `score float64` is consistent with the call in `format`.
