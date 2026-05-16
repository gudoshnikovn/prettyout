package registry

import (
	_ "embed"
	"sort"

	"gopkg.in/yaml.v3"
)

//go:embed builtin.yaml
var builtinYAML []byte

type LauncherConfig struct {
	SkipFlags    []string `yaml:"skip_flags"`
	ValueFlags   []string `yaml:"value_flags"`
	PrefixArgs   []string `yaml:"prefix_args"`
	ToolPosition string   `yaml:"tool_position"`
}

type ToolConfig struct {
	Plugin               string   `yaml:"plugin"`
	InterceptSubcommands []string `yaml:"intercept_subcommands"`
	OutputArgs           []string `yaml:"output_args"`
	PassthroughFlags     []string `yaml:"passthrough_flags"`
	Launchers            []string `yaml:"launchers"`
}

type Registry struct {
	Launchers map[string]LauncherConfig `yaml:"launchers"`
	Tools     map[string]ToolConfig     `yaml:"tools"`
}

func LoadBuiltin() (*Registry, error) {
	var r Registry
	if err := yaml.Unmarshal(builtinYAML, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Merge adds custom tools from user config. Custom entries override builtins.
func (r *Registry) Merge(custom map[string]ToolConfig) {
	if r.Tools == nil {
		r.Tools = map[string]ToolConfig{}
	}
	for name, cfg := range custom {
		r.Tools[name] = cfg
	}
}

// SortedToolNames returns tool names in alphabetical order.
func (r *Registry) SortedToolNames() []string {
	names := make([]string, 0, len(r.Tools))
	for name := range r.Tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ShouldIntercept returns true if subcommand should be intercepted.
// An empty InterceptSubcommands list means intercept all subcommands.
func (tc ToolConfig) ShouldIntercept(subcommand string) bool {
	if len(tc.InterceptSubcommands) == 0 {
		return true
	}
	for _, s := range tc.InterceptSubcommands {
		if s == subcommand {
			return true
		}
	}
	return false
}

// HasPassthroughFlag returns true if any arg matches a passthrough flag.
func (tc ToolConfig) HasPassthroughFlag(args []string) bool {
	for _, arg := range args {
		for _, pf := range tc.PassthroughFlags {
			if arg == pf {
				return true
			}
		}
	}
	return false
}
