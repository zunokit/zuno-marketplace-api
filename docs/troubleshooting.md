# Troubleshooting Guide

## Connection Issues

### "connection refused" to PostgreSQL

**Serverless Mode:**
- Check DATABASE_URL is correct
- Verify Supabase project is active (not paused)
- Check network connectivity
- Try: `psql $DATABASE_URL` to test connection

**Docker Mode:**
- Verify Docker is running: `docker ps`
- Check postgres container is up: `docker compose ps postgres`
- Restart: `docker compose restart postgres`

### Redis connection timeout

**Serverless Mode:**
- Verify REDIS_URL format: `redis://host:port`
- Check Upstash dashboard - database should be active
- Test with: `redis-cli -u $REDIS_URL PING`

**Docker Mode:**
- Check redis container: `docker compose ps redis`
- Verify port not in use: `lsof -i :6379`

### RabbitMQ connection failed

**Serverless Mode:**
- Verify CLOUDAMQP_URL format: `amqp://user:pass@host/vhost`
- Check CloudAMQP dashboard - instance should be running
- Note: Free instances pause after inactivity

**Docker Mode:**
- Check rabbitmq container: `docker compose ps rabbitmq`
- Management UI: http://localhost:15672 (guest/guest)

## Environment Issues

### ".env file not found"

```bash
# Run setup script
./scripts/setup-env.sh
```

### "INFRA_MODE not set"

Add to `.env`:

```bash
INFRA_MODE=serverless  # or docker
```

### Connection string working in dev but not production

- Check if using localhost URL in production
- Verify production environment has correct env vars
- Check firewall/security group rules

## Provider-Specific Issues

### Supabase

**"Project paused"**

- Free projects pause after 1 week of inactivity
- Click "Resume" in Supabase dashboard

**"Connection rate limit"**

- Free tier allows 60 concurrent connections
- Close idle connections in your code

### Upstash

**"Rate limit exceeded"**

- Free tier: 10K commands/day
- Check usage in Upstash dashboard
- Consider upgrading or switching to Docker mode

### CloudAMQP

**"Queue limit reached"**

- Free tier: 100 queues max
- Clean up unused queues in management UI
- Or upgrade to higher tier

**"Message limit reached"**

- Free tier: 10K messages
- Messages auto-delete after 28 days on free tier
- Monitor usage in dashboard

## Performance Issues

### Slow response times

**Serverless Mode:**

- Check provider region selection (closer = faster)
- Upstash HTTP has latency vs native Redis
- Consider connection pooling

**Docker Mode:**

- Check Docker resource limits
- Increase memory in Docker Desktop settings

### High memory usage

**Serverless Mode:**

- Should be minimal (no Docker overhead)
- Check for connection leaks
- Profile with: `pprof`

**Docker Mode:**

- Check container stats: `docker stats`
- Limit container memory in docker-compose.yml

## Getting Help

1. Check [Serverless Development Guide](./serverless-development.md)
2. Search existing [GitHub Issues](https://github.com/zunokit/zuno-marketplace-api/issues)
3. Create new issue with:
   - INFRA_MODE setting
   - Full error message
   - Output of `./scripts/health-check.sh`
