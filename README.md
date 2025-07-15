# Flying Cup

A preview deployment PaaS system for GitHub Pull Request previews using Docker Container.


## Requirements

Flying Cup is designed to run entirely on Docker. Please ensure the following before you begin:

- **Docker Engine**: Installed and running on your server or development machine. [Get Docker](https://docs.docker.com/get-docker/)
- **docker-compose**: Installed (if not included with your Docker installation). [Get docker-compose](https://docs.docker.com/compose/install/)
- **Minimum System Resources**:
  - **Memory**: At least 2 GB RAM (4 GB recommended for multiple concurrent PRs)
  - **Storage**: At least 2 GB free disk space (more required for large repositories or many deployments)
- **Network**: Publicly accessible IP address and open ports (default: 80 for web, 9000 for Traefik dashboard)
- **Repository Requirements**:
  - The application repository you want to deploy **must include a valid `Dockerfile`** at its root or project directory. Flying Cup will use this Dockerfile to build and run preview containers.
  - See the [`example/`](./example) directory in this repository for sample apps (Node.js and Go) that meet this requirement.

> ⚠️ **Note:** Flying Cup will not work with repositories that do not have a Dockerfile. Please ensure your app is containerized before using this system.


## Features

- Automatic GitHub webhook handling for PR events
- Traefik integration for secure preview URLs with automatic SSL
- Docker-based deployment with Traefik routing
- Automatic cleanup of preview deployments
- Environment-based configuration (HTTP for local, HTTPS for production)

## Quick Start

### 1. Setup

```bash
# Clone the repository
git clone <repository-url>
cd flying-cup

# Run setup script
chmod +x setup.sh
./setup.sh

# Edit configuration
nano .env
```

### 2. Configuration

Edit the `.env` file with your settings:

```bash
# Environment (local/staging/production)
ENVIRONMENT=local

# Domain for PR previews
DOMAIN=preview.ngodingo.web.id

# GitHub Configuration
GITHUB_APP_ID=your-github-app-id
GITHUB_WEBHOOK_SECRET=your-webhook-secret
GITHUB_TOKEN=your-github-token

# Ports Configuration
PORT=80
DASHBOARD_PORT=9000
```

### 3. Start Services

```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f
```

### 4. Access Points

- **Traefik Dashboard**: `http://localhost:9000`
- **Flying Cup Controller**: `http://localhost`
- **PR Previews**: `https://{repo}-{pr}-{number}.{your-domain}`

## Architecture

### Services

1. **Controller**: Flying Cup main application
   - Handles GitHub webhooks
   - Manages PR deployments
   - No external port exposure (routed through Traefik)

2. **Traefik**: Reverse proxy and load balancer
   - Handles all external traffic
   - Automatic SSL certificates
   - Routes to appropriate containers

### Network Flow

```
Internet → Traefik (Port 80/443) → Controller (Internal)
         → Traefik (Port 80/443) → PR Containers (Internal)
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | `local` | Environment (local/staging/production) |
| `DOMAIN` | `preview.ngodingo.web.id` | Domain for PR previews |
| `GITHUB_APP_ID` | - | GitHub App ID |
| `GITHUB_WEBHOOK_SECRET` | - | GitHub webhook secret |
| `GITHUB_TOKEN` | - | GitHub personal access token |
| `PORT` | `80` | Port for web traffic |
| `DASHBOARD_PORT` | `9000` | Port for Traefik dashboard |

### Example .env file

```bash
# Production setup
ENVIRONMENT=production
DOMAIN=preview.mycompany.com
PORT=80
DASHBOARD_PORT=9000

# Development setup
ENVIRONMENT=local
DOMAIN=localhost
PORT=8080
DASHBOARD_PORT=9000
```

## DNS Setup

For production, configure your DNS with wildcard records:

```
*.preview.ngodingo.web.id  A  YOUR_SERVER_IP
preview.ngodingo.web.id    A  YOUR_SERVER_IP
```

## Development

```bash
# Run locally with environment variables
export $(cat .env | xargs)
go run .

# Or run with docker-compose
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs controller
docker-compose logs traefik
```

## Preview URL Format

PR preview URLs follow this format:
```
{protocol}://{repo-name}-{pr-title}-{pr-number}.{domain}
```

### Examples:

**Local Development:**
```
http://myapp-feature-login-123.localhost
```

**Production:**
```
https://myapp-feature-login-123.preview.ngodingo.web.id
```

## GitHub Webhook Setup

1. **Create GitHub App** or use Personal Access Token
2. **Configure Webhook** in your repository:
   - URL: `https://your-domain/webhook/github`
   - Content type: `application/x-www-form-urlencoded`
   - Events: `Pull requests`
   - Secret: Use the same value as `GITHUB_WEBHOOK_SECRET`

3. **Set Webhook Secret** in your `.env` file

## Troubleshooting

### Check Configuration

```bash
# Verify environment variables are loaded
docker-compose exec controller env | grep -E "(ENVIRONMENT|DOMAIN|GITHUB_)"

# Check logs
docker-compose logs controller
docker-compose logs traefik
```

### Port Conflicts

If you encounter port conflicts, modify the `.env` file:

```bash
# Use different ports
PORT=8080
DASHBOARD_PORT=9001
```

### Check Traefik Dashboard

Access the Traefik dashboard to see routing configuration:
- URL: `http://localhost:9000` (or your configured dashboard port)
- Shows all active routes and services

### Verify DNS

Test your DNS configuration:
```bash
# Test wildcard DNS
nslookup test.preview.ngodingo.web.id

# Test main domain
nslookup preview.ngodingo.web.id
```

### Common Issues

1. **Webhook 401 errors**: Check `GITHUB_WEBHOOK_SECRET` matches GitHub webhook configuration
2. **Container not found**: Ensure Docker is running and web network exists
3. **Preview not accessible**: Check Traefik dashboard for routing issues
4. **SSL errors**: Verify DNS points to your server and Let's Encrypt can reach it

## Environment-Specific Configuration

### Local Development
```bash
ENVIRONMENT=local
DOMAIN=localhost
PORT=8080
# Results in: http://repo-pr-title-123.localhost
```

### Staging
```bash
ENVIRONMENT=staging
DOMAIN=staging.preview.ngodingo.web.id
PORT=80
# Results in: https://repo-pr-title-123.staging.preview.ngodingo.web.id
```

### Production
```bash
ENVIRONMENT=production
DOMAIN=preview.ngodingo.web.id
PORT=80
# Results in: https://repo-pr-title-123.preview.ngodingo.web.id
```

## Security Considerations

- **Webhook Secret**: Always use a strong, unique webhook secret
- **GitHub Token**: Use GitHub App tokens instead of personal access tokens when possible
- **Environment Variables**: Never commit `.env` files to version control
- **Network Security**: Ensure Traefik dashboard is not publicly accessible in production

## License

This project is licensed under Creative Commons Attribution-NonCommercial 4.0 International License.

- ✅ **Personal use**: Free for personal projects
- ✅ **Educational use**: Free for schools, universities
- ✅ **Non-profit use**: Free for charities, open source
- ❌ **Commercial use**: Cannot sell or use in for-profit companies

For more information, see the [LICENSE](LICENSE) file.