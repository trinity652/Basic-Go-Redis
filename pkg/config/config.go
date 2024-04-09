// File: pkg/config/config.go

package config

import (
	"encoding/json"
	"os"
)

// Config defines the structure of the configuration parameters
type Config struct {
	ServerHost string `json:"server_host"`
	ServerPort string `json:"server_port"`
	LogLevel   string `json:"log_level"`
}

// LoadConfig reads configuration from a file
func LoadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
