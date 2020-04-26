package cmd

import (
	"gs/config"
	"gs/service"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use: "generate",
	Aliases: []string{
		"gen",
		"g",
	},
	Short: "Generate services",
	RunE: func(cmd *cobra.Command, args []string) error {
		if b, _ := cmd.Flags().GetBool("debug"); b {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return generateServices(args...)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

func generateServices(services ...string) error {
	cfg, err := config.Read()
	if err != nil {
		return err
	}
	for _, svc := range services {
		if svcCfg, ok := cfg.Services[svc]; ok {
			err := service.Generate(svc, svcCfg, cfg.Module)
			if err != nil {
				return err
			}
		} else {
			logrus.Warnf("service `%s` does not exits in the configuration file", svc)
		}
	}
	if len(services) == 0 {
		for name, svcCfg := range cfg.Services {
			err := service.Generate(name, svcCfg, cfg.Module)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
