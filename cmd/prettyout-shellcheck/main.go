package main

import (
	"encoding/json"
	"fmt"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

type shellcheckIssue struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Level   string `json:"level"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	formatter.RunWithConfig("shellcheck", format)
}

func format(data []byte, cfg formatter.Config) error {
	var issues []shellcheckIssue
	if err := json.Unmarshal(data, &issues); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return formatter.Render(toIssues(issues, cfg), cfg)
}

func toIssues(raw []shellcheckIssue, cfg formatter.Config) []formatter.Issue {
	out := make([]formatter.Issue, 0, len(raw))
	for _, r := range raw {
		out = append(out, formatter.Issue{
			Rule:     scCode(r.Code),
			File:     formatter.ResolvePath(r.File, cfg),
			Line:     r.Line,
			Message:  formatter.Truncate(r.Message, cfg.MaxMessageLength),
			Severity: r.Level,
		})
	}
	return out
}

func scCode(code int) string {
	return fmt.Sprintf("SC%d", code)
}
