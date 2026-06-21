// prettyout-example is a reference plugin demonstrating the full plugin pattern.
//
// Copy this file, rename the package, and adapt the JSON structs and
// grouping logic for your tool. The helpers in pkg/formatter handle all
// the display details so you only need to write the parsing and grouping.
//
// Build and test:
//
//	go build -o /tmp/prettyout-example ./cmd/prettyout-example/
//	your-tool --json ... | /tmp/prettyout-example
//
// Register in your config to intercept the tool automatically:
//
//	custom_tools:
//	  mytool:
//	    plugin: prettyout-example
//	    output_args: [--json]
package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

// --- JSON structs: adapt these to your tool's output format ---

// toolIssue mirrors one entry from your tool's JSON output.
// Fields vary per tool; use `json:",omitempty"` for optional ones.
type toolIssue struct {
	RuleID   string `json:"rule_id"`  // the rule/check that fired
	Message  string `json:"message"`  // human-readable description
	Severity string `json:"severity"` // "error", "warning", "info", etc.
	File     string `json:"file"`     // path to the file (absolute or relative)
	Line     int    `json:"line"`     // 1-based line number
}

// --- Entry point ---

func main() {
	// RunWithConfig reads stdin, loads the tool's section from config
	// (~/.config/prettyout/config.yaml and .prettyout.yaml), and calls
	// your format function. On error it prints to stderr and exits 1.
	formatter.RunWithConfig("example", format)
}

// --- Core formatter ---

func format(data []byte, cfg formatter.Config) error {
	// Step 1: parse your tool's JSON.
	var issues []toolIssue
	if err := json.Unmarshal(data, &issues); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Step 2: dispatch on group_by.
	// "rule" (default) groups by rule code; "file" groups by file path.
	if cfg.GroupBy == "file" {
		return formatByFile(issues, cfg)
	}
	return formatByRule(issues, cfg)
}

// --- Group by rule (default) ---

type ruleEntry struct {
	severity  string
	message   string
	fileLines map[string][]int // resolved path → line numbers
}

func formatByRule(issues []toolIssue, cfg formatter.Config) error {
	rules := map[string]*ruleEntry{}
	var ruleOrder []string

	for _, iss := range issues {
		rule := iss.RuleID
		if rule == "" {
			rule = "unknown"
		}
		// ResolvePath converts absolute paths to relative-from-CWD.
		// It also respects cfg.Extra["basename_only"] if you set it.
		path := formatter.ResolvePath(iss.File, cfg)

		if _, ok := rules[rule]; !ok {
			rules[rule] = &ruleEntry{fileLines: map[string][]int{}}
			ruleOrder = append(ruleOrder, rule)
		}
		r := rules[rule]
		if r.message == "" {
			// Truncate caps message length at cfg.MaxMessageLength (0 = unlimited).
			r.message = formatter.Truncate(iss.Message, cfg.MaxMessageLength)
			r.severity = iss.Severity
		}
		r.fileLines[path] = append(r.fileLines[path], iss.Line)
	}

	// Count occurrences per rule (for sorting and display).
	ruleCounts := make(map[string]int, len(ruleOrder))
	for _, rule := range ruleOrder {
		n := 0
		for _, lines := range rules[rule].fileLines {
			n += len(lines)
		}
		ruleCounts[rule] = n
	}

	// Apply cfg.OnlyRules filter, then sort.
	// SortOrder supports "" (alpha), "alpha", or "count" (descending).
	ruleOrder = formatter.FilterRuleOrder(ruleOrder, cfg.OnlyRules)
	ruleOrder = formatter.SortOrder(ruleOrder, ruleCounts, cfg.Sort)

	totalFiles := countDistinctFiles(issues, cfg)

	// Stats mode: compact count table, useful for large codebases.
	if cfg.Stats {
		ruleFileCount := make(map[string]int, len(ruleOrder))
		ruleMessages := make(map[string]string, len(ruleOrder))
		for _, rule := range ruleOrder {
			r := rules[rule]
			ruleFileCount[rule] = len(r.fileLines)
			ruleMessages[rule] = r.message
		}
		formatter.PrintStats(ruleOrder, ruleCounts, ruleFileCount, ruleMessages, totalFiles, cfg)
		return nil
	}

	displayedIssues := 0
	for _, rule := range ruleOrder {
		r := rules[rule]

		// Skip rules with no files matching the OnlyFiles filter.
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
		displayedIssues += count

		// Rule header: [SEVERITY] rule-name (N) — message
		col := formatter.SeverityColor(r.severity, cfg.Colors)
		reset := ""
		if cfg.Colors {
			reset = "\033[0m"
		}
		label := formatter.SeverityLabel(r.severity)
		if cfg.Colors {
			fmt.Printf("%s[%s]%s %s (%d) — %s\n", col, label, reset, rule, count, r.message)
		} else {
			fmt.Printf("[%s] %s (%d) — %s\n", label, rule, count, r.message)
		}
		fmt.Println("Affected files:")

		// Sort files alphabetically within the rule group.
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
			// FormatLines produces "line 5" or "lines 5, 12, 30".
			fmt.Printf("  - %s — %s\n", f, formatter.FormatLines(ls))
		}
		fmt.Println(formatter.Divider)
	}

	fmt.Println(formatter.Summary(displayedIssues, len(ruleOrder), totalFiles))
	return nil
}

// --- Group by file ---

func formatByFile(issues []toolIssue, cfg formatter.Config) error {
	type lineEntry struct {
		rule    string
		line    int
		message string
	}

	fileMap := map[string][]lineEntry{}
	var fileOrder []string
	allRules := map[string]struct{}{}

	for _, iss := range issues {
		rule := iss.RuleID
		if rule == "" {
			rule = "unknown"
		}
		allRules[rule] = struct{}{}
		path := formatter.ResolvePath(iss.File, cfg)
		if _, ok := fileMap[path]; !ok {
			fileOrder = append(fileOrder, path)
		}
		fileMap[path] = append(fileMap[path], lineEntry{
			rule:    rule,
			line:    iss.Line,
			message: formatter.Truncate(iss.Message, cfg.MaxMessageLength),
		})
	}

	// Apply OnlyFiles filter, then sort.
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

		// Apply OnlyRules filter.
		var kept []lineEntry
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
			kept = append(kept, e)
		}
		if len(kept) == 0 {
			continue
		}

		sort.Slice(kept, func(i, j int) bool { return kept[i].line < kept[j].line })
		fmt.Printf("%s — %d %s\n", f, len(kept), formatter.Plural(len(kept), "issue", "issues"))

		prevRule := ""
		for _, e := range kept {
			msg := ""
			if e.rule != prevRule {
				msg = " — " + e.message
				prevRule = e.rule
			}
			fmt.Printf("  %s  line %d%s\n", e.rule, e.line, msg)
		}
		fmt.Println(formatter.Divider)
		totalIssues += len(kept)
	}

	fmt.Println(formatter.Summary(totalIssues, len(allRules), len(fileOrder)))
	return nil
}

// --- Helpers ---

func countDistinctFiles(issues []toolIssue, cfg formatter.Config) int {
	seen := map[string]struct{}{}
	for _, iss := range issues {
		seen[formatter.ResolvePath(iss.File, cfg)] = struct{}{}
	}
	return len(seen)
}

