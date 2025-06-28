package main

import (
	"context"
	"log"
	"os"
	"os/exec"

	"github.com/docker/docker/client"
	"github.com/karindrlainux/flying-cup/pkg/deployment"
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

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))

	e.POST("/webhook/github", webhook.HandleGithubWebhook(
		config.Github.WebhookSecret, deployment.DeployPR, deployment.Cleanup,
	))

	e.Logger.Fatal(e.Start(":8080"))
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
