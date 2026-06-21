package launcher

import (
	"testing"

	"github.com/gudoshnikovn/prettyout/internal/registry"
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
