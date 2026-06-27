package cmd

import (
	"fmt"
	"strings"

	"github.com/cloud-byte/air/internal/agents"
	"github.com/spf13/cobra"
)

func newAgentsCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "agents",
		Short: "Fan one canonical AGENTS.md out to every harness (replaces link_agents.*)",
		Long: "agents points each coding tool's instruction file (CLAUDE.md, GEMINI.md,\n" +
			".github/copilot-instructions.md, .cursor/rules/…, …) at one canonical\n" +
			"AGENTS.md. Mode is OS-aware: symlink on macOS/Linux, copy on Windows.",
	}
	c.AddCommand(newAgentsListCmd(), newAgentsLinkCmd(), newAgentsUnlinkCmd())
	return c
}

type agentsFlags struct {
	canonical string
	harness   []string
	all       bool
	mode      string
	dryRun    bool
	stubs     bool
	root      string
	templates string
}

// resolve expands any named sets in --harness into concrete harness names.
func (f *agentsFlags) resolve() error {
	if len(f.harness) == 0 {
		return nil
	}
	hs, err := resolveHarnesses(strings.Join(f.harness, ","))
	if err != nil {
		return err
	}
	f.harness = hs
	return nil
}

func (f *agentsFlags) options() agents.Options {
	return agents.Options{
		Root:         f.root,
		Canonical:    f.canonical,
		Harnesses:    f.harness,
		All:          f.all,
		Mode:         f.mode,
		DryRun:       f.dryRun,
		Stubs:        f.stubs,
		TemplatesDir: f.templates,
	}
}

func bindAgentsFlags(c *cobra.Command, f *agentsFlags) {
	c.Flags().StringVar(&f.canonical, "canonical", "AGENTS.md", "source-of-truth file")
	c.Flags().StringSliceVar(&f.harness, "harness", nil, "harness names and/or set names, e.g. claude,copilot or 'frontend' (default: a common set)")
	c.Flags().BoolVar(&f.all, "all", false, "every known harness")
	c.Flags().StringVar(&f.mode, "mode", "auto", "auto|symlink|copy (auto = copy on Windows, else symlink)")
	c.Flags().BoolVar(&f.dryRun, "dry-run", false, "print the plan; change nothing")
	c.Flags().StringVar(&f.root, "root", "", "project root (default: current directory)")
	c.Flags().StringVar(&f.templates, "templates", "", "dir to seed AGENTS.md / sidecars from")
}

func newAgentsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List known harnesses and their instruction paths",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			w := cmd.OutOrStdout()
			for _, h := range agents.Known {
				fmt.Fprintf(w, "  %-16s %s\n", h.Name, h.Path)
			}
			fmt.Fprintf(w, "\nresolved mode on this OS: %s\n", agents.ResolveMode("auto"))
		},
	}
}

func newAgentsLinkCmd() *cobra.Command {
	f := &agentsFlags{}
	c := &cobra.Command{
		Use:   "link",
		Short: "Link/copy each harness file to the canonical AGENTS.md",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			f.stubs = stubsFlag(cmd)
			if err := f.resolve(); err != nil {
				return err
			}
			rep, err := agents.Link(f.options())
			if err != nil {
				return err
			}
			printReport(cmd, rep)
			return nil
		},
	}
	bindAgentsFlags(c, f)
	c.Flags().Bool("stubs", false, "also create referenced sidecars (opinions.md, voice.md)")
	return c
}

func newAgentsUnlinkCmd() *cobra.Command {
	f := &agentsFlags{}
	c := &cobra.Command{
		Use:   "unlink",
		Short: "Remove only the harness links this tool manages",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := f.resolve(); err != nil {
				return err
			}
			rep, err := agents.Unlink(f.options())
			if err != nil {
				return err
			}
			printReport(cmd, rep)
			return nil
		},
	}
	bindAgentsFlags(c, f)
	return c
}

func stubsFlag(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("stubs")
	return v
}

func printReport(cmd *cobra.Command, rep *agents.Report) {
	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "canonical %s · mode %s\n", rep.Canonical, rep.Mode)
	for _, a := range rep.Actions {
		line := fmt.Sprintf("  %-7s %s", a.Verb, a.Path)
		if a.Note != "" {
			line += "  (" + a.Note + ")"
		}
		fmt.Fprintln(w, line)
	}
}
