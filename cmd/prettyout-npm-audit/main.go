package main

import (
	"encoding/json"
	"fmt"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

type npmVuln struct {
	Name         string          `json:"name"`
	Severity     string          `json:"severity"`
	Range        string          `json:"range"`
	FixAvailable json.RawMessage `json:"fixAvailable"`
}

type npmReport struct {
	Vulnerabilities map[string]npmVuln `json:"vulnerabilities"`
}

func main() {
	formatter.RunWithConfig("npm-audit", format)
}

func fixLabel(raw json.RawMessage) string {
	if raw == nil {
		return "no fix"
	}
	// Could be bool or object
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		if b {
			return "fix available"
		}
		return "no fix"
	}
	// object form: {name, version, isSemVerMajor}
	var obj struct {
		IsSemVerMajor bool `json:"isSemVerMajor"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil {
		if obj.IsSemVerMajor {
			return "breaking fix"
		}
		return "fix available"
	}
	return "no fix"
}


func format(data []byte, cfg formatter.Config) error {
	var report npmReport
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return formatter.Render(toIssues(report, cfg), cfg)
}

func toIssues(report npmReport, cfg formatter.Config) []formatter.Issue {
	var out []formatter.Issue
	for name, v := range report.Vulnerabilities {
		v.Name = name
		sev := v.Severity
		if sev == "" {
			sev = "info"
		}

		fix := fixLabel(v.FixAvailable)
		rangeStr := v.Range
		if rangeStr == "" {
			rangeStr = "any"
		}

		msg := fmt.Sprintf("range: %s [%s]", rangeStr, fix)
		out = append(out, formatter.Issue{
			Rule:     v.Name, // use dependency name as rule
			File:     formatter.ResolvePath("package.json", cfg),
			Message:  msg,
			Severity: sev,
		})
	}
	return out
}
