# Remaining Plugins Integration Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** For each of the 12 remaining plugins (mypy, bandit, pylint, eslint, biome, stylelint, shellcheck, hadolint, golangci-lint, cargo-clippy, trivy, semgrep) — run real tool output through the plugin inside Docker, find and fix bugs, verify all scenarios from CLAUDE.md.

---

## Progress (as of 2026-05-17)

**Status: PAUSED after Task 9. 79 PASS, 0 FAIL in test/run.sh.**

### Completed tasks
- ✅ Task 0: Docker baseline (image built, golang:1.23→1.25, Node 18→20 in Dockerfile)
- ✅ Task 1: mypy — fixed `lines N`→`line N`, line dedup
- ✅ Task 2: bandit — fixed ANSI leak for LOW severity
- ✅ Task 3: pylint — fixed dedup, singular line, summary count
- ✅ Task 4: eslint — fixed broken eslint.config (@eslint/js import), singular line
- ✅ Task 5: biome — fixed `location.path` type (string not struct), `description`→`message`
- ✅ Task 6: stylelint — fixed JSON on stderr (not stdout), Node 20, ResolvePath for absolute paths
- ✅ Task 7: shellcheck — clean fixture fix only (no plugin bugs)
- ✅ Task 8: hadolint — clean fixture fix only (no plugin bugs)
- ✅ Task 9: golangci-lint — graceful empty input, go.mod fix, linter flag `--disable-all --enable=ineffassign`

### Remaining tasks
- ⬜ Task 10: cargo-clippy — deeper testing (multi-span, filtering, rustc vs clippy codes)
- ⬜ Task 11: trivy — full checklist (no file/line model)
- ⬜ Task 12: semgrep — add to Dockerfile, test, fix
- ⬜ Task 13: Final integration run + update CLAUDE.md bugs table

### Notes for resuming
- Image `prettyout-test` may need rebuild (Dockerfile changed for Node 20): `docker build -t prettyout-test -f test/Dockerfile .`
- Branch: `feature/plugin-coverage`
- Start with Task 10 (cargo-clippy) — use subagent-driven-development skill

**Architecture:** Docker container has all tools pre-installed and plugins pre-built. We mount the project source (`-v $(pwd):/project`) so plugin code edits on the host can be rebuilt inside the container without rebuilding the image. Each task follows the same research → verify → fix → commit cycle from CLAUDE.md.

**Tech Stack:** Go (plugins), Docker (isolation), bash (test scripts). All tools run in `prettyout-test` Docker image (`test/Dockerfile`).

---

## Shared Workflow (read before every task)

Every task below follows this exact pattern. Internalize it.

**Step A — Start the container:**
```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker build -t prettyout-test -f test/Dockerfile .   # only if image not yet built
docker run -it --rm -v $(pwd):/project prettyout-test bash
```

**Step B — Rebuild the plugin inside the container** (after any code fix on host):
```bash
go build -o /usr/local/bin/prettyout-<tool> /project/cmd/prettyout-<tool>/
```

**Step C — Run the tool and inspect raw JSON** (inside container):
```bash
<tool> <json-flag> <fixture-file> 2>/dev/null | cat     # raw JSON
<tool> <json-flag> <fixture-file> 2>/dev/null | prettyout-<tool>  # plugin output
```

**Step D — Check every item from the CLAUDE.md checklist:**
- [ ] `group_by: rule` — rules sorted, lines collapsed per file (`lines 3, 7` not separate)
- [ ] `group_by: file` — files sorted, rules+lines listed under each file
- [ ] Occurrence count in header: `RULE (N) — message`
- [ ] Summary line: `N issues · M rules · K files` (singular/plural correct)
- [ ] Paths: relative from CWD, not absolute, not just basename
- [ ] Singular/plural: `line 5` not `lines 5`, `1 issue` not `1 issues`
- [ ] Clean run → `0 issues · 0 rules · 0 files`
- [ ] Empty file → no crash
- [ ] Syntax error → shows correctly, no crash
- [ ] Same-line dedup → line shown once, not repeated
- [ ] Empty stdin → error to stderr, exit 1 (NDJSON tools: `0 issues` acceptable)
- [ ] Invalid JSON → error to stderr, exit 1
- [ ] Severity prefix shown if tool reports it: `[ERROR]`, `[WARN]`, `[INFO]`
- [ ] `colors: false` → no ANSI codes in output
- [ ] `max_message_length: 20` → messages truncated with `...`

**Step E — Update `test/run.sh`** with concrete assertions for the tool.

**Step F — Commit.**

---

## Task 0: Build Docker image and verify baseline

**Files:** `test/Dockerfile`, `test/run.sh`

- [ ] **Build the image:**
```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
docker build -t prettyout-test -f test/Dockerfile .
```
Expected: build completes, all tools and plugins installed.

- [ ] **Run baseline test to see starting state:**
```bash
docker run --rm prettyout-test /project/test/run.sh
```
Note which tools show PASS/FAIL/SKIP. This is the baseline.

- [ ] **Commit baseline result** (no code changes, just verify image builds):
```bash
git add test/
git commit -m "test: verify Docker image builds and baseline run.sh passes"
```

---

## Task 1: mypy

**Files:**
- Modify: `cmd/prettyout-mypy/main.go` (fix bugs found)
- Modify: `test/run.sh` (add assertions)

**Fixtures to create inside container:**
```python
# /tmp/t-mypy/errors.py
x: int = "not an int"

def add(a: int, b: int) -> int:
    return a + b

result = add("hello", 42)

def returns_int() -> int:
    pass  # missing return

def note_parent() -> str:
    return 42  # type error
```
```python
# /tmp/t-mypy/clean.py
def add(a: int, b: int) -> int:
    return a + b

result: int = add(1, 2)
```
```python
# /tmp/t-mypy/same_line.py
x: int = "a"; y: int = "b"  # two errors same line
```
```python
# /tmp/t-mypy/empty.py
```

- [ ] **Start container and create fixtures:**
```bash
docker run -it --rm -v $(pwd):/project prettyout-test bash
mkdir -p /tmp/t-mypy && cd /tmp/t-mypy
# paste fixture files above
```

- [ ] **Capture raw JSON and study it:**
```bash
mypy --output=json errors.py 2>/dev/null | python3 -m json.tool | head -60
```
Key things to verify in the JSON:
- Is `code` always present or can it be null/missing?
- What `severity` values appear (`error`, `warning`, `note`)?
- Are `note`-severity lines tied to the preceding error?
- What does `file` look like (relative? absolute?)?
- What does an empty/clean run output? (mypy exits 0 with no output lines)

- [ ] **Run through plugin and note every discrepancy:**
```bash
mypy --output=json errors.py 2>/dev/null | prettyout-mypy
mypy --output=json clean.py 2>/dev/null | prettyout-mypy    # expect: 0 issues
mypy --output=json same_line.py 2>/dev/null | prettyout-mypy # expect: no dup lines
echo -n "" | prettyout-mypy                                   # expect: 0 issues (NDJSON)
echo "not json" | prettyout-mypy                              # expect: 0 issues (NDJSON skips)
```

- [ ] **Run full CLAUDE.md checklist** (Step D above). Fix any failures in `/project/cmd/prettyout-mypy/main.go` on the host.

- [ ] **Test config options:**
```bash
printf 'settings:\n  mypy:\n    group_by: file\n' > .prettyout.yaml
mypy --output=json errors.py 2>/dev/null | prettyout-mypy
printf 'settings:\n  mypy:\n    colors: false\n' > .prettyout.yaml
mypy --output=json errors.py 2>/dev/null | prettyout-mypy | cat
printf 'settings:\n  mypy:\n    max_message_length: 15\n' > .prettyout.yaml
mypy --output=json errors.py 2>/dev/null | prettyout-mypy
```

- [ ] **Update `test/run.sh` mypy section** with assertions based on real output observed.

- [ ] **Rebuild plugin and re-run run.sh to confirm PASS:**
```bash
go build -o /usr/local/bin/prettyout-mypy /project/cmd/prettyout-mypy/
/project/test/run.sh 2>&1 | grep -A1 "mypy"
```

- [ ] **Commit:**
```bash
git add cmd/prettyout-mypy/main.go test/run.sh
git commit -m "fix(mypy): bugs found via real tool testing in Docker"
```

---

## Task 2: bandit

**Files:**
- Modify: `cmd/prettyout-bandit/main.go`
- Modify: `test/run.sh`

**Fixtures:**
```python
# /tmp/t-bandit/errors.py
import hashlib
import subprocess
import os

hashlib.md5(b"data")
subprocess.call("ls -la", shell=True)
password = "hunter2"
os.system("rm -rf /tmp/test")
```
```python
# /tmp/t-bandit/clean.py
import hashlib

hashlib.sha256(b"data")
```
```python
# /tmp/t-bandit/empty.py
```

- [ ] **Capture raw JSON:**
```bash
bandit -f json errors.py 2>/dev/null | python3 -m json.tool
bandit -f json clean.py 2>/dev/null | python3 -m json.tool   # what does clean look like?
bandit -f json empty.py 2>/dev/null | python3 -m json.tool   # what does empty look like?
```
Key things to verify:
- Does `filename` have leading `./`? (our plugin strips it — check this works)
- Is `issue_severity` always `HIGH`/`MEDIUM`/`LOW` uppercase?
- Is `issue_confidence` always present?
- What does the `errors` array look like for unparseable files?
- What does clean output look like? (`{"results": []}`)

- [ ] **Run through plugin and check all CLAUDE.md items** (Step D).

- [ ] **Test confidence display:** bandit plugin shows `B324 (MEDIUM/HIGH)` — verify this format appears.

- [ ] **Update `test/run.sh` bandit section.**

- [ ] **Rebuild and verify:**
```bash
go build -o /usr/local/bin/prettyout-bandit /project/cmd/prettyout-bandit/
/project/test/run.sh 2>&1 | grep -A1 "bandit"
```

- [ ] **Commit:**
```bash
git add cmd/prettyout-bandit/main.go test/run.sh
git commit -m "fix(bandit): bugs found via real tool testing in Docker"
```

---

## Task 3: pylint

**Files:**
- Modify: `cmd/prettyout-pylint/main.go`
- Modify: `test/run.sh`

**Fixtures:**
```python
# /tmp/t-pylint/errors.py
import os
import sys

x = undefined_variable
very_long_function_name_that_exceeds_limits(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

class MyClass:
    def method_without_docstring(self, a, b, c, d, e, f):
        return a+b
```
```python
# /tmp/t-pylint/clean.py
"""A clean module."""


def add(a: int, b: int) -> int:
    """Add two numbers."""
    return a + b
```
```python
# /tmp/t-pylint/empty.py
```

- [ ] **Capture raw JSON:**
```bash
pylint --output-format=json errors.py 2>/dev/null | python3 -m json.tool | head -60
pylint --output-format=json clean.py 2>/dev/null | python3 -m json.tool
pylint --output-format=json empty.py 2>/dev/null | python3 -m json.tool
```
Key things to verify:
- Is `message-id` always present (or can it be missing)?
- What `type` values appear? (`error`, `warning`, `convention`, `refactor`, `fatal`)
- Is `path` relative or absolute?
- What does pylint output for a completely clean file? (might still warn about style)
- Does `endLine`/`endColumn` exist or is it `null`?

- [ ] **Verify `message-id/symbol` display:** plugin shows `E0602/undefined-variable` — confirm format correct.

- [ ] **Run full CLAUDE.md checklist** (Step D).

- [ ] **Update `test/run.sh`** pylint section.

- [ ] **Rebuild and verify:**
```bash
go build -o /usr/local/bin/prettyout-pylint /project/cmd/prettyout-pylint/
/project/test/run.sh 2>&1 | grep -A1 "pylint"
```

- [ ] **Commit:**
```bash
git add cmd/prettyout-pylint/main.go test/run.sh
git commit -m "fix(pylint): bugs found via real tool testing in Docker"
```

---

## Task 4: eslint

**Files:**
- Modify: `cmd/prettyout-eslint/main.go`
- Modify: `test/run.sh`

**Fixtures:**
```javascript
// /tmp/t-eslint/eslint.config.mjs  (flat config, eslint 9+)
import js from "@eslint/js";
export default [
  js.configs.recommended,
  { rules: { "no-console": "warn", "eqeqeq": "error" } }
];
```
```javascript
// /tmp/t-eslint/errors.js
var x = undefined_var
if (x == "hello") { console.log(x) }
```
```javascript
// /tmp/t-eslint/warnings_only.js
console.log("hello")   // no-console = warn, nothing else
```
```javascript
// /tmp/t-eslint/clean.js
const x = 1;
if (x === 1) {
  process.stdout.write("hello\n");
}
```
```javascript
// /tmp/t-eslint/parse_error.js
const x = {   // unclosed brace — parse error
```

- [ ] **Capture raw JSON:**
```bash
eslint --format=json errors.js 2>/dev/null | python3 -m json.tool
eslint --format=json warnings_only.js 2>/dev/null | python3 -m json.tool
eslint --format=json clean.js 2>/dev/null | python3 -m json.tool
eslint --format=json parse_error.js 2>/dev/null | python3 -m json.tool
```
Key things to verify:
- `filePath` — is it absolute? (our plugin calls `ResolvePath` — check it works)
- `ruleId` — can it be `null`? (parse errors have null ruleId)
- `severity`: 1=warning, 2=error — verify mapping correct
- File with 0 messages — does eslint include it in the array or omit it?
- What does `parse_error.js` look like in JSON? (`fatal: true`, `ruleId: null`)

- [ ] **Run full CLAUDE.md checklist** (Step D).

- [ ] **Test mixed severity** (warnings + errors in same run):
```bash
eslint --format=json errors.js warnings_only.js 2>/dev/null | prettyout-eslint
```
Check: severity prefix shows correctly per rule.

- [ ] **Update `test/run.sh`** eslint section.

- [ ] **Rebuild and verify:**
```bash
go build -o /usr/local/bin/prettyout-eslint /project/cmd/prettyout-eslint/
/project/test/run.sh 2>&1 | grep -A1 "eslint"
```

- [ ] **Commit:**
```bash
git add cmd/prettyout-eslint/main.go test/run.sh
git commit -m "fix(eslint): bugs found via real tool testing in Docker"
```

---

## Task 5: biome

**Files:**
- Modify: `cmd/prettyout-biome/main.go`
- Modify: `test/run.sh`

**Fixtures:**
```json
// /tmp/t-biome/biome.json
{
  "linter": {
    "enabled": true,
    "rules": { "recommended": true }
  }
}
```
```typescript
// /tmp/t-biome/errors.ts
var x = 1
if (x == "1") { console.log("bad") }
debugger;
```
```typescript
// /tmp/t-biome/clean.ts
const x: number = 1;
if (x === 1) {
  console.log("good");
}
```
```typescript
// /tmp/t-biome/empty.ts
```

- [ ] **Capture raw JSON:**
```bash
biome check --reporter=json errors.ts 2>/dev/null | python3 -m json.tool
biome check --reporter=json clean.ts 2>/dev/null | python3 -m json.tool
biome check --reporter=json empty.ts 2>/dev/null | python3 -m json.tool
```
Key things to verify:
- `location.path.file` — is it relative or absolute? Relative to what?
- `location.span` — byte offsets only, no line numbers. Verify plugin handles "no line" correctly (shows just filename, no "line N")
- `severity` field — what values appear? (`"error"` | `"warning"` | `"information"` | `"hint"`)
- What does a clean run look like? Does `diagnostics` array exist and is empty, or key is missing?
- `category` — what format? (`lint/suspicious/noDoubleEquals` — does our plugin show full or truncated?)

- [ ] **Run full CLAUDE.md checklist** (Step D). Note: biome has no line numbers, so checklist item about `line N` display must verify "filename only" output.

- [ ] **Update `test/run.sh`** biome section.

- [ ] **Rebuild and verify:**
```bash
go build -o /usr/local/bin/prettyout-biome /project/cmd/prettyout-biome/
/project/test/run.sh 2>&1 | grep -A1 "biome"
```

- [ ] **Commit:**
```bash
git add cmd/prettyout-biome/main.go test/run.sh
git commit -m "fix(biome): bugs found via real tool testing in Docker"
```

---

## Task 6: stylelint

**Files:**
- Modify: `cmd/prettyout-stylelint/main.go`
- Modify: `test/run.sh`

**Fixtures:**
```json
// /tmp/t-stylelint/.stylelintrc.json
{ "extends": "stylelint-config-standard" }
```
```css
/* /tmp/t-stylelint/errors.css */
a { color: #gggggg; FONT-SIZE: 12px }
.foo { color:red }
```
```css
/* /tmp/t-stylelint/warnings.css */
a { color: pink }
```
```css
/* /tmp/t-stylelint/clean.css */
a {
  color: #fff;
  font-size: 12px;
}
```
```css
/* /tmp/t-stylelint/parse_error.css */
a { color: {{{broken
```
```css
/* /tmp/t-stylelint/empty.css */
```

- [ ] **Capture raw JSON:**
```bash
stylelint --formatter=json errors.css 2>/dev/null | python3 -m json.tool
stylelint --formatter=json clean.css 2>/dev/null | python3 -m json.tool
stylelint --formatter=json parse_error.css 2>/dev/null | python3 -m json.tool
stylelint --formatter=json empty.css 2>/dev/null | python3 -m json.tool
```
Key things to verify:
- `warnings` array — does it contain both errors and warnings (confusing name)?
- `parseErrors` array — what does it look like? Does it appear in the `warnings` array too or only in `parseErrors`?
- `source` field — relative or absolute path?
- `severity` field on each warning — `"error"` or `"warning"`?
- `errored` field on file object — boolean, when is it true?

- [ ] **Run full CLAUDE.md checklist** (Step D).

- [ ] **Verify parse error handling:** `parseErrors` non-empty should produce a `parse-error` rule entry.

- [ ] **Update `test/run.sh`** stylelint section.

- [ ] **Rebuild and verify:**
```bash
go build -o /usr/local/bin/prettyout-stylelint /project/cmd/prettyout-stylelint/
/project/test/run.sh 2>&1 | grep -A1 "stylelint"
```

- [ ] **Commit:**
```bash
git add cmd/prettyout-stylelint/main.go test/run.sh
git commit -m "fix(stylelint): bugs found via real tool testing in Docker"
```

---

## Task 7: shellcheck

**Files:**
- Modify: `cmd/prettyout-shellcheck/main.go`
- Modify: `test/run.sh`

**Fixtures:**
```bash
# /tmp/t-shellcheck/errors.sh
#!/bin/bash
x=`echo hello`
if [ $x == "hello" ]; then
    echo $x
fi
for f in $(ls /tmp); do
    echo $f
done
```
```bash
# /tmp/t-shellcheck/warnings.sh
#!/bin/bash
# SC2034: variable assigned but not used
unused_var="hello"
echo "done"
```
```bash
# /tmp/t-shellcheck/clean.sh
#!/bin/bash
x=$(echo hello)
if [ "$x" = "hello" ]; then
    echo "$x"
fi
```
```bash
# /tmp/t-shellcheck/empty.sh
```

- [ ] **Capture raw JSON:**
```bash
shellcheck --format=json errors.sh 2>/dev/null | python3 -m json.tool
shellcheck --format=json clean.sh 2>/dev/null | python3 -m json.tool
shellcheck --format=json empty.sh 2>/dev/null | python3 -m json.tool
shellcheck --format=json errors.sh warnings.sh 2>/dev/null | python3 -m json.tool
```
Key things to verify:
- `code` — integer. Verify plugin formats as `SC2006` (not `2006`)
- `level` — what values appear? (`error`, `warning`, `info`, `style`)
- `file` — relative or absolute?
- Clean/empty file — does shellcheck output `[]` or nothing?
- Multiple files — is the array flat (all issues together) or nested by file?

- [ ] **Run full CLAUDE.md checklist** (Step D).

- [ ] **Verify SC prefix:** output must show `SC2006`, not bare `2006`.

- [ ] **Verify `style` and `info` levels** show correct `[INFO]` prefix with blue color.

- [ ] **Update `test/run.sh`** shellcheck section.

- [ ] **Rebuild and verify:**
```bash
go build -o /usr/local/bin/prettyout-shellcheck /project/cmd/prettyout-shellcheck/
/project/test/run.sh 2>&1 | grep -A1 "shellcheck"
```

- [ ] **Commit:**
```bash
git add cmd/prettyout-shellcheck/main.go test/run.sh
git commit -m "fix(shellcheck): bugs found via real tool testing in Docker"
```

---

## Task 8: hadolint

**Files:**
- Modify: `cmd/prettyout-hadolint/main.go`
- Modify: `test/run.sh`

**Fixtures:**
```dockerfile
# /tmp/t-hadolint/Dockerfile.errors
FROM ubuntu
RUN apt-get install vim
RUN apt-get install curl
ADD . /app
```
```dockerfile
# /tmp/t-hadolint/Dockerfile.warnings
FROM ubuntu:24.04
RUN wget https://example.com/file.sh | bash
```
```dockerfile
# /tmp/t-hadolint/Dockerfile.clean
FROM ubuntu:24.04
RUN apt-get update && apt-get install -y --no-install-recommends \
        vim \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY . .
```
```dockerfile
# /tmp/t-hadolint/Dockerfile.empty
FROM scratch
```

- [ ] **Capture raw JSON:**
```bash
hadolint --format=json Dockerfile.errors 2>/dev/null | python3 -m json.tool
hadolint --format=json Dockerfile.clean 2>/dev/null | python3 -m json.tool
hadolint --format=json Dockerfile.empty 2>/dev/null | python3 -m json.tool
hadolint --format=json Dockerfile.errors Dockerfile.warnings 2>/dev/null | python3 -m json.tool
```
Key things to verify:
- `file` field — is it the filename as given on CLI? Relative or absolute?
- `level` — values: `error`, `warning`, `info`, `style`, `ignore`? Does `ignore` actually appear?
- `code` — is it always `DL####` or can it be `SC####` (shellcheck codes embedded)?
- Clean Dockerfile — outputs `[]` or nothing?
- Multiple files — flat array?

- [ ] **Run full CLAUDE.md checklist** (Step D). Verify `ignore` level items are skipped.

- [ ] **Update `test/run.sh`** hadolint section.

- [ ] **Rebuild and verify:**
```bash
go build -o /usr/local/bin/prettyout-hadolint /project/cmd/prettyout-hadolint/
/project/test/run.sh 2>&1 | grep -A1 "hadolint"
```

- [ ] **Commit:**
```bash
git add cmd/prettyout-hadolint/main.go test/run.sh
git commit -m "fix(hadolint): bugs found via real tool testing in Docker"
```

---

## Task 9: golangci-lint

**Files:**
- Modify: `cmd/prettyout-golangci/main.go`
- Modify: `test/run.sh`

**Fixtures:**
```go
// /tmp/t-golangci/main.go
package main

import (
	"errors"
	"fmt"
	"os"
)

func riskyOp() error {
	return errors.New("fail")
}

func main() {
	riskyOp()          // error return ignored — errcheck
	x := 1
	fmt.Println("hi")
	os.Exit(0)
	_ = x
}
```
```go
// /tmp/t-golangci/go.mod
module golangci_test

go 1.21
```
```go
// /tmp/t-golangci/clean.go  (separate package to avoid conflicts)
package main  // same package — just add to main.go above instead
```

- [ ] **Capture raw JSON:**
```bash
cd /tmp/t-golangci
golangci-lint run --out-format=json 2>/dev/null | python3 -m json.tool | head -80
```
Key things to verify:
- `Issues` field — is it `null` or `[]` when clean? (spec said it can be `null` — verify!)
- `FromLinter` — linter name format (e.g. `errcheck`, `unused`)
- `Pos.Filename` — relative to module root or absolute?
- `Severity` field — is it always empty string or can it be `error`/`warning`?
- `Text` — message format
- What linters are enabled by default? (errcheck, govet, staticcheck, etc.)

- [ ] **Test clean run** (golangci-lint needs a module with no issues):
```bash
cat > /tmp/t-golangci-clean/main.go << 'GO'
package main

import "fmt"

func main() {
    fmt.Println("hello")
}
GO
cat > /tmp/t-golangci-clean/go.mod << 'MOD'
module clean_test
go 1.21
MOD
cd /tmp/t-golangci-clean && golangci-lint run --out-format=json 2>/dev/null | prettyout-golangci
```

- [ ] **Verify null Issues handling:** if `Issues` is `null` in JSON, plugin must output `0 issues`, not crash.

- [ ] **Run full CLAUDE.md checklist** (Step D).

- [ ] **Update `test/run.sh`** golangci section.

- [ ] **Rebuild and verify:**
```bash
go build -o /usr/local/bin/prettyout-golangci /project/cmd/prettyout-golangci/
/project/test/run.sh 2>&1 | grep -A1 "golangci"
```

- [ ] **Commit:**
```bash
git add cmd/prettyout-golangci/main.go test/run.sh
git commit -m "fix(golangci): bugs found via real tool testing in Docker"
```

---

## Task 10: cargo-clippy (deepen testing)

**Files:**
- Modify: `cmd/prettyout-cargo-clippy/main.go` (if bugs found)
- Modify: `test/run.sh`

cargo-clippy was partially tested locally (we verified basic structure). This task deepens testing inside Docker with more edge cases.

**Fixtures:**
```rust
// /tmp/t-cargo/src/main.rs  (errors)
fn main() {
    let x = vec![1, 2, 3];
    let _ = x;
    let mut v = Vec::new();
    v.push(1);
    v.push(2);
    return;
}
```
```rust
// /tmp/t-cargo/src/no_code.rs  (warnings only, no clippy codes)
#[allow(dead_code)]
fn unused_function() {}
```
```rust
// /tmp/t-cargo-clean/src/main.rs  (clean)
fn main() {
    println!("Hello!");
}
```

- [ ] **Test multi-span issue** (vec_init_then_push spans multiple lines):
```bash
cargo clippy --message-format=json 2>/dev/null | python3 -c "
import json, sys
for line in sys.stdin:
    obj = json.loads(line)
    if obj.get('reason') == 'compiler-message':
        msg = obj['message']
        print(msg.get('code'), [s['line_start'] for s in msg.get('spans',[])])
"
```
Verify: plugin shows `line 4` (first line of span), not `lines 4, 5, 6`.

- [ ] **Test compiler-artifact and build-finished filtering:**
```bash
cargo clippy --message-format=json 2>/dev/null | grep '"reason"' | sort | uniq -c
```
Verify: only `compiler-message` entries processed; `compiler-artifact` and `build-finished` skipped.

- [ ] **Test rustc vs clippy codes:**
```bash
cargo clippy --message-format=json 2>/dev/null | python3 -c "
import json, sys
for line in sys.stdin:
    obj = json.loads(line)
    if obj.get('reason') == 'compiler-message':
        code = obj['message'].get('code')
        print(code)
"
```
Verify: `unused_variables` (rustc, no clippy:: prefix) shows as `unused_variables`, not `clippy::unused_variables`.

- [ ] **Run full CLAUDE.md checklist** (Step D).

- [ ] **Update `test/run.sh`** cargo section with additional assertions.

- [ ] **Commit if any fixes made:**
```bash
git add cmd/prettyout-cargo-clippy/main.go test/run.sh
git commit -m "fix(cargo-clippy): deeper testing in Docker, fix edge cases"
```

---

## Task 11: trivy

**Files:**
- Modify: `cmd/prettyout-trivy/main.go`
- Modify: `test/run.sh`

trivy is a security scanner with a fundamentally different output model (no file/line, groups by severity/CVE).

**Test scenarios:**
```bash
# Scan a known-vulnerable image (small, fast)
trivy image --format=json --quiet alpine:3.11 2>/dev/null | python3 -m json.tool | head -100

# Scan a clean directory (no vulns)
mkdir -p /tmp/clean-dir
trivy fs --format=json --quiet /tmp/clean-dir 2>/dev/null | python3 -m json.tool

# Scan a directory with a requirements.txt containing known vuln
mkdir -p /tmp/t-trivy-fs
echo "django==2.0" > /tmp/t-trivy-fs/requirements.txt
trivy fs --format=json --quiet /tmp/t-trivy-fs 2>/dev/null | python3 -m json.tool
```

- [ ] **Capture and study JSON structure:**
Key things to verify:
- `Results` — array; each entry has `Target`, `Type`, `Vulnerabilities`
- `Vulnerabilities` — can it be `null` (not `[]`) when no vulns for that target?
- `Severity` — `CRITICAL`, `HIGH`, `MEDIUM`, `LOW`, `UNKNOWN` — case always uppercase?
- `FixedVersion` — can it be empty string `""` when no fix available?
- `VulnerabilityID` — always a CVE ID or can it be other formats?
- For filesystem scan: does `Results` array exist even when empty?

- [ ] **Run through plugin:**
```bash
trivy image --format=json --quiet alpine:3.11 2>/dev/null | prettyout-trivy
trivy fs --format=json --quiet /tmp/clean-dir 2>/dev/null | prettyout-trivy
```

- [ ] **Verify output format:**
```
CRITICAL (N vulnerabilities)
  CVE-XXXX-YYYY — package version → fix: fixed_version
HIGH (N vulnerabilities)
  ...
```
Check: correct severity ordering (CRITICAL > HIGH > MEDIUM > LOW > UNKNOWN), "no fix available" shown for empty `FixedVersion`.

- [ ] **Run full CLAUDE.md checklist** — note: trivy has no file/line model, so group_by config does not apply. Verify plugin handles group_by config gracefully (ignores it or has its own grouping).

- [ ] **Update `test/run.sh`** trivy section.

- [ ] **Rebuild and verify:**
```bash
go build -o /usr/local/bin/prettyout-trivy /project/cmd/prettyout-trivy/
/project/test/run.sh 2>&1 | grep -A1 "trivy"
```

- [ ] **Commit:**
```bash
git add cmd/prettyout-trivy/main.go test/run.sh
git commit -m "fix(trivy): bugs found via real tool testing in Docker"
```

---

## Task 12: semgrep

semgrep is not yet in the Dockerfile. This task adds it, then tests.

**Files:**
- Modify: `test/Dockerfile` (add semgrep)
- Modify: `cmd/prettyout-semgrep/main.go`
- Modify: `test/run.sh`

- [ ] **Add semgrep to Dockerfile** — find the Python tools section and add:
```dockerfile
RUN uv tool install semgrep && \
    go build -o /usr/local/bin/prettyout-semgrep /project/cmd/prettyout-semgrep/
```
Also add the build line to the plugin build section at bottom of Dockerfile.

- [ ] **Rebuild Docker image:**
```bash
docker build -t prettyout-test -f test/Dockerfile .
```

**Fixtures:**
```python
# /tmp/t-semgrep/errors.py
import subprocess
user_input = input("cmd: ")
subprocess.call(user_input, shell=True)   # command injection

password = "hardcoded_secret_123"
```
```python
# /tmp/t-semgrep/clean.py
import subprocess
subprocess.call(["ls", "-la"])
```

- [ ] **Find a fast semgrep rule set** (avoid downloading large rule packs):
```bash
# Use a minimal local rule instead of the full registry
cat > /tmp/semgrep-rules.yaml << 'YAML'
rules:
  - id: hardcoded-password
    pattern: $X = "..."
    message: Possible hardcoded secret
    languages: [python]
    severity: WARNING
YAML
semgrep --config /tmp/semgrep-rules.yaml --json errors.py 2>/dev/null | python3 -m json.tool
```

- [ ] **Capture raw JSON and study it:**
Key things to verify:
- `results` array — field names: `check_id`, `path`, `start.line`, `extra.message`, `extra.severity`
- `extra.severity` — uppercase `ERROR`/`WARNING`/`INFO` or lowercase?
- `path` — relative or absolute?
- `errors` array — when does it have entries?
- Clean run — `"results": []` and `"errors": []`?

- [ ] **Run through plugin:**
```bash
semgrep --config /tmp/semgrep-rules.yaml --json errors.py 2>/dev/null | prettyout-semgrep
semgrep --config /tmp/semgrep-rules.yaml --json clean.py 2>/dev/null | prettyout-semgrep
```

- [ ] **Run full CLAUDE.md checklist** (Step D).

- [ ] **Update `test/run.sh`** with semgrep section using local rules (not registry — avoids network and auth).

- [ ] **Rebuild plugin and verify:**
```bash
go build -o /usr/local/bin/prettyout-semgrep /project/cmd/prettyout-semgrep/
/project/test/run.sh 2>&1 | grep -A1 "semgrep"
```

- [ ] **Commit:**
```bash
git add test/Dockerfile cmd/prettyout-semgrep/main.go test/run.sh
git commit -m "fix(semgrep): add to Dockerfile, test via Docker, fix bugs"
```

---

## Task 13: Final integration run + update CLAUDE.md bugs table

**Files:**
- Modify: `test/run.sh` (final pass — ensure all sections complete)
- Modify: `CLAUDE.md` (update Common Bugs table with newly found issues)

- [ ] **Run full test suite inside Docker:**
```bash
docker run --rm prettyout-test /project/test/run.sh
```
Expected: all PASS, 0 FAIL.

- [ ] **If any FAIL remain** — fix the plugin, rebuild inside container, re-run until all pass.

- [ ] **Update `CLAUDE.md` Common Bugs table** with any new bugs found during this work. For each bug: what it was, which plugin, what the fix was.

- [ ] **Final commit:**
```bash
git add test/run.sh CLAUDE.md
git commit -m "test: all plugins passing integration tests in Docker"
```

- [ ] **Print summary of all changes made** (which plugins had bugs, what was fixed).
