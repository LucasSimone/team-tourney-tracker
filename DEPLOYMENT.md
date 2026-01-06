# PRODUCTION DEPLOYMENT CHECKLIST

## Pre-Deployment Review

### Security ✅
- [x] JWT secret externalized to environment variable
- [x] All write operations require authentication
- [x] CORS restricted to configurable origin
- [x] Email uses proper TLS certificate validation
- [x] Error messages are generic (no info leakage)
- [x] Hardcoded credentials removed
- [x] API URLs configurable via environment
- [x] .gitignore configured to protect secrets
- [x] Default credentials documented with warnings

### Documentation ✅
- [x] README.md created (setup, deployment, troubleshooting)
- [x] .env.example comprehensive and detailed
- [x] PRODUCTION_CHANGES.md documents all fixes
- [x] SECURITY_FIXES.md explains security improvements
- [x] API endpoints documented
- [x] Environment variables documented
- [x] Deployment instructions provided

### Code Quality ✅
- [x] Backend compiles (Go 1.21)
- [x] Frontend builds (Vue 3 + Vite)
- [x] All routes protected where needed
- [x] Consistent error handling
- [x] CORS headers standardized

---

## Pre-Production Setup

### 1. Environment Configuration

```bash
# Generate strong JWT secret
JWT_SECRET=$(openssl rand -base64 32)
echo "Your JWT_SECRET: $JWT_SECRET"

# Example .env file
cat > .env << EOF
# Security
JWT_SECRET=$JWT_SECRET
CORS_ORIGIN=https://yourdomain.com

# API Configuration
VITE_API_URL=https://api.yourdomain.com

# Email (Gmail with app password)
SMTP_HOST=smtp.gmail.com
SENDER_EMAIL=your-email@gmail.com
SENDER_APP_PASSWORD=xxxx xxxx xxxx xxxx
BACKUP_EMAIL=backup@example.com
EOF
```

### 2. Gmail App Password Setup

- [ ] Enable 2-Step Verification on Google Account
- [ ] Go to https://myaccount.google.com/apppasswords
- [ ] Select "Mail" and your device type
- [ ] Copy generated 16-character password
- [ ] Set as `SENDER_APP_PASSWORD` in `.env`

### 3. SSL/TLS Certificate

- [ ] Obtain certificate (Let's Encrypt recommended)
- [ ] Configure certificate paths
- [ ] Set up HTTPS reverse proxy
- [ ] Redirect HTTP to HTTPS

### 4. Database Preparation

- [ ] Backup existing database (if migrating)
- [ ] Verify database location accessible
- [ ] Check disk space for backups
- [ ] Configure backup storage

---

## Docker Deployment

### Option A: Using Docker Compose (Recommended)

```bash
# 1. Build images
docker-compose build

# 2. Verify images
docker images | grep team-tourney

# 3. Start services
docker-compose up -d

# 4. Verify services running
docker-compose ps

# 5. Check logs
docker-compose logs -f backend
docker-compose logs -f frontend

# 6. Test API
curl http://localhost:8080/teams
```

### Option B: Docker Hub Deployment

```bash
# 1. Build and tag
docker build -t your-registry/backend:v1.0.0 ./backend
docker build -t your-registry/frontend:v1.0.0 ./frontend

# 2. Push to registry
docker push your-registry/backend:v1.0.0
docker push your-registry/frontend:v1.0.0

# 3. Pull and run on production
docker pull your-registry/backend:v1.0.0
docker pull your-registry/frontend:v1.0.0

# 4. Run with environment variables
docker run -d \
  -e JWT_SECRET=$JWT_SECRET \
  -e CORS_ORIGIN=$CORS_ORIGIN \
  -e VITE_API_URL=$VITE_API_URL \
  -e SENDER_APP_PASSWORD=$SENDER_APP_PASSWORD \
  -e SENDER_EMAIL=$SENDER_EMAIL \
  -e BACKUP_EMAIL=$BACKUP_EMAIL \
  -e SMTP_HOST=$SMTP_HOST \
  -p 8080:8080 \
  -v /data/db:/db \
  your-registry/backend:v1.0.0
```

---

## Configuration Verification

### Before Starting Services

```bash
# 1. Verify environment variables are set
printenv | grep -E "JWT_SECRET|CORS_ORIGIN|VITE_API_URL|SENDER"

# 2. Check .env file exists and has correct permissions
ls -la .env
chmod 600 .env

# 3. Verify database directory exists
mkdir -p /db
chmod 755 /db

# 4. Test API connectivity (locally)
curl http://localhost:8080/teams || echo "API not responding"

# 5. Test frontend build
cd frontend && npm run build && cd ..
```

---

## Post-Deployment Verification

### Functionality Tests

- [ ] Frontend loads at https://yourdomain.com
- [ ] Login page accessible
- [ ] Default credentials work (lucas/password, baddyboi/password)
- [ ] Can view standings/games as public user
- [ ] Can login and record matches as user
- [ ] Can manage teams/players as admin
- [ ] Database backups trigger via admin panel
- [ ] Email backups send to BACKUP_EMAIL

### Security Tests

```bash
# 1. CORS is restricted
curl -H "Origin: https://evil.com" \
  http://localhost:8080/teams \
  -v 2>&1 | grep "Access-Control"
# Should NOT include Access-Control-Allow-Origin header

curl -H "Origin: https://yourdomain.com" \
  http://localhost:8080/teams \
  -v 2>&1 | grep "Access-Control"
# Should include Access-Control-Allow-Origin header

# 2. Write operations require authentication
curl -X POST http://localhost:8080/teams \
  -H "Content-Type: application/json" \
  -d '{"name": "Test"}' \
  -v 2>&1 | grep "401\|403"
# Should return 401 Unauthorized

# 3. Invalid tokens rejected
curl -X POST http://localhost:8080/teams \
  -H "Authorization: Bearer invalid" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test"}' \
  -v 2>&1 | grep "401"
# Should return 401 Unauthorized
```

### Performance Tests

- [ ] API response time < 500ms
- [ ] Frontend loads in < 3 seconds
- [ ] No 500 errors in logs
- [ ] Database queries efficient
- [ ] Memory usage stable

---

## Ongoing Operations

### Daily
- [ ] Monitor application logs for errors
- [ ] Check database size
- [ ] Verify backups are sending

### Weekly
- [ ] Review access logs for suspicious activity
- [ ] Check authentication failures
- [ ] Verify backup emails arrive
- [ ] Monitor disk space usage

### Monthly
- [ ] Update dependencies: `go get -u ./...` and `npm update`
- [ ] Review security logs
- [ ] Backup database to external storage
- [ ] Test disaster recovery procedure

### Quarterly
- [ ] Security audit of custom code
- [ ] Review and update documentation
- [ ] Performance optimization review
- [ ] Load testing if scale increases

---

## Troubleshooting

### API Won't Start

```bash
# Check logs
docker logs team-tourney-tracker_backend_1

# Verify port 8080 is free
lsof -i :8080

# Check environment variables
docker inspect team-tourney-tracker_backend_1 | grep -A 20 Env

# Verify database directory
ls -la /db/
```

### Frontend Can't Connect to API

```bash
# Check API is running
curl http://localhost:8080/teams

# Verify VITE_API_URL
grep VITE_API_URL .env

# Check CORS errors in browser console

# Verify CORS_ORIGIN setting
echo "CORS_ORIGIN is set to: $CORS_ORIGIN"
```

### Email Backups Not Sending

```bash
# Verify credentials
grep SENDER .env | grep -v SENDER_EMAIL

# Test SMTP connection
telnet smtp.gmail.com 587

# Check backend logs for email errors
docker logs team-tourney-tracker_backend_1 | grep -i email

# Verify Gmail app password (not regular password)
# Must have 16 characters
```

### Database Locked Error

```bash
# Stop all services
docker-compose stop

# Check for open connections
lsof /db/sports.db

# Remove lock files if safe
rm /db/sports.db-wal /db/sports.db-shm

# Start services again
docker-compose up -d
```

---

## Rollback Procedure

### If Issues Encountered

```bash
# 1. Stop current services
docker-compose down

# 2. Restore from backup
cp /backups/sports.db.backup /db/sports.db

# 3. Revert to previous version
docker-compose pull  # Gets last known good version
docker-compose up -d

# 4. Verify services
docker-compose ps
curl http://localhost:8080/teams
```

---

## Success Criteria

✅ **Application is production-ready when:**

- [ ] All services running without errors
- [ ] API responds on correct endpoint
- [ ] Frontend loads and authenticates
- [ ] CORS properly restricted
- [ ] JWT tokens required on write operations
- [ ] Email backups configured and sending
- [ ] No hardcoded secrets in deployment
- [ ] Logs contain useful debug information
- [ ] Database persists between restarts
- [ ] SSL/TLS certificate valid

---

## Support & Documentation

| Document | Purpose |
|----------|---------|
| `README.md` | Setup, deployment, API documentation |
| `SECURITY_FIXES.md` | Security improvements explained |
| `PRODUCTION_CHANGES.md` | Technical changelog |
| `.env.example` | Configuration reference |
| Source code comments | Implementation details |

---

## Contact & Escalation

**If issues occur:**

1. Check relevant documentation
2. Review application logs
3. Verify environment configuration
4. Test with curl commands
5. Check browser console for frontend issues
6. Consult troubleshooting section

---

**Last Updated**: January 6, 2026  
**Version**: 1.0.0 (Production Ready)
