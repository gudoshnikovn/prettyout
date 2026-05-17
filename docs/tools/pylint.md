# pylint

**What it checks:** Python linter for errors, style, and code smells.

## Without prettyout

### Default output
```
************* Module errors
errors.py:1:0: C0114: Missing module docstring (missing-module-docstring)
errors.py:2:4: E0602: Undefined variable 'undefined_var' (undefined-variable)
errors.py:1:0: W0611: Unused import os (unused-import)

-----------------------------------
Your code has been rated at 0.00/10
```

### JSON (what CI/CD sees)
```json
[
    {
        "type": "convention",
        "module": "errors",
        "obj": "",
        "line": 1,
        "column": 0,
        "endLine": null,
        "endColumn": null,
        "path": "errors.py",
        "symbol": "missing-module-docstring",
        "message": "Missing module docstring",
        "message-id": "C0114"
    },
    {
        "type": "error",
        "module": "errors",
        "obj": "",
        "line": 2,
        "column": 4,
        "endLine": 2,
        "endColumn": 17,
        "path": "errors.py",
        "symbol": "undefined-variable",
        "message": "Undefined variable 'undefined_var'",
        "message-id": "E0602"
    },
    ...
]
```

## With prettyout

### Group by rule (default)
```
C0114/missing-module-docstring (1) — Missing module docstring
Affected files:
  - errors.py — line 1
────────────────────────────────────────────────
E0602/undefined-variable (1) — Undefined variable 'undefined_var'
Affected files:
  - errors.py — line 2
────────────────────────────────────────────────
W0611/unused-import (1) — Unused import os
Affected files:
  - errors.py — line 1
────────────────────────────────────────────────
3 issues · 3 rules · 1 file
```

### Group by file
```
errors.py — 3 issues
  C0114/missing-module-docstring  line 1 — Missing module docstring
  W0611/unused-import  line 1 — Unused import os
  E0602/undefined-variable  line 2 — Undefined variable 'undefined_var'
────────────────────────────────────────────────
3 issues · 3 rules · 1 file
```

## What prettyout improves

- **Consistent format**: pylint has its own output style → prettyout produces the same structure as all other supported tools
- **Rule ID + name linked**: raw output shows message ID and symbolic name separately → prettyout formats as `E0602/undefined-variable` keeping both together
- **Rule grouping**: raw output lists violations line-by-line → group-by-rule shows all files affected by the same rule at a glance
- **Occurrence counts**: no count in rule header → `E0602/undefined-variable (3) — message` shows how widespread each issue is
- **Summary line**: raw output ends with a rating score → `N issues · M rules · K files` replaces it with a consistent summary
