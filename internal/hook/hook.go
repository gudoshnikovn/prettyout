package hook

import (
	"fmt"
	"strings"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

// Generate returns shell code to be eval'd by the user's shell.
func Generate(shell string, reg *registry.Registry, cfg *config.Config) string {
	if cfg.CIMode == "never" {
		return ""
	}

	var b strings.Builder
	names := reg.SortedToolNames()

	for _, name := range names {
		tc := reg.Tools[name]
		plugin := resolvePlugin(name, tc, cfg)
		writeTool(&b, shell, name, tc, plugin, cfg.CIMode)
	}

	// collect tools per launcher and generate launcher wrappers
	byLauncher := map[string][]string{}
	for _, name := range names {
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
	default:
		fmt.Fprintf(b, "%scat \"$_ef\" >&2; rm -f \"$_ef\"\n", indent)
	}
}

// writeLauncher generates a wrapper function for launchers like uvx, npx, pipx.
func writeLauncher(b *strings.Builder, shell, launcherName string, lc registry.LauncherConfig, tools []string, reg *registry.Registry, cfg *config.Config) {
	fmt.Fprintf(b, "\n%s() {\n", launcherName)
	fmt.Fprintf(b, "  local _tool_arg=\"\" _skip_next=0\n")

	fmt.Fprintf(b, "  local _args=(\"$@\")\n")
	fmt.Fprintf(b, "  for _a in \"${_args[@]}\"; do\n")
	fmt.Fprintf(b, "    if (( _skip_next )); then _skip_next=0; continue; fi\n")

	if len(lc.ValueFlags) > 0 {
		fmt.Fprintf(b, "    case \"$_a\" in\n")
		fmt.Fprintf(b, "      %s) _skip_next=1; continue ;;\n", strings.Join(lc.ValueFlags, "|"))
		fmt.Fprintf(b, "    esac\n")
	}

	if len(lc.SkipFlags) > 0 {
		fmt.Fprintf(b, "    case \"$_a\" in\n")
		fmt.Fprintf(b, "      %s) continue ;;\n", strings.Join(lc.SkipFlags, "|"))
		fmt.Fprintf(b, "    esac\n")
	}

	if len(lc.PrefixArgs) > 0 {
		fmt.Fprintf(b, "    case \"$_a\" in\n")
		fmt.Fprintf(b, "      %s) continue ;;\n", strings.Join(lc.PrefixArgs, "|"))
		fmt.Fprintf(b, "    esac\n")
	}

	fmt.Fprintf(b, "    [[ \"$_a\" == -* ]] && continue\n")
	fmt.Fprintf(b, "    _tool_arg=\"$_a\"; break\n")
	fmt.Fprintf(b, "  done\n")
	fmt.Fprintf(b, "  local _toolname=\"${_tool_arg%%%%@*}\"\n") // strip @version (%%%% = literal %% in Printf)

	fmt.Fprintf(b, "  case \"$_toolname\" in\n")

	for _, toolName := range tools {
		tc := reg.Tools[toolName]
		plugin := resolvePlugin(toolName, tc, cfg)
		flags := strings.Join(tc.JSONFlags, " ")

		fmt.Fprintf(b, "    %s)\n", toolName)
		fmt.Fprintf(b, "      if prettyout _enabled %s 2>/dev/null; then\n", toolName)

		if len(tc.PassthroughFlags) > 0 {
			fmt.Fprintf(b, "        for _ptf in \"$@\"; do\n")
			fmt.Fprintf(b, "          case \"$_ptf\" in\n")
			fmt.Fprintf(b, "            %s) command %s \"$@\"; return $? ;;\n",
				strings.Join(tc.PassthroughFlags, "|"), launcherName)
			fmt.Fprintf(b, "          esac\n")
			fmt.Fprintf(b, "        done\n")
		}

		if len(tc.InterceptSubcommands) == 0 {
			writeLauncherInterceptBlock(b, shell, launcherName, flags, plugin, "        ")
		} else {
			// find subcommand: first non-flag arg after tool name in original $@
			fmt.Fprintf(b, "        local _past_tool=0 _sub=\"\"\n")
			fmt.Fprintf(b, "        for _a in \"$@\"; do\n")
			fmt.Fprintf(b, "          if (( _past_tool )) && [[ \"$_a\" != -* ]]; then _sub=\"$_a\"; break; fi\n")
			fmt.Fprintf(b, "          [[ \"${_a%%%%@*}\" == %q ]] && _past_tool=1\n", toolName)
			fmt.Fprintf(b, "        done\n")
			fmt.Fprintf(b, "        case \"$_sub\" in\n")
			fmt.Fprintf(b, "          %s)\n", strings.Join(tc.InterceptSubcommands, "|"))
			writeLauncherInterceptBlock(b, shell, launcherName, flags, plugin, "            ")
			fmt.Fprintf(b, "            ;;\n") // terminates the inner case arm
			fmt.Fprintf(b, "        esac\n")
		}

		fmt.Fprintf(b, "      fi ;;\n") // close if + terminate outer case arm
	}

	fmt.Fprintf(b, "  esac\n")
	fmt.Fprintf(b, "  command %s \"$@\"\n", launcherName)
	fmt.Fprintf(b, "}\n")
}

func writeLauncherInterceptBlock(b *strings.Builder, shell, launcherName, flags, plugin, indent string) {
	fmt.Fprintf(b, "%slocal _ef; _ef=$(mktemp)\n", indent)
	fmt.Fprintf(b, "%scommand %s \"$@\" %s 2>\"$_ef\" | %s\n", indent, launcherName, flags, plugin)
	switch shell {
	case "zsh":
		fmt.Fprintf(b, "%slocal _r=${pipestatus[1]}; cat \"$_ef\" >&2; rm -f \"$_ef\"; return $_r\n", indent)
	case "bash":
		fmt.Fprintf(b, "%slocal _r=${PIPESTATUS[0]}; cat \"$_ef\" >&2; rm -f \"$_ef\"; return $_r\n", indent)
	default:
		fmt.Fprintf(b, "%scat \"$_ef\" >&2; rm -f \"$_ef\"\n", indent)
	}
}
