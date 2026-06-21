package formatter

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestRenderDispatch(t *testing.T) {
	issues := []Issue{{Rule: "A", File: "f1.go", Line: 1}}
	
	// Test GroupBy="" (Rule)
	outRule := captureOutput(func() {
		Render(issues, Config{GroupBy: ""})
	})
	if !bytes.Contains([]byte(outRule), []byte("Affected files:")) {
		t.Errorf("Expected RenderByRule behavior, got:\n%s", outRule)
	}

	// Test GroupBy="file"
	outFile := captureOutput(func() {
		Render(issues, Config{GroupBy: "file"})
	})
	if bytes.Contains([]byte(outFile), []byte("Affected files:")) {
		t.Errorf("Expected RenderByFile behavior, got:\n%s", outFile)
	}
}

func TestRenderByRule(t *testing.T) {
	issues := []Issue{
		{Rule: "R1", File: "f1.go", Line: 1, Message: "msg1", Severity: "error", Note: "HIGH"},
		{Rule: "R1", File: "f1.go", Line: 1, Message: "msg1", Severity: "error", Note: "HIGH"}, // duplicate
		{Rule: "R1", File: "f2.go", Line: 2, Message: "msg1", Severity: "error", Note: "HIGH"},
		{Rule: "R2", File: "f1.go", Line: 3, Message: "msg2"}, // no severity/note
	}

	tests := []struct {
		name    string
		cfg     Config
		issues  []Issue
		wantOut []string
	}{
		{
			name:    "zero issues",
			cfg:     Config{},
			issues:  []Issue{},
			wantOut: []string{"0 issues", "0 rules"},
		},
		{
			name:    "multiple rules, filters",
			cfg:     Config{OnlyRules: []string{"R1"}, OnlyFiles: []string{"f1.go"}},
			issues:  issues,
			wantOut: []string{"[ERROR]", "(HIGH)", "msg1", "f1.go"},
		},
		{
			name:    "stats true",
			cfg:     Config{Stats: true},
			issues:  issues,
			wantOut: []string{"COUNT", "RULE", "FILES", "DESCRIPTION"},
		},
		{
			name:    "colors true",
			cfg:     Config{Colors: true},
			issues:  issues,
			wantOut: []string{"\033["},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureOutput(func() {
				RenderByRule(tt.issues, tt.cfg)
			})
			for _, want := range tt.wantOut {
				if !bytes.Contains([]byte(out), []byte(want)) {
					t.Errorf("output missing %q, got:\n%s", want, out)
				}
			}
		})
	}
}

func TestRenderByFile(t *testing.T) {
	issues := []Issue{
		{Rule: "R1", File: "f1.go", Line: 1, Message: "msg1", Severity: "error", Note: "HIGH"},
		{Rule: "R1", File: "f1.go", Line: 1, Message: "msg1", Severity: "error", Note: "HIGH"}, // duplicate
		{Rule: "R1", File: "f2.go", Line: 2, Message: "msg1", Severity: "error", Note: "HIGH"},
		{Rule: "R2", File: "f1.go", Line: 3, Message: "msg2"},
	}

	tests := []struct {
		name    string
		cfg     Config
		issues  []Issue
		wantOut []string
	}{
		{
			name:    "zero issues",
			cfg:     Config{GroupBy: "file"},
			issues:  []Issue{},
			wantOut: []string{"0 issues", "0 rules"},
		},
		{
			name:    "multiple files, filters",
			cfg:     Config{GroupBy: "file", OnlyRules: []string{"R1"}, OnlyFiles: []string{"f1.go"}},
			issues:  issues,
			wantOut: []string{"f1.go — 1 issue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureOutput(func() {
				RenderByFile(tt.issues, tt.cfg)
			})
			for _, want := range tt.wantOut {
				if !bytes.Contains([]byte(out), []byte(want)) {
					t.Errorf("output missing %q, got:\n%s", want, out)
				}
			}
		})
	}
}
