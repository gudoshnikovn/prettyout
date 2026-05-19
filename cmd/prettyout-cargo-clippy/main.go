package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type clippySpan struct {
	FileName  string `json:"file_name"`
	LineStart int    `json:"line_start"`
	IsPrimary bool   `json:"is_primary"`
}

type clippyCode struct {
	Code string `json:"code"`
}

type clippyMessage struct {
	Level   string       `json:"level"`
	Message string       `json:"message"`
	Code    *clippyCode  `json:"code"`
	Spans   []clippySpan `json:"spans"`
}

type clippyLine struct {
	Reason  string        `json:"reason"`
	Message clippyMessage `json:"message"`
}

type ruleEntry struct {
	severity string
	message  string
	fileLines map[string][]int
}

func severityLabel(level string) string {
	switch level {
	case "error":
		return "ERROR"
	case "warning":
		return "WARN"
	default:
		return "INFO"
	}
}

func main() {
	formatter.RunWithConfig("cargo-clippy", format)
}

func format(data []byte, cfg formatter.Config) error {
	lines, err := formatter.ParseNDJSON(data)
	if err != nil {
		return fmt.Errorf("invalid NDJSON: %w", err)
	}

	var items []clippyLine
	for _, raw := range lines {
		var cl clippyLine
		if err := json.Unmarshal(raw, &cl); err != nil {
			continue
		}
		if cl.Reason != "compiler-message" {
			continue
		}
		// skip note and help levels
		if cl.Message.Level == "note" || cl.Message.Level == "help" {
			continue
		}
		items = append(items, cl)
	}

	if cfg.GroupBy == "file" {
		return formatByFile(items, cfg)
	}
	return formatByRule(items, cfg)
}

func ruleCode(cl clippyLine) string {
	if cl.Message.Code != nil && cl.Message.Code.Code != "" {
		return cl.Message.Code.Code
	}
	return "rustc"
}

func primarySpan(cl clippyLine) (string, int) {
	for _, s := range cl.Message.Spans {
		if s.IsPrimary {
			return s.FileName, s.LineStart
		}
	}
	if len(cl.Message.Spans) > 0 {
		return cl.Message.Spans[0].FileName, cl.Message.Spans[0].LineStart
	}
	return "", 0
}

func formatByRule(items []clippyLine, cfg formatter.Config) error {
	rules := map[string]*ruleEntry{}
	ruleOrder := []string{}

	for _, cl := range items {
		rule := ruleCode(cl)
		rawFile, line := primarySpan(cl)
		file := rawFile
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
			r.message = truncate(cl.Message.Message, cfg.MaxMessageLength)
			r.severity = cl.Message.Level
		}
		if line > 0 {
			r.fileLines[file] = append(r.fileLines[file], line)
		} else {
			// ensure file appears even without a line
			if _, ok := r.fileLines[file]; !ok {
				r.fileLines[file] = nil
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
	for _, cl := range items {
		f, _ := primarySpan(cl)
		if f != "" {
			totalFiles[formatter.ResolvePath(f, cfg)] = struct{}{}
		}
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

		files := make([]string, 0, len(r.fileLines))
		for f := range r.fileLines {
			files = append(files, f)
		}
		sort.Strings(files)

		for _, f := range files {
			if !formatter.MatchesFileFilter(f, cfg.OnlyFiles) {
				continue
			}
			ls := r.fileLines[f]
			sort.Ints(ls)
			if len(ls) == 0 {
				fmt.Printf("  - %s\n", f)
			} else {
				fmt.Printf("  - %s — %s\n", f, formatLines(ls))
			}
		}
		fmt.Println("────────────────────────────────────────────────")
	}

	fmt.Println(formatter.Summary(displayedIssues, len(ruleOrder), len(totalFiles)))
	return nil
}

func formatByFile(items []clippyLine, cfg formatter.Config) error {
	type lineEntry struct {
		rule    string
		line    int
		message string
	}
	fileMap := map[string][]lineEntry{}
	fileOrder := []string{}
	allRules := map[string]struct{}{}

	for _, cl := range items {
		rule := ruleCode(cl)
		rawFile, line := primarySpan(cl)
		file := rawFile
		if file == "" {
			file = "unknown"
		} else {
			file = formatter.ResolvePath(file, cfg)
		}
		allRules[rule] = struct{}{}
		if _, ok := fileMap[file]; !ok {
			fileOrder = append(fileOrder, file)
		}
		fileMap[file] = append(fileMap[file], lineEntry{rule: rule, line: line, message: truncate(cl.Message.Message, cfg.MaxMessageLength)})
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

func formatLines(ls []int) string {
	if len(ls) == 1 {
		return fmt.Sprintf("line %d", ls[0])
	}
	parts := make([]string, len(ls))
	for i, l := range ls {
		parts[i] = fmt.Sprintf("%d", l)
	}
	return "lines " + strings.Join(parts, ", ")
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
