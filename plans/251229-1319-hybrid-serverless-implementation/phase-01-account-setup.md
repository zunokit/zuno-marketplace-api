# Phase 1: Account Setup & Configuration

**Priority**: P1
**Status**: Pending
**Effort**: 1 hour

## Context Links

- Research Report: `plans/reports/brainstormer-251229-1302-hybrid-serverless-research.md`
- GitHub Issue: [#30](https://github.com/zunokit/zuno-marketplace-api/issues/30)

## Overview

Create accounts on Supabase, Upstash, and CloudAMQP. Generate connection strings for local development environment configuration.

**Goal**: Obtain all connection strings needed for environment configuration.

## Key Insights

1. All three providers have generous free tiers suitable for development
2. CloudAMQP Little Lemur is specifically designed for dev/hobby use
3. Supabase includes additional features (auth, storage) - may be useful later
4. Connection strings use standard formats - no vendor lock-in

## Requirements

### Functional Requirements

- FR1: Create Supabase project with PostgreSQL database
- FR2: Create Upstash Redis database
- FR3: Create CloudAMQP instance (Little Lemur plan)
- FR4: Generate and securely store connection strings
- FR5: Document all connection details in secure location

### Non-Functional Requirements

- NFR1: Use GitHub/GitLab OAuth for account creation (no passwords to manage)
- NFR2: Use default/free tier plans
- NFR3: Select closest region to developer location (reduces latency)
- NFR4: Save connection strings to password manager, not git repo

## Architecture

```
Developer Laptop
    │
    ├── Supabase (PostgreSQL)
    │   ├── Host: db.xxx.supabase.co
    │   ├── Port: 5432
    │   └── Database: postgres
    │
    ├── Upstash (Redis)
    │   ├── Host: xxx.upstash.io
    │   └── Port: 6379
    │
    └── CloudAMQP (RabbitMQ)
        ├── Host: xxx.rmq.cloudamqp.com
        ├── Port: 5672
        └── VHost: xxx
```

## Implementation Steps

### Step 1: Supabase Account & Project

1. **Create Account**

   - Go to: https://supabase.com
   - Click "Start your project"
   - Sign in with GitHub (recommended)

2. **Create New Project**

   - Click "New Project"
   - Organization: Select or create default
   - Name: `zuno-api`
   - Database Password: Generate secure password (save to password manager)
   - Region: Select closest to you (e.g., Southeast Asia for Vietnam)
   - Pricing Plan: **Free** (default)

3. **Get Connection String**
   - Go to Project Settings → Database
   - Find "Connection string" section
   - Select "URI" tab
   - Copy connection string (format below):
   ```
   postgresql://postgres:[YOUR-PASSWORD]@db.xxx.supabase.co:5432/postgres
   ```
   - Save to password manager as `SUPABASE_DATABASE_URL`

### Step 2: Upstash Account & Redis Database

1. **Create Account**

   - Go to: https://upstash.com
   - Click "Sign Up"
   - Sign in with GitHub/GitLab

2. **Create Redis Database**

   - Go to Dashboard
   - Click "Create Database"
   - Name: `zuno-marketplace-redis`
   - Region: Select closest to you
   - Pricing: **Free** (default)

3. **Get Connection Details**
   - Go to Database Details
   - Copy "REST API URL" and "REST API Token"
   - Or use "Redis UPSTASH_REST_URL" and "UPSTASH_REST_TOKEN"
   - Save to password manager

### Step 3: CloudAMQP Account & Instance

1. **Create Account**

   - Go to: https://www.cloudamqp.com
   - Click "Sign Up"
   - Sign in with GitHub/GitLab

2. **Create Instance (Little Lemur Plan)**

   - Go to Dashboard
   - Click "Create New Instance"
   - Name: `zuno-marketplace-rabbitmq`
   - Region: Select closest to you
   - Plan: **Little Lemur** (FREE)
   - Confirm: 100 queues, 10K messages sufficient for dev

3. **Get Connection Details**
   - Go to Instance Details
   - Copy "AMQP URL" (format):
   ```
   amqp://user:password@xxx.rmq.cloudamqp.com/vhost
   ```
   - Save to password manager as `CLOUDAMQP_URL`

### Step 4: Document Connection Details

Create local file (DO NOT commit to git):

```bash
# Create secure local file (add to .gitignore)
touch ~/supabase-upstash-cloudamqp-connections.txt
chmod 600 ~/supabase-upstash-cloudamqp-connections.txt
```

Add to file:

```
# Supabase (PostgreSQL)
SUPABASE_DATABASE_URL=postgresql://postgres:[PASSWORD]@db.xxx.supabase.co:5432/postgres
SUPABASE_ANON_KEY=[from dashboard]
SUPABASE_SERVICE_ROLE_KEY=[from dashboard]

# Upstash (Redis)
UPSTASH_REDIS_REST_URL=https://xxx.upstash.io
UPSTASH_REDIS_REST_TOKEN=AXxX...xXxX

# CloudAMQP (RabbitMQ)
CLOUDAMQP_URL=amqp://xxx:xxx@xxx.rmq.cloudamqp.com/xxx
```

## Todo List

- [ ] Create Supabase account
- [ ] Create Supabase project
- [ ] Save Supabase connection string to password manager
- [ ] Create Upstash account
- [ ] Create Upstash Redis database
- [ ] Save Upstash connection details to password manager
- [ ] Create CloudAMQP account
- [ ] Create CloudAMQP instance (Little Lemur)
- [ ] Save CloudAMQP URL to password manager
- [ ] Create local connection details file (chmod 600)
- [ ] Add connection details file to .gitignore

## Success Criteria

- [ ] All three accounts created with OAuth
- [ ] All connection strings obtained and stored securely
- [ ] Connection details file exists locally with correct permissions
- [ ] Free tier confirmed for all services
- [ ] Regions selected closest to developer location

## Risk Assessment

| Risk                                 | Probability | Impact | Mitigation                          |
| ------------------------------------ | ----------- | ------ | ----------------------------------- |
| Email already registered on provider | Low         | Low    | Use existing account, login         |
| Region not available                 | Low         | Low    | Select nearest available region     |
| Free tier exhausted during setup     | Very Low    | Low    | Free tier sufficient for single dev |

## Security Considerations

- **Auth**: OAuth preferred over passwords (GitHub/GitLab)
- **Storage**: Password manager only, never commit to git
- **File Permissions**: Connection file chmod 600, user read-only
- **Rotation**: Document how to rotate connection strings if needed

## Next Steps

- Proceed to [Phase 2: Environment Configuration](./phase-02-environment-config.md)
- Have connection strings ready for .env files

## Unresolved Questions

- Should we use Supabase Auth for future user authentication?
- Do we need to create separate environments (dev, staging)?
- What's the process for rotating compromised connection strings?
