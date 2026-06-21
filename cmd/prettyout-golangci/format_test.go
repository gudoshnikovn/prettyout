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
  "Issues": [
    {"FromLinter":"errcheck","Text":"Error return value not checked","Pos":{"Filename":"foo.go","Line":10}},
    {"FromLinter":"errcheck","Text":"Error return value not checked","Pos":{"Filename":"foo.go","Line":20}},
    {"FromLinter":"unused","Text":"Function is unused","Pos":{"Filename":"bar.go","Line":5}}
  ]
}`

func TestFormat_clean(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(""), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "0 issues") {
		t.Errorf("empty stdin: want 0 issues, got:\n%s", out)
	}
}

func TestFormat_whitespaceOnly(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte("   \n  "), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "0 issues") {
		t.Errorf("whitespace-only stdin: want 0 issues, got:\n%s", out)
	}
}

func TestFormat_byRule(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "errcheck") {
		t.Errorf("want errcheck, got:\n%s", out)
	}
	if !strings.Contains(out, "unused") {
		t.Errorf("want unused, got:\n%s", out)
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
	if !strings.Contains(out, "foo.go") {
		t.Errorf("want foo.go, got:\n%s", out)
	}
	if !strings.Contains(out, "bar.go") {
		t.Errorf("want bar.go, got:\n%s", out)
	}
}

func TestFormat_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.OnlyRules = []string{"unused"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "errcheck") {
		t.Errorf("errcheck should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "unused") {
		t.Errorf("unused should appear, got:\n%s", out)
	}
}

func TestFormat_onlyFiles(t *testing.T) {
	cfg := noColors()
	cfg.OnlyFiles = []string{"bar.go"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "foo.go") {
		t.Errorf("foo.go should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "bar.go") {
		t.Errorf("bar.go should appear, got:\n%s", out)
	}
}

func TestFormat_byFile_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.GroupBy = "file"
	cfg.OnlyRules = []string{"errcheck"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	// bar.go only has unused, should be skipped
	if strings.Contains(out, "bar.go") {
		t.Errorf("byFile+onlyRules: bar.go should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "foo.go") {
		t.Errorf("byFile+onlyRules: foo.go should appear, got:\n%s", out)
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
	if !strings.Contains(out, "errcheck") {
		t.Errorf("stats: want errcheck, got:\n%s", out)
	}
}

func TestFormat_invalidJSON(t *testing.T) {
	cfg := noColors()
	err := format([]byte("not json"), cfg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
