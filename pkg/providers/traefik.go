package providers

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/karindrlainux/flying-cup/pkg/docker"
	"github.com/karindrlainux/flying-cup/pkg/git"
	"github.com/karindrlainux/flying-cup/pkg/types"
	"github.com/karindrlainux/flying-cup/pkg/webhook"
)

// TraefikProvider implements Provider for Traefik
type TraefikProvider struct {
	config *Config
	// Traefik-specific fields
	deployments map[string]*TraefikDeployment
	mu          sync.Mutex
}

// TraefikDeployment represents a deployment managed by Traefik
type TraefikDeployment struct {
	ID          string
	Name        string
	Domain      string
	ContainerID string
	Status      string
}

// NewTraefikProvider creates a new Traefik provider
func NewTraefikProvider(config *Config) *TraefikProvider {
	return &TraefikProvider{
		config:      config,
		deployments: make(map[string]*TraefikDeployment),
	}
}

// Init initializes the Traefik provider
func (t *TraefikProvider) Init(config interface{}) error {
	// Ensure web network exists
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}

	return t.ensureWebNetwork(context.Background(), cli)
}

// CreateDeployment creates a new deployment using Traefik
func (t *TraefikProvider) CreateDeployment(ctx context.Context, webhook *webhook.GithubPRWebhook) (string, error) {
	log.Printf("Creating Traefik deployment for PR #%d", webhook.Number)

	// Generate domain for this PR
	previewDomain, err := t.createTraefikDeployment(webhook)
	if err != nil {
		return "", fmt.Errorf("failed to create Traefik deployment: %w", err)
	}

	// Build and run container
	containerID, err := t.buildAndRunContainer(ctx, webhook, previewDomain)
	if err != nil {
		return "", fmt.Errorf("failed to build and run container: %w", err)
	}

	// Update deployment with container ID
	t.mu.Lock()
	deploymentKey := fmt.Sprintf("%s-pr-%s-%d", webhook.Repository.Name, webhook.PullRequest.Title, webhook.Number)
	if deployment, exists := t.deployments[deploymentKey]; exists {
		deployment.ContainerID = containerID
		deployment.Status = "running"
	}
	t.mu.Unlock()

	// Get protocol based on environment
	protocol := "https" // Default to HTTPS for production
	if t.config.Environment == "local" {
		protocol = "http"
	}

	return fmt.Sprintf("%s://%s", protocol, previewDomain), nil
}

// CleanupDeployment removes a Traefik deployment
func (t *TraefikProvider) CleanupDeployment(ctx context.Context, repoName, prName string, prNumber int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	deploymentKey := fmt.Sprintf("%s-pr-%s-%d", repoName, prName, prNumber)
	deployment, exists := t.deployments[deploymentKey]
	if !exists {
		return fmt.Errorf("deployment not found: %s", deploymentKey)
	}

	log.Printf("Cleaning up Traefik deployment: %s", deploymentKey)

	// Stop and remove container if it exists
	if deployment.ContainerID != "" {
		if err := t.stopAndRemoveContainer(deployment.ContainerID); err != nil {
			log.Printf("Warning: failed to stop container: %v", err)
		}
	}

	delete(t.deployments, deploymentKey)
	log.Printf("Traefik deployment removed: %s", deploymentKey)
	return nil
}

// GetDeploymentStatus returns the status of a deployment
func (t *TraefikProvider) GetDeploymentStatus(ctx context.Context, deploymentID string) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	deployment, exists := t.deployments[deploymentID]
	if !exists {
		return "", fmt.Errorf("deployment not found: %s", deploymentID)
	}

	return deployment.Status, nil
}

// Helper methods

func (t *TraefikProvider) createTraefikDeployment(webhook *webhook.GithubPRWebhook) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	deploymentKey := fmt.Sprintf("%s-pr-%s-%d", webhook.Repository.Name, webhook.PullRequest.Title, webhook.Number)

	// Generate subdomain for this PR
	// Format: reponame-prname-prnumber.domain
	cleanRepoName := t.sanitizeForDomain(webhook.Repository.Name)
	cleanPRName := t.sanitizeForDomain(webhook.PullRequest.Title)
	subdomain := fmt.Sprintf("%s-%s-%d", cleanRepoName, cleanPRName, webhook.Number)
	domain := fmt.Sprintf("%s.%s", subdomain, t.config.Domain)

	log.Printf("Creating Traefik deployment: %s", deploymentKey)
	log.Printf("Preview URL will be: https://%s", domain)

	// Store deployment info
	deployment := &TraefikDeployment{
		ID:     deploymentKey,
		Name:   deploymentKey,
		Domain: domain,
		Status: "pending",
	}

	t.deployments[deploymentKey] = deployment

	log.Printf("Traefik deployment configured: %s", domain)
	return domain, nil
}

func (t *TraefikProvider) buildAndRunContainer(ctx context.Context, webhook *webhook.GithubPRWebhook, domain string) (string, error) {
	// Generate code name
	codeName := fmt.Sprintf("pr-%s-%d", webhook.Repository.Name, webhook.Number)
	repoPath := fmt.Sprintf("./repos/%s", codeName)

	// Clone repository
	err := git.CloneRepository(ctx, webhook.Repository.CloneUrl, webhook.PullRequest.Head.Ref, repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	// Create Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("failed to create docker client: %w", err)
	}

	// Build Docker image
	dockerBuilder := &docker.DockerBuilder{Client: cli}
	app := &types.App{
		Name:          codeName,
		SourcePath:    repoPath,
		ContainerPort: "8080",
	}

	imageTag, err := dockerBuilder.BuildImage(ctx, app, "Dockerfile", true)
	if err != nil {
		return "", fmt.Errorf("failed to build Docker image: %w", err)
	}

	// Generate Traefik labels
	labels := t.generateTraefikLabels(webhook, domain)

	// Run container with Traefik integration
	dockerRunner := &docker.DockerRunner{Client: cli}
	containerID, err := dockerRunner.RunContainerWithTraefik(ctx, app, imageTag, codeName, labels)
	if err != nil {
		return "", fmt.Errorf("failed to run Docker container: %w", err)
	}

	log.Printf("Container %s started with ID %s", codeName, containerID)
	return containerID, nil
}

func (t *TraefikProvider) generateTraefikLabels(webhook *webhook.GithubPRWebhook, domain string) map[string]string {
	deploymentKey := fmt.Sprintf("%s-pr-%s-%d", webhook.Repository.Name, webhook.PullRequest.Title, webhook.Number)

	return map[string]string{
		// Enable Traefik for this container
		"traefik.enable": "true",

		// HTTP Router configuration - simplified for local testing
		"traefik.http.routers." + deploymentKey + ".rule":        fmt.Sprintf("Host(`%s`)", domain),
		"traefik.http.routers." + deploymentKey + ".entrypoints": "web",

		// Service configuration - use internal container port
		"traefik.http.services." + deploymentKey + ".loadbalancer.server.port": "8080",

		// Add metadata labels
		"flying-cup.deployment": deploymentKey,
		"flying-cup.repo":       webhook.Repository.Name,
		"flying-cup.pr":         fmt.Sprintf("%d", webhook.Number),
		"flying-cup.domain":     domain,
	}
}

func (t *TraefikProvider) ensureWebNetwork(ctx context.Context, cli *client.Client) error {
	networks, err := cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	// Check if web network already exists
	for _, net := range networks {
		if net.Name == "web" {
			log.Printf("Web network already exists")
			return nil
		}
	}

	// Create web network if it doesn't exist
	log.Printf("Creating web network for Traefik connectivity")
	_, err = cli.NetworkCreate(ctx, "web", network.CreateOptions{
		Driver: "bridge",
		Labels: map[string]string{
			"com.flying-cup.managed": "true",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create web network: %w", err)
	}

	log.Printf("Web network created successfully")
	return nil
}

func (t *TraefikProvider) stopAndRemoveContainer(containerID string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}

	ctx := context.Background()

	// Stop the container
	if err := cli.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		log.Printf("Warning: failed to stop container %s: %v", containerID, err)
	}

	// Remove the container
	if err := cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	log.Printf("Container stopped and removed: %s", containerID)
	return nil
}

func (t *TraefikProvider) sanitizeForDomain(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	re := regexp.MustCompile(`[^a-z0-9-]`)
	s = re.ReplaceAllString(s, "")
	return s
}
