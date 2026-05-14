# prettyout — Architecture

## How It Works

prettyout wraps CLI tools transparently: you run `ruff check .` as usual, the shell function intercepts it, runs the tool with JSON output, pipes the JSON through the formatter plugin, and prints pretty grouped output. If something goes wrong or you're in CI, the original tool runs unchanged.

```
$ ruff check .
     ↓  (shell function intercepts)
command ruff check . --output-format=json  2>"$tmpfile"
     ↓  (JSON piped to plugin)
prettyout-ruff
     ↓  (pretty output to stdout, stderr passed through separately)
E501 - Line too long
Affected files:
  - foo.py — line 42
```

---

## Components

### Registry (`internal/registry/`)

The registry is a YAML file embedded in the binary (`builtin.yaml`). It describes every supported tool and launcher:

```yaml
tools:
  ruff:
    plugin: prettyout-ruff
    intercept_subcommands: [check]   # only these; others pass through
    output_args: [--output-format=json]
    passthrough_flags: [--watch, -W] # streaming — skip interception entirely
    launchers: [uvx, pipx]           # also wrap these launchers for this tool

launchers:
  uvx:
    skip_flags: [--no-cache, --no-project, ...]
    value_flags: [--python, --with, --from]  # flags that consume the next arg
    tool_position: first_non_flag
  pipx:
    prefix_args: [run]               # "pipx run <tool>" — skip "run" before tool name
    ...
```

**`intercept_subcommands`** — if set, only listed subcommands are intercepted. `ruff format` hits no match in the case statement and falls through to `command ruff "$@"`. If the list is empty/absent, all invocations are intercepted.

**`passthrough_flags`** — if any of these flags appear in the args, the original tool runs unchanged. Used for `--watch` which streams output and can't be buffered.

**`launchers`** — which launcher wrappers should also intercept this tool (e.g. `uvx ruff check`).

Users can't edit `builtin.yaml` directly, but can add tools via `custom_tools` in their config.

---

### Hook Generator (`internal/hook/`)

`hook.Generate(shell, registry, config)` produces shell functions for every enabled tool and every relevant launcher.

**Direct tool wrapper** (zsh example for ruff):
```bash
ruff() { prettyout _run ruff "$@"; }
uvx() { prettyout _run uvx "$@"; }
```

All interception logic — enabled checks, CI mode, passthrough flags, subcommand matching, process orchestration — is handled by the `prettyout _run` runner.

---

### Runner (`internal/runner/`)

`runner.RunFromArgs(args)` is called by `prettyout _run <tool> [args...]`. It:

1. Loads registry + config
2. Calls `Decide(toolName, args, reg, cfg, isTTY())` — pure logic that returns a `Decision`
3. Calls `Execute(decision)` — runs the real tool, optionally piping through the formatter

`Decide` handles: CI mode check, enabled check, passthrough-flag detection, subcommand matching, launcher arg parsing (via `internal/launcher`), plugin resolution.

`Execute` handles: passthrough (exec with inherited stdio) or intercept (two-process pipe with stderr separation and exit code from tool).

---

### Config (`internal/config/`)

Two config files, merged at load time. Per-project overrides global.

- Global: `~/.config/prettyout/config.yaml`
- Per-project: `.prettyout.yaml` in the current directory

```yaml
enabled:
  ruff: true
  basedpyright: false

# Replace the formatter entirely (any executable: binary, script, etc.)
plugins:
  ruff: ~/scripts/my-ruff-formatter.py

# Configure the built-in formatter
settings:
  ruff:
    colors: true
    max_message_length: 100   # truncate long messages

# Register a tool not in the built-in registry
custom_tools:
  mycooltool:
    plugin: ~/scripts/mycooltool-fmt
    intercept_subcommands: [lint]
    output_args: [--json]

# CI behavior: "auto" | "always" | "never"
# auto  = skip interception when stdout is not a TTY
# always = pretty output even in CI/pipes
# never = disable prettyout globally
ci_mode: auto
```

Merge semantics: each map is merged key-by-key (project wins per key). `ci_mode` from project only overrides if explicitly set to something other than `"auto"`.

---

### Formatter Plugin API (`pkg/formatter/`)

A plugin is any executable that reads JSON from stdin and writes formatted text to stdout.

```go
// Simple: just transform bytes
formatter.Run(func(data []byte) error { ... })

// With config: reads settings from config files
formatter.RunWithConfig("ruff", func(data []byte, cfg formatter.Config) error { ... })
```

`formatter.Config` carries the resolved settings for the tool:

```go
type Config struct {
    GroupBy          string         // "rule" | "file", default "rule"
    Colors           bool           // default true
    MaxMessageLength int            // 0 = unlimited
    Extra            map[string]any // tool-specific settings
}
```

`RunWithConfig` loads `~/.config/prettyout/config.yaml` and `.prettyout.yaml`, merges them, extracts the `settings.<toolname>` block, and applies defaults for anything not set.

Plugin resolution order:
1. User override in `plugins.<toolname>` in config
2. `prettyout-<toolname>` found in `$PATH`
3. Error: no formatter found

---

### Plugins (`cmd/prettyout-ruff`, `cmd/prettyout-basedpyright`)

Standalone binaries. Each reads the tool's JSON format, groups diagnostics by rule code, and prints them with file/line info. They use `filepath.Base()` on filenames — no hardcoded path stripping.

---

## Data Flow Summary

```
shell startup
  └─ eval "$(prettyout hook zsh)"
       └─ reads registry + config
       └─ generates ruff(), uvx(), basedpyright() functions

$ ruff check .
  └─ ruff() shell function
       └─ prettyout _run ruff check .
            └─ Decide: enabled? TTY? passthrough flag? subcommand matches? → intercept
            └─ Execute: ruff check . --output-format=json 2>tmpfile | prettyout-ruff
            └─ cat tmpfile >&2
            └─ return ruff exit code

$ ruff format .
  └─ ruff() shell function
       └─ prettyout _run ruff format .
            └─ Decide: subcommand "format" not in intercept_subcommands → passthrough
            └─ Execute: exec ruff format . (inherited stdio)
```

---

## Adding a New Tool

1. Add an entry to `internal/registry/builtin.yaml` (or to `custom_tools` in config for local use).
2. Write a plugin binary `prettyout-<toolname>` using `formatter.RunWithConfig`.
3. Put the binary in `$PATH`.
4. Run `prettyout enable <toolname>`.

The hook is regenerated automatically on the next shell startup (since `.zshrc` evals `prettyout hook zsh`).
