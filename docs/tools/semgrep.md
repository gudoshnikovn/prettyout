# semgrep

**What it checks:** Semantic code analysis — finds bugs, security issues, and policy violations using custom rules.

## Without prettyout

### Default output

```
                   
┌─────────────────┐
│ 3 Code Findings │
└─────────────────┘
             
    errors.py
   ❯❯❱ shell-injection
          Possible shell injection
                                  
            3┆ subprocess.call(user_input, shell=True)
   
    ❯❱ hardcoded-password
          Possible hardcoded secret in password
                                               
            5┆ password = "hardcoded_secret_123"
            ⋮┆----------------------------------------
    ❯❱ hardcoded-password
          Possible hardcoded secret in api_key
                                              
            6┆ api_key = "another_secret_456"
```

### JSON (what CI/CD sees)

```json
{
  "version": "1.163.0",
  "results": [
    {
      "check_id": "shell-injection",
      "path": "errors.py",
      "start": {"line": 3, "col": 1},
      "end": {"line": 3, "col": 40},
      "extra": {
        "message": "Possible shell injection",
        "severity": "ERROR"
      }
    },
    {
      "check_id": "hardcoded-password",
      "path": "errors.py",
      "start": {"line": 5, "col": 1},
      "extra": {
        "message": "Possible hardcoded secret in password",
        "severity": "WARNING"
      }
    },
    ...
  ],
  "errors": [],
  "paths": {"scanned": ["errors.py"]}
}
```

## With prettyout

### Group by rule (default)

```
[WARN] hardcoded-password (2) — Possible hardcoded secret in password
Affected files:
  - errors.py — lines 5, 6
────────────────────────────────────────────────
[ERROR] shell-injection (1) — Possible shell injection
Affected files:
  - errors.py — line 3
────────────────────────────────────────────────
3 issues · 2 rules · 1 file
```

### Group by file

```
errors.py — 3 issues
  shell-injection  line 3 — Possible shell injection
  hardcoded-password  line 5 — Possible hardcoded secret in password
  hardcoded-password  line 6
────────────────────────────────────────────────
3 issues · 2 rules · 1 file
```

## What prettyout improves

- **Consistent format**: semgrep has a unique rich output → prettyout produces the same structure as all other supported tools
- **Rule code from check_id**: raw JSON uses `check_id` field as the rule identifier → prettyout uses it directly as the rule code for clean display
- **Severity normalized**: semgrep severity comes from `extra.severity` and may be uppercase or absent → prettyout normalizes and shows `[ERROR]` / `[WARN]` / `[INFO]` in the rule header
- **Rule grouping**: raw output lists matches one by one → group-by-rule shows all files affected by the same check
- **Summary line**: raw output ends with stats about rules run → `N issues · M rules · K files` is clean and consistent
