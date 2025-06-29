package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Github GithubConfig `yaml:"github"`
}

type ServerConfig struct {
	Environment string `yaml:"environment"`
	BaseURL string `yaml:"base_url"`
}

type GithubConfig struct {
	AppID         string `yaml:"app_id"`
	WebhookSecret string `yaml:"webhook_secret"`
	Token         string `yaml:"token"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)

	if config.Server.BaseURL == "" {
		return nil, fmt.Errorf("server base url is required")
	}

	if config.Server.Environment == "" || config.Server.Environment == "local" {
		config.Server.BaseURL = "http://localhost"
	}

	if config.Github.WebhookSecret == "" {
		return nil, fmt.Errorf("github webhook secret is required")
	}

	if config.Github.Token == "" {
		return nil, fmt.Errorf("github token is required")
	}

	return &config, err
}
