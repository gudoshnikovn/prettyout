package hook

import (
	"fmt"
	"strings"

	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

// Generate returns shell code to be eval'd by the user's shell.
func Generate(shell string, tools map[string]registry.ToolConfig) string {
	var b strings.Builder

	for name, cfg := range tools {
		flags := strings.Join(cfg.JSONFlags, " ")
		plugin := cfg.PluginName

		switch shell {
		case "zsh":
			fmt.Fprintf(&b, `
%s() {
  if prettyout _enabled %s 2>/dev/null; then
    command %s "$@" %s 2>&1 | %s
    local ret=${pipestatus[1]}
    return $ret
  fi
  command %s "$@"
}
`, name, name, name, flags, plugin, name)

		case "bash":
			fmt.Fprintf(&b, `
%s() {
  if prettyout _enabled %s 2>/dev/null; then
    command %s "$@" %s 2>&1 | %s
    return ${PIPESTATUS[0]}
  fi
  command %s "$@"
}
`, name, name, name, flags, plugin, name)
		}
	}

	return b.String()
}
