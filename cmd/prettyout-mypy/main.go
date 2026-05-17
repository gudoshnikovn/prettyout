package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type mypyMsg struct {
	File     string  `json:"file"`
	Line     int     `json:"line"`
	Severity string  `json:"severity"`
	Message  string  `json:"message"`
	Code     *string `json:"code"`
}

type ruleEntry struct {
	severity  string
	message   string
	fileLines map[string]map[int]struct{}
}

func main() {
	formatter.RunWithConfig("mypy", format)
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

func format(data []byte, cfg formatter.Config) error {
	lines, err := formatter.ParseNDJSON(data)
	if err != nil {
		return fmt.Errorf("invalid NDJSON: %w", err)
	}

	var msgs []mypyMsg
	for _, raw := range lines {
		var m mypyMsg
		if err := json.Unmarshal(raw, &m); err != nil {
			continue
		}
		// skip notes
		if m.Severity == "note" {
			continue
		}
		msgs = append(msgs, m)
	}

	if cfg.GroupBy == "file" {
		return formatByFile(msgs, cfg)
	}
	return formatByRule(msgs, cfg)
}

func codeStr(m mypyMsg) string {
	if m.Code != nil && *m.Code != "" {
		return *m.Code
	}
	return "error"
}

func formatByRule(msgs []mypyMsg, cfg formatter.Config) error {
	rules := map[string]*ruleEntry{}
	ruleOrder := []string{}

	for _, m := range msgs {
		rule := codeStr(m)
		if _, ok := rules[rule]; !ok {
			rules[rule] = &ruleEntry{fileLines: map[string]map[int]struct{}{}}
			ruleOrder = append(ruleOrder, rule)
		}
		r := rules[rule]
		if r.message == "" {
			r.message = truncate(m.Message, cfg.MaxMessageLength)
			r.severity = m.Severity
		}
		if r.fileLines[m.File] == nil {
			r.fileLines[m.File] = map[int]struct{}{}
		}
		r.fileLines[m.File][m.Line] = struct{}{}
	}

	sort.Strings(ruleOrder)

	totalFiles := map[string]struct{}{}
	for _, m := range msgs {
		totalFiles[m.File] = struct{}{}
	}

	for _, rule := range ruleOrder {
		r := rules[rule]
		count := 0
		for _, lineSet := range r.fileLines {
			count += len(lineSet)
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

		files := make([]string, 0, len(r.fileLines))
		for f := range r.fileLines {
			files = append(files, f)
		}
		sort.Strings(files)

		for _, f := range files {
			lineSet := r.fileLines[f]
			ls := make([]int, 0, len(lineSet))
			for l := range lineSet {
				ls = append(ls, l)
			}
			sort.Ints(ls)
			lineStrs := make([]string, len(ls))
			for i, l := range ls {
				lineStrs[i] = fmt.Sprintf("%d", l)
			}
			lineWord := formatter.Plural(len(ls), "line", "lines")
			fmt.Printf("  - %s — %s %s\n", f, lineWord, strings.Join(lineStrs, ", "))
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	fmt.Println(formatter.Summary(len(msgs), len(rules), len(totalFiles)))
	return nil
}

func formatByFile(msgs []mypyMsg, cfg formatter.Config) error {
	type lineEntry struct {
		rule    string
		line    int
		message string
	}
	fileMap := map[string][]lineEntry{}
	fileOrder := []string{}
	allRules := map[string]struct{}{}

	for _, m := range msgs {
		rule := codeStr(m)
		allRules[rule] = struct{}{}
		if _, ok := fileMap[m.File]; !ok {
			fileOrder = append(fileOrder, m.File)
		}
		fileMap[m.File] = append(fileMap[m.File], lineEntry{rule: rule, line: m.Line, message: truncate(m.Message, cfg.MaxMessageLength)})
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

	fmt.Println(formatter.Summary(len(msgs), len(allRules), len(fileOrder)))
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
