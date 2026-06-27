package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/cloud-byte/air/internal/personas"
	"github.com/spf13/cobra"
	"truenas-scale-1.tail5a208d.ts.net/Cloud-Byte-Consulting/teo"
)

func newPersonasCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "personas",
		Short: "Scaffold and list persona packs (replaces scaffold-personas.sh)",
	}
	c.AddCommand(newPersonasScaffoldCmd(), newPersonasListCmd())
	return c
}

func newPersonasListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the persona specs",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			w := cmd.OutOrStdout()
			if teoEnabled() {
				d := teo.New()
				d.Count(len(personas.Specs))
				b := d.Block("personas", "id", "role", "gate")
				for _, s := range personas.Specs {
					b.Row(s.ID, s.Role, s.Tier)
				}
				fmt.Fprint(w, d.String())
				return
			}
			for _, s := range personas.Specs {
				fmt.Fprintf(w, "  %-28s %-7s gate=%s\n", s.ID, s.Role, s.Tier)
			}
			fmt.Fprintf(w, "\n%d personas\n", len(personas.Specs))
		},
	}
}

func newPersonasScaffoldCmd() *cobra.Command {
	var tmpl, out string
	c := &cobra.Command{
		Use:   "scaffold",
		Short: "Generate persona packs (pod files + persona.yaml) from the template",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := repoRoot()
			if err != nil {
				return err
			}
			if tmpl == "" {
				tmpl = filepath.Join(root, "pod-bundle", "templates", "pod-bundle")
			}
			if out == "" {
				out = filepath.Join(root, "pod-bundle", "personas")
			}
			made, err := personas.Scaffold(tmpl, out)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			for _, id := range made {
				fmt.Fprintf(w, "  scaffolded %s\n", id)
			}
			fmt.Fprintf(w, "%d persona packs -> %s\n", len(made), out)
			return nil
		},
	}
	c.Flags().StringVar(&tmpl, "template", "", "pod template dir (default: pod-bundle/templates/pod-bundle)")
	c.Flags().StringVar(&out, "out", "", "output dir (default: pod-bundle/personas)")
	return c
}
