package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
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
				r.message = truncate(m.Message, cfg.MaxMessageLength)
				r.severity = severityStr(m.Severity)
			}
			r.fileLines[fp] = append(r.fileLines[fp], m.Line)
		}
	}

	sort.Strings(ruleOrder)

	for _, rid := range ruleOrder {
		r := rules[rid]
		count := 0
		for _, lines := range r.fileLines {
			count += len(lines)
		}
		color := formatter.SeverityColor(r.severity, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		if cfg.Colors {
			fmt.Printf("%s\033[1;35m%s\033[0m%s (%d) — %s\n", color, rid, reset, count, r.message)
		} else {
			fmt.Printf("%s (%d) — %s\n", rid, count, r.message)
		}
		fmt.Println("Affected files:")

		filesSorted := make([]string, 0, len(r.fileLines))
		for fp := range r.fileLines {
			filesSorted = append(filesSorted, fp)
		}
		sort.Strings(filesSorted)

		for _, fp := range filesSorted {
			lines := r.fileLines[fp]
			sort.Ints(lines)
			lineStrs := make([]string, len(lines))
			for i, l := range lines {
				lineStrs[i] = fmt.Sprintf("%d", l)
			}
			fmt.Printf("  - %s — lines %s\n", fp, strings.Join(lineStrs, ", "))
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	fmt.Println(formatter.Summary(totalIssues, len(rules), len(totalFiles)))
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
		entries := make([]lineEntry, 0, len(f.Messages))
		for _, m := range f.Messages {
			rid := ruleId(m)
			allRules[rid] = struct{}{}
			entries = append(entries, lineEntry{rule: rid, line: m.Line, message: truncate(m.Message, cfg.MaxMessageLength)})
			totalIssues++
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
		fmt.Println("────────────────────────────────────────────────")
		fileCount++
	}

	fmt.Println(formatter.Summary(totalIssues, len(allRules), fileCount))
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
