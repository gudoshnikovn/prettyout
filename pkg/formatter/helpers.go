package formatter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
