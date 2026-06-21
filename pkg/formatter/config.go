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
	Sort             string   // "" | "alpha" (default) | "count"
	OnlyRules        []string // if non-empty, only show rules in this list
	OnlyFiles        []string // if non-empty, only show files matching these prefixes
	Stats            bool     // PO_STATS=1: compact per-rule count table
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
	Stats            *bool          `yaml:"stats"`
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
	if s.Stats != nil {
		cfg.Stats = *s.Stats
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
