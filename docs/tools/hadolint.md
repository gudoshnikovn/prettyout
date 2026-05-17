# hadolint

**What it checks:** Dockerfile linter — enforces best practices in Docker image builds.

## Without prettyout

### Default output

```
Dockerfile.errors:1 DL3006 warning: Always tag the version of an image explicitly
Dockerfile.errors:2 DL3008 warning: Pin versions in apt get install. Instead of `apt-get install <package>` use `apt-get install <package>=<version>`
Dockerfile.errors:2 DL3015 info: Avoid additional packages by specifying `--no-install-recommends`
Dockerfile.errors:2 DL3014 warning: Use the `-y` switch to avoid manual input `apt-get -y install <package>`
Dockerfile.errors:3 DL3059 info: Multiple consecutive `RUN` instructions. Consider consolidation.
Dockerfile.errors:3 DL3008 warning: Pin versions in apt get install. Instead of `apt-get install <package>` use `apt-get install <package>=<version>`
Dockerfile.errors:3 DL3015 info: Avoid additional packages by specifying `--no-install-recommends`
Dockerfile.errors:3 DL3014 warning: Use the `-y` switch to avoid manual input `apt-get -y install <package>`
```

### JSON (what CI/CD sees)

```json
[{"code":"DL3006","column":1,"file":"Dockerfile.errors","level":"warning","line":1,"message":"Always tag the version of an image explicitly"},
{"code":"DL3008","column":1,"file":"Dockerfile.errors","level":"warning","line":2,"message":"Pin versions in apt get install. Instead of `apt-get install <package>` use `apt-get install <package>=<version>`"},
{"code":"DL3015","column":1,"file":"Dockerfile.errors","level":"info","line":2,"message":"Avoid additional packages by specifying `--no-install-recommends`"},
{"code":"DL3014","column":1,"file":"Dockerfile.errors","level":"warning","line":2,"message":"Use the `-y` switch to avoid manual input `apt-get -y install <package>`"},
{"code":"DL3059","column":1,"file":"Dockerfile.errors","level":"info","line":3,"message":"Multiple consecutive `RUN` instructions. Consider consolidation."},
...
]
```

## With prettyout

### Group by rule (default)

```
DL3006 (1) — Always tag the version of an image explicitly
Affected files:
  - Dockerfile.errors — lines 1
────────────────────────────────────────────────
DL3008 (2) — Pin versions in apt get install. Instead of `apt-get install <package>` use `apt-get install <package>=<version>`
Affected files:
  - Dockerfile.errors — lines 2, 3
────────────────────────────────────────────────
DL3014 (2) — Use the `-y` switch to avoid manual input `apt-get -y install <package>`
Affected files:
  - Dockerfile.errors — lines 2, 3
────────────────────────────────────────────────
DL3015 (2) — Avoid additional packages by specifying `--no-install-recommends`
Affected files:
  - Dockerfile.errors — lines 2, 3
────────────────────────────────────────────────
DL3059 (1) — Multiple consecutive `RUN` instructions. Consider consolidation.
Affected files:
  - Dockerfile.errors — lines 3
────────────────────────────────────────────────
8 issues · 5 rules · 1 file
```

### Group by file

```
Dockerfile.errors — 8 issues
  DL3006  line 1 — Always tag the version of an image explicitly
  DL3008  line 2 — Pin versions in apt get install. Instead of `apt-get install <package>` use `apt-get install <package>=<version>`
  DL3015  line 2 — Avoid additional packages by specifying `--no-install-recommends`
  DL3014  line 2 — Use the `-y` switch to avoid manual input `apt-get -y install <package>`
  DL3059  line 3 — Multiple consecutive `RUN` instructions. Consider consolidation.
  DL3008  line 3 — Pin versions in apt get install. Instead of `apt-get install <package>` use `apt-get install <package>=<version>`
  DL3015  line 3 — Avoid additional packages by specifying `--no-install-recommends`
  DL3014  line 3 — Use the `-y` switch to avoid manual input `apt-get -y install <package>`
────────────────────────────────────────────────
8 issues · 5 rules · 1 file
```

## What prettyout improves

- **Consistent format**: hadolint has its own output style → prettyout produces the same structure as all other supported tools
- **Mixed code prefixes unified**: hadolint emits both `DL` (Dockerfile) and `SC` (ShellCheck) codes → prettyout shows them uniformly in the same grouped format
- **Rule grouping**: raw output lists violations line-by-line → group-by-rule shows all affected Dockerfiles per rule
- **Occurrence counts**: no count in raw output → `DL3008 (2) — message` shows how widespread each rule is
- **Summary line**: raw output has no summary → `N issues · M rules · K files` at the end
