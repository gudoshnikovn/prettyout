package main

import (
	"fmt"
	"os"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/install"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

func runInstall(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "prettyout install: tool name required")
		os.Exit(1)
	}
	name := args[0]
	reg, _ := registry.LoadBuiltin()
	cfg := config.Load()
	reg.Merge(cfg.CustomTools)
	tc, ok := reg.Tools[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "prettyout: unknown tool %q\n", name)
		os.Exit(1)
	}
	if tc.Install.Go == "" {
		fmt.Fprintf(os.Stderr, "prettyout: no install method for %q\n", name)
		os.Exit(1)
	}
	fmt.Printf("Installing %s...\n", tc.Plugin)
	fmt.Printf("→ go install %s@latest\n", tc.Install.Go)
	if err := install.InstallPlugin(tc); err != nil {
		os.Exit(1)
	}
	fmt.Printf("✓ Installed. Run: prettyout enable %s\n", name)
}

func runUpdate(args []string) {
	reg, _ := registry.LoadBuiltin()
	cfg := config.Load()
	reg.Merge(cfg.CustomTools)

	if len(args) > 0 {
		name := args[0]
		tc, ok := reg.Tools[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "prettyout: unknown tool %q\n", name)
			os.Exit(1)
		}
		if tc.Install.Go == "" {
			fmt.Fprintf(os.Stderr, "prettyout: no install method for %q\n", name)
			os.Exit(1)
		}
		fmt.Printf("Updating %s...\n", tc.Plugin)
		fmt.Printf("→ go install %s@latest\n", tc.Install.Go)
		if err := install.InstallPlugin(tc); err != nil {
			os.Exit(1)
		}
		fmt.Println("✓ Done.")
		return
	}

	fmt.Println("Updating installed plugins...")
	for _, name := range reg.SortedToolNames() {
		tc := reg.Tools[name]
		if tc.Install.Go == "" || !install.IsInstalled(tc) {
			continue
		}
		fmt.Printf("  %s ... ", tc.Plugin)
		if err := install.InstallPlugin(tc); err != nil {
			fmt.Println("✗")
		} else {
			fmt.Println("✓")
		}
	}
	fmt.Println("Done.")
}

func runUpgrade() {
	if err := install.UpgradeSelf(); err != nil {
		fmt.Fprintf(os.Stderr, "prettyout: upgrade failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ prettyout updated.")
}
