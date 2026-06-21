package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(fn func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

func setupTempHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SHELL", "/bin/zsh")
}

func TestPrintUsage(t *testing.T) {
	out := captureOutput(func() { printUsage() })
	if !strings.Contains(out, "prettyout setup") {
		t.Errorf("printUsage: want 'prettyout setup', got:\n%s", out)
	}
}

func TestShellBase(t *testing.T) {
	cases := []struct{ path, want string }{
		{"/bin/zsh", "zsh"},
		{"/usr/bin/bash", "bash"},
		{"zsh", "zsh"},
		{"", ""},
	}
	for _, c := range cases {
		if got := shellBase(c.path); got != c.want {
			t.Errorf("shellBase(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestShellName_fromEnv(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	if got := shellName(); got != "bash" {
		t.Errorf("shellName() = %q, want bash", got)
	}
}

func TestShellName_empty(t *testing.T) {
	t.Setenv("SHELL", "")
	if got := shellName(); got != "zsh" {
		t.Errorf("shellName() with empty SHELL = %q, want zsh", got)
	}
}

func TestRunHook_zsh(t *testing.T) {
	setupTempHome(t)
	out := captureOutput(func() { runHook([]string{"zsh"}) })
	if !strings.Contains(out, "prettyout") {
		t.Errorf("runHook zsh: want prettyout in hook, got:\n%s", out)
	}
}

func TestRunHook_defaultsToZsh(t *testing.T) {
	setupTempHome(t)
	out := captureOutput(func() { runHook(nil) })
	if !strings.Contains(out, "prettyout") {
		t.Errorf("runHook no args: want prettyout in hook, got:\n%s", out)
	}
}

func TestRunList(t *testing.T) {
	setupTempHome(t)
	out := captureOutput(func() { runList(nil) })
	if !strings.Contains(out, "ruff") {
		t.Errorf("runList: want ruff in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Status") {
		t.Errorf("runList: want 'Status' header, got:\n%s", out)
	}
}

func TestRunList_available(t *testing.T) {
	setupTempHome(t)
	out := captureOutput(func() { runList([]string{"--available"}) })
	if !strings.Contains(out, "ruff") {
		t.Errorf("runList --available: want ruff, got:\n%s", out)
	}
	if !strings.Contains(out, "Plugin") {
		t.Errorf("runList --available: want 'Plugin' header, got:\n%s", out)
	}
}

func TestRunStatus(t *testing.T) {
	setupTempHome(t)
	out := captureOutput(func() { runStatus() })
	if !strings.Contains(out, "prettyout v") {
		t.Errorf("runStatus: want version, got:\n%s", out)
	}
	if !strings.Contains(out, "Tools:") {
		t.Errorf("runStatus: want 'Tools:', got:\n%s", out)
	}
}

func TestRunCompletion_zsh(t *testing.T) {
	setupTempHome(t)
	out := captureOutput(func() { runCompletion([]string{"zsh"}) })
	if out == "" {
		t.Error("runCompletion zsh: want non-empty output")
	}
}

func TestRunCompletion_bash(t *testing.T) {
	setupTempHome(t)
	out := captureOutput(func() { runCompletion([]string{"bash"}) })
	if out == "" {
		t.Error("runCompletion bash: want non-empty output")
	}
}

func TestRunCompletions_tools(t *testing.T) {
	setupTempHome(t)
	out := captureOutput(func() { runCompletions([]string{"tools"}) })
	if !strings.Contains(out, "ruff") {
		t.Errorf("runCompletions tools: want ruff, got:\n%s", out)
	}
}

func TestRunEnable_knownTool(t *testing.T) {
	setupTempHome(t)
	out := captureOutput(func() { runEnable([]string{"ruff"}) })
	if !strings.Contains(out, "Enabled") {
		t.Errorf("runEnable ruff: want 'Enabled', got:\n%s", out)
	}
}

func TestRunEnable_all(t *testing.T) {
	setupTempHome(t)
	out := captureOutput(func() { runEnable([]string{"--all"}) })
	// Either some tools were enabled, or none were found installed
	if out == "" {
		t.Error("runEnable --all: want non-empty output")
	}
}

func TestRunDisable_tool(t *testing.T) {
	setupTempHome(t)
	out := captureOutput(func() { runDisable([]string{"ruff"}) })
	if !strings.Contains(out, "Disabled") {
		t.Errorf("runDisable ruff: want 'Disabled', got:\n%s", out)
	}
}

func TestRunDisable_all(t *testing.T) {
	setupTempHome(t)
	out := captureOutput(func() { runDisable([]string{"--all"}) })
	if !strings.Contains(out, "Disabled all") {
		t.Errorf("runDisable --all: want 'Disabled all', got:\n%s", out)
	}
}

func TestMustLoadBuiltin(t *testing.T) {
	reg := mustLoadBuiltin()
	if _, ok := reg.Tools["ruff"]; !ok {
		t.Error("mustLoadBuiltin: want ruff in registry")
	}
}
