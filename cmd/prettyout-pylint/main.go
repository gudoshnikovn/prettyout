package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type pylintMsg struct {
	Type          string `json:"type"`
	Line          int    `json:"line"`
	Path          string `json:"path"`
	Symbol        string `json:"symbol"`
	Message       string `json:"message"`
	MessageID     string `json:"message-id"` // json format
	MessageIDJson2 string `json:"messageId"`  // json2 format
}

func (m *pylintMsg) msgID() string {
	if m.MessageIDJson2 != "" {
		return m.MessageIDJson2
	}
	return m.MessageID
}

type pylintJSON2 struct {
	Messages   []pylintMsg `json:"messages"`
	Statistics struct {
		Score float64 `json:"score"`
	} `json:"statistics"`
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

func severityLabel(t string) string {
	switch t {
	case "error", "fatal":
		return "ERROR"
	case "warning":
		return "WARN"
	default:
		return "INFO"
	}
}

func format(data []byte, cfg formatter.Config) error {
	var wrapper pylintJSON2
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	msgs := wrapper.Messages
	score := wrapper.Statistics.Score

	if cfg.GroupBy == "file" {
		return formatByFile(msgs, cfg)
	}
	return formatByRule(msgs, cfg, score)
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
		return m.msgID() + "/" + m.Symbol
	}
	return m.msgID()
}

func formatByRule(msgs []pylintMsg, cfg formatter.Config, score float64) error {
	rules := map[string]*ruleEntry{}
	ruleOrder := []string{}

	for _, m := range msgs {
		key := m.msgID()
		if key == "" {
			key = "unknown"
		}
		fp := formatter.ResolvePath(m.Path, cfg)
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
		if r.fileLines[fp] == nil {
			r.fileLines[fp] = map[int]struct{}{}
		}
		r.fileLines[fp][m.Line] = struct{}{}
	}

	ruleCounts := make(map[string]int, len(ruleOrder))
	for _, key := range ruleOrder {
		n := 0
		for _, lineSet := range rules[key].fileLines {
			n += len(lineSet)
		}
		ruleCounts[key] = n
	}
	ruleOrder = formatter.FilterRuleOrder(ruleOrder, cfg.OnlyRules)
	ruleOrder = formatter.SortOrder(ruleOrder, ruleCounts, cfg.Sort)

	displayedIssues := 0
	for _, key := range ruleOrder {
		displayedIssues += ruleCounts[key]
	}

	totalFiles := map[string]struct{}{}
	for _, m := range msgs {
		totalFiles[formatter.ResolvePath(m.Path, cfg)] = struct{}{}
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
			if m.msgID() == key {
				msgType = m.Type
				break
			}
		}
		col := pylintColor(msgType, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		label := severityLabel(msgType)
		if cfg.Colors {
			fmt.Printf("%s[%s]%s %s (%d) — %s\n", col, label, reset, r.display, count, r.message)
		} else {
			fmt.Printf("[%s] %s (%d) — %s\n", label, r.display, count, r.message)
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

	fmt.Println(formatter.Summary(displayedIssues, len(ruleOrder), len(totalFiles)))
	fmt.Printf("  ↳ rated %.2f/10\n", score)
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
		if m.msgID() == "" {
			key = "unknown"
		}
		allRules[key] = struct{}{}
		fp := formatter.ResolvePath(m.Path, cfg)
		if _, ok := fileMap[fp]; !ok {
			fileOrder = append(fileOrder, fp)
			fileSeen[fp] = map[ruleLineKey]struct{}{}
		}
		rlk := ruleLineKey{rule: key, line: m.Line}
		if _, seen := fileSeen[fp][rlk]; seen {
			continue
		}
		fileSeen[fp][rlk] = struct{}{}
		fileMap[fp] = append(fileMap[fp], lineEntry{rule: key, line: m.Line, message: truncate(m.Message, cfg.MaxMessageLength)})
	}

	filteredOrder := fileOrder[:0:0]
	for _, f := range fileOrder {
		if formatter.MatchesFileFilter(f, cfg.OnlyFiles) {
			filteredOrder = append(filteredOrder, f)
		}
	}
	fileOrder = filteredOrder
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
		sort.Slice(filteredEntries, func(i, j int) bool {
			if filteredEntries[i].line != filteredEntries[j].line {
				return filteredEntries[i].line < filteredEntries[j].line
			}
			return filteredEntries[i].rule < filteredEntries[j].rule
		})
		fmt.Printf("%s — %d %s\n", f, len(filteredEntries), formatter.Plural(len(filteredEntries), "issue", "issues"))
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
