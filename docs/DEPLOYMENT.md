# Baseplate Deployment Guide

Complete guide for deploying Baseplate in development, staging, and production environments.

## Table of Contents

- [Requirements](#requirements)
- [Quick Start (Development)](#quick-start-development)
- [Environment Configuration](#environment-configuration)
- [Docker Deployment](#docker-deployment)
- [Production Deployment](#production-deployment)
- [Monitoring and Logging](#monitoring-and-logging)
- [Backup and Recovery](#backup-and-recovery)
- [Troubleshooting](#troubleshooting)

## Requirements

### System Requirements

**Minimum**:
- **CPU**: 2 cores
- **RAM**: 2GB
- **Disk**: 10GB
- **OS**: Linux, macOS, or Windows (with WSL2)

**Recommended (Production)**:
- **CPU**: 4+ cores
- **RAM**: 8GB+
- **Disk**: 50GB+ SSD
- **OS**: Linux (Ubuntu 22.04 LTS or similar)

### Software Requirements

**Required**:
- **Go**: 1.25.1 or higher
- **PostgreSQL**: 15 or higher
- **Docker**: 20.10+ (for containerized deployment)
- **Docker Compose**: 2.0+ (for local development)

**Optional**:
- **Nginx**: For reverse proxy
- **Certbot**: For SSL certificates
- **Systemd**: For service management

### Network Requirements

**Ports**:
- `8080`: Baseplate API (default, configurable)
- `5432`: PostgreSQL (if not using Docker)
- `443`: HTTPS (production)
- `80`: HTTP → HTTPS redirect (production)

## Quick Start (Development)

### 1. Clone Repository

```bash
git clone https://github.com/your-org/baseplate.git
cd baseplate
```

### 2. Set Environment Variables

```bash
# Generate secure JWT secret
export JWT_SECRET=$(openssl rand -base64 32)

# Optional: Override defaults
export SERVER_PORT=8080
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=user
export DB_PASSWORD=password
export DB_NAME=baseplate
```

### 3. Start PostgreSQL

```bash
make db-up
```

Wait 3 seconds for migrations to complete.

### 4. Initialize Super Admin (Required)

```bash
# Set super admin credentials
export SUPER_ADMIN_EMAIL="admin@example.com"
export SUPER_ADMIN_PASSWORD="secure-password-here"

# Create initial super admin
make init-superadmin
```

This creates the first super admin user who can manage teams and other users.

### 5. Run Baseplate

```bash
make run
```

### 6. Verify

```bash
curl http://localhost:8080/api/health
# Response: {"status":"ok"}
```

You're ready! See [API Documentation](./API.md) for usage.

---

## Environment Configuration

### Environment Variables

| Variable | Default | Description | Required |
|----------|---------|-------------|----------|
| `JWT_SECRET` | - | JWT signing secret | **Yes** |
| `SERVER_PORT` | `8080` | HTTP server port | No |
| `GIN_MODE` | `debug` | Gin mode (`debug` or `release`) | No |
| `DB_HOST` | `localhost` | PostgreSQL host | No |
| `DB_PORT` | `5432` | PostgreSQL port | No |
| `DB_USER` | `user` | PostgreSQL username | No |
| `DB_PASSWORD` | `password` | PostgreSQL password | No |
| `DB_NAME` | `baseplate` | PostgreSQL database | No |
| `DB_SSL_MODE` | `disable` | PostgreSQL SSL mode | No |
| `JWT_EXPIRATION_HOURS` | `24` | JWT token lifetime (hours) | No |
| `SUPER_ADMIN_EMAIL` | - | Initial super admin email | **Yes (for init)** |
| `SUPER_ADMIN_PASSWORD` | - | Initial super admin password | **Yes (for init)** |

### Configuration File (.env)

Create `.env` file in project root:

```bash
# Security
JWT_SECRET=your-secure-secret-here-minimum-32-characters

# Server
SERVER_PORT=8080
GIN_MODE=debug

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=baseplate
DB_SSL_MODE=disable

# JWT
JWT_EXPIRATION_HOURS=24

# Super Admin Setup
SUPER_ADMIN_EMAIL=admin@example.com
SUPER_ADMIN_PASSWORD=secure-password-here
```

**Load with**:
```bash
export $(cat .env | xargs)
```

**Never commit `.env` to version control!** Add to `.gitignore`:
```gitignore
.env
.env.*
!.env.example
```

---

## Docker Deployment

### Local Development with Docker

**docker-compose.yaml** (included):

```yaml
version: '3.8'

services:
  db:
    image: postgres:15-alpine
    container_name: baseplate_db
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: baseplate
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user -d baseplate"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

**Commands**:
```bash
# Start database
make db-up
# or: docker-compose up -d

# Stop database
make db-down
# or: docker-compose down

# Reset database (deletes all data!)
make db-reset
# or: docker-compose down -v && docker-compose up -d
```

---

### Full Docker Deployment

Create `Dockerfile` for Baseplate application:

```dockerfile
# Build stage
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Run application
CMD ["./server"]
```

**docker-compose.prod.yaml**:

```yaml
version: '3.8'

services:
  db:
    image: postgres:15-alpine
    restart: always
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER} -d ${DB_NAME}"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - baseplate

  app:
    build: .
    restart: always
    ports:
      - "8080:8080"
    environment:
      JWT_SECRET: ${JWT_SECRET}
      GIN_MODE: release
      DB_HOST: db
      DB_PORT: 5432
      DB_USER: ${DB_USER}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: ${DB_NAME}
      DB_SSL_MODE: disable
      JWT_EXPIRATION_HOURS: 24
    depends_on:
      db:
        condition: service_healthy
    networks:
      - baseplate

volumes:
  postgres_data:

networks:
  baseplate:
```

**Deploy**:
```bash
# Build and start
docker-compose -f docker-compose.prod.yaml up -d

# View logs
docker-compose -f docker-compose.prod.yaml logs -f app

# Stop
docker-compose -f docker-compose.prod.yaml down
```

---

## Production Deployment

### Deployment Checklist

#### Pre-Deployment

- [ ] Generate strong JWT_SECRET (32+ characters)
- [ ] Set GIN_MODE=release
- [ ] Configure DB_SSL_MODE=require
- [ ] Set up HTTPS/TLS certificates
- [ ] Configure firewall rules
- [ ] Set up monitoring and logging
- [ ] Configure automated backups
- [ ] Test database migrations
- [ ] Review and apply security best practices
- [ ] Set up secrets management

#### Post-Deployment

- [ ] Verify health check endpoint
- [ ] Test authentication flow
- [ ] Verify database connectivity
- [ ] Check SSL certificate validity
- [ ] Monitor resource usage
- [ ] Test backup restoration
- [ ] Configure alerting
- [ ] Document deployment

---

### Production Build

```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -a -installsuffix cgo \
  -ldflags="-w -s" \
  -o server \
  ./cmd/server

# Binary size ~15MB

# Copy to server
scp server user@your-server:/opt/baseplate/
scp -r migrations user@your-server:/opt/baseplate/
```

---

### Systemd Service

Create `/etc/systemd/system/baseplate.service`:

```ini
[Unit]
Description=Baseplate API Server
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=baseplate
Group=baseplate
WorkingDirectory=/opt/baseplate

# Environment
Environment="JWT_SECRET=your-secret-here"
Environment="GIN_MODE=release"
Environment="SERVER_PORT=8080"
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_USER=baseplate"
Environment="DB_PASSWORD=secure-password"
Environment="DB_NAME=baseplate"
Environment="DB_SSL_MODE=require"

# Or load from file
EnvironmentFile=/opt/baseplate/.env

# Start
ExecStart=/opt/baseplate/server

# Restart policy
Restart=always
RestartSec=5

# Limits
LimitNOFILE=65536

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=baseplate

[Install]
WantedBy=multi-user.target
```

**Commands**:
```bash
# Reload systemd
sudo systemctl daemon-reload

# Start service
sudo systemctl start baseplate

# Enable on boot
sudo systemctl enable baseplate

# Check status
sudo systemctl status baseplate

# View logs
sudo journalctl -u baseplate -f

# Restart
sudo systemctl restart baseplate
```

---

### Nginx Reverse Proxy

Create `/etc/nginx/sites-available/baseplate`:

```nginx
# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name api.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    # SSL certificates
    ssl_certificate /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.yourdomain.com/privkey.pem;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Logging
    access_log /var/log/nginx/baseplate_access.log;
    error_log /var/log/nginx/baseplate_error.log;

    # Proxy to Baseplate
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;

        # Headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        # Buffering
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
        proxy_busy_buffers_size 8k;
    }

    # Health check endpoint (no authentication)
    location /api/health {
        proxy_pass http://localhost:8080/api/health;
        access_log off;
    }
}
```

**Enable and restart**:
```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/baseplate /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

---

### SSL Certificates (Let's Encrypt)

```bash
# Install Certbot
sudo apt update
sudo apt install certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx -d api.yourdomain.com

# Auto-renewal (certbot installs cron job automatically)
# Test renewal
sudo certbot renew --dry-run

# Check renewal timer
sudo systemctl status certbot.timer
```

---

### Database Setup

**Production PostgreSQL**:

```bash
# Install PostgreSQL
sudo apt update
sudo apt install postgresql-15

# Create user and database
sudo -u postgres psql

postgres=# CREATE USER baseplate WITH PASSWORD 'secure-password';
postgres=# CREATE DATABASE baseplate OWNER baseplate;
postgres=# GRANT ALL PRIVILEGES ON DATABASE baseplate TO baseplate;
postgres=# \q

# Run migrations
psql -U baseplate -d baseplate -f migrations/001_initial.sql

# Configure SSL
# Edit /etc/postgresql/15/main/postgresql.conf
ssl = on
ssl_cert_file = '/etc/ssl/certs/server.crt'
ssl_key_file = '/etc/ssl/private/server.key'

# Restart PostgreSQL
sudo systemctl restart postgresql
```

---

### Secrets Management

**Environment File** (simple):
```bash
# /opt/baseplate/.env
JWT_SECRET=your-secure-secret
DB_PASSWORD=secure-password

# Secure permissions
chmod 600 /opt/baseplate/.env
chown baseplate:baseplate /opt/baseplate/.env
```

**AWS Secrets Manager**:
```bash
# Store secret
aws secretsmanager create-secret \
  --name baseplate/jwt-secret \
  --secret-string "your-secret-here"

# Retrieve in startup script
JWT_SECRET=$(aws secretsmanager get-secret-value \
  --secret-id baseplate/jwt-secret \
  --query SecretString \
  --output text)

export JWT_SECRET
```

---

## Monitoring and Logging

### Health Check

```bash
# Health endpoint
curl https://api.yourdomain.com/api/health

# Response: {"status":"ok"}

# Use in monitoring tools
# Uptime Robot, Pingdom, etc.
```

---

### Application Logs

**Systemd Journal**:
```bash
# View recent logs
sudo journalctl -u baseplate -n 100

# Follow logs
sudo journalctl -u baseplate -f

# Filter by time
sudo journalctl -u baseplate --since "1 hour ago"

# Export logs
sudo journalctl -u baseplate > baseplate.log
```

**Log Aggregation** (e.g., ELK Stack, Grafana Loki):
```bash
# Install filebeat
sudo apt install filebeat

# Configure filebeat to ship logs
# /etc/filebeat/filebeat.yml
filebeat.inputs:
  - type: journald
    id: baseplate
    include_matches:
      - _SYSTEMD_UNIT=baseplate.service

output.logstash:
  hosts: ["logstash:5044"]
```

---

### Metrics and Monitoring

**Prometheus Integration** (add to Baseplate):

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

**Key Metrics**:
- Request rate (requests/second)
- Response time (p50, p95, p99)
- Error rate (4xx, 5xx responses)
- Database connection pool usage
- Active API keys usage

---

## Backup and Recovery

### Database Backup

**Automated Backup Script**:

```bash
#!/bin/bash
# /opt/baseplate/scripts/backup.sh

BACKUP_DIR="/opt/baseplate/backups"
DATE=$(date +%Y%m%d_%H%M%S)
FILENAME="baseplate_backup_$DATE.sql.gz"

# Create backup
pg_dump -U baseplate baseplate | gzip > "$BACKUP_DIR/$FILENAME"

# Keep only last 30 days
find "$BACKUP_DIR" -name "baseplate_backup_*.sql.gz" -mtime +30 -delete

# Upload to S3 (optional)
aws s3 cp "$BACKUP_DIR/$FILENAME" s3://your-bucket/baseplate-backups/
```

**Cron Job**:
```bash
# Daily backup at 2 AM
0 2 * * * /opt/baseplate/scripts/backup.sh
```

---

### Recovery

```bash
# Stop application
sudo systemctl stop baseplate

# Restore from backup
gunzip < backup_20240115_020000.sql.gz | psql -U baseplate -d baseplate

# Restart application
sudo systemctl start baseplate

# Verify
curl https://api.yourdomain.com/api/health
```

---

## Troubleshooting

### Common Issues

#### Application Won't Start

**Check logs**:
```bash
sudo journalctl -u baseplate -n 50
```

**Common causes**:
- Missing JWT_SECRET
- Database connection failure
- Port already in use
- Permission issues

---

#### Database Connection Errors

**Test connection**:
```bash
psql -h localhost -U baseplate -d baseplate
```

**Check PostgreSQL status**:
```bash
sudo systemctl status postgresql
```

**Review PostgreSQL logs**:
```bash
sudo tail -f /var/log/postgresql/postgresql-15-main.log
```

---

#### 502 Bad Gateway (Nginx)

**Causes**:
- Baseplate not running
- Wrong proxy_pass port
- Firewall blocking connection

**Check**:
```bash
# Is Baseplate listening?
sudo netstat -tulpn | grep 8080

# Test direct connection
curl http://localhost:8080/api/health

# Check Nginx error log
sudo tail -f /var/log/nginx/baseplate_error.log
```

---

#### High Memory Usage

**Check memory**:
```bash
# Application memory
ps aux | grep server

# Database memory
ps aux | grep postgres
```

**Tune connection pool** (internal/storage/postgres/client.go):
```go
db.SetMaxOpenConns(25)  // Reduce if memory constrained
db.SetMaxIdleConns(5)
```

---

#### Slow Queries

**Enable slow query log** in PostgreSQL:
```postgresql.conf
log_min_duration_statement = 1000  # Log queries > 1s
```

**Analyze queries**:
```sql
SELECT * FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 10;
```

**See**: [DATABASE.md](./DATABASE.md) for optimization tips

---

### Getting Help

- **Documentation**: Check [API.md](./API.md), [ARCHITECTURE.md](./ARCHITECTURE.md)
- **Logs**: Always include logs when reporting issues
- **GitHub Issues**: Open issue with reproduction steps
- **Community**: Join discussions (if available)

---

For development setup, see [DEVELOPMENT.md](./DEVELOPMENT.md).
For security hardening, see [SECURITY.md](./SECURITY.md).
For database operations, see [DATABASE.md](./DATABASE.md).
