package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gudoshnikovn/prettyout/internal/config"
	"github.com/gudoshnikovn/prettyout/internal/registry"
)

func TestCheckHook_found(t *testing.T) {
	f := filepath.Join(t.TempDir(), ".zshrc")
	os.WriteFile(f, []byte(`eval "$(prettyout hook zsh)"`+"\n"), 0644)
	c := checkHook(f)
	if !c.OK {
		t.Errorf("expected hook found, got: %s", c.Message)
	}
}

func TestCheckHook_notFound(t *testing.T) {
	f := filepath.Join(t.TempDir(), ".zshrc")
	os.WriteFile(f, []byte("# empty\n"), 0644)
	c := checkHook(f)
	if c.OK {
		t.Error("expected hook not found")
	}
	if c.Hint == "" {
		t.Error("expected hint for missing hook")
	}
}

func TestCheckHook_fileMissing(t *testing.T) {
	c := checkHook(filepath.Join(t.TempDir(), "nonexistent"))
	if c.OK {
		t.Error("missing rc file should fail hook check")
	}
}

func TestCheckTool_pluginInstalled(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "prettyout-ruff")
	os.WriteFile(fake, []byte("#!/bin/sh\n"), 0755)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	tc := registry.ToolConfig{Plugin: "prettyout-ruff"}
	c := checkTool("ruff", tc, false)
	if !c.OK {
		t.Errorf("expected OK when plugin found and tool disabled: %s", c.Message)
	}
}

func TestCheckTool_pluginMissing(t *testing.T) {
	tc := registry.ToolConfig{Plugin: "prettyout-definitely-not-installed-xyz"}
	c := checkTool("sometool", tc, false)
	if c.OK {
		t.Error("expected not OK when plugin missing")
	}
	if c.Hint == "" {
		t.Error("expected hint for missing plugin")
	}
}

func TestCheckConfigFile_valid(t *testing.T) {
	f := filepath.Join(t.TempDir(), "config.yaml")
	os.WriteFile(f, []byte("ci_mode: auto\n"), 0644)
	c := checkConfigFile(f, "global")
	if !c.OK {
		t.Errorf("expected valid config: %s", c.Message)
	}
}

func TestCheckConfigFile_syntaxError(t *testing.T) {
	f := filepath.Join(t.TempDir(), "config.yaml")
	os.WriteFile(f, []byte("invalid: yaml: {broken\n"), 0644)
	c := checkConfigFile(f, "global")
	if c.OK {
		t.Error("invalid YAML should fail check")
	}
	if c.Hint == "" {
		t.Error("expected hint describing parse error")
	}
}

func TestCheckConfigFile_missingFile(t *testing.T) {
	c := checkConfigFile(filepath.Join(t.TempDir(), "nonexistent.yaml"), "project")
	if !c.OK {
		t.Error("missing optional config file should be OK")
	}
}

func TestDetectShell_fromEnv(t *testing.T) {
	t.Setenv("SHELL", "/usr/bin/zsh")
	if got := detectShell(); got != "zsh" {
		t.Errorf("got %q, want zsh", got)
	}
}

func TestDetectShell_bash(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	if got := detectShell(); got != "bash" {
		t.Errorf("got %q, want bash", got)
	}
}

func TestDetectShell_empty(t *testing.T) {
	t.Setenv("SHELL", "")
	if got := detectShell(); got != "zsh" {
		t.Errorf("empty SHELL: got %q, want zsh", got)
	}
}

func TestRcFilePath_zsh(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	p := rcFilePath("zsh")
	if filepath.Base(p) != ".zshrc" {
		t.Errorf("zsh rc path: got %q, want .zshrc", filepath.Base(p))
	}
}

func TestRcFilePath_bash(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	p := rcFilePath("bash")
	if filepath.Base(p) != ".bashrc" {
		t.Errorf("bash rc path: got %q, want .bashrc", filepath.Base(p))
	}
}

func TestRcFilePath_unknown_defaultsToZsh(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	p := rcFilePath("fish")
	if filepath.Base(p) != ".zshrc" {
		t.Errorf("unknown shell: got %q, want .zshrc", filepath.Base(p))
	}
}

func TestRun_returnsChecks(t *testing.T) {
	dir := t.TempDir()
	// Create a fake plugin so checkTool can find it
	fake := filepath.Join(dir, "prettyout-ruff")
	os.WriteFile(fake, []byte("#!/bin/sh\n"), 0755)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HOME", dir)

	reg := &registry.Registry{
		Tools: map[string]registry.ToolConfig{
			"ruff": {Plugin: "prettyout-ruff"},
		},
		Launchers: map[string]registry.LauncherConfig{},
	}
	from_config := &config.Config{
		Enabled:     map[string]bool{"ruff": true},
		Plugins:     map[string]string{},
		Settings:    map[string]config.FormatterSettings{},
		CustomTools: map[string]registry.ToolConfig{},
		CIMode:      "always",
	}

	checks := Run(reg, from_config)
	if len(checks) == 0 {
		t.Fatal("Run should return at least one check")
	}
	// Should include a hook check, tool checks, and config checks
	names := make(map[string]bool, len(checks))
	for _, c := range checks {
		names[c.Name] = true
	}
	if !names["tool-ruff"] {
		t.Error("expected tool-ruff check")
	}
	if !names["config-global"] {
		t.Error("expected config-global check")
	}
	if !names["config-project"] {
		t.Error("expected config-project check")
	}
}
