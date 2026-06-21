package main

import (
	"encoding/json"
	"fmt"
	"sort"

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

type ruleEntry struct {
	severity    string
	message     string
	files       map[string]struct{}
	occurrences int
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
		if file != "" {
			file = formatter.ResolvePath(file, cfg)
		}

		if _, ok := rules[rule]; !ok {
			rules[rule] = &ruleEntry{files: map[string]struct{}{}}
			ruleOrder = append(ruleOrder, rule)
		}
		r := rules[rule]
		if r.message == "" {
			r.message = formatter.Truncate(d.Message, cfg.MaxMessageLength)
			r.severity = d.Severity
		}
		r.occurrences++
		if file != "" {
			r.files[file] = struct{}{}
		}
	}

	ruleCounts := make(map[string]int, len(ruleOrder))
	for _, rule := range ruleOrder {
		ruleCounts[rule] = rules[rule].occurrences
	}
	ruleOrder = formatter.FilterRuleOrder(ruleOrder, cfg.OnlyRules)
	ruleOrder = formatter.SortOrder(ruleOrder, ruleCounts, cfg.Sort)

	displayedIssues := 0
	for _, rule := range ruleOrder {
		displayedIssues += ruleCounts[rule]
	}

	totalFiles := map[string]struct{}{}
	for _, d := range diags {
		if d.Location.Path != "" {
			totalFiles[formatter.ResolvePath(d.Location.Path, cfg)] = struct{}{}
		}
	}

	if cfg.Stats {
		ruleFileCount := make(map[string]int, len(ruleOrder))
		ruleMessages := make(map[string]string, len(ruleOrder))
		for _, rule := range ruleOrder {
			r := rules[rule]
			ruleFileCount[rule] = len(r.files)
			ruleMessages[rule] = r.message
		}
		formatter.PrintStats(ruleOrder, ruleCounts, ruleFileCount, ruleMessages, len(totalFiles), cfg)
		return nil
	}

	for _, rule := range ruleOrder {
		r := rules[rule]

		hasMatchingFile := false
		for f := range r.files {
			if formatter.MatchesFileFilter(f, cfg.OnlyFiles) {
				hasMatchingFile = true
				break
			}
		}
		if !hasMatchingFile {
			continue
		}

		col := formatter.SeverityColor(r.severity, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		label := formatter.SeverityLabel(r.severity)
		if cfg.Colors {
			fmt.Printf("%s[%s]%s %s (%d) — %s\n", col, label, reset, rule, r.occurrences, r.message)
		} else {
			fmt.Printf("[%s] %s (%d) — %s\n", label, rule, r.occurrences, r.message)
		}

		if len(r.files) > 0 {
			fmt.Println("Affected files:")
			files := make([]string, 0, len(r.files))
			for f := range r.files {
				files = append(files, f)
			}
			sort.Strings(files)
			for _, f := range files {
				if !formatter.MatchesFileFilter(f, cfg.OnlyFiles) {
					continue
				}
				fmt.Printf("  - %s\n", f)
			}
		}
		fmt.Println(formatter.Divider)
	}

	fmt.Println(formatter.Summary(displayedIssues, len(ruleOrder), len(totalFiles)))
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
		} else {
			file = formatter.ResolvePath(file, cfg)
		}
		rule := d.Category
		if rule == "" {
			rule = "unknown"
		}
		allRules[rule] = struct{}{}
		if _, ok := fileMap[file]; !ok {
			fileOrder = append(fileOrder, file)
		}
		fileMap[file] = append(fileMap[file], entry{rule: rule, message: formatter.Truncate(d.Message, cfg.MaxMessageLength)})
	}

	filtered := fileOrder[:0:0]
	for _, f := range fileOrder {
		if formatter.MatchesFileFilter(f, cfg.OnlyFiles) {
			filtered = append(filtered, f)
		}
	}
	fileOrder = filtered
	sort.Strings(fileOrder)

	totalIssues := 0
	for _, file := range fileOrder {
		entries := fileMap[file]
		var filteredEntries []entry
		for _, e := range entries {
			if len(cfg.OnlyRules) > 0 {
				found := false
				for _, r := range cfg.OnlyRules {
					if e.rule == r {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			filteredEntries = append(filteredEntries, e)
		}
		if len(filteredEntries) == 0 {
			continue
		}
		fmt.Printf("%s — %d %s\n", file, len(filteredEntries), formatter.Plural(len(filteredEntries), "issue", "issues"))
		prevRule := ""
		for _, e := range filteredEntries {
			msg := ""
			if e.rule != prevRule {
				msg = " — " + e.message
				prevRule = e.rule
			}
			fmt.Printf("  %s%s\n", e.rule, msg)
		}
		fmt.Println(formatter.Divider)
		totalIssues += len(filteredEntries)
	}

	fmt.Println(formatter.Summary(totalIssues, len(allRules), len(fileOrder)))
	return nil
}


