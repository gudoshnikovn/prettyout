# prettyout Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite prettyout with a registry-driven, launcher-aware, plugin-extensible architecture that fixes the ruff-format breakage, adds uvx/pipx support, separates stderr, and enables formatter configuration.

**Architecture:** A YAML registry embedded in the binary describes every supported tool (which subcommands to intercept, JSON flags, passthrough flags, launchers). A separate user config enables/disables tools, overrides formatters, and holds per-tool visual settings. Hook generation reads both and emits subcommand-aware shell functions with launcher wrappers.

**Tech Stack:** Go 1.25, gopkg.in/yaml.v3, go:embed, zsh/bash

---

## File Map

| File | Action | Purpose |
|---|---|---|
| `internal/registry/builtin.yaml` | CREATE | Embedded tool + launcher definitions |
| `internal/registry/registry.go` | REWRITE | Load builtin YAML, ToolConfig/LauncherConfig types, Merge/ShouldIntercept/HasPassthroughFlag |
| `internal/registry/registry_test.go` | CREATE | Tests for registry loading and helpers |
| `internal/launcher/launcher.go` | CREATE | ExtractTool: parse tool name + subcommand from launcher args |
| `internal/launcher/launcher_test.go` | CREATE | Tests for uvx/pipx/npx arg parsing |
| `internal/config/config.go` | REWRITE | New schema: Enabled, Plugins, Settings, CustomTools, CIMode; global+project merge |
| `internal/config/config_test.go` | CREATE | Tests for load and merge |
| `pkg/formatter/config.go` | CREATE | Config struct, LoadConfig(toolName) reading config files directly |
| `pkg/formatter/formatter.go` | MODIFY | Add RunWithConfig |
| `pkg/formatter/config_test.go` | CREATE | Tests for config loading and defaults |
| `internal/hook/hook.go` | REWRITE | Generate subcommand-aware + launcher-aware shell functions |
| `internal/hook/hook_test.go` | REWRITE | Tests for generated hook content |
| `cmd/prettyout-ruff/main.go` | MODIFY | Use RunWithConfig, remove backend/ hardcode |
| `cmd/prettyout-basedpyright/main.go` | MODIFY | Use RunWithConfig, remove backend/ hardcode |
| `cmd/prettyout/main.go` | MODIFY | Wire new registry+config, rename Hooks→Enabled |

---

## Task 1: Registry — builtin YAML + Go types

**Files:**
- Create: `internal/registry/builtin.yaml`
- Create: `internal/registry/registry.go`
- Create: `internal/registry/registry_test.go`
- Delete content of: `internal/registry/registry.go` (full rewrite)

- [ ] **Step 1: Write the failing tests**

```go
// internal/registry/registry_test.go
package registry

import "testing"

func TestBuiltinLoads(t *testing.T) {
	r, err := LoadBuiltin()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.Tools["ruff"]; !ok {
		t.Error("ruff not in built-in registry")
	}
	if _, ok := r.Launchers["uvx"]; !ok {
		t.Error("uvx not in built-in launchers")
	}
}

func TestShouldIntercept_withSubcommands(t *testing.T) {
	tc := ToolConfig{InterceptSubcommands: []string{"check"}}
	if !tc.ShouldIntercept("check") {
		t.Error("should intercept 'check'")
	}
	if tc.ShouldIntercept("format") {
		t.Error("should not intercept 'format'")
	}
}

func TestShouldIntercept_noRestriction(t *testing.T) {
	tc := ToolConfig{}
	if !tc.ShouldIntercept("anything") {
		t.Error("empty intercept_subcommands means intercept all")
	}
}

func TestHasPassthroughFlag(t *testing.T) {
	tc := ToolConfig{PassthroughFlags: []string{"--watch", "-W"}}
	if !tc.HasPassthroughFlag([]string{"check", "--watch", "."}) {
		t.Error("should detect --watch")
	}
	if tc.HasPassthroughFlag([]string{"check", "."}) {
		t.Error("no passthrough flag present")
	}
}

func TestMerge_addsCustomTool(t *testing.T) {
	r, _ := LoadBuiltin()
	r.Merge(map[string]ToolConfig{
		"mycooltool": {Plugin: "prettyout-mycooltool"},
	})
	if _, ok := r.Tools["mycooltool"]; !ok {
		t.Error("custom tool not merged")
	}
}

func TestMerge_overridesBuiltin(t *testing.T) {
	r, _ := LoadBuiltin()
	r.Merge(map[string]ToolConfig{
		"ruff": {Plugin: "my-ruff-override"},
	})
	if r.Tools["ruff"].Plugin != "my-ruff-override" {
		t.Error("custom tool should override builtin")
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd /Users/gudoshnikov_na/Programming/Agents/prettyout
go test ./internal/registry/... -v
```

Expected: compile error (types don't exist yet).

- [ ] **Step 3: Create `internal/registry/builtin.yaml`**

```yaml
launchers:
  uvx:
    skip_flags: [--no-cache, --no-project, --isolated, -q, --quiet]
    value_flags: [--python, -p, --with, --from]
    tool_position: first_non_flag
  pipx:
    value_flags: [--python, --spec, --index-url]
    prefix_args: [run]
    tool_position: first_non_flag
  npx:
    skip_flags: [--yes, -y, --no, --legacy-peer-deps]
    value_flags: [-p, --package]
    tool_position: first_non_flag

tools:
  ruff:
    plugin: prettyout-ruff
    intercept_subcommands: [check]
    json_flags: [--output-format=json]
    passthrough_flags: [--watch, -W]
    launchers: [uvx, pipx]
  basedpyright:
    plugin: prettyout-basedpyright
    json_flags: [--outputjson]
    launchers: [uvx]
```

- [ ] **Step 4: Rewrite `internal/registry/registry.go`**

```go
package registry

import (
	_ "embed"
	"sort"

	"gopkg.in/yaml.v3"
)

//go:embed builtin.yaml
var builtinYAML []byte

type LauncherConfig struct {
	SkipFlags    []string `yaml:"skip_flags"`
	ValueFlags   []string `yaml:"value_flags"`
	PrefixArgs   []string `yaml:"prefix_args"`
	ToolPosition string   `yaml:"tool_position"`
}

type ToolConfig struct {
	Plugin               string   `yaml:"plugin"`
	InterceptSubcommands []string `yaml:"intercept_subcommands"`
	JSONFlags            []string `yaml:"json_flags"`
	PassthroughFlags     []string `yaml:"passthrough_flags"`
	Launchers            []string `yaml:"launchers"`
}

type Registry struct {
	Launchers map[string]LauncherConfig `yaml:"launchers"`
	Tools     map[string]ToolConfig     `yaml:"tools"`
}

func LoadBuiltin() (*Registry, error) {
	var r Registry
	if err := yaml.Unmarshal(builtinYAML, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Merge adds custom tools from user config. Custom entries override builtins.
func (r *Registry) Merge(custom map[string]ToolConfig) {
	for name, cfg := range custom {
		r.Tools[name] = cfg
	}
}

// SortedToolNames returns tool names in alphabetical order.
func (r *Registry) SortedToolNames() []string {
	names := make([]string, 0, len(r.Tools))
	for name := range r.Tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ShouldIntercept returns true if subcommand should be intercepted.
// An empty InterceptSubcommands list means intercept all subcommands.
func (tc ToolConfig) ShouldIntercept(subcommand string) bool {
	if len(tc.InterceptSubcommands) == 0 {
		return true
	}
	for _, s := range tc.InterceptSubcommands {
		if s == subcommand {
			return true
		}
	}
	return false
}

// HasPassthroughFlag returns true if any arg matches a passthrough flag.
func (tc ToolConfig) HasPassthroughFlag(args []string) bool {
	for _, arg := range args {
		for _, pf := range tc.PassthroughFlags {
			if arg == pf {
				return true
			}
		}
	}
	return false
}
```

- [ ] **Step 5: Run tests — expect pass**

```bash
go test ./internal/registry/... -v
```

Expected: all 6 tests PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/registry/
git commit -m "feat: registry loaded from embedded YAML with subcommand/passthrough support"
```

---

## Task 2: Launcher arg parsing

**Files:**
- Create: `internal/launcher/launcher.go`
- Create: `internal/launcher/launcher_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// internal/launcher/launcher_test.go
package launcher

import (
	"testing"

	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

func uvxCfg() registry.LauncherConfig {
	return registry.LauncherConfig{
		SkipFlags:    []string{"--no-cache", "--no-project", "--isolated", "-q", "--quiet"},
		ValueFlags:   []string{"--python", "-p", "--with", "--from"},
		ToolPosition: "first_non_flag",
	}
}

func pipxCfg() registry.LauncherConfig {
	return registry.LauncherConfig{
		ValueFlags:   []string{"--python", "--spec", "--index-url"},
		PrefixArgs:   []string{"run"},
		ToolPosition: "first_non_flag",
	}
}

func TestExtractTool_simple(t *testing.T) {
	tool, sub := ExtractTool(uvxCfg(), []string{"ruff", "check", "."})
	if tool != "ruff" || sub != "check" {
		t.Errorf("got tool=%q sub=%q, want ruff/check", tool, sub)
	}
}

func TestExtractTool_versionSuffix(t *testing.T) {
	tool, sub := ExtractTool(uvxCfg(), []string{"ruff@0.5.0", "check", "."})
	if tool != "ruff" || sub != "check" {
		t.Errorf("got tool=%q sub=%q, want ruff/check", tool, sub)
	}
}

func TestExtractTool_withSkipFlag(t *testing.T) {
	tool, sub := ExtractTool(uvxCfg(), []string{"--no-cache", "ruff", "check", "."})
	if tool != "ruff" || sub != "check" {
		t.Errorf("got tool=%q sub=%q, want ruff/check", tool, sub)
	}
}

func TestExtractTool_withValueFlag(t *testing.T) {
	// --python consumes the next arg "3.11"
	tool, sub := ExtractTool(uvxCfg(), []string{"--python", "3.11", "ruff", "check", "."})
	if tool != "ruff" || sub != "check" {
		t.Errorf("got tool=%q sub=%q, want ruff/check", tool, sub)
	}
}

func TestExtractTool_pipxRun(t *testing.T) {
	tool, sub := ExtractTool(pipxCfg(), []string{"run", "ruff", "check", "."})
	if tool != "ruff" || sub != "check" {
		t.Errorf("got tool=%q sub=%q, want ruff/check", tool, sub)
	}
}

func TestExtractTool_noSubcommand(t *testing.T) {
	tool, sub := ExtractTool(uvxCfg(), []string{"basedpyright", "."})
	if tool != "basedpyright" {
		t.Errorf("got tool=%q, want basedpyright", tool)
	}
	// "." is returned as the first positional after tool name
	if sub != "." {
		t.Errorf("got sub=%q, want '.'", sub)
	}
}

func TestExtractTool_empty(t *testing.T) {
	tool, sub := ExtractTool(uvxCfg(), []string{})
	if tool != "" || sub != "" {
		t.Errorf("empty args should return empty strings")
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./internal/launcher/... -v
```

Expected: compile error (package doesn't exist).

- [ ] **Step 3: Create `internal/launcher/launcher.go`**

```go
package launcher

import (
	"strings"

	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

// ExtractTool parses launcher arguments to find the tool name and first
// positional argument after it (subcommand). Returns ("", "") if no tool found.
func ExtractTool(lc registry.LauncherConfig, args []string) (toolName, subcommand string) {
	remaining := args

	// consume prefix_args (e.g. "run" for pipx)
	for _, prefix := range lc.PrefixArgs {
		if len(remaining) > 0 && remaining[0] == prefix {
			remaining = remaining[1:]
		}
	}

	skipFlags := make(map[string]bool, len(lc.SkipFlags))
	for _, f := range lc.SkipFlags {
		skipFlags[f] = true
	}
	valueFlags := make(map[string]bool, len(lc.ValueFlags))
	for _, f := range lc.ValueFlags {
		valueFlags[f] = true
	}

	toolFound := false
	skipNext := false
	for _, arg := range remaining {
		if skipNext {
			skipNext = false
			continue
		}
		if valueFlags[arg] {
			skipNext = true
			continue
		}
		if skipFlags[arg] {
			continue
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if !toolFound {
			toolName = strings.SplitN(arg, "@", 2)[0]
			toolFound = true
			continue
		}
		subcommand = arg
		return
	}
	return
}
```

- [ ] **Step 4: Run tests — expect pass**

```bash
go test ./internal/launcher/... -v
```

Expected: all 7 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/launcher/
git commit -m "feat: launcher arg parsing for uvx/pipx/npx with version-suffix and value-flag support"
```

---

## Task 3: Config — new schema with global + project merge

**Files:**
- Rewrite: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// internal/config/config_test.go
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
	// no files exist → defaults
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

	project := defaults() // empty project config

	merged := mergeConfigs(global, project)
	if !merged.Enabled["ruff"] {
		t.Error("global ruff should survive empty project config")
	}
	if merged.Plugins["ruff"] != "prettyout-ruff" {
		t.Error("global plugin should survive empty project config")
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./internal/config/... -v
```

Expected: compile error (new functions not defined).

- [ ] **Step 3: Rewrite `internal/config/config.go`**

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

type FormatterSettings struct {
	GroupBy          string         `yaml:"group_by"`
	Colors           *bool          `yaml:"colors"`
	MaxMessageLength *int           `yaml:"max_message_length"`
	Extra            map[string]any `yaml:"extra"`
}

type Config struct {
	Enabled     map[string]bool                    `yaml:"enabled"`
	Plugins     map[string]string                  `yaml:"plugins"`
	Settings    map[string]FormatterSettings       `yaml:"settings"`
	CustomTools map[string]registry.ToolConfig     `yaml:"custom_tools"`
	CIMode      string                             `yaml:"ci_mode"`
}

func defaults() *Config {
	return &Config{
		Enabled:     map[string]bool{},
		Plugins:     map[string]string{},
		Settings:    map[string]FormatterSettings{},
		CustomTools: map[string]registry.ToolConfig{},
		CIMode:      "auto",
	}
}

// Load returns the merged result of the global config and the per-project config.
// Per-project values take precedence over global values.
func Load() *Config {
	global := loadFile(globalConfigPath())
	project := loadFile(projectConfigPath())
	return mergeConfigs(global, project)
}

func loadFile(path string) *Config {
	cfg := defaults()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg
	}
	// ensure maps are never nil after unmarshal
	if cfg.Enabled == nil {
		cfg.Enabled = map[string]bool{}
	}
	if cfg.Plugins == nil {
		cfg.Plugins = map[string]string{}
	}
	if cfg.Settings == nil {
		cfg.Settings = map[string]FormatterSettings{}
	}
	if cfg.CustomTools == nil {
		cfg.CustomTools = map[string]registry.ToolConfig{}
	}
	if cfg.CIMode == "" {
		cfg.CIMode = "auto"
	}
	return cfg
}

// mergeConfigs returns a config where project values override global values key-by-key.
func mergeConfigs(global, project *Config) *Config {
	result := *global
	result.Enabled = copyBoolMap(global.Enabled)
	result.Plugins = copyStringMap(global.Plugins)
	result.Settings = copySettingsMap(global.Settings)
	result.CustomTools = copyToolMap(global.CustomTools)

	for k, v := range project.Enabled {
		result.Enabled[k] = v
	}
	for k, v := range project.Plugins {
		result.Plugins[k] = v
	}
	for k, v := range project.Settings {
		result.Settings[k] = v
	}
	for k, v := range project.CustomTools {
		result.CustomTools[k] = v
	}
	if project.CIMode != "" && project.CIMode != "auto" {
		result.CIMode = project.CIMode
	}
	return &result
}

func Save(cfg *Config) {
	path := globalConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "prettyout: cannot create config dir: %v\n", err)
		os.Exit(1)
	}
	data, _ := yaml.Marshal(cfg)
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "prettyout: cannot write config: %v\n", err)
		os.Exit(1)
	}
}

func globalConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "prettyout", "config.yaml")
}

func projectConfigPath() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".prettyout.yaml")
}

func copyBoolMap(m map[string]bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func copyStringMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func copySettingsMap(m map[string]FormatterSettings) map[string]FormatterSettings {
	out := make(map[string]FormatterSettings, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func copyToolMap(m map[string]registry.ToolConfig) map[string]registry.ToolConfig {
	out := make(map[string]registry.ToolConfig, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
```

- [ ] **Step 4: Run tests — expect pass**

```bash
go test ./internal/config/... -v
```

Expected: all 5 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: config schema with enabled/plugins/settings/custom_tools/ci_mode and project-level override"
```

---

## Task 4: pkg/formatter — Config + RunWithConfig

**Files:**
- Create: `pkg/formatter/config.go`
- Create: `pkg/formatter/config_test.go`
- Modify: `pkg/formatter/formatter.go`

- [ ] **Step 1: Write the failing tests**

```go
// pkg/formatter/config_test.go
package formatter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.GroupBy != "rule" {
		t.Errorf("default GroupBy = %q, want rule", cfg.GroupBy)
	}
	if !cfg.Colors {
		t.Error("default Colors should be true")
	}
}

func TestLoadConfig_usesDefaults_whenNoFile(t *testing.T) {
	// Point to a nonexistent home to simulate no config
	t.Setenv("HOME", t.TempDir())
	cfg := LoadConfig("ruff")
	if cfg.GroupBy != "rule" {
		t.Errorf("want default GroupBy=rule, got %q", cfg.GroupBy)
	}
}

func TestLoadConfig_readsGroupBy(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "prettyout")
	_ = os.MkdirAll(dir, 0755)
	_ = os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`
settings:
  ruff:
    group_by: file
    max_message_length: 80
`), 0644)

	cfg := LoadConfig("ruff")
	if cfg.GroupBy != "file" {
		t.Errorf("want group_by=file, got %q", cfg.GroupBy)
	}
	if cfg.MaxMessageLength != 80 {
		t.Errorf("want max_message_length=80, got %d", cfg.MaxMessageLength)
	}
}

func TestLoadConfig_colorsExplicitFalse(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "prettyout")
	_ = os.MkdirAll(dir, 0755)
	_ = os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`
settings:
  ruff:
    colors: false
`), 0644)

	cfg := LoadConfig("ruff")
	if cfg.Colors {
		t.Error("colors should be false when explicitly set to false")
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./pkg/formatter/... -v
```

Expected: compile errors (Config, LoadConfig, DefaultConfig not defined).

- [ ] **Step 3: Create `pkg/formatter/config.go`**

```go
package formatter

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds resolved formatter settings for a single tool.
type Config struct {
	GroupBy          string
	Colors           bool
	MaxMessageLength int
	Extra            map[string]any
}

func DefaultConfig() Config {
	return Config{
		GroupBy: "rule",
		Colors:  true,
	}
}

// LoadConfig reads formatter settings for toolName from the global config
// (~/.config/prettyout/config.yaml) and the per-project config (.prettyout.yaml),
// with per-project values taking precedence.
func LoadConfig(toolName string) Config {
	cfg := DefaultConfig()
	for _, path := range configPaths() {
		applyFile(path, toolName, &cfg)
	}
	return cfg
}

type rawSettings struct {
	GroupBy          string         `yaml:"group_by"`
	Colors           *bool          `yaml:"colors"`
	MaxMessageLength *int           `yaml:"max_message_length"`
	Extra            map[string]any `yaml:"extra"`
}

type rawConfigFile struct {
	Settings map[string]rawSettings `yaml:"settings"`
}

func applyFile(path, toolName string, cfg *Config) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var f rawConfigFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return
	}
	s, ok := f.Settings[toolName]
	if !ok {
		return
	}
	if s.GroupBy != "" {
		cfg.GroupBy = s.GroupBy
	}
	if s.Colors != nil {
		cfg.Colors = *s.Colors
	}
	if s.MaxMessageLength != nil {
		cfg.MaxMessageLength = *s.MaxMessageLength
	}
	for k, v := range s.Extra {
		if cfg.Extra == nil {
			cfg.Extra = map[string]any{}
		}
		cfg.Extra[k] = v
	}
}

func configPaths() []string {
	home, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()
	return []string{
		filepath.Join(home, ".config", "prettyout", "config.yaml"),
		filepath.Join(cwd, ".prettyout.yaml"),
	}
}
```

- [ ] **Step 4: Add `RunWithConfig` to `pkg/formatter/formatter.go`**

```go
package formatter

import (
	"fmt"
	"io"
	"os"
)

// Run reads all of stdin and passes it to transform.
func Run(transform func([]byte) error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "prettyout: read error: %v\n", err)
		os.Exit(1)
	}
	if err := transform(data); err != nil {
		fmt.Fprintf(os.Stderr, "prettyout: %v\n", err)
		os.Exit(1)
	}
}

// RunWithConfig reads all of stdin and passes it along with loaded formatter
// settings to transform. toolName must match the key in prettyout config settings.
func RunWithConfig(toolName string, transform func([]byte, Config) error) {
	cfg := LoadConfig(toolName)
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "prettyout: read error: %v\n", err)
		os.Exit(1)
	}
	if err := transform(data, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "prettyout: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 5: Run tests — expect pass**

```bash
go test ./pkg/formatter/... -v
```

Expected: all 4 tests PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/formatter/
git commit -m "feat: formatter Config with LoadConfig and RunWithConfig"
```

---

## Task 5: Hook — subcommand-aware direct wrappers

**Files:**
- Rewrite: `internal/hook/hook.go`
- Rewrite: `internal/hook/hook_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// internal/hook/hook_test.go
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
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./internal/hook/... -v
```

Expected: compile error (Generate signature changed).

- [ ] **Step 3: Rewrite `internal/hook/hook.go`**

```go
package hook

import (
	"fmt"
	"strings"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

// Generate returns shell code to be eval'd by the user's shell.
func Generate(shell string, reg *registry.Registry, cfg *config.Config) string {
	var b strings.Builder

	for _, name := range reg.SortedToolNames() {
		tc := reg.Tools[name]
		plugin := resolvePlugin(name, tc, cfg)
		writeTool(&b, shell, name, tc, plugin, cfg.CIMode)
	}

	// collect tools per launcher
	byLauncher := map[string][]string{}
	for _, name := range reg.SortedToolNames() {
		tc := reg.Tools[name]
		for _, l := range tc.Launchers {
			if _, ok := reg.Launchers[l]; ok {
				byLauncher[l] = append(byLauncher[l], name)
			}
		}
	}
	for launcherName, tools := range byLauncher {
		lc := reg.Launchers[launcherName]
		writeLauncher(&b, shell, launcherName, lc, tools, reg, cfg)
	}

	return b.String()
}

func resolvePlugin(toolName string, tc registry.ToolConfig, cfg *config.Config) string {
	if override, ok := cfg.Plugins[toolName]; ok && override != "" {
		return override
	}
	return tc.Plugin
}

func writeTool(b *strings.Builder, shell, name string, tc registry.ToolConfig, plugin, ciMode string) {
	flags := strings.Join(tc.JSONFlags, " ")

	fmt.Fprintf(b, "\n%s() {\n", name)

	if ciMode == "auto" {
		fmt.Fprintf(b, "  [[ -t 1 ]] || { command %s \"$@\"; return $?; }\n", name)
	}

	fmt.Fprintf(b, "  if prettyout _enabled %s 2>/dev/null; then\n", name)

	if len(tc.PassthroughFlags) > 0 {
		fmt.Fprintf(b, "    for _ptf in \"$@\"; do\n")
		fmt.Fprintf(b, "      case \"$_ptf\" in\n")
		fmt.Fprintf(b, "        %s) command %s \"$@\"; return $? ;;\n",
			strings.Join(tc.PassthroughFlags, "|"), name)
		fmt.Fprintf(b, "      esac\n")
		fmt.Fprintf(b, "    done\n")
	}

	if len(tc.InterceptSubcommands) == 0 {
		// intercept all subcommands — no case statement, no ;;
		writeInterceptBlock(b, shell, name, flags, plugin, "    ")
	} else {
		fmt.Fprintf(b, "    case \"${1:-}\" in\n")
		fmt.Fprintf(b, "      %s)\n", strings.Join(tc.InterceptSubcommands, "|"))
		writeInterceptBlock(b, shell, name, flags, plugin, "        ")
		fmt.Fprintf(b, "        ;;\n") // terminates the case arm
		fmt.Fprintf(b, "    esac\n")
	}

	fmt.Fprintf(b, "  fi\n")
	fmt.Fprintf(b, "  command %s \"$@\"\n", name)
	fmt.Fprintf(b, "}\n")
}

func writeInterceptBlock(b *strings.Builder, shell, toolName, flags, plugin, indent string) {
	fmt.Fprintf(b, "%slocal _ef; _ef=$(mktemp)\n", indent)
	fmt.Fprintf(b, "%scommand %s \"$@\" %s 2>\"$_ef\" | %s\n", indent, toolName, flags, plugin)
	switch shell {
	case "zsh":
		fmt.Fprintf(b, "%slocal _r=${pipestatus[1]}; cat \"$_ef\" >&2; rm -f \"$_ef\"; return $_r\n", indent)
	case "bash":
		fmt.Fprintf(b, "%slocal _r=${PIPESTATUS[0]}; cat \"$_ef\" >&2; rm -f \"$_ef\"; return $_r\n", indent)
	}
}

// writeLauncher generates a wrapper function for launchers like uvx, npx, pipx.
// It is a stub here — launcher wrapper generation is in Task 6.
func writeLauncher(b *strings.Builder, shell, launcherName string, lc registry.LauncherConfig, tools []string, reg *registry.Registry, cfg *config.Config) {
	// implemented in Task 6
}
```

- [ ] **Step 4: Run tests — expect pass (launcher tests will be pending until Task 6)**

```bash
go test ./internal/hook/... -v -run "TestGenerate_contains|TestGenerate_subcommand|TestGenerate_passthrough|TestGenerate_ciMode|TestGenerate_plugin|TestGenerate_stderr"
```

Expected: all 6 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/hook/
git commit -m "feat: hook generates subcommand-aware, stderr-separated, ci_mode-aware shell functions"
```

---

## Task 6: Hook — launcher wrappers (uvx, pipx, npx)

**Files:**
- Modify: `internal/hook/hook.go` (implement `writeLauncher`)
- Modify: `internal/hook/hook_test.go` (add launcher tests)

- [ ] **Step 1: Add launcher tests to `internal/hook/hook_test.go`**

Add these test functions to the existing file:

```go
func TestGenerate_launcherWrapper(t *testing.T) {
	reg := &registry.Registry{
		Tools: map[string]registry.ToolConfig{
			"ruff": {
				Plugin:               "prettyout-ruff",
				JSONFlags:            []string{"--output-format=json"},
				InterceptSubcommands: []string{"check"},
				PassthroughFlags:     []string{"--watch"},
				Launchers:            []string{"uvx"},
			},
		},
		Launchers: map[string]registry.LauncherConfig{
			"uvx": {
				SkipFlags:  []string{"--no-cache"},
				ValueFlags: []string{"--python"},
			},
		},
	}
	got := Generate("zsh", reg, minimalConfig())
	if !strings.Contains(got, "uvx()") {
		t.Error("expected uvx() wrapper function")
	}
	if !strings.Contains(got, "prettyout _enabled ruff") {
		t.Error("expected ruff enabled check inside uvx wrapper")
	}
	if !strings.Contains(got, "--output-format=json") {
		t.Error("expected ruff JSON flag inside uvx wrapper")
	}
	if !strings.Contains(got, "command uvx") {
		t.Error("expected fallback 'command uvx'")
	}
}

func TestGenerate_launcherVersionStrip(t *testing.T) {
	reg := &registry.Registry{
		Tools: map[string]registry.ToolConfig{
			"ruff": {
				Plugin:    "prettyout-ruff",
				JSONFlags: []string{"--output-format=json"},
				Launchers: []string{"uvx"},
			},
		},
		Launchers: map[string]registry.LauncherConfig{
			"uvx": {},
		},
	}
	got := Generate("zsh", reg, minimalConfig())
	// Generated hook must strip @version from tool arg
	if !strings.Contains(got, `%%@*`) {
		t.Error("expected @version stripping (%%@*) in launcher wrapper")
	}
}
```

- [ ] **Step 2: Run new tests to confirm they fail**

```bash
go test ./internal/hook/... -v -run "TestGenerate_launcher"
```

Expected: FAIL — `writeLauncher` is a stub.

- [ ] **Step 3: Implement `writeLauncher` in `internal/hook/hook.go`**

Replace the stub `writeLauncher` function with:

```go
func writeLauncher(b *strings.Builder, shell, launcherName string, lc registry.LauncherConfig, tools []string, reg *registry.Registry, cfg *config.Config) {
	fmt.Fprintf(b, "\n%s() {\n", launcherName)
	fmt.Fprintf(b, "  # parse tool name from launcher args\n")
	fmt.Fprintf(b, "  local _tool_arg=\"\" _skip_next=0\n")

	// generate skip logic for prefix_args
	for _, pa := range lc.PrefixArgs {
		fmt.Fprintf(b, "  # skip prefix arg %q\n", pa)
	}

	// generate the arg-scanning loop to find tool name
	fmt.Fprintf(b, "  local _args=(\"$@\")\n")
	fmt.Fprintf(b, "  for _a in \"${_args[@]}\"; do\n")
	fmt.Fprintf(b, "    if (( _skip_next )); then _skip_next=0; continue; fi\n")

	if len(lc.ValueFlags) > 0 {
		fmt.Fprintf(b, "    case \"$_a\" in\n")
		fmt.Fprintf(b, "      %s) _skip_next=1; continue ;;\n", strings.Join(lc.ValueFlags, "|"))
		fmt.Fprintf(b, "    esac\n")
	}

	if len(lc.SkipFlags) > 0 {
		fmt.Fprintf(b, "    case \"$_a\" in\n")
		fmt.Fprintf(b, "      %s) continue ;;\n", strings.Join(lc.SkipFlags, "|"))
		fmt.Fprintf(b, "    esac\n")
	}

	// skip prefix_args (e.g. "run" for pipx)
	if len(lc.PrefixArgs) > 0 {
		fmt.Fprintf(b, "    case \"$_a\" in\n")
		fmt.Fprintf(b, "      %s) continue ;;\n", strings.Join(lc.PrefixArgs, "|"))
		fmt.Fprintf(b, "    esac\n")
	}

	fmt.Fprintf(b, "    [[ \"$_a\" == -* ]] && continue\n")
	fmt.Fprintf(b, "    _tool_arg=\"$_a\"; break\n")
	fmt.Fprintf(b, "  done\n")
	fmt.Fprintf(b, "  local _toolname=\"${_tool_arg%%%%@*}\"\n") // strip @version

	fmt.Fprintf(b, "  case \"$_toolname\" in\n")

	for _, toolName := range tools {
		tc := reg.Tools[toolName]
		plugin := resolvePlugin(toolName, tc, cfg)
		flags := strings.Join(tc.JSONFlags, " ")

		fmt.Fprintf(b, "    %s)\n", toolName)
		fmt.Fprintf(b, "      if prettyout _enabled %s 2>/dev/null; then\n", toolName)

		if len(tc.PassthroughFlags) > 0 {
			fmt.Fprintf(b, "        for _ptf in \"$@\"; do\n")
			fmt.Fprintf(b, "          case \"$_ptf\" in\n")
			fmt.Fprintf(b, "            %s) command %s \"$@\"; return $? ;;\n",
				strings.Join(tc.PassthroughFlags, "|"), launcherName)
			fmt.Fprintf(b, "          esac\n")
			fmt.Fprintf(b, "        done\n")
		}

		if len(tc.InterceptSubcommands) == 0 {
			writeLauncherInterceptBlock(b, shell, launcherName, toolName, flags, plugin, "        ")
		} else {
			// find subcommand: first non-flag arg after tool name
			fmt.Fprintf(b, "        local _past_tool=0 _sub=\"\"\n")
			fmt.Fprintf(b, "        for _a in \"$@\"; do\n")
			fmt.Fprintf(b, "          if (( _past_tool )) && [[ \"$_a\" != -* ]]; then _sub=\"$_a\"; break; fi\n")
			fmt.Fprintf(b, "          [[ \"${_a%%%%@*}\" == %q ]] && _past_tool=1\n", toolName)
			fmt.Fprintf(b, "        done\n")
			fmt.Fprintf(b, "        case \"$_sub\" in\n")
			fmt.Fprintf(b, "          %s)\n", strings.Join(tc.InterceptSubcommands, "|"))
			writeLauncherInterceptBlock(b, shell, launcherName, toolName, flags, plugin, "            ")
			fmt.Fprintf(b, "            ;;\n") // terminates the inner case arm
			fmt.Fprintf(b, "        esac\n")
		}

		fmt.Fprintf(b, "      fi ;;\n")
	}

	fmt.Fprintf(b, "  esac\n")
	fmt.Fprintf(b, "  command %s \"$@\"\n", launcherName)
	fmt.Fprintf(b, "}\n")
}

func writeLauncherInterceptBlock(b *strings.Builder, shell, launcherName, toolName, flags, plugin, indent string) {
	fmt.Fprintf(b, "%slocal _ef; _ef=$(mktemp)\n", indent)
	fmt.Fprintf(b, "%scommand %s \"$@\" %s 2>\"$_ef\" | %s\n", indent, launcherName, flags, plugin)
	switch shell {
	case "zsh":
		fmt.Fprintf(b, "%slocal _r=${pipestatus[1]}; cat \"$_ef\" >&2; rm -f \"$_ef\"; return $_r\n", indent)
	case "bash":
		fmt.Fprintf(b, "%slocal _r=${PIPESTATUS[0]}; cat \"$_ef\" >&2; rm -f \"$_ef\"; return $_r\n", indent)
	}
}
```

- [ ] **Step 4: Run all hook tests — expect pass**

```bash
go test ./internal/hook/... -v
```

Expected: all 8 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/hook/
git commit -m "feat: hook generates launcher wrappers (uvx/pipx/npx) with tool name and subcommand detection"
```

---

## Task 7: Update prettyout-ruff plugin

**Files:**
- Modify: `cmd/prettyout-ruff/main.go`

- [ ] **Step 1: Rewrite `cmd/prettyout-ruff/main.go`**

Remove `backend/` path stripping and use `RunWithConfig`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type location struct {
	Row int `json:"row"`
}

type issue struct {
	Code        string   `json:"code"`
	Message     string   `json:"message"`
	Filename    string   `json:"filename"`
	Location    location `json:"location"`
	EndLocation location `json:"end_location"`
}

type ruleInfo struct {
	message string
	items   map[string]string // key (file\x00lineInfo) -> display lineInfo
}

func main() {
	formatter.RunWithConfig("ruff", format)
}

func format(data []byte, cfg formatter.Config) error {
	var issues []issue
	if err := json.Unmarshal(data, &issues); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	rules := map[string]*ruleInfo{}

	for _, iss := range issues {
		code := iss.Code
		if code == "" {
			code = "UNKNOWN"
		}

		filename := filepath.Base(iss.Filename)

		lineInfo := fmt.Sprintf("line %d", iss.Location.Row)
		if iss.Location.Row != iss.EndLocation.Row {
			lineInfo = fmt.Sprintf("lines %d-%d", iss.Location.Row, iss.EndLocation.Row)
		}

		key := filename + "\x00" + lineInfo

		if _, ok := rules[code]; !ok {
			rules[code] = &ruleInfo{items: map[string]string{}}
		}
		r := rules[code]
		if r.message == "" {
			r.message = truncate(iss.Message, cfg.MaxMessageLength)
		}
		r.items[key] = lineInfo
	}

	codes := make([]string, 0, len(rules))
	for code := range rules {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	for _, code := range codes {
		r := rules[code]
		if cfg.Colors {
			fmt.Printf("\033[1;33m%s\033[0m - \033[1m%s\033[0m\n", code, r.message)
		} else {
			fmt.Printf("%s - %s\n", code, r.message)
		}
		fmt.Println("Affected files:")

		keys := make([]string, 0, len(r.items))
		for k := range r.items {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			sep := strings.IndexByte(key, 0)
			file := key[:sep]
			line := key[sep+1:]
			fmt.Printf("  - %s — %s\n", file, line)
		}
		fmt.Println("----------------------------------------")
	}
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
```

- [ ] **Step 2: Build to verify no compile errors**

```bash
go build ./cmd/prettyout-ruff/...
```

Expected: builds with no errors.

- [ ] **Step 3: Commit**

```bash
git add cmd/prettyout-ruff/
git commit -m "feat: prettyout-ruff uses RunWithConfig, respects colors/max_message_length, removes backend/ hardcode"
```

---

## Task 8: Update prettyout-basedpyright plugin

**Files:**
- Modify: `cmd/prettyout-basedpyright/main.go`

- [ ] **Step 1: Rewrite `cmd/prettyout-basedpyright/main.go`**

```go
package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type position struct {
	Line int `json:"line"`
}

type diagRange struct {
	Start position `json:"start"`
	End   position `json:"end"`
}

type diagnostic struct {
	Rule    string    `json:"rule"`
	File    string    `json:"file"`
	Message string    `json:"message"`
	Range   diagRange `json:"range"`
}

type report struct {
	GeneralDiagnostics []diagnostic `json:"generalDiagnostics"`
}

type ruleInfo struct {
	message string
	items   map[string]struct{}
}

func main() {
	formatter.RunWithConfig("basedpyright", format)
}

func format(data []byte, cfg formatter.Config) error {
	var r report
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	rules := map[string]*ruleInfo{}

	for _, d := range r.GeneralDiagnostics {
		code := d.Rule
		if code == "" {
			code = "StaticAnalysis"
		}

		filename := filepath.Base(d.File)

		start := d.Range.Start.Line + 1
		end := d.Range.End.Line + 1
		lineInfo := fmt.Sprintf("line %d", start)
		if start != end {
			lineInfo = fmt.Sprintf("lines %d-%d", start, end)
		}

		key := filename + "\x00" + lineInfo

		if _, ok := rules[code]; !ok {
			rules[code] = &ruleInfo{items: map[string]struct{}{}}
		}
		ri := rules[code]
		if ri.message == "" {
			ri.message = truncate(d.Message, cfg.MaxMessageLength)
		}
		ri.items[key] = struct{}{}
	}

	codes := make([]string, 0, len(rules))
	for code := range rules {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	for _, code := range codes {
		ri := rules[code]
		if cfg.Colors {
			fmt.Printf("\033[1;34m%s\033[0m - \033[1m%s\033[0m\n", code, ri.message)
		} else {
			fmt.Printf("%s - %s\n", code, ri.message)
		}
		fmt.Println("Affected files:")

		keys := make([]string, 0, len(ri.items))
		for k := range ri.items {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			sep := strings.IndexByte(key, 0)
			file := key[:sep]
			line := key[sep+1:]
			fmt.Printf("  - %s — %s\n", file, line)
		}
		fmt.Println("--------------------------------------------------")
	}
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
```

- [ ] **Step 2: Build to verify**

```bash
go build ./cmd/prettyout-basedpyright/...
```

Expected: builds with no errors.

- [ ] **Step 3: Commit**

```bash
git add cmd/prettyout-basedpyright/
git commit -m "feat: prettyout-basedpyright uses RunWithConfig, respects colors/max_message_length, removes backend/ hardcode"
```

---

## Task 9: Update main CLI

**Files:**
- Modify: `cmd/prettyout/main.go`

- [ ] **Step 1: Rewrite `cmd/prettyout/main.go`**

Key changes: use new registry+config, rename `Hooks` → `Enabled`, wire `runHook` to new `Generate` signature.

```go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/hook"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "setup":
		runSetup()
	case "hook":
		runHook(args)
	case "enable":
		runEnable(args)
	case "disable":
		runDisable(args)
	case "list":
		runList()
	case "_enabled":
		runEnabled(args)
	default:
		fmt.Fprintf(os.Stderr, "prettyout: unknown command %q\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage:
  prettyout setup              Add shell hook to your rc file
  prettyout hook <shell>       Print shell hook code (zsh|bash)
  prettyout enable <tool>      Enable pretty output for a tool
  prettyout disable <tool>     Disable pretty output for a tool
  prettyout list               Show all tools and their status
  prettyout _enabled <tool>    Exit 0 if tool is enabled (internal)`)
}

func runSetup() {
	shell := shellName()
	rcFile := rcFilePath(shell)

	line := fmt.Sprintf(`eval "$(prettyout hook %s)"`, shell)

	data, _ := os.ReadFile(rcFile)
	if strings.Contains(string(data), "prettyout hook") {
		fmt.Println("prettyout hook already present in", rcFile)
		return
	}

	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "prettyout: cannot open %s: %v\n", rcFile, err)
		os.Exit(1)
	}
	defer f.Close()
	fmt.Fprintf(f, "\n# prettyout\n%s\n", line)
	fmt.Printf("Added to %s\nRun: source %s\n", rcFile, rcFile)
}

func runHook(args []string) {
	shell := "zsh"
	if len(args) > 0 {
		shell = args[0]
	}
	reg, err := registry.LoadBuiltin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "prettyout: failed to load registry: %v\n", err)
		os.Exit(1)
	}
	cfg := config.Load()
	reg.Merge(cfg.CustomTools)
	fmt.Print(hook.Generate(shell, reg, cfg))
}

func runEnable(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "prettyout enable: tool name required")
		os.Exit(1)
	}
	tool := args[0]
	reg, _ := registry.LoadBuiltin()
	cfg := config.Load()
	reg.Merge(cfg.CustomTools)
	if _, ok := reg.Tools[tool]; !ok {
		fmt.Fprintf(os.Stderr, "prettyout: unknown tool %q\n", tool)
		fmt.Fprintln(os.Stderr, "Known tools:", strings.Join(reg.SortedToolNames(), ", "))
		os.Exit(1)
	}
	cfg.Enabled[tool] = true
	config.Save(cfg)
	fmt.Printf("Enabled prettyout for %s\n", tool)
}

func runDisable(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "prettyout disable: tool name required")
		os.Exit(1)
	}
	cfg := config.Load()
	cfg.Enabled[args[0]] = false
	config.Save(cfg)
	fmt.Printf("Disabled prettyout for %s\n", args[0])
}

func runList() {
	reg, _ := registry.LoadBuiltin()
	cfg := config.Load()
	reg.Merge(cfg.CustomTools)
	fmt.Println("Tool             Status")
	fmt.Println("---------------  -------")
	for _, name := range reg.SortedToolNames() {
		status := "disabled"
		if cfg.Enabled[name] {
			status = "enabled"
		}
		fmt.Printf("%-16s %s\n", name, status)
	}
}

func runEnabled(args []string) {
	if len(args) == 0 {
		os.Exit(1)
	}
	cfg := config.Load()
	if cfg.Enabled[args[0]] {
		os.Exit(0)
	}
	os.Exit(1)
}

func shellName() string {
	shell := shellBase(os.Getenv("SHELL"))
	if shell == "" {
		return "zsh"
	}
	return shell
}

func shellBase(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

func rcFilePath(shell string) string {
	home, _ := os.UserHomeDir()
	switch shell {
	case "zsh":
		return home + "/.zshrc"
	case "bash":
		return home + "/.bashrc"
	default:
		fmt.Fprintf(os.Stderr, "prettyout: unsupported shell %q (only zsh and bash)\n", shell)
		os.Exit(1)
		return ""
	}
}
```

- [ ] **Step 2: Build entire project**

```bash
go build ./...
```

Expected: all three binaries build with no errors.

- [ ] **Step 3: Run all tests**

```bash
go test ./...
```

Expected: all tests PASS.

- [ ] **Step 4: Commit**

```bash
git add cmd/prettyout/
git commit -m "feat: wire new registry+config into main CLI, rename Hooks→Enabled"
```

---

## Task 10: Smoke test

Verify the full system works end-to-end.

- [ ] **Step 1: Build all binaries**

```bash
go build -o /tmp/prettyout ./cmd/prettyout/
go build -o /tmp/prettyout-ruff ./cmd/prettyout-ruff/
go build -o /tmp/prettyout-basedpyright ./cmd/prettyout-basedpyright/
```

Expected: all three binaries created.

- [ ] **Step 2: Verify hook output for zsh**

```bash
/tmp/prettyout hook zsh
```

Expected output contains:
- `ruff()` function with `case "${1:-}"` and `check)` case
- `--output-format=json` flag
- `mktemp` for stderr separation
- No `2>&1`
- `uvx()` wrapper with `ruff` case

- [ ] **Step 3: Verify ruff format passes through**

```bash
# Generate hook and source it in a subshell
eval "$(/tmp/prettyout hook zsh)"
# Check hook only intercepts 'check', not 'format'
# (dry-run: just look at what the ruff function does for 'format')
type ruff   # in zsh: shows the function definition
```

Expected: the `ruff` function only calls the formatter for `check`, falls through to `command ruff` for `format`.

- [ ] **Step 4: Verify prettyout-ruff formats JSON correctly**

```bash
echo '[{"code":"E501","message":"Line too long","filename":"/project/src/foo.py","location":{"row":10,"column":0},"end_location":{"row":10,"column":120}}]' \
  | /tmp/prettyout-ruff
```

Expected: colored output like:
```
E501 - Line too long
Affected files:
  - foo.py — line 10
----------------------------------------
```

- [ ] **Step 5: Verify enable/disable/list commands**

```bash
/tmp/prettyout list
/tmp/prettyout enable ruff
/tmp/prettyout list
/tmp/prettyout disable ruff
```

Expected: list shows `ruff` and `basedpyright`; enable/disable update the status shown by list.

- [ ] **Step 6: Final commit**

```bash
git add -A
git commit -m "chore: smoke test verified — prettyout redesign complete"
```
