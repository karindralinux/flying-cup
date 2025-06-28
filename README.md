# Flying Cup - Preview Deployment PaaS (On Progress)

An open-source, self-hostable service that automatically creates deployment previews for Pull Requests on GitHub.

## Features

- 🚀 **Automatic PR Deployment** - Deploy previews when PRs are opened
- 🧹 **Automatic Cleanup** - Remove previews when PRs are closed
- �� **Docker-based** - Uses Docker for consistent deployments

## Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ (for development)
- GitHub repository with webhook access

## Quick Start

1. **Clone the repository:**
   ```bash
   git clone https://github.com/your-username/flying-cup.git
   cd flying-cup
   ```

2. **Configure the application:**
   ```bash
   cp config.yaml.example config.yaml
   # Edit config.yaml with your GitHub webhook secret
   ```

3. **Run the application:**
   ```bash
   go run main.go config.go
   ```

4. **Set up GitHub webhook:**
   - Go to your repository settings
   - Add webhook URL: `http://your-server:8080/webhook/github`
   - Set content type to `application/x-www-form-urlencoded`
   - Add your webhook secret

## Configuration

### config.yaml

```yaml
github:
  app_id: "your-github-app-id"
  webhook_secret: "your-github-webhook-secret"
```

## How It Works

1. **PR Opened** → GitHub sends webhook → Flying Cup clones repository → Builds Docker image → Deploys container → Preview available
2. **PR Closed** → GitHub sends webhook → Flying Cup stops container → Removes container and image → Cleans up repository

## Development

```bash
# Run locally
go run main.go config.go

# Build binary
go build -o flying-cup
```

## License

This project is licensed under Creative Commons Attribution-NonCommercial 4.0 International License.

- ✅ **Personal use**: Free for personal projects
- ✅ **Educational use**: Free for schools, universities
- ✅ **Non-profit use**: Free for charities, open source
- ❌ **Commercial use**: Cannot sell or use in for-profit companies

For more information, see the [LICENSE](LICENSE) file.