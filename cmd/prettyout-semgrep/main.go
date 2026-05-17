package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type semgrepStart struct {
	Line int `json:"line"`
}

type semgrepExtra struct {
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

type semgrepResult struct {
	CheckID string       `json:"check_id"`
	Path    string       `json:"path"`
	Start   semgrepStart `json:"start"`
	Extra   semgrepExtra `json:"extra"`
}

type semgrepReport struct {
	Results []semgrepResult   `json:"results"`
	Errors  []json.RawMessage `json:"errors"`
}

type ruleEntry struct {
	severity string
	message  string
	fileLines map[string][]int
}

func main() {
	formatter.RunWithConfig("semgrep", format)
}

func format(data []byte, cfg formatter.Config) error {
	var report semgrepReport
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if len(report.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "semgrep: %d error(s) encountered during scan\n", len(report.Errors))
	}

	if cfg.GroupBy == "file" {
		return formatByFile(report.Results, cfg)
	}
	return formatByRule(report.Results, cfg)
}

func severityLabel(sev string) string {
	switch strings.ToUpper(sev) {
	case "ERROR", "FATAL":
		return "ERROR"
	case "WARNING", "WARN":
		return "WARN"
	case "INFO", "INFORMATION":
		return "INFO"
	default:
		if sev == "" {
			return "INFO"
		}
		return strings.ToUpper(sev)
	}
}

func shortCheckID(checkID string) string {
	parts := strings.Split(checkID, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return checkID
}

func formatByRule(results []semgrepResult, cfg formatter.Config) error {
	rules := map[string]*ruleEntry{}
	ruleOrder := []string{}

	for _, r := range results {
		rule := r.CheckID
		if rule == "" {
			rule = "unknown"
		}
		if _, ok := rules[rule]; !ok {
			rules[rule] = &ruleEntry{fileLines: map[string][]int{}}
			ruleOrder = append(ruleOrder, rule)
		}
		re := rules[rule]
		if re.message == "" {
			re.message = truncate(r.Extra.Message, cfg.MaxMessageLength)
			re.severity = r.Extra.Severity
		}
		rp := formatter.ResolvePath(r.Path, cfg)
		re.fileLines[rp] = append(re.fileLines[rp], r.Start.Line)
	}

	sort.Strings(ruleOrder)

	totalFiles := map[string]struct{}{}
	for _, r := range results {
		totalFiles[formatter.ResolvePath(r.Path, cfg)] = struct{}{}
	}

	for _, rule := range ruleOrder {
		re := rules[rule]
		count := 0
		for _, lines := range re.fileLines {
			count += len(lines)
		}
		col := formatter.SeverityColor(strings.ToLower(re.severity), cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		display := shortCheckID(rule)
		sevLabel := severityLabel(re.severity)
		if cfg.Colors {
			fmt.Printf("%s[%s]%s %s (%d) — %s\n", col, sevLabel, reset, display, count, re.message)
		} else {
			fmt.Printf("[%s] %s (%d) — %s\n", sevLabel, display, count, re.message)
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
			lineWord := "lines"
			if len(ls) == 1 {
				lineWord = "line"
			}
			fmt.Printf("  - %s — %s %s\n", f, lineWord, strings.Join(lineStrs, ", "))
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	fmt.Println(formatter.Summary(len(results), len(rules), len(totalFiles)))
	return nil
}

func formatByFile(results []semgrepResult, cfg formatter.Config) error {
	type lineEntry struct {
		rule    string
		line    int
		message string
	}
	fileMap := map[string][]lineEntry{}
	fileOrder := []string{}
	allRules := map[string]struct{}{}

	for _, r := range results {
		rule := r.CheckID
		if rule == "" {
			rule = "unknown"
		}
		allRules[rule] = struct{}{}
		rp := formatter.ResolvePath(r.Path, cfg)
		if _, ok := fileMap[rp]; !ok {
			fileOrder = append(fileOrder, rp)
		}
		fileMap[rp] = append(fileMap[rp], lineEntry{rule: shortCheckID(rule), line: r.Start.Line, message: truncate(r.Extra.Message, cfg.MaxMessageLength)})
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

	fmt.Println(formatter.Summary(len(results), len(allRules), len(fileOrder)))
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
