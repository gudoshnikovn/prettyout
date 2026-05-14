package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gudoshnikov_na/prettyout/pkg/formatter"
)

type location struct {
	Row int `json:"row"`
}

type issue struct {
	Code        string   `json:"code"`
	Message     string   `json:"message"`
	Filename    string   `json:"filename"`
	Location    location `json:"location"`
	EndLocation location `json:"end_location"`
}

type ruleInfo struct {
	message string
	items   map[string]string // key (file\x00lineInfo) -> display lineInfo
}

func main() {
	formatter.RunWithConfig("ruff", format)
}

func format(data []byte, cfg formatter.Config) error {
	var issues []issue
	if err := json.Unmarshal(data, &issues); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	rules := map[string]*ruleInfo{}

	for _, iss := range issues {
		code := iss.Code
		if code == "" {
			code = "UNKNOWN"
		}

		filename := filepath.Base(iss.Filename)

		lineInfo := fmt.Sprintf("line %d", iss.Location.Row)
		if iss.Location.Row != iss.EndLocation.Row {
			lineInfo = fmt.Sprintf("lines %d-%d", iss.Location.Row, iss.EndLocation.Row)
		}

		key := filename + "\x00" + lineInfo

		if _, ok := rules[code]; !ok {
			rules[code] = &ruleInfo{items: map[string]string{}}
		}
		r := rules[code]
		if r.message == "" {
			r.message = truncate(iss.Message, cfg.MaxMessageLength)
		}
		r.items[key] = lineInfo
	}

	codes := make([]string, 0, len(rules))
	for code := range rules {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	for _, code := range codes {
		r := rules[code]
		if cfg.Colors {
			fmt.Printf("\033[1;33m%s\033[0m - \033[1m%s\033[0m\n", code, r.message)
		} else {
			fmt.Printf("%s - %s\n", code, r.message)
		}
		fmt.Println("Affected files:")

		keys := make([]string, 0, len(r.items))
		for k := range r.items {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			sep := strings.IndexByte(key, 0)
			file := key[:sep]
			line := key[sep+1:]
			fmt.Printf("  - %s — %s\n", file, line)
		}
		fmt.Println("----------------------------------------")
	}
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
