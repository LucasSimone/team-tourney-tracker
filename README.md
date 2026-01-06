# Team Tourney Tracker

A web application for managing and tracking tournament results, team standings, and player statistics.

## Features

- **Team Management**: Create and manage tournament teams
- **Player Tracking**: Track players across teams with detailed statistics
- **Match Recording**: Record match results and live scoring
- **Standings**: View real-time team standings and win percentages
- **Player Statistics**: Track individual player performance metrics
- **Database Backups**: Automated weekly local backups with download capability
- **Admin Dashboard**: Comprehensive admin interface for managing all data

## Technology Stack

### Backend
- **Language**: Go 1.21
- **Framework**: Standard Go `net/http`
- **Database**: SQLite3
- **Authentication**: JWT (JSON Web Tokens)
- **Email**: Go Mail with Gmail SMTP

### Frontend
- **Framework**: Vue 3 with Composition API
- **Language**: TypeScript
- **Build Tool**: Vite
- **Router**: Vue Router 4

## Prerequisites

- Docker and Docker Compose
- Or manually:
  - Go 1.21+
  - Node.js 18+
  - npm or yarn

## Quick Start with Docker

1. **Clone the repository**
   ```bash
   git clone https://github.com/LucasSimone/team-tourney-tracker.git
   cd team-tourney-tracker
   ```

2. **Create environment file**
   ```bash
   cp .env.example .env
   ```

3. **Configure environment variables** (see Configuration section)

4. **Build and run with Docker**
   ```bash
   docker-compose up --build
   ```

The application will be available at:
- Frontend: `http://localhost:5173` (or your configured URL)
- Backend API: `http://localhost:8080`

## Manual Setup (Development)

### Backend Setup

1. Navigate to backend directory
   ```bash
   cd backend
   ```

2. Install dependencies
   ```bash
   go mod download
   ```

3. Build and run
   ```bash
   go build -o server .
   ./server
   ```

The backend will run on `http://localhost:8080`

### Frontend Setup

1. Navigate to frontend directory
   ```bash
   cd frontend
   ```

2. Install dependencies
   ```bash
   npm install
   ```

3. Run development server
   ```bash
   npm run dev
   ```

The frontend will run on `http://localhost:5173`

## Configuration

### Environment Variables

Create a `.env` file in the root directory with the following variables:

#### Core Configuration
```bash
# JWT Secret Key (generates warning if not set)
JWT_SECRET=your-strong-secret-key-here

# CORS Origin (restrict API access to specific domain)
CORS_ORIGIN=https://your-domain.com

# API URL for Frontend (build-time configuration)
VITE_API_URL=http://localhost:8080
```

#### Optional: Email Backups
```bash
# Gmail SMTP Configuration (optional - local backups work without this)
SMTP_HOST=smtp.gmail.com
SENDER_EMAIL=your-email@gmail.com
SENDER_APP_PASSWORD=xxxx xxxx xxxx xxxx
BACKUP_EMAIL=recipient@example.com
SMTP_PORT=587  # or 465 for implicit TLS
```

### Database Backups

**How Backups Work:**
- ✅ Local backups are **always enabled** - stored in `/db/backups/`
- ✅ Weekly automatic backups run in the background
- ✅ Admins can manually trigger backups from Admin → Backups panel
- ✅ Admins can download backups directly from the UI
- 📧 Email backups are **optional** - only if SMTP is configured

**Accessing Backups:**
1. Log in as admin
2. Go to Admin → Backups
3. Create manual backups or download existing ones

**Note on Cloud Providers:**
DigitalOcean blocks SMTP ports by default. See [DIGITALOCEAN_DEPLOYMENT.md](./DIGITALOCEAN_DEPLOYMENT.md) for solutions.

### Gmail App Password Setup (Optional)

To enable email backups with Gmail:

1. Go to [Google Account Settings](https://myaccount.google.com/apppasswords)
2. Select "Mail" and your device type
3. Copy the generated 16-character password
4. Add to `.env` as `SENDER_APP_PASSWORD`

## Default Users

Two default user accounts are pre-configured:

| Username  | Password | Role  |
|-----------|----------|-------|
| lucas     | password | admin |
| baddyboi  | password | user  |

⚠️ **Security Warning**: Change these credentials or use the admin panel to create new users and delete defaults.

## API Endpoints

### Authentication
- `POST /auth/login` - Login with username and password
- `POST /auth/register` - Register a new user
- `GET/POST/DELETE /auth/users` - Manage users (admin only)

### Backups
- `POST /admin/backup` - Create a new database backup
- `GET /admin/backup` - List available backups
- `GET /admin/backup/download?file=filename` - Download a backup file

### Teams
- `GET /teams` - List all teams
- `POST /teams` - Create team
- `GET /teams/:id` - Get team details
- `PUT /teams/:id` - Update team
- `DELETE /teams/:id` - Delete team
- `GET /teams/:id/players` - Get team players

### Players
- `GET /players` - List all players
- `POST /players` - Create player
- `GET /players/:id` - Get player details
- `PUT /players/:id` - Update player
- `DELETE /players/:id` - Delete player
- `GET /players/stats` - Get player statistics

### Seasons
- `GET /seasons` - List all seasons
- `POST /seasons` - Create season
- `GET /seasons/:id` - Get season details
- `PUT /seasons/:id` - Update season
- `DELETE /seasons/:id` - Delete season

### Matches
- `GET /matches` - List matches
- `POST /matches` - Record match result
- `GET /matches/:id` - Get match details
- `PUT /matches/:id` - Update match
- `DELETE /matches/:id` - Delete match

### Admin
- `POST /admin/backup` - Trigger database backup (admin only)

## Authentication

The application uses JWT (JSON Web Tokens) for authentication:

1. Send credentials to `/auth/login`
2. Receive JWT token in response
3. Include token in subsequent requests: `Authorization: Bearer <token>`
4. Token expires after 7 days

Token validation is enforced on protected endpoints.

## Database

SQLite database is stored in `/db/sports.db`

The database schema includes:
- `teams` - Team information
- `players` - Player information
- `team_players` - Team-player associations
- `seasons` - Tournament seasons
- `matches` - Match results
- `users` - User accounts with roles

## Deployment

### Production Checklist

- [ ] Set `JWT_SECRET` environment variable to a strong random value
- [ ] Set `CORS_ORIGIN` to your production domain
- [ ] Set `VITE_API_URL` to production API endpoint
- [ ] Configure Gmail app password for backups
- [ ] Use HTTPS in production
- [ ] Review and update default user credentials
- [ ] Enable database backups
- [ ] Set up log aggregation
- [ ] Configure backup storage/retrieval mechanism

### Docker Deployment

```bash
# Build production images
docker-compose build

# Push to registry (e.g., Docker Hub)
docker tag team-tourney-tracker-backend:latest your-registry/backend:latest
docker push your-registry/backend:latest

docker tag team-tourney-tracker-frontend:latest your-registry/frontend:latest
docker push your-registry/frontend:latest

# Deploy to production
docker-compose -f docker-compose.yml pull
docker-compose up -d
```

## Troubleshooting

### Backend won't start
- Check port 8080 is not in use
- Verify environment variables are set correctly
- Check logs: `docker logs team-tourney-tracker_backend_1`

### Frontend can't connect to API
- Verify `VITE_API_URL` is set correctly
- Check CORS_ORIGIN matches your frontend URL
- Ensure backend is running and accessible
- Check browser console for CORS errors

### Database backups not accessible
- Backups are stored locally in `/db/backups/` by default
- Check that the directory exists and has write permissions
- Use the Admin → Backups panel to create and download backups
- No SMTP configuration needed for local backups

### Email backups not working (optional)
- Email backups are optional - local backups work without SMTP
- If SMTP is blocked, check with your hosting provider
- On DigitalOcean, request SMTP port unblock via support ticket
- Alternatively, use SendGrid API or stick with local backups

### Database locked error
- Only one process should access database at a time
- Stop other running instances
- Restart the application

## Development

### Adding a new feature

1. Create a new database table (if needed) in `initDB()`
2. Add handler function in backend
3. Register route in `main.go`
4. Create corresponding Vue component in frontend
5. Add route in `frontend/src/router.ts`
6. Test with frontend

### Testing

Run backend tests:
```bash
cd backend
go test ./...
```

Run frontend development server with hot reload:
```bash
cd frontend
npm run dev
```

## Security Considerations

- **Never commit `.env` files** to version control
- **Use environment variables** for all secrets
- **Keep dependencies updated**: `go get -u ./...` and `npm update`
- **Use HTTPS in production**
- **Implement rate limiting** for production
- **Regular database backups** enabled
- **Review access logs** periodically
- **Update default credentials** immediately

## License

[Your License Here]

## Contributing

[Contributing Guidelines]

## Support

For issues and questions, please open a GitHub issue.

## Changelog

### Version 0.1.0
- Initial release
- Basic team and player management
- Match recording and standings
- Player statistics
- Database backup functionality
