package registry

import "sort"

type ToolConfig struct {
	JSONFlags  []string
	PluginName string
}

var Tools = map[string]ToolConfig{
	"ruff": {
		JSONFlags:  []string{"--output-format=json"},
		PluginName: "prettyout-ruff",
	},
	"basedpyright": {
		JSONFlags:  []string{"--outputjson"},
		PluginName: "prettyout-basedpyright",
	},
}

func SortedTools() []string {
	names := make([]string, 0, len(Tools))
	for name := range Tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
