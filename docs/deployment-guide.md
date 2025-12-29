# Deployment Guide

**Version**: 1.0.0 | **Last Updated**: 2025-12-29

---

## Overview

This guide covers deployment procedures for Zuno NFT Marketplace API with Sentry integration for release tracking and error monitoring.

---

## Prerequisites

### GitHub Secrets Configuration

Before deploying, configure the following secrets in GitHub repository settings (`Settings → Secrets and variables → Actions`):

| Secret Name | Description | How to Get |
|------------|-------------|------------|
| `SENTRY_AUTH_TOKEN` | Authentication token for Sentry CLI | https://sentry.io/settings/auth-tokens/ (required scopes: `project:releases`, `project:write`) |
| `SENTRY_DSN` | Sentry Data Source Name | From Sentry project settings |
| `SENTRY_ORG` | Sentry organization slug | Your Sentry org (e.g., `zuno`) |
| `SENTRY_PROJECT` | Sentry project slug | Your Sentry project (e.g., `zuno-marketplace-api`) |

### Environment Variables

Copy `.env.example` to `.env` and configure:

```env
# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DATABASE=nft_marketplace

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest

# Sentry
SENTRY_DSN=https://your-dsn@sentry.io/project-id
SENTRY_ENVIRONMENT=development
SENTRY_TRACES_SAMPLE_RATE=0.1
```

---

## CI/CD Pipeline

### GitHub Actions Workflows

The repository uses GitHub Actions for CI/CD:

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | Push to `main`, `develop-claude`, PR | Lint, test, build |
| `deploy-staging.yml` | Push to `main`, `develop`, `feature/add-sentry` | Deploy to staging with Sentry tracking |
| `pr.yml` | Pull request | PR quality checks |

---

## Deployment with Sentry

### Staging Deployment

Automatic deployment triggered on push to `main`, `develop`, or `feature/add-sentry` branches:

1. **Checkout code** with full git history
2. **Run tests** to verify build
3. **Build services** using Makefile
4. **Create Sentry Release** with associated commits
5. **Deploy to staging** environment
6. **Notify Sentry** of deploy completion

```bash
# Manual trigger (requires GitHub Actions workflow_dispatch)
git push origin feature/add-sentry
```

### Manual Deploy with Sentry

For manual deployments (e.g., production):

```bash
# Install Sentry CLI
npm install -g @sentry/cli

# Get version
VERSION=$(git describe --tags --always --dirty)

# Create release with commits
sentry-cli releases new "$VERSION" \
  --org $SENTRY_ORG \
  --project $SENTRY_PROJECT
sentry-cli releases set-commits "$VERSION" --auto
sentry-cli releases finalize "$VERSION"

# Deploy (your deployment command)
kubectl apply -f k8s/

# Notify Sentry of deploy
sentry-cli releases deploys "$VERSION" new \
  --env production \
  --name "Production Deploy"
```

---

## Viewing Deploy-Specific Errors

### In Sentry UI

1. Navigate to **Sentry → Releases**
2. Select the release version (e.g., `v1.0.0`, `feature-add-sentry-123abcd`)
3. View **Issues** introduced since this release
4. Filter by **Environment** (staging/production)

### Release Health Metrics

Each release shows:
- **New Issues**: Errors introduced after this release
- **Unresolved Issues**: Total open issues for this release
- **Sessions**: User sessions affected
- **Crash Free Rate**: Percentage of error-free sessions

---

## Rollback Procedures

### Identify Bad Release

1. Go to Sentry → Releases
2. Sort by **Issues** or **Crash Rate**
3. Identify release with spike in errors

### Rollback Deployment

```bash
# Rollback to previous version
kubectl rollout undo deployment/auth-service
kubectl rollout undo deployment/user-service
kubectl rollout undo deployment/wallet-service
kubectl rollout undo deployment/graphql-gateway

# Or rollback to specific revision
kubectl rollout undo deployment/auth-service --to-revision=3
```

### Notify Sentry of Rollback

```bash
# Mark previous release as deployed
PREV_VERSION=$(git describe --tags --always~1)
sentry-cli releases deploys "$PREV_VERSION" new \
  --env production \
  --name "Rollback Deploy" \
  --deployed-at "$(date +%s)"
```

---

## Build Version Injection

Services expose version information via build flags:

```bash
# Check version
./build/auth-service --version
# Output: Version: v1.0.0-dirty BuildTime: 2025-12-29T12:00:00Z

# Service logs include version
[INFO] Starting auth-service Version: v1.0.0-dirty BuildTime: 2025-12-29T12:00:00Z
```

### Version Format

- `v1.0.0` - Tagged release
- `v1.0.0-5-g123abcd` - 5 commits after tag
- `feature-add-sentry-123abcd` - Branch-based version
- `-dirty` - Uncommitted changes

---

## Troubleshooting

### Workflow Fails at Sentry Release Step

**Cause**: Missing or invalid `SENTRY_AUTH_TOKEN`

**Solution**:
1. Verify token exists in GitHub Secrets
2. Ensure token has `project:releases` and `project:write` scopes
3. Regenerate token if expired

### Commits Not Associated with Release

**Cause**: Git checkout without full history

**Solution**: Ensure workflow uses `fetch-depth: 0`

```yaml
- name: Checkout code
  uses: actions/checkout@v4
  with:
    fetch-depth: 0  # Full history required
```

### Wrong Release Version in Sentry

**Cause**: Inconsistent version naming

**Solution**: Use `git describe --tags --always --dirty` consistently across all steps

---

## Monitoring & Alerts

### Sentry Alerts

Configure alerts in Sentry:
1. Go to **Settings → Alerts**
2. Create alert rules for:
   - New error spikes
   - High error rate (>5%)
   - Release regressions

### GitHub Actions Notifications

Enable notifications in:
- **Repository → Settings → Notifications**
- Configure email/Slack for workflow failures

---

## Next Steps

- [ ] Configure production deployment workflow
- [ ] Set up Sentry alerting rules
- [ ] Configure Slack notifications for deployments
- [ ] Add smoke tests to post-deploy verification

---

**Related Docs**:
- [System Architecture](system-architecture.md)
- [Code Standards](code-standards.md)
- [Project Roadmap](project-roadmap.md)
