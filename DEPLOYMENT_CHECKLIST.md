# Production Deployment Checklist

Use this checklist to ensure your DigitalOcean deployment is secure and properly configured.

## Pre-Deployment (Local Testing)

- [ ] All critical security fixes applied
- [ ] Code tested locally with `docker-compose up`
- [ ] Environment variables documented in `.env.example`
- [ ] README.md up to date
- [ ] Git repository clean and latest code committed
- [ ] No sensitive data in `.git` history (check with `git log --all -S 'password'`)

## DigitalOcean Droplet Setup

### Initial Configuration
- [ ] Droplet created (Ubuntu 22.04 LTS recommended)
- [ ] SSH key added (not password)
- [ ] Connected via SSH without password
- [ ] Timezone set correctly
- [ ] Automatic security updates enabled

### System Security
- [ ] Firewall enabled (ufw)
- [ ] Only necessary ports open (22 for SSH, 80 for HTTP, 443 for HTTPS)
- [ ] Fail2ban installed (optional but recommended)
- [ ] Root login disabled via SSH
- [ ] SSH password authentication disabled

### Software Installation
- [ ] Docker installed
- [ ] Docker Compose installed
- [ ] Nginx installed and running
- [ ] Certbot installed
- [ ] Git installed

## Domain & DNS

- [ ] Domain registrar pointed to DigitalOcean nameservers (or A records added)
- [ ] DNS records created for main domain:
  - [ ] `tourney.yourdomain.com` → Droplet IP
  - [ ] `www.tourney.yourdomain.com` → Droplet IP (optional)
- [ ] DNS propagation verified with `nslookup`
- [ ] Additional domains created if hosting multiple apps

## SSL/TLS Certificate

- [ ] Certbot installed
- [ ] Let's Encrypt certificate requested for main domain
- [ ] Certificate auto-renewal tested (`certbot renew --dry-run`)
- [ ] Certbot timer enabled (`systemctl status certbot.timer`)
- [ ] Additional certificates for other domains

## Application Deployment

### Repository & Configuration
- [ ] Repository cloned to `/apps/team-tourney-tracker`
- [ ] `.env` file created from `.env.example`
- [ ] All required environment variables set:
  - [ ] `JWT_SECRET` (generated with openssl)
  - [ ] `CORS_ORIGIN` (set to production domain)
  - [ ] `VITE_API_URL` (set to production API endpoint)
  - [ ] `SMTP_HOST`, `SENDER_EMAIL`, `SENDER_APP_PASSWORD`
  - [ ] `BACKUP_EMAIL`
- [ ] `.env` file permissions set correctly (`600`)
- [ ] `.env` file NOT in git (verified in `.gitignore`)

### Docker Configuration
- [ ] `docker-compose.prod.yml` in place
- [ ] Production ports configured (localhost only)
- [ ] Volume mounts configured for data persistence
- [ ] Resource limits set
- [ ] Health checks configured
- [ ] Containers successfully built and started

### Nginx Configuration
- [ ] Nginx config copied from `nginx.conf.example` to `/etc/nginx/sites-available/tourney-tracker`
- [ ] Domain name replaced in config
- [ ] SSL certificate paths correct
- [ ] Config tested (`nginx -t`)
- [ ] Site enabled (`ln -s sites-available sites-enabled`)
- [ ] Nginx reloaded
- [ ] Website accessible via HTTPS
- [ ] HTTP redirects to HTTPS
- [ ] Security headers present

## Application Testing

### Functionality Tests
- [ ] Frontend loads at domain URL
- [ ] Frontend connects to backend API
- [ ] Login works with test credentials
- [ ] Can create teams (admin)
- [ ] Can create players (admin)
- [ ] Can view standings
- [ ] Can view games
- [ ] Can track matches

### Security Tests
- [ ] HTTPS enforced (HTTP redirects to HTTPS)
- [ ] SSL certificate valid (no warnings)
- [ ] CORS only allows production domain
- [ ] Unauthenticated users cannot POST/PUT/DELETE
- [ ] Non-admin users cannot access admin operations
- [ ] Default admin credentials should be changed

### Performance Tests
- [ ] Page load time < 3 seconds
- [ ] API responses quick (< 1 second)
- [ ] No memory leaks (check `docker stats`)
- [ ] CPU usage reasonable

## Data & Backups

- [ ] Database backup script created and tested
- [ ] Backup directory created (`/backups`)
- [ ] Cron job scheduled for daily backups
- [ ] Manual backup created and verified restorable
- [ ] Old backups automatically deleted (30-day retention)
- [ ] Email backups configured and tested

## Monitoring & Maintenance

- [ ] Log rotation configured
- [ ] Access logs monitored for errors
- [ ] Disk usage monitored (`df -h`)
- [ ] Container memory/CPU monitored
- [ ] Update plan established (weekly check for security updates)
- [ ] Uptime monitoring service configured (optional)
- [ ] Alert system for outages (optional)

## Multiple Applications (if hosting more than one)

- [ ] Additional apps follow same structure under `/apps`
- [ ] Each app has own `.env` file
- [ ] Each app has own Docker Compose configuration
- [ ] Each app uses different ports (8080/5173, 8081/5174, etc.)
- [ ] Nginx configs created for each domain
- [ ] SSL certificates obtained for each domain
- [ ] Each app deployed and tested independently
- [ ] Nginx load balanced between apps correctly

## Documentation

- [ ] Deployment steps documented
- [ ] Admin access credentials recorded (secure location)
- [ ] Database backup/restore procedure documented
- [ ] Update procedure documented
- [ ] Troubleshooting guide created
- [ ] Emergency recovery plan documented

## Final Production Checklist

- [ ] Change default admin credentials
- [ ] Delete any test user accounts
- [ ] Enable database backups
- [ ] Set up monitoring/alerts
- [ ] Create admin account documentation
- [ ] Brief team on access procedures
- [ ] Document any customizations made
- [ ] Set up runbook for common issues

## Post-Deployment

- [ ] Monitor application for 24-48 hours
- [ ] Check logs daily for errors
- [ ] Verify daily backup completion
- [ ] Test recovery procedure (optional but recommended)
- [ ] Plan maintenance window (if needed)
- [ ] Celebrate 🎉

## Common Issues & Solutions

### Issue: "Connection refused" on domain
**Solution**: 
1. Verify DNS propagation: `nslookup yourdomain.com`
2. Check Nginx running: `systemctl status nginx`
3. Check containers running: `docker ps`
4. Check Nginx logs: `tail -f /var/log/nginx/error.log`

### Issue: SSL certificate error
**Solution**:
1. Verify cert exists: `ls /etc/letsencrypt/live/yourdomain.com/`
2. Check cert expiration: `certbot certificates`
3. Renew if needed: `certbot renew`
4. Check Nginx config: `nginx -t`

### Issue: Backend not accessible
**Solution**:
1. Check backend running: `docker ps | grep backend`
2. Check logs: `docker logs tourney-tracker-backend`
3. Test manually: `curl http://localhost:8080/teams`
4. Check firewall: `ufw status`

### Issue: Database errors in logs
**Solution**:
1. Check database file exists: `ls -la /apps/team-tourney-tracker/db/`
2. Check permissions: `ls -la /apps/team-tourney-tracker/`
3. Verify volume mount: `docker inspect tourney-tracker-backend | grep Mounts`
4. Check disk space: `df -h`

### Issue: High memory usage
**Solution**:
1. Check resource limits: `docker stats`
2. Increase limits in docker-compose.prod.yml
3. Restart containers: `./deploy.sh restart`
4. Monitor: `watch -n 1 docker stats`

## Support Resources

- **Docker Docs**: https://docs.docker.com/
- **Docker Compose Docs**: https://docs.docker.com/compose/
- **Nginx Docs**: https://nginx.org/en/docs/
- **Certbot Docs**: https://certbot.eff.org/docs/
- **DigitalOcean Docs**: https://docs.digitalocean.com/
- **Let's Encrypt**: https://letsencrypt.org/

## Success Criteria

✅ Application is deployed and accessible via HTTPS  
✅ All functionality works as expected  
✅ Users can log in and use the app  
✅ Admin functions protected and working  
✅ Database backups automated and tested  
✅ SSL certificate valid and auto-renewing  
✅ Logs monitored and clean  
✅ Downtime plan documented  

**You're ready for production!**
