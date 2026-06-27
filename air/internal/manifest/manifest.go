// Package manifest loads and queries the AIR harness bill of materials
// (harness.manifest.yaml).
package manifest

import (
	"fmt"

	"github.com/spf13/viper"
)

// Component is one entry in the bill of materials.
type Component struct {
	ID        string   `mapstructure:"id"`
	Repo      string   `mapstructure:"repo"`
	Path      string   `mapstructure:"path"`
	Version   string   `mapstructure:"version"`
	Kind      string   `mapstructure:"kind"` // oci | binary | npm | pipx | archive
	Artifact  string   `mapstructure:"artifact"`
	Lifecycle string   `mapstructure:"lifecycle"` // tool | service | content
	Layer     string   `mapstructure:"layer"`     // platform | core
	Personas  []string `mapstructure:"personas"`
}

// Harness is the top-level identity block.
type Harness struct {
	Name    string `mapstructure:"name"`
	Release string `mapstructure:"release"`
}

// Manifest is the parsed harness.manifest.yaml.
type Manifest struct {
	Harness    Harness           `mapstructure:"harness"`
	Defaults   map[string]string `mapstructure:"defaults"`
	Components []Component       `mapstructure:"components"`
}

// Load reads and validates a manifest file.
func Load(path string) (*Manifest, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read manifest %q: %w", path, err)
	}
	var m Manifest
	if err := v.Unmarshal(&m); err != nil {
		return nil, fmt.Errorf("parse manifest %q: %w", path, err)
	}
	if len(m.Components) == 0 {
		return nil, fmt.Errorf("manifest %q has no components", path)
	}
	return &m, nil
}

// ServesPersona reports whether the component is included for the given persona
// id. A component tagged "all" serves everyone; an empty persona disables the
// filter (everything matches).
func (c Component) ServesPersona(persona string) bool {
	if persona == "" {
		return true
	}
	for _, p := range c.Personas {
		if p == "all" || p == persona {
			return true
		}
	}
	return false
}

// Action returns the install verb implied by a component's lifecycle.
func (c Component) Action() string {
	switch c.Lifecycle {
	case "service":
		return "fetch+run"
	case "content":
		return "extract"
	default:
		return "fetch"
	}
}
