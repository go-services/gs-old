package main

import (
	service "addsvc/add"
	"addsvc/add/gen"
	"addsvc/add/gen/cmd"
	"addsvc/add/middleware"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func main() {
	// Simple logger
	logger := log.NewLogfmtLogger(os.Stdout)
	// Http Router
	router := mux.NewRouter()

	// Create the (sparse) metrics we'll use in the service. They, too, are
	// dependencies that we pass to components that use them.
	var ints, chars metrics.Counter
	{
		// Business-level metrics.
		ints = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "example",
			Subsystem: "addsvc",
			Name:      "integers_summed",
			Help:      "Total count of integers summed via the Sum method.",
		}, []string{})
		chars = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "example",
			Subsystem: "addsvc",
			Name:      "characters_concatenated",
			Help:      "Total count of characters concatenated via the Concat method.",
		}, []string{})
	}

	// Make service
	svc := gen.MakeService(
		service.New(),
		middleware.LoggingMiddleware(logger),
		middleware.InstrumentingMiddleware(ints, chars),
	)

	http.DefaultServeMux.Handle("/metrics", promhttp.Handler())

	// Make endpoints
	eps := gen.MakeEndpoints(svc)

	// Make transports
	transports := gen.MakeTransports(eps)

	// Run service
	cmd.Run(transports, router, logger)
}
