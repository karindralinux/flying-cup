package providers

import (
	"context"

	"github.com/karindrlainux/flying-cup/pkg/webhook"
)

// Provider defines the interface for different deployment backends
type Provider interface {
	// Initialize the provider with configuration
	Init(config interface{}) error

	// Create a deployment and return the preview URL
	CreateDeployment(ctx context.Context, webhook *webhook.GithubPRWebhook) (string, error)

	// Clean up a deployment
	CleanupDeployment(ctx context.Context, repoName, prName string, prNumber int) error

	// Get deployment status
	GetDeploymentStatus(ctx context.Context, deploymentID string) (string, error)
}

// Config holds common configuration for all providers
type Config struct {
	Domain      string
	Port        int
	Environment string
}

// Type defines supported deployment providers
type Type string

const (
	TypeTraefik    Type = "traefik"
	TypeNginx      Type = "nginx"
	TypeCloudflare Type = "cloudflare"
)
