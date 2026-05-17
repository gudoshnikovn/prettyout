# eslint

**What it checks:** JavaScript and TypeScript linter for style, errors, and best practices.

## Without prettyout

### Default output
```

/tmp/t-eslint/errors.js
  2:9  error  'undefined_var' is not defined  no-undef

✖ 1 problem (1 error, 0 warnings)
```

### JSON (what CI/CD sees)
```json
[{"filePath":"/tmp/t-eslint/errors.js","messages":[{"ruleId":"no-undef","severity":2,
"message":"'undefined_var' is not defined.","line":2,"column":9,"nodeType":"Identifier",
"messageId":"undef","endLine":2,"endColumn":22}],"suppressedMessages":[],
"errorCount":1,"fatalErrorCount":0,"warningCount":0,"fixableErrorCount":0,
"fixableWarningCount":0,"source":"var x = 1\nvar y = undefined_var\nx + y\n",
"usedDeprecatedRules":[]}]
```

## With prettyout

### Group by rule (default)
```
no-undef (1) — 'undefined_var' is not defined.
Affected files:
  - errors.js — line 2
────────────────────────────────────────────────
1 issue · 1 rule · 1 file
```

### Group by file
```
errors.js — 1 issue
  no-undef  line 2 — 'undefined_var' is not defined.
────────────────────────────────────────────────
1 issue · 1 rule · 1 file
```

## What prettyout improves

- **Consistent format**: eslint has its own columnar output style → prettyout produces the same structure as all other supported tools
- **Parse errors surfaced**: parse errors have `null` ruleId in JSON → prettyout shows them as `parse-error` rule, clearly visible
- **Rule grouping**: raw output groups by file, not rule → group-by-rule shows all files where the same rule fires
- **Occurrence counts**: no count in rule header → `no-unused-vars (4) — message` shows how widespread each rule is
- **Summary line**: raw output ends with a total count line → `N issues · M rules · K files` is consistent with all other tools
