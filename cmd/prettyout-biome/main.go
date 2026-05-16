package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type biomePath struct {
	File string `json:"file"`
}

type biomeLocation struct {
	Path biomePath `json:"path"`
}

type biomeDiagnostic struct {
	Category    string        `json:"category"`
	Description string        `json:"description"`
	Severity    string        `json:"severity"`
	Location    biomeLocation `json:"location"`
}

type biomeReport struct {
	Diagnostics []biomeDiagnostic `json:"diagnostics"`
}

type ruleEntry struct {
	severity string
	message  string
	files    map[string]struct{}
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
		file := d.Location.Path.File

		if _, ok := rules[rule]; !ok {
			rules[rule] = &ruleEntry{files: map[string]struct{}{}}
			ruleOrder = append(ruleOrder, rule)
		}
		r := rules[rule]
		if r.message == "" {
			r.message = truncate(d.Description, cfg.MaxMessageLength)
			r.severity = d.Severity
		}
		if file != "" {
			r.files[file] = struct{}{}
		}
	}

	sort.Strings(ruleOrder)

	totalFiles := map[string]struct{}{}
	for _, d := range diags {
		if d.Location.Path.File != "" {
			totalFiles[d.Location.Path.File] = struct{}{}
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
		if cfg.Colors {
			fmt.Printf("%s%s%s (%d) — %s\n", col, rule, reset, occurrences, r.message)
		} else {
			fmt.Printf("%s (%d) — %s\n", rule, occurrences, r.message)
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

	fmt.Printf("%d issues · %d rules · %d files\n", len(diags), len(rules), len(totalFiles))
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
		file := d.Location.Path.File
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
		fileMap[file] = append(fileMap[file], entry{rule: rule, message: truncate(d.Description, cfg.MaxMessageLength)})
	}

	sort.Strings(fileOrder)

	for _, file := range fileOrder {
		entries := fileMap[file]
		fmt.Printf("%s — %d issues\n", file, len(entries))
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

	fmt.Printf("%d issues · %d rules · %d files\n", len(diags), len(allRules), len(fileOrder))
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

