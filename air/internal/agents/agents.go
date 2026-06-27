// Package agents fans one canonical agent-instructions file (AGENTS.md) out to
// every harness-specific path. It replaces the link_agents.{sh,ps1,py,nu} scripts
// with a single cross-platform implementation that picks symlink vs copy based on
// the running OS (Windows defaults to copy, where symlinks are unreliable).
package agents

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Harness maps a known coding agent to the instruction file it reads.
type Harness struct {
	Name string
	Path string
}

// Known is the full set of harnesses (order = display order).
var Known = []Harness{
	{"codex", "AGENTS.md"}, // also opencode, amp, and the AGENTS.md standard
	{"claude", "CLAUDE.md"},
	{"gemini", "GEMINI.md"},
	{"copilot", ".github/copilot-instructions.md"},
	{"cursor", ".cursor/rules/agents.mdc"},
	{"windsurf", ".windsurf/rules/agents.md"},
	{"cline", ".clinerules/agents.md"},
	{"kiro", ".kiro/steering/agents.md"},
	{"cursor-legacy", ".cursorrules"},
	{"windsurf-legacy", ".windsurfrules"},
}

// Defaults is the common set used when neither --harness nor --all is given.
var Defaults = []string{"claude", "gemini", "copilot", "cursor", "windsurf"}

var ruleDir = map[string]bool{"cursor": true, "windsurf": true, "cline": true, "kiro": true}

var sidecars = map[string]string{"opinions.md": "opinions.template.md", "voice.md": "voice.template.md"}

// Options configures a Link/Unlink run.
type Options struct {
	Root         string   // project root the files are written under (default: cwd)
	Canonical    string   // canonical file, relative to Root (default: AGENTS.md)
	Harnesses    []string // explicit subset; empty => Defaults (or all if All)
	All          bool
	Mode         string // "auto" | "symlink" | "copy"; auto => copy on Windows, else symlink
	DryRun       bool
	Stubs        bool
	TemplatesDir string // optional: seed canonical/sidecars from here
}

// Action records one planned/performed operation for reporting.
type Action struct{ Harness, Path, Verb, Note string }

// Report is the outcome of a run.
type Report struct {
	Canonical string
	Mode      string
	Actions   []Action
}

// ResolveMode turns "auto"/"" into the OS-appropriate concrete mode.
func ResolveMode(mode string) string {
	if mode == "" || mode == "auto" {
		if runtime.GOOS == "windows" {
			return "copy"
		}
		return "symlink"
	}
	return mode
}

func pathFor(name string) (string, bool) {
	for _, h := range Known {
		if h.Name == name {
			return h.Path, true
		}
	}
	return "", false
}

// ResolveTargets returns the (harness,path) pairs to operate on, skipping any
// whose path equals the canonical file.
func ResolveTargets(opts Options) ([]Harness, error) {
	var names []string
	switch {
	case opts.All:
		for _, h := range Known {
			names = append(names, h.Name)
		}
	case len(opts.Harnesses) > 0:
		names = opts.Harnesses
	default:
		names = Defaults
	}
	canon := filepath.Clean(opts.Canonical)
	var out []Harness
	for _, n := range names {
		p, ok := pathFor(n)
		if !ok {
			return nil, fmt.Errorf("unknown harness %q (run `air agents list`)", n)
		}
		if filepath.Clean(p) == canon {
			continue
		}
		out = append(out, Harness{n, p})
	}
	return out, nil
}

// Link fans the canonical file out to each target.
func Link(opts Options) (*Report, error) {
	opts = withDefaults(opts)
	targets, err := ResolveTargets(opts)
	if err != nil {
		return nil, err
	}
	rep := &Report{Canonical: opts.Canonical, Mode: ResolveMode(opts.Mode)}
	if err := ensureCanonical(opts, rep); err != nil {
		return nil, err
	}
	for _, t := range targets {
		if err := linkOne(opts, rep, t); err != nil {
			return nil, err
		}
	}
	if opts.Stubs {
		if err := createStubs(opts, rep); err != nil {
			return nil, err
		}
	}
	return rep, nil
}

// Unlink removes only the links this tool manages (those pointing at canonical).
func Unlink(opts Options) (*Report, error) {
	opts = withDefaults(opts)
	targets, err := ResolveTargets(opts)
	if err != nil {
		return nil, err
	}
	rep := &Report{Canonical: opts.Canonical, Mode: ResolveMode(opts.Mode)}
	canonAbs := abs(opts.Root, opts.Canonical)
	for _, t := range targets {
		full := filepath.Join(opts.Root, t.Path)
		fi, err := os.Lstat(full)
		if err != nil {
			continue
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			rep.Actions = append(rep.Actions, Action{t.Name, t.Path, "skip", "real file, not a managed link"})
			continue
		}
		dest, _ := os.Readlink(full)
		if abs(filepath.Dir(full), dest) != canonAbs {
			rep.Actions = append(rep.Actions, Action{t.Name, t.Path, "skip", "links elsewhere"})
			continue
		}
		if opts.DryRun {
			rep.Actions = append(rep.Actions, Action{t.Name, t.Path, "remove", "dry-run"})
			continue
		}
		if err := os.Remove(full); err != nil {
			return nil, err
		}
		note := ""
		if bak := full + ".bak"; exists(bak) {
			_ = os.Rename(bak, full)
			note = "restored .bak"
		}
		rep.Actions = append(rep.Actions, Action{t.Name, t.Path, "removed", note})
	}
	return rep, nil
}

func withDefaults(o Options) Options {
	if o.Root == "" {
		o.Root, _ = os.Getwd()
	}
	if o.Canonical == "" {
		o.Canonical = "AGENTS.md"
	}
	return o
}

func ensureCanonical(opts Options, rep *Report) error {
	full := filepath.Join(opts.Root, opts.Canonical)
	if exists(full) {
		return nil
	}
	if opts.DryRun {
		rep.Actions = append(rep.Actions, Action{"-", opts.Canonical, "create", "dry-run (from template)"})
		return nil
	}
	if tmpl := templatePath(opts, "AGENTS.template.md"); tmpl != "" {
		if err := copyFile(tmpl, full); err != nil {
			return err
		}
		rep.Actions = append(rep.Actions, Action{"-", opts.Canonical, "create", "seeded from template"})
		return nil
	}
	if err := os.WriteFile(full, []byte("# Agent Instructions\n\nShared instructions for all agents.\n"), 0o644); err != nil {
		return err
	}
	rep.Actions = append(rep.Actions, Action{"-", opts.Canonical, "create", "empty starter"})
	return nil
}

func linkOne(opts Options, rep *Report, t Harness) error {
	mode := ResolveMode(opts.Mode)
	full := filepath.Join(opts.Root, t.Path)
	linkDir := filepath.Dir(full)
	canonAbs := abs(opts.Root, opts.Canonical)
	relTarget, _ := filepath.Rel(linkDir, canonAbs)

	note := ""
	if ruleDir[t.Name] {
		note = "rules-dir harness: add frontmatter for strict always-apply"
	}

	// Idempotency: already linked correctly?
	if fi, err := os.Lstat(full); err == nil && fi.Mode()&os.ModeSymlink != 0 && mode == "symlink" {
		if dest, _ := os.Readlink(full); abs(linkDir, dest) == canonAbs {
			rep.Actions = append(rep.Actions, Action{t.Name, t.Path, "ok", "already linked"})
			return nil
		}
	}
	realExists := exists(full) && !isSymlink(full)

	if opts.DryRun {
		verb := mode
		if realExists {
			note = "backup existing -> .bak; " + note
		}
		rep.Actions = append(rep.Actions, Action{t.Name, t.Path, verb, strings.TrimSpace(note)})
		return nil
	}
	if err := os.MkdirAll(linkDir, 0o755); err != nil {
		return err
	}
	if realExists {
		bak := full + ".bak"
		if exists(bak) {
			rep.Actions = append(rep.Actions, Action{t.Name, t.Path, "skip", ".bak exists; left untouched"})
			return nil
		}
		if err := os.Rename(full, bak); err != nil {
			return err
		}
	} else if isSymlink(full) {
		_ = os.Remove(full)
	}

	if mode == "copy" {
		if err := copyFile(filepath.Join(opts.Root, opts.Canonical), full); err != nil {
			return err
		}
		rep.Actions = append(rep.Actions, Action{t.Name, t.Path, "copy", strings.TrimSpace(note)})
	} else {
		if err := os.Symlink(relTarget, full); err != nil {
			return fmt.Errorf("symlink %s: %w (try --mode copy)", t.Path, err)
		}
		rep.Actions = append(rep.Actions, Action{t.Name, t.Path, "link", strings.TrimSpace(note)})
	}
	return nil
}

func createStubs(opts Options, rep *Report) error {
	base := filepath.Dir(abs(opts.Root, opts.Canonical))
	for name, tmplName := range sidecars {
		dest := filepath.Join(base, name)
		if exists(dest) {
			rep.Actions = append(rep.Actions, Action{"sidecar", name, "ok", "exists"})
			continue
		}
		if opts.DryRun {
			rep.Actions = append(rep.Actions, Action{"sidecar", name, "create", "dry-run"})
			continue
		}
		if tmpl := templatePath(opts, tmplName); tmpl != "" {
			if err := copyFile(tmpl, dest); err != nil {
				return err
			}
		} else if err := os.WriteFile(dest, []byte("# "+name+"\n"), 0o644); err != nil {
			return err
		}
		rep.Actions = append(rep.Actions, Action{"sidecar", name, "create", "from template; fill in"})
	}
	return nil
}

// templatePath looks for a bundled template in the configured dir or common spots.
func templatePath(opts Options, name string) string {
	var dirs []string
	if opts.TemplatesDir != "" {
		dirs = append(dirs, opts.TemplatesDir)
	}
	dirs = append(dirs, filepath.Join(opts.Root, "templates"))
	for _, d := range dirs {
		p := filepath.Join(d, name)
		if exists(p) {
			return p
		}
	}
	return ""
}

func abs(base, p string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Clean(filepath.Join(base, p))
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

func isSymlink(p string) bool {
	fi, err := os.Lstat(p)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
