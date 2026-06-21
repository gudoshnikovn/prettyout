package main

import (
	"encoding/json"
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

const npmJSON = `{
  "vulnerabilities": {
    "lodash": {"name":"lodash","severity":"high","range":"<4.17.21","fixAvailable":true},
    "axios": {"name":"axios","severity":"moderate","range":"<0.21.4","fixAvailable":false},
    "moment": {"name":"moment","severity":"critical","range":"<2.29.4","fixAvailable":{"isSemVerMajor":true}}
  }
}`

func TestFormat_clean(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(`{"vulnerabilities":{}}`), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "0 issues") {
		t.Errorf("clean run: want 0 issues, got:\n%s", out)
	}
}

func TestFormat_groupsBySeverity(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(npmJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "lodash") {
		t.Errorf("want lodash, got:\n%s", out)
	}
	if !strings.Contains(out, "axios") {
		t.Errorf("want axios, got:\n%s", out)
	}
	if !strings.Contains(out, "moment") {
		t.Errorf("want moment, got:\n%s", out)
	}
	if !strings.Contains(out, "[ERROR] lodash") {
		t.Errorf("want lodash, got:\n%s", out)
	}
	if !strings.Contains(out, "[WARN] axios") {
		t.Errorf("want axios, got:\n%s", out)
	}
	if !strings.Contains(out, "[ERROR] moment") {
		t.Errorf("want moment, got:\n%s", out)
	}
}

func TestFormat_fixLabels(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(npmJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "fix available") {
		t.Errorf("want 'fix available' for lodash, got:\n%s", out)
	}
	if !strings.Contains(out, "no fix") {
		t.Errorf("want 'no fix' for axios, got:\n%s", out)
	}
	if !strings.Contains(out, "breaking fix") {
		t.Errorf("want 'breaking fix' for moment, got:\n%s", out)
	}
}

func TestFormat_invalidJSON(t *testing.T) {
	cfg := noColors()
	err := format([]byte("not json"), cfg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFixLabel_nil(t *testing.T) {
	if got := fixLabel(nil); got != "no fix" {
		t.Errorf("fixLabel(nil) = %q, want 'no fix'", got)
	}
}

func TestFixLabel_objectFalse(t *testing.T) {
	raw := []byte(`{"isSemVerMajor":false}`)
	if got := fixLabel(raw); got != "fix available" {
		t.Errorf("fixLabel(isSemVerMajor:false) = %q, want 'fix available'", got)
	}
}

func TestFormat_withColors(t *testing.T) {
	cfg := formatter.DefaultConfig()
	cfg.Colors = true
	out := captureOutput(func() {
		if err := format([]byte(npmJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "\033[") {
		t.Errorf("colors=true: want ANSI codes, got:\n%s", out)
	}
}

func TestFormat_statsMode(t *testing.T) {
	cfg := noColors()
	cfg.Stats = true
	out := captureOutput(func() {
		if err := format([]byte(npmJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "COUNT") || !strings.Contains(out, "RULE") {
		t.Errorf("stats: want stats table header, got:\n%s", out)
	}
	if !strings.Contains(out, "3 issues") {
		t.Errorf("stats: want 3 issues summary, got:\n%s", out)
	}
}

func TestFormat_emptySeverity(t *testing.T) {
	// Vulnerability with no severity field → defaults to "info".
	input := `{"vulnerabilities":{"pkg":{"name":"pkg","severity":"","range":"*","fixAvailable":false}}}`
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "[INFO]") {
		t.Errorf("empty severity: want '[INFO]', got:\n%s", out)
	}
}
func TestFixLabel_invalidRaw(t *testing.T) {
	// A JSON value that's neither bool nor object (e.g., number) → falls through to "no fix"
	raw := json.RawMessage("42")
	if got := fixLabel(raw); got != "no fix" {
		t.Errorf("fixLabel(42) = %q, want 'no fix'", got)
	}
}
