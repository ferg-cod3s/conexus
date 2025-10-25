#!/bin/bash

# Conexus Deployment Script
# This script deploys Conexus to production using Docker Compose

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
IMAGE_NAME="conexus"
CONTAINER_NAME="conexus-prod"
COMPOSE_FILE="docker-compose.prod.yml"
BACKUP_DIR="./backups"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker is not running. Please start Docker first."
        exit 1
    fi
}

# Create necessary directories
setup_directories() {
    log_info "Setting up directories..."
    mkdir -p $BACKUP_DIR
    mkdir -p ./data
    mkdir -p ./ssl
}

# Backup existing data
backup_data() {
    if [ -d "./data" ] && [ "$(ls -A ./data)" ]; then
        log_info "Backing up existing data..."
        BACKUP_FILE="$BACKUP_DIR/conexus-backup-$(date +%Y%m%d-%H%M%S).tar.gz"
        tar -czf "$BACKUP_FILE" ./data
        log_info "Backup created: $BACKUP_FILE"
    fi
}

# Build Docker image
build_image() {
    log_info "Building Docker image..."
    docker build -t $IMAGE_NAME:latest .
    docker tag $IMAGE_NAME:latest $IMAGE_NAME:$(date +%Y%m%d-%H%M%S)
}

# Deploy the application
deploy() {
    log_info "Deploying Conexus..."
    
    # Stop existing containers
    if docker ps -q -f name=$CONTAINER_NAME | grep -q .; then
        log_warn "Stopping existing container..."
        docker compose -f $COMPOSE_FILE down
    fi
    
    # Start new containers
    log_info "Starting new containers..."
    docker compose -f $COMPOSE_FILE up -d
    
    # Wait for health check
    log_info "Waiting for service to be healthy..."
    for i in {1..30}; do
        if curl -f http://localhost:8080/health > /dev/null 2>&1; then
            log_info "Service is healthy!"
            break
        fi
        if [ $i -eq 30 ]; then
            log_error "Service failed to become healthy within 30 seconds"
            docker compose -f $COMPOSE_FILE logs
            exit 1
        fi
        sleep 1
    done
}

# Show deployment status
show_status() {
    log_info "Deployment status:"
    docker compose -f $COMPOSE_FILE ps
    
    log_info "Service health:"
    curl -s http://localhost:8080/health | jq . 2>/dev/null || curl -s http://localhost:8080/health
}

# Main execution
main() {
    log_info "Starting Conexus deployment..."
    
    check_docker
    setup_directories
    backup_data
    build_image
    deploy
    show_status
    
    log_info "Deployment completed successfully!"
    log_info "Conexus is now running at http://localhost:8080"
    log_info "MCP endpoint: http://localhost:8080/mcp"
}

# Handle script arguments
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "stop")
        log_info "Stopping Conexus..."
        docker compose -f $COMPOSE_FILE down
        ;;
    "logs")
        docker compose -f $COMPOSE_FILE logs -f
        ;;
    "status")
        show_status
        ;;
    "backup")
        backup_data
        ;;
    *)
        echo "Usage: $0 {deploy|stop|logs|status|backup}"
        exit 1
        ;;
esac