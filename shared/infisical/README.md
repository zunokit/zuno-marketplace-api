# Infisical Secrets Management Integration

This package provides integration with [Infisical](https://infisical.com) for cloud-based secrets management, replacing local `.env` files with a secure, centralized secrets store.

## Overview

Instead of managing secrets in local environment files, this integration allows you to:

- Store all secrets securely in Infisical cloud
- Access secrets with automatic caching for performance
- Fall back to environment variables if Infisical is unavailable
- Support multiple environments (dev, staging, prod)
- Automatic token refresh and authentication

## Quick Start

### 1. Set Up Infisical

1. Create an account at [infisical.com](https://infisical.com)
2. Create a project and add your secrets
3. Create a Machine Identity with Universal Auth
4. Get your Client ID and Client Secret

### 2. Configure Environment Variables

Add to your `.env.development`:

```bash
# Enable Infisical
INFISICAL_ENABLED=true
INFISICAL_PROJECT_ID=your-project-id
INFISICAL_ENVIRONMENT=dev
INFISICAL_SECRET_PATH=/
INFISICAL_UNIVERSAL_AUTH_CLIENT_ID=your-client-id
INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET=your-client-secret

# Optional
INFISICAL_SITE_URL=https://app.infisical.com
INFISICAL_CACHE_TTL=300
```

### 3. Initialize in Your Service

Update your service's `main.go`:

```go
package main

import (
    "context"
    "log"
    
    "github.com/zunokit/zuno-marketplace-api/shared/env"
)

func main() {
    // Initialize Infisical (optional, falls back to env vars if not configured)
    ctx := context.Background()
    if err := env.InitInfisical(ctx); err != nil {
        log.Printf("Infisical not initialized (using env vars): %v", err)
    } else {
        log.Println("✅ Infisical secrets management enabled")
    }
    
    // Continue with normal service initialization
    cfg := config.Load()
    // ...
}
```

### 4. Use Secrets

Secrets are automatically retrieved with fallback:

```go
// This will check: 1) Env var, 2) Infisical, 3) Fallback value
dbURL := env.GetString("DATABASE_URL", "postgresql://localhost:5432/mydb")

// Or prioritize Infisical over env vars
jwtSecret := env.GetStringFromInfisical("JWT_SECRET", "default-secret")
```

## Configuration Options

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `INFISICAL_ENABLED` | No | `false` | Enable Infisical integration |
| `INFISICAL_PROJECT_ID` | Yes* | - | Your Infisical project ID |
| `INFISICAL_ENVIRONMENT` | Yes* | - | Environment slug (dev, staging, prod) |
| `INFISICAL_SECRET_PATH` | No | `/` | Path to secrets in Infisical |
| `INFISICAL_UNIVERSAL_AUTH_CLIENT_ID` | Yes* | - | Machine identity client ID |
| `INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET` | Yes* | - | Machine identity client secret |
| `INFISICAL_SITE_URL` | No | `https://app.infisical.com` | Infisical instance URL |
| `INFISICAL_CACHE_TTL` | No | `300` | Cache expiry in seconds |

*Required only if `INFISICAL_ENABLED=true`

## API Reference

### InitInfisical

```go
func InitInfisical(ctx context.Context) error
```

Initializes the global Infisical client. Safe to call multiple times (idempotent). Returns an error if Infisical is not configured or authentication fails.

### GetString

```go
func GetString(key, fallback string) string
```

Retrieves a string value with the following priority:
1. Environment variable
2. Infisical secret (if enabled and initialized)
3. Fallback value

### GetStringFromInfisical

```go
func GetStringFromInfisical(key, fallback string) string
```

Retrieves a string value with Infisical prioritized:
1. Infisical secret (if enabled and initialized)
2. Environment variable
3. Fallback value

### GetInt / GetBool

Similar to `GetString` but with type conversion.

### InvalidateInfisicalCache

```go
func InvalidateInfisicalCache()
```

Clears the secret cache, forcing fresh fetches on next access.

## Authentication Methods

The integration supports multiple authentication methods:

### Universal Auth (Recommended)

Uses Client ID and Client Secret from a Machine Identity.

```bash
INFISICAL_UNIVERSAL_AUTH_CLIENT_ID=xxx
INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET=yyy
```

### Other Methods

The Infisical SDK also supports:
- GCP ID Token Auth
- AWS IAM Auth
- Azure Auth
- Kubernetes Auth
- JWT Auth
- LDAP Auth
- OCI Auth

See the [Infisical Go SDK documentation](https://infisical.com/docs/sdks/languages/go) for details.

## Caching

Secrets are cached in memory to reduce API calls:

- **Default TTL**: 300 seconds (5 minutes)
- **Cache invalidation**: Call `env.InvalidateInfisicalCache()`
- **Per-secret caching**: Each secret is cached independently

## Error Handling

If Infisical fails or is not configured:

- `InitInfisical` returns an error (non-fatal)
- `GetString` falls back to environment variables
- `GetString` falls back to default value if env var not set

This ensures your application continues to work even if Infisical is unavailable.

## Migration from Local .env Files

To migrate from local `.env` files to Infisical:

1. **Export existing secrets** from your `.env` files
2. **Import to Infisical** via the web UI or CLI
3. **Set environment variables** for Infisical connection
4. **Keep .env files** as fallback during migration
5. **Remove .env files** once fully migrated

Example migration script:

```bash
#!/bin/bash
# export_secrets.sh

# Read .env file and format for Infisical import
while IFS='=' read -r key value; do
    [[ -z "$key" || "$key" =~ ^# ]] && continue
    echo "infisical secrets set $key=$value --env=dev"
done < .env.development
```

## Best Practices

1. **Always provide fallbacks** for critical secrets
2. **Use different projects** for different environments
3. **Rotate credentials** regularly using Machine Identities
4. **Monitor cache hit rates** for performance tuning
5. **Keep INFISICAL_ENABLED=false** in development if not using Infisical
6. **Use secret paths** to organize secrets by service

## Security Considerations

- Secrets are encrypted in transit (HTTPS/TLS)
- Secrets are cached only in memory (not disk)
- Authentication tokens auto-refresh
- Supports audit logging in Infisical
- Use Machine Identities with minimal permissions

## Troubleshooting

### "infisical not initialized"

- Check `INFISICAL_ENABLED=true`
- Verify all required variables are set
- Check authentication credentials

### "failed to retrieve secret"

- Verify secret exists in Infisical
- Check `INFISICAL_SECRET_PATH` is correct
- Ensure Machine Identity has access to the secret

### Slow performance

- Increase `INFISICAL_CACHE_TTL` (default: 300s)
- Use `GetString` instead of fetching individual secrets repeatedly

## Examples

### Auth Service

```go
func main() {
    ctx := context.Background()
    env.InitInfisical(ctx)
    
    cfg := &Config{
        JWT: JWTConfig{
            Secret: env.GetString("JWT_SECRET", "dev-secret"),
            // Falls back to Infisical if env var not set
        },
        Database: DatabaseConfig{
            URL: env.MustGetString("DATABASE_URL"),
            // Panics if not found in Infisical or env
        },
    }
}
```

### With Custom Context

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

if err := env.InitInfisical(ctx); err != nil {
    log.Fatal(err)
}
```

## License

MIT
