package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

// Check is the result of a single doctor check.
type Check struct {
	Name    string
	OK      bool
	Message string
	Hint    string
}

// Run executes all checks and returns results. No I/O — callers format and print.
func Run(reg *registry.Registry, cfg *config.Config) []Check {
	var checks []Check

	shell := detectShell()
	rc := rcFilePath(shell)
	checks = append(checks, checkHook(rc))

	for _, name := range reg.SortedToolNames() {
		if !cfg.Enabled[name] {
			continue
		}
		checks = append(checks, checkPlugin(name, reg.Tools[name]))
	}

	home, _ := os.UserHomeDir()
	checks = append(checks, checkConfigFile(filepath.Join(home, ".config", "prettyout", "config.yaml"), "global"))

	cwd, _ := os.Getwd()
	checks = append(checks, checkConfigFile(filepath.Join(cwd, ".prettyout.yaml"), "project"))

	return checks
}

func checkHook(rcPath string) Check {
	data, err := os.ReadFile(rcPath)
	if err != nil {
		return Check{
			Name:    "hook",
			OK:      false,
			Message: fmt.Sprintf("Shell hook not found in %s", rcPath),
			Hint:    "prettyout setup",
		}
	}
	if strings.Contains(string(data), "prettyout hook") {
		return Check{Name: "hook", OK: true, Message: fmt.Sprintf("Shell hook present in %s", rcPath)}
	}
	return Check{
		Name:    "hook",
		OK:      false,
		Message: fmt.Sprintf("Shell hook not found in %s", rcPath),
		Hint:    "prettyout setup",
	}
}

func checkPlugin(name string, tc registry.ToolConfig) Check {
	_, err := exec.LookPath(tc.Plugin)
	if err == nil {
		return Check{Name: "plugin-" + name, OK: true, Message: tc.Plugin + " found in PATH"}
	}
	return Check{
		Name:    "plugin-" + name,
		OK:      false,
		Message: tc.Plugin + " not found in PATH",
		Hint:    "prettyout install " + name,
	}
}

func checkConfigFile(path, label string) Check {
	data, err := os.ReadFile(path)
	if err != nil {
		return Check{Name: "config-" + label, OK: true, Message: label + " config: not present (OK)"}
	}
	var raw interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Check{
			Name:    "config-" + label,
			OK:      false,
			Message: fmt.Sprintf("Config parse error: %s", path),
			Hint:    err.Error(),
		}
	}
	return Check{Name: "config-" + label, OK: true, Message: label + " config: OK"}
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if i := strings.LastIndexByte(shell, '/'); i >= 0 {
		shell = shell[i+1:]
	}
	if shell == "" {
		return "zsh"
	}
	return shell
}

func rcFilePath(shell string) string {
	home, _ := os.UserHomeDir()
	if shell == "bash" {
		return filepath.Join(home, ".bashrc")
	}
	return filepath.Join(home, ".zshrc")
}
