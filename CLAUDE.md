# prettyout — Claude Instructions

## Project Overview

prettyout intercepts CLI linters/scanners, runs them with JSON output, and pipes the JSON through a formatter plugin that prints a grouped, colorized human-readable summary. Architecture: `docs/architecture.md`. Roadmap: `ROADMAP.md`.

---

## Plugin Testing Protocol

**Every new or modified plugin MUST be tested against real tool output before committing.**

Do NOT write plugins based on documentation alone and skip testing. JSON formats have undocumented edge cases, null fields, and tool-version quirks that only show up with real output.

### Step 1 — Research: capture real JSON

Create temporary test files (do not commit them) that cover these cases:

| Case | What to create |
|------|----------------|
| Clean | File with no issues at all |
| Warnings only | Issues at warning severity |
| Errors only | Issues at error severity |
| Mixed | Both errors and warnings in same run |
| Multiple files | Issues spread across ≥2 files |
| Multiple issues same line | Two rules triggering on the same line |
| Syntax error | File that can't be parsed by the tool |
| Empty file | Zero-byte or whitespace-only file |

Run with the tool's JSON flag and save the raw output:

```sh
<tool> <json-flag> <files> 2>/tmp/tool-stderr.json > /tmp/tool-output.json
cat /tmp/tool-output.json
```

Study the actual JSON. Document in the spec:
- Top-level shape (array vs object)
- Fields used: rule code, message, file path, line number, severity
- Which fields can be null or missing
- Whether file paths are absolute or relative
- Whether it's standard JSON or NDJSON (one object per line)
- Exit codes (0 = clean? 1 = found issues?)

### Step 2 — Implement the plugin

Write `cmd/prettyout-<tool>/main.go` using `formatter.RunWithConfig`. See existing plugins as reference.

Key helpers in `pkg/formatter`:
- `ResolvePath(path, cfg)` — relative path from CWD; respects `basename_only` config
- `SeverityColor(sev, colors)` — ANSI color for severity string
- `ParseNDJSON(data)` — line-by-line JSON parser for tools like mypy, cargo-clippy

### Step 3 — Test: pipe real JSON through the plugin

Build and test against every case from Step 1:

```sh
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
go build -o /tmp/prettyout-<tool> ./cmd/prettyout-<tool>/

# Normal run
<tool> <json-flag> <test-files> 2>/dev/null | /tmp/prettyout-<tool>

# Clean run (no issues)
<tool> <json-flag> <clean-files> 2>/dev/null | /tmp/prettyout-<tool>
```

### Step 4 — Verify this checklist

Check every item before committing:

**Output correctness**
- [ ] `group_by: rule` (default): rules sorted, files listed under each rule, lines collapsed per file (`lines 30, 52` not `line 30` + `line 52` separately)
- [ ] `group_by: file`: files sorted, rules+lines listed under each file
- [ ] Occurrence count in rule header: `B904 (8) — message`
- [ ] Summary line at end: `N issues · M rules · K files`

**Paths**
- [ ] Paths show relative from CWD, not absolute and not just basename
- [ ] Test from a directory different from the files' location to verify RelPath logic

**Singular/plural**
- [ ] `line 5` (singular) vs `lines 3, 7` (plural) — never `lines 5` for one line
- [ ] `1 file` vs `2 files` in summary — never `1 files`

**Edge cases**
- [ ] Clean run → `0 issues · 0 rules · 0 files` (or tool-appropriate "nothing found" message)
- [ ] Empty file → no crash, handled gracefully
- [ ] Syntax error in file → shows error correctly, no crash
- [ ] Multiple diagnostics at the same line → line shown once (deduplicated), not repeated
- [ ] Empty stdin → error message to stderr, exit 1 (NDJSON tools: `0 issues` is acceptable)
- [ ] Invalid JSON → error message to stderr, exit 1

**Severity (for tools that report it)**
- [ ] Severity prefix shown in group header: `[ERROR]`, `[WARN]`, `[INFO]`
- [ ] Correct color: red=error, yellow=warning, blue=info
- [ ] Mixed severities in one rule group → show highest severity

**Colors**
- [ ] Colors on by default (when stdout is TTY)
- [ ] `colors: false` in config disables all ANSI codes

### Step 5 — Add registry entry

Add to `internal/registry/builtin.yaml`:
- Correct `output_args` for JSON flag
- `intercept_subcommands` if the tool has subcommands
- `passthrough_flags` for any streaming/watch modes
- `launchers` list

Then run `go test ./...` to confirm registry parses correctly.

---

## Common Bugs Found in This Project

Things that have bitten us before — check these explicitly:

| Bug | Where it appeared | Fix |
|-----|-------------------|-----|
| `lines 5` for single occurrence | cargo-clippy | `if len == 1: "line N"` |
| `1 files` in summary | Multiple plugins | `plural("file", n)` helper |
| Multiline messages shown raw | basedpyright | Take `strings.Split(msg, "\n")[0]` |
| Duplicate line numbers | basedpyright (parse errors) | Use `map[int]struct{}` not `[]int` |
| Ugly `../../..` paths | /tmp symlink on macOS | Real usage is fine; test from project dir |
| Plugins written from docs without testing | All new plugins | Always run Step 1-4 above |

---

## Adding a New Plugin — Checklist

1. [ ] Read the tool's JSON output documentation (if it exists)
2. [ ] Write test files covering all cases in Step 1
3. [ ] Run the tool and save real JSON output
4. [ ] Document the JSON format in the spec (`docs/superpowers/specs/`)
5. [ ] Implement `cmd/prettyout-<tool>/main.go`
6. [ ] Run through the full Step 3-4 checklist above
7. [ ] Add registry entry to `internal/registry/builtin.yaml`
8. [ ] Run `go build ./...` and `go test ./...`
9. [ ] Commit with a message describing what was tested and what edge cases were found
