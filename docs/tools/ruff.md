# ruff

**What it checks:** Python linter for style errors, unused imports, and code quality.

## Without prettyout

### Default output
```
F401 [*] `os` imported but unused
 --> errors.py:1:8
  |
1 | import os
  |        ^^
2 | import sys
  |
help: Remove unused import: `os`

F401 [*] `sys` imported but unused
 --> errors.py:2:8
  |
1 | import os
2 | import sys
  |        ^^^
3 |
4 | def foo():
  |
help: Remove unused import: `sys`

Found 2 errors.
[*] 2 fixable with the `--fix` option.
```

### JSON (what CI/CD sees)
```json
[
  {
    "cell": null,
    "code": "F401",
    "end_location": {
      "column": 10,
      "row": 1
    },
    "filename": "/tmp/t-ruff/errors.py",
    "fix": {
      "applicability": "safe",
      "edits": [
        {
          "content": "",
          "end_location": {
            "column": 1,
            "row": 2
          },
          "location": {
            "column": 1,
            "row": 1
          }
        }
      ],
      "message": "Remove unused import: `os`"
    },
    "location": {
      "column": 8,
      "row": 1
    },
```

## With prettyout

### Group by rule (default)
```
F401 (2) — `os` imported but unused
Affected files:
  - errors.py — lines 1, 2
────────────────────────────────────────────────
2 issues · 1 rule · 1 file
```

### Group by file
```
errors.py — 2 issues
  F401  line 1  — `os` imported but unused
  F401  line 2
────────────────────────────────────────────────
2 issues · 1 rule · 1 file
```

## What prettyout improves

- **Consistent format**: ruff has its own output style → prettyout produces the same structure as all other supported tools
- **Rule grouping**: raw output lists violations line-by-line → group-by-rule view shows all affected files per rule at a glance
- **Occurrence counts**: no summary in raw output → `E501 (4) — message` shows how widespread each rule is
- **Summary line**: must scroll to the end to see counts → `N issues · M rules · K files` always at the end
