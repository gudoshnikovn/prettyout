package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/hook"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
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
		runList()
	case "_enabled":
		runEnabled(args)
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
	shell := filepath.Base(os.Getenv("SHELL"))
	if shell == "" {
		shell = "zsh"
	}

	var rcFile string
	home, _ := os.UserHomeDir()
	switch shell {
	case "zsh":
		rcFile = filepath.Join(home, ".zshrc")
	case "bash":
		rcFile = filepath.Join(home, ".bashrc")
	default:
		fmt.Fprintf(os.Stderr, "prettyout: unsupported shell %q (only zsh and bash)\n", shell)
		os.Exit(1)
	}

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
	fmt.Print(hook.Generate(shell, registry.Tools))
}

func runEnable(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "prettyout enable: tool name required")
		os.Exit(1)
	}
	tool := args[0]
	if _, ok := registry.Tools[tool]; !ok {
		fmt.Fprintf(os.Stderr, "prettyout: unknown tool %q\n", tool)
		fmt.Fprintln(os.Stderr, "Known tools:", knownTools())
		os.Exit(1)
	}
	cfg := config.Load()
	cfg.Hooks[tool] = true
	config.Save(cfg)
	fmt.Printf("Enabled prettyout for %s\n", tool)
}

func runDisable(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "prettyout disable: tool name required")
		os.Exit(1)
	}
	tool := args[0]
	cfg := config.Load()
	cfg.Hooks[tool] = false
	config.Save(cfg)
	fmt.Printf("Disabled prettyout for %s\n", tool)
}

func runList() {
	cfg := config.Load()
	fmt.Println("Tool             Status")
	fmt.Println("---------------  -------")
	tools := registry.SortedTools()
	for _, name := range tools {
		status := "disabled"
		if cfg.Hooks[name] {
			status = "enabled"
		}
		fmt.Printf("%-16s %s\n", name, status)
	}
}

func runEnabled(args []string) {
	if len(args) == 0 {
		os.Exit(1)
	}
	cfg := config.Load()
	if cfg.Hooks[args[0]] {
		os.Exit(0)
	}
	os.Exit(1)
}

func knownTools() string {
	return strings.Join(registry.SortedTools(), ", ")
}
