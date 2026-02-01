# Deployment Guide

## Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- At least 2GB of RAM available
- At least 1GB of disk space

## Quick Deployment

### 1. Clone the Repository
```bash
git clone <repository-url>
cd HopfieldAssigmentProblemSolver
```

### 2. Configure Environment Variables
```bash
cp env.example .env
# Edit .env as needed
```

### 3. Start Services
```bash
# Build and start all services
docker-compose up -d

# Verify services are running
docker-compose ps
```

### 4. Verify Deployment
```bash
# Check service health
curl http://localhost:8080/health

# Test problem solving
curl -X POST http://localhost:8080/solve \
  -H "Content-Type: application/json" \
  -d '{"cost_matrix": [[1,2],[3,4]]}'
```

## Advanced Configuration

### Environment Variables

| Variable | Description | Default Value |
|----------|-------------|---------------|
| `ENV` | Execution environment | `development` |
| `PORT` | API Gateway port | `8080` |
| `GIN_MODE` | Gin mode (debug/release) | `debug` |
| `HOPFIELD_SERVICE_URL` | Hopfield service URL | `http://hopfield-service:5000` |
| `LOG_LEVEL` | Logging level | `info` |

### Nginx Configuration

The `nginx/nginx.conf` file contains the reverse proxy configuration. You can modify:

- **Rate Limiting**: Adjust `limit_req_zone` to control traffic
- **Timeouts**: Modify `proxy_*_timeout` to adjust timeouts
- **SSL**: Uncomment HTTPS section and configure certificates

### Horizontal Scaling

To scale horizontally:

```bash
# Scale API Gateway
docker-compose up -d --scale api-gateway=3

# Scale Hopfield service
docker-compose up -d --scale hopfield-service=2
```

**Note**: Make sure to configure a load balancer (like Nginx) to distribute traffic between instances.

## Monitoring

### Health Checks

Services include automatic health checks:

```bash
# Check status of all services
docker-compose ps

# View logs of a specific service
docker-compose logs hopfield-service
docker-compose logs api-gateway
```

### Performance Metrics

```bash
# View resource usage
docker stats

# View logs in real time
docker-compose logs -f
```

## Troubleshooting

### Common Issues

#### 1. Service won't start
```bash
# Check logs
docker-compose logs <service-name>

# Restart service
docker-compose restart <service-name>
```

#### 2. Connection error between services
```bash
# Check Docker network
docker network ls
docker network inspect hopfieldassigmentproblemsolver_hopfield-network
```

#### 3. Port already in use
```bash
# Change ports in docker-compose.yml
# Or stop services using the ports
sudo lsof -i :8080
sudo lsof -i :5000
```

#### 4. Memory issues
```bash
# Increase memory limits in docker-compose.yml
services:
  hopfield-service:
    deploy:
      resources:
        limits:
          memory: 1G
```

### Logs and Debugging

```bash
# View all logs
docker-compose logs

# View logs of a specific service
docker-compose logs hopfield-service
docker-compose logs api-gateway

# Follow logs in real time
docker-compose logs -f --tail=100

# Execute shell in a container
docker-compose exec hopfield-service /bin/bash
docker-compose exec api-gateway /bin/sh
```

## Production Deployment

### 1. Security Configuration

```bash
# Switch to production mode
export GIN_MODE=release
export FLASK_ENV=production

# Configure SSL/TLS
# Copy certificates to nginx/ssl/
cp your-cert.pem nginx/ssl/cert.pem
cp your-key.pem nginx/ssl/key.pem

# Uncomment HTTPS section in nginx.conf
```

### 2. Network Configuration

```yaml
# In docker-compose.yml, add external network configuration
networks:
  hopfield-network:
    external: true
    name: production-network
```

### 3. Data Persistence

```yaml
# Add volumes for logs and data
volumes:
  logs:
    driver: local
  data:
    driver: local
```

### 4. Monitoring and Alerts

- Configure Prometheus for metrics
- Configure Grafana for visualization
- Configure alerts for service failures
- Implement centralized logging (ELK Stack)

### 5. Backup and Recovery

```bash
# Configuration backup
tar -czf backup-$(date +%Y%m%d).tar.gz .

# Restore from backup
tar -xzf backup-20231201.tar.gz
```

## Maintenance

### Updates

```bash
# Update images
docker-compose pull

# Rebuild services
docker-compose build --no-cache

# Apply updates
docker-compose up -d
```

### Cleanup

```bash
# Clean stopped containers
docker-compose down

# Clean unused images
docker image prune

# Clean unused volumes
docker volume prune
```

### Code Updates

```bash
# Pull changes
git pull origin main

# Rebuild and restart
docker-compose down
docker-compose build
docker-compose up -d
```
