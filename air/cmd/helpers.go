package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloud-byte/air/internal/manifest"
	"github.com/spf13/viper"
)

const manifestName = "harness.manifest.yaml"

// manifestPath resolves the manifest location: the --manifest flag / AIR_MANIFEST
// env if set, otherwise searched upward from the current directory.
func manifestPath() (string, error) {
	if p := viper.GetString("manifest"); p != "" {
		return p, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		candidate := filepath.Join(dir, manifestName)
		if fileExists(candidate) {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("%s not found (searched from %s upward); pass --manifest", manifestName, cwd)
}

// repoRoot is the directory containing the manifest.
func repoRoot() (string, error) {
	p, err := manifestPath()
	if err != nil {
		return "", err
	}
	return filepath.Dir(p), nil
}

func loadManifest() (*manifest.Manifest, error) {
	p, err := manifestPath()
	if err != nil {
		return nil, err
	}
	return manifest.Load(p)
}

// personaFile returns the persona.yaml path for an id.
func personaFile(id string) (string, error) {
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "pod-bundle", "personas", id, "persona.yaml"), nil
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// teoEnabled reports whether output should be Token-Efficient. TEO is the
// default; --human / --format human (or AIR_HUMAN / AIR_FORMAT=human) opts out.
// --teo / --format teo force it.
func teoEnabled() bool {
	if viper.GetBool("teo") {
		return true
	}
	if viper.GetBool("human") {
		return false
	}
	switch strings.ToLower(viper.GetString("format")) {
	case "human", "text":
		return false
	}
	return true
}
