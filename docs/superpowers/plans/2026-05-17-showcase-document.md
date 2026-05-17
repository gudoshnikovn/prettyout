# Showcase Document: prettyout vs. Raw Tool Output

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce `docs/showcase.md` — a single document showing every supported tool: the fixture code, raw tool output, prettyout-formatted output, and analysis of what prettyout improves. Target audience: developers evaluating prettyout for their team.

**Architecture:** One agent per tool generates the output by running real commands inside Docker, then assembles the markdown. A final agent stitches the per-tool sections into one document.

**Tech Stack:** Docker (`prettyout-test` image), bash, Go (plugins already built in image).

---

## Specification: what `docs/showcase.md` must contain

### Document structure

```
# prettyout — Showcase

Introduction (3–4 sentences): what prettyout does and why.

---

## Table of Contents
  one line per tool with anchor links

---

## ruff
## mypy
## basedpyright
## bandit
## pylint
## eslint
## biome
## stylelint
## shellcheck
## hadolint
## golangci-lint
## cargo clippy
## trivy
## semgrep

---

## Summary: What prettyout solves
```

---

### Per-tool section spec

Each tool section must follow **exactly** this structure (order matters):

```markdown
## <Tool Name>

**What it checks:** One sentence. (e.g. "Python linter for style, errors, and imports.")

### The Code

A code block with the fixture file(s) used. If multiple files, show each with its filename as a comment/header. Keep it short — enough to produce 2–4 interesting violations.

### Without prettyout

Two sub-sections:

#### Default output

The output a developer normally sees when running the tool **without** `--output-format=json`. This is `<tool> <default-flags> <files>` with color codes stripped (use `--no-color` or `| cat` or similar). If the tool has no human-readable default (only JSON), skip this sub-section and note it.

#### JSON output (what CI/CD sees)

The raw JSON output produced by `<tool> <json-flag> <files>`. Truncated to ≤30 lines if very long. Show with `...` if truncated. This is the input prettyout reads.

### With prettyout

Two sub-sections showing the two main views:

#### Group by rule (default)

Output of: `<tool> <json-flag> <files> 2>/dev/null | prettyout-<tool>`

#### Group by file

Output of the same with `.prettyout.yaml`:
```yaml
settings:
  <tool>:
    group_by: file
```

### What prettyout improves

Bullet list, 3–6 items. Each item names a specific problem and how prettyout addresses it. Use this template:

- **<problem name>**: <one sentence describing the raw problem> → <one sentence describing what prettyout does instead>

**Standard bullets to include (adapt wording per tool):**
- **Consistent format**: Each tool has its own output style → prettyout produces the same structure regardless of tool
- **Rule grouping**: Raw output lists violations line-by-line → group-by-rule view shows all affected files per rule at a glance
- **Occurrence counts**: No summary in raw output → `F401 (4) — message` shows how widespread each rule is
- **Summary line**: Must scroll to the end to see counts → `N issues · M rules · K files` always at the end
- **Singular/plural**: (only if relevant) Raw output may show `1 issues` → prettyout always uses correct grammar

**Tool-specific bullets to add where relevant:**
- ruff: nothing extra
- mypy: `note`-level messages stripped (clutter)
- basedpyright: duplicate line numbers from parse errors deduplicated; severity prefix [ERROR]/[WARN]
- bandit: severity + confidence shown together (`B324 (MEDIUM/HIGH)`)
- pylint: `E0602/undefined-variable` format links ID + readable name
- eslint: parse errors (null ruleId) shown as `parse-error` rule
- biome: no line numbers (byte offsets only) — prettyout shows filename-only entries cleanly
- stylelint: JSON output on stderr — prettyout handles this transparently
- shellcheck: integer code formatted as `SC2006` prefix
- hadolint: mix of DL and SC codes shown uniformly
- golangci-lint: `Issues: null` on empty run handled gracefully
- cargo clippy: NDJSON stream filtered to `compiler-message` only
- trivy: groups by severity (CRITICAL→HIGH→MEDIUM→LOW→UNKNOWN), shows "no fix available"
- semgrep: check_id as rule code, extra.severity normalized

### How to verify

A code block with the exact commands to reproduce both outputs. This section must be runnable by anyone with Docker:

```bash
# Build image (once)
docker build -t prettyout-test -f test/Dockerfile .

# Run (from project root)
docker run --rm -v $(pwd):/project prettyout-test bash -c "
cd /tmp && mkdir t && cd t

cat > errors.py << 'PY'
<fixture code>
PY

# Without prettyout (default output)
<tool> <files>

# Without prettyout (JSON)
<tool> <json-flag> <files>

# With prettyout (group by rule)
<tool> <json-flag> <files> 2>/dev/null | prettyout-<tool>

# With prettyout (group by file)
printf 'settings:\n  <tool>:\n    group_by: file\n' > .prettyout.yaml
<tool> <json-flag> <files> 2>/dev/null | prettyout-<tool>
rm .prettyout.yaml
"
```
```

---

### Summary section spec

At the end of the document, a section `## Summary: What prettyout solves` with:

1. **A comparison table** across all tools:

| Problem | Tools affected | prettyout solution |
|---------|---------------|-------------------|
| Inconsistent output format between tools | All | Uniform grouped format |
| No violation count summary | ruff, mypy, eslint, ... | `N issues · M rules · K files` |
| Line-by-line output (same rule repeated) | All | Group by rule or file |
| JSON only readable by machines | All | Human-readable with colors |
| Severity not shown | basedpyright, mypy, ... | `[ERROR]`/`[WARN]`/`[INFO]` prefix |
| Duplicate violations | basedpyright, pylint | Set-based dedup |
| Wrong singular/plural | many | `formatter.Plural()` everywhere |

2. **Two short paragraphs:**
   - When to use group-by-rule (refactoring: "fix all F401 across the project")
   - When to use group-by-file (code review: "what's wrong with this file")

---

## How to generate the document

The agent must run all outputs **inside Docker** using the `prettyout-test` image (all tools + plugins pre-installed).

### Important: capturing outputs correctly

| Tool | JSON flag | Output goes to | prettyout pipe |
|------|-----------|---------------|----------------|
| ruff | `--output-format=json` | stdout | `2>/dev/null \| prettyout-ruff` |
| mypy | `--output=json` | stdout | `2>/dev/null \| prettyout-mypy` |
| basedpyright | `--outputjson` | stdout | `2>/dev/null \| prettyout-basedpyright` |
| bandit | `-f json` | stdout | `2>/dev/null \| prettyout-bandit` |
| pylint | `--output-format=json` | stdout | `2>/dev/null \| prettyout-pylint` |
| eslint | `--format=json` | stdout | `2>/dev/null \| prettyout-eslint` |
| biome | `check --reporter=json` | stdout | `2>/dev/null \| prettyout-biome` |
| stylelint | `--formatter=json` | **stderr** | `2>&1 >/dev/null \| prettyout-stylelint` |
| shellcheck | `--format=json` | stdout | `2>/dev/null \| prettyout-shellcheck` |
| hadolint | `--format=json` | stdout | `2>/dev/null \| prettyout-hadolint` |
| golangci-lint | `--out-format=json --disable-all --enable=ineffassign` | stdout | `2>/dev/null \| prettyout-golangci` |
| cargo clippy | `--message-format=json` | stdout | `2>/dev/null \| prettyout-cargo-clippy` |
| trivy | `fs --format=json --quiet` | stdout | `2>/dev/null \| prettyout-trivy` |
| semgrep | `--config rules.yaml --json` | stdout | `2>/dev/null \| prettyout-semgrep` |

### Strip ANSI codes for markdown

All output captured for the document must have ANSI codes stripped. Either:
- Use `colors: false` in `.prettyout.yaml` when capturing prettyout output
- OR pipe through `sed 's/\x1b\[[0-9;]*m//g'`

Default tool output (non-JSON) may also have colors — strip those too.

### Fixture requirements per tool

Fixtures must be **minimal but interesting** — enough to show 2–4 violations of 2–3 different rules. The same fixture files from `test/fixtures/<tool>/errors.*` can be reused if they already exist. Otherwise create inline.

---

## Task 0: Setup

- [ ] **Check if `test/fixtures/` exists** from the restructure plan. If yes, reuse those fixture files. If not, create inline fixtures per tool.

- [ ] **Verify Docker image is current:**
```bash
docker build -t prettyout-test -f test/Dockerfile . 2>&1 | tail -3
```

- [ ] **Verify all plugins are built in the image:**
```bash
docker run --rm prettyout-test bash -c "ls /usr/local/bin/prettyout-*"
```

---

## Task 1: Generate per-tool sections

For each tool, run a Docker command to capture outputs, then write the markdown section.

**One subtask per tool.** Run them sequentially. Write the section directly into a temp file, then at the end assemble into `docs/showcase.md`.

Template command per tool (adapt per the table above):

```bash
docker run --rm -v $(pwd):/project prettyout-test bash -c "
mkdir -p /tmp/t-<tool> && cd /tmp/t-<tool>

# Write fixture files
cat > errors.<ext> << 'FIXTURE'
<fixture code>
FIXTURE

# Capture default output (strip colors)
echo '=== DEFAULT ===' && <tool> <files> 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g' | head -20

# Capture JSON output  
echo '=== JSON ===' && <tool> <json-flag> <files> 2>/dev/null | head -30

# Capture prettyout group-by-rule
echo '=== PRETTYOUT RULE ===' && printf 'settings:\n  <tool>:\n    colors: false\n' > .prettyout.yaml && <tool> <json-flag> <files> 2>/dev/null | prettyout-<tool> && rm .prettyout.yaml

# Capture prettyout group-by-file
echo '=== PRETTYOUT FILE ===' && printf 'settings:\n  <tool>:\n    group_by: file\n    colors: false\n' > .prettyout.yaml && <tool> <json-flag> <files> 2>/dev/null | prettyout-<tool> && rm .prettyout.yaml
"
```

Tools to cover (in order):
1. ruff
2. mypy
3. basedpyright
4. bandit
5. pylint
6. eslint (needs eslint.config.mjs in CWD)
7. biome (needs biome.json in CWD)
8. stylelint (needs .stylelintrc.json; pipe is `2>&1 >/dev/null`)
9. shellcheck
10. hadolint
11. golangci-lint (needs go.mod; use `--disable-all --enable=ineffassign`)
12. cargo clippy (needs Cargo.toml + src/)
13. trivy (use `trivy fs --format=json --quiet .` on dir with requirements.txt)
14. semgrep (needs local rules.yaml)

---

## Task 2: Assemble `docs/showcase.md`

- [ ] Create `docs/showcase.md` with intro + table of contents
- [ ] Insert each per-tool section in order
- [ ] Write the final `## Summary: What prettyout solves` section with comparison table
- [ ] Review: every tool section has all 5 sub-sections (code, default output, JSON, prettyout-rule, prettyout-file, improvements, how to verify)
- [ ] Check: all code blocks have correct language tags (` ```python `, ` ```json `, ` ```bash `, ` ``` ` for tool output)
- [ ] Check: ANSI codes absent from all output blocks

- [ ] **Commit:**
```bash
git add docs/showcase.md
git commit -m "docs: add prettyout showcase with before/after for all 14 tools"
```

---

## Quality bar for the document

The agent writing this document should ask: "Would a developer unfamiliar with prettyout understand the value immediately from this section?"

Each section should pass this test:
- [ ] The fixture code is short enough to read in 10 seconds
- [ ] The raw JSON block is ugly enough to make the contrast obvious
- [ ] The prettyout output is readable at a glance
- [ ] The "What prettyout improves" bullets are concrete (name the problem, name the solution — not "it's better")
- [ ] The "How to verify" block is copy-paste runnable with Docker

---

## Output file

**`docs/showcase.md`** — committed to the repo, kept up to date as new tools are added.

Approximate length: 600–900 lines (14 tools × ~50 lines each + intro + summary).
