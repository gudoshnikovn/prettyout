package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gudoshnikovn/prettyout/internal/config"
	"github.com/gudoshnikovn/prettyout/internal/install"
	"github.com/gudoshnikovn/prettyout/internal/registry"
)

const version = "0.3.0"

func runStatus() {
	reg, _ := registry.LoadBuiltin()
	cfg := config.Load()
	reg.Merge(cfg.CustomTools)

	shell := shellName()
	rc := rcFilePath(shell)
	hookStatus := "✗"
	if data, err := os.ReadFile(rc); err == nil && strings.Contains(string(data), "prettyout hook") {
		hookStatus = "✓"
	}

	fmt.Printf("prettyout v%s\n\n", version)
	fmt.Printf("Shell:   %s\n", shell)
	fmt.Printf("Hook:    %s %s\n", hookStatus, rc)
	fmt.Printf("Config:  %s\n\n", config.GlobalConfigPath())
	fmt.Println("Tools:")
	for _, name := range reg.SortedToolNames() {
		tc := reg.Tools[name]
		status := "disabled"
		if cfg.Enabled[name] {
			status = "enabled"
		}
		instStatus := "✗ not found"
		if install.IsInstalled(tc) {
			instStatus = "✓ installed"
		}
		fmt.Printf("  %-16s %-10s %-25s %s\n", name, status, tc.Plugin, instStatus)
	}
}
