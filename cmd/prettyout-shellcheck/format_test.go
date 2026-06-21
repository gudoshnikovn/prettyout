package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
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

func noColors() formatter.Config {
	cfg := formatter.DefaultConfig()
	cfg.Colors = false
	return cfg
}

const twoFileJSON = `[
  {"file":"foo.sh","line":10,"level":"warning","code":2034,"message":"x appears unused"},
  {"file":"foo.sh","line":20,"level":"warning","code":2034,"message":"x appears unused"},
  {"file":"bar.sh","line":5,"level":"error","code":1091,"message":"not following source"}
]`

func TestFormat_clean(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte("[]"), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "0 issues") {
		t.Errorf("clean run: want 0 issues, got:\n%s", out)
	}
}

func TestFormat_byRule(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "SC2034") {
		t.Errorf("want SC2034, got:\n%s", out)
	}
	if !strings.Contains(out, "SC1091") {
		t.Errorf("want SC1091, got:\n%s", out)
	}
	if !strings.Contains(out, "3 issues · 2 rules · 2 files") {
		t.Errorf("want summary, got:\n%s", out)
	}
}

func TestFormat_byFile(t *testing.T) {
	cfg := noColors()
	cfg.GroupBy = "file"
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "foo.sh") {
		t.Errorf("want foo.sh, got:\n%s", out)
	}
	if !strings.Contains(out, "bar.sh") {
		t.Errorf("want bar.sh, got:\n%s", out)
	}
}

func TestFormat_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.OnlyRules = []string{"SC1091"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "SC2034") {
		t.Errorf("SC2034 should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "SC1091") {
		t.Errorf("SC1091 should appear, got:\n%s", out)
	}
}

func TestFormat_onlyFiles(t *testing.T) {
	cfg := noColors()
	cfg.OnlyFiles = []string{"bar.sh"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "foo.sh") {
		t.Errorf("foo.sh should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "bar.sh") {
		t.Errorf("bar.sh should appear, got:\n%s", out)
	}
}

func TestFormat_invalidJSON(t *testing.T) {
	cfg := noColors()
	err := format([]byte("not json"), cfg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFormat_statsMode(t *testing.T) {
	cfg := noColors()
	cfg.Stats = true
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "SC2034") {
		t.Errorf("stats: want SC2034, got:\n%s", out)
	}
}

func TestShellcheckColor_withColors(t *testing.T) {
	if shellcheckColor("error", true) == "" {
		t.Error("shellcheckColor(error, true): want ANSI code")
	}
	if shellcheckColor("warning", true) == "" {
		t.Error("shellcheckColor(warning, true): want ANSI code")
	}
	if shellcheckColor("info", true) == "" {
		t.Error("shellcheckColor(info, true): want ANSI code (dim)")
	}
	if shellcheckColor("error", false) != "" {
		t.Error("shellcheckColor(error, false): want empty")
	}
}

func TestFormat_byFile_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.GroupBy = "file"
	cfg.OnlyRules = []string{"SC2034"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	// bar.sh only has SC1091, should be skipped
	if strings.Contains(out, "bar.sh") {
		t.Errorf("byFile+onlyRules: bar.sh should be filtered, got:\n%s", out)
	}
}

func TestScCode(t *testing.T) {
	if got := scCode(2034); got != "SC2034" {
		t.Errorf("got %q, want SC2034", got)
	}
}
