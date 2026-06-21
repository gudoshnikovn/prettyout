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

const twoFileNDJSON = `{"file":"foo.py","line":10,"severity":"error","message":"Incompatible types","code":"assignment"}
{"file":"foo.py","line":20,"severity":"error","message":"Incompatible types","code":"assignment"}
{"file":"bar.py","line":5,"severity":"error","message":"Cannot find implementation","code":"attr-defined"}
`

func TestFormat_clean(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(""), cfg); err != nil {
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
		if err := format([]byte(twoFileNDJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "assignment") {
		t.Errorf("want 'assignment' rule, got:\n%s", out)
	}
	if !strings.Contains(out, "attr-defined") {
		t.Errorf("want 'attr-defined' rule, got:\n%s", out)
	}
	if !strings.Contains(out, "3 issues") {
		t.Errorf("want 3 issues in summary, got:\n%s", out)
	}
}

func TestFormat_byFile(t *testing.T) {
	cfg := noColors()
	cfg.GroupBy = "file"
	out := captureOutput(func() {
		if err := format([]byte(twoFileNDJSON), cfg); err != nil {
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

func TestFormat_skipsNotes(t *testing.T) {
	cfg := noColors()
	input := `{"file":"a.py","line":1,"severity":"note","message":"Note here","code":"note"}
{"file":"a.py","line":2,"severity":"error","message":"Real error","code":"misc"}
`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "1 issue") {
		t.Errorf("note should be skipped, want 1 issue, got:\n%s", out)
	}
}

func TestFormat_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.OnlyRules = []string{"attr-defined"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileNDJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "assignment") {
		t.Errorf("assignment should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "attr-defined") {
		t.Errorf("attr-defined should appear, got:\n%s", out)
	}
}

func TestFormat_statsMode(t *testing.T) {
	cfg := noColors()
	cfg.Stats = true
	out := captureOutput(func() {
		if err := format([]byte(twoFileNDJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "assignment") {
		t.Errorf("stats: want assignment rule, got:\n%s", out)
	}
}

func TestCodeStr_fallback(t *testing.T) {
	// When Code is nil, codeStr should return "error".
	m := mypyMsg{Code: nil}
	if got := codeStr(m); got != "error" {
		t.Errorf("codeStr(nil) = %q, want error", got)
	}
}

func TestCodeStr_emptyString(t *testing.T) {
	empty := ""
	m := mypyMsg{Code: &empty}
	if got := codeStr(m); got != "error" {
		t.Errorf("codeStr(empty) = %q, want error", got)
	}
}

func TestFormat_byFile_onlyRules(t *testing.T) {
	cfg := noColors()
	cfg.GroupBy = "file"
	cfg.OnlyRules = []string{"assignment"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileNDJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	// bar.py only has attr-defined, should be skipped
	if strings.Contains(out, "bar.py") {
		t.Errorf("byFile+onlyRules: bar.py should be filtered, got:\n%s", out)
	}
	if !strings.Contains(out, "foo.py") {
		t.Errorf("byFile+onlyRules: foo.py should appear, got:\n%s", out)
	}
}

func TestFormat_onlyFiles(t *testing.T) {
	cfg := noColors()
	cfg.OnlyFiles = []string{"bar.py"}
	out := captureOutput(func() {
		if err := format([]byte(twoFileNDJSON), cfg); err != nil {
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
