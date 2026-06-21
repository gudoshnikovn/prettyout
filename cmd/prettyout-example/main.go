// prettyout-example is a reference plugin demonstrating the full plugin pattern.
//
// Copy this file, rename the package, and adapt the JSON structs and
// grouping logic for your tool. The helpers in pkg/formatter handle all
// the display details so you only need to write the parsing and grouping.
//
// Build and test:
//
//	go build -o /tmp/prettyout-example ./cmd/prettyout-example/
//	your-tool --json ... | /tmp/prettyout-example
//
// Register in your config to intercept the tool automatically:
//
//	custom_tools:
//	  mytool:
//	    plugin: prettyout-example
//	    output_args: [--json]
package main

import (
	"encoding/json"
	"fmt"

	"github.com/gudoshnikovn/prettyout/pkg/formatter"
)

// --- JSON structs: adapt these to your tool's output format ---

// toolIssue mirrors one entry from your tool's JSON output.
// Fields vary per tool; use `json:",omitempty"` for optional ones.
type toolIssue struct {
	RuleID   string `json:"rule_id"`  // the rule/check that fired
	Message  string `json:"message"`  // human-readable description
	Severity string `json:"severity"` // "error", "warning", "info", etc.
	File     string `json:"file"`     // path to the file (absolute or relative)
	Line     int    `json:"line"`     // 1-based line number
}

// --- Entry point ---

func main() {
	// RunWithConfig reads stdin, loads the tool's section from config
	// (~/.config/prettyout/config.yaml and .prettyout.yaml), and calls
	// your format function. On error it prints to stderr and exits 1.
	formatter.RunWithConfig("example", format)
}

// --- Core formatter ---

func format(data []byte, cfg formatter.Config) error {
	// Step 1: parse your tool's JSON.
	var issues []toolIssue
	if err := json.Unmarshal(data, &issues); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Step 2: convert to generic issues and render.
	return formatter.Render(toIssues(issues, cfg), cfg)
}

// toIssues converts your tool's specific JSON structs into the generic Issue format.
func toIssues(issues []toolIssue, cfg formatter.Config) []formatter.Issue {
	out := make([]formatter.Issue, 0, len(issues))
	for _, iss := range issues {
		out = append(out, formatter.Issue{
			Rule: iss.RuleID,
			File: formatter.ResolvePath(iss.File, cfg),
			Line: iss.Line,
			// Truncate caps message length at cfg.MaxMessageLength (0 = unlimited).
			Message:  formatter.Truncate(iss.Message, cfg.MaxMessageLength),
			Severity: iss.Severity,
		})
	}
	return out
}
