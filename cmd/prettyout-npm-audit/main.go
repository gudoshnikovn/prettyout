package main

import (
	"encoding/json"
	"fmt"
	"sort"

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

var npmSeverityOrder = []string{"critical", "high", "moderate", "low", "info"}

func npmSeverityRank(s string) int {
	for i, v := range npmSeverityOrder {
		if v == s {
			return i
		}
	}
	return len(npmSeverityOrder)
}

func npmColor(sev string, colors bool) string {
	if !colors {
		return ""
	}
	switch sev {
	case "critical", "high":
		return "\033[1;31m"
	case "moderate":
		return "\033[1;33m"
	default:
		return "\033[1;34m"
	}
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

	if len(report.Vulnerabilities) == 0 {
		fmt.Println("No vulnerabilities found.")
		return nil
	}

	// Group by severity
	bySeverity := map[string][]npmVuln{}
	for name, v := range report.Vulnerabilities {
		v.Name = name
		sev := v.Severity
		if sev == "" {
			sev = "info"
		}
		bySeverity[sev] = append(bySeverity[sev], v)
	}

	sevs := make([]string, 0, len(bySeverity))
	for s := range bySeverity {
		sevs = append(sevs, s)
	}
	sort.Slice(sevs, func(i, j int) bool {
		return npmSeverityRank(sevs[i]) < npmSeverityRank(sevs[j])
	})

	if cfg.Stats {
		maxCount := 0
		for _, sev := range sevs {
			if c := len(bySeverity[sev]); c > maxCount {
				maxCount = c
			}
		}
		width := len(fmt.Sprintf("%d", maxCount))
		for _, sev := range sevs {
			c := len(bySeverity[sev])
			col := npmColor(sev, cfg.Colors)
			reset := ""
			if cfg.Colors {
				reset = "\033[0m"
			}
			fmt.Printf("%*d  %s%s%s\n", width, c, col, sev, reset)
		}
		fmt.Println()
		fmt.Printf("%d vulnerabilities\n", len(report.Vulnerabilities))
		return nil
	}

	total := len(report.Vulnerabilities)

	for _, sev := range sevs {
		vulns := bySeverity[sev]
		sort.Slice(vulns, func(i, j int) bool { return vulns[i].Name < vulns[j].Name })

		col := npmColor(sev, cfg.Colors)
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

		for _, v := range vulns {
			fix := fixLabel(v.FixAvailable)
			rangeStr := v.Range
			if rangeStr == "" {
				rangeStr = "any"
			}
			fmt.Printf("  %s  range: %s  [%s]\n", v.Name, rangeStr, fix)
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	fmt.Printf("%d vulnerabilities\n", total)
	return nil
}
