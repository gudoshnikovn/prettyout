package runner

import (
	"testing"

	"github.com/gudoshnikovn/prettyout/internal/config"
	"github.com/gudoshnikovn/prettyout/internal/registry"
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
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), enabledCfg("ruff"), true, false)
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
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), enabledCfg(), true, false)
	if d.Intercept {
		t.Error("disabled tool should not intercept")
	}
}

func TestDecide_wrongSubcommand(t *testing.T) {
	d := Decide("ruff", []string{"format", "."}, ruffRegistry(), enabledCfg("ruff"), true, false)
	if d.Intercept {
		t.Error("ruff format should not intercept")
	}
}

func TestDecide_passthroughFlag(t *testing.T) {
	d := Decide("ruff", []string{"check", "--watch", "."}, ruffRegistry(), enabledCfg("ruff"), true, false)
	if d.Intercept {
		t.Error("--watch should cause passthrough")
	}
}

func TestDecide_ciModeNever(t *testing.T) {
	cfg := enabledCfg("ruff")
	cfg.CIMode = "never"
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), cfg, true, false)
	if d.Intercept {
		t.Error("ci_mode=never should not intercept")
	}
}

func TestDecide_ciModeAutoNotTTY(t *testing.T) {
	cfg := enabledCfg("ruff")
	cfg.CIMode = "auto"
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), cfg, false, false)
	if d.Intercept {
		t.Error("ci_mode=auto + not TTY should not intercept")
	}
}

func TestDecide_ciModeAutoTTY(t *testing.T) {
	cfg := enabledCfg("ruff")
	cfg.CIMode = "auto"
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), cfg, true, false)
	if !d.Intercept {
		t.Error("ci_mode=auto + TTY should intercept")
	}
}

func TestDecide_unknownTool(t *testing.T) {
	d := Decide("unknown", []string{"check", "."}, ruffRegistry(), enabledCfg("unknown"), true, false)
	if d.Intercept {
		t.Error("unknown tool should not intercept")
	}
}

func TestDecide_pluginOverride(t *testing.T) {
	cfg := enabledCfg("ruff")
	cfg.Plugins["ruff"] = "/home/user/my-ruff.py"
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), cfg, true, false)
	if d.Plugin != "/home/user/my-ruff.py" {
		t.Errorf("plugin = %q, want /home/user/my-ruff.py", d.Plugin)
	}
}

func TestDecide_launcher_intercepts(t *testing.T) {
	cfg := enabledCfg("ruff")
	d := Decide("uvx", []string{"ruff", "check", "."}, uvxRegistry(), cfg, true, false)
	if !d.Intercept {
		t.Fatal("uvx ruff check should intercept")
	}
	if d.RealCmd != "uvx" {
		t.Errorf("RealCmd = %q, want uvx", d.RealCmd)
	}
}

func TestDecide_launcher_unknownTool(t *testing.T) {
	cfg := enabledCfg("ruff")
	d := Decide("uvx", []string{"sometool", "check", "."}, uvxRegistry(), cfg, true, false)
	if d.Intercept {
		t.Error("unknown tool via launcher should not intercept")
	}
}

func TestDecide_launcher_disabledTool(t *testing.T) {
	d := Decide("uvx", []string{"ruff", "check", "."}, uvxRegistry(), enabledCfg(), true, false)
	if d.Intercept {
		t.Error("disabled tool via launcher should not intercept")
	}
}

func TestDecide_noArgs(t *testing.T) {
	d := Decide("ruff", []string{}, ruffRegistry(), enabledCfg("ruff"), true, false)
	if d.Intercept {
		t.Error("no args: no subcommand to intercept")
	}
}

func TestDecide_outputArgConflict(t *testing.T) {
	d := Decide("ruff", []string{"check", "--output-format=github", "."}, ruffRegistry(), enabledCfg("ruff"), true, false)
	if d.Intercept {
		t.Error("user-specified --output-format should cause passthrough")
	}
}

func TestDecide_outputArgConflict_sameValue(t *testing.T) {
	d := Decide("ruff", []string{"check", "--output-format=json", "."}, ruffRegistry(), enabledCfg("ruff"), true, false)
	if d.Intercept {
		t.Error("user specifying same output format should still be passthrough")
	}
}

func TestDecide_outputArgNoConflict(t *testing.T) {
	d := Decide("ruff", []string{"check", "--select=E501", "."}, ruffRegistry(), enabledCfg("ruff"), true, false)
	if !d.Intercept {
		t.Error("unrelated flags should not prevent intercept")
	}
}

func TestHasOutputArgConflict_withEquals(t *testing.T) {
	if !hasOutputArgConflict([]string{"--output-format=json"}, []string{"--output-format=github"}) {
		t.Error("expected conflict")
	}
}

func TestHasOutputArgConflict_noEquals(t *testing.T) {
	if !hasOutputArgConflict([]string{"--outputjson"}, []string{"--outputjson"}) {
		t.Error("exact match should conflict")
	}
	if hasOutputArgConflict([]string{"--outputjson"}, []string{"--outputjson-extra"}) {
		t.Error("prefix should not conflict when no '=' in output_arg")
	}
}

func TestHasOutputArgConflict_noConflict(t *testing.T) {
	if hasOutputArgConflict([]string{"--output-format=json"}, []string{"check", "."}) {
		t.Error("no conflict expected")
	}
}

func TestHasOutputArgConflict_noFalsePositive(t *testing.T) {
	// --output-format-extra is a different flag, should not conflict with --output-format key
	if hasOutputArgConflict([]string{"--output-format=json"}, []string{"--output-format-extra=foo"}) {
		t.Error("different flag with same prefix should not conflict")
	}
}

func TestExecute_passthrough(t *testing.T) {
	// passthrough: echo hello should print "hello\n" and exit 0
	d := Decision{
		Intercept:    false,
		RealCmd:      "echo",
		OriginalArgs: []string{"hello"},
	}
	code := Execute(d)
	if code != 0 {
		t.Errorf("passthrough exit code = %d, want 0", code)
	}
}

func TestExecute_passthroughExitCode(t *testing.T) {
	// passthrough with non-zero exit: "false" exits 1
	d := Decision{
		Intercept:    false,
		RealCmd:      "false",
		OriginalArgs: []string{},
	}
	code := Execute(d)
	if code != 1 {
		t.Errorf("false exit code = %d, want 1", code)
	}
}

func TestExecute_intercept(t *testing.T) {
	// intercept: pipe "echo hello" through "cat" — should exit 0
	d := Decision{
		Intercept:       true,
		RealCmd:         "echo",
		TransformedArgs: []string{"hello"},
		Plugin:          "cat",
	}
	code := Execute(d)
	if code != 0 {
		t.Errorf("intercept exit code = %d, want 0", code)
	}
}

func TestExecute_interceptExitCode(t *testing.T) {
	// intercept: tool exits non-zero, plugin is cat — exit code is from tool
	d := Decision{
		Intercept:       true,
		RealCmd:         "false",
		TransformedArgs: []string{},
		Plugin:          "cat",
	}
	code := Execute(d)
	if code != 1 {
		t.Errorf("false via intercept = %d, want 1", code)
	}
}

func TestDecide_poRaw_forcesPassthrough(t *testing.T) {
	t.Setenv("PO_RAW", "1")
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), enabledCfg("ruff"), true, false)
	if d.Intercept {
		t.Error("PO_RAW=1 should force passthrough")
	}
}

func TestDecide_preCommit_treatedAsTTY(t *testing.T) {
	t.Setenv("PRE_COMMIT", "1")
	cfg := enabledCfg("ruff")
	cfg.CIMode = "auto"
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), cfg, false, false)
	if !d.Intercept {
		t.Error("PRE_COMMIT=1 should treat non-TTY as TTY")
	}
}

func TestExecute_returnsPluginError_whenToolSucceeds(t *testing.T) {
	// Integration-level concern; covered by verifying Execute checks pluginCmd exit code.
	t.Skip("covered by integration test in test/run.sh")
}

// debug=true tests — cover all the fmt.Fprintf(os.Stderr, ...) branches in Decide.

func TestDecide_debug_direct_intercept(t *testing.T) {
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), enabledCfg("ruff"), true, true)
	if !d.Intercept {
		t.Fatal("expected intercept")
	}
}

func TestDecide_debug_direct_passthrough_ciNever(t *testing.T) {
	cfg := enabledCfg("ruff")
	cfg.CIMode = "never"
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), cfg, true, true)
	if d.Intercept {
		t.Error("ci_mode=never with debug should still passthrough")
	}
}

func TestDecide_debug_direct_passthrough_passthroughFlag(t *testing.T) {
	d := Decide("ruff", []string{"check", "--watch", "."}, ruffRegistry(), enabledCfg("ruff"), true, true)
	if d.Intercept {
		t.Error("passthrough flag with debug should still passthrough")
	}
}

func TestDecide_debug_direct_passthrough_outputConflict(t *testing.T) {
	d := Decide("ruff", []string{"check", "--output-format=github", "."}, ruffRegistry(), enabledCfg("ruff"), true, true)
	if d.Intercept {
		t.Error("output arg conflict with debug should still passthrough")
	}
}

func TestDecide_debug_direct_passthrough_wrongSubcommand(t *testing.T) {
	d := Decide("ruff", []string{"format", "."}, ruffRegistry(), enabledCfg("ruff"), true, true)
	if d.Intercept {
		t.Error("wrong subcommand with debug should still passthrough")
	}
}

func TestDecide_debug_direct_passthrough_unknownTool(t *testing.T) {
	d := Decide("unknown", []string{"check", "."}, ruffRegistry(), enabledCfg("unknown"), true, true)
	if d.Intercept {
		t.Error("unknown tool with debug should passthrough")
	}
}

func TestDecide_debug_direct_passthrough_disabledTool(t *testing.T) {
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), enabledCfg(), true, true)
	if d.Intercept {
		t.Error("disabled tool with debug should passthrough")
	}
}

func TestDecide_debug_poRaw(t *testing.T) {
	t.Setenv("PO_RAW", "1")
	d := Decide("ruff", []string{"check", "."}, ruffRegistry(), enabledCfg("ruff"), true, true)
	if d.Intercept {
		t.Error("PO_RAW with debug should passthrough")
	}
}

func TestDecide_debug_launcher_intercept(t *testing.T) {
	d := Decide("uvx", []string{"ruff", "check", "."}, uvxRegistry(), enabledCfg("ruff"), true, true)
	if !d.Intercept {
		t.Fatal("uvx ruff check with debug should intercept")
	}
}

func TestDecide_debug_launcher_unknownTool(t *testing.T) {
	d := Decide("uvx", []string{"sometool", "."}, uvxRegistry(), enabledCfg("ruff"), true, true)
	if d.Intercept {
		t.Error("unknown launcher tool with debug should passthrough")
	}
}

func TestDecide_debug_launcher_disabledTool(t *testing.T) {
	d := Decide("uvx", []string{"ruff", "check", "."}, uvxRegistry(), enabledCfg(), true, true)
	if d.Intercept {
		t.Error("disabled launcher tool with debug should passthrough")
	}
}

func TestDecide_launcher_passthroughFlag(t *testing.T) {
	reg := &registry.Registry{
		Tools: map[string]registry.ToolConfig{
			"ruff": {
				Plugin:           "prettyout-ruff",
				OutputArgs:       []string{"--output-format=json"},
				PassthroughFlags: []string{"--watch"},
				Launchers:        []string{"uvx"},
			},
		},
		Launchers: map[string]registry.LauncherConfig{
			"uvx": {},
		},
	}
	d := Decide("uvx", []string{"ruff", "--watch", "check", "."}, reg, enabledCfg("ruff"), true, false)
	if d.Intercept {
		t.Error("passthrough flag via launcher should passthrough")
	}
}

func TestDecide_launcher_outputArgConflict(t *testing.T) {
	d := Decide("uvx", []string{"ruff", "check", "--output-format=github", "."}, uvxRegistry(), enabledCfg("ruff"), true, false)
	if d.Intercept {
		t.Error("output arg conflict via launcher should passthrough")
	}
}

func TestDecide_launcher_wrongSubcommand(t *testing.T) {
	d := Decide("uvx", []string{"ruff", "format", "."}, uvxRegistry(), enabledCfg("ruff"), true, false)
	if d.Intercept {
		t.Error("wrong subcommand via launcher should passthrough")
	}
}

func TestExecute_interceptJSONOnStderr(t *testing.T) {
	// Tool writes to stderr; JSONOnStderr=true pipes tool stderr to plugin stdin.
	d := Decision{
		Intercept:       true,
		RealCmd:         "sh",
		TransformedArgs: []string{"-c", "echo hello >&2"},
		Plugin:          "cat",
		JSONOnStderr:    true,
	}
	code := Execute(d)
	if code != 0 {
		t.Errorf("JSONOnStderr intercept exit code = %d, want 0", code)
	}
}

func TestExecute_interceptCommandNotFound(t *testing.T) {
	d := Decision{
		Intercept:       true,
		RealCmd:         "definitely-not-a-real-command-xyz-abc",
		TransformedArgs: []string{},
		Plugin:          "cat",
	}
	code := Execute(d)
	if code != 127 {
		t.Errorf("not-found intercept exit code = %d, want 127", code)
	}
}

func TestExecute_interceptPluginNotFound(t *testing.T) {
	d := Decision{
		Intercept:       true,
		RealCmd:         "echo",
		TransformedArgs: []string{"hello"},
		Plugin:          "definitely-not-a-real-plugin-xyz-abc",
	}
	code := Execute(d)
	if code != 1 {
		t.Errorf("plugin not found exit code = %d, want 1", code)
	}
}
