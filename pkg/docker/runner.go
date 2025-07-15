package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"

	"github.com/docker/docker/client"
	sharedTypes "github.com/karindrlainux/flying-cup/pkg/types"
)

type DockerRunner struct {
	Client *client.Client
}

// RunContainerWithTraefik runs a container with Traefik integration
func (d *DockerRunner) RunContainerWithTraefik(ctx context.Context, app *sharedTypes.App, imageTag, containerName string, labels map[string]string) (string, error) {
	containerPortBind := fmt.Sprintf("%s/tcp", app.ContainerPort)

	containerConfig := &container.Config{
		Image:  imageTag,
		Labels: labels,
		ExposedPorts: map[nat.Port]struct{}{
			nat.Port(containerPortBind): {},
		},
	}

	hostConfig := &container.HostConfig{
		// No PortBindings needed - Traefik handles routing
		NetworkMode: "web", // Explicitly attach to the traefik network
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
		AutoRemove: false, // Don't auto-remove for PR deployments
	}

	// Remove existing container if it exists
	d.RemoveContainerIfExists(ctx, app)

	// Create container
	resp, err := d.Client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	err = d.Client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		// Cleanup on failure
		d.Client.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Printf("✅ Container %s started with ID %s\n", containerName, resp.ID)
	return resp.ID, nil
}

// RunContainer runs a container (legacy method for backward compatibility)
func (d *DockerRunner) RunContainer(ctx context.Context, app *sharedTypes.App, imageTag, containerName string) (string, error) {
	containerPortBind := fmt.Sprintf("%s/tcp", app.ContainerPort)

	containerConfig := &container.Config{
		Image: imageTag,
		ExposedPorts: map[nat.Port]struct{}{
			nat.Port(containerPortBind): {},
		},
	}

	hostConfig := &container.HostConfig{
		// No PortBindings needed - use internal container port
		AutoRemove: true,
	}

	d.RemoveContainerIfExists(ctx, app)

	resp, err := d.Client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)

	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	err = d.Client.ContainerStart(ctx, resp.ID, container.StartOptions{})

	if err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	return resp.ID, nil
}

// StopAndRemoveContainer stops and removes a container by ID
func (d *DockerRunner) StopAndRemoveContainer(ctx context.Context, containerID string) error {
	// Stop the container
	if err := d.Client.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		fmt.Printf("⚠️  Warning: failed to stop container %s: %v\n", containerID, err)
	}

	// Remove the container
	if err := d.Client.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	fmt.Printf("✅ Container stopped and removed: %s\n", containerID)
	return nil
}

// RemoveContainerIfExists removes a container if it exists
func (d *DockerRunner) RemoveContainerIfExists(ctx context.Context, app *sharedTypes.App) error {
	containers, err := d.Client.ContainerList(ctx, container.ListOptions{All: true})

	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if name == "/"+app.Name {
				fmt.Printf("🧹 Removing existing container: %s\n", name)

				// Stop the container
				if err := d.Client.ContainerStop(ctx, c.ID, container.StopOptions{}); err != nil {
					fmt.Printf("⚠️  Warning: failed to stop container %s: %v\n", c.ID, err)
				}

				// Remove the container
				if err := d.Client.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true}); err != nil {
					return fmt.Errorf("failed to remove container: %w", err)
				}

				// Wait for container to be fully removed
				for i := 0; i < 10; i++ {
					time.Sleep(1 * time.Second)

					// Check if container still exists
					containers, _ := d.Client.ContainerList(ctx, container.ListOptions{All: true})
					exists := false
					for _, c2 := range containers {
						for _, name2 := range c2.Names {
							if name2 == "/"+app.Name {
								exists = true
								break
							}
						}
					}

					if !exists {
						fmt.Printf("✅ Container removed: %s\n", name)
						return nil
					}
				}

				return fmt.Errorf("timeout waiting for container removal: %s", name)
			}
		}
	}

	return nil
}

// ListContainers lists all containers with optional filtering
func (d *DockerRunner) ListContainers(ctx context.Context, all bool) ([]types.Container, error) {
	return d.Client.ContainerList(ctx, container.ListOptions{All: all})
}

// GetContainerInfo gets detailed information about a container
func (d *DockerRunner) GetContainerInfo(ctx context.Context, containerID string) (container.InspectResponse, error) {
	return d.Client.ContainerInspect(ctx, containerID)
}
