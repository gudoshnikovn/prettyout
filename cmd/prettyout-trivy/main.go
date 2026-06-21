package main

import (
	"encoding/json"
	"fmt"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

type trivyVuln struct {
	VulnerabilityID  string `json:"VulnerabilityID"`
	PkgName          string `json:"PkgName"`
	InstalledVersion string `json:"InstalledVersion"`
	FixedVersion     string `json:"FixedVersion"`
	Severity         string `json:"Severity"`
}

type trivyResult struct {
	Target          string      `json:"Target"`
	Vulnerabilities []trivyVuln `json:"Vulnerabilities"`
}

type trivyReport struct {
	Results []trivyResult `json:"Results"`
}

func main() {
	formatter.RunWithConfig("trivy", format)
}

func format(data []byte, cfg formatter.Config) error {
	var report trivyReport
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return formatter.Render(toIssues(report, cfg), cfg)
}

func toIssues(report trivyReport, cfg formatter.Config) []formatter.Issue {
	var out []formatter.Issue
	for _, result := range report.Results {
		target := formatter.ResolvePath(result.Target, cfg)
		for _, v := range result.Vulnerabilities {
			msg := fmt.Sprintf("%s %s", v.PkgName, v.InstalledVersion)
			if v.FixedVersion != "" {
				msg += " → fix: " + v.FixedVersion
			} else {
				msg += " → no fix available"
			}

			sev := v.Severity
			if sev == "" {
				sev = "UNKNOWN"
			}
			out = append(out, formatter.Issue{
				Rule:     v.VulnerabilityID,
				File:     target,
				Message:  msg,
				Severity: sev,
			})
		}
	}
	return out
}
