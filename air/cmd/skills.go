package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cloud-byte/air/internal/skills"
	"github.com/cloud-byte/air/internal/targets"
	"github.com/spf13/cobra"
	"github.com/cloud-byte-consulting/teo"
)

func newSkillsCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "skills",
		Short: "Manage one canonical skills folder shared across AI coding tools",
		Long: "Skills live once under pod-bundle/skills/<name>/SKILL.md. Make them usable\n" +
			"in Claude, Cursor, Copilot, and Gemini by symlink (link) or copy (sync).",
	}
	c.AddCommand(newSkillsListCmd(), newSkillsSyncCmd(), newSkillsLinkCmd())
	return c
}

func canonicalDir(flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "pod-bundle", "skills"), nil
}

func outDir(flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "pod-bundle"), nil
}

func toolList(flag string) []string {
	if flag == "" {
		return skills.Tools
	}
	return strings.Split(flag, ",")
}

func newSkillsListCmd() *cobra.Command {
	var dir string
	c := &cobra.Command{
		Use:   "list",
		Short: "List canonical skills and their descriptions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			d, err := canonicalDir(dir)
			if err != nil {
				return err
			}
			found, err := skills.Discover(d)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			if teoEnabled() {
				doc := teo.New()
				doc.Count(len(found))
				b := doc.Block("skills", "name", "description")
				for _, s := range found {
					b.Row(s.Name, s.Description)
				}
				fmt.Fprint(w, doc.String())
				return nil
			}
			for _, s := range found {
				fmt.Fprintf(w, "  %-32s %s\n", s.Name, truncate(s.Description, 70))
			}
			fmt.Fprintf(w, "\n%d skills in %s\n", len(found), d)
			return nil
		},
	}
	c.Flags().StringVar(&dir, "skills-dir", "", "canonical skills dir (default: pod-bundle/skills)")
	return c
}

func newSkillsSyncCmd() *cobra.Command {
	var dir, out, tools, harness string
	c := &cobra.Command{
		Use:   "sync",
		Short: "Copy skills into each tool's format (portable; works on Windows)",
		Long: "sync writes per-tool copies from the canonical skills folder:\n" +
			"  claude  -> .claude/skills/<name>/   (full dir, incl. references)\n" +
			"  cursor  -> .cursor/rules/<name>.mdc\n" +
			"  copilot -> .github/instructions/<name>.instructions.md\n" +
			"  gemini  -> .gemini/skills/<name>.md\n" +
			"Copies preserve each tool's frontmatter. Re-run after editing a skill.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProjection(cmd, dir, out, tools, harness, false)
		},
	}
	c.Flags().StringVar(&dir, "skills-dir", "", "canonical skills dir (default: pod-bundle/skills)")
	c.Flags().StringVar(&out, "out", "", "output root (default: pod-bundle)")
	c.Flags().StringVar(&tools, "tools", "", "comma list: claude,cursor,copilot,gemini (default: all)")
	c.Flags().StringVar(&harness, "harness", "", "target harnesses/sets, e.g. claude,copilot or 'frontend' (maps to skills tools)")
	return c
}

func newSkillsLinkCmd() *cobra.Command {
	var dir, out, tools, harness string
	c := &cobra.Command{
		Use:   "link",
		Short: "Symlink each tool location to the canonical skills (single source)",
		Long: "link points the tool locations at pod-bundle/skills via symlinks, so\n" +
			"editing a skill once is reflected everywhere — no copies to maintain:\n" +
			"  claude  -> .claude/skills        (one directory symlink; full fidelity)\n" +
			"  cursor/copilot/gemini            (per-skill file symlinks)\n" +
			"Caveat: a symlinked file shares the body but NOT each tool's trigger\n" +
			"frontmatter (Cursor alwaysApply, Copilot applyTo). Symlinks need\n" +
			"Developer Mode / core.symlinks on Windows — use `sync` if unavailable.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProjection(cmd, dir, out, tools, harness, true)
		},
	}
	c.Flags().StringVar(&dir, "skills-dir", "", "canonical skills dir (default: pod-bundle/skills)")
	c.Flags().StringVar(&out, "out", "", "output root (default: pod-bundle)")
	c.Flags().StringVar(&tools, "tools", "", "comma list: claude,cursor,copilot,gemini (default: all)")
	c.Flags().StringVar(&harness, "harness", "", "target harnesses/sets, e.g. claude,copilot or 'frontend' (maps to skills tools)")
	return c
}

func runProjection(cmd *cobra.Command, dirFlag, outFlag, toolsFlag, harnessFlag string, link bool) error {
	dir, err := canonicalDir(dirFlag)
	if err != nil {
		return err
	}
	out, err := outDir(outFlag)
	if err != nil {
		return err
	}
	tools := toolList(toolsFlag)
	if harnessFlag != "" {
		hs, herr := resolveHarnesses(harnessFlag)
		if herr != nil {
			return herr
		}
		tools = targets.SkillsTools(hs)
		if len(tools) == 0 {
			return fmt.Errorf("no skills-capable harnesses in %q (skills tools: claude, cursor, copilot, gemini)", harnessFlag)
		}
	}
	verb := "Copied"
	fn := skills.Sync
	if link {
		verb, fn = "Linked", skills.Link
	}
	counts, err := fn(dir, out, tools)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "%s skills from %s -> %s\n", verb, dir, out)
	for _, t := range tools {
		fmt.Fprintf(w, "  %-8s %d\n", t, counts[strings.TrimSpace(t)])
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
