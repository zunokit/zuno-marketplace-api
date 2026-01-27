---
title: "Phase 1: Neon Database Integration"
description: "Verify Neon PostgreSQL compatibility and update connection handling"
status: pending
priority: P1
effort: 2h
---

# Phase 1: Neon Database Integration

## Overview

Verify Neon PostgreSQL (serverless) compatibility with existing GORM/pgx setup. Update Tiltfile comments and documentation.

## Context

**Files:**
- Config: `services/*/internal/config/config.go`
- Main: `services/*/cmd/main.go`
- Tiltfile: `Tiltfile.development`, `Tiltfile.production`
- Env: `.env.development.example`, `.env.production.example`

**Current State:**
- Uses GORM with postgres driver (wraps pgx v5)
- Config already has DATABASE_URL support for serverless mode
- Tiltfile comments reference Supabase (update to Neon)

## Requirements

### Functional
- [ ] Verify pgx v5 driver compatible with Neon
- [ ] Test Neon connection string format
- [ ] Update Tiltfile comments (Supabase → Neon)
- [ ] Update env file comments (Supabase → Neon)
- [ ] Document any Neon-specific configs

### Non-Functional
- Connection pooling suitable for serverless
- TLS enabled by default (Neon requires)
- Connection timeout handling

## Architecture

```
┌─────────────────┐
│  main.go        │
│  Load config    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  config.go      │
│  GetDSN()       │───> serverless? use DATABASE_URL
│                 │     docker? build host/port string
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  GORM + pgx     │
│  postgres.Open  │
└─────────────────┘
```

## Implementation Steps

### 1. Update Tiltfile Comments

**File:** `Tiltfile.development`

```diff
- # - PostgreSQL: Supabase (cloud)
+ # - PostgreSQL: Neon (cloud)
```

**File:** `Tiltfile.production` (similar)

### 2. Update Env File Comments

**File:** `.env.development.example`

```diff
- # - Supabase (PostgreSQL)
+ # - Neon (PostgreSQL)
# Get from: Neon Dashboard → Project → Connection Details
# Format: postgresql://[user]:[password]@[neon-host]/[dbname]?sslmode=require
```

### 3. Verify Connection String Format

Neon requires `sslmode=require`. Test in `config.go`:

```go
// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
    if c.Mode == "serverless" && c.URL != "" {
        // Neon requires SSL
        if !strings.Contains(c.URL, "sslmode=") {
            return c.URL + "?sslmode=require"
        }
        return c.URL
    }
    return "host=" + c.Host + " port=" + c.Port + " user=" + c.User +
        " password=" + c.Password + " dbname=" + c.Database + " sslmode=" + c.SSLMode
}
```

### 4. Test Neon Connection

```bash
# Using psql or Neon console
psql $DATABASE_URL -c "SELECT 1"
```

## Related Code Files

**Modify:**
- `Tiltfile.development` - Update comments
- `Tiltfile.production` - Update comments
- `.env.development.example` - Update comments
- `.env.production.example` - Update comments
- `services/auth-service/internal/config/config.go` - Add SSL to URL

**Verify:**
- `services/user-service/internal/config/config.go`
- `services/wallet-service/internal/config/config.go`
- `services/graphql-gateway/internal/config/config.go`

## Todo Checklist

- [ ] Update Tiltfile.development comments (Supabase → Neon)
- [ ] Update Tiltfile.production comments (Supabase → Neon)
- [ ] Update .env.development.example comments
- [ ] Update .env.production.example comments
- [ ] Add sslmode=require to serverless DATABASE_URL in config.GetDSN()
- [ ] Test connection with real Neon instance
- [ ] Document any Neon-specific pooling settings

## Success Criteria

- [ ] All references to Supabase updated to Neon
- [ ] DATABASE_URL with sslmode=require works
- [ ] Connection successful from local dev environment
- [ ] No breaking changes to docker mode

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Neon requires different pgx version | High | Test early, check go.mod compat |
| SSL handshake failures | Medium | Verify sslmode parameter format |
| Connection pooling differences | Low | GORM handles pooling automatically |

## Security Considerations

- SSL mode required (enforced)
- Connection string in env (never commit)
- No credential hardcoding

## Next Steps

- **Phase 2:** Upstash Redis integration
- **Phase 3:** CloudAMQP RabbitMQ integration
- **Phase 4:** Testing & validation
