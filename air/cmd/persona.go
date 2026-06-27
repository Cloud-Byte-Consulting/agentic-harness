package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newPersonaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "persona <id>",
		Short: "Show a persona pack's selection (persona.yaml)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := personaFile(args[0])
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("persona %q not found (%s)", args[0], path)
			}
			fmt.Fprint(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
}
