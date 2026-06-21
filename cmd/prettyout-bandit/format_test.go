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
  "errors": [],
  "results": [
    {"filename":"foo.py","line_number":10,"test_id":"B101","issue_text":"Use of assert","issue_severity":"LOW","issue_confidence":"HIGH"},
    {"filename":"foo.py","line_number":20,"test_id":"B101","issue_text":"Use of assert","issue_severity":"LOW","issue_confidence":"HIGH"},
    {"filename":"bar.py","line_number":5,"test_id":"B602","issue_text":"subprocess with shell","issue_severity":"HIGH","issue_confidence":"MEDIUM"}
  ]
}`

func TestFormat_clean(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(`{"errors":[],"results":[]}`), cfg); err != nil {
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
	if !strings.Contains(out, "B101") {
		t.Errorf("want B101, got:\n%s", out)
	}
	if !strings.Contains(out, "B602") {
		t.Errorf("want B602, got:\n%s", out)
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

func TestFormat_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.OnlyRules = []string{"B602"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "B101") {
		t.Errorf("B101 should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "B602") {
		t.Errorf("B602 should appear, got:\n%s", out)
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
		t.Errorf("foo.py should be filtered, got:\n%s", out)
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
	if !strings.Contains(out, "B101") {
		t.Errorf("stats: want B101, got:\n%s", out)
	}
}

func TestSeverityColor_withColors(t *testing.T) {
	if severityColor("HIGH", true) == "" {
		t.Error("severityColor(HIGH, true): want ANSI code")
	}
	if severityColor("MEDIUM", true) == "" {
		t.Error("severityColor(MEDIUM, true): want ANSI code")
	}
	if severityColor("HIGH", false) != "" {
		t.Error("severityColor(HIGH, false): want empty")
	}
}

func TestFormat_byFile_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.GroupBy = "file"
	cfg.OnlyRules = []string{"B101"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	// bar.py only has B602, should be skipped
	if strings.Contains(out, "bar.py") {
		t.Errorf("byFile+onlyRules: bar.py should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "foo.py") {
		t.Errorf("byFile+onlyRules: foo.py should appear, got:\n%s", out)
	}
}

func TestCleanFilename(t *testing.T) {
	if got := cleanFilename("./foo.py"); got != "foo.py" {
		t.Errorf("got %q, want %q", got, "foo.py")
	}
	if got := cleanFilename("bar.py"); got != "bar.py" {
		t.Errorf("got %q, want %q", got, "bar.py")
	}
}

func TestSeverityColor_low(t *testing.T) {
	// LOW/unknown goes to default branch → returns ""
	if got := severityColor("LOW", true); got != "" {
		t.Errorf("severityColor(LOW, true) = %q, want empty", got)
	}
}

func TestFormat_withErrors(t *testing.T) {
	// Non-empty errors slice triggers the stderr warning branch
	cfg := noColors()
	input := `{
  "errors": [{"filename":"bad.py","reason":"syntax error"}],
  "results": []
}`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "0 issues") {
		t.Errorf("withErrors: want 0 issues in stdout, got:\n%s", out)
	}
}
