package runner

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/launcher"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

type Decision struct {
	Intercept       bool
	RealCmd         string
	OriginalArgs    []string
	TransformedArgs []string
	Plugin          string
}

func Decide(toolName string, args []string, reg *registry.Registry, cfg *config.Config, isTTY bool, debug bool) Decision {
	passthrough := Decision{RealCmd: toolName, OriginalArgs: args}

	if debug {
		fmt.Fprintf(os.Stderr, "[prettyout] tool=%s args=%v\n", toolName, args)
	}

	if cfg.CIMode == "never" {
		if debug {
			fmt.Fprintf(os.Stderr, "[prettyout] ci_mode=never → passthrough\n")
		}
		return passthrough
	}
	if cfg.CIMode == "auto" && !isTTY {
		if debug {
			fmt.Fprintf(os.Stderr, "[prettyout] ci_mode=auto tty=false → passthrough\n")
		}
		return passthrough
	}
	if debug {
		fmt.Fprintf(os.Stderr, "[prettyout] ci_mode=%s tty=%v → proceeding\n", cfg.CIMode, isTTY)
	}

	if lc, isLauncher := reg.Launchers[toolName]; isLauncher {
		innerName, subcommand := launcher.ExtractTool(lc, args)
		if innerName == "" {
			if debug {
				fmt.Fprintf(os.Stderr, "[prettyout] launcher: tool not found in args → passthrough\n")
			}
			return passthrough
		}
		tc, ok := reg.Tools[innerName]
		if !ok {
			if debug {
				fmt.Fprintf(os.Stderr, "[prettyout] launcher: %s not in registry → passthrough\n", innerName)
			}
			return passthrough
		}
		if !cfg.Enabled[innerName] {
			if debug {
				fmt.Fprintf(os.Stderr, "[prettyout] enabled=false → passthrough\n")
			}
			return passthrough
		}
		if debug {
			fmt.Fprintf(os.Stderr, "[prettyout] enabled=true\n")
		}
		if tc.HasPassthroughFlag(args) {
			if debug {
				fmt.Fprintf(os.Stderr, "[prettyout] passthrough_flags: matched → passthrough\n")
			}
			return passthrough
		}
		if debug {
			fmt.Fprintf(os.Stderr, "[prettyout] passthrough_flags: none matched\n")
		}
		if hasOutputArgConflict(tc.OutputArgs, args) {
			if debug {
				fmt.Fprintf(os.Stderr, "[prettyout] output_args conflict: found → passthrough\n")
			}
			return passthrough
		}
		if debug {
			fmt.Fprintf(os.Stderr, "[prettyout] output_args conflict: none\n")
		}
		if !tc.ShouldIntercept(subcommand) {
			if debug {
				fmt.Fprintf(os.Stderr, "[prettyout] subcommand=%s → no match → passthrough\n", subcommand)
			}
			return passthrough
		}
		if debug {
			fmt.Fprintf(os.Stderr, "[prettyout] subcommand=%s → match → intercepting\n", subcommand)
		}
		plugin := resolvePlugin(innerName, tc, cfg)
		if debug {
			fmt.Fprintf(os.Stderr, "[prettyout] plugin=%s\n", plugin)
		}
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

	tc, ok := reg.Tools[toolName]
	if !ok {
		if debug {
			fmt.Fprintf(os.Stderr, "[prettyout] %s not in registry → passthrough\n", toolName)
		}
		return passthrough
	}
	if !cfg.Enabled[toolName] {
		if debug {
			fmt.Fprintf(os.Stderr, "[prettyout] enabled=false → passthrough\n")
		}
		return passthrough
	}
	if debug {
		fmt.Fprintf(os.Stderr, "[prettyout] enabled=true\n")
	}
	if tc.HasPassthroughFlag(args) {
		if debug {
			fmt.Fprintf(os.Stderr, "[prettyout] passthrough_flags: matched → passthrough\n")
		}
		return passthrough
	}
	if debug {
		fmt.Fprintf(os.Stderr, "[prettyout] passthrough_flags: none matched\n")
	}
	if hasOutputArgConflict(tc.OutputArgs, args) {
		if debug {
			fmt.Fprintf(os.Stderr, "[prettyout] output_args conflict: found → passthrough\n")
		}
		return passthrough
	}
	if debug {
		fmt.Fprintf(os.Stderr, "[prettyout] output_args conflict: none\n")
	}
	subcommand := firstPositional(args)
	if !tc.ShouldIntercept(subcommand) {
		if debug {
			fmt.Fprintf(os.Stderr, "[prettyout] subcommand=%s → no match → passthrough\n", subcommand)
		}
		return passthrough
	}
	if debug {
		fmt.Fprintf(os.Stderr, "[prettyout] subcommand=%s → match → intercepting\n", subcommand)
	}
	plugin := resolvePlugin(toolName, tc, cfg)
	if debug {
		fmt.Fprintf(os.Stderr, "[prettyout] plugin=%s\n", plugin)
	}
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

func hasOutputArgConflict(outputArgs, userArgs []string) bool {
	for _, oa := range outputArgs {
		if i := strings.IndexByte(oa, '='); i >= 0 {
			key := oa[:i]
			for _, ua := range userArgs {
				if strings.HasPrefix(ua, key) {
					return true
				}
			}
		} else {
			for _, ua := range userArgs {
				if ua == oa {
					return true
				}
			}
		}
	}
	return false
}

func resolvePlugin(toolName string, tc registry.ToolConfig, cfg *config.Config) string {
	if override, ok := cfg.Plugins[toolName]; ok && override != "" {
		return override
	}
	return tc.Plugin
}

func firstPositional(args []string) string {
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			return a
		}
	}
	return ""
}

func RunFromArgs(args []string) int {
	debug := false
	if len(args) > 0 && args[0] == "--debug" {
		debug = true
		args = args[1:]
	}
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

	decision := Decide(toolName, toolArgs, reg, cfg, isTTY(), debug)
	return Execute(decision)
}

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
		pluginCmd.Stdin.(io.Closer).Close()
		toolCmd.Wait()
		return 1
	}

	toolCmd.Wait()
	pluginCmd.Wait()

	stderrFile.Seek(0, 0)
	io.Copy(os.Stderr, stderrFile)

	if toolCmd.ProcessState != nil {
		return toolCmd.ProcessState.ExitCode()
	}
	return 1
}

func isTTY() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
