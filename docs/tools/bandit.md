# bandit

**What it checks:** Python security linter — finds common security issues in Python code.

## Without prettyout

### Default output
```
Run started:2026-05-17 09:01:07.442712+00:00

Test results:
>> Issue: [B404:blacklist] Consider possible security implications associated with the subprocess module.
   Severity: Low   Confidence: High
   CWE: CWE-78 (https://cwe.mitre.org/data/definitions/78.html)
   More Info: https://bandit.readthedocs.io/en/1.9.4/blacklists/blacklist_imports.html#b404-import-subprocess
   Location: ./errors.py:2:0
1	import hashlib
2	import subprocess
3	hashlib.md5(b"data")

--------------------------------------------------
>> Issue: [B324:hashlib] Use of weak MD5 hash for security. Consider usedforsecurity=False
   Severity: High   Confidence: High
   CWE: CWE-327 (https://cwe.mitre.org/data/definitions/327.html)
   More Info: https://bandit.readthedocs.io/en/1.9.4/plugins/b324_hashlib.html
   Location: ./errors.py:3:0
2	import subprocess
3	hashlib.md5(b"data")
4	subprocess.call("ls", shell=True)

--------------------------------------------------
>> Issue: [B607:start_process_with_partial_path] Starting a process with a partial executable path
   Severity: Low   Confidence: High
   CWE: CWE-78 (https://cwe.mitre.org/data/definitions/78.html)
   More Info: https://bandit.readthedocs.io/en/1.9.4/plugins/b607_start_process_with_partial_path.html
   Location: ./errors.py:4:0
3	hashlib.md5(b"data")
4	subprocess.call("ls", shell=True)

--------------------------------------------------
>> Issue: [B602:subprocess_popen_with_shell_equals_true] subprocess call with shell=True seems safe, but may be changed in the future, consider rewriting without shell
   Severity: Low   Confidence: High
   CWE: CWE-78 (https://cwe.mitre.org/data/definitions/78.html)
   More Info: https://bandit.readthedocs.io/en/1.9.4/plugins/b602_subprocess_popen_with_shell_equals_true.html
   Location: ./errors.py:4:0
3	hashlib.md5(b"data")
4	subprocess.call("ls", shell=True)

--------------------------------------------------

Code scanned:
	Total lines of code: 4
	Total lines skipped (#nosec): 0
	Total potential issues skipped due to specifically being disabled (e.g., #nosec BXXX): 0

Run metrics:
	Total issues (by severity):
		Undefined: 0
		Low: 3
		Medium: 0
		High: 1
	Total issues (by confidence):
		Undefined: 0
		Low: 0
		Medium: 0
		High: 4
Files skipped (0):
```

### JSON (what CI/CD sees)
```json
{
  "errors": [],
  "generated_at": "2026-05-17T09:01:07Z",
  "metrics": {
    "./errors.py": {
      "CONFIDENCE.HIGH": 4,
      "CONFIDENCE.LOW": 0,
      "CONFIDENCE.MEDIUM": 0,
      "CONFIDENCE.UNDEFINED": 0,
      "SEVERITY.HIGH": 1,
      "SEVERITY.LOW": 3,
      "SEVERITY.MEDIUM": 0,
      "SEVERITY.UNDEFINED": 0,
      "loc": 4,
      "nosec": 0,
      "skipped_tests": 0
    },
    "_totals": {
      "CONFIDENCE.HIGH": 4,
      "CONFIDENCE.LOW": 0,
      "CONFIDENCE.MEDIUM": 0,
      "CONFIDENCE.UNDEFINED": 0,
      "SEVERITY.HIGH": 1,
      "SEVERITY.LOW": 3,
      "SEVERITY.MEDIUM": 0,
      "SEVERITY.UNDEFINED": 0,
      "loc": 4,
      "nosec": 0,
      "skipped_tests": 0
    }
...
```

## With prettyout

### Group by rule (default)
```
B324 (HIGH/HIGH) (1) — Use of weak MD5 hash for security. Consider usedforsecurity=False
Affected files:
  - errors.py — lines 3
────────────────────────────────────────────────
B404 (LOW/HIGH) (1) — Consider possible security implications associated with the subprocess module.
Affected files:
  - errors.py — lines 2
────────────────────────────────────────────────
B602 (LOW/HIGH) (1) — subprocess call with shell=True seems safe, but may be changed in the future, consider rewriting without shell
Affected files:
  - errors.py — lines 4
────────────────────────────────────────────────
B607 (LOW/HIGH) (1) — Starting a process with a partial executable path
Affected files:
  - errors.py — lines 4
────────────────────────────────────────────────
4 issues · 4 rules · 1 file
```

### Group by file
```
errors.py — 4 issues
  B404  line 2 — Consider possible security implications associated with the subprocess module.
  B324  line 3 — Use of weak MD5 hash for security. Consider usedforsecurity=False
  B607  line 4 — Starting a process with a partial executable path
  B602  line 4 — subprocess call with shell=True seems safe, but may be changed in the future, consider rewriting without shell
────────────────────────────────────────────────
4 issues · 4 rules · 1 file
```

## What prettyout improves

- **Consistent format**: bandit has a custom text format → prettyout produces the same structure as all other supported tools
- **Severity + confidence together**: raw output shows severity and confidence on separate lines → prettyout shows `B324 (HIGH/HIGH) — message` in the rule header
- **Rule grouping**: raw output lists each issue separately → group-by-rule shows all files affected by the same check code
- **Occurrence counts**: no count in rule header → `B324 (1) — message` shows how widespread each issue is
- **Summary line**: raw output ends with a long metrics block → `N issues · M rules · K files` is clean and consistent
