# golangci-lint

**What it checks:** Go meta-linter — runs multiple Go linters in parallel.

## Without prettyout

### Default output

```
main.go:6:2: ineffectual assignment to x (ineffassign)
	x := 42
	^
```

### JSON (what CI/CD sees)

```json
{
  "Issues": [
    {
      "FromLinter": "ineffassign",
      "Text": "ineffectual assignment to x",
      "Severity": "",
      "SourceLines": ["\tx := 42"],
      "Replacement": null,
      "Pos": {
        "Filename": "main.go",
        "Offset": 43,
        "Line": 6,
        "Column": 2
      },
      "ExpectNoLint": false,
      "ExpectedNoLintLinter": ""
    }
  ],
  "Report": {
    "Linters": [
      {"Name": "asasalint"},
      {"Name": "asciicheck"},
      {"Name": "bidichk"},
      ...
    ]
  }
}
```

## With prettyout

### Group by rule (default)

```
ineffassign (1) — ineffectual assignment to x
Affected files:
  - main.go — lines 6
────────────────────────────────────────────────
1 issue · 1 rule · 1 file
```

### Group by file

```
main.go — 1 issue
  ineffassign  line 6 — ineffectual assignment to x
────────────────────────────────────────────────
1 issue · 1 rule · 1 file
```

## What prettyout improves

- **Consistent format**: golangci-lint has its own output style → prettyout produces the same structure as all other supported tools
- **Empty run handled gracefully**: golangci-lint emits `"Issues": null` (not `[]`) on a clean run → prettyout handles this and shows `0 issues · 0 rules · 0 files`
- **Rule grouping**: raw output lists issues line-by-line → group-by-rule shows all files affected by the same linter check
- **Occurrence counts**: no count in rule header → `ineffassign (2) — message` shows how widespread each issue is
- **Summary line**: raw output has no summary → `N issues · M rules · K files` at the end
