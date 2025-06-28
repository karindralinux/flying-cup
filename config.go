package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Github GithubConfig `yaml:"github"`
}

type GithubConfig struct {
	AppID         string `yaml:"app_id"`
	WebhookSecret string `yaml:"webhook_secret"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)

	if config.Github.WebhookSecret == "" {
		return nil, fmt.Errorf("github webhook secret is required")
	}

	return &config, err
}
