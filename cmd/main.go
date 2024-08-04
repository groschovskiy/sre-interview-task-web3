package main

import (
	"cloud-lb/internal/backend"
	"cloud-lb/internal/config"
)

func main() {
	config := config.ReadConfigFile("config.json")
	proxy := backend.NewReverseProxy(config)

	go proxy.Serve()
	go proxy.LogMetrics()

	proxy.HandleGracefulShutdown()
}
