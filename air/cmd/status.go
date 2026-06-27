package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"truenas-scale-1.tail5a208d.ts.net/Cloud-Byte-Consulting/teo"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "List manifest components and their artifact kinds",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			m, err := loadManifest()
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()

			if teoEnabled() {
				d := teo.New()
				d.Scalar("description", "AIR manifest components")
				d.Scalar("harness", m.Harness.Name)
				d.Scalar("release", m.Harness.Release)
				d.Count(len(m.Components))
				b := d.Block("components", "id", "kind", "layer", "lifecycle")
				for _, c := range m.Components {
					b.Row(c.ID, c.Kind, c.Layer, c.Lifecycle)
				}
				fmt.Fprint(w, d.String())
				return nil
			}

			fmt.Fprintf(w, "%s — release %s\n\n", m.Harness.Name, m.Harness.Release)
			for _, c := range m.Components {
				fmt.Fprintf(w, "  %-34s kind=%-8s layer=%-9s lifecycle=%s\n",
					c.ID, c.Kind, c.Layer, c.Lifecycle)
			}
			return nil
		},
	}
}
