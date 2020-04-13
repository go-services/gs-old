package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "gs",
	SilenceUsage: true,
	Short:        "A tool to help you create microservices",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Used to set debugging mode on")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
