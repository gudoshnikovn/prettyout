# prettyout Roadmap

## Done ✅

**v0.1 — MVP**
- Shell hook (zsh/bash) via `prettyout setup`
- Plugins: `prettyout-ruff`, `prettyout-basedpyright`
- Commands: `setup`, `hook`, `enable`, `disable`, `list`, `_enabled`
- Plugin contract via `pkg/formatter`

**v0.2 — Registry-driven redesign**
- Registry embedded in binary (`builtin.yaml`) — tool + launcher definitions
- Subcommand-aware hooks: only listed subcommands intercepted (`ruff format` passes through)
- Launcher wrappers: `uvx ruff`, `pipx ruff` intercepted via arg-parsing logic
- Stderr separation via `mktemp` — formatter only sees JSON, stderr goes to terminal
- Formatter config: `colors`, `max_message_length`, extensible `extra` map
- Plugin override: replace any formatter with any executable via `plugins:` in config
- Per-project config: `.prettyout.yaml` overrides global
- `ci_mode: auto` — skips interception when stdout is not a TTY
- `custom_tools:` in config — register tools not in built-in registry

---

## Up Next

### Better CLI UX

- `prettyout install <plugin>` — runs `go install github.com/.../prettyout-<plugin>@latest` for the user
- `prettyout list --available` — shows all plugins from `registry.yaml`, not only installed ones; marks which are installed
- Auto-discover installed `prettyout-*` binaries from `$PATH` and show them in `list`
- `prettyout doctor` — diagnoses problems: plugin not in PATH, hook not in rc file, config parse errors
- `prettyout enable <tool>` — if hook not yet in rc file, offer to add it automatically

### More Launchers

Right now: `uvx`, `pipx`, `npx`. Could add:

| Launcher | Used for |
|----------|----------|
| `uv run` | Python scripts/tools via uv |
| `poetry run` | Tools inside poetry envs |
| `bunx` | Bun JS runtime |
| `pnpx` | pnpm equivalent of npx |
| `deno run` | Deno |

Each needs a launcher entry in `builtin.yaml` with its skip/value flags. The hook generator handles the rest.

### More Output Modes

- `group_by: file` — group diagnostics by file instead of by rule (for some workflows this is more useful)
- Summary line at the end: `3 rules, 12 occurrences across 5 files`
- Error counts per rule in the header
- `--no-color` flag respected even when config says `colors: true` (for quick overrides)

### Filtering

- `prettyout enable ruff --only E501,F401` — only intercept specific rule codes
- `prettyout enable ruff --ignore E501` — suppress specific codes
- Stored in config under `settings.ruff.only` / `settings.ruff.ignore`
- The plugin reads these from `cfg.Extra` and filters before printing

---

## More Plugins

Priority order based on ecosystem popularity:

| Plugin | Tool | JSON flag | Notes |
|--------|------|-----------|-------|
| `prettyout-golangci` | golangci-lint | `--out-format json` | Go linter, used very widely |
| `prettyout-eslint` | eslint | `--format json` | JS/TS linter; intercept `--ext` invocations |
| `prettyout-mypy` | mypy | `--output json` | Python type checker; JSON output added in v1.4 |
| `prettyout-shellcheck` | shellcheck | `--format json` | Shell linter |
| `prettyout-cargo-clippy` | cargo clippy | `--message-format json` | Rust; JSON is per-line (not array), needs special parsing |
| `prettyout-semgrep` | semgrep | `--json` | SAST; large JSON output, group by rule makes sense |
| `prettyout-hadolint` | hadolint | `--format json` | Dockerfile linter |
| `prettyout-trivy` | trivy | `--format json` | Security scanner; intercept `image`, `fs` subcommands |
| `prettyout-npm-audit` | npm audit | `--json` | Node dependency audit |
| `prettyout-bandit` | bandit | `-f json` | Python security linter |

**cargo clippy** is special: its JSON output is one object per line (cargo's machine output format), not a JSON array. The plugin needs to read line by line.

**trivy** has multiple subcommands (`image`, `fs`, `repo`, `config`) — registry entry needs `intercept_subcommands` list.

---

## Community & Ecosystem

- `registry.yaml` as public index of community plugins — anyone can submit a PR
- `prettyout search <keyword>` — searches `registry.yaml` by tool name or description
- Plugin template repository with README, example code, CI setup
- GitHub Actions workflow: tag `prettyout-<tool>@vX.Y.Z` → binary published to releases
- Documentation site (probably just GitHub Pages from `docs/`)

---

## Explicit Non-Goals

Things that are out of scope unless there's a compelling reason:

- **Streaming formatters** — tools with `--watch` are pass-through by design; buffering a stream is a different problem
- **IDE integration** — this is a shell-level tool; IDEs have their own plugin ecosystems
- **JSON schema versioning** — formatters are expected to handle tool version differences themselves
- **TUI** (`prettyout tui`) — adds complexity for marginal benefit; `prettyout list` + `prettyout enable/disable` covers the use case
- **Modifying tool output in place** (e.g. auto-fix driven from pretty output) — out of scope
