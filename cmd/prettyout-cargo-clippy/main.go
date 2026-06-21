package main

import (
	"encoding/json"
	"fmt"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

type clippySpan struct {
	FileName  string `json:"file_name"`
	LineStart int    `json:"line_start"`
	IsPrimary bool   `json:"is_primary"`
}

type clippyCode struct {
	Code string `json:"code"`
}

type clippyMessage struct {
	Level   string       `json:"level"`
	Message string       `json:"message"`
	Code    *clippyCode  `json:"code"`
	Spans   []clippySpan `json:"spans"`
}

type clippyLine struct {
	Reason  string        `json:"reason"`
	Message clippyMessage `json:"message"`
}

func main() {
	formatter.RunWithConfig("cargo-clippy", format)
}

func format(data []byte, cfg formatter.Config) error {
	lines, err := formatter.ParseNDJSON(data)
	if err != nil {
		return fmt.Errorf("invalid NDJSON: %w", err)
	}

	var items []clippyLine
	for _, raw := range lines {
		var cl clippyLine
		if err := json.Unmarshal(raw, &cl); err != nil {
			continue
		}
		if cl.Reason != "compiler-message" {
			continue
		}
		// skip note and help levels
		if cl.Message.Level == "note" || cl.Message.Level == "help" {
			continue
		}
		items = append(items, cl)
	}

	return formatter.Render(toIssues(items, cfg), cfg)
}

func toIssues(raw []clippyLine, cfg formatter.Config) []formatter.Issue {
	out := make([]formatter.Issue, 0, len(raw))
	for _, r := range raw {
		rawFile, line := primarySpan(r)
		file := rawFile
		if file == "" {
			file = "unknown"
		} else {
			file = formatter.ResolvePath(file, cfg)
		}
		out = append(out, formatter.Issue{
			Rule:     ruleCode(r),
			File:     file,
			Line:     line,
			Message:  formatter.Truncate(r.Message.Message, cfg.MaxMessageLength),
			Severity: r.Message.Level,
		})
	}
	return out
}

func ruleCode(cl clippyLine) string {
	if cl.Message.Code != nil && cl.Message.Code.Code != "" {
		return cl.Message.Code.Code
	}
	return "rustc"
}

func primarySpan(cl clippyLine) (string, int) {
	for _, s := range cl.Message.Spans {
		if s.IsPrimary {
			return s.FileName, s.LineStart
		}
	}
	if len(cl.Message.Spans) > 0 {
		return cl.Message.Spans[0].FileName, cl.Message.Spans[0].LineStart
	}
	return "", 0
}
