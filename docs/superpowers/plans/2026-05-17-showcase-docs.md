# Showcase Docs Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce `docs/tools/<tool>.md` for all 14 tools and `docs/summary.md` showing real before/after format comparisons.

**Architecture:** For each tool, run a Docker command to capture 4 outputs (default, JSON, prettyout group-by-rule, prettyout group-by-file), then write the markdown file. All fixtures are pre-existing in `test/fixtures/`. A final task writes `docs/summary.md`.

**Tech Stack:** Docker (`prettyout-test` image), bash, existing fixtures in `test/fixtures/`.

---

## File Map

**Create:**
- `docs/tools/ruff.md`
- `docs/tools/mypy.md`
- `docs/tools/basedpyright.md`
- `docs/tools/bandit.md`
- `docs/tools/pylint.md`
- `docs/tools/eslint.md`
- `docs/tools/biome.md`
- `docs/tools/stylelint.md`
- `docs/tools/shellcheck.md`
- `docs/tools/hadolint.md`
- `docs/tools/golangci-lint.md`
- `docs/tools/cargo-clippy.md`
- `docs/tools/trivy.md`
- `docs/tools/semgrep.md`
- `docs/summary.md`

---

## Task 0: Setup

- [ ] **Verify Docker image is current**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker build -t prettyout-test -f test/Dockerfile . 2>&1 | tail -5
```

Expected: build completes (may use cache), last line contains "Successfully built" or "FINISHED".

- [ ] **Verify all plugins are available in the image**

```bash
docker run --rm prettyout-test bash -c "ls /usr/local/bin/prettyout-*"
```

Expected: list includes prettyout-ruff, prettyout-mypy, prettyout-basedpyright, prettyout-bandit, prettyout-pylint, prettyout-eslint, prettyout-biome, prettyout-stylelint, prettyout-shellcheck, prettyout-hadolint, prettyout-golangci, prettyout-cargo-clippy, prettyout-trivy, prettyout-semgrep.

- [ ] **Create docs/tools/ directory**

```bash
mkdir -p /Users/gudoshnikov_na/Programming/Agents/prettyout/docs/tools
```

---

## Task 1: ruff

**Files:** Create `docs/tools/ruff.md`

Fixture: `test/fixtures/ruff/errors.py`

- [ ] **Capture outputs**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  rm -rf /tmp/t-ruff && cp -r /project/test/fixtures/ruff /tmp/t-ruff
  cd /tmp/t-ruff

  echo '=== DEFAULT ==='
  ruff check errors.py 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g'

  echo '=== JSON ==='
  ruff check --output-format=json errors.py 2>/dev/null | head -30

  echo '=== PRETTYOUT RULE ==='
  printf 'settings:\n  ruff:\n    colors: false\n' > .prettyout.yaml
  ruff check --output-format=json errors.py 2>/dev/null | prettyout-ruff
  rm .prettyout.yaml

  echo '=== PRETTYOUT FILE ==='
  printf 'settings:\n  ruff:\n    group_by: file\n    colors: false\n' > .prettyout.yaml
  ruff check --output-format=json errors.py 2>/dev/null | prettyout-ruff
  rm .prettyout.yaml
"
```

- [ ] **Write `docs/tools/ruff.md`** with the captured outputs:

```markdown
# ruff

**What it checks:** Python linter for style errors, unused imports, and code quality.

## Without prettyout

### Default output
\```
<paste === DEFAULT === output here>
\```

### JSON (what CI/CD sees)
\```json
<paste === JSON === output here>
\```

## With prettyout

### Group by rule (default)
\```
<paste === PRETTYOUT RULE === output here>
\```

### Group by file
\```
<paste === PRETTYOUT FILE === output here>
\```

## What prettyout improves

- **Consistent format**: ruff has its own output style → prettyout produces the same structure as all other supported tools
- **Rule grouping**: raw output lists violations line-by-line → group-by-rule view shows all affected files per rule at a glance
- **Occurrence counts**: no summary in raw output → `E501 (4) — message` shows how widespread each rule is
- **Summary line**: must scroll to the end to see counts → `N issues · M rules · K files` always at the end
```

- [ ] **Commit**

```bash
git add docs/tools/ruff.md
git commit -m "docs: add ruff showcase"
```

---

## Task 2: mypy

**Files:** Create `docs/tools/mypy.md`

Fixture: `test/fixtures/mypy/errors.py`

- [ ] **Capture outputs**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  rm -rf /tmp/t-mypy && cp -r /project/test/fixtures/mypy /tmp/t-mypy
  cd /tmp/t-mypy

  echo '=== DEFAULT ==='
  mypy errors.py 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g'

  echo '=== JSON ==='
  mypy --output=json errors.py 2>/dev/null | head -30

  echo '=== PRETTYOUT RULE ==='
  printf 'settings:\n  mypy:\n    colors: false\n' > .prettyout.yaml
  mypy --output=json errors.py 2>/dev/null | prettyout-mypy
  rm .prettyout.yaml

  echo '=== PRETTYOUT FILE ==='
  printf 'settings:\n  mypy:\n    group_by: file\n    colors: false\n' > .prettyout.yaml
  mypy --output=json errors.py 2>/dev/null | prettyout-mypy
  rm .prettyout.yaml
"
```

- [ ] **Write `docs/tools/mypy.md`** with captured outputs:

```markdown
# mypy

**What it checks:** Static type checker for Python.

## Without prettyout

### Default output
\```
<paste === DEFAULT === output here>
\```

### JSON (what CI/CD sees)
\```json
<paste === JSON === output here>
\```

## With prettyout

### Group by rule (default)
\```
<paste === PRETTYOUT RULE === output here>
\```

### Group by file
\```
<paste === PRETTYOUT FILE === output here>
\```

## What prettyout improves

- **Consistent format**: mypy has its own output style → prettyout produces the same structure as all other supported tools
- **Rule grouping**: raw output lists violations line-by-line → group-by-rule shows all files affected by the same error code
- **Occurrence counts**: no summary in raw output → `error-code (4) — message` shows how widespread each issue is
- **Note clutter removed**: mypy emits `note`-level messages as context lines → prettyout strips them, showing only actionable errors and warnings
- **Summary line**: `Found N errors` is hard to scan → `N issues · M rules · K files` at the end is consistent with all other tools
```

- [ ] **Commit**

```bash
git add docs/tools/mypy.md
git commit -m "docs: add mypy showcase"
```

---

## Task 3: basedpyright

**Files:** Create `docs/tools/basedpyright.md`

Fixture: `test/fixtures/basedpyright/errors.py`

- [ ] **Capture outputs**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  rm -rf /tmp/t-bp && cp -r /project/test/fixtures/basedpyright /tmp/t-bp
  cd /tmp/t-bp

  echo '=== DEFAULT ==='
  basedpyright errors.py 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g'

  echo '=== JSON ==='
  basedpyright --outputjson errors.py 2>/dev/null | python3 -c 'import sys,json; d=json.load(sys.stdin); print(json.dumps(d,indent=2))' | head -30

  echo '=== PRETTYOUT RULE ==='
  printf 'settings:\n  basedpyright:\n    colors: false\n' > .prettyout.yaml
  basedpyright --outputjson errors.py 2>/dev/null | prettyout-basedpyright
  rm .prettyout.yaml

  echo '=== PRETTYOUT FILE ==='
  printf 'settings:\n  basedpyright:\n    group_by: file\n    colors: false\n' > .prettyout.yaml
  basedpyright --outputjson errors.py 2>/dev/null | prettyout-basedpyright
  rm .prettyout.yaml
"
```

- [ ] **Write `docs/tools/basedpyright.md`** with captured outputs:

```markdown
# basedpyright

**What it checks:** Strict static type checker for Python, based on pyright.

## Without prettyout

### Default output
\```
<paste === DEFAULT === output here>
\```

### JSON (what CI/CD sees)
\```json
<paste === JSON === output here>
\```

## With prettyout

### Group by rule (default)
\```
<paste === PRETTYOUT RULE === output here>
\```

### Group by file
\```
<paste === PRETTYOUT FILE === output here>
\```

## What prettyout improves

- **Consistent format**: basedpyright output is unique to pyright → prettyout produces the same structure as all other supported tools
- **Severity prefix**: raw output has no severity indicator → prettyout shows `[ERROR]` / `[WARN]` / `[INFO]` in each rule header
- **Duplicate deduplication**: parse errors can produce duplicate line numbers in the same diagnostic → prettyout uses a set-based approach, showing each line once
- **Rule grouping**: raw output lists diagnostics one by one → group-by-rule shows all files affected by the same rule code at a glance
- **Summary line**: raw output ends with a counts paragraph → `N issues · M rules · K files` is consistent with all other tools
```

- [ ] **Commit**

```bash
git add docs/tools/basedpyright.md
git commit -m "docs: add basedpyright showcase"
```

---

## Task 4: bandit

**Files:** Create `docs/tools/bandit.md`

Fixture: `test/fixtures/bandit/errors.py`

- [ ] **Capture outputs**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  rm -rf /tmp/t-bandit && cp -r /project/test/fixtures/bandit /tmp/t-bandit
  cd /tmp/t-bandit

  echo '=== DEFAULT ==='
  bandit errors.py 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g'

  echo '=== JSON ==='
  bandit -f json errors.py 2>/dev/null | head -30

  echo '=== PRETTYOUT RULE ==='
  printf 'settings:\n  bandit:\n    colors: false\n' > .prettyout.yaml
  bandit -f json errors.py 2>/dev/null | prettyout-bandit
  rm .prettyout.yaml

  echo '=== PRETTYOUT FILE ==='
  printf 'settings:\n  bandit:\n    group_by: file\n    colors: false\n' > .prettyout.yaml
  bandit -f json errors.py 2>/dev/null | prettyout-bandit
  rm .prettyout.yaml
"
```

- [ ] **Write `docs/tools/bandit.md`** with captured outputs:

```markdown
# bandit

**What it checks:** Python security linter — finds common security issues in Python code.

## Without prettyout

### Default output
\```
<paste === DEFAULT === output here>
\```

### JSON (what CI/CD sees)
\```json
<paste === JSON === output here>
\```

## With prettyout

### Group by rule (default)
\```
<paste === PRETTYOUT RULE === output here>
\```

### Group by file
\```
<paste === PRETTYOUT FILE === output here>
\```

## What prettyout improves

- **Consistent format**: bandit has a custom text format → prettyout produces the same structure as all other supported tools
- **Severity + confidence together**: raw output shows severity and confidence on separate lines → prettyout shows `B324 (MEDIUM/HIGH) — message` in the rule header
- **Rule grouping**: raw output lists each issue separately → group-by-rule shows all files affected by the same check code
- **Occurrence counts**: no count in rule header → `B324 (3) — message` shows how widespread each issue is
- **Summary line**: raw output ends with a long metrics block → `N issues · M rules · K files` is clean and consistent
```

- [ ] **Commit**

```bash
git add docs/tools/bandit.md
git commit -m "docs: add bandit showcase"
```

---

## Task 5: pylint

**Files:** Create `docs/tools/pylint.md`

Fixture: `test/fixtures/pylint/errors.py`

- [ ] **Capture outputs**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  rm -rf /tmp/t-pylint && cp -r /project/test/fixtures/pylint /tmp/t-pylint
  cd /tmp/t-pylint

  echo '=== DEFAULT ==='
  pylint errors.py 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g'

  echo '=== JSON ==='
  pylint --output-format=json errors.py 2>/dev/null | head -30

  echo '=== PRETTYOUT RULE ==='
  printf 'settings:\n  pylint:\n    colors: false\n' > .prettyout.yaml
  pylint --output-format=json errors.py 2>/dev/null | prettyout-pylint
  rm .prettyout.yaml

  echo '=== PRETTYOUT FILE ==='
  printf 'settings:\n  pylint:\n    group_by: file\n    colors: false\n' > .prettyout.yaml
  pylint --output-format=json errors.py 2>/dev/null | prettyout-pylint
  rm .prettyout.yaml
"
```

- [ ] **Write `docs/tools/pylint.md`** with captured outputs:

```markdown
# pylint

**What it checks:** Python linter for errors, style, and code smells.

## Without prettyout

### Default output
\```
<paste === DEFAULT === output here>
\```

### JSON (what CI/CD sees)
\```json
<paste === JSON === output here>
\```

## With prettyout

### Group by rule (default)
\```
<paste === PRETTYOUT RULE === output here>
\```

### Group by file
\```
<paste === PRETTYOUT FILE === output here>
\```

## What prettyout improves

- **Consistent format**: pylint has its own output style → prettyout produces the same structure as all other supported tools
- **Rule ID + name linked**: raw output shows message ID and symbolic name separately → prettyout formats as `E0602/undefined-variable` keeping both together
- **Rule grouping**: raw output lists violations line-by-line → group-by-rule shows all files affected by the same rule at a glance
- **Occurrence counts**: no count in rule header → `E0602/undefined-variable (3) — message` shows how widespread each issue is
- **Summary line**: raw output ends with a rating score → `N issues · M rules · K files` replaces it with a consistent summary
```

- [ ] **Commit**

```bash
git add docs/tools/pylint.md
git commit -m "docs: add pylint showcase"
```

---

## Task 6: eslint

**Files:** Create `docs/tools/eslint.md`

Fixtures: `test/fixtures/eslint/errors.js` + `test/fixtures/eslint/eslint.config.mjs`

- [ ] **Capture outputs**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  rm -rf /tmp/t-eslint && cp -r /project/test/fixtures/eslint /tmp/t-eslint
  cd /tmp/t-eslint

  echo '=== DEFAULT ==='
  eslint errors.js 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g'

  echo '=== JSON ==='
  eslint --format=json errors.js 2>/dev/null | head -30

  echo '=== PRETTYOUT RULE ==='
  printf 'settings:\n  eslint:\n    colors: false\n' > .prettyout.yaml
  eslint --format=json errors.js 2>/dev/null | prettyout-eslint
  rm .prettyout.yaml

  echo '=== PRETTYOUT FILE ==='
  printf 'settings:\n  eslint:\n    group_by: file\n    colors: false\n' > .prettyout.yaml
  eslint --format=json errors.js 2>/dev/null | prettyout-eslint
  rm .prettyout.yaml
"
```

- [ ] **Write `docs/tools/eslint.md`** with captured outputs:

```markdown
# eslint

**What it checks:** JavaScript and TypeScript linter for style, errors, and best practices.

## Without prettyout

### Default output
\```
<paste === DEFAULT === output here>
\```

### JSON (what CI/CD sees)
\```json
<paste === JSON === output here>
\```

## With prettyout

### Group by rule (default)
\```
<paste === PRETTYOUT RULE === output here>
\```

### Group by file
\```
<paste === PRETTYOUT FILE === output here>
\```

## What prettyout improves

- **Consistent format**: eslint has its own columnar output style → prettyout produces the same structure as all other supported tools
- **Parse errors surfaced**: parse errors have `null` ruleId in JSON → prettyout shows them as `parse-error` rule, clearly visible
- **Rule grouping**: raw output groups by file, not rule → group-by-rule shows all files where the same rule fires
- **Occurrence counts**: no count in rule header → `no-unused-vars (4) — message` shows how widespread each rule is
- **Summary line**: raw output ends with a total count line → `N issues · M rules · K files` is consistent with all other tools
```

- [ ] **Commit**

```bash
git add docs/tools/eslint.md
git commit -m "docs: add eslint showcase"
```

---

## Task 7: biome

**Files:** Create `docs/tools/biome.md`

Fixtures: `test/fixtures/biome/errors.ts` + `test/fixtures/biome/biome.json`

- [ ] **Capture outputs**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  rm -rf /tmp/t-biome && cp -r /project/test/fixtures/biome /tmp/t-biome
  cd /tmp/t-biome

  echo '=== DEFAULT ==='
  biome check errors.ts 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g'

  echo '=== JSON ==='
  biome check --reporter=json errors.ts 2>/dev/null | head -30

  echo '=== PRETTYOUT RULE ==='
  printf 'settings:\n  biome:\n    colors: false\n' > .prettyout.yaml
  biome check --reporter=json errors.ts 2>/dev/null | prettyout-biome
  rm .prettyout.yaml

  echo '=== PRETTYOUT FILE ==='
  printf 'settings:\n  biome:\n    group_by: file\n    colors: false\n' > .prettyout.yaml
  biome check --reporter=json errors.ts 2>/dev/null | prettyout-biome
  rm .prettyout.yaml
"
```

- [ ] **Write `docs/tools/biome.md`** with captured outputs:

```markdown
# biome

**What it checks:** Fast formatter and linter for JavaScript, TypeScript, and JSON.

## Without prettyout

### Default output
\```
<paste === DEFAULT === output here>
\```

### JSON (what CI/CD sees)
\```json
<paste === JSON === output here>
\```

## With prettyout

### Group by rule (default)
\```
<paste === PRETTYOUT RULE === output here>
\```

### Group by file
\```
<paste === PRETTYOUT FILE === output here>
\```

## What prettyout improves

- **Consistent format**: biome has a rich custom output → prettyout produces the same structure as all other supported tools
- **No line numbers in JSON**: biome reports byte offsets, not line/column in its JSON → prettyout shows filename-only entries cleanly without crashing on missing line info
- **Rule grouping**: raw output groups by file → group-by-rule shows all files where the same rule fires
- **Occurrence counts**: no count in rule header → `lint/suspicious/noDoubleEquals (3) — message` shows how widespread each rule is
- **Summary line**: raw JSON has no human summary → `N issues · M rules · K files` at the end
```

- [ ] **Commit**

```bash
git add docs/tools/biome.md
git commit -m "docs: add biome showcase"
```

---

## Task 8: stylelint

**Files:** Create `docs/tools/stylelint.md`

Fixtures: `test/fixtures/stylelint/errors.css`

Note: stylelint writes JSON to **stderr**. The prettyout pipe must be `2>&1 >/dev/null | prettyout-stylelint`.

- [ ] **Capture outputs**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  rm -rf /tmp/t-stylelint && cp -r /project/test/fixtures/stylelint /tmp/t-stylelint
  cd /tmp/t-stylelint
  printf '{\"rules\":{\"color-no-invalid-hex\":true,\"font-family-no-unknown-names\":true,\"property-no-unknown\":true}}' > .stylelintrc.json

  echo '=== DEFAULT ==='
  stylelint errors.css 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g'

  echo '=== JSON ==='
  stylelint --formatter=json errors.css 2>&1 >/dev/null | head -30

  echo '=== PRETTYOUT RULE ==='
  printf 'settings:\n  stylelint:\n    colors: false\n' > .prettyout.yaml
  stylelint --formatter=json errors.css 2>&1 >/dev/null | prettyout-stylelint
  rm .prettyout.yaml

  echo '=== PRETTYOUT FILE ==='
  printf 'settings:\n  stylelint:\n    group_by: file\n    colors: false\n' > .prettyout.yaml
  stylelint --formatter=json errors.css 2>&1 >/dev/null | prettyout-stylelint
  rm .prettyout.yaml
"
```

- [ ] **Write `docs/tools/stylelint.md`** with captured outputs:

```markdown
# stylelint

**What it checks:** CSS/SCSS/Less linter for style errors and best practices.

## Without prettyout

### Default output
\```
<paste === DEFAULT === output here>
\```

### JSON (what CI/CD sees, on stderr)
\```json
<paste === JSON === output here>
\```

## With prettyout

### Group by rule (default)
\```
<paste === PRETTYOUT RULE === output here>
\```

### Group by file
\```
<paste === PRETTYOUT FILE === output here>
\```

## What prettyout improves

- **Consistent format**: stylelint has its own output style → prettyout produces the same structure as all other supported tools
- **JSON on stderr handled**: stylelint outputs JSON to stderr, not stdout — most tools can't consume it directly → prettyout handles this transparently via `2>&1 >/dev/null` piping
- **Absolute paths resolved**: stylelint emits absolute file paths in JSON → prettyout converts them to relative paths from the current directory
- **Rule grouping**: raw output lists violations per file → group-by-rule shows all files affected by the same rule
- **Summary line**: raw output has no count summary → `N issues · M rules · K files` at the end
```

- [ ] **Commit**

```bash
git add docs/tools/stylelint.md
git commit -m "docs: add stylelint showcase"
```

---

## Task 9: shellcheck

**Files:** Create `docs/tools/shellcheck.md`

Fixture: `test/fixtures/shellcheck/errors.sh`

- [ ] **Capture outputs**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  rm -rf /tmp/t-shellcheck && cp -r /project/test/fixtures/shellcheck /tmp/t-shellcheck
  cd /tmp/t-shellcheck

  echo '=== DEFAULT ==='
  shellcheck errors.sh 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g'

  echo '=== JSON ==='
  shellcheck --format=json errors.sh 2>/dev/null | head -30

  echo '=== PRETTYOUT RULE ==='
  printf 'settings:\n  shellcheck:\n    colors: false\n' > .prettyout.yaml
  shellcheck --format=json errors.sh 2>/dev/null | prettyout-shellcheck
  rm .prettyout.yaml

  echo '=== PRETTYOUT FILE ==='
  printf 'settings:\n  shellcheck:\n    group_by: file\n    colors: false\n' > .prettyout.yaml
  shellcheck --format=json errors.sh 2>/dev/null | prettyout-shellcheck
  rm .prettyout.yaml
"
```

- [ ] **Write `docs/tools/shellcheck.md`** with captured outputs:

```markdown
# shellcheck

**What it checks:** Shell script linter — finds bugs and pitfalls in bash/sh scripts.

## Without prettyout

### Default output
\```
<paste === DEFAULT === output here>
\```

### JSON (what CI/CD sees)
\```json
<paste === JSON === output here>
\```

## With prettyout

### Group by rule (default)
\```
<paste === PRETTYOUT RULE === output here>
\```

### Group by file
\```
<paste === PRETTYOUT FILE === output here>
\```

## What prettyout improves

- **Consistent format**: shellcheck has a unique annotated output → prettyout produces the same structure as all other supported tools
- **Rule code formatted**: raw JSON stores rule as an integer (2006) → prettyout formats it as `SC2006` matching shellcheck's own documentation
- **Rule grouping**: raw output lists each warning separately → group-by-rule shows all files affected by the same code
- **Occurrence counts**: no count in raw output → `SC2006 (3) — message` shows how widespread each issue is
- **Summary line**: raw output has no summary → `N issues · M rules · K files` at the end
```

- [ ] **Commit**

```bash
git add docs/tools/shellcheck.md
git commit -m "docs: add shellcheck showcase"
```

---

## Task 10: hadolint

**Files:** Create `docs/tools/hadolint.md`

Fixture: `test/fixtures/hadolint/Dockerfile.errors`

- [ ] **Capture outputs**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  rm -rf /tmp/t-hadolint && cp -r /project/test/fixtures/hadolint /tmp/t-hadolint
  cd /tmp/t-hadolint

  echo '=== DEFAULT ==='
  hadolint Dockerfile.errors 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g'

  echo '=== JSON ==='
  hadolint --format=json Dockerfile.errors 2>/dev/null | head -30

  echo '=== PRETTYOUT RULE ==='
  printf 'settings:\n  hadolint:\n    colors: false\n' > .prettyout.yaml
  hadolint --format=json Dockerfile.errors 2>/dev/null | prettyout-hadolint
  rm .prettyout.yaml

  echo '=== PRETTYOUT FILE ==='
  printf 'settings:\n  hadolint:\n    group_by: file\n    colors: false\n' > .prettyout.yaml
  hadolint --format=json Dockerfile.errors 2>/dev/null | prettyout-hadolint
  rm .prettyout.yaml
"
```

- [ ] **Write `docs/tools/hadolint.md`** with captured outputs:

```markdown
# hadolint

**What it checks:** Dockerfile linter — enforces best practices in Docker image builds.

## Without prettyout

### Default output
\```
<paste === DEFAULT === output here>
\```

### JSON (what CI/CD sees)
\```json
<paste === JSON === output here>
\```

## With prettyout

### Group by rule (default)
\```
<paste === PRETTYOUT RULE === output here>
\```

### Group by file
\```
<paste === PRETTYOUT FILE === output here>
\```

## What prettyout improves

- **Consistent format**: hadolint has its own output style → prettyout produces the same structure as all other supported tools
- **Mixed code prefixes unified**: hadolint emits both `DL` (Dockerfile) and `SC` (ShellCheck) codes → prettyout shows them uniformly in the same grouped format
- **Rule grouping**: raw output lists violations line-by-line → group-by-rule shows all affected Dockerfiles per rule
- **Occurrence counts**: no count in raw output → `DL3008 (2) — message` shows how widespread each rule is
- **Summary line**: raw output has no summary → `N issues · M rules · K files` at the end
```

- [ ] **Commit**

```bash
git add docs/tools/hadolint.md
git commit -m "docs: add hadolint showcase"
```

---

## Task 11: golangci-lint

**Files:** Create `docs/tools/golangci-lint.md`

Fixture: `test/fixtures/golangci-lint/errors/` (contains `go.mod` + `main.go`)

- [ ] **Capture outputs**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  rm -rf /tmp/t-golangci && cp -r /project/test/fixtures/golangci-lint/errors /tmp/t-golangci
  cd /tmp/t-golangci

  echo '=== DEFAULT ==='
  golangci-lint run --disable-all --enable=ineffassign 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g'

  echo '=== JSON ==='
  golangci-lint run --out-format=json --disable-all --enable=ineffassign 2>/dev/null | head -30

  echo '=== PRETTYOUT RULE ==='
  printf 'settings:\n  golangci:\n    colors: false\n' > .prettyout.yaml
  golangci-lint run --out-format=json --disable-all --enable=ineffassign 2>/dev/null | prettyout-golangci
  rm .prettyout.yaml

  echo '=== PRETTYOUT FILE ==='
  printf 'settings:\n  golangci:\n    group_by: file\n    colors: false\n' > .prettyout.yaml
  golangci-lint run --out-format=json --disable-all --enable=ineffassign 2>/dev/null | prettyout-golangci
  rm .prettyout.yaml
"
```

- [ ] **Write `docs/tools/golangci-lint.md`** with captured outputs:

```markdown
# golangci-lint

**What it checks:** Go meta-linter — runs multiple Go linters in parallel.

## Without prettyout

### Default output
\```
<paste === DEFAULT === output here>
\```

### JSON (what CI/CD sees)
\```json
<paste === JSON === output here>
\```

## With prettyout

### Group by rule (default)
\```
<paste === PRETTYOUT RULE === output here>
\```

### Group by file
\```
<paste === PRETTYOUT FILE === output here>
\```

## What prettyout improves

- **Consistent format**: golangci-lint has its own output style → prettyout produces the same structure as all other supported tools
- **Empty run handled gracefully**: golangci-lint emits `"Issues": null` (not `[]`) on a clean run → prettyout handles this and shows `0 issues · 0 rules · 0 files`
- **Rule grouping**: raw output lists issues line-by-line → group-by-rule shows all files affected by the same linter check
- **Occurrence counts**: no count in rule header → `ineffassign (2) — message` shows how widespread each issue is
- **Summary line**: raw output has no summary → `N issues · M rules · K files` at the end
```

- [ ] **Commit**

```bash
git add docs/tools/golangci-lint.md
git commit -m "docs: add golangci-lint showcase"
```

---

## Task 12: cargo clippy

**Files:** Create `docs/tools/cargo-clippy.md`

Fixture: `test/fixtures/cargo-clippy/errors/` (contains `Cargo.toml` + `src/main.rs`)

- [ ] **Capture outputs**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  rm -rf /tmp/t-clippy && cp -r /project/test/fixtures/cargo-clippy/errors /tmp/t-clippy
  cd /tmp/t-clippy

  echo '=== DEFAULT ==='
  cargo clippy 2>&1 | sed 's/\x1b\[[0-9;]*m//g' | grep -v '^error\[' | head -20

  echo '=== JSON ==='
  cargo clippy --message-format=json 2>/dev/null | head -30

  echo '=== PRETTYOUT RULE ==='
  printf 'settings:\n  cargo-clippy:\n    colors: false\n' > .prettyout.yaml
  cargo clippy --message-format=json 2>/dev/null | prettyout-cargo-clippy
  rm .prettyout.yaml

  echo '=== PRETTYOUT FILE ==='
  printf 'settings:\n  cargo-clippy:\n    group_by: file\n    colors: false\n' > .prettyout.yaml
  cargo clippy --message-format=json 2>/dev/null | prettyout-cargo-clippy
  rm .prettyout.yaml
"
```

- [ ] **Write `docs/tools/cargo-clippy.md`** with captured outputs:

```markdown
# cargo clippy

**What it checks:** Rust linter — catches common mistakes and suggests idiomatic Rust.

## Without prettyout

### Default output
\```
<paste === DEFAULT === output here>
\```

### JSON (NDJSON stream, what CI/CD sees)
\```
<paste === JSON === output here>
\```

## With prettyout

### Group by rule (default)
\```
<paste === PRETTYOUT RULE === output here>
\```

### Group by file
\```
<paste === PRETTYOUT FILE === output here>
\```

## What prettyout improves

- **Consistent format**: cargo clippy emits a noisy NDJSON stream mixing compiler messages and metadata → prettyout filters to `compiler-message` type only and produces the same structure as all other supported tools
- **NDJSON parsed correctly**: raw output is one JSON object per line, not a JSON array → prettyout handles NDJSON natively
- **Rule grouping**: raw output lists warnings in order → group-by-rule shows all files affected by the same clippy lint
- **Occurrence counts**: no count in raw output → `clippy::needless_return (2) — message` shows how widespread each lint is
- **Summary line**: raw output ends with compiler summary lines → `N issues · M rules · K files` is clean and consistent
```

- [ ] **Commit**

```bash
git add docs/tools/cargo-clippy.md
git commit -m "docs: add cargo-clippy showcase"
```

---

## Task 13: trivy

**Files:** Create `docs/tools/trivy.md`

Fixture: `test/fixtures/trivy/django/requirements.txt`

- [ ] **Capture outputs**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  rm -rf /tmp/t-trivy && cp -r /project/test/fixtures/trivy/django /tmp/t-trivy
  cd /tmp/t-trivy

  echo '=== DEFAULT ==='
  trivy fs --quiet . 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g' | head -30

  echo '=== JSON ==='
  trivy fs --format=json --quiet . 2>/dev/null | python3 -c 'import sys,json; d=json.load(sys.stdin); print(json.dumps(d,indent=2))' | head -40

  echo '=== PRETTYOUT RULE ==='
  printf 'settings:\n  trivy:\n    colors: false\n' > .prettyout.yaml
  trivy fs --format=json --quiet . 2>/dev/null | prettyout-trivy
  rm .prettyout.yaml

  echo '=== PRETTYOUT FILE ==='
  printf 'settings:\n  trivy:\n    group_by: file\n    colors: false\n' > .prettyout.yaml
  trivy fs --format=json --quiet . 2>/dev/null | prettyout-trivy
  rm .prettyout.yaml
"
```

- [ ] **Write `docs/tools/trivy.md`** with captured outputs:

```markdown
# trivy

**What it checks:** Security scanner — finds vulnerabilities in dependencies, containers, and IaC.

## Without prettyout

### Default output
\```
<paste === DEFAULT === output here>
\```

### JSON (what CI/CD sees)
\```json
<paste === JSON === output here>
\```

## With prettyout

### Group by rule (default — grouped by CVE)
\```
<paste === PRETTYOUT RULE === output here>
\```

### Group by file (grouped by package file)
\```
<paste === PRETTYOUT FILE === output here>
\```

## What prettyout improves

- **Consistent format**: trivy table output is unique to trivy → prettyout produces the same structure as all other supported tools
- **Severity ordering**: raw output mixes severities → prettyout groups by severity in CRITICAL → HIGH → MEDIUM → LOW → UNKNOWN order
- **Fix availability surfaced**: raw output buries "no fix available" in the table → prettyout shows it explicitly in the vulnerability entry
- **Rule grouping**: raw output groups by package file/image → group-by-rule (CVE) shows all packages affected by the same vulnerability
- **Summary line**: raw output ends with a table footer → `N issues · M rules · K files` is clean and consistent
```

- [ ] **Commit**

```bash
git add docs/tools/trivy.md
git commit -m "docs: add trivy showcase"
```

---

## Task 14: semgrep

**Files:** Create `docs/tools/semgrep.md`

Fixtures: `test/fixtures/semgrep/errors.py` + `test/fixtures/semgrep/semgrep-rules.yaml`

- [ ] **Capture outputs**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker run --rm -v $(pwd):/project prettyout-test bash -c "
  rm -rf /tmp/t-semgrep && cp -r /project/test/fixtures/semgrep /tmp/t-semgrep
  cd /tmp/t-semgrep

  echo '=== DEFAULT ==='
  semgrep --config semgrep-rules.yaml errors.py 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g'

  echo '=== JSON ==='
  semgrep --config semgrep-rules.yaml --json errors.py 2>/dev/null | head -30

  echo '=== PRETTYOUT RULE ==='
  printf 'settings:\n  semgrep:\n    colors: false\n' > .prettyout.yaml
  semgrep --config semgrep-rules.yaml --json errors.py 2>/dev/null | prettyout-semgrep
  rm .prettyout.yaml

  echo '=== PRETTYOUT FILE ==='
  printf 'settings:\n  semgrep:\n    group_by: file\n    colors: false\n' > .prettyout.yaml
  semgrep --config semgrep-rules.yaml --json errors.py 2>/dev/null | prettyout-semgrep
  rm .prettyout.yaml
"
```

- [ ] **Write `docs/tools/semgrep.md`** with captured outputs:

```markdown
# semgrep

**What it checks:** Semantic code analysis — finds bugs, security issues, and policy violations using custom rules.

## Without prettyout

### Default output
\```
<paste === DEFAULT === output here>
\```

### JSON (what CI/CD sees)
\```json
<paste === JSON === output here>
\```

## With prettyout

### Group by rule (default)
\```
<paste === PRETTYOUT RULE === output here>
\```

### Group by file
\```
<paste === PRETTYOUT FILE === output here>
\```

## What prettyout improves

- **Consistent format**: semgrep has a unique rich output → prettyout produces the same structure as all other supported tools
- **Rule code from check_id**: raw JSON uses `check_id` field as the rule identifier → prettyout uses it directly as the rule code for clean display
- **Severity normalized**: semgrep severity comes from `extra.severity` and may be uppercase or absent → prettyout normalizes and shows `[ERROR]` / `[WARN]` / `[INFO]` in the rule header
- **Rule grouping**: raw output lists matches one by one → group-by-rule shows all files affected by the same check
- **Summary line**: raw output ends with stats about rules run → `N issues · M rules · K files` is clean and consistent
```

- [ ] **Commit**

```bash
git add docs/tools/semgrep.md
git commit -m "docs: add semgrep showcase"
```

---

## Task 15: Write `docs/summary.md`

**Files:** Create `docs/summary.md`

- [ ] **Write `docs/summary.md`**

```markdown
# prettyout — Output Format Showcase

prettyout is a thin wrapper that intercepts CLI linters and scanners, runs them with JSON output, and formats the result as a grouped, human-readable summary. Install once, then use your tools exactly as before — the output format changes, nothing else does.

All 14 supported tools produce the same output structure: violations grouped by rule (or file), with occurrence counts, severity labels, and a summary line. This makes it easy to spot patterns across a large codebase regardless of which tool found them.

---

## Tools

| Tool | What it checks | Docs |
|------|---------------|------|
| ruff | Python linter for style, errors, and imports | [ruff.md](tools/ruff.md) |
| mypy | Static type checker for Python | [mypy.md](tools/mypy.md) |
| basedpyright | Strict static type checker for Python | [basedpyright.md](tools/basedpyright.md) |
| bandit | Python security linter | [bandit.md](tools/bandit.md) |
| pylint | Python linter for errors, style, and code smells | [pylint.md](tools/pylint.md) |
| eslint | JavaScript and TypeScript linter | [eslint.md](tools/eslint.md) |
| biome | Fast formatter and linter for JS/TS/JSON | [biome.md](tools/biome.md) |
| stylelint | CSS/SCSS/Less linter | [stylelint.md](tools/stylelint.md) |
| shellcheck | Shell script linter | [shellcheck.md](tools/shellcheck.md) |
| hadolint | Dockerfile linter | [hadolint.md](tools/hadolint.md) |
| golangci-lint | Go meta-linter | [golangci-lint.md](tools/golangci-lint.md) |
| cargo clippy | Rust linter | [cargo-clippy.md](tools/cargo-clippy.md) |
| trivy | Security vulnerability scanner | [trivy.md](tools/trivy.md) |
| semgrep | Semantic code analysis | [semgrep.md](tools/semgrep.md) |

---

## What prettyout improves

| Problem | Tools affected | prettyout solution |
|---------|---------------|-------------------|
| Inconsistent output format between tools | All | Uniform grouped format across all 14 tools |
| Line-by-line output (same rule repeated per occurrence) | All | Group by rule or file — see all affected locations at a glance |
| No violation count summary | All | `N issues · M rules · K files` always at the end |
| JSON only readable by machines | All | Human-readable with colors |
| Severity not shown | basedpyright, mypy, semgrep | `[ERROR]` / `[WARN]` / `[INFO]` prefix in rule header |
| Duplicate violations | basedpyright, pylint | Set-based deduplication |
| Wrong singular/plural | Many | `1 file` not `1 files`, `line 5` not `lines 5` |
| JSON on wrong stream | stylelint (stderr) | Handled transparently |
| Null/missing fields in JSON | golangci-lint, cargo clippy, biome | All edge cases handled gracefully |

---

## Group by rule vs group by file

**Use group-by-rule** (the default) when you are doing a refactoring pass: "fix all `F401` unused import warnings across the project." The rule-grouped view shows every file where that rule fires, so you can work through them systematically.

**Use group-by-file** when doing a code review: "what's wrong with this file I just changed?" The file-grouped view shows every violation in a given file in one block, making it easy to address everything in one pass.

Switch between them by adding `.prettyout.yaml` to your project root:

\```yaml
settings:
  ruff:
    group_by: file
\```
```

- [ ] **Commit**

```bash
git add docs/summary.md
git commit -m "docs: add prettyout showcase summary"
```

---

## Self-Review

**Spec coverage:**
- [x] All 14 tools have tasks
- [x] Each task captures 4 outputs (default, JSON, prettyout-rule, prettyout-file)
- [x] Each .md has: what it checks, without prettyout (default + JSON), with prettyout (rule + file), what prettyout improves (3-5 bullets)
- [x] summary.md has: intro, tools table, comparison table, group-by explanation
- [x] All fixture paths are correct and match `test/fixtures/<tool>/`
- [x] stylelint stderr piping handled correctly
- [x] golangci-lint uses `errors/` subdirectory
- [x] cargo clippy uses `errors/` subdirectory

**No placeholders:** All bullets contain concrete problem descriptions and solutions.

**Type consistency:** All plugin names match `cmd/` directory names (prettyout-golangci, prettyout-cargo-clippy, etc.).
