package service

import (
	"errors"
	"gs/config"
	"gs/fs"
	"io/ioutil"
	"os"
	"path"
	"unicode"
	"unicode/utf8"

	"github.com/go-services/annotation"

	"github.com/go-services/source"
	"github.com/spf13/viper"

	"github.com/go-services/code"
)

type Endpoint struct {
	Name string

	Config config.ServiceConfig

	Params  []code.Parameter
	Results []code.Parameter

	Request  *code.Struct
	Response *code.Struct

	RequestImport  *code.Import
	ResponseImport *code.Import

	HttpTransport *HttpTransport

	Middlewares []Middleware

	Annotations []annotation.Annotation
}

func parseEndpoint(method source.InterfaceMethod, serviceImport, serviceName string) (ep *Endpoint, err error) {
	if err = checkEndpointParams(method.Params()); err != nil {
		return nil, err
	}
	if err = checkEndpointResults(method.Results()); err != nil {
		return nil, err
	}
	ep = &Endpoint{
		Name:        method.Name(),
		Annotations: method.Annotations(),
	}

	// this fixes the import for parameters in the same package
	for _, param := range method.Params() {
		param.Type = fixMethodImport(param.Type, serviceImport, serviceName)
		ep.Params = append(ep.Params, param)
	}
	// find the request struct and the import of the request
	ep.Request, ep.RequestImport, err = findRequest(ep.Params, serviceImport, serviceName)
	if err != nil {
		return nil, err
	}

	// this fixes the import for parameters in the same package
	for _, param := range method.Results() {
		param.Type = fixMethodImport(param.Type, serviceImport, serviceName)
		ep.Results = append(ep.Results, param)
	}

	// find the response struct and the import of the response
	ep.Response, ep.ResponseImport, err = findResponse(ep.Results, serviceImport, serviceName)
	if err != nil {
		return nil, err
	}

	ep.HttpTransport, err = ParseHttpTransport(*ep)
	if err != nil {
		return nil, err
	}

	ep.Middlewares = parseMiddleware(source.FindAnnotations("middleware", &method))
	// TODO add middleware parsing
	return
}

func findRequest(params []code.Parameter, serviceImport, serviceName string) (*code.Struct, *code.Import, error) {
	if len(params) < 2 {
		return nil, nil, nil
	}
	request, err := findStruct(params[1].Type)
	if err != nil {
		return nil, nil, err
	}

	// this fixes the import for parameters in the same package
	for inx, field := range request.Fields {
		field.Type = fixMethodImport(field.Type, serviceImport, serviceName)
		request.Fields[inx] = field
	}

	return request, params[1].Type.Import, nil
}

func findResponse(params []code.Parameter, serviceImport, serviceName string) (*code.Struct, *code.Import, error) {
	if len(params) < 2 {
		return nil, nil, nil
	}
	response, err := findStruct(params[0].Type)
	if err != nil {
		return nil, nil, err
	}
	// this fixes the import for parameters in the same package
	for inx, field := range response.Fields {
		field.Type = fixMethodImport(field.Type, serviceImport, serviceName)
		response.Fields[inx] = field
	}
	return response, params[0].Type.Import, nil
}

func isExported(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(ch)
}

func checkEndpointParams(params []code.Parameter) error {
	if len(params) != 1 && len(params) != 2 {
		return errors.New("method must except either the context or the context and the request struct")
	}
	if !(params[0].Type.Qualifier == "Context" &&
		params[0].Type.Import.Path == "context") &&
		params[0].Type.Pointer &&
		params[0].Type.Variadic {
		return errors.New("the first parameter of the method needs to be the context")
	}
	if len(params) == 2 && !isExported(params[1].Type.Qualifier) {
		return errors.New("request needs to be an exported structure")
	}
	return nil
}

func checkEndpointResults(params []code.Parameter) error {
	if (len(params) != 1 && len(params) != 2) ||
		len(params) == 1 && params[0].Type.Qualifier != "error" ||
		len(params) == 2 && params[1].Type.Qualifier != "error" ||
		len(params) == 2 && !params[0].Type.Pointer ||
		len(params) == 2 && !isExported(params[0].Type.Qualifier) {
		return errors.New("method must return either the error or the response pointer and the error")
	}
	return nil
}

func fixMethodImport(tp code.Type, serviceImport, serviceName string) code.Type {
	if tp.Import == nil && isExported(tp.Qualifier) {
		currentPath, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		if viper.GetString(fs.DebugKey) != "" {
			currentPath = path.Join(currentPath, viper.GetString(fs.DebugKey))
		}
		tp.Import = code.NewImportWithFilePath(
			"service",
			serviceImport,
			path.Join(currentPath, serviceName),
		)
	}
	return tp
}

func findStruct(tp code.Type) (*code.Struct, error) {
	notFoundErr := errors.New(
		"could not find structure, make sure that you are using a structure as request/response parameters",
	)
	if tp.Import.FilePath == "" {
		return nil, notFoundErr
	}
	fls, err := ioutil.ReadDir(tp.Import.FilePath)
	if err != nil {
		return nil, err
	}
	if fls == nil {
		return nil, notFoundErr
	}
	for _, file := range fls {
		if file.IsDir() {
			continue
		}
		var fileSource *source.Source
		filePath := path.Join(tp.Import.FilePath, file.Name())
		if src, ok := fileSourceCache[filePath]; ok {
			fileSource = src
		} else {
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			fileSource, err = source.New(string(data))
			fileSourceCache[filePath] = fileSource
			if err != nil {
				return nil, err
			}
		}
		for _, structure := range fileSource.Structures() {
			if structure.Name() == tp.Qualifier {
				return structure.Code().(*code.Struct), nil
			}
		}
	}
	return nil, notFoundErr
}
