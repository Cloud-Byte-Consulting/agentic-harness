package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cloud-byte/air/internal/agents"
	"github.com/cloud-byte/air/internal/targets"
	"github.com/spf13/cobra"
	"github.com/cloud-byte-consulting/teo"
)

// userSets merges custom target sets from (lowest→highest precedence):
//   - ~/.air/targets.yaml          (global, available in any project)
//   - <harness repo>/harness.targets.yaml
//   - ./harness.targets.yaml       (current project)
func userSets() map[string][]string {
	merged := map[string][]string{}
	load := func(path string) {
		if s, err := targets.LoadSets(path); err == nil {
			for k, v := range s {
				merged[k] = v
			}
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		load(filepath.Join(home, ".air", "targets.yaml"))
	}
	if root, err := repoRoot(); err == nil {
		load(filepath.Join(root, "harness.targets.yaml"))
	}
	if cwd, err := os.Getwd(); err == nil {
		load(filepath.Join(cwd, "harness.targets.yaml"))
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

// resolveHarnesses expands a --harness spec (names and/or set names) into a
// concrete, deduped harness list.
func resolveHarnesses(spec string) ([]string, error) {
	return targets.Resolve(spec, userSets())
}

func newTargetsCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "targets",
		Short: "List harnesses and named target sets you can pass to --harness",
	}
	c.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "Show harnesses and named sets",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				w := cmd.OutOrStdout()
				sets := targets.Sets(userSets())
				names := make([]string, 0, len(sets))
				for n := range sets {
					names = append(names, n)
				}
				sort.Strings(names)

				if teoEnabled() {
					d := teo.New()
					hb := d.Block("harnesses", "name", "path", "skills")
					for _, h := range agents.Known {
						hb.Row(h.Name, h.Path, targets.SkillsTool[h.Name] != "")
					}
					sb := d.Block("sets", "name", "members")
					for _, n := range names {
						sb.Row(n, strings.Join(sets[n], "|")) // documented delimiter: | (avoids row comma)
					}
					fmt.Fprint(w, d.String())
					return nil
				}

				fmt.Fprintln(w, "Harnesses:")
				for _, h := range agents.Known {
					tag := ""
					if t := targets.SkillsTool[h.Name]; t != "" {
						tag = "  (skills ✓)"
					}
					fmt.Fprintf(w, "  %-16s %s%s\n", h.Name, h.Path, tag)
				}
				fmt.Fprintln(w, "\nSets:")
				for _, n := range names {
					fmt.Fprintf(w, "  %-12s %s\n", n, strings.Join(sets[n], ", "))
				}
				fmt.Fprintln(w, "\nUse with --harness, e.g.  air agents link --harness claude,copilot")
				fmt.Fprintln(w, "or a set:               air bootstrap --harness frontend")
				fmt.Fprintln(w, "Define your own in harness.targets.yaml (sets: { name: [harness, …] }).")
				return nil
			},
		},
		&cobra.Command{
			Use:   "resolve <spec>",
			Short: "Print the harnesses a spec expands to",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				hs, err := resolveHarnesses(args[0])
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), strings.Join(hs, " "))
				return nil
			},
		},
	)
	return c
}
