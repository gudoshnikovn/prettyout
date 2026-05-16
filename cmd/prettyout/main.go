package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/hook"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
	"github.com/gudoshnikov_na/prettyout/internal/runner"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "setup":
		runSetup()
	case "hook":
		runHook(args)
	case "enable":
		runEnable(args)
	case "disable":
		runDisable(args)
	case "list":
		runList(args)
	case "_enabled":
		runEnabled(args)
	case "_run":
		os.Exit(runner.RunFromArgs(args))
	case "status":
		runStatus()
	default:
		fmt.Fprintf(os.Stderr, "prettyout: unknown command %q\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage:
  prettyout setup              Add shell hook to your rc file
  prettyout hook <shell>       Print shell hook code (zsh|bash)
  prettyout enable <tool>      Enable pretty output for a tool
  prettyout disable <tool>     Disable pretty output for a tool
  prettyout list               Show all tools and their status
  prettyout _enabled <tool>    Exit 0 if tool is enabled (internal)`)
}

func runSetup() {
	shell := shellName()
	rcFile := rcFilePath(shell)

	line := fmt.Sprintf(`eval "$(prettyout hook %s)"`, shell)

	data, _ := os.ReadFile(rcFile)
	if strings.Contains(string(data), "prettyout hook") {
		fmt.Println("prettyout hook already present in", rcFile)
		return
	}

	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "prettyout: cannot open %s: %v\n", rcFile, err)
		os.Exit(1)
	}
	defer f.Close()
	fmt.Fprintf(f, "\n# prettyout\n%s\n", line)
	fmt.Printf("Added to %s\nRun: source %s\n", rcFile, rcFile)
}

func runHook(args []string) {
	shell := "zsh"
	if len(args) > 0 {
		shell = args[0]
	}
	reg, err := registry.LoadBuiltin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "prettyout: failed to load registry: %v\n", err)
		os.Exit(1)
	}
	cfg := config.Load()
	reg.Merge(cfg.CustomTools)
	fmt.Print(hook.Generate(shell, reg, cfg))
}

func runDisable(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "prettyout disable: tool name required")
		os.Exit(1)
	}
	cfg := config.Load()
	cfg.Enabled[args[0]] = false
	config.Save(cfg)
	fmt.Printf("Disabled prettyout for %s\n", args[0])
}

func runEnabled(args []string) {
	if len(args) == 0 {
		os.Exit(1)
	}
	cfg := config.Load()
	if cfg.Enabled[args[0]] {
		os.Exit(0)
	}
	os.Exit(1)
}

func shellName() string {
	shell := shellBase(os.Getenv("SHELL"))
	if shell == "" {
		return "zsh"
	}
	return shell
}

func shellBase(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

func rcFilePath(shell string) string {
	home, _ := os.UserHomeDir()
	switch shell {
	case "zsh":
		return home + "/.zshrc"
	case "bash":
		return home + "/.bashrc"
	default:
		fmt.Fprintf(os.Stderr, "prettyout: unsupported shell %q (only zsh and bash)\n", shell)
		os.Exit(1)
		return ""
	}
}
