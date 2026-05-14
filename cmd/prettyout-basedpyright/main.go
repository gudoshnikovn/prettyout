package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
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
	Rule    string    `json:"rule"`
	File    string    `json:"file"`
	Message string    `json:"message"`
	Range   diagRange `json:"range"`
}

type report struct {
	GeneralDiagnostics []diagnostic `json:"generalDiagnostics"`
}

type ruleInfo struct {
	message string
	items   map[string]struct{}
}

func main() {
	formatter.RunWithConfig("basedpyright", format)
}

func format(data []byte, cfg formatter.Config) error {
	var r report
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	rules := map[string]*ruleInfo{}

	for _, d := range r.GeneralDiagnostics {
		code := d.Rule
		if code == "" {
			code = "StaticAnalysis"
		}

		filename := filepath.Base(d.File)

		start := d.Range.Start.Line + 1
		end := d.Range.End.Line + 1
		lineInfo := fmt.Sprintf("line %d", start)
		if start != end {
			lineInfo = fmt.Sprintf("lines %d-%d", start, end)
		}

		key := filename + "\x00" + lineInfo

		if _, ok := rules[code]; !ok {
			rules[code] = &ruleInfo{items: map[string]struct{}{}}
		}
		ri := rules[code]
		if ri.message == "" {
			ri.message = truncate(d.Message, cfg.MaxMessageLength)
		}
		ri.items[key] = struct{}{}
	}

	codes := make([]string, 0, len(rules))
	for code := range rules {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	for _, code := range codes {
		ri := rules[code]
		if cfg.Colors {
			fmt.Printf("\033[1;34m%s\033[0m - \033[1m%s\033[0m\n", code, ri.message)
		} else {
			fmt.Printf("%s - %s\n", code, ri.message)
		}
		fmt.Println("Affected files:")

		keys := make([]string, 0, len(ri.items))
		for k := range ri.items {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			sep := strings.IndexByte(key, 0)
			file := key[:sep]
			line := key[sep+1:]
			fmt.Printf("  - %s — %s\n", file, line)
		}
		fmt.Println("--------------------------------------------------")
	}
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
