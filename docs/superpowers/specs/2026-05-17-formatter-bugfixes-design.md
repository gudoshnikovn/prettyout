# Formatter Bug Fixes Design

**Goal:** Fix 5 bugs across multiple formatter plugins and update showcase docs.

---

## Fix 1: Severity prefix for 9 tools

Add `[ERROR]` / `[WARN]` / `[INFO]` text prefix to rule group headers in all tools that have severity in their JSON. Currently severity is only used for ANSI color — this adds a visible text label for non-color terminals and clarity.

**Format:** `[ERROR] rule-code (N) — message`

**Affected files:** `cmd/prettyout-mypy`, `cmd/prettyout-eslint`, `cmd/prettyout-stylelint`, `cmd/prettyout-shellcheck`, `cmd/prettyout-hadolint`, `cmd/prettyout-cargo-clippy`, `cmd/prettyout-biome`, `cmd/prettyout-semgrep`, `cmd/prettyout-pylint`

**Normalization per tool:**

| Tool | → ERROR | → WARN | → INFO |
|------|---------|--------|--------|
| mypy | `"error"` | `"warning"` | — |
| eslint | `severity == 2` | `severity == 1` | — |
| stylelint | `"error"` | `"warning"` | — |
| shellcheck | `"error"` | `"warning"` | `"info"`, `"style"` |
| hadolint | `"error"` | `"warning"` | `"info"` |
| cargo clippy | `"error"` | `"warning"` | `"note"`, `"help"` |
| biome | `"error"` | `"warning"` | — |
| semgrep | `"ERROR"` | `"WARNING"` | `"INFO"` |
| pylint | `"error"`, `"fatal"` | `"warning"` | `"convention"`, `"refactor"` |

**Output in group-by-rule header (no colors):**
```
[ERROR] E0602/undefined-variable (1) — Undefined variable 'undefined_var'
```

**Output with colors:** severity bracket uses same ANSI code as the existing color, rest is normal.

**group-by-file:** severity prefix shown per-line entry: `  [WARN] no-unused-vars  line 3 — message`

---

## Fix 2: ruff fixable count

**File:** `cmd/prettyout-ruff/main.go`

The ruff JSON `fix` field is non-null when an issue is auto-fixable. Count these and show after summary.

**JSON field to parse:**
```go
type issue struct {
    // existing fields...
    Fix *struct{} `json:"fix"` // non-null = fixable
}
```

**Output (only shown when fixable count > 0):**
```
2 issues · 1 rule · 1 file
  ↳ 2 fixable with --fix
```

---

## Fix 3: hadolint singular/plural

**File:** `cmd/prettyout-hadolint/main.go`

Replace hardcoded `"lines %s"` with proper singular/plural check.

**Current (broken):**
```go
fmt.Printf("  - %s — lines %s\n", f, strings.Join(lineStrs, ", "))
```

**Fixed:**
```go
lineWord := formatter.Plural(len(ls), "line", "lines")
fmt.Printf("  - %s — %s %s\n", f, lineWord, strings.Join(lineStrs, ", "))
```

---

## Fix 4: pylint json2 + rating

**Files:** `cmd/prettyout-pylint/main.go`, `internal/registry/builtin.yaml`

Switch from `--output-format=json` to `--output-format=json2`. The json2 format wraps messages in an object:

```json
{
  "messages": [
    {"type": "error", "module": "errors", "path": "errors.py", "line": 2, ...}
  ],
  "statistics": {
    "score": 0.0,
    "refactor": 0, "convention": 1, "warning": 1, "error": 1, "fatal": 0
  }
}
```

**Parser change:** unmarshal into `struct { Messages []pylintMsg; Statistics struct { Score float64 } }` instead of `[]pylintMsg`.

**Output after summary:**
```
3 issues · 3 rules · 1 file
  ↳ rated 0.00/10
```

Always shown (even `10.00/10` — users want to know).

**Registry change in `internal/registry/builtin.yaml`:**
```yaml
# Before:
output_args: ["--output-format=json"]
# After:
output_args: ["--output-format=json2"]
```

---

## Fix 5: trivy group_by: file

**File:** `cmd/prettyout-trivy/main.go`

Currently `cfg.GroupBy` is never read — both modes produce identical output.

**group_by: rule (default, current behavior):** group all vulns by severity across all targets. No change.

**group_by: file:** group by `Target` (e.g. `requirements.txt`, `Cargo.lock`). Within each target, show vulns sorted by severity rank then CVE ID. Summary: `N vulnerabilities · M targets`.

```
requirements.txt
  CRITICAL  CVE-2019-19844 — django 2.0 → fix: 1.11.27, 2.2.9, 3.0.1
  CRITICAL  CVE-2020-7471 — django 2.0 → fix: 1.11.28, 2.2.10, 3.0.3
  HIGH      CVE-2018-6188 — django 2.0 → fix: 2.0.2, 1.11.10
  ...
────────────────────────────────────────────────
16 vulnerabilities · 1 target
```

---

## Showcase doc updates

After all formatter fixes, regenerate these per-tool docs using the same Docker capture process:
- `docs/tools/ruff.md` (fixable line)
- `docs/tools/hadolint.md` (singular fix + severity prefix)
- `docs/tools/pylint.md` (rating + severity prefix)
- `docs/tools/mypy.md` (severity prefix)
- `docs/tools/eslint.md` (severity prefix)
- `docs/tools/stylelint.md` (severity prefix)
- `docs/tools/shellcheck.md` (severity prefix)
- `docs/tools/cargo-clippy.md` (severity prefix)
- `docs/tools/biome.md` (severity prefix)
- `docs/tools/semgrep.md` (severity prefix)
- `docs/tools/trivy.md` (group_by: file now different)

Also update `docs/summary.md` comparison table to reflect mypy now having severity prefix.

---

## Testing

Each formatter fix must be verified by running the real tool through Docker:
```bash
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  cd /tmp && cp -r /project/test/fixtures/<tool> t && cd t
  <tool> <json-flag> errors.* 2>/dev/null | /project/<binary>
"
```

Plus `go build ./...` and `go test ./...` after all changes.
