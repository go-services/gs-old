package cmd

import (
	"gs/fs"
	"gs/template"
	"path"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/ozgio/strutil"

	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Create a new project",
	Aliases: []string{
		"p",
	},
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if b, _ := cmd.Flags().GetBool("debug"); b {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return newProject(args[0])
	},
}

func init() {
	newCmd.AddCommand(projectCmd)
}

func newProject(name string) error {
	// we should remove the '_' because of this guide https://blog.golang.org/package-names
	moduleName := strings.ReplaceAll(strutil.ToSnakeCase(name), "_", "")

	if err := fs.CreateFolder(moduleName); err != nil {
		return err
	}

	goMod, err := template.CompileFromPath("project/go.mod.jet", map[string]string{
		"Module": moduleName,
	})
	if err != nil {
		return err
	}

	gitignore, err := template.CompileFromPath("project/gitignore", nil)
	if err != nil {
		return err
	}
	if err := fs.WriteFile(path.Join(moduleName, ".gitignore"), gitignore); err != nil {
		return err
	}
	if err := fs.WriteFile(path.Join(moduleName, "go.mod"), goMod); err != nil {
		return err
	}
	return fs.WriteFile(path.Join(moduleName, "gs.toml"), "")
}
