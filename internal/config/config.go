package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Environment represents the application environment
type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

// Public, tiny struct that contains app configs
type App struct {
	Name        string      `yaml:"name"`
	Port        int         `yaml:"port"`
	Environment Environment `yaml:"environment"`
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

// GetEnvironment returns the current environment from APP_ENV or defaults to development
func GetEnvironment() Environment {
	env := os.Getenv("APP_ENV")
	switch env {
	case "production", "prod":
		return Production
	case "staging", "stage":
		return Staging
	default:
		return Development
	}
}

// Load .env → env vars
// Supports environment-specific .env files (.env.production, .env.staging, .env.development)
func LoadEnv() error {
	env := GetEnvironment()

	// Try to load environment-specific .env file first
	envFile := fmt.Sprintf(".env.%s", env)
	if err := godotenv.Load(envFile); err == nil {
		return nil
	}

	// Fall back to default .env file
	if err := godotenv.Load(); err != nil {
		// In production, .env file might not exist (using system env vars)
		if env == Production {
			return nil
		}
		return errors.New("error loading .env file")
	}

	return nil
}

// LoadConfig loads configuration from environment-specific YAML file
// Priority: config.<env>.yaml → config.yaml
func LoadConfig(baseFile string) (*Config, error) {
	env := GetEnvironment()

	// Try environment-specific config file first
	dir := filepath.Dir(baseFile)
	base := filepath.Base(baseFile)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	envFile := filepath.Join(dir, fmt.Sprintf("%s.%s%s", name, env, ext))

	// Try to load environment-specific config
	data, err := os.ReadFile(envFile)
	if err != nil {
		// Fall back to base config file
		data, err = os.ReadFile(baseFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Set environment from APP_ENV
	config.App.Environment = env

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

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	var errs []error

	// Validate App config
	if c.App.Name == "" {
		errs = append(errs, errors.New("app.name cannot be empty"))
	}
	if c.App.Port < 1 || c.App.Port > 65535 {
		errs = append(errs, fmt.Errorf("app.port must be between 1 and 65535, got %d", c.App.Port))
	}

	// Validate YouTube config
	if c.YouTube.APIKey == "" {
		errs = append(errs, errors.New("youtube.api_key cannot be empty"))
	}
	if c.YouTube.MaxResults < 1 || c.YouTube.MaxResults > 50 {
		errs = append(errs, fmt.Errorf("youtube.max_results must be between 1 and 50, got %d", c.YouTube.MaxResults))
	}

	// Validate Database config
	if c.Database.Path == "" {
		errs = append(errs, errors.New("database.path cannot be empty"))
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration validation failed: %v", errs)
	}

	return nil
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == Development
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.App.Environment == Production
}

// IsStaging returns true if running in staging environment
func (c *Config) IsStaging() bool {
	return c.App.Environment == Staging
}
