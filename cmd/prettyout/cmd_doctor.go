package main

import (
	"fmt"
	"os"

	"github.com/gudoshnikovn/prettyout/internal/config"
	"github.com/gudoshnikovn/prettyout/internal/doctor"
)

func runDoctor() {
	reg := mustLoadBuiltin()
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
