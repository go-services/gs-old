package main

import (
	"gs/cmd"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	setupLogger()
	cmd.Execute()
}

func setupLogger() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(new(logrus.TextFormatter))
	logrus.SetLevel(logrus.InfoLevel)
}
