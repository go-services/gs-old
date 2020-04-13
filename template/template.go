package template

import (
	"bytes"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/CloudyKit/jet"

	"golang.org/x/tools/imports"
)

var log = logrus.WithFields(logrus.Fields{
	"package": "template",
})

func CompileFromPath(tplPath string, data interface{}) (string, error) {
	t, err := getSet().GetTemplate(tplPath)
	if err != nil {
		return "", err
	}
	var templateBuffer bytes.Buffer
	log.Debugf("Executing template %s", tplPath)
	err = t.Execute(&templateBuffer, make(jet.VarMap), data)
	if err != nil {
		return "", err
	}
	return templateBuffer.String(), err
}

func CompileGoFromPath(tplPath string, data interface{}) (string, error) {
	src, err := CompileFromPath(tplPath, data)
	if err != nil {
		return "", err
	}

	prettyCode, err := imports.Process(strings.Replace(tplPath, "jet", "go", -1), []byte(src), nil)
	return string(prettyCode), err
}
