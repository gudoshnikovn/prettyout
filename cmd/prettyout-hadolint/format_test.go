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
  {"line":10,"level":"warning","code":"DL3008","message":"Pin versions in apt get install","file":"Dockerfile"},
  {"line":20,"level":"warning","code":"DL3008","message":"Pin versions in apt get install","file":"Dockerfile"},
  {"line":5,"level":"error","code":"DL3000","message":"Use absolute WORKDIR","file":"Dockerfile.dev"}
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
	if !strings.Contains(out, "DL3008") {
		t.Errorf("want DL3008, got:\n%s", out)
	}
	if !strings.Contains(out, "DL3000") {
		t.Errorf("want DL3000, got:\n%s", out)
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
	if !strings.Contains(out, "Dockerfile") {
		t.Errorf("want Dockerfile, got:\n%s", out)
	}
}

func TestFormat_filtersIgnoreLevel(t *testing.T) {
	cfg := noColors()
	input := `[
    {"line":1,"level":"ignore","code":"DL3999","message":"ignored rule","file":"Dockerfile"},
    {"line":2,"level":"warning","code":"DL3008","message":"Pin versions","file":"Dockerfile"}
  ]`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "DL3999") {
		t.Errorf("ignore-level issue should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "DL3008") {
		t.Errorf("DL3008 should appear, got:\n%s", out)
	}
	if !strings.Contains(out, "1 issue") {
		t.Errorf("want 1 issue in summary, got:\n%s", out)
	}
}

func TestFormat_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.OnlyRules = []string{"DL3000"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "DL3008") {
		t.Errorf("DL3008 should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "DL3000") {
		t.Errorf("DL3000 should appear, got:\n%s", out)
	}
}

func TestFormat_invalidJSON(t *testing.T) {
	cfg := noColors()
	err := format([]byte("not json"), cfg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
