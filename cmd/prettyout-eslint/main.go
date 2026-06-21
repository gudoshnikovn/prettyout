package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

type eslintMessage struct {
	RuleId   *string `json:"ruleId"`
	Severity int     `json:"severity"`
	Message  string  `json:"message"`
	Line     int     `json:"line"`
}

type eslintFile struct {
	FilePath string          `json:"filePath"`
	Messages []eslintMessage `json:"messages"`
}

type ruleEntry struct {
	severity string
	message  string
	fileLines map[string][]int
}

func main() {
	formatter.RunWithConfig("eslint", format)
}

func format(data []byte, cfg formatter.Config) error {
	var files []eslintFile
	if err := json.Unmarshal(data, &files); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if cfg.GroupBy == "file" {
		return formatByFile(files, cfg)
	}
	return formatByRule(files, cfg)
}

func ruleId(m eslintMessage) string {
	if m.RuleId == nil {
		return "parse-error"
	}
	return *m.RuleId
}

func severityStr(s int) string {
	if s == 1 {
		return "warning"
	}
	return "error"
}

func formatByRule(files []eslintFile, cfg formatter.Config) error {
	rules := map[string]*ruleEntry{}
	ruleOrder := []string{}

	totalIssues := 0
	totalFiles := map[string]struct{}{}

	for _, f := range files {
		if len(f.Messages) == 0 {
			continue
		}
		fp := formatter.ResolvePath(f.FilePath, cfg)
		for _, m := range f.Messages {
			rid := ruleId(m)
			totalIssues++
			totalFiles[fp] = struct{}{}

			if _, ok := rules[rid]; !ok {
				rules[rid] = &ruleEntry{fileLines: map[string][]int{}}
				ruleOrder = append(ruleOrder, rid)
			}
			r := rules[rid]
			if r.message == "" {
				r.message = formatter.Truncate(m.Message, cfg.MaxMessageLength)
				r.severity = severityStr(m.Severity)
			}
			r.fileLines[fp] = append(r.fileLines[fp], m.Line)
		}
	}

	ruleCounts := make(map[string]int, len(ruleOrder))
	for _, rid := range ruleOrder {
		r := rules[rid]
		n := 0
		for _, lines := range r.fileLines {
			n += len(lines)
		}
		ruleCounts[rid] = n
	}
	ruleOrder = formatter.FilterRuleOrder(ruleOrder, cfg.OnlyRules)
	ruleOrder = formatter.SortOrder(ruleOrder, ruleCounts, cfg.Sort)

	if cfg.Stats {
		ruleFileCount := make(map[string]int, len(ruleOrder))
		ruleMessages := make(map[string]string, len(ruleOrder))
		for _, rid := range ruleOrder {
			r := rules[rid]
			ruleFileCount[rid] = len(r.fileLines)
			ruleMessages[rid] = r.message
		}
		formatter.PrintStats(ruleOrder, ruleCounts, ruleFileCount, ruleMessages, len(totalFiles), cfg)
		return nil
	}

	for _, rid := range ruleOrder {
		r := rules[rid]

		hasMatchingFile := false
		for fp := range r.fileLines {
			if formatter.MatchesFileFilter(fp, cfg.OnlyFiles) {
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
		color := formatter.SeverityColor(r.severity, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		label := formatter.SeverityLabel(r.severity)
		if cfg.Colors {
			fmt.Printf("%s[%s]%s %s (%d) — %s\n", color, label, reset, rid, count, r.message)
		} else {
			fmt.Printf("[%s] %s (%d) — %s\n", label, rid, count, r.message)
		}
		fmt.Println("Affected files:")

		filesSorted := make([]string, 0, len(r.fileLines))
		for fp := range r.fileLines {
			filesSorted = append(filesSorted, fp)
		}
		sort.Strings(filesSorted)

		for _, fp := range filesSorted {
			if !formatter.MatchesFileFilter(fp, cfg.OnlyFiles) {
				continue
			}
			lines := r.fileLines[fp]
			sort.Ints(lines)
			lineStrs := make([]string, len(lines))
			for i, l := range lines {
				lineStrs[i] = fmt.Sprintf("%d", l)
			}
			fmt.Printf("  - %s — %s %s\n", fp, formatter.Plural(len(lines), "line", "lines"), strings.Join(lineStrs, ", "))
		}
		fmt.Println(formatter.Divider)
	}

	fmt.Println(formatter.Summary(totalIssues, len(ruleOrder), len(totalFiles)))
	return nil
}

func formatByFile(files []eslintFile, cfg formatter.Config) error {
	type lineEntry struct {
		rule    string
		line    int
		message string
	}

	totalIssues := 0
	allRules := map[string]struct{}{}
	fileCount := 0

	for _, f := range files {
		if len(f.Messages) == 0 {
			continue
		}
		fp := formatter.ResolvePath(f.FilePath, cfg)
		if !formatter.MatchesFileFilter(fp, cfg.OnlyFiles) {
			continue
		}
		entries := make([]lineEntry, 0, len(f.Messages))
		for _, m := range f.Messages {
			rid := ruleId(m)
			if len(cfg.OnlyRules) > 0 {
				found := false
				for _, r := range cfg.OnlyRules {
					if rid == r {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			allRules[rid] = struct{}{}
			entries = append(entries, lineEntry{rule: rid, line: m.Line, message: formatter.Truncate(m.Message, cfg.MaxMessageLength)})
			totalIssues++
		}
		if len(entries) == 0 {
			continue
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].line < entries[j].line })
		fmt.Printf("%s — %d %s\n", fp, len(entries), formatter.Plural(len(entries), "issue", "issues"))
		prevRule := ""
		for _, e := range entries {
			msg := ""
			if e.rule != prevRule {
				msg = " — " + e.message
				prevRule = e.rule
			}
			fmt.Printf("  %s  line %d%s\n", e.rule, e.line, msg)
		}
		fmt.Println(formatter.Divider)
		fileCount++
	}

	fmt.Println(formatter.Summary(totalIssues, len(allRules), fileCount))
	return nil
}

