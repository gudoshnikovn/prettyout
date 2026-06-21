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

// pylint JSON2 format
const twoFileJSON = `{
  "messages": [
    {"type":"warning","line":10,"path":"foo.py","symbol":"line-too-long","message":"Line too long","messageId":"C0301"},
    {"type":"warning","line":20,"path":"foo.py","symbol":"line-too-long","message":"Line too long","messageId":"C0301"},
    {"type":"error","line":5,"path":"bar.py","symbol":"undefined-variable","message":"Undefined variable","messageId":"E0602"}
  ],
  "statistics": {"score": 7.5}
}`

func TestFormat_clean(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(`{"messages":[],"statistics":{"score":10.0}}`), cfg); err != nil {
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
	if !strings.Contains(out, "C0301") {
		t.Errorf("want C0301, got:\n%s", out)
	}
	if !strings.Contains(out, "E0602") {
		t.Errorf("want E0602, got:\n%s", out)
	}
	if !strings.Contains(out, "3 issues · 2 rules · 2 files") {
		t.Errorf("want summary, got:\n%s", out)
	}
	if !strings.Contains(out, "rated 7.50/10") {
		t.Errorf("want score in output, got:\n%s", out)
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

func TestFormat_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.OnlyRules = []string{"E0602"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "C0301") {
		t.Errorf("C0301 should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "E0602") {
		t.Errorf("E0602 should appear, got:\n%s", out)
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
	if !strings.Contains(out, "line-too-long") {
		t.Errorf("stats: want line-too-long, got:\n%s", out)
	}
}

func TestFormat_deduplicatesLines(t *testing.T) {
	cfg := noColors()
	// Same file, same line, same rule twice → should appear once
	input := `{
    "messages": [
      {"type":"error","line":5,"path":"a.py","symbol":"undefined-variable","message":"Undefined","messageId":"E0602"},
      {"type":"error","line":5,"path":"a.py","symbol":"undefined-variable","message":"Undefined","messageId":"E0602"}
    ],
    "statistics": {"score": 0.0}
  }`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "1 issue") {
		t.Errorf("duplicate line should be deduplicated: want 1 issue, got:\n%s", out)
	}
}

func TestFormat_invalidJSON(t *testing.T) {
	cfg := noColors()
	err := format([]byte("not json"), cfg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
