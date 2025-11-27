package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Public, tiny struct that contains app configs
type App struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

// Public, tiny struct that contains YouTube client configs
type YouTube struct {
	APIKey     string `yaml:"api_key"`
	MaxResults int64  `yaml:"max_results"`
}

// Public, tiny struct that contains database configs
type Database struct {
	Path string `yaml:"path"`
}

// Full app config
type Config struct {
	App      App      `yaml:"app"`
	YouTube  YouTube  `yaml:"youtube"`
	Database Database `yaml:"database"`
}

// Load .env → env vars
func LoadEnv() error {
	if err := godotenv.Load(); err != nil {
		return errors.New("error loading .env file")
	}
	return nil
}

// Load config.yaml → Config (without secret)
func LoadConfig(file string) (*Config, error) {
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

// Inject secret from env var into the struct
func InjectEnvVariables(config *Config) error {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return errors.New("YOUTUBE_API_KEY environment variable not set")
	}
	config.YouTube.APIKey = apiKey
	return nil
}
