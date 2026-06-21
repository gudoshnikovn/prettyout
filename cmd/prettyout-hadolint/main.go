package main

import (
	"encoding/json"
	"fmt"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

type hadolintIssue struct {
	Line    int    `json:"line"`
	Level   string `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
	File    string `json:"file"`
}

func main() {
	formatter.RunWithConfig("hadolint", format)
}

func format(data []byte, cfg formatter.Config) error {
	var issues []hadolintIssue
	if err := json.Unmarshal(data, &issues); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Filter out "ignore" level items
	filtered := issues[:0]
	for _, iss := range issues {
		if iss.Level != "ignore" {
			filtered = append(filtered, iss)
		}
	}
	issues = filtered

	return formatter.Render(toIssues(issues, cfg), cfg)
}

func toIssues(raw []hadolintIssue, cfg formatter.Config) []formatter.Issue {
	out := make([]formatter.Issue, 0, len(raw))
	for _, r := range raw {
		out = append(out, formatter.Issue{
			Rule:     r.Code,
			File:     formatter.ResolvePath(r.File, cfg),
			Line:     r.Line,
			Message:  formatter.Truncate(r.Message, cfg.MaxMessageLength),
			Severity: r.Level,
		})
	}
	return out
}
