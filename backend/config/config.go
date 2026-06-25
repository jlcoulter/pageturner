package config

import (
	//"fmt"
	"log"
	"os"

	"go.yaml.in/yaml/v4"
)

type ConfigWrapper struct {
	Config struct {
		Database struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
			DbName   string `yaml:"dbname"`
			SslMode  string `yaml:"sslmode"`
		} `yaml:"database"`
	} `yaml:"config"`
}

func InitConfig() ConfigWrapper {
	var cfg ConfigWrapper

	filePath := "./config.yml"
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read config file %s: %v", filePath, err)
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("Failed to parse YAML config: %v", err)
	}

	return cfg
}
