package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

type pos struct {
	Filename string `json:"Filename"`
	Line     int    `json:"Line"`
}

type golangciIssue struct {
	FromLinter string `json:"FromLinter"`
	Text       string `json:"Text"`
	Pos        pos    `json:"Pos"`
}

type golangciReport struct {
	Issues []golangciIssue `json:"Issues"`
}

func main() {
	formatter.RunWithConfig("golangci-lint", format)
}

func format(data []byte, cfg formatter.Config) error {
	if len(strings.TrimSpace(string(data))) == 0 {
		fmt.Println(formatter.Summary(0, 0, 0))
		return nil
	}
	var report golangciReport
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return formatter.Render(toIssues(report.Issues, cfg), cfg)
}

func toIssues(raw []golangciIssue, cfg formatter.Config) []formatter.Issue {
	out := make([]formatter.Issue, 0, len(raw))
	for _, r := range raw {
		rule := r.FromLinter
		if rule == "" {
			rule = "unknown"
		}
		file := r.Pos.Filename
		if file == "" {
			file = "unknown"
		} else {
			file = formatter.ResolvePath(file, cfg)
		}
		out = append(out, formatter.Issue{
			Rule:    rule,
			File:    file,
			Line:    r.Pos.Line,
			Message: formatter.Truncate(r.Text, cfg.MaxMessageLength),
		})
	}
	return out
}
