package main

import (
	"encoding/json"
	"fmt"
	"os"
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

func severityLabel(sev string) string {
	switch sev {
	case "error":
		return "ERROR"
	case "warning":
		return "WARN"
	default:
		return "INFO"
	}
}

func main() {
	formatter.RunWithConfig("stylelint", format)
}

func format(data []byte, cfg formatter.Config) error {
	var files []stylelintFile
	if len(strings.TrimSpace(string(data))) == 0 {
		fmt.Println(formatter.Summary(0, 0, 0))
		return nil
	}
	if err := json.Unmarshal(data, &files); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(string(data)))
		return fmt.Errorf("stylelint error (non-JSON output)")
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

	ruleCounts := make(map[string]int, len(ruleOrder))
	for _, rule := range ruleOrder {
		n := 0
		for _, lines := range rules[rule].fileLines {
			n += len(lines)
		}
		if n == 0 {
			n = len(rules[rule].fileLines)
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
		totalFiles[iss.source] = struct{}{}
	}

	for _, rule := range ruleOrder {
		r := rules[rule]

		hasMatchingFile := false
		for f := range r.fileLines {
			if formatter.MatchesFileFilter(f, cfg.OnlyFiles) {
				hasMatchingFile = true
				break
			}
		}
		if !hasMatchingFile {
			continue
		}

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
		label := severityLabel(r.severity)
		if cfg.Colors {
			fmt.Printf("%s[%s]%s %s (%d) — %s\n", col, label, reset, rule, count, r.message)
		} else {
			fmt.Printf("[%s] %s (%d) — %s\n", label, rule, count, r.message)
		}
		fmt.Println("Affected files:")

		filesSorted := make([]string, 0, len(r.fileLines))
		for f := range r.fileLines {
			filesSorted = append(filesSorted, f)
		}
		sort.Strings(filesSorted)

		for _, f := range filesSorted {
			if !formatter.MatchesFileFilter(f, cfg.OnlyFiles) {
				continue
			}
			ls := r.fileLines[f]
			if len(ls) == 0 {
				fmt.Printf("  - %s\n", f)
			} else {
				sort.Ints(ls)
				lineStrs := make([]string, len(ls))
				for i, l := range ls {
					lineStrs[i] = fmt.Sprintf("%d", l)
				}
				fmt.Printf("  - %s — %s %s\n", f, formatter.Plural(len(ls), "line", "lines"), strings.Join(lineStrs, ", "))
			}
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	fmt.Println(formatter.Summary(displayedIssues, len(ruleOrder), len(totalFiles)))
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

	filtered := fileOrder[:0:0]
	for _, f := range fileOrder {
		if formatter.MatchesFileFilter(f, cfg.OnlyFiles) {
			filtered = append(filtered, f)
		}
	}
	fileOrder = filtered
	sort.Strings(fileOrder)

	totalIssues := 0
	for _, f := range fileOrder {
		entries := fileMap[f]
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
		fmt.Printf("%s — %d %s\n", f, len(filteredEntries), formatter.Plural(len(filteredEntries), "issue", "issues"))
		prevRule := ""
		for _, e := range filteredEntries {
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
