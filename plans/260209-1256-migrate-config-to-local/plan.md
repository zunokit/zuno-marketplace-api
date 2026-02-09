# Migrate Services to Local Config

## Overview

Migrate collection-service and media-service from shared/config to fully local config.

## Services to Migrate

| Service | Current Pattern | Target Pattern |
|---------|-----------------|----------------|
| collection-service | Hybrid (shared + local) | Fully local |
| media-service | Uses sharedConfig.ServerConfig | Fully local |

## Migration Steps

### Phase 1: Collection-Service

1. Remove `sharedConfig` import
2. Define local `ServerConfig` struct
3. Define local `DatabaseConfig` struct with `Mode`, `URL`, `GetDSN()`
4. Define local `RedisConfig` struct with `GetAddr()`
5. Include `RedisConfig` in main `Config` struct (not separate function)
6. Update `Load()` to return `*Config` (not `*sharedConfig.Config`)
7. Update `main.go` to use new config structure

### Phase 2: Media-Service

1. Remove `sharedConfig` import
2. Define local `ServerConfig` struct
3. Define local `DatabaseConfig` struct
4. Update `Load()` function
5. Update `main.go`

## Success Criteria

- [ ] No service imports `shared/config`
- [ ] Each service owns its config structs
- [ ] All services compile
- [ ] All tests pass
