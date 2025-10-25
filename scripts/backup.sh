#!/bin/bash

# Conexus Backup Script
# This script handles backup and recovery operations for Conexus data

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BACKUP_DIR="./backups"
DATA_DIR="./data"
CONTAINER_NAME="conexus-prod"
DB_FILE="conexus.db"
RETENTION_DAYS=7
COMPOSE_FILE="docker-compose.prod.yml"

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

# Create backup directory
setup_backup_dir() {
    mkdir -p $BACKUP_DIR
    log_info "Backup directory: $BACKUP_DIR"
}

# Backup database
backup_database() {
    local timestamp=$(date +%Y%m%d-%H%M%S)
    local backup_file="$BACKUP_DIR/conexus-db-$timestamp.sqlite"
    
    log_info "Creating database backup..."
    
    if docker ps -q -f name=$CONTAINER_NAME | grep -q .; then
        # Backup from running container
        docker exec $CONTAINER_NAME sqlite3 /data/$DB_FILE ".backup $backup_file"
        log_info "Database backed up to: $backup_file"
    else
        # Backup from local file
        if [ -f "$DATA_DIR/$DB_FILE" ]; then
            cp "$DATA_DIR/$DB_FILE" "$backup_file"
            log_info "Database copied to: $backup_file"
        else
            log_error "Database file not found: $DATA_DIR/$DB_FILE"
            return 1
        fi
    fi
    
    # Compress backup
    gzip "$backup_file"
    log_info "Backup compressed: ${backup_file}.gz"
    
    echo "${backup_file}.gz"
}

# Backup configuration
backup_config() {
    local timestamp=$(date +%Y%m%d-%H%M%S)
    local config_backup="$BACKUP_DIR/config-$timestamp.tar.gz"
    
    log_info "Creating configuration backup..."
    
    tar -czf "$config_backup" \
        docker-compose*.yml \
        config*.yml \
        nginx.conf \
        scripts/ \
        .env* 2>/dev/null || true
    
    log_info "Configuration backed up to: $config_backup"
    echo "$config_backup"
}

# Full backup
backup_full() {
    local timestamp=$(date +%Y%m%d-%H%M%S)
    local full_backup="$BACKUP_DIR/conexus-full-$timestamp.tar.gz"
    
    log_info "Creating full backup..."
    
    # Create temporary directory
    local temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT
    
    # Backup database
    local db_backup=$(backup_database)
    cp "$db_backup" "$temp_dir/"
    
    # Backup configuration
    local config_backup=$(backup_config)
    cp "$config_backup" "$temp_dir/"
    
    # Backup data directory (excluding database)
    if [ -d "$DATA_DIR" ]; then
        tar -czf "$temp_dir/data.tar.gz" -C "$DATA_DIR" --exclude="$DB_FILE" .
    fi
    
    # Create backup manifest
    cat > "$temp_dir/manifest.txt" << EOF
Conexus Backup Manifest
=======================
Backup Date: $(date)
Backup Type: Full
Version: $(git describe --tags --always 2>/dev/null || echo "unknown")
Files:
- $(basename $db_backup)
- $(basename $config_backup)
- data.tar.gz (if exists)
EOF
    
    # Create final backup archive
    tar -czf "$full_backup" -C "$temp_dir" .
    
    log_info "Full backup created: $full_backup"
    
    # Cleanup temporary files
    rm -f "$db_backup" "$config_backup"
    
    echo "$full_backup"
}

# List backups
list_backups() {
    log_info "Available backups:"
    echo ""
    
    if [ -d "$BACKUP_DIR" ]; then
        echo "Database Backups:"
        ls -lh $BACKUP_DIR/conexus-db-*.gz 2>/dev/null || echo "  No database backups found"
        echo ""
        
        echo "Configuration Backups:"
        ls -lh $BACKUP_DIR/config-*.tar.gz 2>/dev/null || echo "  No configuration backups found"
        echo ""
        
        echo "Full Backups:"
        ls -lh $BACKUP_DIR/conexus-full-*.tar.gz 2>/dev/null || echo "  No full backups found"
        echo ""
    else
        echo "No backup directory found"
    fi
}

# Restore database
restore_database() {
    local backup_file="$1"
    
    if [ -z "$backup_file" ]; then
        log_error "Please specify a backup file to restore"
        return 1
    fi
    
    if [ ! -f "$backup_file" ]; then
        log_error "Backup file not found: $backup_file"
        return 1
    fi
    
    log_warn "This will replace the current database. Continue? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        log_info "Restore cancelled"
        return 0
    fi
    
    # Stop services
    log_info "Stopping services..."
    docker compose -f $COMPOSE_FILE down || true
    
    # Create backup of current database
    if [ -f "$DATA_DIR/$DB_FILE" ]; then
        local current_backup="$BACKUP_DIR/conexus-db-before-restore-$(date +%Y%m%d-%H%M%S).sqlite"
        cp "$DATA_DIR/$DB_FILE" "$current_backup"
        log_info "Current database backed up to: $current_backup"
    fi
    
    # Extract and restore
    if [[ "$backup_file" == *.gz ]]; then
        gunzip -c "$backup_file" > "$DATA_DIR/$DB_FILE"
    else
        cp "$backup_file" "$DATA_DIR/$DB_FILE"
    fi
    
    # Set permissions
    chmod 644 "$DATA_DIR/$DB_FILE"
    
    # Start services
    log_info "Starting services..."
    docker compose -f $COMPOSE_FILE up -d
    
    # Wait for services to be healthy
    log_info "Waiting for services to be healthy..."
    sleep 10
    
    if curl -sf http://localhost:8080/health > /dev/null; then
        log_info "Database restored successfully"
    else
        log_error "Service health check failed after restore"
        return 1
    fi
}

# Restore full backup
restore_full() {
    local backup_file="$1"
    
    if [ -z "$backup_file" ]; then
        log_error "Please specify a full backup file to restore"
        return 1
    fi
    
    if [ ! -f "$backup_file" ]; then
        log_error "Backup file not found: $backup_file"
        return 1
    fi
    
    log_warn "This will restore the entire system. Continue? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        log_info "Restore cancelled"
        return 0
    fi
    
    # Create temporary directory
    local temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT
    
    # Extract backup
    tar -xzf "$backup_file" -C "$temp_dir"
    
    # Stop services
    log_info "Stopping services..."
    docker compose -f $COMPOSE_FILE down || true
    
    # Restore database
    if [ -f "$temp_dir/conexus-db-*.gz" ]; then
        local db_file=$(ls $temp_dir/conexus-db-*.gz)
        gunzip -c "$db_file" > "$DATA_DIR/$DB_FILE"
        log_info "Database restored"
    fi
    
    # Restore data
    if [ -f "$temp_dir/data.tar.gz" ]; then
        tar -xzf "$temp_dir/data.tar.gz" -C "$DATA_DIR"
        log_info "Data directory restored"
    fi
    
    # Restore configuration (optional)
    log_info "Configuration files found in backup. Restore them? (y/N)"
    read -r restore_config
    if [[ "$restore_config" =~ ^[Yy]$ ]]; then
        if [ -f "$temp_dir/config-*.tar.gz" ]; then
            local config_file=$(ls $temp_dir/config-*.tar.gz)
            tar -xzf "$config_file" -C ./
            log_info "Configuration restored"
        fi
    fi
    
    # Start services
    log_info "Starting services..."
    docker compose -f $COMPOSE_FILE up -d
    
    # Wait for services to be healthy
    log_info "Waiting for services to be healthy..."
    sleep 15
    
    if curl -sf http://localhost:8080/health > /dev/null; then
        log_info "Full restore completed successfully"
    else
        log_error "Service health check failed after restore"
        return 1
    fi
}

# Cleanup old backups
cleanup_backups() {
    log_info "Cleaning up old backups (retention: $RETENTION_DAYS days)..."
    
    # Clean up database backups
    find $BACKUP_DIR -name "conexus-db-*.gz" -mtime +$RETENTION_DAYS -delete 2>/dev/null || true
    
    # Clean up configuration backups
    find $BACKUP_DIR -name "config-*.tar.gz" -mtime +$RETENTION_DAYS -delete 2>/dev/null || true
    
    # Clean up full backups
    find $BACKUP_DIR -name "conexus-full-*.tar.gz" -mtime +$RETENTION_DAYS -delete 2>/dev/null || true
    
    log_info "Backup cleanup completed"
}

# Verify backup integrity
verify_backup() {
    local backup_file="$1"
    
    if [ -z "$backup_file" ]; then
        log_error "Please specify a backup file to verify"
        return 1
    fi
    
    if [ ! -f "$backup_file" ]; then
        log_error "Backup file not found: $backup_file"
        return 1
    fi
    
    log_info "Verifying backup integrity: $backup_file"
    
    # Check if file is readable
    if ! gzip -t "$backup_file" 2>/dev/null && ! tar -tf "$backup_file" > /dev/null 2>&1; then
        log_error "Backup file is corrupted or not in expected format"
        return 1
    fi
    
    # For database backups, check SQLite integrity
    if [[ "$backup_file" == conexus-db-*.gz ]]; then
        local temp_db=$(mktemp)
        trap "rm -f $temp_db" EXIT
        
        gunzip -c "$backup_file" > "$temp_db"
        
        if sqlite3 "$temp_db" "PRAGMA integrity_check;" | grep -q "ok"; then
            log_info "Database backup integrity check passed"
        else
            log_error "Database backup integrity check failed"
            return 1
        fi
    fi
    
    log_info "Backup verification completed successfully"
}

# Schedule automatic backups
schedule_backups() {
    log_info "Setting up automatic backup schedule..."
    
    # Create cron job
    local cron_entry="0 2 * * * cd $(pwd) && ./scripts/backup.sh auto >> ./logs/backup.log 2>&1"
    
    # Add to crontab
    (crontab -l 2>/dev/null | grep -v "backup.sh"; echo "$cron_entry") | crontab -
    
    log_info "Automatic backup scheduled for daily at 2:00 AM"
    log_info "Cron entry: $cron_entry"
}

# Automatic backup (for cron)
auto_backup() {
    log_info "Starting automatic backup..."
    
    setup_backup_dir
    backup_full
    cleanup_backups
    
    log_info "Automatic backup completed"
}

# Main execution
main() {
    echo -e "${BLUE}Conexus Backup Tool${NC}"
    echo "===================="
    
    # Create backup directory
    setup_backup_dir
    
    case "${1:-help}" in
        "db")
            backup_database
            ;;
        "config")
            backup_config
            ;;
        "full")
            backup_full
            ;;
        "list")
            list_backups
            ;;
        "restore-db")
            restore_database "$2"
            ;;
        "restore-full")
            restore_full "$2"
            ;;
        "cleanup")
            cleanup_backups
            ;;
        "verify")
            verify_backup "$2"
            ;;
        "schedule")
            schedule_backups
            ;;
        "auto")
            auto_backup
            ;;
        "help"|*)
            echo "Usage: $0 {db|config|full|list|restore-db|restore-full|cleanup|verify|schedule|auto|help}"
            echo ""
            echo "Commands:"
            echo "  db           - Backup database only"
            echo "  config       - Backup configuration files"
            echo "  full         - Create full backup (database + config + data)"
            echo "  list         - List available backups"
            echo "  restore-db   - Restore database from backup"
            echo "  restore-full - Restore full system from backup"
            echo "  cleanup      - Remove old backups (retention: $RETENTION_DAYS days)"
            echo "  verify       - Verify backup integrity"
            echo "  schedule     - Schedule automatic daily backups"
            echo "  auto         - Automatic backup (for cron jobs)"
            echo "  help         - Show this help"
            echo ""
            echo "Examples:"
            echo "  $0 full                    # Create full backup"
            echo "  $0 restore-db backup.gz    # Restore database"
            echo "  $0 verify backup.gz        # Verify backup integrity"
            echo "  $0 schedule                 # Schedule daily backups"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"