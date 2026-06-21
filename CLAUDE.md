# prettyout — Claude Instructions

## Project Overview

`prettyout` intercepts CLI linters/scanners, runs them with JSON output, and pipes the JSON through a formatter plugin that prints a grouped, colorized human-readable summary. 

- **Architecture**: `docs/architecture.md`
- **Roadmap**: `ROADMAP.md`

## Writing a Plugin

Adding a new tool is incredibly simple thanks to the generic render engine:
1. Create `cmd/prettyout-<tool>/main.go`.
2. Unmarshal the tool's JSON output.
3. Convert the data to `[]formatter.Issue`.
4. Call `formatter.Render(issues, cfg)`.
5. Register the tool in `internal/registry/builtin.yaml`.

The generic render engine (`pkg/formatter/render.go`) automatically handles grouping, filtering, sorting, summary generation, and colors.
