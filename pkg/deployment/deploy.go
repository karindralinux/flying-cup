package deployment

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/karindrlainux/flying-cup/pkg/providers"
	"github.com/karindrlainux/flying-cup/pkg/webhook"
)

// Deploy from pull request webhook
func DeployPullRequest(ctx context.Context, webhook *webhook.GithubPRWebhook, provider providers.Provider) (string, error) {
	log.Printf("Starting deployment for PR #%d", webhook.Number)
	log.Printf("Repository: %s", webhook.Repository.Name)
	log.Printf("Branch: %s", webhook.PullRequest.Head.Ref)
	log.Printf("PR Title: %s", webhook.PullRequest.Title)

	// Use the provider to create deployment
	previewURL, err := provider.CreateDeployment(ctx, webhook)
	if err != nil {
		return "", fmt.Errorf("failed to create deployment: %w", err)
	}

	// Wait for deployment to be ready
	log.Printf("Waiting for deployment to be ready...")
	time.Sleep(5 * time.Second)

	log.Printf("✅ Successfully deployed PR #%d", webhook.Number)
	log.Printf("🌐 Preview available at: %s", previewURL)

	return previewURL, nil
}
