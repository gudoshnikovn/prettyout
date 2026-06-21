package main

import (
	"encoding/json"
	"fmt"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

type mypyMsg struct {
	File     string  `json:"file"`
	Line     int     `json:"line"`
	Severity string  `json:"severity"`
	Message  string  `json:"message"`
	Code     *string `json:"code"`
}

func main() {
	formatter.RunWithConfig("mypy", format)
}

func format(data []byte, cfg formatter.Config) error {
	lines, err := formatter.ParseNDJSON(data)
	if err != nil {
		return fmt.Errorf("invalid NDJSON: %w", err)
	}

	var msgs []mypyMsg
	for _, raw := range lines {
		var m mypyMsg
		if err := json.Unmarshal(raw, &m); err != nil {
			continue
		}
		// skip notes
		if m.Severity == "note" {
			continue
		}
		msgs = append(msgs, m)
	}

	return formatter.Render(toIssues(msgs, cfg), cfg)
}

func toIssues(raw []mypyMsg, cfg formatter.Config) []formatter.Issue {
	out := make([]formatter.Issue, 0, len(raw))
	for _, r := range raw {
		out = append(out, formatter.Issue{
			Rule:     codeStr(r),
			File:     formatter.ResolvePath(r.File, cfg),
			Line:     r.Line,
			Message:  formatter.Truncate(r.Message, cfg.MaxMessageLength),
			Severity: r.Severity,
		})
	}
	return out
}

func codeStr(m mypyMsg) string {
	if m.Code != nil && *m.Code != "" {
		return *m.Code
	}
	return "error"
}
