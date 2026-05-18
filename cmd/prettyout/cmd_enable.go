package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/install"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

func runEnable(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "prettyout enable: tool name required (or --all)")
		os.Exit(1)
	}

	reg, _ := registry.LoadBuiltin()
	cfg := config.Load()
	reg.Merge(cfg.CustomTools)

	if args[0] == "--all" {
		enableAll(reg, cfg)
		return
	}

	tool := args[0]
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

func enableAll(reg *registry.Registry, cfg *config.Config) {
	var enabled []string
	var skipped []string

	for _, name := range reg.SortedToolNames() {
		tc := reg.Tools[name]
		_, toolFound := exec.LookPath(name)
		pluginFound := install.IsInstalled(tc)
		if toolFound == nil && pluginFound {
			cfg.Enabled[name] = true
			enabled = append(enabled, name)
		} else {
			skipped = append(skipped, name)
		}
	}

	if len(enabled) == 0 {
		fmt.Println("No tools enabled — install the tool and its plugin first.")
		return
	}

	config.Save(cfg)
	fmt.Printf("Enabled: %s\n", strings.Join(enabled, ", "))
	if len(skipped) > 0 {
		fmt.Printf("Skipped (not installed): %s\n", strings.Join(skipped, ", "))
	}
}
