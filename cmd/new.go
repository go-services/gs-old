package cmd

import (
	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gs/assets"
	"gs/fs"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

var newCmd = &cobra.Command{
	Use: "new",
	Aliases: []string{
		"n",
	},
	Short: "New",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var projectCmd = &cobra.Command{
	Use: "project",
	Aliases: []string{
		"p",
	},
	Args:  cobra.ExactArgs(1),
	Short: "New project",
	RunE: func(cmd *cobra.Command, args []string) error {
		if b, _ := cmd.Flags().GetBool("debug"); b {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return generateProject(args[0])
	},
}

func init() {
	newCmd.AddCommand(projectCmd)
	rootCmd.AddCommand(newCmd)
}

func generateProject(name string) error {
	formattedName := strcase.ToKebab(name)
	if exists, _ := fs.Exists(formattedName); exists {
		logrus.Errorf("Folder %s already exists", formattedName)
		return nil
	}
	err := assets.ParseAndWriteTemplate(
		"project/package.json.tmpl",
		path.Join(formattedName, "package.json"),
		map[string]string{
			"Name": formattedName,
		},
	)
	if err != nil {
		return err
	}

	err = assets.ParseAndWriteTemplate(
		"project/sst.config.ts.tmpl",
		path.Join(formattedName, "sst.config.ts"),
		map[string]string{
			"Name": formattedName,
		},
	)
	if err != nil {
		return err
	}

	_ = fs.CreateFolder(path.Join(formattedName, "stacks", "gen"))
	err = assets.ParseAndWriteTemplate(
		"project/stacks/gen/gen.ts.tmpl",
		path.Join(formattedName, "stacks", "gen", "index.ts"),
		nil,
	)
	if err != nil {
		return err
	}

	_ = fs.CreateFolder(path.Join(formattedName, "numbers"))
	err = assets.ParseAndWriteTemplate(
		"project/numbers/service.go.tmpl",
		path.Join(formattedName, "numbers", "service.go"),
		map[string]string{
			"Module": strcase.ToSnake(formattedName),
		},
	)
	if err != nil {
		return err
	}

	err = assets.ParseAndWriteTemplate(
		"project/go.mod.tmpl",
		path.Join(formattedName, "go.mod"),
		map[string]string{
			"Module":  strcase.ToSnake(formattedName),
			"Version": strings.TrimPrefix(runtime.Version(), "go"),
		},
	)
	if err != nil {
		return err
	}

	err = os.Chdir(formattedName)
	if err != nil {
		return err
	}

	log.Infof("Generating project %s", formattedName)
	cmd := exec.Command("gs", "generate")

	err = cmd.Run()

	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Installing dependencies")
	npmCmd := exec.Command("npm", "install")
	npmCmd.Stdout = os.Stdout
	npmCmd.Stderr = os.Stderr
	npmCmd.Stdin = os.Stdin
	err = npmCmd.Run()

	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Project %s created", formattedName)
	log.Infof("Run `cd %s && npm dev` to start the app", formattedName)

	return nil
}
