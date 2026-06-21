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
  "results": [
    {"check_id":"rules.python.dangerous-eval","path":"foo.py","start":{"line":10},"extra":{"message":"use of eval","severity":"ERROR"}},
    {"check_id":"rules.python.dangerous-eval","path":"foo.py","start":{"line":20},"extra":{"message":"use of eval","severity":"ERROR"}},
    {"check_id":"rules.python.hardcoded-secret","path":"bar.py","start":{"line":5},"extra":{"message":"hardcoded secret","severity":"WARNING"}}
  ],
  "errors": []
}`

func TestFormat_clean(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(`{"results":[],"errors":[]}`), cfg); err != nil {
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
	// display shows shortCheckID (last component after last ".")
	if !strings.Contains(out, "dangerous-eval") {
		t.Errorf("want dangerous-eval, got:\n%s", out)
	}
	if !strings.Contains(out, "hardcoded-secret") {
		t.Errorf("want hardcoded-secret, got:\n%s", out)
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
	if !strings.Contains(out, "foo.py") {
		t.Errorf("want foo.py, got:\n%s", out)
	}
	if !strings.Contains(out, "bar.py") {
		t.Errorf("want bar.py, got:\n%s", out)
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
	if !strings.Contains(out, "dangerous-eval") {
		t.Errorf("stats: want dangerous-eval, got:\n%s", out)
	}
}

func TestShortCheckID(t *testing.T) {
	cases := []struct{ in, want string }{
		{"rules.python.dangerous-eval", "dangerous-eval"},
		{"simple", "simple"},
		{"a.b.c.d", "d"},
	}
	for _, c := range cases {
		got := shortCheckID(c.in)
		if got != c.want {
			t.Errorf("shortCheckID(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
