package service

import (
	"strings"

	"github.com/go-services/code"

	"github.com/sirupsen/logrus"

	"github.com/go-services/annotation"
)

type paramType string
type requestFormat string

var log = logrus.WithFields(logrus.Fields{
	"package": "service",
})

const (
	URL   paramType = "URL"
	QUERY paramType = "QUERY"
	BODY  paramType = "BODY"
)

const (
	JSON requestFormat = "JSON"
	XML  requestFormat = "XML"
	FORM requestFormat = "FORM"
)

type ParamParser struct {
	Fn      string
	NoError bool
}

var typeFuncMap = map[string]*ParamParser{
	"[]string": {
		Fn:      "StringToStringArray",
		NoError: true,
	},
	"int": {
		Fn: "StringToInt",
	},
	"[]int": {
		Fn: "StringToIntArray",
	},
	"float64": {
		Fn: "StringToFloat64",
	},
	"[]float64": {
		Fn: "StringToFloat64Array",
	},
	"float32": {
		Fn: "StringToFloat32",
	},
	"[]float32": {
		Fn: "StringToFloat32Array",
	},
	"bool": {
		Fn: "StringToBool",
	},
}

type HttpRequestParam struct {
	// this is the field name
	Field string
	// this is the name given in the url param or query param
	Name string
	// this is the field type
	Type code.Type
	// is this parameter optional
	Required bool
	// this tells us of it is a URL param or a Query param
	ParamType paramType
	// parameter parse function
	Parser *ParamParser
}

type HttpRequest struct {
	// the format the data is
	Format requestFormat
	// if the request has any query params
	HasUrl bool
	// if the request has body portion
	HasBody bool
	// all the extra params
	Params []HttpRequestParam
}

type HttpMethodRoute struct {
	Name    string
	Methods []string
	Route   string
}

type HttpTransport struct {
	Request        *HttpRequest
	ResponseFormat string
	MethodRoutes   []HttpMethodRoute
}

func ParseHttpTransport(endpoint Endpoint) (*HttpTransport, error) {
	httpAnnotations := findAnnotations("http", endpoint.Annotations)
	if len(httpAnnotations) == 0 {
		return nil, nil
	}

	return &HttpTransport{
		MethodRoutes:   parseMethodRoutes(httpAnnotations[0]),
		Request:        parseHttpRequest(endpoint),
		ResponseFormat: string(httpResponseFormat(httpAnnotations[0].Get("response").String())),
	}, nil
}

func httpResponseFormat(format string) requestFormat {
	if format == "" {
		return JSON
	}
	switch requestFormat(strings.ToUpper(format)) {
	case XML:
		return XML
	default:
		return JSON
	}
}
func parseHttpRequest(endpoint Endpoint) *HttpRequest {
	if endpoint.Request == nil {
		return nil
	}

	httpAnnotations := findAnnotations("http", endpoint.Annotations)
	httpEncode := httpAnnotations[0]
	annotationFormat := httpEncode.Get("request").String()
	format := JSON
	if annotationFormat != "" {
		switch requestFormat(strings.ToUpper(annotationFormat)) {
		case XML:
			format = XML
		case FORM:
			format = FORM
		case JSON:
			format = JSON
		default:
			log.WithField("endpoint", endpoint.Name).Info("The request format is not supported `json` will be used as default")
		}
	}
	request := &HttpRequest{
		Format: format,
	}
	parseHttpRequestParams(endpoint.Request, request)
	return request
}

func parseHttpRequestParams(req *code.Struct, request *HttpRequest) {
	for _, field := range req.Fields {
		if !isExported(field.Name) || field.Tags == nil {
			continue
		}

		gsUrl := getTag("url", *field.Tags)
		gsQuery := getTag("query", *field.Tags)
		gsBody := getTag("body", *field.Tags)

		tp := field.Type.String()

		if gsUrl != "" {
			if !isUrlTypeSupported(tp) {
				log.WithField("field", field.Name).WithField("type", field.Type.String()).Warn("Field type not supported for url")
				continue
			}
			var parser *ParamParser = nil
			if tp != "string" {
				parser = typeFuncMap[tp]
			}
			name, required := getParameter(gsUrl)
			request.Params = append(request.Params, HttpRequestParam{
				Field:     field.Name,
				Name:      name,
				Type:      field.Type,
				Required:  required,
				ParamType: URL,
				Parser:    parser,
			})
			request.HasUrl = true
		}
		if gsQuery != "" {
			if !isQueryTypeSupported(tp) {
				log.WithField("field", field.Name).WithField("type", field.Type.String()).Warn("Field type not supported for query")
				continue
			}
			var parser *ParamParser = nil
			if tp != "string" {
				parser = typeFuncMap[tp]
			}

			name, required := getParameter(gsQuery)
			request.Params = append(request.Params, HttpRequestParam{
				Field:     field.Name,
				Name:      name,
				Type:      field.Type,
				Required:  required,
				ParamType: QUERY,
				Parser:    parser,
			})
		}
		if gsBody != "" {
			name, required := getParameter(gsBody)
			format := JSON
			switch requestFormat(strings.ToUpper(name)) {
			case XML:
				format = XML
			case FORM:
				format = FORM
			case JSON:
				format = JSON
			default:
				log.WithField("endpoint", field.Name).Info("The request format is not supported `json` will be used as default")
			}
			request.Params = append(request.Params, HttpRequestParam{
				Field:     field.Name,
				Name:      string(format),
				Required:  required,
				ParamType: BODY,
			})
			request.HasBody = true
		}
	}
	return
}

func parseMethodRoutes(httpAnnotation annotation.Annotation) (routes []HttpMethodRoute) {
	keepTrailingSlash := httpAnnotation.Get("keepTrailingSlash").Bool()
	var methodsPrepared []string
	for _, method := range strings.Split(httpAnnotation.Get("methods").String(), ",") {
		methodsPrepared = append(methodsPrepared, strings.ToUpper(strings.TrimSpace(method)))
	}
	route := httpAnnotation.Get("route").String()
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	methodRoute := HttpMethodRoute{
		Name:    httpAnnotation.Get("name").String(),
		Methods: methodsPrepared,
		Route:   route,
	}
	routes = append(
		routes,
		methodRoute,
	)
	if !keepTrailingSlash {
		if strings.HasSuffix(methodRoute.Route, "/") {
			route = strings.TrimSuffix(route, "/")
		} else {
			route += "/"
		}
		methodRoute.Route = route
		routes = append(
			routes,
			methodRoute,
		)
	}
	return
}

func isQueryTypeSupported(tp string) bool {
	var supportedQueryTypes = []string{
		"string",
		"[]string",
		"int",
		"[]int",
		"bool",
		"float32",
		"[]float32",
		"float64",
		"[]float64",
	}
	found := false
	for _, supportedType := range supportedQueryTypes {
		if supportedType == tp {
			found = true
			break
		}
	}
	return found
}

func isUrlTypeSupported(tp string) bool {
	var supportedUrlTypes = []string{"string", "int", "float32", "float64"}
	found := false
	for _, supportedType := range supportedUrlTypes {
		if supportedType == tp {
			found = true
			break
		}
	}
	return found
}
