package deployment

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strconv"

	"github.com/docker/docker/client"
	"github.com/karindrlainux/flying-cup/pkg/docker"
	"github.com/karindrlainux/flying-cup/pkg/git"
	"github.com/karindrlainux/flying-cup/pkg/types"
	"github.com/karindrlainux/flying-cup/pkg/webhook"
)

func Cleanup(ctx context.Context, webhook *webhook.GithubPRWebhook) error {
	log.Printf("Cleaning up deployment for PR #%d", webhook.Number)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	// Get container name and retrieve port from running container
	containerName := fmt.Sprintf("pr-%s-%d", webhook.Repository.Name, webhook.Number)
	dockerRunner := &docker.DockerRunner{Client: cli}

	hostPortStr, err := dockerRunner.GetContainerHostPort(ctx, containerName)
	if err != nil {
		return fmt.Errorf("failed to retrieve host port for container %s: %w", containerName, err)
	}

	// Release the port
	if hostPort, err := strconv.Atoi(hostPortStr); err == nil {
		portManager.ReleasePort(hostPort)
		log.Printf("Released port %d for PR #%d", hostPort, webhook.Number)
	}

	// Stop & Remove Container
	err = dockerRunner.RemoveContainerIfExists(ctx, &types.App{Name: containerName})

	if err != nil {
		return fmt.Errorf("failed to remove Docker container: %w", err)
	}

	// Stop & Remove Image
	dockerBuilder := &docker.DockerBuilder{Client: cli}
	imageTag := fmt.Sprintf("pr-%d", webhook.Number)
	err = dockerBuilder.RemoveImage(ctx, imageTag)

	if err != nil {
		return fmt.Errorf("failed to remove Docker image: %w", err)
	}

	// Stop & Remove Repo
	repoPath := filepath.Join("./repos", fmt.Sprintf("pr-%d", webhook.Number))
	err = git.RemoveClonedRepository(ctx, repoPath)

	if err != nil {
		return fmt.Errorf("failed to remove cloned repository: %w", err)
	}

	return nil
}
