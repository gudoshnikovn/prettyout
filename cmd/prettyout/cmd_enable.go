package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

func runEnable(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "prettyout enable: tool name required")
		os.Exit(1)
	}
	tool := args[0]
	reg, _ := registry.LoadBuiltin()
	cfg := config.Load()
	reg.Merge(cfg.CustomTools)
	if _, ok := reg.Tools[tool]; !ok {
		fmt.Fprintf(os.Stderr, "prettyout: unknown tool %q\n", tool)
		fmt.Fprintln(os.Stderr, "Known tools:", strings.Join(reg.SortedToolNames(), ", "))
		os.Exit(1)
	}
	cfg.Enabled[tool] = true
	config.Save(cfg)
	fmt.Printf("Enabled prettyout for %s\n", tool)

	rc := rcFilePath(shellName())
	data, _ := os.ReadFile(rc)
	if !strings.Contains(string(data), "prettyout hook") {
		fmt.Fprintf(os.Stderr, "⚠ Hook not found in %s — run: prettyout setup\n", rc)
	}
}
