package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

type stylelintWarning struct {
	Line     int    `json:"line"`
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Text     string `json:"text"`
}

type stylelintFile struct {
	Source      string             `json:"source"`
	ParseErrors []json.RawMessage  `json:"parseErrors"`
	Warnings    []stylelintWarning `json:"warnings"`
}

func main() {
	formatter.RunWithConfig("stylelint", format)
}

func format(data []byte, cfg formatter.Config) error {
	var files []stylelintFile
	if len(strings.TrimSpace(string(data))) == 0 {
		fmt.Println(formatter.Summary(0, 0, 0))
		return nil
	}
	if err := json.Unmarshal(data, &files); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(string(data)))
		return fmt.Errorf("stylelint error (non-JSON output)")
	}

	var allIssues []formatter.Issue
	for _, f := range files {
		source := formatter.ResolvePath(f.Source, cfg)
		for range f.ParseErrors {
			allIssues = append(allIssues, formatter.Issue{File: source, Rule: "parse-error", Severity: "error", Line: 0, Message: "CSS parse error"})
		}
		for _, w := range f.Warnings {
			rule := w.Rule
			if rule == "" {
				rule = "unknown"
			}
			allIssues = append(allIssues, formatter.Issue{
				File:     source,
				Rule:     rule,
				Severity: w.Severity,
				Line:     w.Line,
				Message:  formatter.Truncate(w.Text, cfg.MaxMessageLength),
			})
		}
	}

	return formatter.Render(allIssues, cfg)
}
