package cmd

import (
	"gs/watch"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch is used to hot reload your microservices",
	RunE: func(cmd *cobra.Command, args []string) error {
		if b, _ := cmd.Flags().GetBool("debug"); b {
			logrus.SetLevel(logrus.DebugLevel)
		}
		p, _ := cmd.Flags().GetInt("port")
		if err := generateServices(); err != nil {
			return err
		}
		watch.Run(p)
		return nil
	},
}

func init() {
	watchCmd.Flags().IntP("port", "p", 8888, "the port to run the proxy")
	rootCmd.AddCommand(watchCmd)
}
