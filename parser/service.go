package parser

import (
	"errors"
	"fmt"
	"github.com/go-services/annotation"
	"github.com/go-services/code"
	"github.com/go-services/source"
	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
	"gs/config"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

var log = logrus.WithFields(logrus.Fields{
	"package": "api",
})

// this is used from findStruct() if the file was already read we don't
// want to spend all the time to read and parse it again
var fileSourceCache = map[string]*source.Source{}

type Endpoint struct {
	Name string

	Config config.GSConfig

	Params  []code.Parameter
	Results []code.Parameter

	Request  *code.Struct
	Response *code.Struct

	RequestImport  *code.Import
	ResponseImport *code.Import

	HttpTransport *HttpTransport

	Annotations []annotation.Annotation
}

type Service struct {
	Name      string
	Interface string
	Config    config.GSConfig
	Import    string
	Package   string
	BaseRoute string

	Endpoints   []Endpoint
	Annotations []annotation.Annotation
}

func (a Endpoint) PackageName() string {
	return strcase.ToSnake(a.Name)
}

func (a Endpoint) CamelCaseName() string {
	return strcase.ToCamel(a.Name)

}

func parseService(file AnnotatedFile) (*Service, error) {
	cnf := config.Get()
	inf, svcAnnotation := findServiceInterface(&file.Src)
	if inf == nil {
		log.Debugf("File `%s` has no services interface", file.Path)
		return nil, nil
	}
	name := strcase.ToSnake(svcAnnotation.Get("name").String())
	if name == "" {
		name = strcase.ToSnake(inf.Name())
	}

	var newMth *source.Function
	for _, fn := range file.Src.Functions() {
		if fn.Name() == "New" && len(fn.Results()) == 1 && fn.Results()[0].Type.Qualifier == inf.Name() {
			newMth = &fn
			break
		}
	}
	if newMth == nil {
		return nil, errors.New(fmt.Sprintf("service `%s` has no New method, each service needs to have a New method that returns a Service", name))
	}

	impt := getPackageImport(cnf.Module, file.Path, file.Src.Package())

	route := svcAnnotation.Get("route").String()
	if route == "" {
		log.Warnf("Service `%s` has no base route, the service name will be used as route", name)
		route = name
	}
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	if strings.HasSuffix(route, "/") {
		route = route[:len(route)-1]
	}
	service := &Service{
		Interface:   inf.Name(),
		BaseRoute:   route,
		Config:      *cnf,
		Name:        name,
		Import:      impt,
		Package:     file.Src.Package(),
		Annotations: inf.Annotations(),
	}

	for _, method := range filterMethods(inf.Methods()) {
		ep, err := parseEndpoint(method, route, service.Import, file.Path)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error while parsing endpoint `%s.%s` : %s", inf.Name(), method.Name(), err.Error()))
		}

		for _, ex := range service.Endpoints {
			if ex.HttpTransport.MethodRoutes[0].Route == ep.HttpTransport.MethodRoutes[0].Route && ex.HttpTransport.MethodRoutes[0].Methods[0] == ep.HttpTransport.MethodRoutes[0].Methods[0] {
				return nil, errors.New(fmt.Sprintf("endpoint `%s.%s` has the same route and method as `%s`", service.Interface, ep.Name, ex.Name))
			}
		}

		service.Endpoints = append(service.Endpoints, *ep)
	}
	return service, nil

}

func filterMethods(methods []source.InterfaceMethod) (list []source.InterfaceMethod) {
	for _, method := range methods {
		httpAnnotations := findAnnotations("http", method.Annotations())
		if isExported(method.Name()) {
			if len(httpAnnotations) > 0 {
				list = append(list, method)
			}
		} else {
			if len(httpAnnotations) > 0 {
				log.Warnf("Method `%s` is not exported and will be ignored", method.Name())
			}
		}
	}
	return list
}

func FindServices(files []AnnotatedFile) (services []Service, err error) {
	for _, file := range files {
		svc, err := parseService(file)
		if err != nil {
			return nil, err
		}
		if svc == nil {
			continue
		}
		for _, sv := range services {
			if sv.Name == svc.Name {
				return nil, errors.New(fmt.Sprintf("service `%s` is defined more than once", svc.Name))
			}
			if sv.BaseRoute == svc.BaseRoute {
				return nil, errors.New(fmt.Sprintf("service `%s` has the same base route as `%s`", svc.Name, sv.Name))
			}
		}
		services = append(services, *svc)
	}
	return
}

func getPackageImport(module string, pth string, pkg string) string {
	dir := filepath.Dir(filepath.Dir(pth))
	if dir == "." {
		return fmt.Sprintf("%s/%s", module, pkg)
	}
	return fmt.Sprintf("%s/%s/%s", module, dir, pkg)
}

func findServiceInterface(src *source.Source) (*source.Interface, *annotation.Annotation) {
	for _, inf := range src.Interfaces() {
		annotations := findAnnotations("service", inf.Annotations())
		if len(annotations) > 0 {
			if len(annotations) > 1 {
				log.Warnf("Interface `%s` has more than one services annotation last one will be used", inf.Name())
			}
			return &inf, &annotations[len(annotations)-1]
		}
	}
	return nil, nil
}

func parseEndpoint(method source.InterfaceMethod, serviceRoute, serviceImport, filePath string) (ep *Endpoint, err error) {
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
		param.Type = fixMethodImport(param.Type, serviceImport, filePath)
		ep.Params = append(ep.Params, param)
	}
	// find the request struct and the import of the request
	ep.Request, ep.RequestImport, err = findRequest(ep.Params, serviceImport, filePath)
	if err != nil {
		return nil, err
	}

	// this fixes the import for parameters in the same package
	for _, param := range method.Results() {
		param.Type = fixMethodImport(param.Type, serviceImport, filePath)
		ep.Results = append(ep.Results, param)
	}

	// find the response struct and the import of the response
	ep.Response, ep.ResponseImport, err = findResponse(ep.Results, serviceImport, filePath)
	if err != nil {
		return nil, err
	}

	ep.HttpTransport, err = parseHttpTransport(serviceRoute, *ep)
	if err != nil {
		return nil, err
	}
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

func fixMethodImport(tp code.Type, imp, filePath string) code.Type {
	if filepath.Ext(filePath) != "" {
		filePath = filepath.Dir(filePath)
	}
	if tp.Import == nil && isExported(tp.Qualifier) {
		tp.Import = code.NewImportWithFilePath(
			"",
			imp,
			filePath,
		)
	}
	return tp
}

func fixStructFieldImport(tp code.Type, alias, structImport, structFile string) code.Type {
	if tp.Import == nil && isExported(tp.Qualifier) {
		tp.Import = code.NewImportWithFilePath(
			alias,
			structImport,
			structFile,
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
		if file.IsDir() || filepath.Ext(file.Name()) != ".go" {
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
				strc := structure.Code().(*code.Struct)
				for inx, field := range strc.Fields {
					field.Type = fixStructFieldImport(field.Type, tp.Import.Alias, tp.Import.Path, tp.Import.FilePath)
					if field.Name == "" {
						field.Name = field.Type.Qualifier
					}
					strc.Fields[inx] = field
				}
				return strc, nil
			}
		}
	}
	return nil, notFoundErr
}
