package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	AuthToken       []string `json:"valid_tokens"`
	BackendServers  []string `json:"backend_servers"`
	LoadBalanceAlgo string   `json:"load_balancing_algorithm"`
}

func ReadConfigFile(filename string) *Config {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open configuration file: %v", err)
	}
	defer file.Close()

	config := &Config{}
	if err := json.NewDecoder(file).Decode(config); err != nil {
		log.Fatalf("Failed to parse configuration file: %v", err)
	}

	return config
}
