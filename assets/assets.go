package assets

import (
	"embed"
	"fmt"
	"github.com/iancoleman/strcase"
	"golang.org/x/tools/imports"
	"gs/fs"
	"strings"
	"text/template"
)

//go:embed templates/*
var folder embed.FS

var CustomFunctions = template.FuncMap{
	"lowerFirst":          strcase.ToLowerCamel,
	"title":               strcase.ToCamel,
	"camelCase":           strcase.ToLowerCamel,
	"httpResponseEncoder": GetHttpResponseEncodeFunction,
	"httpRequestDecoder":  GetHttpRequestDecoderFunction,
}

func GetHttpResponseEncodeFunction(tp string) string {
	return map[string]string{
		"JSON": "JsonEncoder",
		"XML":  "XmlEncoder",
	}[tp]
}

func GetHttpRequestDecoderFunction(tp string) string {
	return map[string]string{
		"JSON": "JsonDecoder",
		"XML":  "XmlDecoder",
		"FORM": "FormDecoder",
	}[tp]
}

func ReadTemplate(path string) (string, error) {
	content, err := folder.ReadFile("templates/" + path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func ParseTemplate(path string, data interface{}) (string, error) {
	tmplContent, err := ReadTemplate(path)
	if err != nil {
		return "", err
	}

	tmpl := template.New(path).Funcs(CustomFunctions)
	tmpl, err = tmpl.Parse(tmplContent)
	if err != nil {
		return "", err
	}

	buf := &strings.Builder{}
	err = tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func ParseAndWriteTemplate(path, outPath string, data interface{}) error {
	content, err := ParseTemplate(path, data)
	if err != nil {
		return err
	}
	isGoOutput := strings.HasSuffix(outPath, ".go")
	outContent := content
	if isGoOutput {
		formatted, err := imports.Process(path, []byte(content), &imports.Options{
			Comments: true,
		})
		outContent = string(formatted)
		if err != nil {
			fmt.Println(content)
			return err
		}
	}
	return fs.WriteFile(outPath, outContent)
}
