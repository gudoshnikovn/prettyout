package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Hooks map[string]bool `yaml:"hooks"`
}

func Load() *Config {
	cfg := &Config{Hooks: map[string]bool{}}
	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg
	}
	_ = yaml.Unmarshal(data, cfg)
	if cfg.Hooks == nil {
		cfg.Hooks = map[string]bool{}
	}
	return cfg
}

func Save(cfg *Config) {
	path := configPath()
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

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "prettyout", "config.yaml")
}
