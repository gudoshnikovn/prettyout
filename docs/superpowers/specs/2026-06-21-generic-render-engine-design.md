# Generic Render Engine — Design Spec

**Date:** 2026-06-21  
**Status:** Approved

## Problem

Every plugin repeats 200–250 lines of identical grouping and rendering logic in `formatByRule` and `formatByFile`. Only the JSON parsing and field extraction differ between plugins. This makes plugins hard to read, new plugins slow to write, and bugs likely to appear in one copy but not others.

## Goal

Extract the shared render engine into `pkg/formatter`. Each plugin shrinks to: parse JSON → map to `[]formatter.Issue` → `formatter.Render(issues, cfg)` → optional plugin-specific extras.

**Target:** plugins go from 200–320 lines → 50–80 lines.

---

## New API: `pkg/formatter/render.go`

### `Issue` type

```go
type Issue struct {
    Rule     string // machine key: used for grouping, filtering, sorting
    Display  string // label shown in output; falls back to Rule if empty
    File     string // already resolved (caller runs ResolvePath before mapping)
    Line     int
    Message  string
    Severity string // "error"/"warning"/"info" variants → SeverityColor/SeverityLabel
    Note     string // optional extra in rule header, e.g. "HIGH/MEDIUM" for bandit
}
```

### Render functions

```go
// Render dispatches to RenderByRule or RenderByFile based on cfg.GroupBy.
func Render(issues []Issue, cfg Config) error

func RenderByRule(issues []Issue, cfg Config) error
func RenderByFile(issues []Issue, cfg Config) error
```

---

## Render Behavior

### `RenderByRule`

1. **Build buckets:** `map[rule] → {display, message, severity, note, map[file] → set[line]}`.  
   Line deduplication via `map[int]struct{}` — always on, no per-plugin logic needed.

2. **Count + filter + sort:**
   ```
   ruleCounts[rule] = total lines across all files
   ruleOrder = FilterRuleOrder(ruleOrder, cfg.OnlyRules)
   ruleOrder = SortOrder(ruleOrder, ruleCounts, cfg.Sort)
   ```

3. **Stats mode:** if `cfg.Stats` → call `PrintStats`, return.

4. **Print each rule:**
   - Skip if no file passes `MatchesFileFilter`
   - Header format:
     ```
     if severity != "":   [SEVERITY] display (note) (N) — message
     if severity == "":   display (note) (N) — message
     ```
     `(note)` is omitted when `Issue.Note` is empty.  
     Color from `SeverityColor(severity, cfg.Colors)`.
   - Print `Affected files:` then sorted files with `FormatLines(lines)`.
   - Print `Divider`.

5. **Summary:** `Summary(displayedIssues, len(ruleOrder), totalFiles)`.

### `RenderByFile`

1. **Build buckets:** `map[file] → []lineEntry{rule, display, line, message}`.  
   Deduplicate `(rule, line)` per file.

2. **Filter + sort:** filter files by `MatchesFileFilter`, sort alphabetically.

3. **Print each file:**
   - Skip if no entry passes `OnlyRules` filter.
   - Header: `file — N issue(s)`
   - Entries sorted by line, then rule. For each: `  rule  line N — message` (message only on first occurrence of each rule).

4. **Summary:** `Summary(totalIssues, distinctRules, len(filteredFiles))`.

---

## Plugin Pattern After Migration

```go
// map tool-specific struct to []formatter.Issue
func toIssues(raw []toolResult, cfg formatter.Config) []formatter.Issue {
    out := make([]formatter.Issue, 0, len(raw))
    for _, r := range raw {
        out = append(out, formatter.Issue{
            Rule:     r.Code,
            File:     formatter.ResolvePath(r.Filename, cfg),
            Line:     r.Location.Row,
            Message:  formatter.Truncate(r.Message, cfg.MaxMessageLength),
            Severity: r.Severity,
        })
    }
    return out
}

func format(data []byte, cfg formatter.Config) error {
    var raw []toolResult
    if err := json.Unmarshal(data, &raw); err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }
    return formatter.Render(toIssues(raw, cfg), cfg)
}
```

Plugins that need post-summary extras call `Render` then print their line:
```go
if err := formatter.Render(issues, cfg); err != nil {
    return err
}
printFixHint(raw) // ruff
// or: fmt.Printf("  ↳ rated %.2f/10\n", score)  // pylint
return nil
```

---

## Per-Plugin Migration Notes

| Plugin | Issue.Display | Issue.Severity | Issue.Note | Post-render extra |
|---|---|---|---|---|
| ruff | Rule | `""` (no label, yellow color) | — | `printFixHint` |
| pylint | `C0301/line-too-long` | mapped from type field | — | `↳ rated N/10` |
| bandit | Rule | `IssueSeverity` (HIGH/MEDIUM/LOW) | `"HIGH/MEDIUM"` | — |
| semgrep | `shortCheckID(CheckID)` | `extra.severity` | — | — |
| basedpyright | Rule | `"error"/"warning"` | — | — |
| mypy | Rule | `"error"/"warning"` | — | — |
| shellcheck | Rule | severity field | — | — |
| golangci | Rule | severity field | — | — |
| eslint | Rule | severity int→string | — | — |
| hadolint | Rule | severity field | — | — |
| cargo-clippy | Rule | `"warning"` | — | — |
| stylelint | Rule | severity field | — | — |
| trivy | Rule | severity field | — | — |
| npm-audit | Rule | severity field | — | — |

**bandit** uses `cleanFilename` before `ResolvePath` — keep that in `toIssues`.  
**basedpyright / mypy / shellcheck** use `ParseNDJSON` — keep that in `format`, mapping stays the same.  
**stylelint** reads from stderr — no change to I/O, only mapping changes.  
**golangci** empty-stdin guard — stays in `format` before calling `toIssues`.

---

## What Gets Deleted

From each plugin after migration:
- Local `ruleEntry` / `ruleGroup` / `fileLines` structs
- `formatByRule` function
- `formatByFile` function
- Local `severityColor` duplicates (bandit, semgrep) — replaced by `formatter.SeverityColor`
- `countDistinctFiles` (ruff)

`pkg/formatter` gains no deletions — only additions.

---

## Files Changed

| File | Change |
|---|---|
| `pkg/formatter/render.go` | **New** — `Issue` type + `Render` / `RenderByRule` / `RenderByFile` (~130 lines) |
| `pkg/formatter/render_test.go` | **New** — table-driven tests for all render paths |
| `cmd/prettyout-*/main.go` | **Shrink** × 14 plugins — remove local render logic, add `toIssues` mapping |

---

## Test Plan for `render.go`

Cover in `render_test.go`:
- `RenderByRule`: zero issues, one rule/one file, multiple rules, multiple files per rule, `OnlyRules` filter, `OnlyFiles` filter, `Sort: count`, `Stats: true`, line deduplication, `Colors: true/false`
- `RenderByFile`: same set
- `Render` dispatch: GroupBy="" → rule, GroupBy="file" → file
- Header formats: severity present, severity empty, Note present, Note empty

Plugin-level test: existing `go test ./...` must still pass after migration.
