// Package cmd implements the `air` CLI (Cobra commands, Viper config).
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRootCmd builds the root command and wires Viper (flags + AIR_* env vars).
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "air",
		Short:         "AIR Harness installer & lifecycle manager",
		Long:          "air reads harness.manifest.yaml (the bill of materials) and a team\nprofile, then resolves, installs, and inspects the harness components.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().String("manifest", "", "path to harness.manifest.yaml (default: search upward from cwd)")
	root.PersistentFlags().String("persona", "", "restrict to a single persona id")
	// Output is Token-Efficient by default; opt out with --human (or --format human).
	root.PersistentFlags().String("format", "teo", "output format: teo (default) or human")
	root.PersistentFlags().Bool("human", false, "shorthand for --format human")
	root.PersistentFlags().Bool("teo", false, "shorthand for --format teo (the default)")

	v := viper.GetViper()
	v.SetEnvPrefix("AIR")
	v.AutomaticEnv() // AIR_MANIFEST, AIR_PROFILE, AIR_PERSONA, AIR_FORMAT, AIR_HUMAN
	_ = v.BindPFlag("manifest", root.PersistentFlags().Lookup("manifest"))
	_ = v.BindPFlag("persona", root.PersistentFlags().Lookup("persona"))
	_ = v.BindPFlag("format", root.PersistentFlags().Lookup("format"))
	_ = v.BindPFlag("human", root.PersistentFlags().Lookup("human"))
	_ = v.BindPFlag("teo", root.PersistentFlags().Lookup("teo"))

	root.AddCommand(
		newStatusCmd(),
		newPersonaCmd(),
		newPersonasCmd(),
		newSkillsCmd(),
		newAgentsCmd(),
		newTargetsCmd(),
		newPackageCmd(),
		newVersionCmd(),
	)
	return root
}

// Execute runs the root command; callers handle the returned error / exit code.
func Execute() error {
	return NewRootCmd().Execute()
}
