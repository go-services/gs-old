package service

import (
	"errors"
	"fmt"
	"gs/config"
	"gs/fs"
	"gs/template"
	"path"
	"strings"

	"github.com/go-services/annotation"

	"github.com/ozgio/strutil"

	"github.com/spf13/viper"

	"github.com/go-services/source"
)

type Middleware struct {
	Alias  string
	Import string
	Method string
}

type Service struct {
	Interface string
	Config    config.ServiceConfig
	Import    string
	Package   string

	Endpoints   []Endpoint
	Middlewares []Middleware
}

// this is used from findStruct() if the file was already read we don't
// want to spend all the time to read and parse it again
var fileSourceCache map[string]*source.Source

func Generate(config config.ServiceConfig, module string) error {
	fileSourceCache = map[string]*source.Source{}

	src, err := readServiceSource(config.Name)
	if err != nil {
		return err
	}

	inf := findServiceInterface(src)
	if inf == nil {
		return fmt.Errorf(
			"error while parsing service : %s",
			"Could not find service interface, make sure you are using @service()",
		)
	}

	service := Service{
		Interface: inf.Name(),
		Config:    config,

		Import:  fmt.Sprintf("%s/%s", module, config.Name),
		Package: src.Package(),
	}

	for _, method := range filterMethods(inf.Methods()) {
		ep, err := parseEndpoint(method, service.Import, service.Name())
		if err != nil {
			return err
		}
		service.Endpoints = append(service.Endpoints, *ep)
	}
	service.Middlewares = parseMiddleware(source.FindAnnotations("middleware", inf))

	return service.generateFiles()
}

func (s *Service) generateEndpoints() error {
	for _, endpoint := range s.Endpoints {
		endpointFile := strutil.ToSnakeCase(endpoint.Name) + ".go"
		files := map[string]string{
			"service/gen/endpoint/definitions/method.jet": s.GetPath("gen", "endpoint", "definitions", endpointFile),
			"service/gen/endpoint/method.jet":             s.GetPath("gen", "endpoint", endpointFile),
		}
		if endpoint.HttpTransport != nil {
			files["service/gen/transport/http/method.jet"] = s.GetPath("gen", "transport", "http", endpointFile)
		}

		for k, v := range files {
			src, err := template.CompileGoFromPath(k, struct {
				Endpoint Endpoint
				Service  Service
			}{
				Endpoint: endpoint,
				Service:  *s,
			})
			if err != nil {
				return err
			}
			err = fs.WriteFile(v, src)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (s *Service) generateFiles() error {
	err := fs.DeleteFolder(s.GetPath("gen"))
	if err != nil {
		return err
	}
	files := map[string]string{
		"service/gen/service.jet":             s.GetPath("gen", "gen.go"),
		"service/gen/service/service.jet":     s.GetPath("gen", "service", "service.go"),
		"service/gen/cmd/cmd.jet":             s.GetPath("gen", "cmd", "cmd.go"),
		"service/gen/errors/errors.jet":       s.GetPath("gen", "errors", "errors.go"),
		"service/gen/errors/http.jet":         s.GetPath("gen", "errors", "http.go"),
		"service/gen/utils/utils.jet":         s.GetPath("gen", "utils", "utils.go"),
		"service/gen/endpoint/endpoint.jet":   s.GetPath("gen", "endpoint", "endpoint.go"),
		"service/gen/transport/transport.jet": s.GetPath("gen", "transport", "transport.go"),
		"service/gen/transport/http/http.jet": s.GetPath("gen", "transport", "http", "http.go"),
	}

	for k, v := range files {
		src, err := template.CompileGoFromPath(k, s)
		if err != nil {
			return err
		}
		err = fs.WriteFile(v, src)
		if err != nil {
			return err
		}
	}
	if err := s.generateEndpoints(); err != nil {
		return err
	}
	if err := s.generateCmd(); err != nil {
		return err
	}
	return nil
}

func (s Service) generateCmd() error {
	if b, err := fs.Exists(s.GetPath("cmd", "main.go")); err != nil {
		return err
	} else if b {
		return nil
	}
	src, err := template.CompileGoFromPath(
		"service/cmd/main.jet",
		s,
	)
	if err != nil {
		return err
	}
	return fs.WriteFile(s.GetPath("cmd", "main.go"), src)
}

func (s *Service) GetPath(pth ...string) string {
	return path.Join(append([]string{s.Name()}, pth...)...)
}
func (s *Service) Name() string {
	return s.Config.Name
}

func readServiceSource(name string) (*source.Source, error) {
	data, err := fs.ReadFile(fmt.Sprintf("%s/service.go", name))
	if err != nil {
		return nil, errors.New("A read error occurred. Please update your code..: " + err.Error())
	}
	src, err := source.New(data)
	if err != nil {
		return nil, errors.New("A read error occurred. Please update your code..: " + err.Error())
	}
	return src, nil
}

func findServiceInterface(src *source.Source) *source.Interface {
	for _, inf := range src.Interfaces() {
		annotations := source.FindAnnotations(viper.GetString(config.ServiceAnnotation), &inf)
		if len(annotations) > 0 {
			return &inf
		}
	}
	return nil
}

func filterMethods(methods []source.InterfaceMethod) (list []source.InterfaceMethod) {
	for _, method := range methods {
		if isExported(method.Name()) {
			list = append(list, method)
		}
	}
	return list
}

func parseMiddleware(annotations []annotation.Annotation) (mdw []Middleware) {
	packages := map[string]string{}
	for _, v := range annotations {
		middlewarePath := v.Get("path").String()
		pathParts := strings.Split(middlewarePath, ".")
		if len(pathParts) == 1 {
			mdw = append(mdw, Middleware{
				Alias:  "service",
				Method: pathParts[0],
			})
			continue
		}
		middleware := Middleware{
			Alias:  "",
			Method: pathParts[len(pathParts)-1],
		}
		middlewarePackage := strings.Join(pathParts[:len(pathParts)-1], "/")
		if v, ok := packages[middlewarePackage]; ok {
			middleware.Alias = v
		} else {
			packages[middlewarePackage] = fmt.Sprintf("mdw%d", len(packages)+1)
			middleware.Alias = packages[middlewarePackage]
		}
		middleware.Import = middlewarePackage
		mdw = append(mdw, middleware)
	}
	return
}

func New(name string) error {
	cfg, err := config.Read()
	if err != nil {
		return err
	}

	// we should remove the '_' because of this guide https://blog.golang.org/package-names
	serviceName := strings.ReplaceAll(strutil.ToSnakeCase(name), "_", "")

	if err := fs.CreateFolder(serviceName); err != nil {
		return err
	}

	data := map[string]string{
		"Name": serviceName,
	}

	src, err := template.CompileGoFromPath("service/service.jet", data)
	if err != nil {
		return err
	}
	err = fs.WriteFile(path.Join(serviceName, "service.go"), src)
	if err != nil {
		return err
	}

	port := 8000
	for _, v := range cfg.Services {
		if v.Http.Port == port {
			port += 1
		}
	}
	cfg.Services = append(cfg.Services, config.ServiceConfig{
		Name: serviceName,
		Http: config.HttpConfig{
			Port: port,
		},
	})
	return config.Update(*cfg)
}
