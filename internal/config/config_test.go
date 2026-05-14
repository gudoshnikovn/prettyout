package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeYAML(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad_defaults(t *testing.T) {
	cfg := loadFile("/nonexistent/path.yaml")
	if cfg.CIMode != "auto" {
		t.Errorf("default CIMode = %q, want auto", cfg.CIMode)
	}
	if cfg.Enabled == nil {
		t.Error("Enabled map must not be nil")
	}
}

func TestLoad_enabled(t *testing.T) {
	dir := t.TempDir()
	path := writeYAML(t, dir, "config.yaml", `
enabled:
  ruff: true
  basedpyright: false
`)
	cfg := loadFile(path)
	if !cfg.Enabled["ruff"] {
		t.Error("ruff should be enabled")
	}
	if cfg.Enabled["basedpyright"] {
		t.Error("basedpyright should be disabled")
	}
}

func TestLoad_customTools(t *testing.T) {
	dir := t.TempDir()
	path := writeYAML(t, dir, "config.yaml", `
custom_tools:
  mycooltool:
    plugin: ~/scripts/mycooltool-fmt
    json_flags: [--json]
`)
	cfg := loadFile(path)
	if _, ok := cfg.CustomTools["mycooltool"]; !ok {
		t.Error("custom tool not loaded")
	}
}

func TestMerge_projectOverridesGlobal(t *testing.T) {
	global := defaults()
	global.Enabled["ruff"] = true
	global.CIMode = "always"

	project := defaults()
	project.Enabled["ruff"] = false
	project.CIMode = "never"

	merged := mergeConfigs(global, project)
	if merged.Enabled["ruff"] {
		t.Error("project should override global: ruff disabled")
	}
	if merged.CIMode != "never" {
		t.Errorf("project CIMode should win, got %q", merged.CIMode)
	}
}

func TestMerge_projectDoesNotClearGlobal(t *testing.T) {
	global := defaults()
	global.Enabled["ruff"] = true
	global.Plugins["ruff"] = "prettyout-ruff"

	project := defaults()

	merged := mergeConfigs(global, project)
	if !merged.Enabled["ruff"] {
		t.Error("global ruff should survive empty project config")
	}
	if merged.Plugins["ruff"] != "prettyout-ruff" {
		t.Error("global plugin should survive empty project config")
	}
}
