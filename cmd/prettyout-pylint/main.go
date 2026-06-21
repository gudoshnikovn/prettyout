package main

import (
	"encoding/json"
	"fmt"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

type pylintMsg struct {
	Type           string `json:"type"`
	Line           int    `json:"line"`
	Path           string `json:"path"`
	Symbol         string `json:"symbol"`
	Message        string `json:"message"`
	MessageID      string `json:"message-id"` // json format
	MessageIDJson2 string `json:"messageId"`  // json2 format
}

func (m *pylintMsg) msgID() string {
	if m.MessageIDJson2 != "" {
		return m.MessageIDJson2
	}
	return m.MessageID
}

type pylintJSON2 struct {
	Messages   []pylintMsg `json:"messages"`
	Statistics struct {
		Score float64 `json:"score"`
	} `json:"statistics"`
}

type ruleEntry struct {
	severity  string
	message   string
	display   string
	fileLines map[string]map[int]struct{}
}

func main() {
	formatter.RunWithConfig("pylint", format)
}

func format(data []byte, cfg formatter.Config) error {
	var wrapper pylintJSON2
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	msgs := wrapper.Messages
	score := wrapper.Statistics.Score

	if err := formatter.Render(toIssues(msgs, cfg), cfg); err != nil {
		return err
	}
	fmt.Printf("  ↳ rated %.2f/10\n", score)
	return nil
}

func toIssues(raw []pylintMsg, cfg formatter.Config) []formatter.Issue {
	out := make([]formatter.Issue, 0, len(raw))
	for _, r := range raw {
		out = append(out, formatter.Issue{
			Rule:     r.msgID(),
			Display:  ruleDisplay(r),
			File:     formatter.ResolvePath(r.Path, cfg),
			Line:     r.Line,
			Message:  formatter.Truncate(r.Message, cfg.MaxMessageLength),
			Severity: pylintSeverity(r.Type),
		})
	}
	return out
}

func pylintSeverity(t string) string {
	switch t {
	case "error", "fatal":
		return "error"
	case "warning":
		return "warning"
	default:
		return "info"
	}
}

func ruleDisplay(m pylintMsg) string {
	if m.Symbol != "" {
		return m.msgID() + "/" + m.Symbol
	}
	return m.msgID()
}
