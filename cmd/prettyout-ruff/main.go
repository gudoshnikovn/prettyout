package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
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

func format(data []byte, cfg formatter.Config) error {
	var issues []issue
	if err := json.Unmarshal(data, &issues); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if err := formatter.Render(toIssues(issues, cfg), cfg); err != nil {
		return err
	}
	printFixHint(issues)
	return nil
}

func toIssues(raw []issue, cfg formatter.Config) []formatter.Issue {
	out := make([]formatter.Issue, 0, len(raw))
	for _, r := range raw {
		out = append(out, formatter.Issue{
			Rule:    r.Code,
			File:    formatter.ResolvePath(r.Filename, cfg),
			Line:    r.Location.Row,
			Message: formatter.Truncate(r.Message, cfg.MaxMessageLength),
		})
	}
	return out
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
