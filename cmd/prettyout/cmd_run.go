package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gudoshnikovn/prettyout/internal/runner"
)

// runRun handles `prettyout run [--po-flags...] tool [tool-args...]`.
// Prettyout flags are consumed and set as PO_* env vars.
// The first non-flag argument is the tool name; the rest are passed to the tool.
func runRun(args []string) {
	var toolAndArgs []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--raw":
			os.Setenv("PO_RAW", "1")
		case arg == "--group-by" && i+1 < len(args):
			i++
			os.Setenv("PO_GROUP_BY", args[i])
		case strings.HasPrefix(arg, "--group-by="):
			os.Setenv("PO_GROUP_BY", strings.TrimPrefix(arg, "--group-by="))
		case arg == "--sort" && i+1 < len(args):
			i++
			os.Setenv("PO_SORT", args[i])
		case strings.HasPrefix(arg, "--sort="):
			os.Setenv("PO_SORT", strings.TrimPrefix(arg, "--sort="))
		case arg == "--only-rules" && i+1 < len(args):
			i++
			os.Setenv("PO_ONLY_RULES", args[i])
		case strings.HasPrefix(arg, "--only-rules="):
			os.Setenv("PO_ONLY_RULES", strings.TrimPrefix(arg, "--only-rules="))
		case arg == "--only-files" && i+1 < len(args):
			i++
			os.Setenv("PO_ONLY_FILES", args[i])
		case strings.HasPrefix(arg, "--only-files="):
			os.Setenv("PO_ONLY_FILES", strings.TrimPrefix(arg, "--only-files="))
		default:
			// First non-po arg = tool name; rest = tool args
			toolAndArgs = args[i:]
			i = len(args) // stop loop
		}
	}
	if len(toolAndArgs) == 0 {
		fmt.Fprintln(os.Stderr, "prettyout run: tool name required")
		fmt.Fprintln(os.Stderr, "Usage: prettyout run [--raw] [--group-by=file|rule] [--sort=count|alpha] [--only-rules=A,B] [--only-files=src/] <tool> [args...]")
		os.Exit(1)
	}
	os.Exit(runner.RunFromArgs(toolAndArgs))
}
