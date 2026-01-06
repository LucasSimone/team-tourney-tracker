# Email Backup Troubleshooting Guide

## Problem
The application is timing out when trying to send email backups to Gmail's SMTP server.

Error: `dial tcp 172.253.63.108:587: i/o timeout`

## Root Cause
DigitalOcean blocks outbound SMTP connections by default for security reasons to prevent spam.

## Solution Options

### Option 1: Disable Backup Emails (Recommended for Testing)
If you don't need automated email backups yet, simply don't set `BACKUP_EMAIL`:

```bash
# Edit .env
nano .env

# Comment out or remove:
# BACKUP_EMAIL=your-email@example.com
```

This disables the automatic backup feature. You can:
- Still trigger manual backups from the admin panel (if you set it up later)
- Use Docker volumes to persist the database
- Back up manually when needed

**Restart the app:**
```bash
docker-compose down
docker-compose up -d
```

---

### Option 2: Request SMTP Access from DigitalOcean
If you need email backups, request SMTP access:

1. Go to [DigitalOcean Support](https://cloud.digitalocean.com/support/tickets)
2. Submit a ticket requesting SMTP relay access
3. Explain your use case (database backups, transactional emails)
4. They'll typically enable it within 24 hours

**Note:** This is free but takes time.

---

### Option 3: Use SendGrid or Mailgun (Fastest)
Use a third-party email service instead of Gmail:

#### Option 3A: SendGrid (Recommended)
1. Sign up at [SendGrid](https://sendgrid.com) (free tier available)
2. Get your API key
3. Update your `.env`:

```bash
SMTP_HOST=smtp.sendgrid.net
SENDER_EMAIL=apikey  # Note: literally "apikey"
SENDER_APP_PASSWORD=SG.your_sendgrid_api_key_here
BACKUP_EMAIL=your-email@example.com
```

#### Option 3B: Mailgun
1. Sign up at [Mailgun](https://www.mailgun.com) (free tier: 100 emails/day)
2. Get your SMTP credentials from domain settings
3. Update `.env`:

```bash
SMTP_HOST=smtp.mailgun.org
SENDER_EMAIL=postmaster@your-domain.com
SENDER_APP_PASSWORD=your_mailgun_password
BACKUP_EMAIL=your-email@example.com
```

---

### Option 4: Use DigitalOcean Spaces + S3 (More Secure)
Instead of emailing backups, upload to cloud storage:

This requires code changes to `backup.go` - more complex but very secure.

---

## Recommended Action

For now, I recommend **Option 1 (Disable backups temporarily)**:

```bash
cd /apps/team-tourney-tracker

# Edit .env
nano .env

# Find and comment out these lines:
# BACKUP_EMAIL=your-email@example.com

# Save (Ctrl+X, Y, Enter)

# Restart
docker-compose down
docker-compose up -d

# Check logs
docker-compose logs -f
```

The application will start normally without the email timeout error.

---

## Later: Enable Email Backups

Once you decide on an option:

1. Update `.env` with your email service credentials
2. Restart: `docker-compose down && docker-compose up -d`
3. Test by triggering a backup from the admin panel

---

## Current Status

Your application is **fully functional** - the email backup failure doesn't affect normal operation:
- ✅ Frontend running on port 5173
- ✅ Backend running on port 8080
- ✅ All data operations work
- ❌ Email backups fail (non-critical for now)

The warning in logs is just informational since `BACKUP_EMAIL` is configured. Once you disable it or configure proper SMTP, the warning disappears.
