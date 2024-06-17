package main

import (
	"cloud-lb/internal/backend"
)

func main() {
	config := backend.ReadConfigFile("config.json")
	proxy := backend.NewReverseProxy(config)

	go proxy.Serve()

	backend.HandleGracefulShutdown(proxy)
}
