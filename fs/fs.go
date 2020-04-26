package fs

import (
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/spf13/afero"
)

var log = logrus.WithFields(logrus.Fields{
	"package": "fs",
})

var fs afero.Fs
var testFs afero.Fs

func appFs() afero.Fs {
	if testFs != nil {
		log.Debug("Using test filesystem")
		return testFs
	}
	if fs == nil {
		fs = afero.NewOsFs()
	}
	return fs
}

func DeleteFolder(path string) error {
	log.Debugf("Deleting `%s` with files", path)
	return appFs().RemoveAll(path)
}

func CreateFolder(path string) error {
	log.Debugf("Creating `%s`", path)
	b, _ := afero.Exists(appFs(), path)
	if b {
		return fmt.Errorf("folder `%s` already exists", path)
	}
	return appFs().MkdirAll(path, 0755)
}

func WriteFile(path, data string) error {
	log.Debugf("Writing `%s`", path)
	dir := filepath.Dir(path)
	b, _ := afero.Exists(appFs(), dir)
	if !b {
		err := CreateFolder(dir)
		if err != nil {
			return err
		}
	}
	return afero.WriteFile(appFs(), path, []byte(data), 0644)
}

func ReadFile(path string) (string, error) {
	b, err := afero.ReadFile(appFs(), path)
	return string(b), err
}

func Exists(path string) (bool, error) {
	return afero.Exists(appFs(), path)
}
