package main

import (
	"encoding/json"
	"fmt"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

type eslintMessage struct {
	RuleId   *string `json:"ruleId"`
	Severity int     `json:"severity"`
	Message  string  `json:"message"`
	Line     int     `json:"line"`
}

type eslintFile struct {
	FilePath string          `json:"filePath"`
	Messages []eslintMessage `json:"messages"`
}

func main() {
	formatter.RunWithConfig("eslint", format)
}

func format(data []byte, cfg formatter.Config) error {
	var files []eslintFile
	if err := json.Unmarshal(data, &files); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return formatter.Render(toIssues(files, cfg), cfg)
}

func toIssues(files []eslintFile, cfg formatter.Config) []formatter.Issue {
	var out []formatter.Issue
	for _, f := range files {
		if len(f.Messages) == 0 {
			continue
		}
		fp := formatter.ResolvePath(f.FilePath, cfg)
		for _, m := range f.Messages {
			out = append(out, formatter.Issue{
				Rule:     ruleId(m),
				File:     fp,
				Line:     m.Line,
				Message:  formatter.Truncate(m.Message, cfg.MaxMessageLength),
				Severity: severityStr(m.Severity),
			})
		}
	}
	return out
}

func ruleId(m eslintMessage) string {
	if m.RuleId == nil {
		return "parse-error"
	}
	return *m.RuleId
}

func severityStr(s int) string {
	if s == 1 {
		return "warning"
	}
	return "error"
}
