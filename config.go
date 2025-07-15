package main

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server ServerConfig
	Github GithubConfig
}

type ServerConfig struct {
	Environment string
	Domain      string
	Port        int
}

type GithubConfig struct {
	AppID         string
	WebhookSecret string
	Token         string
}

func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Environment: getEnv("ENVIRONMENT", "local"),
			Domain:      getEnv("DOMAIN", "localhost"),
			Port:        getEnvAsInt("PORT", 80),
		},
		Github: GithubConfig{
			AppID:         getEnv("GITHUB_APP_ID", ""),
			WebhookSecret: getEnv("GITHUB_WEBHOOK_SECRET", ""),
			Token:         getEnv("GITHUB_TOKEN", ""),
		},
	}

	// Validate required fields
	if config.Github.WebhookSecret == "" {
		return nil, fmt.Errorf("GITHUB_WEBHOOK_SECRET is required")
	}

	if config.Github.Token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is required")
	}

	if config.Server.Domain == "" {
		return nil, fmt.Errorf("DOMAIN is required")
	}

	return config, nil
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Domain getter method
func (c *Config) GetDomain() string {
	return c.Server.Domain
}

// GetProtocol returns http or https based on environment
func (c *Config) GetProtocol() string {
	if c.Server.Environment == "local" {
		return "http"
	}
	return "https"
}
