#!/bin/bash

# Conexus Monitoring Script
# This script provides monitoring and health check functionality for Conexus deployments

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
HEALTH_URL="http://localhost:8080/health"
METRICS_URL="http://localhost:8080/metrics"
LOG_FILE="./logs/monitor.log"
ALERT_THRESHOLD_CPU=80
ALERT_THRESHOLD_MEMORY=80
ALERT_THRESHOLD_DISK=90

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] $1" >> $LOG_FILE
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [WARN] $1" >> $LOG_FILE
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] $1" >> $LOG_FILE"
}

# Create log directory
mkdir -p ./logs

# Health check
check_health() {
    log_info "Checking service health..."
    
    if curl -sf $HEALTH_URL > /dev/null; then
        local health_response=$(curl -s $HEALTH_URL)
        local status=$(echo $health_response | jq -r '.status // "unknown"')
        local version=$(echo $health_response | jq -r '.version // "unknown"')
        
        if [ "$status" = "healthy" ]; then
            log_info "Service is healthy (version: $version)"
            return 0
        else
            log_error "Service status: $status"
            return 1
        fi
    else
        log_error "Service is not responding"
        return 1
    fi
}

# Metrics collection
collect_metrics() {
    log_info "Collecting metrics..."
    
    # System metrics
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%us,//')
    local memory_usage=$(free | grep Mem | awk '{printf("%.1f"), $3/$2 * 100.0}')
    local disk_usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
    
    # Docker container metrics (if running)
    local container_metrics=""
    if command -v docker &> /dev/null; then
        local container_id=$(docker ps -q --filter "name=conexus" | head -1)
        if [ ! -z "$container_id" ]; then
            local container_cpu=$(docker stats --no-stream --format "table {{.CPUPerc}}" $container_id | tail -1 | sed 's/%//')
            local container_memory=$(docker stats --no-stream --format "table {{.MemPerc}}" $container_id | tail -1 | sed 's/%//')
            container_metrics="Container CPU: ${container_cpu}%, Container Memory: ${container_memory}%"
        fi
    fi
    
    # Application metrics (if available)
    local app_metrics=""
    if curl -sf $METRICS_URL > /dev/null 2>&1; then
        local request_count=$(curl -s $METRICS_URL | grep "conexus_requests_total" | awk '{sum += $NF} END {print sum}')
        local error_count=$(curl -s $METRICS_URL | grep "conexus_requests_total" | grep "status=\"5\"" | awk '{sum += $NF} END {print sum}')
        app_metrics="Requests: ${request_count:-0}, Errors: ${error_count:-0}"
    fi
    
    # Log metrics
    log_info "System Metrics - CPU: ${cpu_usage}%, Memory: ${memory_usage}%, Disk: ${disk_usage}%"
    if [ ! -z "$container_metrics" ]; then
        log_info "$container_metrics"
    fi
    if [ ! -z "$app_metrics" ]; then
        log_info "$app_metrics"
    fi
    
    # Check thresholds
    if (( $(echo "$cpu_usage > $ALERT_THRESHOLD_CPU" | bc -l) )); then
        log_warn "CPU usage above threshold: ${cpu_usage}%"
    fi
    if (( $(echo "$memory_usage > $ALERT_THRESHOLD_MEMORY" | bc -l) )); then
        log_warn "Memory usage above threshold: ${memory_usage}%"
    fi
    if [ "$disk_usage" -gt "$ALERT_THRESHOLD_DISK" ]; then
        log_warn "Disk usage above threshold: ${disk_usage}%"
    fi
}

# Check logs for errors
check_logs() {
    log_info "Checking for errors in logs..."
    
    if command -v docker &> /dev/null; then
        local error_count=$(docker logs conexus 2>&1 | grep -i error | wc -l)
        local warn_count=$(docker logs conexus 2>&1 | grep -i warn | wc -l)
        
        log_info "Log summary - Errors: $error_count, Warnings: $warn_count"
        
        if [ "$error_count" -gt 0 ]; then
            log_warn "Found $error_count errors in container logs"
            docker logs conexus 2>&1 | grep -i error | tail -5
        fi
    fi
}

# Performance test
performance_test() {
    log_info "Running performance test..."
    
    local start_time=$(date +%s.%N)
    local response_time=$(curl -o /dev/null -s -w '%{time_total}' $HEALTH_URL)
    local end_time=$(date +%s.%N)
    
    local duration=$(echo "$end_time - $start_time" | bc)
    
    log_info "Performance metrics - Response time: ${response_time}s, Total time: ${duration}s"
    
    # Alert on slow response
    if (( $(echo "$response_time > 1.0" | bc -l) )); then
        log_warn "Slow response time detected: ${response_time}s"
    fi
}

# Database health check
check_database() {
    log_info "Checking database health..."
    
    if command -v docker &> /dev/null; then
        local db_size=$(docker exec conexus ls -la /data/conexus.db 2>/dev/null | awk '{print $5}' || echo "0")
        local db_size_mb=$((db_size / 1024 / 1024))
        
        log_info "Database size: ${db_size_mb}MB"
        
        # Check if database is accessible
        if docker exec conexus sqlite3 /data/conexus.db "SELECT 1;" > /dev/null 2>&1; then
            log_info "Database is accessible"
        else
            log_error "Database is not accessible"
        fi
    fi
}

# Generate report
generate_report() {
    log_info "Generating monitoring report..."
    
    local report_file="./reports/monitor-report-$(date +%Y%m%d-%H%M%S).txt"
    mkdir -p ./reports
    
    {
        echo "Conexus Monitoring Report"
        echo "========================"
        echo "Generated: $(date)"
        echo ""
        
        echo "Health Status:"
        check_health
        echo ""
        
        echo "System Metrics:"
        collect_metrics
        echo ""
        
        echo "Database Status:"
        check_database
        echo ""
        
        echo "Performance Test:"
        performance_test
        echo ""
        
        echo "Recent Errors:"
        check_logs
        echo ""
        
    } > $report_file
    
    log_info "Report generated: $report_file"
}

# Continuous monitoring
monitor_continuous() {
    log_info "Starting continuous monitoring (Ctrl+C to stop)..."
    
    while true; do
        echo -e "\n${BLUE}=== $(date) ===${NC}"
        
        check_health
        collect_metrics
        check_database
        
        sleep 60
    done
}

# Alert setup
setup_alerts() {
    log_info "Setting up alert configuration..."
    
    cat > ./alerts.yml << EOF
# Conexus Alert Configuration
alerts:
  - name: "High CPU Usage"
    condition: "cpu_usage > 80"
    severity: "warning"
    action: "log"
  
  - name: "High Memory Usage" 
    condition: "memory_usage > 80"
    severity: "warning"
    action: "log"
    
  - name: "Service Down"
    condition: "health_check != 200"
    severity: "critical"
    action: "restart"
    
  - name: "Disk Space Low"
    condition: "disk_usage > 90"
    severity: "critical"
    action: "log"
EOF
    
    log_info "Alert configuration created: ./alerts.yml"
}

# Main execution
main() {
    echo -e "${BLUE}Conexus Monitoring Tool${NC}"
    echo "========================"
    
    case "${1:-check}" in
        "check")
            check_health
            collect_metrics
            check_database
            performance_test
            ;;
        "continuous")
            monitor_continuous
            ;;
        "report")
            generate_report
            ;;
        "logs")
            check_logs
            ;;
        "setup-alerts")
            setup_alerts
            ;;
        "all")
            check_health
            collect_metrics
            check_database
            performance_test
            check_logs
            generate_report
            ;;
        *)
            echo "Usage: $0 {check|continuous|report|logs|setup-alerts|all}"
            echo ""
            echo "Commands:"
            echo "  check         - Run basic health and metrics checks"
            echo "  continuous    - Run continuous monitoring (every 60s)"
            echo "  report        - Generate detailed monitoring report"
            echo "  logs          - Check for errors in logs"
            echo "  setup-alerts  - Create alert configuration"
            echo "  all           - Run all checks and generate report"
            exit 1
            ;;
    esac
}

# Install dependencies if needed
if ! command -v jq &> /dev/null; then
    log_warn "jq is not installed. Install it for better JSON parsing."
fi

if ! command -v bc &> /dev/null; then
    log_warn "bc is not installed. Install it for floating point calculations."
fi

# Run main function
main "$@"