package main

import (
	"fmt"
	"os"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/doctor"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

func runDoctor() {
	reg, _ := registry.LoadBuiltin()
	cfg := config.Load()
	reg.Merge(cfg.CustomTools)

	checks := doctor.Run(reg, cfg)
	issues := 0
	for _, c := range checks {
		if c.OK {
			fmt.Printf("✓ %s\n", c.Message)
		} else {
			fmt.Printf("✗ %s\n", c.Message)
			if c.Hint != "" {
				fmt.Printf("  → %s\n", c.Hint)
			}
			issues++
		}
	}
	if issues > 0 {
		fmt.Printf("\n%d issue(s) found.\n", issues)
		os.Exit(1)
	}
}
