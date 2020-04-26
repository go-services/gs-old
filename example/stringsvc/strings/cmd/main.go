package main

import (
	service "stringsvc/strings"
	"stringsvc/strings/gen"
)

func main() {
	gen.New(service.New()).Run()
}
