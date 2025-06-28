package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	sharedTypes "github.com/karindrlainux/flying-cup/pkg/types"
)

type DockerBuilder struct {
	Client *client.Client
}

func (d *DockerBuilder) BuildImage(ctx context.Context, app *sharedTypes.App, dockerfile string, nonCache bool) (string, error) {

	imageTag := fmt.Sprintf("%s:latest", app.Name)

	args := []string{"build", "-t", imageTag}

	if nonCache {
		args = append(args, "--no-cache")
	}

	fmt.Printf("Building image %s with dockerfile %s\n", imageTag, dockerfile)

	fmt.Printf("Check %s != %s\n", dockerfile, "Dockerfile")
	if dockerfile != "Dockerfile" {
		args = append(args, "-f", dockerfile)
	}

	args = append(args, app.SourcePath)

	fmt.Println(args)

	cmd := exec.CommandContext(ctx, "docker", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to build docker image: %w", err)
	}

	return imageTag, nil
}

func (d *DockerBuilder) RemoveImage(ctx context.Context, imageTag string) error {

	_, err := d.Client.ImageRemove(ctx, imageTag, image.RemoveOptions{Force: true})

	if err != nil {
		return fmt.Errorf("failed to remove Docker image: %w", err)
	}

	return nil
}
