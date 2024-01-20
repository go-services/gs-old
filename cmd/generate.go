package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gs/config"
	"gs/generate"
	"gs/parser"
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
		return generateServices()
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

func generateServices() error {
	cnf := config.Get()
	if cnf.Module == "" {
		logrus.Error("Not in the root of the module")
		return nil
	}

	files, err := parser.ParseFiles(".")
	if err != nil {
		return err
	}
	services, err := parser.FindServices(files)
	if err != nil {
		return err
	}

	err = generate.Common()
	if err != nil {
		return err
	}

	svcGen := generate.NewServiceGenerator(services)
	err = svcGen.Generate()
	if err != nil {
		return err
	}
	return nil
}
