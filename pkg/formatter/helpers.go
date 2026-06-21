package formatter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const ansiReset = "\033[0m"

// ResolvePath returns a display-friendly path:
// - if basename_only extra config is true: returns filepath.Base(path)
// - if path is absolute: returns relative from CWD (falls back to path on error)
// - if path is already relative: returns as-is
func ResolvePath(path string, cfg Config) string {
	if bn, ok := cfg.Extra["basename_only"]; ok {
		if v, ok := bn.(bool); ok && v {
			return filepath.Base(path)
		}
	}
	if filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return path
		}
		rel, err := filepath.Rel(cwd, path)
		if err != nil {
			return path
		}
		return rel
	}
	return path
}

// SeverityColor returns the opening ANSI escape for the given severity string.
// Returns empty string if colors is false or severity is unknown.
func SeverityColor(sev string, colors bool) string {
	if !colors {
		return ""
	}
	switch sev {
	case "error", "ERROR", "fatal", "FATAL":
		return "\033[1;31m"
	case "warning", "WARNING", "WARN", "warn":
		return "\033[1;33m"
	case "information", "info", "INFO", "note", "NOTE", "style", "hint":
		return "\033[1;34m"
	default:
		return ""
	}
}

// ParseNDJSON parses newline-delimited JSON (one JSON object per line).
// Skips blank lines and lines that aren't valid JSON.
func ParseNDJSON(data []byte) ([]json.RawMessage, error) {
	var results []json.RawMessage
	start := 0
	for i := 0; i <= len(data); i++ {
		if i == len(data) || data[i] == '\n' {
			line := data[start:i]
			start = i + 1
			// trim whitespace
			trimmed := trimSpace(line)
			if len(trimmed) == 0 {
				continue
			}
			if !json.Valid(trimmed) {
				continue
			}
			results = append(results, json.RawMessage(trimmed))
		}
	}
	return results, nil
}

// Plural returns singular when n==1, plural otherwise.
func Plural(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

// Summary returns a standard "N issues · M rules · K files" summary line.
func Summary(issues, rules, files int) string {
	return fmt.Sprintf("%d %s · %d %s · %d %s",
		issues, Plural(issues, "issue", "issues"),
		rules, Plural(rules, "rule", "rules"),
		files, Plural(files, "file", "files"),
	)
}

func trimSpace(b []byte) []byte {
	start := 0
	end := len(b)
	for start < end && (b[start] == ' ' || b[start] == '\t' || b[start] == '\r') {
		start++
	}
	for end > start && (b[end-1] == ' ' || b[end-1] == '\t' || b[end-1] == '\r') {
		end--
	}
	return b[start:end]
}

// SortOrder sorts a slice of rule/file names by the given strategy.
// sortBy: "" or "alpha" → alphabetical; "count" → descending by count, alpha tiebreak.
// Returns a new sorted slice; does not modify the input.
func SortOrder(order []string, counts map[string]int, sortBy string) []string {
	out := make([]string, len(order))
	copy(out, order)
	if sortBy == "count" {
		sort.Slice(out, func(i, j int) bool {
			ci, cj := counts[out[i]], counts[out[j]]
			if ci != cj {
				return ci > cj // descending
			}
			return out[i] < out[j] // alpha tiebreak
		})
	} else {
		sort.Strings(out)
	}
	return out
}

// FilterRuleOrder returns only the rules that are in onlyRules.
// If onlyRules is empty, returns order unchanged.
func FilterRuleOrder(order []string, onlyRules []string) []string {
	if len(onlyRules) == 0 {
		return order
	}
	allowed := make(map[string]bool, len(onlyRules))
	for _, r := range onlyRules {
		allowed[r] = true
	}
	out := order[:0:0] // empty slice, same backing if reused
	for _, r := range order {
		if allowed[r] {
			out = append(out, r)
		}
	}
	return out
}

// MatchesFileFilter returns true if path starts with any prefix in onlyFiles.
// If onlyFiles is empty, returns true (no filter).
func MatchesFileFilter(path string, onlyFiles []string) bool {
	if len(onlyFiles) == 0 {
		return true
	}
	for _, prefix := range onlyFiles {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// PrintStats prints a compact per-rule count table sorted by count descending,
// followed by a summary line. ruleFiles may be nil to suppress the files column.
func PrintStats(
	ruleOrder []string,
	ruleCounts map[string]int,
	ruleFiles map[string]int,
	ruleMessages map[string]string,
	totalFiles int,
	cfg Config,
) {
	sorted := make([]string, len(ruleOrder))
	copy(sorted, ruleOrder)
	sort.Slice(sorted, func(i, j int) bool {
		ci, cj := ruleCounts[sorted[i]], ruleCounts[sorted[j]]
		if ci != cj {
			return ci > cj
		}
		return sorted[i] < sorted[j]
	})

	totalIssues := 0
	maxCount := 0
	maxRuleLen := 0
	for _, r := range sorted {
		c := ruleCounts[r]
		totalIssues += c
		if c > maxCount {
			maxCount = c
		}
		if len(r) > maxRuleLen {
			maxRuleLen = len(r)
		}
	}
	countWidth := max(len(fmt.Sprintf("%d", maxCount)), len("COUNT"))

	var fileNumWidth, fileColWidth int
	if ruleFiles != nil {
		maxFC := 0
		for _, r := range sorted {
			if fc := ruleFiles[r]; fc > maxFC {
				maxFC = fc
			}
		}
		fileNumWidth = len(fmt.Sprintf("%d", maxFC))
		fileColWidth = max(fileNumWidth+6, len("FILES")) // N + space + "files"(5)
	}

	if ruleFiles != nil {
		fmt.Printf("%*s  %-*s  %-*s  %s\n", countWidth, "COUNT", maxRuleLen, "RULE", fileColWidth, "FILES", "DESCRIPTION")
	} else {
		fmt.Printf("%*s  %-*s  %s\n", countWidth, "COUNT", maxRuleLen, "RULE", "DESCRIPTION")
	}

	for _, r := range sorted {
		count := ruleCounts[r]
		msg := ruleMessages[r]
		if ruleFiles != nil {
			fc := ruleFiles[r]
			fileStr := fmt.Sprintf("%*d %s", fileNumWidth, fc, Plural(fc, "file", "files"))
			if msg != "" {
				fmt.Printf("%*d  %-*s  %-*s  — %s\n", countWidth, count, maxRuleLen, r, fileColWidth, fileStr, msg)
			} else {
				fmt.Printf("%*d  %-*s  %s\n", countWidth, count, maxRuleLen, r, fileStr)
			}
		} else {
			if msg != "" {
				fmt.Printf("%*d  %-*s  — %s\n", countWidth, count, maxRuleLen, r, msg)
			} else {
				fmt.Printf("%*d  %s\n", countWidth, count, r)
			}
		}
	}

	fmt.Println()
	fmt.Println(Summary(totalIssues, len(sorted), totalFiles))
}
