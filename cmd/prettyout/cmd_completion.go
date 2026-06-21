package main

import (
	"fmt"
	"os"

	"github.com/gudoshnikovn/prettyout/internal/completion"
	"github.com/gudoshnikovn/prettyout/internal/config"
	"github.com/gudoshnikovn/prettyout/internal/registry"
)

func runCompletion(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "prettyout completion: shell name required (zsh|bash)")
		os.Exit(1)
	}
	reg, _ := registry.LoadBuiltin()
	script, err := completion.Generate(args[0], reg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "prettyout: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(script)
}

// runCompletions handles `prettyout _completions tools` — used by shell completion scripts.
func runCompletions(args []string) {
	if len(args) == 0 || args[0] != "tools" {
		os.Exit(1)
	}
	reg, _ := registry.LoadBuiltin()
	cfg := config.Load()
	reg.Merge(cfg.CustomTools)
	for _, name := range reg.SortedToolNames() {
		fmt.Println(name)
	}
}
