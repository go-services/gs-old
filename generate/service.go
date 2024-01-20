package generate

import (
	"fmt"
	"gs/assets"
	"gs/fs"
	"gs/parser"
	"log"
	"os/exec"
	"path"
)

type ServiceGenerator interface {
	Generate() error
}

type EndpointMethodData struct {
	Module           string
	Service          string
	ServiceImport    string
	ServiceInterface string
	Endpoint         parser.Endpoint
}

type serviceGenerator struct {
	services []parser.Service
}

func NewServiceGenerator(services []parser.Service) ServiceGenerator {
	return &serviceGenerator{
		services: services,
	}
}

func generateEndpoints(sv parser.Service, endpointPath string) error {
	definitionsFolder := path.Join(endpointPath, "definitions")
	if exists, _ := fs.Exists(definitionsFolder); !exists {
		_ = fs.CreateFolder(definitionsFolder)
	}
	// add method definition
	for _, ep := range sv.Endpoints {
		err := assets.ParseAndWriteTemplate(
			"services/endpoints/definitions/method.go.tmpl",
			path.Join(
				definitionsFolder,
				fmt.Sprintf("%s.go", ep.PackageName()),
			),
			ep,
		)
		if err != nil {
			return err
		}
		err = assets.ParseAndWriteTemplate(
			"services/endpoints/method.go.tmpl",
			path.Join(
				endpointPath,
				fmt.Sprintf("%s.go", ep.PackageName()),
			),
			EndpointMethodData{
				Module:           sv.Config.Module,
				Service:          sv.Package,
				ServiceImport:    sv.Import,
				ServiceInterface: sv.Interface,
				Endpoint:         ep,
			},
		)
		if err != nil {
			return err
		}
	}
	err := assets.ParseAndWriteTemplate(
		"services/endpoints/options.go.tmpl",
		path.Join(
			endpointPath,
			"options$.go",
		),
		sv,
	)
	if err != nil {
		return err
	}
	err = assets.ParseAndWriteTemplate(
		"services/endpoints/endpoint.go.tmpl",
		path.Join(
			endpointPath,
			"endpoint$.go",
		),
		sv,
	)
	if err != nil {
		return err
	}
	return nil
}

func generateHttpTransport(svc parser.Service, httpTransportPath string) error {
	globalTransportPth := path.Join(genPath(), "transport", "http")
	if exists, _ := fs.Exists(globalTransportPth); !exists {
		_ = fs.CreateFolder(globalTransportPth)
	}
	err := assets.ParseAndWriteTemplate(
		"transport/http/http.go.tmpl",
		path.Join(
			globalTransportPth,
			"http.go",
		),
		nil,
	)
	if err != nil {
		return err
	}

	for _, ep := range svc.Endpoints {
		err = assets.ParseAndWriteTemplate(
			"services/transport/http/method.go.tmpl",
			path.Join(
				httpTransportPath,
				fmt.Sprintf("%s.go", ep.PackageName()),
			),
			EndpointMethodData{
				Module:           svc.Config.Module,
				Service:          svc.Package,
				ServiceImport:    svc.Import,
				ServiceInterface: svc.Interface,
				Endpoint:         ep,
			},
		)
		if err != nil {
			return err
		}
	}
	err = assets.ParseAndWriteTemplate(
		"services/transport/http/http.go.tmpl",
		path.Join(
			httpTransportPath,
			"http$.go",
		),
		svc,
	)
	if err != nil {
		return err
	}

	return nil
}

func generateService(svc parser.Service) error {
	pth := path.Join(genPath(), "services", svc.Package)
	epFolder := path.Join(pth, "endpoint")
	httpTransportPath := path.Join(pth, "transport", "http")
	if exists, _ := fs.Exists(epFolder); !exists {
		_ = fs.CreateFolder(epFolder)
	}
	err := generateEndpoints(svc, epFolder)
	if err != nil {
		return err
	}

	err = generateHttpTransport(svc, httpTransportPath)
	if err != nil {
		return err
	}

	err = assets.ParseAndWriteTemplate(
		"services/service.go.tmpl",
		path.Join(
			pth,
			"service.go",
		),
		svc,
	)
	return err
}

func (g serviceGenerator) Generate() error {
	for _, svc := range g.services {
		err := generateService(svc)
		if err != nil {
			return err
		}
		lambdaHandler := path.Join(cmdPath(), svc.Name)
		if exists, _ := fs.Exists(lambdaHandler); !exists {
			_ = fs.CreateFolder(lambdaHandler)
		}

		lambdaHandler = path.Join(lambdaHandler, "lambda.go")
		if exists, _ := fs.Exists(lambdaHandler); !exists {
			err = assets.ParseAndWriteTemplate(
				"cmd/lambda/service.go.tmpl",
				lambdaHandler,
				svc,
			)
			if err != nil {
				return err
			}
		}
	}
	cmd := exec.Command("go", "mod", "tidy")

	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}

	genStack := path.Join("stacks", "gen")
	if exists, _ := fs.Exists(genStack); !exists {
		if exists, _ := fs.Exists(genStack); !exists {
			_ = fs.CreateFolder(genStack)
		}
	}
	err = assets.ParseAndWriteTemplate(
		"project/stacks/gen/gen.ts.tmpl",
		path.Join(genStack, "index.ts"),
		g.services,
	)
	if err != nil {
		return err
	}

	return nil
}
