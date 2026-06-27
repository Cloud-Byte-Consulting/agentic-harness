package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"truenas-scale-1.tail5a208d.ts.net/Cloud-Byte-Consulting/teo"
)

// Version is set at build time via -ldflags "-X .../cmd.Version=...".
var Version = "dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the air version and harness release",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			w := cmd.OutOrStdout()
			m, mErr := loadManifest()

			if teoEnabled() {
				d := teo.New().Scalar("air", Version)
				if mErr == nil {
					d.Scalar("harness", m.Harness.Name).Scalar("release", m.Harness.Release)
				}
				fmt.Fprint(w, d.String())
				return
			}

			fmt.Fprintf(w, "air %s\n", Version)
			if mErr == nil {
				fmt.Fprintf(w, "harness %s release %s\n", m.Harness.Name, m.Harness.Release)
			}
		},
	}
}
