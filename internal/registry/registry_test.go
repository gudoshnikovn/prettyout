package registry

import "testing"

func TestBuiltinLoads(t *testing.T) {
	r, err := LoadBuiltin()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.Tools["ruff"]; !ok {
		t.Error("ruff not in built-in registry")
	}
	if _, ok := r.Launchers["uvx"]; !ok {
		t.Error("uvx not in built-in launchers")
	}
}

func TestShouldIntercept_withSubcommands(t *testing.T) {
	tc := ToolConfig{InterceptSubcommands: []string{"check"}}
	if !tc.ShouldIntercept("check") {
		t.Error("should intercept 'check'")
	}
	if tc.ShouldIntercept("format") {
		t.Error("should not intercept 'format'")
	}
}

func TestShouldIntercept_noRestriction(t *testing.T) {
	tc := ToolConfig{}
	if !tc.ShouldIntercept("anything") {
		t.Error("empty intercept_subcommands means intercept all")
	}
}

func TestHasPassthroughFlag(t *testing.T) {
	tc := ToolConfig{PassthroughFlags: []string{"--watch", "-W"}}
	if !tc.HasPassthroughFlag([]string{"check", "--watch", "."}) {
		t.Error("should detect --watch")
	}
	if tc.HasPassthroughFlag([]string{"check", "."}) {
		t.Error("no passthrough flag present")
	}
}

func TestMerge_addsCustomTool(t *testing.T) {
	r, _ := LoadBuiltin()
	r.Merge(map[string]ToolConfig{
		"mycooltool": {Plugin: "prettyout-mycooltool"},
	})
	if _, ok := r.Tools["mycooltool"]; !ok {
		t.Error("custom tool not merged")
	}
}

func TestMerge_overridesBuiltin(t *testing.T) {
	r, _ := LoadBuiltin()
	r.Merge(map[string]ToolConfig{
		"ruff": {Plugin: "my-ruff-override"},
	})
	if r.Tools["ruff"].Plugin != "my-ruff-override" {
		t.Error("custom tool should override builtin")
	}
}

func TestSortedToolNames(t *testing.T) {
	r, _ := LoadBuiltin()
	names := r.SortedToolNames()
	if len(names) == 0 {
		t.Error("expected at least one tool name")
	}
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("names not sorted: %q before %q", names[i-1], names[i])
		}
	}
}
