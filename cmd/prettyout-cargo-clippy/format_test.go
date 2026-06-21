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

// clippyNDJSON returns two compiler-message lines for two files and one non-compiler-message.
const clippyNDJSON = `{"reason":"compiler-message","message":{"level":"warning","message":"unused variable","code":{"code":"unused_variables"},"spans":[{"file_name":"src/foo.rs","line_start":10,"is_primary":true}]}}
{"reason":"compiler-message","message":{"level":"warning","message":"unused variable","code":{"code":"unused_variables"},"spans":[{"file_name":"src/foo.rs","line_start":20,"is_primary":true}]}}
{"reason":"compiler-message","message":{"level":"error","message":"mismatched types","code":{"code":"E0308"},"spans":[{"file_name":"src/bar.rs","line_start":5,"is_primary":true}]}}
{"reason":"build-finished","success":false}
`

func TestFormat_clean(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		// Only a build-finished line, no compiler-messages
		if err := format([]byte(`{"reason":"build-finished","success":true}`+"\n"), cfg); err != nil {
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
		if err := format([]byte(clippyNDJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "unused_variables") {
		t.Errorf("want unused_variables, got:\n%s", out)
	}
	if !strings.Contains(out, "E0308") {
		t.Errorf("want E0308, got:\n%s", out)
	}
	if !strings.Contains(out, "3 issues · 2 rules · 2 files") {
		t.Errorf("want summary, got:\n%s", out)
	}
}

func TestFormat_byFile(t *testing.T) {
	cfg := noColors()
	cfg.GroupBy = "file"
	out := captureOutput(func() {
		if err := format([]byte(clippyNDJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "src/foo.rs") {
		t.Errorf("want src/foo.rs, got:\n%s", out)
	}
	if !strings.Contains(out, "src/bar.rs") {
		t.Errorf("want src/bar.rs, got:\n%s", out)
	}
}

func TestFormat_skipsNoteAndHelp(t *testing.T) {
	cfg := noColors()
	input := `{"reason":"compiler-message","message":{"level":"note","message":"a note","code":null,"spans":[]}}
{"reason":"compiler-message","message":{"level":"help","message":"a help","code":null,"spans":[]}}
{"reason":"compiler-message","message":{"level":"warning","message":"real warning","code":{"code":"W0001"},"spans":[{"file_name":"a.rs","line_start":1,"is_primary":true}]}}
`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "1 issue") {
		t.Errorf("note/help should be skipped, want 1 issue, got:\n%s", out)
	}
}

func TestFormat_statsMode(t *testing.T) {
	cfg := noColors()
	cfg.Stats = true
	out := captureOutput(func() {
		if err := format([]byte(clippyNDJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "unused_variables") {
		t.Errorf("stats: want unused_variables, got:\n%s", out)
	}
}

func TestRuleCode_nil(t *testing.T) {
	cl := clippyLine{Message: clippyMessage{Code: nil}}
	if got := ruleCode(cl); got != "rustc" {
		t.Errorf("ruleCode(nil code) = %q, want rustc", got)
	}
}

func TestPrimarySpan_noSpans(t *testing.T) {
	cl := clippyLine{Message: clippyMessage{Spans: nil}}
	file, line := primarySpan(cl)
	if file != "" || line != 0 {
		t.Errorf("primarySpan(no spans) = (%q, %d), want ('', 0)", file, line)
	}
}

func TestPrimarySpan_fallbackToFirst(t *testing.T) {
	cl := clippyLine{Message: clippyMessage{Spans: []clippySpan{
		{FileName: "a.rs", LineStart: 5, IsPrimary: false},
		{FileName: "b.rs", LineStart: 10, IsPrimary: false},
	}}}
	file, line := primarySpan(cl)
	if file != "a.rs" || line != 5 {
		t.Errorf("primarySpan(no primary) = (%q, %d), want (a.rs, 5)", file, line)
	}
}

func TestFormat_byFile_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.GroupBy = "file"
	cfg.OnlyRules = []string{"E0308"}
	out := captureOutput(func() {
		if err := format([]byte(clippyNDJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	// src/foo.rs only has unused_variables, should be skipped
	if strings.Contains(out, "src/foo.rs") {
		t.Errorf("byFile+onlyRules: foo.rs should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "src/bar.rs") {
		t.Errorf("byFile+onlyRules: bar.rs should appear, got:\n%s", out)
	}
}

func TestFormat_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.OnlyRules = []string{"E0308"}
	out := captureOutput(func() {
		if err := format([]byte(clippyNDJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "unused_variables") {
		t.Errorf("unused_variables should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "E0308") {
		t.Errorf("E0308 should appear, got:\n%s", out)
	}
}
