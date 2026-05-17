# Showcase Documentation Design

**Goal:** Produce per-tool format comparison docs and a summary page showing what prettyout improves over raw tool output. Target audience: developers evaluating prettyout, both on GitHub and a future website.

---

## File Structure

```
docs/
  tools/
    ruff.md
    mypy.md
    basedpyright.md
    bandit.md
    pylint.md
    eslint.md
    biome.md
    stylelint.md
    shellcheck.md
    hadolint.md
    golangci-lint.md
    cargo-clippy.md
    trivy.md
    semgrep.md
  summary.md
```

`summary.md` is the entry point. Each `tools/<tool>.md` is a self-contained comparison page.

---

## Per-Tool File Structure (`docs/tools/<tool>.md`)

```markdown
# <Tool Name>

**What it checks:** One sentence.

## Without prettyout

### Default output
<tool output with colors stripped>

### JSON (what CI/CD sees)
<raw JSON, truncated to ~15 lines with "..." if longer>

## With prettyout

### Group by rule (default)
<prettyout output, colors: false>

### Group by file
<prettyout output with group_by: file, colors: false>

## What prettyout improves

- **<problem name>**: <what's wrong in raw output> → <what prettyout does instead>
- ... (3–5 bullets, concrete and specific)
```

All outputs are generated from real commands inside Docker using the `prettyout-test` image. Fixtures are reused from `test/fixtures/<tool>/`. ANSI codes are stripped for markdown.

---

## `docs/summary.md` Structure

```markdown
# prettyout — Output Format Showcase

3–4 sentences: what prettyout does, how it works (install once, use tools normally, output format changes).

## Tools

| Tool | What it checks | Docs |
|------|---------------|------|
| ruff | Python linter for style, errors, and imports | [ruff.md](tools/ruff.md) |
| ... | ... | ... |

## What prettyout improves

| Problem | Tools affected | prettyout solution |
|---------|---------------|-------------------|
| Inconsistent output format between tools | All | Uniform grouped format |
| Line-by-line output (same rule repeated) | All | Group by rule or file |
| No violation count summary | ruff, mypy, eslint, ... | `N issues · M rules · K files` |
| JSON only readable by machines | All | Human-readable with colors |
| Severity not shown | basedpyright, mypy, ... | `[ERROR]`/`[WARN]`/`[INFO]` prefix |
| Duplicate violations | basedpyright, pylint | Set-based deduplication |
| Wrong singular/plural | many | Correct grammar always |

## Group by rule vs group by file

Two short paragraphs: when to use each (group-by-rule for refactoring across files, group-by-file for code review).
```

---

## Generation Approach

All outputs are captured by running Docker commands against the `prettyout-test` image (all tools + plugins pre-installed). Fixture files from `test/fixtures/<tool>/` are reused where they exist.

| Tool | JSON flag | Pipe |
|------|-----------|------|
| ruff | `--output-format=json` | `2>/dev/null \| prettyout-ruff` |
| mypy | `--output=json` | `2>/dev/null \| prettyout-mypy` |
| basedpyright | `--outputjson` | `2>/dev/null \| prettyout-basedpyright` |
| bandit | `-f json` | `2>/dev/null \| prettyout-bandit` |
| pylint | `--output-format=json` | `2>/dev/null \| prettyout-pylint` |
| eslint | `--format=json` | `2>/dev/null \| prettyout-eslint` |
| biome | `check --reporter=json` | `2>/dev/null \| prettyout-biome` |
| stylelint | `--formatter=json` | `2>&1 >/dev/null \| prettyout-stylelint` |
| shellcheck | `--format=json` | `2>/dev/null \| prettyout-shellcheck` |
| hadolint | `--format=json` | `2>/dev/null \| prettyout-hadolint` |
| golangci-lint | `--out-format=json --disable-all --enable=ineffassign` | `2>/dev/null \| prettyout-golangci` |
| cargo clippy | `--message-format=json` | `2>/dev/null \| prettyout-cargo-clippy` |
| trivy | `fs --format=json --quiet` | `2>/dev/null \| prettyout-trivy` |
| semgrep | `--config rules.yaml --json` | `2>/dev/null \| prettyout-semgrep` |

ANSI codes stripped via `colors: false` in `.prettyout.yaml` when capturing prettyout output; default tool output piped through `sed 's/\x1b\[[0-9;]*m//g'`.
