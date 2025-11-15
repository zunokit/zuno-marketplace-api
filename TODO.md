# 📋 TODO - Next Implementation Plan

## 🔴 Critical (MVP Blockers)

### 1. Fix Test Compilation Errors
- [ ] Update SIWE service tests to match library API
- [ ] Fix `siwe.Message` struct initialization (use constructors instead of direct field access)
- [ ] Fix `GetExpirationTime()` return type handling
- [ ] Run and verify all unit tests pass: `go test ./... -v`

### 2. Docker Build & Deployment
- [ ] Complete docker-compose build (currently in progress)
- [ ] Fix any remaining build errors
- [ ] Test all services start correctly: `docker-compose ps`
- [ ] Verify service health checks
- [ ] Test inter-service gRPC communication

### 3. GraphQL Gateway Implementation
- [ ] Install GraphQL dependencies: `gqlgen`, `chi`
- [ ] Define GraphQL schema (`schema.graphqls`)
  - Auth queries/mutations (getNonce, verifySiwe, refreshSession)
  - User queries (me, getUser)
  - Wallet queries (myWallets, getWallets)
- [ ] Generate GraphQL resolvers: `go run github.com/99designs/gqlgen generate`
- [ ] Implement resolver logic (call gRPC services)
- [ ] Add cookie handling for refresh tokens
- [ ] Test GraphQL playground

### 4. Database Migrations
- [ ] Install golang-migrate: `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`
- [ ] Create migration files from existing `db/up.sql` files
- [ ] Test migrations: `make migrate-up`, `make migrate-down`
- [ ] Add migration commands to Makefile

---

## 🟡 High Priority (Post-MVP)

### 5. Frontend Integration
- [ ] Update frontend with RainbowKit + Custom SIWE flow
- [ ] Implement getNonce API call
- [ ] Implement wallet signature flow
- [ ] Implement verifySiwe API call
- [ ] Handle JWT tokens (localStorage + HttpOnly cookies)
- [ ] Implement auto-refresh on 401
- [ ] Test full authentication flow end-to-end

### 6. Testing & Quality
- [ ] Add integration tests for each service
- [ ] Complete E2E auth flow test with real signatures
- [ ] Add repository layer unit tests
- [ ] Add gRPC handler unit tests
- [ ] Set up test database for integration tests
- [ ] Add CI/CD test automation

### 7. Security Hardening
- [ ] Add rate limiting (Redis-based)
  - GetNonce: 5 req/min per IP
  - VerifySiwe: 10 req/min per account
- [ ] Add request validation middleware
- [ ] Implement IP geo-validation (optional)
- [ ] Add user agent fingerprinting
- [ ] Security audit of SIWE implementation
- [ ] Add CORS configuration for production

### 8. Monitoring & Observability
- [ ] Add structured logging with zerolog/zap
- [ ] Implement request ID propagation
- [ ] Add Prometheus metrics endpoints
- [ ] Set up Grafana dashboards
- [ ] Integrate Sentry for error tracking
- [ ] Add distributed tracing (Jaeger/Zipkin)

---

## 🟢 Medium Priority (Enhancement)

### 9. Session Management Features
- [ ] List active sessions endpoint
- [ ] Revoke specific session by ID
- [ ] Revoke all sessions except current
- [ ] Add device fingerprinting
- [ ] Session activity tracking

### 10. User Profile Management
- [ ] Complete user profile CRUD operations
- [ ] Add username availability check
- [ ] Implement avatar upload (IPFS/Pinata)
- [ ] Add profile privacy settings
- [ ] User preferences management

### 11. Wallet Management Features
- [ ] Multiple wallet support per user
- [ ] Primary wallet switching
- [ ] Wallet labels/nicknames
- [ ] Wallet activity audit log
- [ ] Wallet verification status

### 12. Additional Auth Features
- [ ] Email linking (optional)
- [ ] 2FA support (TOTP)
- [ ] Social auth (Twitter, Discord)
- [ ] Magic link authentication
- [ ] Passkey/WebAuthn support

---

## 🔵 Low Priority (Nice to Have)

### 13. Developer Experience
- [ ] Add Swagger/OpenAPI docs for REST endpoints
- [ ] Add gRPC reflection (already done, verify)
- [ ] Create Postman collection for APIs
- [ ] Add example GraphQL queries
- [ ] Write API usage documentation

### 14. Performance Optimization
- [ ] Add Redis caching for user data
- [ ] Implement database connection pooling tuning
- [ ] Add gRPC client connection pooling
- [ ] Optimize database queries (indexes, EXPLAIN)
- [ ] Add CDN for static assets

### 15. Infrastructure
- [ ] Set up Kubernetes production deployment
- [ ] Create Helm charts for services
- [ ] Set up production secrets management (Vault/Sealed Secrets)
- [ ] Configure auto-scaling policies
- [ ] Set up database backups

### 16. Documentation
- [ ] API documentation (GraphQL schema docs)
- [ ] Architecture decision records (ADRs)
- [ ] Deployment guide
- [ ] Troubleshooting guide
- [ ] Contributing guidelines

---

## 📦 Immediate Next Steps (Order of Execution)

1. **Fix tests** → `go test ./... -v` passes
2. **Complete Docker build** → All services running
3. **Implement GraphQL schema** → Gateway functional
4. **Test end-to-end** → Manual testing with curl/Postman
5. **Frontend integration** → Connect RainbowKit
6. **Production deployment** → K8s + monitoring

---

## 🐛 Known Issues to Fix

- [ ] SIWE test compilation errors (GetExpirationTime type mismatch)
- [ ] Proto package conflicts (auth.pb.go and user.pb.go in same directory)
- [ ] Docker Compose watch mode removed (restore for Tilt only)
- [ ] Kubernetes not configured (Tilt requires K8s cluster)
- [ ] Missing golang-migrate integration
- [ ] No rate limiting implemented yet
- [ ] GraphQL gateway is minimal placeholder

---

## 📚 Reference Documentation

- **SIWE Spec**: https://docs.login.xyz/
- **gqlgen**: https://gqlgen.com/
- **GORM**: https://gorm.io/
- **gRPC Go**: https://grpc.io/docs/languages/go/
- **RainbowKit**: https://www.rainbowkit.com/

---

## ✅ Completed (Reference)

- [x] Project cleanup (removed 8,534 lines of bloat)
- [x] Microservices architecture setup
- [x] Proto definitions for all services
- [x] Auth service (SIWE, JWT, Sessions)
- [x] User service (User, Profile management)
- [x] Wallet service (Wallet linking)
- [x] GORM models for all entities
- [x] Repository layer (with idempotent operations)
- [x] gRPC servers for all services
- [x] Docker Compose configuration
- [x] Tiltfile for hot reload
- [x] Basic unit test structure
- [x] E2E test structure
- [x] .env configuration
- [x] Makefile with development commands

---

**Last Updated**: 2025-11-15
**Status**: MVP Backend Complete - Ready for Testing & Frontend Integration
