# DigitalOcean Droplet Setup & Deployment Guide

This guide covers setting up a DigitalOcean droplet to host Team Tourney Tracker and other lightweight applications with HTTPS support.

## Architecture Overview

```
DigitalOcean Droplet (Ubuntu 22.04 LTS)
├── Nginx (Reverse Proxy + SSL/TLS)
│   ├── app1.yourdomain.com → localhost:3001
│   ├── app2.yourdomain.com → localhost:3002
│   └── tourney.yourdomain.com → localhost:3003
├── Docker / Docker Compose (Container Orchestration)
│   ├── team-tourney-tracker
│   ├── other-app-1
│   └── other-app-2
├── Certbot (Let's Encrypt SSL)
└── System Monitoring (Optional)
```

## Step 1: Initial Droplet Setup

### 1.1 Create the Droplet

1. Log in to DigitalOcean
2. Click "Create" → "Droplets"
3. Choose:
   - **Image**: Ubuntu 22.04 LTS
   - **Size**: Recommended minimum $5-6/month (1GB RAM, 25GB SSD)
   - **Region**: Choose closest to your users
   - **Auth**: SSH Key (recommended over password)
4. Add a hostname (e.g., `tourney-app-server`)
5. Click "Create Droplet"

### 1.2 Connect via SSH

```bash
ssh root@your_droplet_ip
```

### 1.3 Initial Security Setup

```bash
# Update system packages
apt-get update && apt-get upgrade -y

# Set timezone
timedatectl set-timezone America/New_York  # Or your timezone

# Enable automatic security updates
apt-get install -y unattended-upgrades
dpkg-reconfigure -plow unattended-upgrades

# Configure firewall
ufw enable
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow http
ufw allow https
ufw status
```

## Step 2: Install Docker & Docker Compose

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Add current user to docker group
usermod -aG docker root

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Verify installation
docker --version
docker-compose --version
```

## Step 3: Install Nginx & SSL (Certbot)

```bash
# Install Nginx
apt-get install -y nginx

# Install Certbot for Let's Encrypt SSL
apt-get install -y certbot python3-certbot-nginx

# Start Nginx
systemctl start nginx
systemctl enable nginx

# Test firewall rules
ufw status
```

## Step 4: Set Up DNS

1. Go to your domain registrar (GoDaddy, Namecheap, etc.)
2. Create an A record pointing to your droplet IP:
   - **Type**: A
   - **Name**: @ (or subdomain like `tourney`)
   - **Value**: Your droplet's IP address
   - **TTL**: 3600

3. For subdomains, create additional A records:
   - `tourney.yourdomain.com` → droplet IP
   - `api.yourdomain.com` → droplet IP (if needed)

Wait for DNS to propagate (5-30 minutes):
```bash
nslookup tourney.yourdomain.com
```

## Step 5: Configure Nginx as Reverse Proxy

Create `/etc/nginx/sites-available/tourney-tracker`:

```nginx
server {
    listen 80;
    server_name tourney.yourdomain.com;

    # Temporary - will be replaced by HTTPS
    location / {
        proxy_pass http://localhost:5173;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Enable the site:
```bash
ln -s /etc/nginx/sites-available/tourney-tracker /etc/nginx/sites-enabled/
nginx -t  # Test config
systemctl restart nginx
```

## Step 6: Set Up HTTPS with Let's Encrypt

```bash
# Get SSL certificate
certbot --nginx -d tourney.yourdomain.com

# Follow prompts:
# - Enter email address
# - Agree to terms
# - Choose to redirect HTTP to HTTPS (recommended)

# Verify certificate
certbot certificates
```

Certbot automatically updates your Nginx config with HTTPS!

## Step 7: Deploy Team Tourney Tracker

### 7.1 Prepare Application Directory

```bash
# Create apps directory
mkdir -p /apps
cd /apps

# Clone your repo
git clone https://github.com/LucasSimone/team-tourney-tracker.git
cd team-tourney-tracker

# Create .env file
cp .env.example .env
nano .env  # Edit with your configuration
```

### 7.2 Configure Environment Variables

Edit `/apps/team-tourney-tracker/.env`:

```bash
# Security
JWT_SECRET=<generate-with: openssl rand -base64 32>
CORS_ORIGIN=https://tourney.yourdomain.com

# Frontend
VITE_API_URL=https://tourney.yourdomain.com/api

# Email Backups (OPTIONAL - can skip for now)
# Note: If you get timeout errors, Gmail's port 587 may be blocked
# Use port 465 instead: SMTP_HOST=smtp.gmail.com:465
SMTP_HOST=smtp.gmail.com
SENDER_EMAIL=your-email@gmail.com
SENDER_APP_PASSWORD=xxxx xxxx xxxx xxxx
BACKUP_EMAIL=backup@yourdomain.com
```

**If you get email timeout errors:**
```bash
# Try using port 465 (SMTPS) instead of 587
SMTP_HOST=smtp.gmail.com:465
# or
SMTP_PORT=465

# Or simply disable email backups for now (comment out the SMTP_* lines)
# and enable them later
```

### 7.3 Update Docker Compose for Production

Create `/apps/team-tourney-tracker/docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  backend:
    build: ./backend
    container_name: tourney-backend
    ports:
      - "8080:8080"
    volumes:
      - ./db:/db
    restart: unless-stopped
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - CORS_ORIGIN=${CORS_ORIGIN}
      - SMTP_HOST=${SMTP_HOST}
      - SENDER_EMAIL=${SENDER_EMAIL}
      - SENDER_APP_PASSWORD=${SENDER_APP_PASSWORD}
      - BACKUP_EMAIL=${BACKUP_EMAIL}

  frontend:
    build:
      context: ./frontend
      args:
        - VITE_API_URL=${VITE_API_URL}
    container_name: tourney-frontend
    ports:
      - "5173:5173"
    restart: unless-stopped
    environment:
      - VITE_API_URL=${VITE_API_URL}
```

### 7.4 Start the Application

```bash
cd /apps/team-tourney-tracker

# Build and start containers
docker-compose -f docker-compose.prod.yml up -d

# View logs
docker-compose logs -f

# Check status
docker-compose ps
```

### 7.5 Update Nginx Config for Production

Replace `/etc/nginx/sites-available/tourney-tracker`:

```nginx
upstream tourney_frontend {
    server localhost:5173;
}

upstream tourney_backend {
    server localhost:8080;
}

server {
    listen 80;
    server_name tourney.yourdomain.com;
    return 301 https://$server_name$request_uri;  # Redirect to HTTPS
}

server {
    listen 443 ssl http2;
    server_name tourney.yourdomain.com;

    # SSL certificates (created by Certbot)
    ssl_certificate /etc/letsencrypt/live/tourney.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/tourney.yourdomain.com/privkey.pem;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Frontend
    location / {
        proxy_pass http://tourney_frontend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_redirect off;
    }

    # Backend API
    location /api {
        proxy_pass http://tourney_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 60s;
        proxy_connect_timeout 60s;
    }

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml text/javascript application/json application/javascript application/xml+rss application/rss+xml;

    # Logging
    access_log /var/log/nginx/tourney-access.log;
    error_log /var/log/nginx/tourney-error.log;
}
```

Reload Nginx:
```bash
nginx -t
systemctl reload nginx
```

## Step 8: Hosting Multiple Apps

### 8.1 Directory Structure for Multiple Apps

```bash
/apps/
├── team-tourney-tracker/
│   ├── docker-compose.prod.yml
│   └── .env
├── app-2/
│   ├── docker-compose.prod.yml
│   └── .env
└── app-3/
    ├── docker-compose.prod.yml
    └── .env
```

### 8.2 Nginx Config for Multiple Apps

Create separate config files for each app:

```bash
# /etc/nginx/sites-available/app1
server {
    listen 443 ssl http2;
    server_name app1.yourdomain.com;
    ssl_certificate /etc/letsencrypt/live/app1.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/app1.yourdomain.com/privkey.pem;
    
    location / {
        proxy_pass http://localhost:3001;
        # ... proxy settings ...
    }
}

# /etc/nginx/sites-available/app2
server {
    listen 443 ssl http2;
    server_name app2.yourdomain.com;
    ssl_certificate /etc/letsencrypt/live/app2.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/app2.yourdomain.com/privkey.pem;
    
    location / {
        proxy_pass http://localhost:3002;
        # ... proxy settings ...
    }
}
```

Enable all sites:
```bash
ln -s /etc/nginx/sites-available/app1 /etc/nginx/sites-enabled/
ln -s /etc/nginx/sites-available/app2 /etc/nginx/sites-enabled/
systemctl reload nginx
```

### 8.3 Get SSL for Multiple Domains

```bash
# Single certificate for multiple domains
certbot --nginx -d tourney.yourdomain.com -d app1.yourdomain.com -d app2.yourdomain.com

# Or get individual certificates
certbot --nginx -d app1.yourdomain.com
certbot --nginx -d app2.yourdomain.com
```

## Step 9: Database Backups & Persistence

### 9.1 Backup Strategy

Create `/apps/team-tourney-tracker/backup.sh`:

```bash
#!/bin/bash
BACKUP_DIR="/backups"
APP_DIR="/apps/team-tourney-tracker"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Create backup directory if it doesn't exist
mkdir -p $BACKUP_DIR

# Copy database
cp $APP_DIR/db/sports.db $BACKUP_DIR/sports_db_$TIMESTAMP.db

# Keep only last 30 days of backups
find $BACKUP_DIR -name "sports_db_*.db" -mtime +30 -delete

echo "Backup completed: sports_db_$TIMESTAMP.db"
```

Schedule with cron:
```bash
# Run daily at 2 AM
0 2 * * * /apps/team-tourney-tracker/backup.sh
```

## Step 10: Monitoring & Maintenance

### 10.1 Check Application Status

```bash
# View running containers
docker ps

# View logs
docker-compose -f /apps/team-tourney-tracker/docker-compose.prod.yml logs -f

# Check disk usage
df -h
du -sh /apps/*

# Check memory/CPU
free -h
top
```

### 10.2 Update Applications

```bash
cd /apps/team-tourney-tracker

# Pull latest code
git pull origin main

# Rebuild and restart
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml up -d --build
```

### 10.3 SSL Certificate Auto-Renewal

Certbot automatically renews certificates 30 days before expiry. Verify:

```bash
certbot renew --dry-run  # Test renewal
systemctl status certbot.timer  # Check renewal timer
```

## Step 11: Production Checklist

Before going live, verify:

- [ ] SSL certificate installed and working
- [ ] CORS_ORIGIN set to production domain
- [ ] JWT_SECRET set to strong random value
- [ ] Database backups configured
- [ ] Email backups working
- [ ] Firewall rules configured
- [ ] Automatic updates enabled
- [ ] SSH key-based auth enabled
- [ ] Root login disabled
- [ ] Nginx gzip compression enabled
- [ ] Security headers in place
- [ ] DNS records pointing correctly
- [ ] Application accessible via domain
- [ ] Admin credentials changed from defaults

## Step 12: Troubleshooting

### Port Already in Use
```bash
# Find what's using the port
lsof -i :8080

# Kill the process
kill -9 <PID>
```

### Nginx Not Reloading
```bash
# Test config syntax
nginx -t

# View logs
tail -f /var/log/nginx/error.log
```

### Container Won't Start
```bash
# View container logs
docker logs <container_id>

# Rebuild from scratch
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml up -d --build
```

### SSL Certificate Issues
```bash
# Check certificate status
certbot certificates

# Renew immediately if needed
certbot renew --force-renewal
```

### Email Backup Timeout Errors

If you see `i/o timeout` errors on port 587:

```bash
# Option 1: Try port 465 (SMTPS)
nano /apps/team-tourney-tracker/.env

# Change:
SMTP_HOST=smtp.gmail.com:465
# OR
SMTP_PORT=465

# Restart:
docker-compose -f docker-compose.prod.yml restart backend

# Check logs:
docker-compose -f docker-compose.prod.yml logs backend
```

If port 465 doesn't work either:

```bash
# Option 2: Disable email backups for now
nano /apps/team-tourney-tracker/.env

# Comment out all SMTP_* and BACKUP_EMAIL lines:
# SMTP_HOST=...
# SENDER_EMAIL=...
# etc.

docker-compose -f docker-compose.prod.yml restart backend
```

**Why this happens:** Some cloud providers (DigitalOcean, AWS, etc.) block outbound SMTP ports (25, 587) by default to prevent spam. Port 465 (SMTPS) often works better. You can also:
- Use a different email service (SendGrid, Mailgun, etc.)
- Request DigitalOcean to unblock SMTP ports (they may require email verification)
- Use manual backups via the admin panel instead

## Cost Estimate

**DigitalOcean:**
- Basic Droplet ($5-6/month): 1GB RAM, 25GB SSD - handles this app well
- Larger Droplet ($12/month): 2GB RAM, 50GB SSD - for multiple apps

**Domain:**
- Typically $10-15/year (or transfer for free with some registrars)

**SSL/TLS:**
- Let's Encrypt (Free! via Certbot)

**Total:** ~$60-90/year for reliable hosting

## Next Steps

1. Create DigitalOcean droplet
2. Complete initial setup (Step 1-3)
3. Configure DNS (Step 4)
4. Set up Nginx + SSL (Step 5-6)
5. Deploy application (Step 7)
6. Set up monitoring and backups

Refer to individual app documentation for specific configuration needs.
