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
  {"code":"E501","message":"line too long","filename":"foo.py","location":{"row":10},"end_location":{"row":10}},
  {"code":"E501","message":"line too long","filename":"foo.py","location":{"row":20},"end_location":{"row":20}},
  {"code":"F401","message":"unused import","filename":"bar.py","location":{"row":5},"end_location":{"row":5}}
]`

func TestFormat_clean(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte("[]"), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "0 issues") {
		t.Errorf("clean run: want 0 issues in output, got:\n%s", out)
	}
}

func TestFormat_byRule(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "E501 (2)") {
		t.Errorf("want E501 (2), got:\n%s", out)
	}
	if !strings.Contains(out, "F401 (1)") {
		t.Errorf("want F401 (1), got:\n%s", out)
	}
	if !strings.Contains(out, "lines 10, 20") {
		t.Errorf("want 'lines 10, 20', got:\n%s", out)
	}
	if !strings.Contains(out, "line 5") {
		t.Errorf("want 'line 5', got:\n%s", out)
	}
	if !strings.Contains(out, "3 issues · 2 rules · 2 files") {
		t.Errorf("want summary '3 issues · 2 rules · 2 files', got:\n%s", out)
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
	if !strings.Contains(out, "bar.py") {
		t.Errorf("want bar.py in output, got:\n%s", out)
	}
	if !strings.Contains(out, "foo.py") {
		t.Errorf("want foo.py in output, got:\n%s", out)
	}
	if !strings.Contains(out, "3 issues · 2 rules · 2 files") {
		t.Errorf("want summary, got:\n%s", out)
	}
}

func TestFormat_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.OnlyRules = []string{"F401"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "E501") {
		t.Errorf("E501 should be filtered out, got:\n%s", out)
	}
	if !strings.Contains(out, "F401") {
		t.Errorf("F401 should appear, got:\n%s", out)
	}
}

func TestFormat_onlyFiles(t *testing.T) {
	cfg := noColors()
	cfg.OnlyFiles = []string{"bar.py"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "foo.py") {
		t.Errorf("foo.py should be filtered out, got:\n%s", out)
	}
	if !strings.Contains(out, "bar.py") {
		t.Errorf("bar.py should appear, got:\n%s", out)
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
	if !strings.Contains(out, "E501") {
		t.Errorf("stats: want E501, got:\n%s", out)
	}
	if !strings.Contains(out, "3 issues") {
		t.Errorf("stats: want summary, got:\n%s", out)
	}
}

func TestFormat_fixHint_safeOnly(t *testing.T) {
	cfg := noColors()
	input := `[{"code":"E501","message":"too long","filename":"a.py","location":{"row":1},"end_location":{"row":1},"fix":{"applicability":"safe"}}]`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "fixable with --fix") {
		t.Errorf("want safe fix hint, got:\n%s", out)
	}
}

func TestFormat_fixHint_unsafeOnly(t *testing.T) {
	cfg := noColors()
	input := `[{"code":"E501","message":"too long","filename":"a.py","location":{"row":1},"end_location":{"row":1},"fix":{"applicability":"unsafe"}}]`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "--unsafe-fixes") {
		t.Errorf("want unsafe fix hint, got:\n%s", out)
	}
}

func TestFormat_fixHint_both(t *testing.T) {
	cfg := noColors()
	input := `[
    {"code":"E501","message":"too long","filename":"a.py","location":{"row":1},"end_location":{"row":1},"fix":{"applicability":"safe"}},
    {"code":"F401","message":"unused","filename":"a.py","location":{"row":2},"end_location":{"row":2},"fix":{"applicability":"unsafe"}}
  ]`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "--fix") || !strings.Contains(out, "--unsafe-fixes") {
		t.Errorf("want both fix hints, got:\n%s", out)
	}
}

func TestFormat_byFile_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.GroupBy = "file"
	cfg.OnlyRules = []string{"E501"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	// bar.py only has F401, should be skipped
	if strings.Contains(out, "bar.py") {
		t.Errorf("byFile+onlyRules: bar.py should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "foo.py") {
		t.Errorf("byFile+onlyRules: foo.py should appear, got:\n%s", out)
	}
}

func TestFormat_singleLineSingular(t *testing.T) {
	cfg := noColors()
	input := `[{"code":"E501","message":"too long","filename":"a.py","location":{"row":5},"end_location":{"row":5}}]`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "lines 5") {
		t.Errorf("single line should say 'line 5' not 'lines 5', got:\n%s", out)
	}
	if !strings.Contains(out, "line 5") {
		t.Errorf("want 'line 5', got:\n%s", out)
	}
}
