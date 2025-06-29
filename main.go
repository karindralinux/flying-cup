package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/docker/docker/client"
	"github.com/karindrlainux/flying-cup/pkg/deployment"
	"github.com/karindrlainux/flying-cup/pkg/notification"
	"github.com/karindrlainux/flying-cup/pkg/webhook"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func init() {
	checkSystemRequirements()
}

func main() {
	e := echo.New()

	config, err := LoadConfig("config.yaml")

	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	notifier := notification.NewGithubNotifier(config.Github.Token)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))

	e.POST("/webhook/github", webhook.HandleGithubWebhook(
		config.Github.WebhookSecret,
		// On PR Opened
		func(ctx context.Context, webhook *webhook.GithubPRWebhook) error {

			log.Printf("🚀 Starting deployment process for PR #%d", webhook.Number)
			log.Printf("📋 Deployment details:")
			log.Printf("   - Repository: %s", webhook.Repository.Name)
			log.Printf("   - Branch: %s", webhook.PullRequest.Head.Ref)
			log.Printf("   - PR Title: %s", webhook.PullRequest.Title)
			log.Printf("   - Author: %s", webhook.Sender.Username)

			previewURL, err := deployment.DeployPR(ctx, webhook)

			if err != nil {
				log.Printf("❌ Error deploying PR #%d: %v", webhook.Number, err)

				log.Printf("📤 Send deployment failure notification for PR #%d (%s)", webhook.Number, webhook.Repository.Name)
				failureComment := createDeploymentFailureComment(webhook, err.Error())

				err = notifier.CreateCommentPR(ctx, webhook, failureComment)

				if err != nil {
					log.Printf("❌ Error sending deployment failure notification for PR #%d (%s): %v", webhook.Number, webhook.Repository.Name, err)
				}

				return err
			}

			log.Printf("✅ Deployment successful for PR #%d (%s)", webhook.Number, webhook.Repository.Name)
			log.Printf("🌐 Preview URL: %s", previewURL)

			successComment := createDeploymentSuccessComment(webhook, previewURL)

			err = notifier.CreateCommentPR(ctx, webhook, successComment)

			if err != nil {
				log.Printf("❌ Error sending deployment success notification for PR #%d (%s): %v", webhook.Number, webhook.Repository.Name, err)
			}

			log.Printf("✅ Deployment success notification sent for PR #%d (%s)", webhook.Number, webhook.Repository.Name)

			return nil
		},
		// On PR Closed
		func(ctx context.Context, webhook *webhook.GithubPRWebhook) error {
			log.Printf("🧹 Cleaning up deployment for PR #%d (%s)", webhook.Number, webhook.Repository.Name)
			log.Printf("📋 Cleanup details:")
			log.Printf("   - Repository: %s", webhook.Repository.Name)
			log.Printf("   - PR: #%d", webhook.Number)

			err := deployment.Cleanup(ctx, webhook)

			if err != nil {
				log.Printf("❌ Error cleaning up deployment for PR #%d (%s): %v", webhook.Number, webhook.Repository.Name, err)
			}

			log.Printf("📤 Send deployment cleanup success notification for PR #%d (%s)", webhook.Number, webhook.Repository.Name)

			successComment := createCleanupSuccessComment(webhook)

			err = notifier.CreateCommentPR(ctx, webhook, successComment)

			if err != nil {
				log.Printf("❌ Error sending deployment cleanup success notification for PR #%d (%s): %v", webhook.Number, webhook.Repository.Name, err)
			}

			log.Printf("✅ Deployment cleanup completed for PR #%d (%s)", webhook.Number, webhook.Repository.Name)

			return nil
		},
	))

	e.Logger.Fatal(e.Start(":8080"))
}

func createDeploymentFailureComment(webhook *webhook.GithubPRWebhook, errorMessage string) string {
	return fmt.Sprintf(`## ❌ Deployment failed
	
**Error :** %s

**Details :**
- Repository : %s
- Branch : %s
- PR Title : %s
- Triggered by : %s

Please check your app and deployment configuration.
	`, errorMessage, webhook.Repository.Name, webhook.PullRequest.Head.Ref, webhook.PullRequest.Title, webhook.Sender.Username)
}

func createDeploymentSuccessComment(webhook *webhook.GithubPRWebhook, previewURL string) string {
	return fmt.Sprintf(`## 🚀 Preview Deployment Successful!

Your preview is now available at: **%s**

**Details:**
- Repository: %s
- Branch: %s
- PR: #%d

The preview will be automatically cleaned up when this PR is closed.`, previewURL, webhook.Repository.Name, webhook.PullRequest.Head.Ref, webhook.Number)
}

func createCleanupSuccessComment(webhook *webhook.GithubPRWebhook) string {
	return fmt.Sprintf(`## 🧹 Preview Cleanup Completed

The preview deployment for PR #%d has been successfully cleaned up.

**Cleaned up resources:**
- Docker container
- Docker image
- Cloned repository`,
		webhook.Number)
}

func checkSystemRequirements() {
	// Check system requirements
	log.Println("Checking system requirements...")

	// Skip Docker CLI checks when running in container
	// The Docker socket is mounted, so we can connect directly
	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		log.Println("Docker socket found - running in container environment ✅")
		log.Println("Skipping Docker CLI checks (not needed in container)")

		// Test Docker API connection
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Printf("Warning: Could not create Docker client: %v", err)
		} else {
			// Test the connection
			_, err = cli.Ping(context.Background())
			if err != nil {
				log.Printf("Warning: Could not ping Docker daemon: %v", err)
			} else {
				log.Println("Docker API connection successful ✅")
			}
		}
		return
	}

	// Only check Docker CLI when running on host
	dockerInstalled := exec.Command("docker", "--version")

	if err := dockerInstalled.Run(); err != nil {
		log.Fatalf("Docker is not installed. Please install Docker and try again.")
	}

	log.Println("Docker is installed ✅")

	// Check if Docker is running
	dockerRunning := exec.Command("docker", "ps")

	if err := dockerRunning.Run(); err != nil {
		log.Fatalf("Docker is not running. Please start Docker and try again.")
	}

	log.Println("Docker is running ✅")

	log.Println("All system requirements checked ✅")
}
