package main

import (
	"fmt"
	"os/exec"

	"github.com/gudoshnikovn/prettyout/internal/config"
	"github.com/gudoshnikovn/prettyout/internal/install"
	"github.com/gudoshnikovn/prettyout/internal/registry"
)

func runList(args []string) {
	available := len(args) > 0 && args[0] == "--available"
	reg, _ := registry.LoadBuiltin()
	cfg := config.Load()
	reg.Merge(cfg.CustomTools)

	if available {
		fmt.Println("Tool             Status     Tool  Plugin")
		fmt.Println("---------------  ---------  ----  ----------------------")
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
			toolBin := "✗"
			if _, err := exec.LookPath(name); err == nil {
				toolBin = "✓"
			}
			pluginBin := "✗  prettyout install " + name
			if install.IsInstalled(tc) {
				pluginBin = "✓"
			}
			fmt.Printf("%-16s %-10s %-4s  %s\n", name, status, toolBin, pluginBin)
		}
		return
	}

	fmt.Println("Tool             Status     Binary  Plugin")
	fmt.Println("---------------  ---------  ------  ----------------------")
	for _, name := range reg.SortedToolNames() {
		tc := reg.Tools[name]
		status := "disabled"
		if cfg.Enabled[name] {
			status = "enabled"
		}
		bin := "✗"
		if _, err := exec.LookPath(name); err == nil {
			bin = "✓"
		}
		fmt.Printf("%-16s %-10s %-6s  %s\n", name, status, bin, tc.Plugin)
	}
}
