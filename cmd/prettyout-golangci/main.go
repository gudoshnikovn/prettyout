package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type pos struct {
	Filename string `json:"Filename"`
	Line     int    `json:"Line"`
}

type golangciIssue struct {
	FromLinter string `json:"FromLinter"`
	Text       string `json:"Text"`
	Pos        pos    `json:"Pos"`
}

type golangciReport struct {
	Issues []golangciIssue `json:"Issues"`
}

type ruleEntry struct {
	message string
	// file -> sorted lines
	fileLines map[string][]int
}

func main() {
	formatter.RunWithConfig("golangci-lint", format)
}

func format(data []byte, cfg formatter.Config) error {
	if len(strings.TrimSpace(string(data))) == 0 {
		fmt.Println(formatter.Summary(0, 0, 0))
		return nil
	}
	var report golangciReport
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if cfg.GroupBy == "file" {
		return formatByFile(report.Issues, cfg)
	}
	return formatByRule(report.Issues, cfg)
}

func formatByRule(issues []golangciIssue, cfg formatter.Config) error {
	rules := map[string]*ruleEntry{}
	ruleOrder := []string{}

	for _, iss := range issues {
		rule := iss.FromLinter
		if rule == "" {
			rule = "unknown"
		}
		file := iss.Pos.Filename
		if file == "" {
			file = "unknown"
		}

		if _, ok := rules[rule]; !ok {
			rules[rule] = &ruleEntry{fileLines: map[string][]int{}}
			ruleOrder = append(ruleOrder, rule)
		}
		r := rules[rule]
		if r.message == "" {
			r.message = truncate(iss.Text, cfg.MaxMessageLength)
		}
		r.fileLines[file] = append(r.fileLines[file], iss.Pos.Line)
	}

	sort.Strings(ruleOrder)

	totalIssues := 0
	totalFiles := map[string]struct{}{}
	for _, iss := range issues {
		totalIssues++
		f := iss.Pos.Filename
		if f == "" {
			f = "unknown"
		}
		totalFiles[f] = struct{}{}
	}

	for _, rule := range ruleOrder {
		r := rules[rule]
		count := 0
		for _, lines := range r.fileLines {
			count += len(lines)
		}
		if cfg.Colors {
			fmt.Printf("\033[1;36m%s\033[0m (%d) — %s\n", rule, count, r.message)
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
			lines := r.fileLines[f]
			sort.Ints(lines)
			lineStrs := make([]string, len(lines))
			for i, l := range lines {
				lineStrs[i] = fmt.Sprintf("%d", l)
			}
			fmt.Printf("  - %s — lines %s\n", f, strings.Join(lineStrs, ", "))
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	fmt.Println(formatter.Summary(totalIssues, len(rules), len(totalFiles)))
	return nil
}

func formatByFile(issues []golangciIssue, cfg formatter.Config) error {
	type lineEntry struct {
		rule    string
		line    int
		message string
	}
	fileMap := map[string][]lineEntry{}
	fileOrder := []string{}

	for _, iss := range issues {
		rule := iss.FromLinter
		if rule == "" {
			rule = "unknown"
		}
		file := iss.Pos.Filename
		if file == "" {
			file = "unknown"
		}
		if _, ok := fileMap[file]; !ok {
			fileOrder = append(fileOrder, file)
		}
		fileMap[file] = append(fileMap[file], lineEntry{
			rule:    rule,
			line:    iss.Pos.Line,
			message: truncate(iss.Text, cfg.MaxMessageLength),
		})
	}

	sort.Strings(fileOrder)

	totalIssues := len(issues)
	rules := map[string]struct{}{}
	for _, iss := range issues {
		r := iss.FromLinter
		if r == "" {
			r = "unknown"
		}
		rules[r] = struct{}{}
	}

	for _, file := range fileOrder {
		entries := fileMap[file]
		sort.Slice(entries, func(i, j int) bool { return entries[i].line < entries[j].line })
		fmt.Printf("%s — %d %s\n", file, len(entries), formatter.Plural(len(entries), "issue", "issues"))
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

	fmt.Println(formatter.Summary(totalIssues, len(rules), len(fileOrder)))
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
