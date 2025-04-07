package main

import "github.com/Isaac-Franklyn/dist-kvstore/internal/application/adapters/httpserver"

func main() {

	server := httpserver.NewHTTPServer()
	server.Start()

}
