package cmd

import (
	"errors"
	"gs/config"
	"gs/service"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use: "generate",
	Aliases: []string{
		"gen",
		"g",
	},
	Short: "Generate service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Read()
		if err != nil {
			return err
		}
		var svc *config.ServiceConfig
		for _, v := range cfg.Services {
			if v.Name == args[0] {
				svc = &v
			}
		}
		if svc == nil {
			return errors.New("service with this mane does not exits in the configuration file")
		}
		return service.Generate(*svc, cfg.Module)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
