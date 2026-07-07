package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var BuildVersion = "dev"

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nuon-ext-linter",
		Short: "Lint Nuon app config directories for best practices",
		Long: `Lint Nuon app config directories for best practices and common errors.

Checks for missing labels, overly broad permissions, sandbox configurations
using branches instead of tags, and other best practice violations.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		newLintCmd(),
		newInitCmd(),
		newRulesCmd(),
		&cobra.Command{
			Use:   "version",
			Short: "Print the extension version",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(BuildVersion)
			},
		},
	)

	return cmd
}
