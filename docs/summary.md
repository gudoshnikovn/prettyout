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

```yaml
settings:
  ruff:
    group_by: file
```
