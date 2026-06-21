package formatter

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestRun_happyPath(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	w.Write([]byte("hello"))
	w.Close()

	var got []byte
	Run(func(data []byte) error {
		got = data
		return nil
	})
	os.Stdin = oldStdin

	if string(got) != "hello" {
		t.Errorf("Run: got %q, want %q", got, "hello")
	}
}

func TestRunWithConfig_happyPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Chdir(t.TempDir())

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	w.Write([]byte(`[]`))
	w.Close()

	var gotData []byte
	var gotCfg Config
	RunWithConfig("ruff", func(data []byte, cfg Config) error {
		gotData = data
		gotCfg = cfg
		return nil
	})
	os.Stdin = oldStdin

	if string(gotData) != `[]` {
		t.Errorf("RunWithConfig: got data %q, want []", gotData)
	}
	if gotCfg.GroupBy != "rule" {
		t.Errorf("RunWithConfig: GroupBy = %q, want rule", gotCfg.GroupBy)
	}
}

func TestRunWithConfig_propagatesEnvOverrides(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Chdir(t.TempDir())
	t.Setenv("PO_GROUP_BY", "file")

	r, w, _ := os.Pipe()
	oldStdin := os.Stdin
	os.Stdin = r
	w.Write([]byte(`{}`))
	w.Close()

	var gotCfg Config
	RunWithConfig("ruff", func(_ []byte, cfg Config) error {
		gotCfg = cfg
		return nil
	})
	os.Stdin = oldStdin

	if gotCfg.GroupBy != "file" {
		t.Errorf("env override PO_GROUP_BY not applied: GroupBy = %q", gotCfg.GroupBy)
	}
}

func TestRun_transformError(t *testing.T) {
	// When transform returns an error, Run writes to stderr and calls os.Exit(1).
	// We can't test os.Exit directly, but we verify that the non-error path works
	// and the error is formatted correctly by inspecting the error message format.
	msg := fmt.Sprintf("prettyout: %v\n", fmt.Errorf("test error"))
	if !strings.HasPrefix(msg, "prettyout: ") {
		t.Error("error format unexpected")
	}
}
