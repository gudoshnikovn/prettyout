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
  {
    "source": "foo.css",
    "parseErrors": [],
    "warnings": [
      {"line":10,"rule":"color-no-invalid-hex","severity":"error","text":"Invalid hex color"},
      {"line":20,"rule":"color-no-invalid-hex","severity":"error","text":"Invalid hex color"}
    ]
  },
  {
    "source": "bar.css",
    "parseErrors": [],
    "warnings": [
      {"line":5,"rule":"unit-no-unknown","severity":"warning","text":"Unknown unit"}
    ]
  }
]`

func TestFormat_cleanEmptyInput(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(""), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "0 issues") {
		t.Errorf("empty input: want 0 issues, got:\n%s", out)
	}
}

func TestFormat_cleanNoWarnings(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(`[{"source":"a.css","parseErrors":[],"warnings":[]}]`), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "0 issues") {
		t.Errorf("no warnings: want 0 issues, got:\n%s", out)
	}
}

func TestFormat_byRule(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "color-no-invalid-hex") {
		t.Errorf("want color-no-invalid-hex, got:\n%s", out)
	}
	if !strings.Contains(out, "unit-no-unknown") {
		t.Errorf("want unit-no-unknown, got:\n%s", out)
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
	if !strings.Contains(out, "foo.css") {
		t.Errorf("want foo.css, got:\n%s", out)
	}
	if !strings.Contains(out, "bar.css") {
		t.Errorf("want bar.css, got:\n%s", out)
	}
}

func TestFormat_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.OnlyRules = []string{"unit-no-unknown"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "color-no-invalid-hex") {
		t.Errorf("color-no-invalid-hex should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "unit-no-unknown") {
		t.Errorf("unit-no-unknown should appear, got:\n%s", out)
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
	if !strings.Contains(out, "color-no-invalid-hex") {
		t.Errorf("stats: want color-no-invalid-hex, got:\n%s", out)
	}
}

func TestFormat_nonJSONInput(t *testing.T) {
	cfg := noColors()
	err := format([]byte("stylelint: command failed"), cfg)
	if err == nil {
		t.Error("expected error for non-JSON input")
	}
}

func TestFormat_parseError_noLine(t *testing.T) {
	// parseErrors in stylelint output create issues with line: 0
	// This covers the else branch in formatByRule when iss.line == 0
	cfg := noColors()
	input := `[{"source":"bad.css","parseErrors":[{}],"warnings":[]}]`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "parse-error") {
		t.Errorf("parse error: want 'parse-error' in output, got:\n%s", out)
	}
}

func TestFormat_withColors(t *testing.T) {
	cfg := formatter.DefaultConfig()
	cfg.Colors = true
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "\033[") {
		t.Errorf("withColors: want ANSI codes in output, got:\n%s", out)
	}
}

func TestFormat_byFile_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.GroupBy = "file"
	cfg.OnlyRules = []string{"color-no-invalid-hex"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	// bar.css only has unit-no-unknown, should be filtered out
	if strings.Contains(out, "bar.css") {
		t.Errorf("byFile+onlyRules: bar.css should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "foo.css") {
		t.Errorf("byFile+onlyRules: foo.css should appear, got:\n%s", out)
	}
}
