package main

import (
	"github.com/sirupsen/logrus"
	"gs/cmd"
	"os"
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
