package main

import (
	"gs/cmd"
	"gs/config"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

func main() {
	setupLogger()
	envDefaults()
	cmd.Execute()
}

func setupLogger() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(new(logrus.TextFormatter))
	logrus.SetLevel(logrus.InfoLevel)
}

func envDefaults() {
	viper.SetDefault(config.GSConfigFileName, "gs.json")
	viper.SetDefault(config.ServiceAnnotation, "service")

	viper.AutomaticEnv()
}
