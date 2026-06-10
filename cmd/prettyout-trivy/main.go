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

	if cfg.GroupBy == "file" {
		return formatByFile(report, cfg)
	}

	allowedRules := make(map[string]bool, len(cfg.OnlyRules))
	for _, r := range cfg.OnlyRules {
		allowedRules[r] = true
	}

	// Collect all vulns grouped by severity, respecting OnlyFiles and OnlyRules
	bySeverity := map[string][]trivyVuln{}
	for _, result := range report.Results {
		target := formatter.ResolvePath(result.Target, cfg)
		if !formatter.MatchesFileFilter(target, cfg.OnlyFiles) {
			continue
		}
		for _, v := range result.Vulnerabilities {
			if len(allowedRules) > 0 && !allowedRules[v.VulnerabilityID] {
				continue
			}
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

func formatByFile(report trivyReport, cfg formatter.Config) error {
	totalVulns := 0
	targets := 0

	for _, result := range report.Results {
		if len(result.Vulnerabilities) == 0 {
			continue
		}
		target := formatter.ResolvePath(result.Target, cfg)
		if !formatter.MatchesFileFilter(target, cfg.OnlyFiles) {
			continue
		}
		targets++
		vulns := result.Vulnerabilities
		totalVulns += len(vulns)

		fmt.Println(target)

		// Sort by severity rank then CVE ID
		sort.Slice(vulns, func(i, j int) bool {
			ri := severityRank(vulns[i].Severity)
			rj := severityRank(vulns[j].Severity)
			if ri != rj {
				return ri < rj
			}
			return vulns[i].VulnerabilityID < vulns[j].VulnerabilityID
		})

		for _, v := range vulns {
			sev := v.Severity
			if sev == "" {
				sev = "UNKNOWN"
			}
			col := trivyColor(sev, cfg.Colors)
			reset := ""
			if cfg.Colors {
				reset = "\033[0m"
			}
			if v.FixedVersion != "" {
				fmt.Printf("  %s%-8s%s %s — %s %s → fix: %s\n", col, sev, reset, v.VulnerabilityID, v.PkgName, v.InstalledVersion, v.FixedVersion)
			} else {
				fmt.Printf("  %s%-8s%s %s — %s %s → no fix available\n", col, sev, reset, v.VulnerabilityID, v.PkgName, v.InstalledVersion)
			}
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	if totalVulns == 0 {
		fmt.Println("No vulnerabilities found.")
		return nil
	}

	fmt.Printf("%d %s · %d %s\n",
		totalVulns, formatter.Plural(totalVulns, "vulnerability", "vulnerabilities"),
		targets, formatter.Plural(targets, "target", "targets"))
	return nil
}
