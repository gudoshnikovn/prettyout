package install

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gudoshnikovn/prettyout/internal/registry"
)

func TestIsInstalled_found(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "prettyout-test-binary")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	tc := registry.ToolConfig{Plugin: "prettyout-test-binary"}
	if !IsInstalled(tc) {
		t.Error("should detect binary in PATH")
	}
}

func TestIsInstalled_notFound(t *testing.T) {
	tc := registry.ToolConfig{Plugin: "prettyout-definitely-not-installed-xyz"}
	if IsInstalled(tc) {
		t.Error("should not find non-existent binary")
	}
}
