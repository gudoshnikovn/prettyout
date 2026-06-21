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

Write `cmd/prettyout-<tool>/main.go` using `formatter.RunWithConfig`. 

You no longer need to manually group, filter, or sort data. Simply:
1. Unmarshal your tool's output.
2. Convert it into a `[]formatter.Issue` slice.
3. Call `formatter.Render(issues, cfg)`.

Key helpers in `pkg/formatter`:
- `ResolvePath(path, cfg)` — relative path from CWD; respects `basename_only` config
- `ParseNDJSON(data)` — line-by-line JSON parser for tools like mypy, cargo-clippy
- `Truncate(message, maxLen)` — safely shorten messages

Config fields plugins should respect (set by runtime flags and config file):
All the heavy lifting (`cfg.GroupBy`, `cfg.Sort`, `cfg.OnlyRules`, `cfg.OnlyFiles`, `cfg.Colors`, `cfg.Stats`) is automatically handled inside `formatter.Render()`.

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
- [ ] Ensure your `toIssues` function correctly populates `Rule`, `Message`, `File`, `Line`, and `Severity`.
- [ ] If `Line` is missing or irrelevant, leave it as `0`. The generic render engine will handle it.
- [ ] Map your tool's severity to standard strings (e.g., "error", "warning", "info", "CRITICAL", "HIGH").
- [ ] Ensure you're passing absolute paths or tool paths through `formatter.ResolvePath(path, cfg)`.

**Edge cases**
- [ ] Clean run → `0 issues · 0 rules · 0 files`
- [ ] Empty file → no crash, handled gracefully
- [ ] Syntax error in file → shows error correctly, no crash
- [ ] Invalid JSON → error message to stderr, exit 1

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
| Multiline messages shown raw | basedpyright | Take `strings.Split(msg, "\n")[0]` |
| Duplicate line numbers | basedpyright (parse errors) | Make sure parsing is correct, generic render handles deduplication by file+line |
| Ugly `../../..` paths | /tmp symlink on macOS | Real usage is fine; test from project dir |
| Plugins written from docs without testing | All new plugins | Always run Step 1-4 above |
| Missing severity prefix | semgrep | Provide severity to `formatter.Issue{Severity: ...}`; engine will format it |
| Paths shown raw instead of relative | semgrep | Always pass paths through `ResolvePath(path, cfg)`, not `result.path` directly |
| Wrong field for file path in imports | eslint | `result.filePath` is the top-level key, not a per-message field |
| `location.path` treated as object `{file: string}`, actually plain string in biome 2.x | biome | biome 2.x changed `location.path` from `{file: string}` to a plain string — access `d.Location.Path` directly, not `.file` |
| stylelint writes JSON to stderr, not stdout; paths are absolute | stylelint | Pipe with `2>&1 >/dev/null` so stderr reaches the plugin's stdin; also wrap all source paths with `ResolvePath(f.Source, cfg)` since stylelint emits absolute paths |
| Empty stdout causes JSON parse crash | golangci-lint | golangci-lint exits with code 3 and produces no stdout on infrastructure errors (e.g. Go version mismatch); guard with `if len(strings.TrimSpace(string(data))) == 0` and treat as 0 issues |
| `cp -r` without clearing dest causes stale test fixtures | test/run.sh, Docker | Run `rm -rf dest && cp -r src dest` when copying test fixtures into container |

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
