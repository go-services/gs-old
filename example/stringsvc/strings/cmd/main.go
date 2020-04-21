package main

import (
	"os"
	service "stringsvc/strings"
	"stringsvc/strings/gen"
	"stringsvc/strings/gen/cmd"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func main() {
	// Simple logger
	logger := log.NewLogfmtLogger(os.Stdout)
	// Http Router
	router := mux.NewRouter()

	// Make service
	svc := gen.MakeService(service.New())

	// Make endpoints
	eps := gen.MakeEndpoints(svc)

	// Make transports
	transports := gen.MakeTransports(eps)

	// Run service
	cmd.Run(transports, router, logger)
}
