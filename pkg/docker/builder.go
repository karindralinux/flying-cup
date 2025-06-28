package docker

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	sharedTypes "github.com/karindrlainux/flying-cup/pkg/types"
)

type DockerBuilder struct {
	Client *client.Client
}

// createTarArchive creates a tar.gz archive from the given directory
func (d *DockerBuilder) createTarArchive(sourceDir string) (io.ReadCloser, error) {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		gw := gzip.NewWriter(pw)
		defer gw.Close()

		tw := tar.NewWriter(gw)
		defer tw.Close()

		err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip .git directory and other unnecessary files
			if info.IsDir() && (info.Name() == ".git" || info.Name() == "node_modules") {
				return filepath.SkipDir
			}

			// Calculate relative path for tar header
			relPath, err := filepath.Rel(sourceDir, path)
			if err != nil {
				return err
			}

			// Skip root directory
			if relPath == "." {
				return nil
			}

			header, err := tar.FileInfoHeader(info, relPath)
			if err != nil {
				return err
			}

			header.Name = relPath

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			// Write file content if it's a regular file
			if !info.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()

				_, err = io.Copy(tw, file)
				return err
			}

			return nil
		})

		if err != nil {
			pw.CloseWithError(err)
		}
	}()

	return pr, nil
}

func (d *DockerBuilder) BuildImage(ctx context.Context, app *sharedTypes.App, dockerfile string, nonCache bool) (string, error) {

	imageTag := fmt.Sprintf("%s:latest", app.Name)

	fmt.Printf("Building image %s with dockerfile %s\n", imageTag, dockerfile)

	// Build options
	buildOptions := types.ImageBuildOptions{
		Dockerfile: dockerfile,
		Tags:       []string{imageTag},
		NoCache:    nonCache,
		Remove:     true,
	}

	// Create tar archive from the cloned repository
	buildContext, err := d.createTarArchive(app.SourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to create build context: %w", err)
	}
	defer buildContext.Close()

	// Build the image using Docker API
	buildResponse, err := d.Client.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return "", fmt.Errorf("failed to build docker image: %w", err)
	}
	defer buildResponse.Body.Close()

	// Stream the build output
	_, err = io.Copy(os.Stdout, buildResponse.Body)
	if err != nil {
		return "", fmt.Errorf("failed to stream build output: %w", err)
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
