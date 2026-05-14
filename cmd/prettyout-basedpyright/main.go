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
	formatter.Run(format)
}

func format(data []byte) error {
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

		filepath := d.File
		if idx := strings.Index(filepath, "backend/"); idx >= 0 {
			filepath = filepath[idx+len("backend/"):]
		} else if idx := strings.LastIndex(filepath, "/"); idx >= 0 {
			filepath = filepath[idx+1:]
		}

		start := d.Range.Start.Line + 1
		end := d.Range.End.Line + 1
		lineInfo := fmt.Sprintf("line %d", start)
		if start != end {
			lineInfo = fmt.Sprintf("lines %d-%d", start, end)
		}

		key := filepath + "\x00" + lineInfo

		if _, ok := rules[code]; !ok {
			rules[code] = &ruleInfo{items: map[string]struct{}{}}
		}
		ri := rules[code]
		if ri.message == "" {
			ri.message = d.Message
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
		msg := ri.message
		if len(msg) > 100 {
			msg = msg[:100] + "..."
		}
		fmt.Printf("\033[1;34m%s\033[0m - \033[1m%s\033[0m\n", code, msg)
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
