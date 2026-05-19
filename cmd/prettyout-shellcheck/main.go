package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type shellcheckIssue struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Level   string `json:"level"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ruleEntry struct {
	severity string
	message  string
	fileLines map[string][]int
}

func severityLabel(level string) string {
	switch level {
	case "error":
		return "ERROR"
	case "warning":
		return "WARN"
	default:
		return "INFO"
	}
}

func main() {
	formatter.RunWithConfig("shellcheck", format)
}

func format(data []byte, cfg formatter.Config) error {
	var issues []shellcheckIssue
	if err := json.Unmarshal(data, &issues); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if cfg.GroupBy == "file" {
		return formatByFile(issues, cfg)
	}
	return formatByRule(issues, cfg)
}

func scCode(code int) string {
	return fmt.Sprintf("SC%d", code)
}

func shellcheckColor(level string, colors bool) string {
	if !colors {
		return ""
	}
	switch level {
	case "error":
		return "\033[1;31m"
	case "warning":
		return "\033[1;33m"
	default:
		return "\033[2m" // dim for info/style
	}
}

func formatByRule(issues []shellcheckIssue, cfg formatter.Config) error {
	rules := map[string]*ruleEntry{}
	ruleOrder := []string{}

	for _, iss := range issues {
		rule := scCode(iss.Code)
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

	ruleCounts := make(map[string]int, len(ruleOrder))
	for _, rule := range ruleOrder {
		r := rules[rule]
		n := 0
		for _, lines := range r.fileLines {
			n += len(lines)
		}
		ruleCounts[rule] = n
	}
	ruleOrder = formatter.FilterRuleOrder(ruleOrder, cfg.OnlyRules)
	ruleOrder = formatter.SortOrder(ruleOrder, ruleCounts, cfg.Sort)

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
		col := shellcheckColor(r.severity, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		label := severityLabel(r.severity)
		if cfg.Colors {
			fmt.Printf("%s[%s]%s %s (%d) — %s\n", col, label, reset, rule, count, r.message)
		} else {
			fmt.Printf("[%s] %s (%d) — %s\n", label, rule, count, r.message)
		}
		fmt.Println("Affected files:")

		files := make([]string, 0, len(r.fileLines))
		for f := range r.fileLines {
			files = append(files, f)
		}
		sort.Strings(files)

		for _, f := range files {
			if !formatter.MatchesFileFilter(f, cfg.OnlyFiles) {
				continue
			}
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

func formatByFile(issues []shellcheckIssue, cfg formatter.Config) error {
	type lineEntry struct {
		rule    string
		line    int
		message string
	}
	fileMap := map[string][]lineEntry{}
	fileOrder := []string{}
	allRules := map[string]struct{}{}

	for _, iss := range issues {
		rule := scCode(iss.Code)
		allRules[rule] = struct{}{}
		if _, ok := fileMap[iss.File]; !ok {
			fileOrder = append(fileOrder, iss.File)
		}
		fileMap[iss.File] = append(fileMap[iss.File], lineEntry{rule: rule, line: iss.Line, message: truncate(iss.Message, cfg.MaxMessageLength)})
	}

	filteredFileOrder := fileOrder[:0:0]
	for _, f := range fileOrder {
		if formatter.MatchesFileFilter(f, cfg.OnlyFiles) {
			filteredFileOrder = append(filteredFileOrder, f)
		}
	}
	fileOrder = filteredFileOrder
	sort.Strings(fileOrder)

	for _, f := range fileOrder {
		entries := fileMap[f]
		sort.Slice(entries, func(i, j int) bool { return entries[i].line < entries[j].line })
		prevRule := ""
		filteredEntries := entries[:0:0]
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
		fmt.Printf("%s — %d %s\n", f, len(filteredEntries), formatter.Plural(len(filteredEntries), "issue", "issues"))
		for _, e := range filteredEntries {
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
