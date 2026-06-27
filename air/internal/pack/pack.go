// Package pack builds the AIR release artifacts (replaces package-harness.sh):
// it validates each persona pack, tars+gzips each one and a top-level harness
// bundle, and writes a manifest.json with SHA256 hashes. Stdlib only, so it runs
// the same on every OS.
package pack

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloud-byte/air/internal/manifest"
)

// Entry is one produced artifact.
type Entry struct {
	Name     string `json:"id,omitempty"`
	Artifact string `json:"artifact"`
	SHA256   string `json:"sha256"`
}

// Report is the packaging outcome (also serialized to dist/manifest.json).
type Report struct {
	Harness  string  `json:"harness"`
	Release  string  `json:"release"`
	Bundle   Entry   `json:"bundle"`
	Personas []Entry `json:"personas"`
}

var requiredPodFiles = []string{"persona.yaml", "README.md", "pod.md", "behavior.md", "sources.md", "workflows.md"}

// Package writes artifacts under outDir and returns the report.
func Package(repoRoot, outDir string) (*Report, error) {
	release := "dev"
	if m, err := manifest.Load(filepath.Join(repoRoot, "harness.manifest.yaml")); err == nil {
		release = m.Harness.Release
	}
	if err := os.RemoveAll(outDir); err != nil {
		return nil, err
	}
	rep := &Report{Harness: "AIR", Release: release}

	personasRoot := filepath.Join(repoRoot, "pod-bundle", "personas")
	entries, err := os.ReadDir(personasRoot)
	if err != nil {
		return nil, fmt.Errorf("read personas: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(personasRoot, e.Name())
		for _, f := range requiredPodFiles {
			if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
				return nil, fmt.Errorf("persona %q missing %s", e.Name(), f)
			}
		}
		artifact := filepath.Join(outDir, "personas", e.Name()+".tar.gz")
		if err := tarGz(repoRoot, []string{rel(repoRoot, dir)}, artifact, nil); err != nil {
			return nil, err
		}
		sum, err := sha256File(artifact)
		if err != nil {
			return nil, err
		}
		rep.Personas = append(rep.Personas, Entry{e.Name(), relOut(outDir, artifact), sum})
	}

	// Harness bundle: manifest + init + profiles + pod-bundle (no dot dirs / dist).
	bundle := filepath.Join(outDir, "air-harness-"+release+".tar.gz")
	roots := []string{"harness.manifest.yaml", "harness.init.yaml", "profiles", "pod-bundle", "bin"}
	if err := tarGz(repoRoot, existing(repoRoot, roots), bundle, skipNoise); err != nil {
		return nil, err
	}
	sum, err := sha256File(bundle)
	if err != nil {
		return nil, err
	}
	rep.Bundle = Entry{Artifact: relOut(outDir, bundle), SHA256: sum}

	data, _ := json.MarshalIndent(rep, "", "  ")
	if err := os.WriteFile(filepath.Join(outDir, "manifest.json"), append(data, '\n'), 0o644); err != nil {
		return nil, err
	}
	return rep, nil
}

// skipNoise excludes hidden dirs (generated tool views) and build output.
func skipNoise(relPath string) bool {
	for _, part := range strings.Split(filepath.ToSlash(relPath), "/") {
		if strings.HasPrefix(part, ".") || part == "dist" {
			return true
		}
	}
	return false
}

func existing(root string, paths []string) []string {
	var out []string
	for _, p := range paths {
		if _, err := os.Stat(filepath.Join(root, p)); err == nil {
			out = append(out, p)
		}
	}
	return out
}

// tarGz writes a gzipped tar of the given roots (relative to base) into dst.
func tarGz(base string, roots []string, dst string, skip func(string) bool) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	for _, root := range roots {
		abs := filepath.Join(base, root)
		err := filepath.Walk(abs, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			name, _ := filepath.Rel(base, p)
			name = filepath.ToSlash(name)
			if skip != nil && skip(name) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if !info.Mode().IsRegular() {
				return nil // skip dirs (implicit) and symlinks
			}
			hdr, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			hdr.Name = name
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			in, err := os.Open(p)
			if err != nil {
				return err
			}
			defer in.Close()
			_, err = io.Copy(tw, in)
			return err
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func sha256File(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func rel(base, p string) string { r, _ := filepath.Rel(base, p); return filepath.ToSlash(r) }

func relOut(outDir, p string) string { r, _ := filepath.Rel(outDir, p); return filepath.ToSlash(r) }
