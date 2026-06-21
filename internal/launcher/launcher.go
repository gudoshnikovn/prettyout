package launcher

import (
	"strings"

	"github.com/gudoshnikovn/prettyout/internal/registry"
)

// ExtractTool parses launcher arguments to find the tool name and first
// positional argument after it (subcommand). Returns ("", "") if no tool found.
func ExtractTool(lc registry.LauncherConfig, args []string) (toolName, subcommand string) {
	remaining := args

	// consume prefix_args (e.g. "run" for pipx)
	for _, prefix := range lc.PrefixArgs {
		if len(remaining) > 0 && remaining[0] == prefix {
			remaining = remaining[1:]
		}
	}

	skipFlags := make(map[string]bool, len(lc.SkipFlags))
	for _, f := range lc.SkipFlags {
		skipFlags[f] = true
	}
	valueFlags := make(map[string]bool, len(lc.ValueFlags))
	for _, f := range lc.ValueFlags {
		valueFlags[f] = true
	}

	toolFound := false
	skipNext := false
	for _, arg := range remaining {
		if skipNext {
			skipNext = false
			continue
		}
		if valueFlags[arg] {
			skipNext = true
			continue
		}
		if skipFlags[arg] {
			continue
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if !toolFound {
			toolName = strings.SplitN(arg, "@", 2)[0]
			toolFound = true
			continue
		}
		subcommand = arg
		return
	}
	return
}
