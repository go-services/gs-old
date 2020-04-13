package watch

import (
	"fmt"
	"gs/config"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
)

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	// parse the url
	url, err := url.Parse(target)
	if err != nil {
		panic(err)
	}
	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host
	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}

func handleServiceRequest(service config.ServiceConfig) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		address := service.Http.Url
		if address == "" {
			address = "http://localhost"
		}
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/"+service.Name)
		serveReverseProxy(fmt.Sprintf("%s:%d", address, service.Http.Port), res, req)
	}
}

func proxyServices(port int) {
	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}
	r := mux.NewRouter()
	log.Infof("Starting server at: http://localhost:%d", port)
	for _, v := range cfg.Services {
		log.Infof("Proxy requests for service %s to http://localhost:%d/%s", v.Name, port, v.Name)
		r.HandleFunc(fmt.Sprintf("/%s", v.Name), handleServiceRequest(v))
		r.HandleFunc(fmt.Sprintf("/%s/", v.Name), handleServiceRequest(v))
		r.HandleFunc(fmt.Sprintf(`/%s/{rest:[a-zA-Z0-9=\-\/]+}`, v.Name), handleServiceRequest(v))
	}
	http.Handle("/", r)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}
