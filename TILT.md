# Tilt/Kubernetes Development (Advanced)

> **Note**: Tilt is for advanced users who want hot reload in Kubernetes.
> For simple development, use `make dev` (Docker Compose) instead.

## Prerequisites

- Docker Desktop with Kubernetes enabled
- Tilt installed: https://docs.tilt.dev/install.html

## Setup Steps

### Step 1: Create Secrets File

**First time only:**

```bash
cd E:\zuno-marketplace-api
cp infra/development/k8s/secrets.yaml.example infra/development/k8s/secrets.yaml
```

> **Note**: `secrets.yaml` is gitignored. Keep your secrets safe!

### Step 2: Start Tilt

Open a terminal and run:

```bash
cd E:\zuno-marketplace-api
tilt up
```

**Keep this terminal open!** Don't close it.

Go to Tilt UI: http://localhost:10350

Wait until ALL services are green (✓).

### Step 2: Run Migrations

Open a **NEW terminal** and run:

```bash
cd E:\zuno-marketplace-api
make migrate
```

### Step 3: Access Services

- **Tilt UI**: http://localhost:10350
- **GraphQL Playground**: http://localhost:8081/graphql
- **PostgreSQL**: localhost:5433 (auto port-forwarded by Tilt)

## Stop Tilt

Press `Ctrl+C` in the Tilt terminal, then run:

```bash
tilt down
```

## Troubleshooting

### Port 10350 already in use

```bash
# Kill existing Tilt
taskkill //F //IM tilt.exe

# Start fresh
tilt up
```

### Services not starting

Check Tilt UI at http://localhost:10350 for error logs in each service.

### Database connection errors

Make sure Tilt auto port-forward is working:
```bash
netstat -ano | findstr :5433
# Should show tilt.exe using port 5433
```

## Hot Reload

With Tilt, when you edit code:
- Changes are automatically detected
- Services rebuild and restart
- No need to restart manually

## Why Use Tilt?

✅ Hot reload - see changes instantly
✅ Production-like environment (Kubernetes)
✅ Better resource isolation
✅ Multi-service orchestration

❌ More complex setup
❌ Requires Kubernetes knowledge
❌ Slower first startup

**For most development, use `make dev` instead!**
