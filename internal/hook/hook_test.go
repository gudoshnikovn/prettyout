package hook

import (
	"strings"
	"testing"

	"github.com/gudoshnikovn/prettyout/internal/config"
	"github.com/gudoshnikovn/prettyout/internal/registry"
)

func minimalConfig() *config.Config {
	return &config.Config{
		Enabled:     map[string]bool{},
		Plugins:     map[string]string{},
		Settings:    map[string]config.FormatterSettings{},
		CustomTools: map[string]registry.ToolConfig{},
		CIMode:      "always",
	}
}

func ruffReg() *registry.Registry {
	return &registry.Registry{
		Tools: map[string]registry.ToolConfig{
			"ruff": {
				Plugin:     "prettyout-ruff",
				OutputArgs: []string{"--output-format=json"},
				Launchers:  []string{"uvx"},
			},
		},
		Launchers: map[string]registry.LauncherConfig{
			"uvx": {},
		},
	}
}

func TestGenerate_toolOneLiner(t *testing.T) {
	got := Generate("zsh", ruffReg(), minimalConfig())
	want := "ruff() { prettyout _run ruff \"$@\"; }"
	if !strings.Contains(got, want) {
		t.Errorf("expected %q in hook output\ngot:\n%s", want, got)
	}
}

func TestGenerate_launcherOneLiner(t *testing.T) {
	got := Generate("zsh", ruffReg(), minimalConfig())
	want := "uvx() { prettyout _run uvx \"$@\"; }"
	if !strings.Contains(got, want) {
		t.Errorf("expected %q in hook output\ngot:\n%s", want, got)
	}
}

func TestGenerate_ciModeNever_returnsEmpty(t *testing.T) {
	cfg := minimalConfig()
	cfg.CIMode = "never"
	got := Generate("zsh", ruffReg(), cfg)
	if got != "" {
		t.Errorf("ci_mode=never should return empty string, got %q", got)
	}
}

func TestGenerate_ciModeAuto_noTTYCheckInShell(t *testing.T) {
	cfg := minimalConfig()
	cfg.CIMode = "auto"
	got := Generate("zsh", ruffReg(), cfg)
	if strings.Contains(got, "[[ -t 1 ]]") {
		t.Error("TTY check moved to runner; hook must not contain [[ -t 1 ]]")
	}
}

func TestGenerate_noDuplicateLaunchers(t *testing.T) {
	reg := &registry.Registry{
		Tools: map[string]registry.ToolConfig{
			"ruff":         {Launchers: []string{"uvx"}},
			"basedpyright": {Launchers: []string{"uvx"}},
		},
		Launchers: map[string]registry.LauncherConfig{"uvx": {}},
	}
	got := Generate("zsh", reg, minimalConfig())
	count := strings.Count(got, "uvx() {")
	if count != 1 {
		t.Errorf("uvx() should appear exactly once, got %d times", count)
	}
}
