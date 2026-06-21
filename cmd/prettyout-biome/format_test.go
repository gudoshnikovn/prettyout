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

const twoFileJSON = `{
  "diagnostics": [
    {"category":"lint/correctness/noUnusedVariables","message":"unused variable","severity":"warning","location":{"path":"foo.js"}},
    {"category":"lint/correctness/noUnusedVariables","message":"unused variable","severity":"warning","location":{"path":"foo.js"}},
    {"category":"lint/style/noVar","message":"use let/const","severity":"error","location":{"path":"bar.js"}}
  ]
}`

func TestFormat_clean(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(`{"diagnostics":[]}`), cfg); err != nil {
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
	if !strings.Contains(out, "noUnusedVariables") {
		t.Errorf("want noUnusedVariables, got:\n%s", out)
	}
	if !strings.Contains(out, "noVar") {
		t.Errorf("want noVar, got:\n%s", out)
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
	cfg.OnlyRules = []string{"lint/style/noVar"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "noUnusedVariables") {
		t.Errorf("noUnusedVariables should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "noVar") {
		t.Errorf("noVar should appear, got:\n%s", out)
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

func TestFormat_statsMode(t *testing.T) {
	cfg := noColors()
	cfg.Stats = true
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "noUnusedVariables") {
		t.Errorf("stats: want noUnusedVariables, got:\n%s", out)
	}
}

func TestFormat_invalidJSON(t *testing.T) {
	cfg := noColors()
	err := format([]byte("not json"), cfg)
	if err == nil {
		t.Error("expected error for invalid JSON")
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
	// formatByFile uses full category name (not shortCategoryID)
	cfg.OnlyRules = []string{"lint/correctness/noUnusedVariables"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	// bar.js only has lint/style/noVar → filtered out
	if strings.Contains(out, "bar.js") {
		t.Errorf("byFile+onlyRules: bar.js should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "foo.js") {
		t.Errorf("byFile+onlyRules: foo.js should appear, got:\n%s", out)
	}
}
