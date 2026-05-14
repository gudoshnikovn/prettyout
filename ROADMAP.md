# prettyout Roadmap

## v0.1 — MVP ✅
- Shell hook (zsh/bash) via `prettyout setup`
- Plugins: `prettyout-ruff`, `prettyout-basedpyright`
- Commands: `setup`, `hook`, `enable`, `disable`, `list`, `_enabled`
- Plugin contract via `pkg/formatter`

---

## v0.2 — Better UX
- `prettyout enable <tool>` автоматически добавляет хук в rc-файл если его нет
- `prettyout install <plugin>` запускает `go install` за пользователя
- `prettyout list --available` показывает все плагины из registry.yaml (не только установленные)
- Auto-discover установленных плагинов `prettyout-*` из `$PATH`

---

## v0.3 — More Plugins

| Plugin | Tool | JSON flag |
|--------|------|-----------|
| `prettyout-eslint` | eslint | `--format json` |
| `prettyout-golangci` | golangci-lint | `--out-format json` |
| `prettyout-mypy` | mypy | `--output json` |
| `prettyout-shellcheck` | shellcheck | `--format json` |
| `prettyout-cargo-clippy` | cargo clippy | `--message-format json` |
| `prettyout-semgrep` | semgrep | `--json` |
| `prettyout-bandit` | bandit | `-f json` |
| `prettyout-trivy` | trivy | `--format json` |
| `prettyout-npm-audit` | npm audit | `--json` |
| `prettyout-hadolint` | hadolint | `--format json` |

---

## v0.4 — Power Features
- Фильтрация по кодам: `prettyout enable ruff --only E501,F401`
- Per-project конфиг: `.prettyout.yaml` в корне проекта (override глобального)
- TUI для управления плагинами (`prettyout tui`)
- `prettyout run <tool> [args]` как явная альтернатива хуку

---

## v1.0 — Community & Ecosystem
- Документация и шаблон для написания плагинов
- GitHub Actions workflow для публикации community-плагинов
- `prettyout search <keyword>` — поиск по registry
- IDE интеграции (VS Code extension)
