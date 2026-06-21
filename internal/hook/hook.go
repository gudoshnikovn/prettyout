package hook

import (
	"fmt"
	"strings"

	"github.com/gudoshnikovn/prettyout/internal/completion"
	"github.com/gudoshnikovn/prettyout/internal/config"
	"github.com/gudoshnikovn/prettyout/internal/registry"
)

// Generate returns shell code to be eval'd by the user's shell.
// Each tool and launcher gets a one-liner that delegates to `prettyout _run`.
// Completion is included so a single eval handles both interception and tab completion.
func Generate(shell string, reg *registry.Registry, cfg *config.Config) string {
	if cfg.CIMode == "never" {
		return ""
	}

	var b strings.Builder
	names := reg.SortedToolNames()

	for _, name := range names {
		fmt.Fprintf(&b, "\n%s() { prettyout _run %s \"$@\"; }\n", name, name)
	}

	launchers := make(map[string]struct{})
	for _, name := range names {
		for _, l := range reg.Tools[name].Launchers {
			if _, ok := reg.Launchers[l]; ok {
				launchers[l] = struct{}{}
			}
		}
	}
	for l := range launchers {
		fmt.Fprintf(&b, "\n%s() { prettyout _run %s \"$@\"; }\n", l, l)
	}

	if comp, err := completion.Generate(shell, reg); err == nil {
		fmt.Fprintf(&b, "\n%s", comp)
	}

	return b.String()
}
