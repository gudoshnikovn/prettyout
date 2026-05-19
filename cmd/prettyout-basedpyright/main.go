package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type position struct {
	Line int `json:"line"`
}

type diagRange struct {
	Start position `json:"start"`
	End   position `json:"end"`
}

type diagnostic struct {
	Rule     string    `json:"rule"`
	File     string    `json:"file"`
	Severity string    `json:"severity"`
	Message  string    `json:"message"`
	Range    diagRange `json:"range"`
}

type report struct {
	GeneralDiagnostics []diagnostic `json:"generalDiagnostics"`
}

func main() {
	formatter.RunWithConfig("basedpyright", format)
}

const divider = "────────────────────────────────────────────────"

// severityRank returns a numeric rank for comparing severities (higher = more severe).
func severityRank(sev string) int {
	switch sev {
	case "error", "ERROR", "fatal", "FATAL":
		return 3
	case "warning", "WARNING", "warn", "WARN":
		return 2
	case "information", "info", "INFO":
		return 1
	default:
		return 0
	}
}

// severityLabel returns the bracket label for a severity string.
func severityLabel(sev string) string {
	switch sev {
	case "error", "ERROR", "fatal", "FATAL":
		return "ERROR"
	case "warning", "WARNING", "warn", "WARN":
		return "WARN"
	case "information", "info", "INFO":
		return "INFO"
	default:
		upper := ""
		for _, c := range sev {
			if c >= 'a' && c <= 'z' {
				upper += string(c - 32)
			} else {
				upper += string(c)
			}
		}
		return upper
	}
}

func format(data []byte, cfg formatter.Config) error {
	var r report
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if cfg.GroupBy == "file" {
		return formatByFile(r.GeneralDiagnostics, cfg)
	}
	return formatByRule(r.GeneralDiagnostics, cfg)
}

type ruleGroup struct {
	message     string
	maxSeverity string
	files       map[string]map[int]struct{} // path -> set of lines (deduped)
	fileOrder   []string
}

func formatByRule(diags []diagnostic, cfg formatter.Config) error {
	rules := map[string]*ruleGroup{}
	var ruleOrder []string

	for _, d := range diags {
		code := d.Rule
		if code == "" {
			code = "StaticAnalysis"
		}
		path := formatter.ResolvePath(d.File, cfg)
		line := d.Range.Start.Line + 1

		if _, ok := rules[code]; !ok {
			rules[code] = &ruleGroup{files: map[string]map[int]struct{}{}}
			ruleOrder = append(ruleOrder, code)
		}
		rg := rules[code]
		if rg.message == "" {
			rg.message = truncate(d.Message, cfg.MaxMessageLength)
		}
		if severityRank(d.Severity) > severityRank(rg.maxSeverity) {
			rg.maxSeverity = d.Severity
		}
		if _, ok := rg.files[path]; !ok {
			rg.fileOrder = append(rg.fileOrder, path)
			rg.files[path] = map[int]struct{}{}
		}
		rg.files[path][line] = struct{}{}
	}

	ruleCounts := make(map[string]int, len(ruleOrder))
	for _, code := range ruleOrder {
		rg := rules[code]
		n := 0
		for _, lines := range rg.files {
			n += len(lines)
		}
		ruleCounts[code] = n
	}
	ruleOrder = formatter.FilterRuleOrder(ruleOrder, cfg.OnlyRules)
	ruleOrder = formatter.SortOrder(ruleOrder, ruleCounts, cfg.Sort)

	totalFiles := countDistinctFiles(diags, cfg)

	for _, code := range ruleOrder {
		rg := rules[code]

		hasMatchingFile := false
		for _, path := range rg.fileOrder {
			if formatter.MatchesFileFilter(path, cfg.OnlyFiles) {
				hasMatchingFile = true
				break
			}
		}
		if !hasMatchingFile {
			continue
		}

		count := 0
		for _, lines := range rg.files {
			count += len(lines)
		}

		sevLabel := severityLabel(rg.maxSeverity)
		sevPrefix := fmt.Sprintf("[%s] ", sevLabel)
		if cfg.Colors {
			color := formatter.SeverityColor(rg.maxSeverity, true)
			reset := "\033[0m"
			if color != "" {
				sevPrefix = fmt.Sprintf("%s[%s]%s ", color, sevLabel, reset)
			}
			fmt.Printf("%s\033[1;34m%s\033[0m (%d) — \033[1m%s\033[0m\n", sevPrefix, code, count, rg.message)
		} else {
			fmt.Printf("%s%s (%d) — %s\n", sevPrefix, code, count, rg.message)
		}
		fmt.Println("Affected files:")

		sortedFiles := make([]string, len(rg.fileOrder))
		copy(sortedFiles, rg.fileOrder)
		sort.Strings(sortedFiles)

		for _, path := range sortedFiles {
			if !formatter.MatchesFileFilter(path, cfg.OnlyFiles) {
				continue
			}
			lineSet := rg.files[path]
			lines := make([]int, 0, len(lineSet))
			for l := range lineSet {
				lines = append(lines, l)
			}
			sort.Ints(lines)
			fmt.Printf("  - %s — %s\n", path, formatLines(lines))
		}
		fmt.Println(divider)
	}

	fmt.Println(formatter.Summary(len(diags), len(ruleOrder), totalFiles))
	return nil
}

type issueEntry struct {
	code     string
	severity string
	message  string
	line     int
}

func formatByFile(diags []diagnostic, cfg formatter.Config) error {
	type fileGroup struct {
		path   string
		issues []issueEntry
	}

	files := map[string]*fileGroup{}
	var fileOrder []string

	for _, d := range diags {
		code := d.Rule
		if code == "" {
			code = "StaticAnalysis"
		}
		path := formatter.ResolvePath(d.File, cfg)
		line := d.Range.Start.Line + 1

		if _, ok := files[path]; !ok {
			files[path] = &fileGroup{path: path}
			fileOrder = append(fileOrder, path)
		}
		files[path].issues = append(files[path].issues, issueEntry{
			code:     code,
			severity: d.Severity,
			message:  truncate(d.Message, cfg.MaxMessageLength),
			line:     line,
		})
	}

	filteredFileOrder := fileOrder[:0:0]
	for _, f := range fileOrder {
		if formatter.MatchesFileFilter(f, cfg.OnlyFiles) {
			filteredFileOrder = append(filteredFileOrder, f)
		}
	}
	fileOrder = filteredFileOrder
	sort.Strings(fileOrder)

	ruleSeen := map[string]struct{}{}
	for _, d := range diags {
		code := d.Rule
		if code == "" {
			code = "StaticAnalysis"
		}
		ruleSeen[code] = struct{}{}
	}

	for _, path := range fileOrder {
		fg := files[path]
		sort.Slice(fg.issues, func(i, j int) bool {
			return fg.issues[i].line < fg.issues[j].line
		})

		// Pre-filter by OnlyRules
		filteredIssues := fg.issues[:0:0]
		for _, e := range fg.issues {
			if len(cfg.OnlyRules) > 0 {
				found := false
				for _, r := range cfg.OnlyRules {
					if e.code == r {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			filteredIssues = append(filteredIssues, e)
		}
		if len(filteredIssues) == 0 {
			continue
		}

		fmt.Printf("%s — %d %s\n", fg.path, len(filteredIssues), formatter.Plural(len(filteredIssues), "issue", "issues"))

		lastCode := ""
		for _, e := range filteredIssues {
			sevLabel := severityLabel(e.severity)
			sevPrefix := fmt.Sprintf("[%s] ", sevLabel)
			if cfg.Colors {
				color := formatter.SeverityColor(e.severity, true)
				reset := "\033[0m"
				if color != "" {
					sevPrefix = fmt.Sprintf("%s[%s]%s ", color, sevLabel, reset)
				}
			}

			msg := ""
			if e.code != lastCode {
				msg = "  — " + e.message
				lastCode = e.code
			}

			if cfg.Colors {
				fmt.Printf("  %s\033[1;34m%s\033[0m  line %d%s\n", sevPrefix, e.code, e.line, msg)
			} else {
				fmt.Printf("  %s%s  line %d%s\n", sevPrefix, e.code, e.line, msg)
			}
		}
		fmt.Println(divider)
	}

	fmt.Println(formatter.Summary(len(diags), len(ruleSeen), len(fileOrder)))
	return nil
}

func countDistinctFiles(diags []diagnostic, cfg formatter.Config) int {
	seen := map[string]struct{}{}
	for _, d := range diags {
		seen[formatter.ResolvePath(d.File, cfg)] = struct{}{}
	}
	return len(seen)
}

func formatLines(lines []int) string {
	if len(lines) == 1 {
		return fmt.Sprintf("line %d", lines[0])
	}
	parts := make([]string, len(lines))
	for i, l := range lines {
		parts[i] = fmt.Sprintf("%d", l)
	}
	return "lines " + strings.Join(parts, ", ")
}

func truncate(s string, max int) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
