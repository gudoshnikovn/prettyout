package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type hadolintIssue struct {
	Line    int    `json:"line"`
	Level   string `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
	File    string `json:"file"`
}

type ruleEntry struct {
	severity string
	message  string
	fileLines map[string][]int
}

func main() {
	formatter.RunWithConfig("hadolint", format)
}

func format(data []byte, cfg formatter.Config) error {
	var issues []hadolintIssue
	if err := json.Unmarshal(data, &issues); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Filter out "ignore" level items
	filtered := issues[:0]
	for _, iss := range issues {
		if iss.Level != "ignore" {
			filtered = append(filtered, iss)
		}
	}
	issues = filtered

	if cfg.GroupBy == "file" {
		return formatByFile(issues, cfg)
	}
	return formatByRule(issues, cfg)
}

func hadolintColor(level string, colors bool) string {
	if !colors {
		return ""
	}
	switch level {
	case "error":
		return "\033[1;31m"
	case "warning":
		return "\033[1;33m"
	default:
		return "\033[1;34m"
	}
}

func formatByRule(issues []hadolintIssue, cfg formatter.Config) error {
	rules := map[string]*ruleEntry{}
	ruleOrder := []string{}

	for _, iss := range issues {
		rule := iss.Code
		if rule == "" {
			rule = "unknown"
		}
		if _, ok := rules[rule]; !ok {
			rules[rule] = &ruleEntry{fileLines: map[string][]int{}}
			ruleOrder = append(ruleOrder, rule)
		}
		r := rules[rule]
		if r.message == "" {
			r.message = truncate(iss.Message, cfg.MaxMessageLength)
			r.severity = iss.Level
		}
		r.fileLines[iss.File] = append(r.fileLines[iss.File], iss.Line)
	}

	sort.Strings(ruleOrder)

	totalFiles := map[string]struct{}{}
	for _, iss := range issues {
		totalFiles[iss.File] = struct{}{}
	}

	for _, rule := range ruleOrder {
		r := rules[rule]
		count := 0
		for _, lines := range r.fileLines {
			count += len(lines)
		}
		col := hadolintColor(r.severity, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		if cfg.Colors {
			fmt.Printf("%s%s%s (%d) — %s\n", col, rule, reset, count, r.message)
		} else {
			fmt.Printf("%s (%d) — %s\n", rule, count, r.message)
		}
		fmt.Println("Affected files:")

		files := make([]string, 0, len(r.fileLines))
		for f := range r.fileLines {
			files = append(files, f)
		}
		sort.Strings(files)

		for _, f := range files {
			ls := r.fileLines[f]
			sort.Ints(ls)
			lineStrs := make([]string, len(ls))
			for i, l := range ls {
				lineStrs[i] = fmt.Sprintf("%d", l)
			}
			fmt.Printf("  - %s — lines %s\n", f, strings.Join(lineStrs, ", "))
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	fmt.Println(formatter.Summary(len(issues), len(rules), len(totalFiles)))
	return nil
}

func formatByFile(issues []hadolintIssue, cfg formatter.Config) error {
	type lineEntry struct {
		rule    string
		line    int
		message string
	}
	fileMap := map[string][]lineEntry{}
	fileOrder := []string{}
	allRules := map[string]struct{}{}

	for _, iss := range issues {
		rule := iss.Code
		if rule == "" {
			rule = "unknown"
		}
		allRules[rule] = struct{}{}
		if _, ok := fileMap[iss.File]; !ok {
			fileOrder = append(fileOrder, iss.File)
		}
		fileMap[iss.File] = append(fileMap[iss.File], lineEntry{rule: rule, line: iss.Line, message: truncate(iss.Message, cfg.MaxMessageLength)})
	}

	sort.Strings(fileOrder)

	for _, f := range fileOrder {
		entries := fileMap[f]
		sort.Slice(entries, func(i, j int) bool { return entries[i].line < entries[j].line })
		fmt.Printf("%s — %d %s\n", f, len(entries), formatter.Plural(len(entries), "issue", "issues"))
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

	fmt.Println(formatter.Summary(len(issues), len(allRules), len(fileOrder)))
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
