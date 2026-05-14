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
