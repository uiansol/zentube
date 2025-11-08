package main

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Youtube struct {
		ApiKey     string `yaml:"api_key"`
		MaxResults int    `yaml:"max_results"`
	} `yaml:"youtube"`
}

func loadEnv() error {
	if err := godotenv.Load(); err != nil {
		return errors.New("error loading .env file")
	}
	return nil
}

func loadConfig(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func injectEnvVariables(config *Config) error {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		panic("YOUTUBE_API_KEY environment variable not set")
	}
	config.Youtube.ApiKey = apiKey
	return nil
}

func main() {
	if err := loadEnv(); err != nil {
		panic(err)
	}

	config, err := loadConfig("configs/config.yaml")
	if err != nil {
		panic(err)
	}

	if err := injectEnvVariables(config); err != nil {
		panic(err)
	}
}
