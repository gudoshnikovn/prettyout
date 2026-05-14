package hook

import (
	"strings"
	"testing"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

func minimalConfig() *config.Config {
	return &config.Config{
		Enabled:     map[string]bool{},
		Plugins:     map[string]string{},
		Settings:    map[string]config.FormatterSettings{},
		CustomTools: map[string]registry.ToolConfig{},
		CIMode:      "always", // no TTY check, simplest for tests
	}
}

func TestGenerate_containsToolFunction(t *testing.T) {
	reg := &registry.Registry{
		Tools: map[string]registry.ToolConfig{
			"ruff": {
				Plugin:    "prettyout-ruff",
				JSONFlags: []string{"--output-format=json"},
			},
		},
		Launchers: map[string]registry.LauncherConfig{},
	}
	got := Generate("zsh", reg, minimalConfig())
	if !strings.Contains(got, "ruff()") {
		t.Error("expected ruff() function in hook")
	}
	if !strings.Contains(got, "prettyout _enabled ruff") {
		t.Error("expected prettyout _enabled ruff check")
	}
	if !strings.Contains(got, "--output-format=json") {
		t.Error("expected JSON flag in hook")
	}
	if !strings.Contains(got, "prettyout-ruff") {
		t.Error("expected plugin name in hook")
	}
	if !strings.Contains(got, "command ruff") {
		t.Error("expected fallback 'command ruff'")
	}
}

func TestGenerate_subcommandCheck(t *testing.T) {
	reg := &registry.Registry{
		Tools: map[string]registry.ToolConfig{
			"ruff": {
				Plugin:               "prettyout-ruff",
				JSONFlags:            []string{"--output-format=json"},
				InterceptSubcommands: []string{"check"},
			},
		},
		Launchers: map[string]registry.LauncherConfig{},
	}
	got := Generate("zsh", reg, minimalConfig())
	if !strings.Contains(got, `case "${1:-}"`) {
		t.Error("expected subcommand case statement")
	}
	if !strings.Contains(got, "check)") {
		t.Error("expected check) case")
	}
}

func TestGenerate_passthroughFlag(t *testing.T) {
	reg := &registry.Registry{
		Tools: map[string]registry.ToolConfig{
			"ruff": {
				Plugin:               "prettyout-ruff",
				JSONFlags:            []string{"--output-format=json"},
				InterceptSubcommands: []string{"check"},
				PassthroughFlags:     []string{"--watch", "-W"},
			},
		},
		Launchers: map[string]registry.LauncherConfig{},
	}
	got := Generate("zsh", reg, minimalConfig())
	if !strings.Contains(got, "--watch") {
		t.Error("expected --watch passthrough check")
	}
}

func TestGenerate_ciModeAuto_addsTTYCheck(t *testing.T) {
	reg := &registry.Registry{
		Tools: map[string]registry.ToolConfig{
			"ruff": {Plugin: "prettyout-ruff", JSONFlags: []string{"--output-format=json"}},
		},
		Launchers: map[string]registry.LauncherConfig{},
	}
	cfg := minimalConfig()
	cfg.CIMode = "auto"
	got := Generate("zsh", reg, cfg)
	if !strings.Contains(got, "[[ -t 1 ]]") {
		t.Error("ci_mode=auto should add TTY check")
	}
}

func TestGenerate_pluginOverride(t *testing.T) {
	reg := &registry.Registry{
		Tools: map[string]registry.ToolConfig{
			"ruff": {Plugin: "prettyout-ruff", JSONFlags: []string{"--output-format=json"}},
		},
		Launchers: map[string]registry.LauncherConfig{},
	}
	cfg := minimalConfig()
	cfg.Plugins["ruff"] = "/home/user/scripts/my-ruff.py"
	got := Generate("zsh", reg, cfg)
	if !strings.Contains(got, "/home/user/scripts/my-ruff.py") {
		t.Error("expected plugin override path in hook")
	}
}

func TestGenerate_stderrSeparated(t *testing.T) {
	reg := &registry.Registry{
		Tools: map[string]registry.ToolConfig{
			"ruff": {Plugin: "prettyout-ruff", JSONFlags: []string{"--output-format=json"}},
		},
		Launchers: map[string]registry.LauncherConfig{},
	}
	got := Generate("zsh", reg, minimalConfig())
	if strings.Contains(got, "2>&1") {
		t.Error("hook must not use 2>&1 — stderr must be separated")
	}
	if !strings.Contains(got, "mktemp") {
		t.Error("expected mktemp for stderr temp file")
	}
}
