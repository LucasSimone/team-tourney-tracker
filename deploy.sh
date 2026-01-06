#!/bin/bash

###############################################################################
# Team Tourney Tracker - Production Deployment Helper Script
# 
# Usage:
#   ./deploy.sh start    - Build and start the application
#   ./deploy.sh stop     - Stop the application
#   ./deploy.sh restart  - Restart the application
#   ./deploy.sh update   - Pull latest code and redeploy
#   ./deploy.sh logs     - View application logs
#   ./deploy.sh status   - Show application status
#   ./deploy.sh backup   - Create manual backup
###############################################################################

set -e

APP_DIR="/apps/team-tourney-tracker"
BACKUP_DIR="/backups"
COMPOSE_FILE="docker-compose.prod.yml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        exit 1
    fi
}

# Start the application
start_app() {
    check_root
    log_info "Starting Team Tourney Tracker..."
    
    cd $APP_DIR
    
    # Check if .env exists
    if [ ! -f .env ]; then
        log_error ".env file not found. Please create it from .env.example"
        exit 1
    fi
    
    # Build and start containers
    docker-compose -f $COMPOSE_FILE up -d --build
    
    log_info "Application started successfully"
    log_info "Frontend: http://localhost:5173"
    log_info "Backend API: http://localhost:8080"
}

# Stop the application
stop_app() {
    check_root
    log_info "Stopping Team Tourney Tracker..."
    
    cd $APP_DIR
    docker-compose -f $COMPOSE_FILE down
    
    log_info "Application stopped"
}

# Restart the application
restart_app() {
    log_info "Restarting Team Tourney Tracker..."
    stop_app
    sleep 2
    start_app
}

# Update application
update_app() {
    check_root
    log_info "Updating Team Tourney Tracker..."
    
    cd $APP_DIR
    
    # Pull latest code
    log_info "Pulling latest code from repository..."
    git pull origin main
    
    # Stop current containers
    log_info "Stopping current containers..."
    docker-compose -f $COMPOSE_FILE down
    
    # Rebuild and start
    log_info "Building and starting updated containers..."
    docker-compose -f $COMPOSE_FILE up -d --build
    
    log_info "Update completed successfully"
}

# View logs
view_logs() {
    cd $APP_DIR
    docker-compose -f $COMPOSE_FILE logs -f
}

# Show status
show_status() {
    log_info "Team Tourney Tracker Status:"
    echo ""
    
    cd $APP_DIR
    docker-compose -f $COMPOSE_FILE ps
    
    echo ""
    log_info "Docker Stats:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}"
}

# Create backup
backup_database() {
    check_root
    log_info "Creating database backup..."
    
    mkdir -p $BACKUP_DIR
    
    TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
    DB_FILE="$APP_DIR/db/sports.db"
    BACKUP_FILE="$BACKUP_DIR/sports_db_$TIMESTAMP.db"
    
    if [ -f "$DB_FILE" ]; then
        cp "$DB_FILE" "$BACKUP_FILE"
        log_info "Backup created: $BACKUP_FILE"
        
        # Keep only last 30 days of backups
        find $BACKUP_DIR -name "sports_db_*.db" -mtime +30 -delete
        log_info "Old backups (>30 days) removed"
    else
        log_warn "Database file not found at $DB_FILE"
    fi
}

# Health check
health_check() {
    log_info "Running health check..."
    
    # Check if containers are running
    cd $APP_DIR
    BACKEND_STATUS=$(docker-compose -f $COMPOSE_FILE ps -q backend)
    FRONTEND_STATUS=$(docker-compose -f $COMPOSE_FILE ps -q frontend)
    
    if [ -z "$BACKEND_STATUS" ]; then
        log_error "Backend container is not running"
        exit 1
    fi
    
    if [ -z "$FRONTEND_STATUS" ]; then
        log_error "Frontend container is not running"
        exit 1
    fi
    
    # Check if backend is responding
    if curl -f http://localhost:8080/teams > /dev/null 2>&1; then
        log_info "✓ Backend API is responding"
    else
        log_error "✗ Backend API is not responding"
        exit 1
    fi
    
    # Check if frontend is responding
    if curl -f http://localhost:5173 > /dev/null 2>&1; then
        log_info "✓ Frontend is responding"
    else
        log_error "✗ Frontend is not responding"
        exit 1
    fi
    
    log_info "Health check passed"
}

# Clean up old Docker images and volumes
cleanup_docker() {
    check_root
    log_warn "This will remove unused Docker images and volumes"
    read -p "Continue? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Cleaning up Docker..."
        docker image prune -a --force
        docker volume prune --force
        log_info "Cleanup complete"
    fi
}

# Main script logic
case "${1:-}" in
    start)
        start_app
        ;;
    stop)
        stop_app
        ;;
    restart)
        restart_app
        ;;
    update)
        update_app
        ;;
    logs)
        view_logs
        ;;
    status)
        show_status
        ;;
    backup)
        backup_database
        ;;
    health)
        health_check
        ;;
    cleanup)
        cleanup_docker
        ;;
    *)
        echo "Team Tourney Tracker - Deployment Helper"
        echo ""
        echo "Usage: $0 {start|stop|restart|update|logs|status|backup|health|cleanup}"
        echo ""
        echo "Commands:"
        echo "  start       - Build and start the application"
        echo "  stop        - Stop the application"
        echo "  restart     - Restart the application"
        echo "  update      - Pull latest code and redeploy"
        echo "  logs        - View application logs"
        echo "  status      - Show application status"
        echo "  backup      - Create manual database backup"
        echo "  health      - Run health check"
        echo "  cleanup     - Clean up old Docker images/volumes"
        echo ""
        exit 1
        ;;
esac
