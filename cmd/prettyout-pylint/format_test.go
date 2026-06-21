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
	if !strings.Contains(out, "C0301") {
		t.Errorf("stats: want C0301, got:\n%s", out)
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

func TestMsgID_json2Field(t *testing.T) {
	// MessageIDJson2 ("messageId" key) takes priority.
	m := pylintMsg{MessageIDJson2: "C0301", MessageID: ""}
	if got := m.msgID(); got != "C0301" {
		t.Errorf("msgID json2 = %q, want C0301", got)
	}
}

func TestMsgID_fallbackToMessageID(t *testing.T) {
	// When MessageIDJson2 is empty, fall back to MessageID ("message-id" key).
	m := pylintMsg{MessageIDJson2: "", MessageID: "W0611"}
	if got := m.msgID(); got != "W0611" {
		t.Errorf("msgID fallback = %q, want W0611", got)
	}
}

func TestPylintSeverity_info(t *testing.T) {
	if got := pylintSeverity("convention"); got != "info" {
		t.Errorf("pylintSeverity(convention) = %q, want info", got)
	}
	if got := pylintSeverity("refactor"); got != "info" {
		t.Errorf("pylintSeverity(refactor) = %q, want info", got)
	}
}



func TestRuleDisplay_withSymbol(t *testing.T) {
	m := pylintMsg{Symbol: "line-too-long", MessageIDJson2: "C0301"}
	if got := ruleDisplay(m); !strings.Contains(got, "C0301") || !strings.Contains(got, "line-too-long") {
		t.Errorf("ruleDisplay with symbol = %q", got)
	}
}

func TestRuleDisplay_noSymbol(t *testing.T) {
	m := pylintMsg{Symbol: "", MessageIDJson2: "C0301"}
	if got := ruleDisplay(m); got != "C0301" {
		t.Errorf("ruleDisplay no symbol = %q, want 'C0301'", got)
	}
}

func TestFormat_byFile_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.GroupBy = "file"
	// generic engine uses Rule (msgID) for filtering
	cfg.OnlyRules = []string{"C0301"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	// bar.py only has E0602/undefined-variable → skipped
	if strings.Contains(out, "bar.py") {
		t.Errorf("byFile+onlyRules: bar.py should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "foo.py") {
		t.Errorf("byFile+onlyRules: foo.py should appear, got:\n%s", out)
	}
}
