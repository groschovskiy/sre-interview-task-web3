package main

import (
	"cloud-lb/internal/backend"
	"cloud-lb/internal/config"
	"cloud-lb/internal/health"
)

func main() {
	config := config.ReadConfigFile("config.json")
	proxy := backend.NewReverseProxy(config)

	go proxy.Serve()

	if config.HealthCheck != nil {
		go health.RunHealthChecks(config.BackendServers, config.HealthCheck)
	}

	proxy.HandleGracefulShutdown()
}
