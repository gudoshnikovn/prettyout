package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type stylelintWarning struct {
	Line     int    `json:"line"`
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Text     string `json:"text"`
}

type stylelintFile struct {
	Source      string             `json:"source"`
	ParseErrors []json.RawMessage  `json:"parseErrors"`
	Warnings    []stylelintWarning `json:"warnings"`
}

type ruleEntry struct {
	severity  string
	message   string
	fileLines map[string][]int
}

type issueItem struct {
	source   string
	rule     string
	severity string
	line     int
	message  string
}

func main() {
	formatter.RunWithConfig("stylelint", format)
}

func format(data []byte, cfg formatter.Config) error {
	var files []stylelintFile
	if err := json.Unmarshal(data, &files); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	var allIssues []issueItem
	for _, f := range files {
		source := formatter.ResolvePath(f.Source, cfg)
		for range f.ParseErrors {
			allIssues = append(allIssues, issueItem{source: source, rule: "parse-error", severity: "error", line: 0, message: "CSS parse error"})
		}
		for _, w := range f.Warnings {
			rule := w.Rule
			if rule == "" {
				rule = "unknown"
			}
			allIssues = append(allIssues, issueItem{
				source:   source,
				rule:     rule,
				severity: w.Severity,
				line:     w.Line,
				message:  truncate(w.Text, cfg.MaxMessageLength),
			})
		}
	}

	if cfg.GroupBy == "file" {
		return formatByFile(allIssues, cfg)
	}
	return formatByRule(allIssues, cfg)
}

func formatByRule(issues []issueItem, cfg formatter.Config) error {
	rules := map[string]*ruleEntry{}
	ruleOrder := []string{}

	for _, iss := range issues {
		rule := iss.rule
		if _, ok := rules[rule]; !ok {
			rules[rule] = &ruleEntry{fileLines: map[string][]int{}}
			ruleOrder = append(ruleOrder, rule)
		}
		r := rules[rule]
		if r.message == "" {
			r.message = iss.message
			r.severity = iss.severity
		}
		if iss.line > 0 {
			r.fileLines[iss.source] = append(r.fileLines[iss.source], iss.line)
		} else {
			if _, ok := r.fileLines[iss.source]; !ok {
				r.fileLines[iss.source] = nil
			}
		}
	}

	sort.Strings(ruleOrder)

	totalFiles := map[string]struct{}{}
	for _, iss := range issues {
		totalFiles[iss.source] = struct{}{}
	}

	for _, rule := range ruleOrder {
		r := rules[rule]
		count := 0
		for _, lines := range r.fileLines {
			count += len(lines)
		}
		if count == 0 {
			count = len(r.fileLines)
		}
		col := formatter.SeverityColor(r.severity, cfg.Colors)
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

		filesSorted := make([]string, 0, len(r.fileLines))
		for f := range r.fileLines {
			filesSorted = append(filesSorted, f)
		}
		sort.Strings(filesSorted)

		for _, f := range filesSorted {
			ls := r.fileLines[f]
			if len(ls) == 0 {
				fmt.Printf("  - %s\n", f)
			} else {
				sort.Ints(ls)
				lineStrs := make([]string, len(ls))
				for i, l := range ls {
					lineStrs[i] = fmt.Sprintf("%d", l)
				}
				fmt.Printf("  - %s — lines %s\n", f, strings.Join(lineStrs, ", "))
			}
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	fmt.Println(formatter.Summary(len(issues), len(rules), len(totalFiles)))
	return nil
}

func formatByFile(issues []issueItem, cfg formatter.Config) error {
	type lineEntry struct {
		rule    string
		line    int
		message string
	}
	fileMap := map[string][]lineEntry{}
	fileOrder := []string{}
	allRules := map[string]struct{}{}

	for _, iss := range issues {
		allRules[iss.rule] = struct{}{}
		if _, ok := fileMap[iss.source]; !ok {
			fileOrder = append(fileOrder, iss.source)
		}
		fileMap[iss.source] = append(fileMap[iss.source], lineEntry{rule: iss.rule, line: iss.line, message: iss.message})
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
			if e.line > 0 {
				fmt.Printf("  %s  line %d%s\n", e.rule, e.line, msg)
			} else {
				fmt.Printf("  %s%s\n", e.rule, msg)
			}
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
