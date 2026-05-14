# prettyout

Pretty, grouped output for CLI tools that support JSON output. Works transparently — install once, use tools as usual.

```
$ ruff check .

E501 - Line too long
Affected files:
  - foo.py — line 42
  - bar.py — lines 10-11
----------------------------------------
reportMissingImports - Import "foo" could not be resolved
Affected files:
  - main.py — line 1
----------------------------------------
```

Supports ruff, basedpyright, uvx/pipx launchers. Extensible: write any executable as a plugin.

## Install

```bash
go install github.com/gudoshnikov_na/prettyout/cmd/prettyout@latest
go install github.com/gudoshnikov_na/prettyout/cmd/prettyout-ruff@latest
go install github.com/gudoshnikov_na/prettyout/cmd/prettyout-basedpyright@latest
```

## Setup

```bash
prettyout setup
source ~/.zshrc   # or restart terminal
```

This adds `eval "$(prettyout hook zsh)"` to your rc file. After that, `ruff`, `basedpyright`, `uvx ruff`, and `pipx ruff` automatically use pretty output.

## Usage

Just use your tools as usual:

```bash
ruff check .
basedpyright .
uvx ruff check .
```

Only intercepted subcommands are wrapped — `ruff format` runs unchanged. Flags like `--watch` are passed through without buffering.

## Manage tools

```bash
prettyout list              # show all tools and their status
prettyout enable ruff       # enable pretty output for ruff
prettyout disable ruff      # fall back to original output
```

## Configuration

`~/.config/prettyout/config.yaml`:

```yaml
enabled:
  ruff: true
  basedpyright: true

settings:
  ruff:
    colors: true
    max_message_length: 100

ci_mode: auto   # auto | always | never
                # auto = skip interception when stdout is not a TTY
```

Per-project override: `.prettyout.yaml` in the project root (same schema, all fields optional).

## Writing a plugin

A plugin is any executable named `prettyout-<toolname>` that reads JSON from stdin and writes formatted text to stdout.

```go
import "github.com/gudoshnikov_na/prettyout/pkg/formatter"

func main() {
    formatter.RunWithConfig("mytool", func(data []byte, cfg formatter.Config) error {
        // parse data, print pretty output
        // cfg.Colors, cfg.MaxMessageLength, cfg.GroupBy are set from user config
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
basedpyright --outputjson . | prettyout-basedpyright
```
