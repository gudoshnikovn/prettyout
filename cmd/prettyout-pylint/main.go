package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type pylintMsg struct {
	Type      string `json:"type"`
	Line      int    `json:"line"`
	Path      string `json:"path"`
	Symbol    string `json:"symbol"`
	Message   string `json:"message"`
	MessageID string `json:"message-id"`
}

type ruleEntry struct {
	severity  string
	message   string
	display   string
	fileLines map[string]map[int]struct{}
}

func main() {
	formatter.RunWithConfig("pylint", format)
}

func format(data []byte, cfg formatter.Config) error {
	var msgs []pylintMsg
	if err := json.Unmarshal(data, &msgs); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if cfg.GroupBy == "file" {
		return formatByFile(msgs, cfg)
	}
	return formatByRule(msgs, cfg)
}

func pylintSeverity(t string) string {
	switch t {
	case "error", "fatal":
		return "error"
	case "warning":
		return "warning"
	default:
		return "info"
	}
}

func pylintColor(t string, colors bool) string {
	if !colors {
		return ""
	}
	switch t {
	case "error", "fatal":
		return "\033[1;31m"
	case "warning":
		return "\033[1;33m"
	default:
		return "\033[2m" // dim
	}
}

func ruleDisplay(m pylintMsg) string {
	if m.Symbol != "" {
		return m.MessageID + "/" + m.Symbol
	}
	return m.MessageID
}

func formatByRule(msgs []pylintMsg, cfg formatter.Config) error {
	rules := map[string]*ruleEntry{}
	ruleOrder := []string{}

	for _, m := range msgs {
		key := m.MessageID
		if key == "" {
			key = "unknown"
		}
		if _, ok := rules[key]; !ok {
			rules[key] = &ruleEntry{fileLines: map[string]map[int]struct{}{}}
			ruleOrder = append(ruleOrder, key)
		}
		r := rules[key]
		if r.message == "" {
			r.message = truncate(m.Message, cfg.MaxMessageLength)
			r.severity = pylintSeverity(m.Type)
			r.display = ruleDisplay(m)
		}
		if r.fileLines[m.Path] == nil {
			r.fileLines[m.Path] = map[int]struct{}{}
		}
		r.fileLines[m.Path][m.Line] = struct{}{}
	}

	sort.Strings(ruleOrder)

	totalFiles := map[string]struct{}{}
	for _, m := range msgs {
		totalFiles[m.Path] = struct{}{}
	}

	for _, key := range ruleOrder {
		r := rules[key]
		count := 0
		for _, lineSet := range r.fileLines {
			count += len(lineSet)
		}

		// find the type for this rule
		msgType := ""
		for _, m := range msgs {
			if m.MessageID == key {
				msgType = m.Type
				break
			}
		}
		col := pylintColor(msgType, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		if cfg.Colors {
			fmt.Printf("%s%s%s (%d) — %s\n", col, r.display, reset, count, r.message)
		} else {
			fmt.Printf("%s (%d) — %s\n", r.display, count, r.message)
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
			label := "lines"
			if len(ls) == 1 {
				label = "line"
			}
			fmt.Printf("  - %s — %s %s\n", f, label, strings.Join(lineStrs, ", "))
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	totalIssues := 0
	for _, r := range rules {
		for _, lineSet := range r.fileLines {
			totalIssues += len(lineSet)
		}
	}
	fmt.Println(formatter.Summary(totalIssues, len(rules), len(totalFiles)))
	return nil
}

func formatByFile(msgs []pylintMsg, cfg formatter.Config) error {
	type lineEntry struct {
		rule    string
		line    int
		message string
	}
	// fileMap stores deduplicated (rule, line) pairs per file
	type ruleLineKey struct {
		rule string
		line int
	}
	fileMap := map[string][]lineEntry{}
	fileSeen := map[string]map[ruleLineKey]struct{}{}
	fileOrder := []string{}
	allRules := map[string]struct{}{}

	for _, m := range msgs {
		key := ruleDisplay(m)
		if m.MessageID == "" {
			key = "unknown"
		}
		allRules[key] = struct{}{}
		if _, ok := fileMap[m.Path]; !ok {
			fileOrder = append(fileOrder, m.Path)
			fileSeen[m.Path] = map[ruleLineKey]struct{}{}
		}
		rlk := ruleLineKey{rule: key, line: m.Line}
		if _, seen := fileSeen[m.Path][rlk]; seen {
			continue
		}
		fileSeen[m.Path][rlk] = struct{}{}
		fileMap[m.Path] = append(fileMap[m.Path], lineEntry{rule: key, line: m.Line, message: truncate(m.Message, cfg.MaxMessageLength)})
	}

	sort.Strings(fileOrder)

	totalIssues := 0
	for _, entries := range fileMap {
		totalIssues += len(entries)
	}

	for _, f := range fileOrder {
		entries := fileMap[f]
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].line != entries[j].line {
				return entries[i].line < entries[j].line
			}
			return entries[i].rule < entries[j].rule
		})
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

	fmt.Println(formatter.Summary(totalIssues, len(allRules), len(fileOrder)))
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
