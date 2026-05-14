# prettyout

Pretty, grouped output for CLI tools that support JSON output (ruff, basedpyright, and more).

## Install

```bash
go install github.com/gudoshnikov_na/prettyout/cmd/prettyout@latest
go install github.com/gudoshnikov_na/prettyout/cmd/prettyout-ruff@latest
go install github.com/gudoshnikov_na/prettyout/cmd/prettyout-basedpyright@latest
```

## Setup (one time)

```bash
prettyout setup
source ~/.zshrc   # or restart terminal
```

This adds a shell hook so `ruff` and `basedpyright` automatically use pretty output.

## Usage

After setup, just use tools as usual:

```bash
ruff check .
basedpyright .
```

## Manage hooks

```bash
prettyout list              # show all tools and their status
prettyout enable ruff       # enable pretty output for ruff
prettyout disable ruff      # disable (fall back to original output)
```

## Use plugins directly

```bash
ruff check . --output-format=json | prettyout-ruff
basedpyright --outputjson . | prettyout-basedpyright
```

## Writing a plugin

A prettyout plugin is any binary named `prettyout-<toolname>` that:
- Reads JSON from stdin
- Writes formatted text to stdout
- Exits with code 0 on success

Use the shared helper:

```go
import "github.com/gudoshnikov_na/prettyout/pkg/formatter"

func main() {
    formatter.Run(func(data []byte) error {
        // parse data, print pretty output
        return nil
    })
}
```

Add your plugin to `registry.yaml` via a PR.
