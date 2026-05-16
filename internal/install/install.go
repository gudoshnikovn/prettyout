package install

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

// IsInstalled reports whether the plugin binary for tc is found in PATH.
func IsInstalled(tc registry.ToolConfig) bool {
	_, err := exec.LookPath(tc.Plugin)
	return err == nil
}

// InstallPlugin runs `go install` for the tool's module path.
func InstallPlugin(tc registry.ToolConfig) error {
	if tc.Install.Go == "" {
		return fmt.Errorf("no install.go path for plugin %q", tc.Plugin)
	}
	cmd := exec.Command("go", "install", tc.Install.Go+"@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type upgrader struct {
	name    string
	cmdStr  string
	detect  func() bool
	upgrade func() error
}

var upgraders = []upgrader{
	{
		name:   "brew upgrade",
		cmdStr: "brew upgrade prettyout",
		detect: func() bool {
			return exec.Command("brew", "list", "prettyout").Run() == nil
		},
		upgrade: func() error {
			cmd := exec.Command("brew", "upgrade", "prettyout")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		},
	},
	{
		name:   "go install",
		cmdStr: "go install github.com/gudoshnikov_na/prettyout/cmd/prettyout@latest",
		detect: func() bool { return true },
		upgrade: func() error {
			cmd := exec.Command("go", "install", "github.com/gudoshnikov_na/prettyout/cmd/prettyout@latest")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		},
	},
}

// UpgradeSelf detects the package manager and upgrades prettyout.
// Prints the detected method and command to stdout before running.
func UpgradeSelf() error {
	for _, u := range upgraders {
		if u.detect() {
			fmt.Printf("Detected: %s\n", u.name)
			fmt.Printf("→ %s\n", u.cmdStr)
			return u.upgrade()
		}
	}
	return fmt.Errorf("no package manager detected")
}
