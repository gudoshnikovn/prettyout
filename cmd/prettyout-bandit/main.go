package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

type banditResult struct {
	Filename        string `json:"filename"`
	LineNumber      int    `json:"line_number"`
	TestID          string `json:"test_id"`
	IssueText       string `json:"issue_text"`
	IssueSeverity   string `json:"issue_severity"`
	IssueConfidence string `json:"issue_confidence"`
}

type banditReport struct {
	Errors  []json.RawMessage `json:"errors"`
	Results []banditResult    `json:"results"`
}

type ruleEntry struct {
	severity   string
	confidence string
	message    string
	fileLines  map[string][]int
}

func main() {
	formatter.RunWithConfig("bandit", format)
}

func format(data []byte, cfg formatter.Config) error {
	var report banditReport
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if len(report.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "bandit: %d file(s) could not be parsed\n", len(report.Errors))
	}

	return formatter.Render(toIssues(report.Results, cfg), cfg)
}

func toIssues(raw []banditResult, cfg formatter.Config) []formatter.Issue {
	out := make([]formatter.Issue, 0, len(raw))
	for _, r := range raw {
		out = append(out, formatter.Issue{
			Rule:     r.TestID,
			File:     formatter.ResolvePath(cleanFilename(r.Filename), cfg),
			Line:     r.LineNumber,
			Message:  formatter.Truncate(r.IssueText, cfg.MaxMessageLength),
			Severity: r.IssueSeverity,
			Note:     fmt.Sprintf("%s/%s", r.IssueSeverity, r.IssueConfidence),
		})
	}
	return out
}

func cleanFilename(f string) string {
	return strings.TrimPrefix(f, "./")
}
