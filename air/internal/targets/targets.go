// Package targets resolves a single "which harnesses" vocabulary shared by every
// command. A target spec is a comma-separated list of harness names and/or named
// sets (e.g. "claude,copilot" or "frontend"), so you can configure several
// harnesses at once and save reusable configurations.
package targets

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloud-byte/air/internal/agents"
	"github.com/spf13/viper"
)

// SkillsTool maps a harness to the skills-projection tool it supports (empty =
// no skills projection for that harness).
var SkillsTool = map[string]string{
	"claude":  "claude",
	"cursor":  "cursor",
	"copilot": "copilot",
	"gemini":  "gemini",
}

// Builtin named sets.
func builtin() map[string][]string {
	return map[string][]string{
		"all":     HarnessNames(),
		"popular": {"claude", "gemini", "copilot", "cursor"},
		"skills":  {"claude", "cursor", "copilot", "gemini"}, // harnesses that support skill projection
	}
}

// HarnessNames is the canonical harness list (from the agents registry).
func HarnessNames() []string {
	out := make([]string, 0, len(agents.Known))
	for _, h := range agents.Known {
		out = append(out, h.Name)
	}
	return out
}

func isHarness(name string) bool {
	for _, h := range agents.Known {
		if h.Name == name {
			return true
		}
	}
	return false
}

// Resolve expands a spec of harness names and/or set names into a deduped,
// ordered harness list. userSets (may be nil) extends/overrides the builtin sets.
// An empty spec resolves to an empty list (callers apply their own default).
func Resolve(spec string, userSets map[string][]string) ([]string, error) {
	sets := builtin()
	for k, v := range userSets {
		sets[k] = v
	}
	seen := map[string]bool{}
	var out []string
	add := func(n string) {
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	for _, tok := range strings.Split(spec, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		if grp, ok := sets[tok]; ok {
			for _, n := range grp {
				if !isHarness(n) {
					return nil, fmt.Errorf("set %q references unknown harness %q", tok, n)
				}
				add(n)
			}
			continue
		}
		if isHarness(tok) {
			add(tok)
			continue
		}
		return nil, fmt.Errorf("unknown harness or set %q (try `air targets list`)", tok)
	}
	return out, nil
}

// SkillsTools maps resolved harnesses to their skills tools, dropping any that
// have no skills projection.
func SkillsTools(harnesses []string) []string {
	var out []string
	for _, h := range harnesses {
		if t := SkillsTool[h]; t != "" {
			out = append(out, t)
		}
	}
	return out
}

// Sets returns the merged builtin + user sets (for display).
func Sets(userSets map[string][]string) map[string][]string {
	m := builtin()
	for k, v := range userSets {
		m[k] = v
	}
	return m
}

type fileSets struct {
	Sets map[string][]string `mapstructure:"sets"`
}

// LoadSets reads user-defined sets from a YAML file (harness.targets.yaml).
// A missing file is not an error (returns nil).
func LoadSets(path string) (map[string][]string, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, nil
	}
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read targets %q: %w", path, err)
	}
	var f fileSets
	if err := v.Unmarshal(&f); err != nil {
		return nil, fmt.Errorf("parse targets %q: %w", path, err)
	}
	return f.Sets, nil
}
