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

func TestFormat_statsMode(t *testing.T) {
	cfg := noColors()
	cfg.Stats = true
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "no-unused-vars") {
		t.Errorf("stats: want no-unused-vars, got:\n%s", out)
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

func TestRuleId_nil(t *testing.T) {
	m := eslintMessage{RuleId: nil, Severity: 2, Message: "Parsing error", Line: 1}
	if got := ruleId(m); got != "parse-error" {
		t.Errorf("ruleId(nil) = %q, want parse-error", got)
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
	cfg.OnlyRules = []string{"no-unused-vars"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	// bar.js only has "semi" → filtered out
	if strings.Contains(out, "bar.js") {
		t.Errorf("byFile+onlyRules: bar.js should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "foo.js") {
		t.Errorf("byFile+onlyRules: foo.js should appear, got:\n%s", out)
	}
}

func TestFormat_parseError(t *testing.T) {
	// Message with null ruleId (parse error) goes through the ruleId==nil path
	cfg := noColors()
	input := `[{"filePath":"a.js","messages":[{"ruleId":null,"severity":2,"message":"Parsing error: unexpected token","line":1}]}]`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "parse-error") {
		t.Errorf("null ruleId: want 'parse-error' rule, got:\n%s", out)
	}
}
