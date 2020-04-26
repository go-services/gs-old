package main

import (
	service "addsvc/add"
	"addsvc/add/gen"
	"addsvc/add/middleware"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

func main() {
	logger := log.NewLogfmtLogger(os.Stdout)

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

	gen.New(
		service.New(),
		gen.Logger(logger),
		gen.ServiceMode(gen.DEBUG),

		gen.ServiceMiddleware(
			middleware.LoggingMiddleware(logger),
			middleware.InstrumentingMiddleware(ints, chars),
		),
	).Run()
}
