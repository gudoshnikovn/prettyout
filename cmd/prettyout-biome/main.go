package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
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

type ruleEntry struct {
	severity string
	message  string
	files    map[string]struct{}
}

func severityLabel(sev string) string {
	switch strings.ToLower(sev) {
	case "error", "fatal":
		return "ERROR"
	case "warning":
		return "WARN"
	default:
		return "INFO"
	}
}

func main() {
	formatter.RunWithConfig("biome", format)
}

func format(data []byte, cfg formatter.Config) error {
	var report biomeReport
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if cfg.GroupBy == "file" {
		return formatByFile(report.Diagnostics, cfg)
	}
	return formatByRule(report.Diagnostics, cfg)
}

func formatByRule(diags []biomeDiagnostic, cfg formatter.Config) error {
	rules := map[string]*ruleEntry{}
	ruleOrder := []string{}

	for _, d := range diags {
		rule := d.Category
		if rule == "" {
			rule = "unknown"
		}
		file := d.Location.Path

		if _, ok := rules[rule]; !ok {
			rules[rule] = &ruleEntry{files: map[string]struct{}{}}
			ruleOrder = append(ruleOrder, rule)
		}
		r := rules[rule]
		if r.message == "" {
			r.message = truncate(d.Message, cfg.MaxMessageLength)
			r.severity = d.Severity
		}
		if file != "" {
			r.files[file] = struct{}{}
		}
	}

	sort.Strings(ruleOrder)

	totalFiles := map[string]struct{}{}
	for _, d := range diags {
		if d.Location.Path != "" {
			totalFiles[d.Location.Path] = struct{}{}
		}
	}

	for _, rule := range ruleOrder {
		r := rules[rule]
		count := len(r.files)
		if count == 0 {
			count = 1
		}
		// count occurrences properly
		occurrences := 0
		for _, d := range diags {
			c := d.Category
			if c == "" {
				c = "unknown"
			}
			if c == rule {
				occurrences++
			}
		}

		col := formatter.SeverityColor(r.severity, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		label := severityLabel(r.severity)
		if cfg.Colors {
			fmt.Printf("%s[%s]%s %s (%d) — %s\n", col, label, reset, rule, occurrences, r.message)
		} else {
			fmt.Printf("[%s] %s (%d) — %s\n", label, rule, occurrences, r.message)
		}

		if len(r.files) > 0 {
			fmt.Println("Affected files:")
			files := make([]string, 0, len(r.files))
			for f := range r.files {
				files = append(files, f)
			}
			sort.Strings(files)
			for _, f := range files {
				fmt.Printf("  - %s\n", f)
			}
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	fmt.Println(formatter.Summary(len(diags), len(rules), len(totalFiles)))
	return nil
}

func formatByFile(diags []biomeDiagnostic, cfg formatter.Config) error {
	type entry struct {
		rule    string
		message string
	}
	fileMap := map[string][]entry{}
	fileOrder := []string{}
	allRules := map[string]struct{}{}

	for _, d := range diags {
		file := d.Location.Path
		if file == "" {
			file = "(unknown)"
		}
		rule := d.Category
		if rule == "" {
			rule = "unknown"
		}
		allRules[rule] = struct{}{}
		if _, ok := fileMap[file]; !ok {
			fileOrder = append(fileOrder, file)
		}
		fileMap[file] = append(fileMap[file], entry{rule: rule, message: truncate(d.Message, cfg.MaxMessageLength)})
	}

	sort.Strings(fileOrder)

	for _, file := range fileOrder {
		entries := fileMap[file]
		fmt.Printf("%s — %d %s\n", file, len(entries), formatter.Plural(len(entries), "issue", "issues"))
		prevRule := ""
		for _, e := range entries {
			msg := ""
			if e.rule != prevRule {
				msg = " — " + e.message
				prevRule = e.rule
			}
			fmt.Printf("  %s%s\n", e.rule, msg)
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	fmt.Println(formatter.Summary(len(diags), len(allRules), len(fileOrder)))
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

