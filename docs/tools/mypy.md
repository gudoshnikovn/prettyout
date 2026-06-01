# mypy

**What it checks:** Static type checker for Python.

## Without prettyout

### Default output
```
errors.py:1: error: Incompatible types in assignment (expression has type "str", variable has type "int")  [assignment]
errors.py:2: error: Unsupported operand types for + ("int" and "str")  [operator]
errors.py:3: error: Incompatible types in assignment (expression has type "str", variable has type "int")  [assignment]
Found 3 errors in 1 file (checked 1 source file)
```

### JSON (what CI/CD sees)
```json
{"file": "errors.py", "line": 1, "column": 9, "end_line": 1, "end_column": 21, "message": "Incompatible types in assignment (expression has type \"str\", variable has type \"int\")", "hint": null, "code": "assignment", "severity": "error"}
{"file": "errors.py", "line": 2, "column": 13, "end_line": 2, "end_column": 20, "message": "Unsupported operand types for + (\"int\" and \"str\")", "hint": null, "code": "operator", "severity": "error"}
{"file": "errors.py", "line": 3, "column": 9, "end_line": 3, "end_column": 19, "message": "Incompatible types in assignment (expression has type \"str\", variable has type \"int\")", "hint": null, "code": "assignment", "severity": "error"}
```

## With prettyout

### Group by rule (default)
```
[ERROR] assignment (2) — Incompatible types in assignment (expression has type "str", variable has type "int")
Affected files:
  - errors.py — lines 1, 3
────────────────────────────────────────────────
[ERROR] operator (1) — Unsupported operand types for + ("int" and "str")
Affected files:
  - errors.py — line 2
────────────────────────────────────────────────
3 issues · 2 rules · 1 file
```

### Group by file
```
errors.py — 3 issues
  assignment  line 1 — Incompatible types in assignment (expression has type "str", variable has type "int")
  operator  line 2 — Unsupported operand types for + ("int" and "str")
  assignment  line 3 — Incompatible types in assignment (expression has type "str", variable has type "int")
────────────────────────────────────────────────
3 issues · 2 rules · 1 file
```

## What prettyout improves

- **Consistent format**: mypy has its own output style → prettyout produces the same structure as all other supported tools
- **Rule grouping**: raw output lists violations line-by-line → group-by-rule shows all files affected by the same error code
- **Occurrence counts**: no summary in raw output → `error-code (4) — message` shows how widespread each issue is
- **Note clutter removed**: mypy emits `note`-level messages as context lines → prettyout strips them, showing only actionable errors and warnings
- **Summary line**: `Found N errors` is hard to scan → `N issues · M rules · K files` at the end is consistent with all other tools
