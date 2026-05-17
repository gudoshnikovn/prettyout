# basedpyright

**What it checks:** Strict static type checker for Python, based on pyright.

## Without prettyout

### Default output
```
/tmp/t-bp/errors.py
  /tmp/t-bp/errors.py:1:10 - error: Type "Literal['not an int']" is not assignable to declared type "int"
    "Literal['not an int']" is not assignable to "int" (reportAssignmentType)
  /tmp/t-bp/errors.py:2:12 - error: Function with declared return type "int" must return value on all code paths
    "None" is not assignable to "int" (reportReturnType)
2 errors, 0 warnings, 0 notes
```

### JSON (what CI/CD sees)
```json
{
  "version": "1.39.4",
  "time": "1779008400653",
  "generalDiagnostics": [
    {
      "file": "/tmp/t-bp/errors.py",
      "severity": "error",
      "message": "Type \"Literal['not an int']\" is not assignable to declared type \"int\"\n  \"Literal['not an int']\" is not assignable to \"int\"",
      "range": {
        "start": {
          "line": 0,
          "character": 9
        },
        "end": {
          "line": 0,
          "character": 21
        }
      },
      "rule": "reportAssignmentType"
    },
    ...
  ]
}
```

## With prettyout

### Group by rule (default)
```
[ERROR] reportAssignmentType (1) — Type "Literal['not an int']" is not assignable to declared type "int"
Affected files:
  - errors.py — line 1
────────────────────────────────────────────────
[ERROR] reportReturnType (1) — Function with declared return type "int" must return value on all code paths
Affected files:
  - errors.py — line 2
────────────────────────────────────────────────
2 issues · 2 rules · 1 file
```

### Group by file
```
errors.py — 2 issues
  [ERROR] reportAssignmentType  line 1  — Type "Literal['not an int']" is not assignable to declared type "int"
  [ERROR] reportReturnType  line 2  — Function with declared return type "int" must return value on all code paths
────────────────────────────────────────────────
2 issues · 2 rules · 1 file
```

## What prettyout improves

- **Consistent format**: basedpyright output is unique to pyright → prettyout produces the same structure as all other supported tools
- **Severity prefix**: raw output has no severity indicator → prettyout shows `[ERROR]` / `[WARN]` / `[INFO]` in each rule header
- **Duplicate deduplication**: parse errors can produce duplicate line numbers in the same diagnostic → prettyout uses a set-based approach, showing each line once
- **Rule grouping**: raw output lists diagnostics one by one → group-by-rule shows all files affected by the same rule code at a glance
- **Summary line**: raw output ends with a counts paragraph → `N issues · M rules · K files` is consistent with all other tools
