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
    "filePath": "foo.js",
    "messages": [
      {"ruleId": "no-unused-vars", "severity": 2, "message": "unused variable", "line": 10},
      {"ruleId": "no-unused-vars", "severity": 2, "message": "unused variable", "line": 20}
    ]
  },
  {
    "filePath": "bar.js",
    "messages": [
      {"ruleId": "semi", "severity": 1, "message": "missing semicolon", "line": 5}
    ]
  }
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
	if !strings.Contains(out, "no-unused-vars") {
		t.Errorf("want no-unused-vars, got:\n%s", out)
	}
	if !strings.Contains(out, "semi") {
		t.Errorf("want semi, got:\n%s", out)
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
	if !strings.Contains(out, "foo.js") {
		t.Errorf("want foo.js, got:\n%s", out)
	}
	if !strings.Contains(out, "bar.js") {
		t.Errorf("want bar.js, got:\n%s", out)
	}
}

func TestFormat_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.OnlyRules = []string{"semi"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "no-unused-vars") {
		t.Errorf("no-unused-vars should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "semi") {
		t.Errorf("semi should appear, got:\n%s", out)
	}
}

func TestFormat_onlyFiles(t *testing.T) {
	cfg := noColors()
	cfg.OnlyFiles = []string{"bar.js"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "foo.js") {
		t.Errorf("foo.js should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "bar.js") {
		t.Errorf("bar.js should appear, got:\n%s", out)
	}
}

func TestFormat_invalidJSON(t *testing.T) {
	cfg := noColors()
	err := format([]byte("not json"), cfg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFormat_emptyFileSkipped(t *testing.T) {
	cfg := noColors()
	input := `[{"filePath":"empty.js","messages":[]}]`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "0 issues") {
		t.Errorf("file with no messages should yield 0 issues, got:\n%s", out)
	}
}
