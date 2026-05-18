package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type location struct {
	Row int `json:"row"`
}

type fixInfo struct {
	Applicability string `json:"applicability"`
}

type issue struct {
	Code        string   `json:"code"`
	Message     string   `json:"message"`
	Filename    string   `json:"filename"`
	Location    location `json:"location"`
	EndLocation location `json:"end_location"`
	Fix         *fixInfo `json:"fix"`
}

func main() {
	formatter.RunWithConfig("ruff", format)
}

const divider = "────────────────────────────────────────────────"

func format(data []byte, cfg formatter.Config) error {
	var issues []issue
	if err := json.Unmarshal(data, &issues); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if cfg.GroupBy == "file" {
		return formatByFile(issues, cfg)
	}
	return formatByRule(issues, cfg)
}

// fileLines tracks lines per file for a rule group.
type fileLines struct {
	file  string
	lines []int
}

type ruleGroup struct {
	message string
	files   map[string]*fileLines // keyed by resolved path
	order   []string              // insertion order for files
}

func formatByRule(issues []issue, cfg formatter.Config) error {
	rules := map[string]*ruleGroup{}
	var ruleOrder []string

	for _, iss := range issues {
		code := iss.Code
		if code == "" {
			code = "UNKNOWN"
		}
		path := formatter.ResolvePath(iss.Filename, cfg)

		if _, ok := rules[code]; !ok {
			rules[code] = &ruleGroup{files: map[string]*fileLines{}}
			ruleOrder = append(ruleOrder, code)
		}
		rg := rules[code]
		if rg.message == "" {
			rg.message = truncate(iss.Message, cfg.MaxMessageLength)
		}
		if _, ok := rg.files[path]; !ok {
			rg.files[path] = &fileLines{file: path}
			rg.order = append(rg.order, path)
		}
		rg.files[path].lines = append(rg.files[path].lines, iss.Location.Row)
	}

	sort.Strings(ruleOrder)

	totalIssues := len(issues)
	totalFiles := countDistinctFiles(issues, cfg)

	for _, code := range ruleOrder {
		rg := rules[code]
		count := 0
		for _, fl := range rg.files {
			count += len(fl.lines)
		}

		if cfg.Colors {
			fmt.Printf("\033[1;33m%s\033[0m (%d) — \033[1m%s\033[0m\n", code, count, rg.message)
		} else {
			fmt.Printf("%s (%d) — %s\n", code, count, rg.message)
		}
		fmt.Println("Affected files:")

		// Sort files alphabetically
		sortedFiles := make([]string, len(rg.order))
		copy(sortedFiles, rg.order)
		sort.Strings(sortedFiles)

		for _, path := range sortedFiles {
			fl := rg.files[path]
			sort.Ints(fl.lines)
			lineLabel := formatLines(fl.lines)
			fmt.Printf("  - %s — %s\n", fl.file, lineLabel)
		}
		fmt.Println(divider)
	}

	fmt.Println(formatter.Summary(totalIssues, len(ruleOrder), totalFiles))
	printFixHint(issues)
	return nil
}

// issueEntry holds a parsed issue with resolved path for file-mode grouping.
type issueEntry struct {
	path    string
	code    string
	message string
	line    int
}

func formatByFile(issues []issue, cfg formatter.Config) error {
	type fileGroup struct {
		path   string
		issues []issueEntry
	}

	files := map[string]*fileGroup{}
	var fileOrder []string

	for _, iss := range issues {
		path := formatter.ResolvePath(iss.Filename, cfg)
		code := iss.Code
		if code == "" {
			code = "UNKNOWN"
		}
		if _, ok := files[path]; !ok {
			files[path] = &fileGroup{path: path}
			fileOrder = append(fileOrder, path)
		}
		files[path].issues = append(files[path].issues, issueEntry{
			path:    path,
			code:    code,
			message: truncate(iss.Message, cfg.MaxMessageLength),
			line:    iss.Location.Row,
		})
	}

	sort.Strings(fileOrder)

	// Count distinct rules for summary
	ruleSeen := map[string]struct{}{}
	for _, iss := range issues {
		code := iss.Code
		if code == "" {
			code = "UNKNOWN"
		}
		ruleSeen[code] = struct{}{}
	}

	for _, path := range fileOrder {
		fg := files[path]
		// Sort by line number
		sort.Slice(fg.issues, func(i, j int) bool {
			return fg.issues[i].line < fg.issues[j].line
		})

		fmt.Printf("%s — %d %s\n", fg.path, len(fg.issues), formatter.Plural(len(fg.issues), "issue", "issues"))

		// Track last seen rule to suppress repeated messages
		lastCode := ""
		for _, e := range fg.issues {
			msg := ""
			if e.code != lastCode {
				msg = "  — " + e.message
				lastCode = e.code
			}
			if cfg.Colors {
				fmt.Printf("  \033[1;33m%s\033[0m  line %d%s\n", e.code, e.line, msg)
			} else {
				fmt.Printf("  %s  line %d%s\n", e.code, e.line, msg)
			}
		}
		fmt.Println(divider)
	}

	fmt.Println(formatter.Summary(len(issues), len(ruleSeen), len(fileOrder)))
	printFixHint(issues)
	return nil
}

func printFixHint(issues []issue) {
	safe, unsafe := 0, 0
	for _, iss := range issues {
		if iss.Fix == nil {
			continue
		}
		if strings.EqualFold(iss.Fix.Applicability, "unsafe") {
			unsafe++
		} else {
			safe++
		}
	}
	if safe > 0 && unsafe > 0 {
		fmt.Printf("  ↳ %d fixable with --fix (%d more with --unsafe-fixes)\n", safe, unsafe)
	} else if safe > 0 {
		fmt.Printf("  ↳ %d fixable with --fix\n", safe)
	} else if unsafe > 0 {
		fmt.Printf("  ↳ %d fixable with --unsafe-fixes\n", unsafe)
	}
}

func countDistinctFiles(issues []issue, cfg formatter.Config) int {
	seen := map[string]struct{}{}
	for _, iss := range issues {
		seen[formatter.ResolvePath(iss.Filename, cfg)] = struct{}{}
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
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
