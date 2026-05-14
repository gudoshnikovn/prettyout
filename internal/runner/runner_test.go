package runner

import (
	"testing"

	"github.com/gudoshnikov_na/prettyout/internal/config"
	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

func ruffRegistry() *registry.Registry {
	return &registry.Registry{
		Tools: map[string]registry.ToolConfig{
			"ruff": {
				Plugin:               "prettyout-ruff",
				OutputArgs:           []string{"--output-format=json"},
				InterceptSubcommands: []string{"check"},
				PassthroughFlags:     []string{"--watch", "-W"},
			},
		},
		Launchers: map[string]registry.LauncherConfig{},
	}
}

func uvxRegistry() *registry.Registry {
	return &registry.Registry{
		Tools: map[string]registry.ToolConfig{
			"ruff": {
				Plugin:               "prettyout-ruff",
				OutputArgs:           []string{"--output-format=json"},
				InterceptSubcommands: []string{"check"},
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
}

func enabledCfg(tools ...string) *config.Config {
	cfg := &config.Config{
		Enabled:     map[string]bool{},
		Plugins:     map[string]string{},
		Settings:    map[string]config.FormatterSettings{},
		CustomTools: map[string]registry.ToolConfig{},
		CIMode:      "always",
	}
	for _, t := range tools {
		cfg.Enabled[t] = true
	}
	return cfg
}

func TestDecide_intercepts(t *testing.T) {
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), enabledCfg("ruff"), true)
	if !d.Intercept {
		t.Fatal("expected intercept")
	}
	if d.Plugin != "prettyout-ruff" {
		t.Errorf("plugin = %q, want prettyout-ruff", d.Plugin)
	}
	if d.TransformedArgs[len(d.TransformedArgs)-1] != "--output-format=json" {
		t.Errorf("output arg not appended: %v", d.TransformedArgs)
	}
}

func TestDecide_disabledTool(t *testing.T) {
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), enabledCfg(), true)
	if d.Intercept {
		t.Error("disabled tool should not intercept")
	}
}

func TestDecide_wrongSubcommand(t *testing.T) {
	d := Decide("ruff", []string{"format", "."}, ruffRegistry(), enabledCfg("ruff"), true)
	if d.Intercept {
		t.Error("ruff format should not intercept")
	}
}

func TestDecide_passthroughFlag(t *testing.T) {
	d := Decide("ruff", []string{"check", "--watch", "."}, ruffRegistry(), enabledCfg("ruff"), true)
	if d.Intercept {
		t.Error("--watch should cause passthrough")
	}
}

func TestDecide_ciModeNever(t *testing.T) {
	cfg := enabledCfg("ruff")
	cfg.CIMode = "never"
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), cfg, true)
	if d.Intercept {
		t.Error("ci_mode=never should not intercept")
	}
}

func TestDecide_ciModeAutoNotTTY(t *testing.T) {
	cfg := enabledCfg("ruff")
	cfg.CIMode = "auto"
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), cfg, false)
	if d.Intercept {
		t.Error("ci_mode=auto + not TTY should not intercept")
	}
}

func TestDecide_ciModeAutoTTY(t *testing.T) {
	cfg := enabledCfg("ruff")
	cfg.CIMode = "auto"
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), cfg, true)
	if !d.Intercept {
		t.Error("ci_mode=auto + TTY should intercept")
	}
}

func TestDecide_unknownTool(t *testing.T) {
	d := Decide("unknown", []string{"check", "."}, ruffRegistry(), enabledCfg("unknown"), true)
	if d.Intercept {
		t.Error("unknown tool should not intercept")
	}
}

func TestDecide_pluginOverride(t *testing.T) {
	cfg := enabledCfg("ruff")
	cfg.Plugins["ruff"] = "/home/user/my-ruff.py"
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), cfg, true)
	if d.Plugin != "/home/user/my-ruff.py" {
		t.Errorf("plugin = %q, want /home/user/my-ruff.py", d.Plugin)
	}
}

func TestDecide_launcher_intercepts(t *testing.T) {
	cfg := enabledCfg("ruff")
	d := Decide("uvx", []string{"ruff", "check", "."}, uvxRegistry(), cfg, true)
	if !d.Intercept {
		t.Fatal("uvx ruff check should intercept")
	}
	if d.RealCmd != "uvx" {
		t.Errorf("RealCmd = %q, want uvx", d.RealCmd)
	}
}

func TestDecide_launcher_unknownTool(t *testing.T) {
	cfg := enabledCfg("ruff")
	d := Decide("uvx", []string{"sometool", "check", "."}, uvxRegistry(), cfg, true)
	if d.Intercept {
		t.Error("unknown tool via launcher should not intercept")
	}
}

func TestDecide_launcher_disabledTool(t *testing.T) {
	d := Decide("uvx", []string{"ruff", "check", "."}, uvxRegistry(), enabledCfg(), true)
	if d.Intercept {
		t.Error("disabled tool via launcher should not intercept")
	}
}

func TestDecide_noArgs(t *testing.T) {
	d := Decide("ruff", []string{}, ruffRegistry(), enabledCfg("ruff"), true)
	if d.Intercept {
		t.Error("no args: no subcommand to intercept")
	}
}
