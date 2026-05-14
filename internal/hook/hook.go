package hook

import (
	"fmt"
	"strings"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

// Generate returns shell code to be eval'd by the user's shell.
func Generate(shell string, reg *registry.Registry, cfg *config.Config) string {
	var b strings.Builder

	for _, name := range reg.SortedToolNames() {
		tc := reg.Tools[name]
		plugin := resolvePlugin(name, tc, cfg)
		writeTool(&b, shell, name, tc, plugin, cfg.CIMode)
	}

	// collect tools per launcher and generate launcher wrappers
	byLauncher := map[string][]string{}
	for _, name := range reg.SortedToolNames() {
		tc := reg.Tools[name]
		for _, l := range tc.Launchers {
			if _, ok := reg.Launchers[l]; ok {
				byLauncher[l] = append(byLauncher[l], name)
			}
		}
	}
	for launcherName, tools := range byLauncher {
		lc := reg.Launchers[launcherName]
		writeLauncher(&b, shell, launcherName, lc, tools, reg, cfg)
	}

	return b.String()
}

func resolvePlugin(toolName string, tc registry.ToolConfig, cfg *config.Config) string {
	if override, ok := cfg.Plugins[toolName]; ok && override != "" {
		return override
	}
	return tc.Plugin
}

func writeTool(b *strings.Builder, shell, name string, tc registry.ToolConfig, plugin, ciMode string) {
	flags := strings.Join(tc.JSONFlags, " ")

	fmt.Fprintf(b, "\n%s() {\n", name)

	if ciMode == "auto" {
		fmt.Fprintf(b, "  [[ -t 1 ]] || { command %s \"$@\"; return $?; }\n", name)
	}

	fmt.Fprintf(b, "  if prettyout _enabled %s 2>/dev/null; then\n", name)

	if len(tc.PassthroughFlags) > 0 {
		fmt.Fprintf(b, "    for _ptf in \"$@\"; do\n")
		fmt.Fprintf(b, "      case \"$_ptf\" in\n")
		fmt.Fprintf(b, "        %s) command %s \"$@\"; return $? ;;\n",
			strings.Join(tc.PassthroughFlags, "|"), name)
		fmt.Fprintf(b, "      esac\n")
		fmt.Fprintf(b, "    done\n")
	}

	if len(tc.InterceptSubcommands) == 0 {
		// intercept all subcommands — no case statement, no ;;
		writeInterceptBlock(b, shell, name, flags, plugin, "    ")
	} else {
		fmt.Fprintf(b, "    case \"${1:-}\" in\n")
		fmt.Fprintf(b, "      %s)\n", strings.Join(tc.InterceptSubcommands, "|"))
		writeInterceptBlock(b, shell, name, flags, plugin, "        ")
		fmt.Fprintf(b, "        ;;\n") // terminates the case arm
		fmt.Fprintf(b, "    esac\n")
	}

	fmt.Fprintf(b, "  fi\n")
	fmt.Fprintf(b, "  command %s \"$@\"\n", name)
	fmt.Fprintf(b, "}\n")
}

// writeInterceptBlock emits the mktemp+pipe+return block. No ;; — caller adds it when inside a case.
func writeInterceptBlock(b *strings.Builder, shell, toolName, flags, plugin, indent string) {
	fmt.Fprintf(b, "%slocal _ef; _ef=$(mktemp)\n", indent)
	fmt.Fprintf(b, "%scommand %s \"$@\" %s 2>\"$_ef\" | %s\n", indent, toolName, flags, plugin)
	switch shell {
	case "zsh":
		fmt.Fprintf(b, "%slocal _r=${pipestatus[1]}; cat \"$_ef\" >&2; rm -f \"$_ef\"; return $_r\n", indent)
	case "bash":
		fmt.Fprintf(b, "%slocal _r=${PIPESTATUS[0]}; cat \"$_ef\" >&2; rm -f \"$_ef\"; return $_r\n", indent)
	}
}

// writeLauncher generates a wrapper function for launchers like uvx, npx, pipx.
// Implemented in Task 6.
func writeLauncher(b *strings.Builder, shell, launcherName string, lc registry.LauncherConfig, tools []string, reg *registry.Registry, cfg *config.Config) {
	// stub — implemented in Task 6
}
