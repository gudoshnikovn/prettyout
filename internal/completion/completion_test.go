package completion

import (
	"strings"
	"testing"
)

func TestGenerate_zsh(t *testing.T) {
	got, err := Generate("zsh", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "compdef _prettyout prettyout") {
		t.Error("zsh script missing compdef declaration")
	}
	if !strings.Contains(got, "prettyout _completions tools") {
		t.Error("zsh script should call prettyout _completions tools for dynamic tool names")
	}
}

func TestGenerate_bash(t *testing.T) {
	got, err := Generate("bash", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "complete -F _prettyout prettyout") {
		t.Error("bash script missing complete declaration")
	}
	if !strings.Contains(got, "prettyout _completions tools") {
		t.Error("bash script should call prettyout _completions tools for dynamic tool names")
	}
}

func TestGenerate_unsupportedShell(t *testing.T) {
	_, err := Generate("fish", nil)
	if err == nil {
		t.Error("unsupported shell should return error")
	}
}

func TestGenerate_zsh_hasSubcommands(t *testing.T) {
	got, _ := Generate("zsh", nil)
	for _, cmd := range []string{"setup", "enable", "disable", "install", "doctor", "status", "completion"} {
		if !strings.Contains(got, cmd) {
			t.Errorf("zsh completion missing subcommand %q", cmd)
		}
	}
}
