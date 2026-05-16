package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
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

var severityOrder = []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "UNKNOWN"}

func severityRank(s string) int {
	for i, v := range severityOrder {
		if v == s {
			return i
		}
	}
	return len(severityOrder)
}

func trivyColor(sev string, colors bool) string {
	if !colors {
		return ""
	}
	switch sev {
	case "CRITICAL", "HIGH":
		return "\033[1;31m"
	case "MEDIUM":
		return "\033[1;33m"
	default:
		return "\033[1;34m"
	}
}

func format(data []byte, cfg formatter.Config) error {
	var report trivyReport
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Collect all vulns grouped by severity
	bySeverity := map[string][]trivyVuln{}
	for _, result := range report.Results {
		for _, v := range result.Vulnerabilities {
			sev := v.Severity
			if sev == "" {
				sev = "UNKNOWN"
			}
			bySeverity[sev] = append(bySeverity[sev], v)
		}
	}

	if len(bySeverity) == 0 {
		fmt.Println("No vulnerabilities found.")
		return nil
	}

	// Sort severity groups by rank
	sevs := make([]string, 0, len(bySeverity))
	for s := range bySeverity {
		sevs = append(sevs, s)
	}
	sort.Slice(sevs, func(i, j int) bool {
		return severityRank(sevs[i]) < severityRank(sevs[j])
	})

	totalVulns := 0
	for _, sev := range sevs {
		vulns := bySeverity[sev]
		totalVulns += len(vulns)
		col := trivyColor(sev, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		label := "vulnerabilities"
		if len(vulns) == 1 {
			label = "vulnerability"
		}
		if cfg.Colors {
			fmt.Printf("%s%s%s (%d %s)\n", col, sev, reset, len(vulns), label)
		} else {
			fmt.Printf("%s (%d %s)\n", sev, len(vulns), label)
		}

		// Sort vulns by VulnerabilityID
		sort.Slice(vulns, func(i, j int) bool {
			return vulns[i].VulnerabilityID < vulns[j].VulnerabilityID
		})

		for _, v := range vulns {
			if v.FixedVersion != "" {
				fmt.Printf("  %s — %s %s → fix: %s\n", v.VulnerabilityID, v.PkgName, v.InstalledVersion, v.FixedVersion)
			} else {
				fmt.Printf("  %s — %s %s → no fix available\n", v.VulnerabilityID, v.PkgName, v.InstalledVersion)
			}
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	fmt.Printf("%d vulnerabilities · %d severity levels\n", totalVulns, len(sevs))
	return nil
}
