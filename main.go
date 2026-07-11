package main

import (
	"clean-template/cmd"
	"clean-template/helpers"
)

func main() {
	// load config
	helpers.SetupConfig()

	// load log
	helpers.SetupLogger()

	// load db
	helpers.SetupMySQL()

	// run grpc
	go cmd.ServeGRPC()

	// run http
	cmd.ServeHTTP()
}
