package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

type position struct {
	Line int `json:"line"`
}

type diagRange struct {
	Start position `json:"start"`
	End   position `json:"end"`
}

type diagnostic struct {
	Rule     string    `json:"rule"`
	File     string    `json:"file"`
	Severity string    `json:"severity"`
	Message  string    `json:"message"`
	Range    diagRange `json:"range"`
}

type report struct {
	GeneralDiagnostics []diagnostic `json:"generalDiagnostics"`
}

func main() {
	formatter.RunWithConfig("basedpyright", format)
}

// firstLine returns s up to the first newline, stripping multiline messages
// that basedpyright sometimes emits.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func format(data []byte, cfg formatter.Config) error {
	var r report
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return formatter.Render(toIssues(r.GeneralDiagnostics, cfg), cfg)
}

func toIssues(raw []diagnostic, cfg formatter.Config) []formatter.Issue {
	out := make([]formatter.Issue, 0, len(raw))
	for _, r := range raw {
		code := r.Rule
		if code == "" {
			code = "StaticAnalysis"
		}
		out = append(out, formatter.Issue{
			Rule:     code,
			File:     formatter.ResolvePath(r.File, cfg),
			Line:     r.Range.Start.Line + 1,
			Message:  formatter.Truncate(firstLine(r.Message), cfg.MaxMessageLength),
			Severity: r.Severity,
		})
	}
	return out
}
