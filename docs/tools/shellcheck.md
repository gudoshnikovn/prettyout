# shellcheck

**What it checks:** Shell script linter — finds bugs and pitfalls in bash/sh scripts.

## Without prettyout

### Default output
```

In errors.sh line 2:
x=`echo hello`
  ^----------^ SC2006 (style): Use $(...) notation instead of legacy backticks `...`.
  ^----------^ SC2116 (style): Useless echo? Instead of 'cmd $(echo foo)', just use 'cmd foo'.

Did you mean: 
x=$(echo hello)


In errors.sh line 3:
if [ $x == "hello" ]; then
     ^-- SC2086 (info): Double quote to prevent globbing and word splitting.

Did you mean: 
if [ "$x" == "hello" ]; then

For more information:
  https://www.shellcheck.net/wiki/SC2086 -- Double quote to prevent globbing ...
  https://www.shellcheck.net/wiki/SC2006 -- Use $(...) notation instead of le...
  https://www.shellcheck.net/wiki/SC2116 -- Useless echo? Instead of 'cmd $(e...
```

### JSON (what CI/CD sees)
```json
[{"file":"errors.sh","line":2,"endLine":2,"column":3,"endColumn":15,"level":"style",
"code":2006,"message":"Use $(...) notation instead of legacy backticks `...`."},
{"file":"errors.sh","line":2,"endLine":2,"column":3,"endColumn":15,"level":"style",
"code":2116,"message":"Useless echo? Instead of 'cmd $(echo foo)', just use 'cmd foo'."},
{"file":"errors.sh","line":3,"endLine":3,"column":6,"endColumn":8,"level":"info",
"code":2086,"message":"Double quote to prevent globbing and word splitting."}]
```

## With prettyout

### Group by rule (default)
```
[INFO] SC2006 (1) — Use $(...) notation instead of legacy backticks `...`.
Affected files:
  - errors.sh — line 2
────────────────────────────────────────────────
[INFO] SC2086 (1) — Double quote to prevent globbing and word splitting.
Affected files:
  - errors.sh — line 3
────────────────────────────────────────────────
[INFO] SC2116 (1) — Useless echo? Instead of 'cmd $(echo foo)', just use 'cmd foo'.
Affected files:
  - errors.sh — line 2
────────────────────────────────────────────────
3 issues · 3 rules · 1 file
```

### Group by file
```
errors.sh — 3 issues
  SC2006  line 2 — Use $(...) notation instead of legacy backticks `...`.
  SC2116  line 2 — Useless echo? Instead of 'cmd $(echo foo)', just use 'cmd foo'.
  SC2086  line 3 — Double quote to prevent globbing and word splitting.
────────────────────────────────────────────────
3 issues · 3 rules · 1 file
```

## What prettyout improves

- **Consistent format**: shellcheck has a unique annotated output → prettyout produces the same structure as all other supported tools
- **Rule code formatted**: raw JSON stores rule as an integer (2006) → prettyout formats it as `SC2006` matching shellcheck's own documentation
- **Rule grouping**: raw output lists each warning separately → group-by-rule shows all files affected by the same code
- **Occurrence counts**: no count in raw output → `SC2006 (3) — message` shows how widespread each issue is
- **Summary line**: raw output has no summary → `N issues · M rules · K files` at the end
