package cmd

import (
	"github.com/sirupsen/logrus"
	"os"

	"github.com/spf13/cobra"
)

var log = logrus.WithFields(logrus.Fields{
	"package": "cmd",
})

var rootCmd = &cobra.Command{
	Use:          "gs",
	SilenceUsage: true,
	Short:        "A tool to help you create services",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Used to set debugging mode on")
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
