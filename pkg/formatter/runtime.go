package formatter

import (
	"os"
	"strings"
)

// ApplyEnvOverrides merges PO_* environment variables on top of cfg.
// Env vars take precedence over config-file settings.
func ApplyEnvOverrides(cfg *Config) {
	if v := os.Getenv("PO_GROUP_BY"); v != "" {
		cfg.GroupBy = v
	}
	if v := os.Getenv("PO_SORT"); v != "" {
		cfg.Sort = v
	}
	if v := os.Getenv("PO_ONLY_RULES"); v != "" {
		cfg.OnlyRules = splitTrimmed(v)
	}
	if v := os.Getenv("PO_ONLY_FILES"); v != "" {
		cfg.OnlyFiles = splitTrimmed(v)
	}
	if os.Getenv("PO_STATS") == "1" {
		cfg.Stats = true
	}
}

func splitTrimmed(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
