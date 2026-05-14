package formatter

import (
	"fmt"
	"io"
	"os"
)

// Run reads all of stdin and passes it to transform. Use this in every plugin's main().
func Run(transform func([]byte) error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "prettyout: read error: %v\n", err)
		os.Exit(1)
	}
	if err := transform(data); err != nil {
		fmt.Fprintf(os.Stderr, "prettyout: %v\n", err)
		os.Exit(1)
	}
}
