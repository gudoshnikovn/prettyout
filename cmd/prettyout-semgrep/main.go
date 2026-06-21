package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

type semgrepStart struct {
	Line int `json:"line"`
}

type semgrepExtra struct {
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

type semgrepResult struct {
	CheckID string       `json:"check_id"`
	Path    string       `json:"path"`
	Start   semgrepStart `json:"start"`
	Extra   semgrepExtra `json:"extra"`
}

type semgrepReport struct {
	Results []semgrepResult   `json:"results"`
	Errors  []json.RawMessage `json:"errors"`
}

func main() {
	formatter.RunWithConfig("semgrep", format)
}

func format(data []byte, cfg formatter.Config) error {
	var report semgrepReport
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if len(report.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "semgrep: %d error(s) encountered during scan\n", len(report.Errors))
	}

	return formatter.Render(toIssues(report.Results, cfg), cfg)
}

func toIssues(raw []semgrepResult, cfg formatter.Config) []formatter.Issue {
	out := make([]formatter.Issue, 0, len(raw))
	for _, r := range raw {
		out = append(out, formatter.Issue{
			Rule:     r.CheckID,
			Display:  shortCheckID(r.CheckID),
			File:     formatter.ResolvePath(r.Path, cfg),
			Line:     r.Start.Line,
			Message:  formatter.Truncate(r.Extra.Message, cfg.MaxMessageLength),
			Severity: r.Extra.Severity,
		})
	}
	return out
}

func shortCheckID(checkID string) string {
	parts := strings.Split(checkID, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return checkID
}
