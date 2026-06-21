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
	if !strings.Contains(out, "No vulnerabilities found") {
		t.Errorf("clean run: want 'No vulnerabilities found', got:\n%s", out)
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
	// critical should appear before high, high before moderate
	critIdx := strings.Index(out, "critical")
	hiIdx := strings.Index(out, "high")
	modIdx := strings.Index(out, "moderate")
	if critIdx < 0 || hiIdx < 0 || modIdx < 0 {
		t.Fatalf("missing severity labels in:\n%s", out)
	}
	if critIdx > hiIdx || hiIdx > modIdx {
		t.Errorf("severity order wrong, got:\n%s", out)
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

func TestNpmSeverityRank(t *testing.T) {
	if npmSeverityRank("critical") >= npmSeverityRank("high") {
		t.Error("critical should rank before high")
	}
	if npmSeverityRank("high") >= npmSeverityRank("moderate") {
		t.Error("high should rank before moderate")
	}
}
