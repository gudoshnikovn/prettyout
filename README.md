# prettyout

Pretty, grouped output for CLI linters and security scanners. Works transparently — install once, use your tools as usual.

```
$ ruff check .

E501 (12) — Line too long
  src/main.py — lines 42, 87, 103
  src/utils.py — lines 15, 29

F401 (3) — Module imported but unused
  src/main.py — line 1
  tests/test_api.py — lines 3, 11

3 issues · 2 rules · 3 files
```

Supports 14 tools across Python, JavaScript/TypeScript, Go, Rust, Shell, Docker, and security scanning.

## Supported tools

| Tool | Language / domain |
|------|-------------------|
| ruff | Python linter |
| mypy | Python type checker |
| basedpyright | Python type checker (strict) |
| bandit | Python security linter |
| pylint | Python linter |
| eslint | JavaScript / TypeScript linter |
| biome | JS / TS / JSON formatter and linter |
| stylelint | CSS / SCSS / Less linter |
| shellcheck | Shell script linter |
| hadolint | Dockerfile linter |
| golangci-lint | Go meta-linter |
| cargo clippy | Rust linter |
| trivy | Security vulnerability scanner |
| semgrep | Semantic code analysis |

## Install

Install the main binary and any plugins you need:

```bash
go install github.com/gudoshnikovn/prettyout/cmd/prettyout@latest

# Python
go install github.com/gudoshnikovn/prettyout/cmd/prettyout-ruff@latest
go install github.com/gudoshnikovn/prettyout/cmd/prettyout-mypy@latest
go install github.com/gudoshnikovn/prettyout/cmd/prettyout-basedpyright@latest
go install github.com/gudoshnikovn/prettyout/cmd/prettyout-bandit@latest
go install github.com/gudoshnikovn/prettyout/cmd/prettyout-pylint@latest

# JavaScript / TypeScript
go install github.com/gudoshnikovn/prettyout/cmd/prettyout-eslint@latest
go install github.com/gudoshnikovn/prettyout/cmd/prettyout-biome@latest
go install github.com/gudoshnikovn/prettyout/cmd/prettyout-stylelint@latest
go install github.com/gudoshnikovn/prettyout/cmd/prettyout-npm-audit@latest

# Shell / Docker
go install github.com/gudoshnikovn/prettyout/cmd/prettyout-shellcheck@latest
go install github.com/gudoshnikovn/prettyout/cmd/prettyout-hadolint@latest

# Go / Rust
go install github.com/gudoshnikovn/prettyout/cmd/prettyout-golangci@latest
go install github.com/gudoshnikovn/prettyout/cmd/prettyout-cargo-clippy@latest

# Security
go install github.com/gudoshnikovn/prettyout/cmd/prettyout-trivy@latest
go install github.com/gudoshnikovn/prettyout/cmd/prettyout-semgrep@latest
```

## Setup

```bash
prettyout setup
source ~/.zshrc   # or restart terminal
```

This adds `eval "$(prettyout hook zsh)"` to your rc file. After that, supported tools automatically produce pretty output.

For bash:

```bash
prettyout setup --shell bash
source ~/.bashrc
```

## Usage

Just use your tools as usual:

```bash
ruff check .
eslint src/
golangci-lint run ./...
trivy fs .
```

Only intercepted subcommands are wrapped — `ruff format` runs unchanged. Flags like `--watch` are passed through without buffering.

## Manage tools

```bash
prettyout list              # show all tools and their status
prettyout enable ruff       # enable pretty output for ruff
prettyout disable ruff      # fall back to original output
prettyout doctor            # diagnose setup problems
```

## Configuration

`~/.config/prettyout/config.yaml`:

```yaml
enabled:
  ruff: true
  eslint: true

settings:
  ruff:
    colors: true
    group_by: rule        # "rule" (default) or "file"
    max_message_length: 100
    only_rules: [E501, F401]   # optional: filter to specific rules
    only_files: [src/]         # optional: filter to specific paths

ci_mode: auto   # auto | always | never
                # auto = skip interception when stdout is not a TTY
```

Per-project override: `.prettyout.yaml` in the project root (same schema, all fields optional).

## Group by rule vs group by file

**Group by rule** (default) — best for fixing one type of issue across the whole codebase:

```
E501 (12) — Line too long
  src/main.py — lines 42, 87
  src/utils.py — line 15
```

**Group by file** — best for reviewing everything wrong in a specific file:

```
src/main.py
  E501 — lines 42, 87
  F401 — line 1
```

Switch per-project:

```yaml
# .prettyout.yaml
settings:
  ruff:
    group_by: file
```

## Writing a plugin

A plugin is any executable named `prettyout-<toolname>` that reads JSON from stdin and writes formatted text to stdout.

```go
import "github.com/gudoshnikovn/prettyout/pkg/formatter"

func main() {
    formatter.RunWithConfig("mytool", func(data []byte, cfg formatter.Config) error {
        // parse data, print pretty output
        // cfg.Colors, cfg.GroupBy, cfg.OnlyRules, cfg.OnlyFiles are set from config
        return nil
    })
}
```

Register the tool in `custom_tools` in your config:

```yaml
custom_tools:
  mytool:
    plugin: prettyout-mytool
    output_args: [--json]
    intercept_subcommands: [lint]  # optional: only intercept these subcommands
```

Or add it to `internal/registry/builtin.yaml` and open a PR.

See [docs/architecture.md](docs/architecture.md) for a full description of how everything works.

## Use plugins directly

```bash
ruff check . --output-format=json | prettyout-ruff
eslint --format json src/ | prettyout-eslint
golangci-lint run --out-format json ./... | prettyout-golangci
trivy fs --format json . | prettyout-trivy
```
