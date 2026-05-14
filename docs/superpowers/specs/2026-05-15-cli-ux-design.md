# CLI UX Improvements — Design Spec

## Goal

Make prettyout convenient and polished as a toolchain before expanding plugin coverage. This covers new commands (`status`, `doctor`, `install`, `update`, `upgrade`, `completion`), enhancements to existing commands (`list --available`, `enable` with hook check), debug mode for `_run`, and a fix for output-arg conflict detection in `Decide`.

---

## User-Facing Behaviour

### `prettyout status`

Informational only. Always exits 0.

```
prettyout v0.3.0

Shell:   zsh
Hook:    ✓ ~/.zshrc
Config:  ~/.config/prettyout/config.yaml

Tools:
  ruff            enabled    prettyout-ruff          ✓ installed
  basedpyright    disabled   prettyout-basedpyright  ✗ not found
```

Shows: binary version, detected shell, hook presence in rc file, config path, all registered tools with enabled status and plugin install status.

---

### `prettyout doctor`

Runs a set of named checks. Exits 0 if all pass, exits 1 if any fail. Each failing check includes a `→` hint.

```
✓ Shell hook present in ~/.zshrc
✓ prettyout-ruff found in PATH
✗ prettyout-basedpyright not found in PATH
  → prettyout install basedpyright
✗ Config parse error: ~/.config/prettyout/config.yaml line 4
  → unexpected key "colour"

2 issues found.
```

Checks performed (in order):
1. Hook line present in rc file (`prettyout hook` in `~/.zshrc` or `~/.bashrc`)
2. For each enabled tool: plugin binary found in `$PATH`
3. Global config file parses without error (if it exists)
4. Per-project config parses without error (if `.prettyout.yaml` exists in cwd)

---

### `prettyout list` / `prettyout list --available`

**Without flag** (current behaviour, enhanced):
```
Tool             Status     Plugin
---------------  ---------  ----------------------
basedpyright     disabled   prettyout-basedpyright
ruff             enabled    prettyout-ruff
```

**With `--available`** — shows all tools from registry, including those with no plugin installed:
```
Tool             Status     Installed
---------------  ---------  ---------
basedpyright     disabled   ✓
eslint           —          ✗  prettyout install eslint
golangci         —          ✗  prettyout install golangci
ruff             enabled    ✓
```

"Installed" = plugin binary found in `$PATH`. `—` in Status = not enabled and no config entry.

---

### `prettyout install <tool>`

Installs the plugin binary for a tool.

```
$ prettyout install golangci
Installing prettyout-golangci...
→ go install github.com/gudoshnikov_na/prettyout/cmd/prettyout-golangci@latest
✓ Installed. Run: prettyout enable golangci
```

Errors:
- Tool not in registry → `prettyout: unknown tool "golangci"` (exit 1)
- No `install.go` entry in registry → `prettyout: no install method for "golangci"` (exit 1)
- `go install` fails → forward stderr, exit 1

Does **not** automatically enable the tool — the user runs `prettyout enable <tool>` separately.

---

### `prettyout update [tool]`

Without argument: updates all plugins that are currently installed (binary found in PATH).
With argument: updates a specific plugin.

```
$ prettyout update
Updating installed plugins...
  prettyout-ruff ... ✓
  prettyout-basedpyright ... ✓
Done.

$ prettyout update ruff
Updating prettyout-ruff...
→ go install github.com/gudoshnikov_na/prettyout/cmd/prettyout-ruff@latest
✓ Done.
```

Skips tools with no `install.go` entry silently (with a note if `--verbose` ever added).

---

### `prettyout upgrade`

Updates the `prettyout` binary itself.

```
$ prettyout upgrade
Detected: go install
→ go install github.com/gudoshnikov_na/prettyout/cmd/prettyout@latest
✓ prettyout updated.
```

Detection order:
1. `brew list prettyout` exits 0 → use `brew upgrade prettyout`
2. Otherwise → use `go install github.com/gudoshnikov_na/prettyout/cmd/prettyout@latest`

Extensible: detection is a list of `Detector` structs in `internal/install`, easy to add `apt`, `scoop`, etc.

---

### `prettyout enable <tool>` — hook check

After enabling, check if hook is present in rc file. If not, warn:

```
$ prettyout enable ruff
Enabled prettyout for ruff.
⚠ Hook not found in ~/.zshrc — run: prettyout setup
```

If hook is already present: no extra output (current behaviour unchanged).

---

### `prettyout completion <shell>`

Outputs a completion script. User adds to their rc file once.

```bash
# ~/.zshrc
eval "$(prettyout completion zsh)"
```

Supported shells: `zsh`, `bash`.

Completions provided:
- Subcommand names (`setup`, `hook`, `enable`, `disable`, `list`, `install`, `update`, `doctor`, `status`, `completion`, `upgrade`)
- Tool names for `enable`, `disable`, `install`, `update` (read from registry at completion time)

---

### `prettyout _run --debug <tool> [args...]`

`--debug` as first arg after `_run` (before tool name). Prints decision trace to stderr.

```
$ prettyout _run --debug ruff check .
[prettyout] tool=ruff args=[check .]
[prettyout] ci_mode=auto tty=true → proceeding
[prettyout] enabled=true
[prettyout] output_args conflict: none
[prettyout] passthrough_flags: none matched
[prettyout] subcommand=check → match → intercepting
[prettyout] plugin=prettyout-ruff
```

Debug output goes to stderr only, does not affect stdout or the tool's exit code.

---

### Output-arg conflict detection in `Decide`

If any user-provided arg shares the same flag name prefix as an `output_args` entry, fall through to passthrough instead of appending a conflicting flag.

Example: `output_args: [--output-format=json]`, user passes `--output-format=github` → conflict detected → passthrough (ruff runs normally with `--output-format=github`, no pretty output).

Detection rule: for each `output_arg`, extract the flag key up to `=`. If any user arg starts with that key, it's a conflict.

```
--output-format=json  → key: --output-format
--outputjson          → no `=`, exact match only
```

---

## Architecture

### New internal packages

**`internal/doctor/`**

```go
type Check struct {
    Name    string
    OK      bool
    Message string
    Hint    string // shown only when !OK
}

func Run(reg *registry.Registry, cfg *config.Config) []Check
```

Each check is a private function returning a `Check`. `Run` executes all and returns results. No I/O — callers format and print.

**`internal/install/`**

```go
// InstallPlugin installs or updates a plugin binary for the named tool.
func InstallPlugin(tc registry.ToolConfig) error

// UpgradeSelf detects the package manager and upgrades prettyout.
func UpgradeSelf() error

// IsInstalled reports whether the plugin binary for tc is in PATH.
func IsInstalled(tc registry.ToolConfig) bool
```

`UpgradeSelf` uses a `[]Detector` slice (each has `Detect() bool` and `Upgrade() error`). Currently two detectors: `brewDetector` and `goInstallDetector`. Adding new ones = appending to the slice.

**`internal/completion/`**

```go
func Generate(shell string, reg *registry.Registry) (string, error)
```

Generates a completion script string for the given shell. The script calls `prettyout _completions <subcommand>` at completion time to get dynamic tool name lists — avoids baking tool names into the static script.

Add new internal subcommand `prettyout _completions tools` that prints tool names (one per line) for use by the completion script.

### Registry changes

Add `Install` field to `ToolConfig`:

```go
type InstallConfig struct {
    Go   string `yaml:"go"`
    Brew string `yaml:"brew"` // reserved for future use
}

type ToolConfig struct {
    // existing fields unchanged
    Install InstallConfig `yaml:"install"`
}
```

`builtin.yaml` additions:
```yaml
ruff:
  install:
    go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-ruff
basedpyright:
  install:
    go: github.com/gudoshnikov_na/prettyout/cmd/prettyout-basedpyright
```

### New command files in `cmd/prettyout/`

| File | Functions |
|---|---|
| `cmd_status.go` | `runStatus()` |
| `cmd_doctor.go` | `runDoctor()` |
| `cmd_install.go` | `runInstall(args)`, `runUpdate(args)`, `runUpgrade()` |
| `cmd_list.go` | `runList(args)` — replaces inline impl in main.go |
| `cmd_completion.go` | `runCompletion(args)` |
| `cmd_enable.go` | `runEnable(args)` — moves from main.go, adds hook check |

`main.go` switch gains: `status`, `install`, `update`, `upgrade`, `completion`, `_completions`. Existing `enable`, `list` cases delegate to new files.

### Runner changes

`RunFromArgs` parses `--debug` as an optional first arg:

```go
// prettyout _run --debug ruff check .
// args[0] == "--debug" → enable debug, shift args by 1
```

`Decide` gains `debug bool` parameter (or a `DecideOptions` struct to avoid future signature churn). When debug=true, each decision point writes a `[prettyout]` line to stderr.

`Decide` gains output-arg conflict check before appending `output_args`.

---

## What This Is Not

- No `prettyout config set/get` — edit the yaml directly
- No `prettyout set <tool> <key> <value>` — plugin settings live in config file
- No streaming/watch support — pass-through by design
- No auto-enable on install — explicit `prettyout enable <tool>` required
- `brew` install support in `install` command deferred — registry field is reserved but unimplemented

---

## File Map Summary

| Action | Path |
|---|---|
| Create | `internal/doctor/doctor.go` |
| Create | `internal/doctor/doctor_test.go` |
| Create | `internal/install/install.go` |
| Create | `internal/install/install_test.go` |
| Create | `internal/completion/completion.go` |
| Create | `internal/completion/completion_test.go` |
| Create | `cmd/prettyout/cmd_status.go` |
| Create | `cmd/prettyout/cmd_doctor.go` |
| Create | `cmd/prettyout/cmd_install.go` |
| Create | `cmd/prettyout/cmd_list.go` |
| Create | `cmd/prettyout/cmd_completion.go` |
| Create | `cmd/prettyout/cmd_enable.go` |
| Modify | `internal/registry/registry.go` — add `InstallConfig`, `Install` field |
| Modify | `internal/registry/builtin.yaml` — add `install:` entries |
| Modify | `internal/runner/runner.go` — `--debug` flag, output-arg conflict check |
| Modify | `internal/runner/runner_test.go` — tests for conflict detection |
| Modify | `cmd/prettyout/main.go` — new cases, delegate existing |
