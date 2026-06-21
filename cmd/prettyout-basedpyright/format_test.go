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

// range.start.line is 0-indexed; displayed as line+1
const twoFileJSON = `{
  "generalDiagnostics": [
    {"rule":"reportMissingImports","file":"foo.py","severity":"error","message":"Import not found","range":{"start":{"line":9},"end":{"line":9}}},
    {"rule":"reportMissingImports","file":"foo.py","severity":"error","message":"Import not found","range":{"start":{"line":19},"end":{"line":19}}},
    {"rule":"reportUndefinedVariable","file":"bar.py","severity":"warning","message":"Variable undefined","range":{"start":{"line":4},"end":{"line":4}}}
  ]
}`

func TestFormat_clean(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(`{"generalDiagnostics":[]}`), cfg); err != nil {
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
	if !strings.Contains(out, "reportMissingImports") {
		t.Errorf("want reportMissingImports, got:\n%s", out)
	}
	if !strings.Contains(out, "reportUndefinedVariable") {
		t.Errorf("want reportUndefinedVariable, got:\n%s", out)
	}
	if !strings.Contains(out, "3 issues · 2 rules · 2 files") {
		t.Errorf("want summary, got:\n%s", out)
	}
	// line 9 → displayed as 10
	if !strings.Contains(out, "10") {
		t.Errorf("want line 10 (0-indexed 9 + 1), got:\n%s", out)
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
	if !strings.Contains(out, "foo.py") {
		t.Errorf("want foo.py, got:\n%s", out)
	}
	if !strings.Contains(out, "bar.py") {
		t.Errorf("want bar.py, got:\n%s", out)
	}
}

func TestFormat_deduplicatesLines(t *testing.T) {
	cfg := noColors()
	// Same file, same line, same rule twice
	input := `{"generalDiagnostics":[
    {"rule":"reportMissingImports","file":"a.py","severity":"error","message":"Import not found","range":{"start":{"line":4},"end":{"line":4}}},
    {"rule":"reportMissingImports","file":"a.py","severity":"error","message":"Import not found","range":{"start":{"line":4},"end":{"line":4}}}
  ]}`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	// The rule group count should be 1 (deduped), not "lines 5, 5"
	if strings.Contains(out, "lines 5, 5") || strings.Contains(out, "lines 5") {
		t.Errorf("duplicate line should be deduplicated, got:\n%s", out)
	}
	// Rule group header should show (1) not (2)
	if !strings.Contains(out, "reportMissingImports (1)") {
		t.Errorf("deduped count should be 1, got:\n%s", out)
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
	if !strings.Contains(out, "reportMissingImports") {
		t.Errorf("stats: want reportMissingImports, got:\n%s", out)
	}
}

func TestFormat_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.OnlyRules = []string{"reportUndefinedVariable"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "reportMissingImports") {
		t.Errorf("reportMissingImports should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "reportUndefinedVariable") {
		t.Errorf("reportUndefinedVariable should appear, got:\n%s", out)
	}
}

func TestFormat_invalidJSON(t *testing.T) {
	cfg := noColors()
	err := format([]byte("not json"), cfg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFirstLine(t *testing.T) {
	if got := firstLine("first\nsecond"); got != "first" {
		t.Errorf("got %q, want %q", got, "first")
	}
	if got := firstLine("no newline"); got != "no newline" {
		t.Errorf("got %q, want %q", got, "no newline")
	}
}
