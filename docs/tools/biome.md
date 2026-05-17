# biome

**What it checks:** Fast formatter and linter for JavaScript, TypeScript, and JSON.

## Without prettyout

### Default output
```
Checked 1 file in 2ms. No fixes applied.
Found 3 errors.
```

### JSON (what CI/CD sees)
```json
{"summary":{"changed":0,"unchanged":1,"matches":0,"duration":2078375,"errors":3,
"warnings":0,"infos":0,"skipped":0,"suggestedFixesSkipped":0,"diagnosticsNotPrinted":0,
"scannerDuration":513417},"diagnostics":[{"severity":"error","message":"Using == may be
unsafe if you are relying on type coercion.","category":"lint/suspicious/noDoubleEquals",
"location":{"path":"errors.ts","start":{"line":2,"column":7},"end":{"line":2,"column":9}},
"advices":[]},{"severity":"error","message":"This is an unexpected use of the debugger
statement.","category":"lint/suspicious/noDebugger","location":{"path":"errors.ts",
"start":{"line":3,"column":1},"end":{"line":3,"column":10}},"advices":[]},
{"severity":"error","message":"Formatter would have printed the following content:",
"category":"format","location":{"path":"errors.ts","start":{"line":0,"column":0},
"end":{"line":0,"column":0}},"advices":[]}],"command":"check"}
```

## With prettyout

### Group by rule (default)
```
format (1) — Formatter would have printed the following content:
Affected files:
  - errors.ts
────────────────────────────────────────────────
lint/suspicious/noDebugger (1) — This is an unexpected use of the debugger statement.
Affected files:
  - errors.ts
────────────────────────────────────────────────
lint/suspicious/noDoubleEquals (1) — Using == may be unsafe if you are relying on type coercion.
Affected files:
  - errors.ts
────────────────────────────────────────────────
3 issues · 3 rules · 1 file
```

### Group by file
```
errors.ts — 3 issues
  lint/suspicious/noDoubleEquals — Using == may be unsafe if you are relying on type coercion.
  lint/suspicious/noDebugger — This is an unexpected use of the debugger statement.
  format — Formatter would have printed the following content:
────────────────────────────────────────────────
3 issues · 3 rules · 1 file
```

## What prettyout improves

- **Consistent format**: biome has a rich custom output → prettyout produces the same structure as all other supported tools
- **No line numbers in JSON**: biome reports byte offsets, not line/column in its JSON → prettyout shows filename-only entries cleanly without crashing on missing line info
- **Rule grouping**: raw output groups by file → group-by-rule shows all files where the same rule fires
- **Occurrence counts**: no count in rule header → `lint/suspicious/noDoubleEquals (3) — message` shows how widespread each rule is
- **Summary line**: raw JSON has no human summary → `N issues · M rules · K files` at the end
