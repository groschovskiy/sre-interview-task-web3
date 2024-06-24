package health

import (
	"cloud-lb/internal/config"
	"log"
	"net/http"
	"time"
)

func healthCheck(server string, config *config.HealthCheckConfig) bool {
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	url := server + config.Path
	req, err := http.NewRequest(config.Method, url, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == config.RespCode
}

func RunHealthChecks(servers []string, config *config.HealthCheckConfig) {
	for {
		log.Println("Runned RunHealthChecks function")
		for _, server := range servers {
			if !healthCheck(server, config) {
				// TODO: - Remove the unhealthy server from the pool or send alert
			}
		}
		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}
