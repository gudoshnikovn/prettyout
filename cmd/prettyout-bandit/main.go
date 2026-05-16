package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type banditResult struct {
	Filename        string `json:"filename"`
	LineNumber      int    `json:"line_number"`
	TestID          string `json:"test_id"`
	IssueText       string `json:"issue_text"`
	IssueSeverity   string `json:"issue_severity"`
	IssueConfidence string `json:"issue_confidence"`
}

type banditReport struct {
	Errors  []json.RawMessage `json:"errors"`
	Results []banditResult    `json:"results"`
}

type ruleEntry struct {
	severity   string
	confidence string
	message    string
	fileLines  map[string][]int
}

func main() {
	formatter.RunWithConfig("bandit", format)
}

func format(data []byte, cfg formatter.Config) error {
	var report banditReport
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if len(report.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "bandit: %d file(s) could not be parsed\n", len(report.Errors))
	}

	if cfg.GroupBy == "file" {
		return formatByFile(report.Results, cfg)
	}
	return formatByRule(report.Results, cfg)
}

func cleanFilename(f string) string {
	return strings.TrimPrefix(f, "./")
}

func severityColor(sev string, colors bool) string {
	if !colors {
		return ""
	}
	switch strings.ToUpper(sev) {
	case "HIGH", "CRITICAL":
		return "\033[1;31m"
	case "MEDIUM":
		return "\033[1;33m"
	default:
		return "\033[0m"
	}
}

func formatByRule(results []banditResult, cfg formatter.Config) error {
	rules := map[string]*ruleEntry{}
	ruleOrder := []string{}

	for _, r := range results {
		rule := r.TestID
		if rule == "" {
			rule = "unknown"
		}
		file := cleanFilename(r.Filename)

		if _, ok := rules[rule]; !ok {
			rules[rule] = &ruleEntry{fileLines: map[string][]int{}}
			ruleOrder = append(ruleOrder, rule)
		}
		re := rules[rule]
		if re.message == "" {
			re.message = truncate(r.IssueText, cfg.MaxMessageLength)
			re.severity = r.IssueSeverity
			re.confidence = r.IssueConfidence
		}
		re.fileLines[file] = append(re.fileLines[file], r.LineNumber)
	}

	sort.Strings(ruleOrder)

	totalFiles := map[string]struct{}{}
	for _, r := range results {
		totalFiles[cleanFilename(r.Filename)] = struct{}{}
	}

	for _, rule := range ruleOrder {
		re := rules[rule]
		count := 0
		for _, lines := range re.fileLines {
			count += len(lines)
		}
		col := severityColor(re.severity, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		header := fmt.Sprintf("%s (%s/%s)", rule, re.severity, re.confidence)
		if cfg.Colors {
			fmt.Printf("%s%s%s (%d) — %s\n", col, header, reset, count, re.message)
		} else {
			fmt.Printf("%s (%d) — %s\n", header, count, re.message)
		}
		fmt.Println("Affected files:")

		files := make([]string, 0, len(re.fileLines))
		for f := range re.fileLines {
			files = append(files, f)
		}
		sort.Strings(files)

		for _, f := range files {
			ls := re.fileLines[f]
			sort.Ints(ls)
			lineStrs := make([]string, len(ls))
			for i, l := range ls {
				lineStrs[i] = fmt.Sprintf("%d", l)
			}
			fmt.Printf("  - %s — lines %s\n", f, strings.Join(lineStrs, ", "))
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	fmt.Printf("%d issues · %d rules · %d files\n", len(results), len(rules), len(totalFiles))
	return nil
}

func formatByFile(results []banditResult, cfg formatter.Config) error {
	type lineEntry struct {
		rule    string
		line    int
		message string
	}
	fileMap := map[string][]lineEntry{}
	fileOrder := []string{}
	allRules := map[string]struct{}{}

	for _, r := range results {
		rule := r.TestID
		if rule == "" {
			rule = "unknown"
		}
		file := cleanFilename(r.Filename)
		allRules[rule] = struct{}{}
		if _, ok := fileMap[file]; !ok {
			fileOrder = append(fileOrder, file)
		}
		fileMap[file] = append(fileMap[file], lineEntry{rule: rule, line: r.LineNumber, message: truncate(r.IssueText, cfg.MaxMessageLength)})
	}

	sort.Strings(fileOrder)

	for _, f := range fileOrder {
		entries := fileMap[f]
		sort.Slice(entries, func(i, j int) bool { return entries[i].line < entries[j].line })
		fmt.Printf("%s — %d issues\n", f, len(entries))
		prevRule := ""
		for _, e := range entries {
			msg := ""
			if e.rule != prevRule {
				msg = " — " + e.message
				prevRule = e.rule
			}
			fmt.Printf("  %s  line %d%s\n", e.rule, e.line, msg)
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	fmt.Printf("%d issues · %d rules · %d files\n", len(results), len(allRules), len(fileOrder))
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
