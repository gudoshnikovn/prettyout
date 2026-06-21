package doctor

import (
	"os"
	"path/filepath"
	"testing"

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
