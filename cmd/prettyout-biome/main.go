package main

import (
	"encoding/json"
	"fmt"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

type biomeLocation struct {
	Path string `json:"path"`
}

type biomeDiagnostic struct {
	Category string        `json:"category"`
	Message  string        `json:"message"`
	Severity string        `json:"severity"`
	Location biomeLocation `json:"location"`
}

type biomeReport struct {
	Diagnostics []biomeDiagnostic `json:"diagnostics"`
}

func main() {
	formatter.RunWithConfig("biome", format)
}

func format(data []byte, cfg formatter.Config) error {
	var report biomeReport
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return formatter.Render(toIssues(report.Diagnostics, cfg), cfg)
}

func toIssues(diags []biomeDiagnostic, cfg formatter.Config) []formatter.Issue {
	out := make([]formatter.Issue, 0, len(diags))
	for _, d := range diags {
		rule := d.Category
		if rule == "" {
			rule = "unknown"
		}
		file := d.Location.Path
		if file == "" {
			file = "(unknown)"
		} else {
			file = formatter.ResolvePath(file, cfg)
		}
		out = append(out, formatter.Issue{
			Rule:     rule,
			File:     file,
			Message:  formatter.Truncate(d.Message, cfg.MaxMessageLength),
			Severity: d.Severity,
		})
	}
	return out
}
