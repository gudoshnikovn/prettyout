# Plugin Improvements & Coverage — Design Spec

**Date:** 2026-05-17  
**Branch:** `feature/plugin-coverage`  
**Scope:** Two parallel tracks — (1) fix and improve existing plugins, (2) research and implement new plugins for full ecosystem coverage.

---

## Part 1 — Existing Plugin Improvements

### 1.1 Path Display

**Current:** `filepath.Base()` → shows only `admin.py`. Loses directory context when multiple files share a name.

**New default:** relative path from CWD → `core_api/views/admin.py`.

Implementation: `relPath, err := filepath.Rel(cwd, absPath); if err != nil { relPath = absPath }`. CWD resolved once at startup via `os.Getwd()`.

**Config option** `basename_only: true` under `settings.<tool>` for users who prefer short names. Stored in `cfg.Extra["basename_only"]` (bool).

Affects both `prettyout-ruff` and `prettyout-basedpyright`.

---

### 1.2 group_by: file (not implemented — implement now)

`cfg.GroupBy` is already parsed from config but ignored in both plugins. Two modes:

**`group_by: rule` (default, current behaviour — improved):**
```
B904 (8) — Within an `except` clause, raise exceptions with `raise ... from err`
Affected files:
  - core_api/views/good.py — lines 30, 52, 85, 135, 163, 199
  - core_api/views/dirs.py — line 107
  - core_api/good_filters_src/filters/checkbox.py — line 52
────────────────────────────────────────────────
```
Lines for the same file are now **collapsed on one row** (comma-separated), not repeated as separate lines. This fixes the current behaviour where `good.py` appears 6 times under B904.

**`group_by: file`:**
```
core_api/views/good.py — 6 issues
  B904  line 30  — Within an `except` clause, raise exceptions with `raise ... from err`
  B904  line 52
  B904  line 85
  B904  line 135
  B904  line 163
  B904  line 199
────────────────────────────────────────────────
core_api/views/admin.py — 3 issues
  B905  line 450  — `zip()` without an explicit `strict=` parameter
  B905  line 493
  B905  line 821
────────────────────────────────────────────────
```
Same rule on consecutive lines: message shown only on first occurrence. Files sorted alphabetically. Issues within a file sorted by line number.

---

### 1.3 basedpyright Severity

basedpyright JSON carries `severity: "error" | "warning" | "information"`. Currently ignored.

**New:** prefix each rule group header with severity indicator:

| Severity    | Color (with colors) | Prefix (no colors) |
|-------------|--------------------|--------------------|
| error       | bold red `[ERROR]` | `[ERROR]`          |
| warning     | bold yellow `[WARN]`| `[WARN]`           |
| information | bold blue `[INFO]` | `[INFO]`           |

Since one rule can have mixed severities (rare but possible), use the **highest severity** seen for the group header. Counts are broken down: `(2 errors, 1 warning)`.

In `group_by: file` mode: show severity per-line instead.

---

### 1.4 Summary Line

Printed after all groups, for both plugins.

```
────────────────────────────────────────────────
29 issues · 12 rules · 8 files
```

Values: total issues (occurrences), distinct rule codes, distinct files. Always printed, no config flag.

---

### 1.5 Occurrence Counts in Rule Header

Already implied in 1.2 examples: `B904 (8)` — number in parens is total occurrences of that rule across all files. Simple to compute while building the rule map.

---

### 1.6 Shared Helper in pkg/formatter

The path-resolution logic (Rel from CWD, basename_only fallback) should live in `pkg/formatter` as `ResolvePath(absPath string, cfg Config) string` so both plugins (and future ones) share it identically.

---

## Part 2 — New Plugins

### 2.1 Research Methodology

For each tool:
1. Write minimal test files covering: clean (no issues), one error, one warning, mixed errors+warnings, edge case (empty file, no newline at EOF, unicode, very long message).
2. Run with JSON output flag.
3. Document actual JSON structure observed.
4. Note corner cases: missing fields, null values, empty arrays, tools that exit non-zero even on success with `--json`.
5. Test files are **not committed** — they are temporary research artifacts. Results are captured in section 2.3 below.

### 2.2 Output Flag and Registry Entry Template

Each new tool needs:
- A registry entry in `internal/registry/builtin.yaml`
- A `cmd/prettyout-<tool>/main.go` binary using `formatter.RunWithConfig`

Registry template:
```yaml
tools:
  <toolname>:
    plugin: prettyout-<toolname>
    intercept_subcommands: [<cmd>]   # omit if all subcommands intercepted
    output_args: [<json-flag>]
    passthrough_flags: []
    launchers: [<launcher>, ...]
    install:
      go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-<toolname>
```

### 2.3 Tool-by-Tool Specifications

---

#### golangci-lint

**JSON flag:** `--out-format json`  
**Launcher:** direct binary (no uvx/pipx)  
**Intercept subcommands:** `run` (only subcommand that lints)

**JSON format:**
```json
{
  "Issues": [
    {
      "FromLinter": "errcheck",
      "Text": "Error return value of `os.Remove` is not checked",
      "Severity": "",
      "Pos": {
        "Filename": "main.go",
        "Line": 10,
        "Column": 5
      },
      "SourceLines": ["  os.Remove(tmpFile)"]
    }
  ],
  "Report": { "Warnings": [], "Linters": [] }
}
```

**Corner cases:**
- `Severity` field is often empty string — treat as "error" by default
- `Pos.Filename` is relative to module root, not absolute — use as-is (no Rel() needed)
- `SourceLines` included but not used in our output
- Issues with no linter name: use `"unknown"` as rule code
- Can return `{"Issues": null}` when no issues (not empty array) — handle null

**Plugin grouping:** by `FromLinter` (linter name acts as rule). Message from `Text`. File from `Pos.Filename`. Line from `Pos.Line`.

**Color:** bold cyan for linter name.

**Registry:**
```yaml
golangci-lint:
  plugin: prettyout-golangci
  intercept_subcommands: [run]
  output_args: [--out-format=json]
  launchers: []
  install:
    go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-golangci
```

---

#### eslint

**JSON flag:** `--format json`  
**Launchers:** npx  
**Intercept subcommands:** none (eslint has no subcommands; all invocations intercepted)

**JSON format:**
```json
[
  {
    "filePath": "/abs/path/to/file.ts",
    "messages": [
      {
        "ruleId": "no-unused-vars",
        "severity": 2,
        "message": "'x' is defined but never used.",
        "line": 1,
        "column": 5,
        "endLine": 1,
        "endColumn": 6
      }
    ],
    "errorCount": 1,
    "warningCount": 0
  }
]
```

**Corner cases:**
- `severity`: 1 = warning, 2 = error. `ruleId` can be null for parser errors — use `"parse-error"`.
- `filePath` is absolute — apply `filepath.Rel(cwd, filePath)`.
- File entry with empty `messages` array means file was checked but clean — skip it.
- Fatal parse errors: `message.fatal = true`, `ruleId = null`.
- eslint exits with code 1 on any lint error — normal, don't treat as failure.

**Plugin grouping:** default `group_by: rule` uses `ruleId`. File mode uses `filePath`. Severity from `severity` field (map 1→warning, 2→error).

**Color:** bold magenta for rule name.

**Registry:**
```yaml
eslint:
  plugin: prettyout-eslint
  output_args: [--format=json]
  launchers: [npx]
  install:
    go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-eslint
```

---

#### biome

**JSON flag:** `--reporter=json`  
**Launchers:** npx  
**Intercept subcommands:** `check`, `lint`, `ci`

**JSON format:**
```json
{
  "summary": {
    "changed": 0,
    "unchanged": 1,
    "duration": {"secs": 0, "nanos": 123},
    "errors": 1,
    "warnings": 0
  },
  "diagnostics": [
    {
      "category": "lint/suspicious/noDoubleEquals",
      "description": "Use === instead of ==.",
      "severity": "error",
      "location": {
        "path": {"file": "src/index.ts"},
        "span": [10, 20],
        "sourceCode": "a == b"
      }
    }
  ]
}
```

**Corner cases:**
- `location.path.file` is relative — use as-is.
- `span` is byte offsets, not line numbers — biome does not provide line numbers in JSON output; omit line info, show only filename.
- `category` is the rule identifier (e.g. `lint/suspicious/noDoubleEquals`); display as-is.
- `severity`: `"error"` | `"warning"` | `"information"` | `"hint"`.

**Registry:**
```yaml
biome:
  plugin: prettyout-biome
  intercept_subcommands: [check, lint, ci]
  output_args: [--reporter=json]
  launchers: [npx]
  install:
    go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-biome
```

---

#### mypy

**JSON flag:** `--output=json` (mypy ≥ 1.4)  
**Launchers:** uvx, pipx  
**Intercept subcommands:** none

**JSON format:** one JSON object per line (NDJSON), not an array:
```json
{"file": "foo.py", "line": 1, "column": 1, "severity": "error", "message": "Incompatible types", "code": "assignment", "hint": null}
{"file": "foo.py", "line": 5, "column": 3, "severity": "note", "message": "Expected type 'int'", "code": null, "hint": null}
```

**Corner cases:**
- Parse line-by-line, not as a single JSON value.
- `severity`: `"error"` | `"warning"` | `"note"`. Notes are supplemental context for the preceding error — they have `code: null`. In output, `note` lines should be attached to their parent error, not shown as a separate rule group.
- `code` can be null for notes and some internal errors — use `"error"` as fallback rule name.
- `file` is relative from project root — use as-is.
- mypy exits 0 if no errors, 1 if errors found — handle both.

**Plugin grouping:** group by `code`. Notes (severity=note) are skipped in grouped view or shown indented under their parent. Severity from `severity` field.

**Color:** bold red for errors, yellow for warnings.

**Registry:**
```yaml
mypy:
  plugin: prettyout-mypy
  output_args: [--output=json]
  launchers: [uvx, pipx]
  install:
    go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-mypy
```

---

#### bandit

**JSON flag:** `-f json`  
**Launchers:** uvx, pipx  
**Intercept subcommands:** none

**JSON format:**
```json
{
  "errors": [],
  "metrics": {"_totals": {"CONFIDENCE.HIGH": 1}},
  "results": [
    {
      "filename": "./test.py",
      "line_number": 5,
      "col_offset": 0,
      "test_id": "B324",
      "test_name": "hashlib",
      "issue_text": "Use of insecure MD2, MD4, MD5, or SHA1 hash function.",
      "issue_severity": "MEDIUM",
      "issue_confidence": "HIGH",
      "line_range": [5],
      "more_info": "https://..."
    }
  ]
}
```

**Corner cases:**
- `filename` may have leading `./` — strip with `strings.TrimPrefix(f, "./")`.
- `issue_severity`: `"LOW"` | `"MEDIUM"` | `"HIGH"`. Use as severity indicator.
- `issue_confidence`: `"LOW"` | `"MEDIUM"` | `"HIGH"`. Show in output alongside severity: `B324 (MEDIUM/HIGH)`.
- `errors` array: if non-empty, files that couldn't be parsed. Log to stderr.
- bandit exits 0 if no issues, 1 if issues found — handle both.

**Plugin grouping:** group by `test_id`. Message from `issue_text`. File from `filename`. Line from `line_number`.

**Color:** bold red for HIGH, yellow for MEDIUM, white for LOW.

**Registry:**
```yaml
bandit:
  plugin: prettyout-bandit
  output_args: [-f, json]
  launchers: [uvx, pipx]
  install:
    go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-bandit
```

---

#### pylint

**JSON flag:** `--output-format=json`  
**Launchers:** uvx, pipx  
**Intercept subcommands:** none

**JSON format:**
```json
[
  {
    "type": "error",
    "module": "mymodule.views",
    "obj": "MyView.get",
    "line": 42,
    "column": 4,
    "endLine": 42,
    "endColumn": 10,
    "path": "mymodule/views.py",
    "symbol": "undefined-variable",
    "message": "Undefined variable 'foo'",
    "message-id": "E0602"
  }
]
```

**Corner cases:**
- `type`: `"error"` | `"warning"` | `"convention"` | `"refactor"` | `"fatal"`. Map to severity.
- `symbol` is the human-readable rule name; `message-id` is the code (e.g. `E0602`). Show both: `E0602/undefined-variable`.
- `path` is relative — use as-is.
- pylint exits with a bitmask exit code (0=no errors, various bits for different issue types) — any exit code is acceptable.
- Can output an empty array `[]` — handle gracefully.

**Plugin grouping:** group by `message-id`. Display as `E0602/undefined-variable`.

**Color:** bold red for errors/fatal, yellow for warnings, dim for convention/refactor.

**Registry:**
```yaml
pylint:
  plugin: prettyout-pylint
  output_args: [--output-format=json]
  launchers: [uvx, pipx]
  install:
    go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-pylint
```

---

#### shellcheck

**JSON flag:** `--format json`  
**Launchers:** none (system binary)  
**Intercept subcommands:** none

**JSON format:**
```json
[
  {
    "file": "script.sh",
    "line": 3,
    "endLine": 3,
    "column": 5,
    "endColumn": 10,
    "level": "warning",
    "code": 2006,
    "message": "Use $(...) notation instead of legacy backtick `...`."
  }
]
```

**Corner cases:**
- `code` is an integer — format as `SC2006`.
- `level`: `"error"` | `"warning"` | `"info"` | `"style"`.
- `file` is relative — use as-is.
- shellcheck exits 0 if no issues, 1 if issues found, 2 on usage error.
- Can output `[]` with exit 0 on clean files.

**Plugin grouping:** group by formatted code `SC{code}`.

**Color:** bold red for errors, yellow for warnings, dim for info/style.

**Registry:**
```yaml
shellcheck:
  plugin: prettyout-shellcheck
  output_args: [--format=json]
  launchers: []
  install:
    go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-shellcheck
```

---

#### hadolint

**JSON flag:** `--format json`  
**Launchers:** none (system binary or docker)  
**Intercept subcommands:** none

**JSON format:**
```json
[
  {
    "line": 3,
    "column": 1,
    "level": "warning",
    "code": "DL3008",
    "message": "Pin versions in apt get install.",
    "file": "Dockerfile"
  }
]
```

**Corner cases:**
- `file` is the filename as given on the command line — could be relative or basename only.
- `level`: `"error"` | `"warning"` | `"info"` | `"style"` | `"ignore"`. Skip `"ignore"` level items.
- `code`: `DL####` for Dockerfile rules, `SC####` for embedded shellcheck rules.
- hadolint exits 0 if no issues, 1 otherwise.

**Plugin grouping:** group by `code`.

**Registry:**
```yaml
hadolint:
  plugin: prettyout-hadolint
  output_args: [--format=json]
  launchers: []
  install:
    go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-hadolint
```

---

#### cargo-clippy

**JSON flag:** `--message-format json`  
**Launchers:** none (cargo subcommand)  
**Intercept subcommands:** `clippy` (i.e. intercept `cargo clippy`)

**JSON format:** NDJSON — one JSON object per line with a `reason` field:
```json
{"reason":"compiler-message","package_id":"foo 0.1.0 (path+file:///...)","manifest_path":"...","message":{"rendered":"warning[clippy::needless_return]: ...\n","code":{"code":"clippy::needless_return","explanation":null},"level":"warning","message":"needless return","spans":[{"file_name":"src/main.rs","line_start":5,"column_start":5,"line_end":5,"column_end":18,"is_primary":true}]}}
{"reason":"build-finished","success":true}
```

**Corner cases:**
- Process only objects where `reason == "compiler-message"`.
- `message.code` can be null for non-clippy compiler messages (e.g. `unused import`) — use `message.level` + first word of `message.message` as fallback key, or use `"rustc"` as the rule.
- `message.spans`: use the span where `is_primary == true` for file/line. Multiple spans possible.
- `message.level`: `"warning"` | `"error"` | `"note"` | `"help"`. Skip `note` and `help` as they are supplemental.
- Tool is `cargo clippy` — intercept `cargo` when subcommand is `clippy`.

**Registry:**
```yaml
cargo:
  plugin: prettyout-cargo-clippy
  intercept_subcommands: [clippy]
  output_args: [--message-format=json]
  launchers: []
  install:
    go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-cargo-clippy
```

---

#### semgrep

**JSON flag:** `--json`  
**Launchers:** none (system binary)  
**Intercept subcommands:** `scan`, `ci` (not `login`, `publish`, etc.)

**JSON format:**
```json
{
  "results": [
    {
      "check_id": "python.lang.security.audit.formatted-sql-query",
      "path": "app/db.py",
      "start": {"line": 10, "col": 5, "offset": 200},
      "end": {"line": 10, "col": 30, "offset": 225},
      "extra": {
        "message": "Detected a formatted string in a SQL statement.",
        "severity": "ERROR",
        "metadata": {"confidence": "HIGH"}
      }
    }
  ],
  "errors": [],
  "stats": {"total_time": 1.2}
}
```

**Corner cases:**
- `check_id` is the rule identifier (namespaced: `python.lang.security.audit.xxx`). Long — truncate display to last segment for header, show full in verbose mode.
- `extra.severity`: `"ERROR"` | `"WARNING"` | `"INFO"`.
- `path` is relative — use as-is.
- `errors` array: semgrep parse errors or rule errors — log to stderr if non-empty.
- semgrep exits 0 if no findings, 1 if findings, other codes for errors.

**Registry:**
```yaml
semgrep:
  plugin: prettyout-semgrep
  intercept_subcommands: [scan, ci]
  output_args: [--json]
  launchers: []
  install:
    go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-semgrep
```

---

#### trivy

**JSON flag:** `--format json`  
**Launchers:** none (system binary)  
**Intercept subcommands:** `image`, `fs`, `repo`, `config`

**JSON format:**
```json
{
  "SchemaVersion": 2,
  "Results": [
    {
      "Target": "ubuntu:22.04 (ubuntu 22.04)",
      "Type": "ubuntu",
      "Vulnerabilities": [
        {
          "VulnerabilityID": "CVE-2023-1234",
          "PkgName": "openssl",
          "InstalledVersion": "1.1.1t-1",
          "FixedVersion": "1.1.1u-1",
          "Severity": "HIGH",
          "Title": "OpenSSL: buffer overflow"
        }
      ],
      "Misconfigurations": null
    }
  ]
}
```

**Corner cases:**
- trivy is a **security scanner**, not a linter. No "file/line" — group by `Severity` then `PkgName`.
- `Vulnerabilities` can be null (not empty array) when no vulns — handle null.
- `Misconfigurations` has a different structure: `{ID, AVDID, Type, Title, Description, Message, Severity, Status}`.
- `Severity`: `"CRITICAL"` | `"HIGH"` | `"MEDIUM"` | `"LOW"` | `"UNKNOWN"`.
- Output format differs from linter plugins — no "file/line" model. Show: severity group → CVE ID → package → installed vs fixed version.
- trivy exits 0 if no vulns, 1 if vulns found.

**Output format example:**
```
HIGH (3 vulnerabilities)
  CVE-2023-1234 — openssl 1.1.1t → fix: 1.1.1u
  CVE-2023-5678 — curl 7.81.0 → fix: 7.88.0
────────────────────────────────────────────────
MEDIUM (1 vulnerability)
  CVE-2022-9999 — zlib 1.2.11 → no fix available
```

**Registry:**
```yaml
trivy:
  plugin: prettyout-trivy
  intercept_subcommands: [image, fs, repo, config]
  output_args: [--format=json]
  launchers: []
  install:
    go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-trivy
```

---

#### npm audit

**JSON flag:** `--json`  
**Launchers:** npx (but typically run directly as `npm audit`)  
**Intercept subcommands:** `audit` (intercept `npm` when subcommand is `audit`)

**JSON format (npm audit report v2):**
```json
{
  "auditReportVersion": 2,
  "vulnerabilities": {
    "lodash": {
      "name": "lodash",
      "severity": "high",
      "isDirect": false,
      "via": ["CVE-2021-23337"],
      "effects": ["my-app"],
      "range": ">=0.0.0 <4.17.21",
      "nodes": ["node_modules/lodash"],
      "fixAvailable": true
    }
  },
  "metadata": {
    "vulnerabilities": {"info": 0, "low": 0, "moderate": 1, "high": 2, "critical": 0, "total": 3}
  }
}
```

**Corner cases:**
- Key in `vulnerabilities` map is the package name — use it as the identifier.
- `severity`: lowercase strings `"info"` | `"low"` | `"moderate"` | `"high"` | `"critical"`.
- `via` can be strings (CVE IDs) or objects (nested vulnerability info) — handle both.
- `fixAvailable`: bool or object with `{name, version, isSemVerMajor}`. Show "fix available" or "breaking fix" or "no fix".
- No file/line concept — group by severity, show package name + range + fix status.
- Tool is `npm audit` — intercept `npm` when subcommand is `audit`.
- npm exits 0 if no vulns, non-zero if vulns found.

**Registry:**
```yaml
npm:
  plugin: prettyout-npm-audit
  intercept_subcommands: [audit]
  output_args: [--json]
  launchers: []
  install:
    go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-npm-audit
```

---

#### stylelint

**JSON flag:** `--formatter json`  
**Launchers:** npx  
**Intercept subcommands:** none

**JSON format:**
```json
[
  {
    "source": "src/styles/main.css",
    "deprecations": [],
    "invalidOptionWarnings": [],
    "parseErrors": [],
    "errored": true,
    "warnings": [
      {
        "line": 5,
        "column": 3,
        "endLine": 5,
        "endColumn": 10,
        "rule": "color-no-invalid-hex",
        "severity": "error",
        "text": "Unexpected invalid hex color \"#gggggg\" (color-no-invalid-hex)",
        "url": "https://stylelint.io/user-guide/rules/color-no-invalid-hex"
      }
    ]
  }
]
```

**Corner cases:**
- `severity`: `"error"` | `"warning"`.
- `source` is relative — use as-is.
- `warnings` array holds both errors and warnings (misleading name).
- `parseErrors` array: if non-empty, CSS couldn't be parsed — show as `parse-error` rule.
- stylelint exits 0 if no issues, 2 if lint errors found, 1 on config error.

**Registry:**
```yaml
stylelint:
  plugin: prettyout-stylelint
  output_args: [--formatter=json]
  launchers: [npx]
  install:
    go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-stylelint
```

---

## Part 3 — Infrastructure Notes

### 3.1 pkg/formatter additions

Add to `pkg/formatter`:
- `ResolvePath(path string, cfg Config) string` — if path is absolute: compute relative from CWD, fall back to absPath on error. If path is already relative: use as-is. If `basename_only` extra config key is true: return `filepath.Base(path)` regardless.
- `SeverityColor(sev string, colors bool) string` — maps `"error"/"warning"/"information"/"note"` to ANSI sequences. Shared across plugins that use severity.

### 3.2 builtin.yaml additions

Add all 11 new tools (golangci-lint, eslint, biome, mypy, bandit, pylint, shellcheck, hadolint, cargo, semgrep, trivy, npm, stylelint) to `internal/registry/builtin.yaml`.

### 3.3 NDJSON parser helper

mypy and cargo-clippy use NDJSON (one JSON object per line). Add `formatter.ParseNDJSON(data []byte) ([]json.RawMessage, error)` to `pkg/formatter` for reuse.

---

## Execution Plan

1. **Improve existing plugins** — implement all of Part 1 in `prettyout-ruff` and `prettyout-basedpyright`, plus `pkg/formatter` helpers.
2. **New plugins** — implement all plugins in Part 2, add registry entries.
3. Both tracks can proceed in parallel.
