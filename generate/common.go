package generate

import (
	"gs/assets"
	"gs/config"
	"gs/fs"
	"path"
	"strings"
)

func genPath() string {
	cnf := config.Get()
	gp := "gen"
	if cnf.GenPath != "" {
		gp = cnf.GenPath
	}
	return gp
}

func cmdPath() string {
	cnf := config.Get()
	cp := "cmd"
	if cnf.CmdPath != "" {
		cp = cnf.CmdPath
	}
	return cp
}

func Common() error {
	gp := genPath()
	if exists, _ := fs.Exists(gp); !exists {
		_ = fs.CreateFolder(gp)
	}

	commonFiles := []string{
		"errors/errors.go.tmpl",
		"errors/http.go.tmpl",
		"utils/utils.go.tmpl",
	}

	for _, file := range commonFiles {
		err := assets.ParseAndWriteTemplate(file, path.Join(gp, strings.TrimSuffix(file, ".tmpl")), nil)
		if err != nil {
			return err
		}
	}
	return nil
}
