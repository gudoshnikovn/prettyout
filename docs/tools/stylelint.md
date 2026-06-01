# stylelint

**What it checks:** CSS/SCSS/Less linter for style errors and best practices.

## Without prettyout

### Default output
```
(no stdout — stylelint writes formatted output only when not using --formatter=json)
```

### JSON (what CI/CD sees, on stderr)
```json
[{"source":"/tmp/t-stylelint/errors.css","deprecations":[],"invalidOptionWarnings":[],
"parseErrors":[],"errored":true,"warnings":[{"line":1,"column":1,"endLine":2,
"endColumn":2,"rule":"font-family-no-unknown-names","severity":"error",
"text":"Unknown rule font-family-no-unknown-names."},{"line":1,"column":12,"endLine":1,
"endColumn":19,"rule":"color-no-invalid-hex","severity":"error",
"text":"Invalid hex color \"#gggggg\" (color-no-invalid-hex)"}]}]
```

## With prettyout

### Group by rule (default)
```
[ERROR] color-no-invalid-hex (1) — Invalid hex color "#gggggg" (color-no-invalid-hex)
Affected files:
  - errors.css — line 1
────────────────────────────────────────────────
[ERROR] font-family-no-unknown-names (1) — Unknown rule font-family-no-unknown-names.
Affected files:
  - errors.css — line 1
────────────────────────────────────────────────
2 issues · 2 rules · 1 file
```

### Group by file
```
errors.css — 2 issues
  font-family-no-unknown-names  line 1 — Unknown rule font-family-no-unknown-names.
  color-no-invalid-hex  line 1 — Invalid hex color "#gggggg" (color-no-invalid-hex)
────────────────────────────────────────────────
2 issues · 2 rules · 1 file
```

## What prettyout improves

- **Consistent format**: stylelint has its own output style → prettyout produces the same structure as all other supported tools
- **JSON on stderr handled**: stylelint outputs JSON to stderr, not stdout — most tools can't consume it directly → prettyout handles this transparently via `2>&1 >/dev/null` piping
- **Absolute paths resolved**: stylelint emits absolute file paths in JSON → prettyout converts them to relative paths from the current directory
- **Rule grouping**: raw output lists violations per file → group-by-rule shows all files affected by the same rule
- **Summary line**: raw output has no count summary → `N issues · M rules · K files` at the end
