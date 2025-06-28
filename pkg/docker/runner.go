package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"

	"github.com/docker/docker/client"
	sharedTypes "github.com/karindrlainux/flying-cup/pkg/types"
)

type DockerRunner struct {
	Client *client.Client
}

func (d *DockerRunner) RunContainer(ctx context.Context, app *sharedTypes.App, imageTag, containerName string) (string, error) {

	containerPortBind := fmt.Sprintf("%s/tcp", app.ContainerPort)

	containerConfig := &container.Config{
		Image: imageTag,
		ExposedPorts: map[nat.Port]struct{}{
			nat.Port(containerPortBind): {},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(containerPortBind): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: app.HostPort,
				},
			},
		},
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

func (d *DockerRunner) RemoveContainerIfExists(ctx context.Context, app *sharedTypes.App) error {

	containers, err := d.Client.ContainerList(ctx, container.ListOptions{})

	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	fmt.Println("Containers:")
	for _, c := range containers {
		fmt.Println(c.Names[0])
		if c.Names[0] == "/"+app.Name {
			err = d.Client.ContainerStop(ctx, c.ID, container.StopOptions{})

			if err != nil {
				return fmt.Errorf("failed to stop and remove container: %w", err)
			}

			// Simple wait - check every second for 10 seconds
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
					fmt.Printf("✅ Container removed: %s\n", "/preview-"+app.Name)
					return nil
				}
			}
		}
	}

	return nil
}

// GetContainerHostPort retrieves the host port from running container
func (d *DockerRunner) GetContainerHostPort(ctx context.Context, containerName string) (string, error) {
	containers, err := d.Client.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+containerName {
				// Get container details to see port bindings
				containerInfo, err := d.Client.ContainerInspect(ctx, container.ID)
				if err != nil {
					return "", fmt.Errorf("failed to inspect container: %w", err)
				}

				// Check port bindings
				for containerPort, bindings := range containerInfo.NetworkSettings.Ports {
					if len(bindings) > 0 {
						fmt.Printf("Found port binding: %s -> %s\n", containerPort, bindings[0].HostPort)
						return bindings[0].HostPort, nil
					}
				}

				return "", fmt.Errorf("no port bindings found for container")
			}
		}
	}

	return "", fmt.Errorf("container not found: %s", containerName)
}
