package main

import (
	"github.com/karindrlainux/flying-cup/pkg/deployment"
	"github.com/karindrlainux/flying-cup/pkg/webhook"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"os/exec"
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
