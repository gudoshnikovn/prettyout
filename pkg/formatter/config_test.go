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

func TestApplyEnvOverrides_groupBy(t *testing.T) {
	t.Setenv("PO_GROUP_BY", "file")
	cfg := DefaultConfig()
	ApplyEnvOverrides(&cfg)
	if cfg.GroupBy != "file" {
		t.Errorf("want GroupBy=file, got %q", cfg.GroupBy)
	}
}

func TestApplyEnvOverrides_sort(t *testing.T) {
	t.Setenv("PO_SORT", "count")
	cfg := DefaultConfig()
	ApplyEnvOverrides(&cfg)
	if cfg.Sort != "count" {
		t.Errorf("want Sort=count, got %q", cfg.Sort)
	}
}

func TestApplyEnvOverrides_onlyRules(t *testing.T) {
	t.Setenv("PO_ONLY_RULES", "E501, F401")
	cfg := DefaultConfig()
	ApplyEnvOverrides(&cfg)
	if len(cfg.OnlyRules) != 2 || cfg.OnlyRules[0] != "E501" || cfg.OnlyRules[1] != "F401" {
		t.Errorf("unexpected OnlyRules: %v", cfg.OnlyRules)
	}
}

func TestApplyEnvOverrides_onlyFiles(t *testing.T) {
	t.Setenv("PO_ONLY_FILES", "src/,tests/")
	cfg := DefaultConfig()
	ApplyEnvOverrides(&cfg)
	if len(cfg.OnlyFiles) != 2 || cfg.OnlyFiles[0] != "src/" || cfg.OnlyFiles[1] != "tests/" {
		t.Errorf("unexpected OnlyFiles: %v", cfg.OnlyFiles)
	}
}

func TestApplyEnvOverrides_noVars(t *testing.T) {
	cfg := DefaultConfig()
	ApplyEnvOverrides(&cfg)
	if cfg.GroupBy != "rule" || cfg.Sort != "" || cfg.OnlyRules != nil {
		t.Error("no env vars should leave config unchanged")
	}
}

func TestLoadConfig_readsSortAndFilters(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "prettyout")
	_ = os.MkdirAll(dir, 0755)
	_ = os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`
settings:
  ruff:
    sort: count
    only_rules:
      - E501
      - F401
    only_files:
      - src/
`), 0644)

	cfg := LoadConfig("ruff")
	if cfg.Sort != "count" {
		t.Errorf("want sort=count, got %q", cfg.Sort)
	}
	if len(cfg.OnlyRules) != 2 || cfg.OnlyRules[0] != "E501" || cfg.OnlyRules[1] != "F401" {
		t.Errorf("unexpected OnlyRules: %v", cfg.OnlyRules)
	}
	if len(cfg.OnlyFiles) != 1 || cfg.OnlyFiles[0] != "src/" {
		t.Errorf("unexpected OnlyFiles: %v", cfg.OnlyFiles)
	}
}
