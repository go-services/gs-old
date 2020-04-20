package main

import (
	service "addsvc/add"
	"addsvc/add/gen"
	"addsvc/add/gen/cmd"
	"os"

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
