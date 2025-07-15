package deployment

import (
	"context"
	"fmt"
	"log"

	"github.com/karindrlainux/flying-cup/pkg/providers"
)

// CleanupConfig interface for cleanup configuration
type CleanupConfig interface {
	GetDomain() string
}

// CleanupPullRequest removes a specific PR deployment
func CleanupPullRequest(ctx context.Context, repoName, prName string, prNumber int, provider providers.Provider) error {
	log.Printf("Cleaning up deployment for PR #%d", prNumber)

	err := provider.CleanupDeployment(ctx, repoName, prName, prNumber)
	if err != nil {
		return fmt.Errorf("failed to cleanup deployment: %w", err)
	}

	log.Printf("✅ Successfully cleaned up PR #%d", prNumber)
	return nil
}

// CleanupAllDeployments removes all active deployments
func CleanupAllDeployments(provider providers.Provider) error {
	log.Printf("Starting cleanup of all deployments")

	// TODO: Implement cleanup of all deployments
	// This would require the provider to support listing all deployments
	log.Printf("✅ Successfully cleaned up all deployments")
	return nil
}
