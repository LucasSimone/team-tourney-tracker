# DigitalOcean Droplet Setup & Deployment Guide

This guide covers setting up a DigitalOcean droplet to host Team Tourney Tracker and other lightweight applications with HTTPS support.

## Architecture Overview

```
DigitalOcean Droplet (Ubuntu 22.04 LTS)
├── Nginx (Reverse Proxy + SSL/TLS)
│   ├── app1.yourdomain.com → localhost:3001
│   ├── app2.yourdomain.com → localhost:3002
│   └── tourney.yourdomain.com → localhost:3003
├── Docker / Dock#### Local Backups (Always Enabled, No Configuration)
- Automatic weekly backups crea#### Export Backups for Long-Term Storage

Since only the 3 most recent backups are kept locally, periodically archive them to external storage:

```bash
# Create a tar archive of all current backups
tar -czf ~/backups_archive_$(date +%Y-%m-%d).tar.gz \
  /var/lib/docker/volumes/team-tourney-tracker_db_volume/_data/backups/

# Download to your computer
scp root@your-droplet-ip:~/backups_archive_*.tar.gz ./

# Or upload to cloud storage (AWS S3, Google Drive, Backblaze, etc.)
```und
- Manual backups triggered from Admin → Backups panel
- Download backups directly from the web UI
- **Only the 3 most recent backups are kept** (older backups auto-deleted)
- Stored at: `/var/lib/docker/volumes/team-tourney-tracker_db_volume/_data/backups/`pose (Container Orchestration)
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

## Step 5: Configure Nginx as Reverse Proxy (HTTP Only - Initial)

Create `/etc/nginx/sites-available/tourney-tracker` with HTTP only (we'll add HTTPS after getting the certificate):

```nginx
server {
    listen 80;
    server_name tourney.yourdomain.com;

    # Frontend
    location / {
        proxy_pass http://localhost:5173;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Backend API
    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Rewrite requests: /api/* → /* (backend expects paths without /api prefix)
        rewrite ^/api(/.*)$ $1 break;
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

**IMPORTANT: Do this AFTER Step 5 and BEFORE deploying the app in Step 7**

```bash
# Get SSL certificate (this updates Nginx config automatically)
certbot --nginx -d tourney.yourdomain.com

# Follow prompts:
# - Enter email address
# - Agree to terms
# - Choose to redirect HTTP to HTTPS (recommended)

# Verify certificate
certbot certificates
```

This will automatically update your Nginx config to use HTTPS!

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
```

### 7.3 Update Docker Compose for Production

Create `/apps/team-tourney-tracker/docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  backend:
    build: ./backend
    container_name: tourney-tracker-backend
    ports:
      - "127.0.0.1:8080:8080"
    volumes:
      - db_volume:/db
    restart: unless-stopped
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - CORS_ORIGIN=${CORS_ORIGIN}

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
      args:
        - VITE_API_URL=${VITE_API_URL}
    container_name: tourney-tracker-frontend
    ports:
      - "127.0.0.1:5173:5173"
    restart: unless-stopped
    environment:
      - VITE_API_URL=${VITE_API_URL}

volumes:
  db_volume:
    driver: local
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

### 7.5 Update Nginx Config for Production (After Certificate is Ready)

Now that you have the SSL certificate, update `/etc/nginx/sites-available/tourney-tracker`:

```nginx
upstream tourney_frontend {
    server 127.0.0.1:5173;
}

upstream tourney_backend {
    server 127.0.0.1:8080;
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
        
        # Rewrite requests: /api/* → /* (backend expects paths without /api prefix)
        rewrite ^/api(/.*)$ $1 break;
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
nginx -t  # Test config
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

### 8.2 Get SSL Certificates for Multiple Domains

```bash
# Get all certificates at once
certbot --nginx -d tourney.yourdomain.com -d app1.yourdomain.com -d app2.yourdomain.com

# Or get individual certificates
certbot --nginx -d app1.yourdomain.com
certbot --nginx -d app2.yourdomain.com
```

### 8.3 Nginx Config for Multiple Apps

Create separate config files for each app:

```bash
# /etc/nginx/sites-available/app1
server {
    listen 80;
    server_name app1.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name app1.yourdomain.com;
    ssl_certificate /etc/letsencrypt/live/app1.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/app1.yourdomain.com/privkey.pem;
    
    location / {
        proxy_pass http://127.0.0.1:3001;
        # ... proxy settings ...
    }
}

# /etc/nginx/sites-available/app2
server {
    listen 80;
    server_name app2.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name app2.yourdomain.com;
    ssl_certificate /etc/letsencrypt/live/app2.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/app2.yourdomain.com/privkey.pem;
    
    location / {
        proxy_pass http://127.0.0.1:3002;
        # ... proxy settings ...
    }
}
```

Enable all sites:
```bash
ln -s /etc/nginx/sites-available/app1 /etc/nginx/sites-enabled/
ln -s /etc/nginx/sites-available/app2 /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
```

## Step 9: Database Backups & Persistence

### 9.1 Understanding Data Storage

Your data is stored in a **Docker named volume** that persists on your droplet's disk:

```
Container Path:  /db/
  ├── sports.db         (main database)
  └── backups/          (backup files)
        ├── sports_backup_2026-01-06_10-30-45.db
        ├── sports_backup_2026-01-05_10-30-45.db
        └── ...

Droplet Path: /var/lib/docker/volumes/team-tourney-tracker_db_volume/_data/
  (Same structure as above)
```

**Key Point:** All data is on your droplet's disk. Backups are automatically deleted after 30 days, so you should periodically copy them to external storage.

### 9.2 Backup System Overview

The application includes a **dual-mode backup system**:

#### Local Backups (Always Enabled, No Configuration)
- Automatic weekly backups created in background
- Manual backups triggered from Admin → Backups panel
- Download backups directly from the web UI
- Backups older than 30 days automatically deleted
- Stored at: `/var/lib/docker/volumes/team-tourney-tracker_db_volume/_data/backups/`

#### Email Backups (Optional)
- Sends backup as email attachment (if SMTP configured)
- Requires working email provider (Gmail, SendGrid, etc.)
- **DigitalOcean blocks SMTP ports by default**

### 9.3 Email Backup Configuration (Optional)

If you want email backups **and have SMTP access**, add to `.env`:

```bash
SENDER_EMAIL=your-email@gmail.com
SENDER_APP_PASSWORD=your-16-char-app-password
BACKUP_EMAIL=backup-recipient@example.com
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587  # or 465 for implicit TLS
```

**To get SMTP access on DigitalOcean:**
1. Contact DigitalOcean support via control panel
2. Request "SMTP port unblock" (port 587 or 465)
3. Explain you need it for database backups
4. Wait for approval (typically 1-2 days)

**If you don't need email backups:** Simply don't set these variables. Local backups will work fine.

### 9.4 Backup Management Tasks

#### View Current Backups
```bash
# SSH into your droplet
ssh root@your-droplet-ip

# View all backup files
ls -lh /var/lib/docker/volumes/team-tourney-tracker_db_volume/_data/backups/

# Count backups
ls /var/lib/docker/volumes/team-tourney-tracker_db_volume/_data/backups/ | wc -l
```

#### Download a Backup to Your Computer
```bash
# From your local machine
scp root@your-droplet-ip:/var/lib/docker/volumes/team-tourney-tracker_db_volume/_data/backups/sports_backup_TIMESTAMP.db ./

# Or download via the web UI: Admin → Backups → Download button
```

#### Create a Backup Now (from droplet)
```bash
# Method 1: Use web UI (Admin → Backups → Create Backup Now)

# Method 2: Command line
docker exec tourney-tracker-backend curl -X POST http://localhost:8080/admin/backup \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"
```

#### Restore from Backup (if needed)
```bash
# Stop the application
cd /apps/team-tourney-tracker
docker-compose -f docker-compose.prod.yml down

# Replace the database with your backup
cp ./db/backups/sports_backup_TIMESTAMP.db ./db/sports.db

# Start the application again
docker-compose -f docker-compose.prod.yml up -d

# Verify restoration worked
docker-compose logs -f backend
```

#### Export Backups for Long-Term Storage

Since backups are deleted after 30 days, periodically archive them:

```bash
# Create a tar archive of all backups
tar -czf ~/backups_archive_$(date +%Y-%m-%d).tar.gz \
  /var/lib/docker/volumes/team-tourney-tracker_db_volume/_data/backups/

# Download to your computer
scp root@your-droplet-ip:~/backups_archive_*.tar.gz ./

# Or upload to cloud storage (AWS S3, Google Drive, etc.)
```

### 9.5 Monitoring Backup Storage

With only 3 backups at a time, storage usage is minimal. Monitor it occasionally:

```bash
# Check volume size
du -sh /var/lib/docker/volumes/team-tourney-tracker_db_volume/_data/

# Check droplet disk usage
df -h

# Check if running low on space
df -h | grep -E '(8[0-9]|9[0-9]|100)%'  # Alerts if >80% used

# List current backups (should be 0-3 files)
ls -lh /var/lib/docker/volumes/team-tourney-tracker_db_volume/_data/backups/
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
- [ ] Local backups enabled (automatic, no config needed)
- [ ] Can create/download backups from Admin panel
- [ ] Email backups configured (optional, SMTP may be blocked)
- [ ] Firewall rules configured
- [ ] Automatic updates enabled
- [ ] SSH key-based auth enabled
- [ ] Root login disabled
- [ ] Nginx gzip compression enabled
- [ ] Security headers in place
- [ ] DNS records pointing correctly
- [ ] Application accessible via domain
- [ ] Admin credentials changed from defaults
- [ ] Backup archive procedure in place (for long-term storage)

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

### Email Backup Timeout Errors (Resolved)

Email backups have been disabled. The application now creates backups to local disk only.

All backup functions work without email:
- Automatic weekly backups ✅
- Manual backup creation from admin panel ✅
- Backup download from admin panel ✅

No SMTP configuration needed!

### Backup Storage Running Full

If `/var/lib/docker/volumes/team-tourney-tracker_db_volume/` is consuming too much disk:

```bash
# Check storage usage
du -sh /var/lib/docker/volumes/team-tourney-tracker_db_volume/_data/

# View all backups (should be max 3 files)
ls -lh /var/lib/docker/volumes/team-tourney-tracker_db_volume/_data/backups/

# If still too large, export to cloud storage and restart
tar -czf ~/tournament_backups_$(date +%Y-%m-%d).tar.gz \
  /var/lib/docker/volumes/team-tourney-tracker_db_volume/_data/backups/
```

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
