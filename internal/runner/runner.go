package runner

import (
	"strings"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/launcher"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

// Decision describes what prettyout _run should do.
// When Intercept is false, run RealCmd with OriginalArgs unchanged.
// When Intercept is true, run RealCmd with TransformedArgs (output_args appended)
// and pipe its stdout through Plugin.
type Decision struct {
	Intercept       bool
	RealCmd         string
	OriginalArgs    []string
	TransformedArgs []string
	Plugin          string
}

// Decide inspects toolName and args against the registry and config to determine
// whether to intercept. isTTY should be true when os.Stdout is a terminal.
func Decide(toolName string, args []string, reg *registry.Registry, cfg *config.Config, isTTY bool) Decision {
	passthrough := Decision{RealCmd: toolName, OriginalArgs: args}

	if cfg.CIMode == "never" {
		return passthrough
	}
	if cfg.CIMode == "auto" && !isTTY {
		return passthrough
	}

	// Launcher path (e.g. uvx, pipx, npx)
	if lc, isLauncher := reg.Launchers[toolName]; isLauncher {
		toolName, subcommand := launcher.ExtractTool(lc, args)
		if toolName == "" {
			return passthrough
		}
		tc, ok := reg.Tools[toolName]
		if !ok {
			return passthrough
		}
		if !cfg.Enabled[toolName] {
			return passthrough
		}
		if tc.HasPassthroughFlag(args) {
			return passthrough
		}
		if !tc.ShouldIntercept(subcommand) {
			return passthrough
		}
		plugin := resolvePlugin(toolName, tc, cfg)
		return Decision{
			Intercept:       true,
			RealCmd:         passthrough.RealCmd,
			OriginalArgs:    args,
			TransformedArgs: append(args, tc.OutputArgs...),
			Plugin:          plugin,
		}
	}

	// Direct tool path (e.g. ruff, basedpyright)
	tc, ok := reg.Tools[toolName]
	if !ok {
		return passthrough
	}
	if !cfg.Enabled[toolName] {
		return passthrough
	}
	if tc.HasPassthroughFlag(args) {
		return passthrough
	}
	subcommand := firstPositional(args)
	if !tc.ShouldIntercept(subcommand) {
		return passthrough
	}
	plugin := resolvePlugin(toolName, tc, cfg)
	return Decision{
		Intercept:       true,
		RealCmd:         toolName,
		OriginalArgs:    args,
		TransformedArgs: append(args, tc.OutputArgs...),
		Plugin:          plugin,
	}
}

func resolvePlugin(toolName string, tc registry.ToolConfig, cfg *config.Config) string {
	if override, ok := cfg.Plugins[toolName]; ok && override != "" {
		return override
	}
	return tc.Plugin
}

// firstPositional returns the first arg that does not start with '-'.
func firstPositional(args []string) string {
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			return a
		}
	}
	return ""
}
