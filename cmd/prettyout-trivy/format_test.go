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

func TestTrivyColor_withColors(t *testing.T) {
	if trivyColor("CRITICAL", true) == "" {
		t.Error("trivyColor(CRITICAL, true): want ANSI code")
	}
	if trivyColor("MEDIUM", true) == "" {
		t.Error("trivyColor(MEDIUM, true): want ANSI code")
	}
	if trivyColor("LOW", true) == "" {
		t.Error("trivyColor(LOW, true): want ANSI code (default blue)")
	}
	if trivyColor("CRITICAL", false) != "" {
		t.Error("trivyColor(CRITICAL, false): want empty")
	}
}

func TestFormat_byFile_onlyFiles_noMatch(t *testing.T) {
	cfg := noColors()
	cfg.GroupBy = "file"
	cfg.OnlyFiles = []string{"other.json"}
	out := captureOutput(func() {
		if err := format([]byte(trivyJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "CVE-2023-001") {
		t.Errorf("byFile+onlyFiles: CVEs should be filtered, got:\n%s", out)
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

func TestSeverityRank_default(t *testing.T) {
	// LOW is at index 3 in severityOrder
	if got := severityRank("LOW"); got != 3 {
		t.Errorf("severityRank(LOW) = %d, want 3", got)
	}
	// UNKNOWN is at index 4 in severityOrder
	if got := severityRank("UNKNOWN"); got != 4 {
		t.Errorf("severityRank(UNKNOWN) = %d, want 4", got)
	}
	// A totally unknown value returns len(severityOrder) = 5
	if got := severityRank("NOTFOUND"); got != 5 {
		t.Errorf("severityRank(NOTFOUND) = %d, want 5", got)
	}
}

func TestFormat_withColors(t *testing.T) {
	cfg := formatter.DefaultConfig()
	cfg.Colors = true
	out := captureOutput(func() {
		if err := format([]byte(trivyJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "\033[") {
		t.Errorf("withColors: want ANSI codes in output, got:\n%s", out)
	}
}

func TestFormat_byFile_withColors(t *testing.T) {
	cfg := formatter.DefaultConfig()
	cfg.Colors = true
	cfg.GroupBy = "file"
	out := captureOutput(func() {
		if err := format([]byte(trivyJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "\033[") {
		t.Errorf("byFile withColors: want ANSI codes in output, got:\n%s", out)
	}
}

func TestFormat_noFixAvailable(t *testing.T) {
	// FixedVersion=="" triggers the "no fix available" branch in both byRule and byFile
	cfg := noColors()
	input := `{
  "Results": [
    {
      "Target": "package.json",
      "Vulnerabilities": [
        {"VulnerabilityID":"CVE-2023-999","PkgName":"moment","InstalledVersion":"2.29.0","FixedVersion":"","Severity":"HIGH"}
      ]
    }
  ]
}`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "no fix available") {
		t.Errorf("no fix: want 'no fix available', got:\n%s", out)
	}
}

func TestFormat_byFile_noFixAvailable(t *testing.T) {
	cfg := noColors()
	cfg.GroupBy = "file"
	input := `{
  "Results": [
    {
      "Target": "package.json",
      "Vulnerabilities": [
        {"VulnerabilityID":"CVE-2023-999","PkgName":"moment","InstalledVersion":"2.29.0","FixedVersion":"","Severity":"HIGH"}
      ]
    }
  ]
}`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "no fix available") {
		t.Errorf("byFile no fix: want 'no fix available', got:\n%s", out)
	}
}

func TestFormat_onlyRules(t *testing.T) {
	// Tests the OnlyRules filter (allowedRules map) in format
	cfg := noColors()
	cfg.OnlyRules = []string{"CVE-2023-001"}
	out := captureOutput(func() {
		if err := format([]byte(trivyJSON), cfg); err != nil {
			t.Error(err)
		}
	})
	if strings.Contains(out, "CVE-2023-002") {
		t.Errorf("CVE-2023-002 should be filtered by OnlyRules, got:\n%s", out)
	}
	if !strings.Contains(out, "CVE-2023-001") {
		t.Errorf("CVE-2023-001 should appear, got:\n%s", out)
	}
}

func TestFormat_emptySeverity(t *testing.T) {
	// Empty Severity → defaults to "UNKNOWN" branch
	cfg := noColors()
	input := `{
  "Results": [
    {
      "Target": "pkg.json",
      "Vulnerabilities": [
        {"VulnerabilityID":"CVE-2023-888","PkgName":"pkg","InstalledVersion":"1.0","FixedVersion":"","Severity":""}
      ]
    }
  ]
}`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "UNKNOWN") {
		t.Errorf("empty severity: want 'UNKNOWN', got:\n%s", out)
	}
}

func TestFormat_byFile_emptySeverityAndEmptyVulns(t *testing.T) {
	// Covers: empty Severity→UNKNOWN in formatByFile, len(Vulns)==0 skip
	cfg := noColors()
	cfg.GroupBy = "file"
	input := `{
  "Results": [
    {
      "Target": "empty.json",
      "Vulnerabilities": []
    },
    {
      "Target": "pkg.json",
      "Vulnerabilities": [
        {"VulnerabilityID":"CVE-2023-777","PkgName":"pkg","InstalledVersion":"1.0","FixedVersion":"","Severity":""}
      ]
    }
  ]
}`
	out := captureOutput(func() {
		if err := format([]byte(input), cfg); err != nil {
			t.Error(err)
		}
	})
	if !strings.Contains(out, "UNKNOWN") {
		t.Errorf("empty severity in byFile: want 'UNKNOWN', got:\n%s", out)
	}
}
