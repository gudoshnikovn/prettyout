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
		} else {
			file = formatter.ResolvePath(file, cfg)
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

	ruleCounts := make(map[string]int, len(ruleOrder))
	for _, rule := range ruleOrder {
		n := 0
		for _, lines := range rules[rule].fileLines {
			n += len(lines)
		}
		ruleCounts[rule] = n
	}
	ruleOrder = formatter.FilterRuleOrder(ruleOrder, cfg.OnlyRules)
	ruleOrder = formatter.SortOrder(ruleOrder, ruleCounts, cfg.Sort)

	displayedIssues := 0
	for _, rule := range ruleOrder {
		displayedIssues += ruleCounts[rule]
	}

	totalFiles := map[string]struct{}{}
	for _, iss := range issues {
		f := iss.Pos.Filename
		if f == "" {
			f = "unknown"
		} else {
			f = formatter.ResolvePath(f, cfg)
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
			if !formatter.MatchesFileFilter(f, cfg.OnlyFiles) {
				continue
			}
			lines := r.fileLines[f]
			sort.Ints(lines)
			lineStrs := make([]string, len(lines))
			for i, l := range lines {
				lineStrs[i] = fmt.Sprintf("%d", l)
			}
			fmt.Printf("  - %s — %s %s\n", f, formatter.Plural(len(lines), "line", "lines"), strings.Join(lineStrs, ", "))
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	fmt.Println(formatter.Summary(displayedIssues, len(ruleOrder), len(totalFiles)))
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
		} else {
			file = formatter.ResolvePath(file, cfg)
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

	filtered := fileOrder[:0:0]
	for _, f := range fileOrder {
		if formatter.MatchesFileFilter(f, cfg.OnlyFiles) {
			filtered = append(filtered, f)
		}
	}
	fileOrder = filtered
	sort.Strings(fileOrder)

	allRules := map[string]struct{}{}
	for _, iss := range issues {
		r := iss.FromLinter
		if r == "" {
			r = "unknown"
		}
		allRules[r] = struct{}{}
	}

	totalIssues := 0
	for _, file := range fileOrder {
		entries := fileMap[file]
		var filteredEntries []lineEntry
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
		sort.Slice(filteredEntries, func(i, j int) bool { return filteredEntries[i].line < filteredEntries[j].line })
		fmt.Printf("%s — %d %s\n", file, len(filteredEntries), formatter.Plural(len(filteredEntries), "issue", "issues"))
		prevRule := ""
		for _, e := range filteredEntries {
			msg := ""
			if e.rule != prevRule {
				msg = " — " + e.message
				prevRule = e.rule
			}
			fmt.Printf("  %s  line %d%s\n", e.rule, e.line, msg)
		}
		fmt.Println("────────────────────────────────────────────────")
		totalIssues += len(filteredEntries)
	}

	fmt.Println(formatter.Summary(totalIssues, len(allRules), len(fileOrder)))
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
