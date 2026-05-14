package main

import (
	"encoding/json"
	"fmt"
	"sort"

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
	items   map[string]struct{} // "filename|line_info"
	order   map[string]string   // key -> display line_info
}

func main() {
	formatter.Run(format)
}

func format(data []byte) error {
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

		filename := iss.Filename
		if idx := findIndex(filename, "backend/"); idx >= 0 {
			filename = filename[idx+len("backend/"):]
		} else if idx := findLastSlash(filename); idx >= 0 {
			filename = filename[idx+1:]
		}

		lineInfo := fmt.Sprintf("line %d", iss.Location.Row)
		if iss.Location.Row != iss.EndLocation.Row {
			lineInfo = fmt.Sprintf("lines %d-%d", iss.Location.Row, iss.EndLocation.Row)
		}

		key := filename + "\x00" + lineInfo

		if _, ok := rules[code]; !ok {
			rules[code] = &ruleInfo{items: map[string]struct{}{}, order: map[string]string{}}
		}
		r := rules[code]
		if r.message == "" {
			r.message = iss.Message
		}
		r.items[key] = struct{}{}
		r.order[key] = lineInfo
	}

	codes := make([]string, 0, len(rules))
	for code := range rules {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	for _, code := range codes {
		r := rules[code]
		fmt.Printf("\033[1;33m%s\033[0m - \033[1m%s\033[0m\n", code, r.message)
		fmt.Println("Affected files:")

		keys := make([]string, 0, len(r.items))
		for k := range r.items {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			sep := findNull(key)
			file := key[:sep]
			line := key[sep+1:]
			fmt.Printf("  - %s — %s\n", file, line)
		}
		fmt.Println("----------------------------------------")
	}
	return nil
}

func findIndex(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func findLastSlash(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			return i
		}
	}
	return -1
}

func findNull(s string) int {
	for i, c := range s {
		if c == 0 {
			return i
		}
	}
	return len(s)
}
