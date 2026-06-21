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

const trivyJSON = `{
  "Results": [
    {
      "Target": "package-lock.json",
      "Vulnerabilities": [
        {"VulnerabilityID":"CVE-2023-001","PkgName":"lodash","InstalledVersion":"4.17.0","FixedVersion":"4.17.21","Severity":"HIGH"},
        {"VulnerabilityID":"CVE-2023-002","PkgName":"axios","InstalledVersion":"0.21.0","FixedVersion":"0.21.4","Severity":"MEDIUM"}
      ]
    }
  ]
}`

func TestFormat_clean(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(`{"Results":[]}`), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "No vulnerabilities found") {
		t.Errorf("clean run: want 'No vulnerabilities found', got:\n%s", out)
	}
}

func TestFormat_bySeverity(t *testing.T) {
	cfg := noColors()
	out := captureOutput(func() {
		if err := format([]byte(trivyJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "CVE-2023-001") {
		t.Errorf("want CVE-2023-001, got:\n%s", out)
	}
	if !strings.Contains(out, "CVE-2023-002") {
		t.Errorf("want CVE-2023-002, got:\n%s", out)
	}
	// HIGH should appear before MEDIUM
	hiIdx := strings.Index(out, "HIGH")
	medIdx := strings.Index(out, "MEDIUM")
	if hiIdx < 0 || medIdx < 0 || hiIdx > medIdx {
		t.Errorf("HIGH should appear before MEDIUM, got:\n%s", out)
	}
}

func TestFormat_byFile(t *testing.T) {
	cfg := noColors()
	cfg.GroupBy = "file"
	out := captureOutput(func() {
		if err := format([]byte(trivyJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "package-lock.json") {
		t.Errorf("want package-lock.json, got:\n%s", out)
	}
	if !strings.Contains(out, "CVE-2023-001") {
		t.Errorf("want CVE-2023-001, got:\n%s", out)
	}
}

func TestFormat_onlyFiles(t *testing.T) {
	cfg := noColors()
	cfg.OnlyFiles = []string{"other-file.json"}
	out := captureOutput(func() {
		if err := format([]byte(trivyJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "CVE-2023-001") {
		t.Errorf("vulns should be filtered when file doesn't match, got:\n%s", out)
	}
	if !strings.Contains(out, "No vulnerabilities found") {
		t.Errorf("want 'No vulnerabilities found' after filtering, got:\n%s", out)
	}
}

func TestFormat_statsMode(t *testing.T) {
	cfg := noColors()
	cfg.Stats = true
	out := captureOutput(func() {
		if err := format([]byte(trivyJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "HIGH") {
		t.Errorf("stats: want HIGH, got:\n%s", out)
	}
	if !strings.Contains(out, "vulnerabilit") {
		t.Errorf("stats: want vulnerability summary, got:\n%s", out)
	}
}

func TestFormat_invalidJSON(t *testing.T) {
	cfg := noColors()
	err := format([]byte("not json"), cfg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestSeverityRank(t *testing.T) {
	if severityRank("CRITICAL") >= severityRank("HIGH") {
		t.Error("CRITICAL should rank before HIGH")
	}
	if severityRank("HIGH") >= severityRank("MEDIUM") {
		t.Error("HIGH should rank before MEDIUM")
	}
	if severityRank("MEDIUM") >= severityRank("LOW") {
		t.Error("MEDIUM should rank before LOW")
	}
}
