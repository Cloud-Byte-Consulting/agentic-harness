package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/cloud-byte/air/internal/pack"
	"github.com/spf13/cobra"
)

func newPackageCmd() *cobra.Command {
	var out string
	c := &cobra.Command{
		Use:   "package",
		Short: "Build release artifacts: persona packs + bundle + manifest.json (replaces package-harness.sh)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := repoRoot()
			if err != nil {
				return err
			}
			if out == "" {
				out = filepath.Join(root, "dist")
			}
			rep, err := pack.Package(root, out)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "AIR %s -> %s\n", rep.Release, out)
			for _, p := range rep.Personas {
				fmt.Fprintf(w, "  persona  %-28s %s…\n", p.Name, p.SHA256[:12])
			}
			fmt.Fprintf(w, "  bundle   %-28s %s…\n", rep.Bundle.Artifact, rep.Bundle.SHA256[:12])
			fmt.Fprintf(w, "%d persona packs + 1 bundle + manifest.json\n", len(rep.Personas))
			return nil
		},
	}
	c.Flags().StringVar(&out, "out", "", "output dir (default: dist/)")
	return c
}
