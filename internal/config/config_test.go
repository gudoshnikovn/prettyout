package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gudoshnikovn/prettyout/internal/registry"
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
    output_args: [--json]
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

func TestLoad_mergesGlobalAndProject(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Write global config
	globalDir := filepath.Join(home, ".config", "prettyout")
	os.MkdirAll(globalDir, 0755)
	writeYAML(t, globalDir, "config.yaml", `
enabled:
  ruff: true
ci_mode: always
`)

	// Write project config in a temp CWD
	projectDir := t.TempDir()
	writeYAML(t, projectDir, ".prettyout.yaml", `
enabled:
  mypy: true
`)

	// Change working directory for the duration of the test
	origDir, _ := os.Getwd()
	os.Chdir(projectDir)
	defer os.Chdir(origDir)

	cfg := Load()
	if !cfg.Enabled["ruff"] {
		t.Error("global ruff should be enabled")
	}
	if !cfg.Enabled["mypy"] {
		t.Error("project mypy should be enabled")
	}
	if cfg.CIMode != "always" {
		t.Errorf("global ci_mode should be preserved, got %q", cfg.CIMode)
	}
}

func TestSave_writesFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := defaults()
	cfg.Enabled["ruff"] = true
	cfg.CIMode = "always"
	Save(cfg)

	path := filepath.Join(home, ".config", "prettyout", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("config file not written: %v", err)
	}
	if string(data) == "" {
		t.Error("config file is empty")
	}

	// Verify round-trip
	loaded := loadFile(path)
	if !loaded.Enabled["ruff"] {
		t.Error("saved ruff enabled not preserved")
	}
	if loaded.CIMode != "always" {
		t.Errorf("saved ci_mode not preserved, got %q", loaded.CIMode)
	}
}

func TestGlobalConfigPath_containsPrettyout(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	p := GlobalConfigPath()
	if !filepath.IsAbs(p) {
		t.Errorf("path should be absolute, got %q", p)
	}
	if filepath.Base(p) != "config.yaml" {
		t.Errorf("expected config.yaml, got %q", filepath.Base(p))
	}
}

func TestMerge_ciModeAutoNotOverridden(t *testing.T) {
	global := defaults()
	global.CIMode = "always"

	project := defaults() // CIMode = "auto" (default)

	merged := mergeConfigs(global, project)
	// project CIMode is "auto" which is the default, so global should win
	if merged.CIMode != "always" {
		t.Errorf("global ci_mode=always should win when project is default auto, got %q", merged.CIMode)
	}
}

func TestLoadFile_invalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := writeYAML(t, dir, "config.yaml", "enabled: {[invalid yaml")
	// yaml.Unmarshal fails → returns defaults
	cfg := loadFile(path)
	if cfg.CIMode != "auto" {
		t.Errorf("invalid YAML: want default CIMode 'auto', got %q", cfg.CIMode)
	}
}

func TestLoadFile_nullFields(t *testing.T) {
	dir := t.TempDir()
	// YAML with explicit null values nullifies the map fields after unmarshal
	path := writeYAML(t, dir, "config.yaml", `
enabled: ~
plugins: ~
settings: ~
custom_tools: ~
ci_mode: ""
`)
	cfg := loadFile(path)
	// nil checks in loadFile should re-initialize all maps
	if cfg.Enabled == nil {
		t.Error("Enabled should not be nil after loadFile nil-check")
	}
	if cfg.Plugins == nil {
		t.Error("Plugins should not be nil after loadFile nil-check")
	}
	if cfg.Settings == nil {
		t.Error("Settings should not be nil after loadFile nil-check")
	}
	if cfg.CustomTools == nil {
		t.Error("CustomTools should not be nil after loadFile nil-check")
	}
	if cfg.CIMode != "auto" {
		t.Errorf("empty ci_mode should default to 'auto', got %q", cfg.CIMode)
	}
}

func TestMergeConfigs_withSettings(t *testing.T) {
	global := defaults()
	global.Settings["ruff"] = FormatterSettings{GroupBy: "file"}

	project := defaults()
	project.Settings["mypy"] = FormatterSettings{GroupBy: "rule"}

	merged := mergeConfigs(global, project)
	if merged.Settings["ruff"].GroupBy != "file" {
		t.Error("global setting for ruff should survive merge")
	}
	if merged.Settings["mypy"].GroupBy != "rule" {
		t.Error("project setting for mypy should be merged in")
	}
}

func TestMergeConfigs_withCustomTools(t *testing.T) {
	global := defaults()
	global.CustomTools["mytool"] = registry.ToolConfig{Plugin: "~/bin/mytool-fmt"}

	project := defaults()
	project.CustomTools["othertool"] = registry.ToolConfig{Plugin: "~/bin/other-fmt"}

	merged := mergeConfigs(global, project)
	if merged.CustomTools["mytool"].Plugin != "~/bin/mytool-fmt" {
		t.Error("global custom tool should survive merge")
	}
	if merged.CustomTools["othertool"].Plugin != "~/bin/other-fmt" {
		t.Error("project custom tool should be merged in")
	}
}

func TestCopySettingsMap_nonEmpty(t *testing.T) {
	m := map[string]FormatterSettings{
		"ruff": {GroupBy: "file"},
		"mypy": {GroupBy: "rule"},
	}
	out := copySettingsMap(m)
	if len(out) != 2 {
		t.Errorf("copySettingsMap: got %d entries, want 2", len(out))
	}
	if out["ruff"].GroupBy != "file" {
		t.Error("copySettingsMap: ruff.GroupBy should be 'file'")
	}
}

func TestCopyToolMap_nonEmpty(t *testing.T) {
	m := map[string]registry.ToolConfig{
		"mytool": {Plugin: "~/bin/mytool"},
	}
	out := copyToolMap(m)
	if len(out) != 1 {
		t.Errorf("copyToolMap: got %d entries, want 1", len(out))
	}
	if out["mytool"].Plugin != "~/bin/mytool" {
		t.Error("copyToolMap: mytool.Plugin not preserved")
	}
}
