package runner

import (
	"io"
	"os"
	"os/exec"
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
		transformed := make([]string, len(args), len(args)+len(tc.OutputArgs))
		copy(transformed, args)
		return Decision{
			Intercept:       true,
			RealCmd:         passthrough.RealCmd,
			OriginalArgs:    args,
			TransformedArgs: append(transformed, tc.OutputArgs...),
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
	transformed := make([]string, len(args), len(args)+len(tc.OutputArgs))
	copy(transformed, args)
	return Decision{
		Intercept:       true,
		RealCmd:         toolName,
		OriginalArgs:    args,
		TransformedArgs: append(transformed, tc.OutputArgs...),
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

// RunFromArgs is the entry point for `prettyout _run <toolName> [args...]`.
// It loads registry and config, calls Decide, executes, and returns the exit code.
func RunFromArgs(args []string) int {
	if len(args) == 0 {
		return 1
	}
	toolName := args[0]
	toolArgs := args[1:]

	reg, err := registry.LoadBuiltin()
	if err != nil {
		return 1
	}
	cfg := config.Load()
	reg.Merge(cfg.CustomTools)

	decision := Decide(toolName, toolArgs, reg, cfg, isTTY())
	return Execute(decision)
}

// Execute runs the decision: passthrough or intercepted pipe.
func Execute(d Decision) int {
	if !d.Intercept {
		cmd := exec.Command(d.RealCmd, d.OriginalArgs...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			if cmd.ProcessState != nil {
				return cmd.ProcessState.ExitCode()
			}
			return 1
		}
		return 0
	}

	stderrFile, err := os.CreateTemp("", "prettyout-stderr-*")
	if err != nil {
		return 1
	}
	defer os.Remove(stderrFile.Name())
	defer stderrFile.Close()

	toolCmd := exec.Command(d.RealCmd, d.TransformedArgs...)
	toolCmd.Stdin = os.Stdin
	toolCmd.Stderr = stderrFile

	pluginCmd := exec.Command(d.Plugin)
	pluginCmd.Stdin, err = toolCmd.StdoutPipe()
	if err != nil {
		return 1
	}
	pluginCmd.Stdout = os.Stdout
	pluginCmd.Stderr = os.Stderr

	if err := toolCmd.Start(); err != nil {
		return 1
	}
	if err := pluginCmd.Start(); err != nil {
		toolCmd.Wait()
		return 1
	}

	toolCmd.Wait()
	pluginCmd.Wait()

	stderrFile.Seek(0, 0)
	io.Copy(os.Stderr, stderrFile)

	return toolCmd.ProcessState.ExitCode()
}

// isTTY reports whether stdout is a terminal.
func isTTY() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
