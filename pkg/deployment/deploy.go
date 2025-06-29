package deployment

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/karindrlainux/flying-cup/pkg/docker"
	"github.com/karindrlainux/flying-cup/pkg/git"
	"github.com/karindrlainux/flying-cup/pkg/port"
	"github.com/karindrlainux/flying-cup/pkg/types"
	"github.com/karindrlainux/flying-cup/pkg/webhook"
)

var portManager = port.NewPortManager(8080, 65535)

func generateCodeName(webhook *webhook.GithubPRWebhook) string {
	return fmt.Sprintf("pr-%s-%d", webhook.Repository.Name, webhook.Number)
}

func DeployPR(ctx context.Context, webhook *webhook.GithubPRWebhook) (string, error) {

	log.Printf("Starting deployment for PR #%d", webhook.Number)
	log.Printf("Repository URL: %s", webhook.Repository.CloneUrl)
	log.Printf("Branch: %s", webhook.PullRequest.Head.Ref)
	log.Printf("PR Title: %s", webhook.PullRequest.Title)

	hostPort, err := portManager.GetAvailablePort()

	if err != nil {
		return "", fmt.Errorf("failed to get available port: %w", err)
	}
	log.Printf("Using host port: %d", hostPort)

	codeName := generateCodeName(webhook)

	repoPath := filepath.Join("./repos", codeName)
	log.Printf("Target path: %s", repoPath)

	err = git.CloneRepository(ctx, webhook.Repository.CloneUrl, webhook.PullRequest.Head.Ref, repoPath)

	if err != nil {
		portManager.ReleasePort(hostPort)
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		portManager.ReleasePort(hostPort)
		return "", fmt.Errorf("failed to create docker client: %w", err)
	}

	dockerBuilder := &docker.DockerBuilder{Client: cli}

	app := &types.App{
		Name:          codeName,
		SourcePath:    repoPath,
		HostPort:      fmt.Sprintf("%d", hostPort),
		ContainerPort: "8080",
	}

	imageTag, err := dockerBuilder.BuildImage(ctx, app, "Dockerfile", true)

	if err != nil {
		portManager.ReleasePort(hostPort)
		return "", fmt.Errorf("failed to build Docker image: %w", err)
	}

	dockerRunner := &docker.DockerRunner{Client: cli}
	containerName := codeName
	containerId, err := dockerRunner.RunContainer(ctx, app, imageTag, containerName)

	if err != nil {
		portManager.ReleasePort(hostPort)
		return "", fmt.Errorf("failed to run Docker container: %w", err)
	}

	log.Printf("Container %s started with ID %s", containerName, containerId)

	log.Printf("Successfully deployed PR #%d (%s) to %s", webhook.Number, webhook.Repository.CloneUrl, containerId)

	previewURL := fmt.Sprintf("http://localhost:%s", fmt.Sprintf("%d", hostPort))
	log.Printf("Preview available at: %s", previewURL)

	return previewURL, nil
}
