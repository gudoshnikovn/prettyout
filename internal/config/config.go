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
	Enabled     map[string]bool                `yaml:"enabled"`
	Plugins     map[string]string              `yaml:"plugins"`
	Settings    map[string]FormatterSettings   `yaml:"settings"`
	CustomTools map[string]registry.ToolConfig `yaml:"custom_tools"`
	CIMode      string                         `yaml:"ci_mode"`
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

// GlobalConfigPath returns the path to the user's global prettyout config file.
func GlobalConfigPath() string {
	return globalConfigPath()
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
