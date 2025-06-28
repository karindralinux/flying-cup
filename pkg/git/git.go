package git

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func CloneRepository(ctx context.Context, cloneUrl string, branch string, targetPath string) error {

	if _, err := os.Stat(targetPath); err == nil {
		if err := os.RemoveAll(targetPath); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	log.Printf("Cloning repository %s (branch: %s) to %s", cloneUrl, branch, targetPath)

	cmd := exec.CommandContext(ctx, "git", "clone", "-b", branch, cloneUrl, targetPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("git clone failed with exit code %d: %s", exitErr.ExitCode(), exitErr.String())
		}
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	log.Printf("Repository %s (branch: %s) successfully cloned", cloneUrl, branch)
	return nil
}

func RemoveClonedRepository(ctx context.Context, targetPath string) error {
	if err := os.RemoveAll(targetPath); err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}

	return nil
}
