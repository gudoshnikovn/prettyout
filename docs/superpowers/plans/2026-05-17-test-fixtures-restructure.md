# Test Fixtures Restructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the monolithic `test/run.sh` inline heredoc approach with a proper fixture directory tree + per-tool manifests + snapshot golden files, driven by a Go test runner.

**Architecture:** Each tool gets `test/fixtures/<tool>/` with real committed fixture files, a `manifest.yaml` describing scenarios and assertions, and `.snap` golden files for exact output comparison. A Go binary (`cmd/test-runner/`) reads the manifests, runs each scenario, checks `contains`/`absent` strings, diffs against snapshots. Supports `--update` to regenerate snapshots when tool output intentionally changes.

**Tech Stack:** Go (runner), YAML (manifests), bash (thin `test/run.sh` wrapper), Docker (isolation).

---

## Why this matters

Current `run.sh` has two problems:
1. **Fixtures live in bash heredocs** — no syntax highlighting, no linting, hard to evolve
2. **No golden files** — when a test fails you can't tell if *our plugin broke* or *the tool changed its output format*

With snapshots: if the tool updates and starts reporting a renamed rule, the snapshot diff makes it obvious. If our plugin breaks the formatting structure, the diff shows that instead.

---

## Directory structure (target state)

```
test/
  fixtures/
    ruff/
      manifest.yaml
      errors.py          # F401 (unused import), F811 (redefinition)
      clean.py           # 0 issues
      same-line.py       # two rules on same line — dedup test
      empty.py           # zero-byte
      errors.snap        # golden: prettyout-ruff output for errors.py
      clean.snap
      same-line.snap
      empty.snap
      errors-group-by-file.snap
      errors-no-colors.snap
      errors-max-msg.snap

    mypy/
      manifest.yaml
      errors.py
      clean.py
      same-line.py
      empty.py
      *.snap

    basedpyright/
      manifest.yaml
      errors.py
      clean.py
      syntax-error.py
      *.snap

    bandit/
      manifest.yaml
      errors.py
      clean.py
      empty.py
      *.snap

    pylint/
      manifest.yaml
      errors.py
      clean.py
      empty.py
      *.snap

    eslint/
      manifest.yaml
      eslint.config.mjs  # flat config (ESLint 9+), committed here
      errors.js
      clean.js
      *.snap

    biome/
      manifest.yaml
      biome.json         # minimal config, committed here
      errors.ts
      clean.ts
      empty.ts
      *.snap

    stylelint/
      manifest.yaml
      .stylelintrc.json  # committed here
      errors.css
      clean.css
      empty.css
      *.snap

    shellcheck/
      manifest.yaml
      errors.sh
      clean.sh
      empty.sh
      *.snap

    hadolint/
      manifest.yaml
      errors.Dockerfile
      clean.Dockerfile
      *.snap

    golangci-lint/
      manifest.yaml
      go.mod             # module golangci_test, go 1.21
      errors.go
      clean.go
      *.snap

    cargo-clippy/
      manifest.yaml
      Cargo.toml
      src/
        errors.rs
        clean.rs
      *.snap

    trivy/
      manifest.yaml
      requirements.txt   # django==2.0 — known CVE for fs scan
      # image scan uses alpine:3.11 — no fixture file needed
      *.snap

    semgrep/
      manifest.yaml
      rules.yaml         # local rule definitions (no registry download)
      errors.py
      clean.py
      *.snap

  runner/
    main.go              # Go test runner binary
    manifest.go          # manifest YAML structs
    runner.go            # scenario execution logic
    snapshot.go          # snapshot read/write/diff logic

  run.sh                 # thin wrapper: builds runner, executes it
  Dockerfile             # updated: build runner binary, CMD uses it
```

---

## manifest.yaml format

Each tool has one manifest. The runner reads it and executes each scenario.

```yaml
# test/fixtures/ruff/manifest.yaml

tool: ruff
command: "ruff check --output-format=json"
plugin: prettyout-ruff

# stderr handling: most tools: stderr discarded (2>/dev/null)
# some tools write JSON to stderr (stylelint): stderr_to_stdout: true
stderr_to_stdout: false

scenarios:
  - name: errors
    files: [errors.py]
    contains: ["F401", " · ", " ("]   # quick sanity checks
    absent: []
    snapshot: errors.snap

  - name: clean
    files: [clean.py]
    contains: ["0 issues"]
    snapshot: clean.snap

  - name: same-line-dedup
    files: [same-line.py]
    # output must contain the line number exactly once, not twice
    snapshot: same-line.snap

  - name: empty-file
    files: [empty.py]
    contains: ["0 issues"]
    snapshot: empty.snap

  - name: errors-group-by-file
    files: [errors.py]
    plugin_config:
      group_by: file
    contains: ["errors.py"]
    snapshot: errors-group-by-file.snap

  - name: errors-no-colors
    files: [errors.py]
    plugin_config:
      colors: false
    absent: ["["]             # no ANSI escape codes
    snapshot: errors-no-colors.snap

  - name: errors-max-msg
    files: [errors.py]
    plugin_config:
      max_message_length: 20
    contains: ["..."]              # truncation marker
    snapshot: errors-max-msg.snap
```

**`plugin_config`** fields (any subset): the runner writes a `.prettyout.yaml` before the run and deletes it after.

**`stderr_to_stdout: true`** (stylelint): runner redirects stderr to stdout when piping.

**`command`**: run from the fixture directory as CWD. `{files}` is replaced by space-joined file names from the scenario. If no `{files}` placeholder, files are appended at the end.

---

## Go runner — behaviour

Binary: built at `test/runner/` or compiled into the Docker image.

```
Usage:
  test-runner [flags]

Flags:
  --fixtures   path to fixtures dir (default: ./test/fixtures)
  --tool       run only this tool (default: all)
  --update     regenerate snapshot files instead of comparing
  --no-snap    skip snapshot comparison, only check contains/absent
  --color      force color output (default: auto-detect TTY)
```

**Per-scenario execution:**
1. `cd test/fixtures/<tool>/`
2. If `plugin_config`: write `.prettyout.yaml` with that config
3. Build command: `<command> <files> | <plugin>` (or with stderr redirect)
4. Capture stdout+exit-code
5. Check `contains` — each string must appear in output
6. Check `absent` — each string must NOT appear in output
7. If `snapshot` and not `--update`: diff output against `<tool>/<snapshot>`. Fail on any diff.
8. If `snapshot` and `--update`: write output to `<tool>/<snapshot>`
9. Cleanup `.prettyout.yaml`

**Output format:**
```
══ ruff ══
  PASS  errors
  PASS  clean
  PASS  same-line-dedup
  FAIL  errors-group-by-file
        snapshot diff:
        - F401 (2) — ...
        + F401 (3) — ...

══ mypy ══
  PASS  errors
  SKIP  clean (tool not installed)

══ Results ══
PASS: 41  FAIL: 1  SKIP: 3
```

**SKIP** when tool binary not found (`has_tool` check per manifest).

---

## Snapshot update workflow

When a tool releases a new version and rule codes/messages change:

```bash
# Update snapshots for one tool
docker run --rm -v $(pwd):/project prettyout-test test-runner --tool=ruff --update

# Review the diff
git diff test/fixtures/ruff/

# If change is expected (tool updated): commit
git add test/fixtures/ruff/*.snap
git commit -m "snap(ruff): update snapshots for ruff 0.X.Y"

# If change is unexpected (our plugin broke): fix the plugin, re-run
```

This is how you distinguish "tool changed" (snapshot update is intentional) from "plugin broke" (snapshot diff is a bug).

---

## Task 0: Write the Go runner

**Files:**
- Create: `test/runner/main.go`
- Create: `test/runner/manifest.go`
- Create: `test/runner/runner.go`
- Create: `test/runner/snapshot.go`

- [ ] **Write `manifest.go`** — structs for manifest YAML:

```go
package main

type Manifest struct {
    Tool          string     `yaml:"tool"`
    Command       string     `yaml:"command"`
    Plugin        string     `yaml:"plugin"`
    StderrToStdout bool      `yaml:"stderr_to_stdout"`
    Scenarios     []Scenario `yaml:"scenarios"`
}

type Scenario struct {
    Name         string            `yaml:"name"`
    Files        []string          `yaml:"files"`
    Contains     []string          `yaml:"contains"`
    Absent       []string          `yaml:"absent"`
    Snapshot     string            `yaml:"snapshot"`
    PluginConfig map[string]string `yaml:"plugin_config"`
}
```

- [ ] **Write `runner.go`** — core execution logic:
  - `runScenario(dir string, m Manifest, s Scenario, update bool) Result`
  - Builds shell command: `<command> <files> 2>/dev/null | <plugin>` (or `2>&1 >/dev/null | <plugin>` if `stderr_to_stdout`)
  - Writes/deletes `.prettyout.yaml` when `PluginConfig` is non-empty:
    ```yaml
    settings:
      <tool>:
        <key>: <value>
    ```
  - Returns `Result{Name, Status, FailReason, Output}`

- [ ] **Write `snapshot.go`** — read/write/diff snapshot files:
  - `loadSnapshot(path string) (string, error)`
  - `saveSnapshot(path, content string) error`
  - `diffSnapshot(expected, got string) string` — returns unified diff string, empty if equal

- [ ] **Write `main.go`** — CLI entry point:
  - Parse flags: `--fixtures`, `--tool`, `--update`, `--no-snap`, `--color`
  - Discover manifests: `glob(fixtures + "/*/manifest.yaml")`
  - Check tool is installed (`exec.LookPath`) — SKIP if not found
  - Run scenarios, collect results, print report, exit 1 on any FAIL

- [ ] **Test the runner compiles:**
```bash
cd /path/to/prettyout
go build ./test/runner/
```

- [ ] **Commit:**
```bash
git add test/runner/
git commit -m "feat(test-runner): add Go-based fixture runner with snapshot support"
```

---

## Task 1: Create fixtures for each tool

For every tool, create the fixture files and manifest. Do them one at a time — commit per tool.

**General pattern for each tool:**

- [ ] Create `test/fixtures/<tool>/` directory
- [ ] Write fixture files (errors, clean, edge cases — see below)
- [ ] Write `manifest.yaml`
- [ ] Run the tool in Docker to verify fixtures trigger what's expected:
  ```bash
  docker run --rm -v $(pwd):/project prettyout-test bash -c "
  cd /project/test/fixtures/<tool>
  <tool-command> <files> 2>/dev/null | prettyout-<tool>
  "
  ```
- [ ] Generate initial snapshots with `--update`:
  ```bash
  docker run --rm -v $(pwd):/project prettyout-test bash -c "
  cd /project && test-runner --tool=<tool> --update
  "
  ```
- [ ] Review generated `.snap` files — verify they look correct
- [ ] Commit: `git add test/fixtures/<tool>/ && git commit -m "test(fixtures): add <tool> fixtures and snapshots"`

**Fixtures per tool:**

| Tool | Fixture files | Edge cases |
|------|--------------|------------|
| ruff | errors.py, clean.py, same-line.py, empty.py | same-line dedup, empty |
| mypy | errors.py, clean.py, same-line.py, empty.py | NDJSON, note-level filtered |
| basedpyright | errors.py, clean.py, syntax-error.py, empty.py | dedup, syntax error |
| bandit | errors.py, clean.py, empty.py | severity levels (HIGH/MEDIUM/LOW) |
| pylint | errors.py, clean.py, empty.py | multiple types (error/warning/convention) |
| eslint | eslint.config.mjs, errors.js, clean.js | parse error (null ruleId) |
| biome | biome.json, errors.ts, clean.ts, empty.ts | no line numbers (byte offsets only) |
| stylelint | .stylelintrc.json, errors.css, clean.css, empty.css | JSON on stderr |
| shellcheck | errors.sh, clean.sh, empty.sh | SC prefix formatting |
| hadolint | errors.Dockerfile, clean.Dockerfile | DL + SC codes mixed |
| golangci-lint | go.mod, errors.go | null Issues field, --disable-all --enable=ineffassign |
| cargo-clippy | Cargo.toml, src/errors.rs, src/clean.rs | NDJSON, multi-span |
| trivy | requirements.txt (for fs scan) | no file/line model, severity ordering |
| semgrep | rules.yaml, errors.py, clean.py | local rules only (no registry) |

---

## Task 2: Update Dockerfile and run.sh

**Files:**
- Modify: `test/Dockerfile` — build test-runner binary, use it as CMD
- Modify: `test/run.sh` — thin wrapper around test-runner

- [ ] **Add test-runner build to Dockerfile:**
```dockerfile
# After project COPY:
RUN go build -o /usr/local/bin/test-runner ./test/runner/
```

- [ ] **Update CMD:**
```dockerfile
CMD ["test-runner", "--fixtures", "/project/test/fixtures"]
```

- [ ] **Rewrite `test/run.sh`** to thin wrapper:
```bash
#!/usr/bin/env bash
# Thin wrapper: builds runner and runs it.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
go build -o /tmp/prettyout-test-runner ./test/runner/
/tmp/prettyout-test-runner --fixtures ./test/fixtures "$@"
```

- [ ] **Test in Docker:**
```bash
docker build -t prettyout-test -f test/Dockerfile .
docker run --rm prettyout-test
```

Expected: all PASS, 0 FAIL.

- [ ] **Commit:**
```bash
git add test/Dockerfile test/run.sh
git commit -m "feat(test): switch to fixture-based runner"
```

---

## Task 3: Remove old inline run.sh content

Once the new runner is working and all tools pass:

- [ ] Delete the old inline test code from `test/run.sh` (replaced by thin wrapper above)
- [ ] Verify nothing is lost — every scenario from old `run.sh` is covered by a manifest scenario
- [ ] Final run: `docker run --rm prettyout-test` — all PASS
- [ ] Commit: `git commit -m "test: remove old inline run.sh, fully replaced by fixture runner"`

---

## Snapshot update runbook (for future use)

When a tool releases a new version and tests fail:

```bash
# 1. See what changed
docker run --rm -v $(pwd):/project prettyout-test bash -c "test-runner --tool=ruff"

# 2. If the change is expected (tool updated rules/format):
docker run --rm -v $(pwd):/project prettyout-test bash -c "test-runner --tool=ruff --update"
git diff test/fixtures/ruff/
git add test/fixtures/ruff/*.snap
git commit -m "snap(ruff): update snapshots for ruff X.Y.Z"

# 3. If the change is unexpected (our plugin broke something):
# — fix cmd/prettyout-ruff/main.go, rebuild, re-run without --update
```

---

## Notes

- **golangci-lint**: manifest `command` should use `--disable-all --enable=ineffassign` to avoid Go version mismatch crashes
- **stylelint**: manifest needs `stderr_to_stdout: true`
- **trivy**: `command` uses `trivy fs --format=json --quiet .` (scans the fixture dir for requirements.txt)
- **semgrep**: `command` uses `semgrep --config rules.yaml --json`
- **cargo-clippy**: `command` is `cargo clippy --message-format=json`; NDJSON, runner must not use `python3 -m json.tool`
- **mypy**: NDJSON; clean run produces no output lines → `0 issues` from plugin
- Snapshot files should be committed to git — they are the source of truth for expected output
- `.prettyout.yaml` is gitignored (it's a temp file written by the runner per scenario)
