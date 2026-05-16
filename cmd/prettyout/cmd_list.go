package main

import (
	"fmt"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/install"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

func runList(args []string) {
	available := len(args) > 0 && args[0] == "--available"
	reg, _ := registry.LoadBuiltin()
	cfg := config.Load()
	reg.Merge(cfg.CustomTools)

	if available {
		fmt.Println("Tool             Status     Installed")
		fmt.Println("---------------  ---------  ---------")
		for _, name := range reg.SortedToolNames() {
			tc := reg.Tools[name]
			status := "—"
			if val, exists := cfg.Enabled[name]; exists {
				if val {
					status = "enabled"
				} else {
					status = "disabled"
				}
			}
			installedStr := "✗  prettyout install " + name
			if install.IsInstalled(tc) {
				installedStr = "✓"
			}
			fmt.Printf("%-16s %-10s %s\n", name, status, installedStr)
		}
		return
	}

	fmt.Println("Tool             Status     Plugin")
	fmt.Println("---------------  ---------  ----------------------")
	for _, name := range reg.SortedToolNames() {
		tc := reg.Tools[name]
		status := "disabled"
		if cfg.Enabled[name] {
			status = "enabled"
		}
		fmt.Printf("%-16s %-10s %s\n", name, status, tc.Plugin)
	}
}
