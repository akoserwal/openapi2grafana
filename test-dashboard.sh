#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
OPENAPI_FILE="sample-openapi.yaml"  # Use sample spec for demo
DASHBOARD_FILE="output_dashboard.json"
BACKUP_DIR="backups"
COMPOSE_FILE="docker-compose.yaml"

# Print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if files exist
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    if [[ ! -f "$OPENAPI_FILE" ]]; then
        print_error "OpenAPI file '$OPENAPI_FILE' not found!"
        exit 1
    fi
    
    if [[ ! -f "$COMPOSE_FILE" ]]; then
        print_error "Docker Compose file '$COMPOSE_FILE' not found!"
        exit 1
    fi
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH!"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed or not in PATH!"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH!"
        exit 1
    fi
    
    print_info "All prerequisites met!"
}

# Backup existing dashboard
backup_dashboard() {
    if [[ -f "$DASHBOARD_FILE" ]]; then
        mkdir -p "$BACKUP_DIR"
        BACKUP_FILE="$BACKUP_DIR/dashboard_$(date +%Y%m%d_%H%M%S).json"
        cp "$DASHBOARD_FILE" "$BACKUP_FILE"
        print_info "Backed up existing dashboard to $BACKUP_FILE"
    fi
}

# Generate the dashboard
generate_dashboard() {
    print_info "Generating Grafana dashboard from OpenAPI spec..."
    
    # Check if we should update or create new
    if [[ -f "$DASHBOARD_FILE" ]]; then
        print_info "Updating existing dashboard..."
        go run main.go "$OPENAPI_FILE" "$DASHBOARD_FILE" --update --datasource prometheus
    else
        print_info "Creating new dashboard..."
        go run main.go "$OPENAPI_FILE" "$DASHBOARD_FILE" --datasource prometheus
    fi
    
    if [[ $? -eq 0 ]]; then
        print_info "Dashboard generated successfully!"
    else
        print_error "Failed to generate dashboard!"
        exit 1
    fi
}

# Validate the generated dashboard
validate_dashboard() {
    print_info "Validating generated dashboard..."
    
    if [[ ! -f "$DASHBOARD_FILE" ]]; then
        print_error "Dashboard file not found!"
        exit 1
    fi
    
    # Basic JSON validation
    if ! jq empty "$DASHBOARD_FILE" 2>/dev/null; then
        print_error "Generated dashboard is not valid JSON!"
        exit 1
    fi
    
    # Check for required fields
    if ! jq -e '.title' "$DASHBOARD_FILE" > /dev/null; then
        print_error "Dashboard missing required 'title' field!"
        exit 1
    fi
    
    if ! jq -e '.panels | length > 0' "$DASHBOARD_FILE" > /dev/null; then
        print_error "Dashboard has no panels!"
        exit 1
    fi
    
    PANEL_COUNT=$(jq '.panels | length' "$DASHBOARD_FILE")
    print_info "Dashboard validation passed! Found $PANEL_COUNT panels."
}

# Start the monitoring stack
start_monitoring() {
    print_info "Starting monitoring stack..."
    
    # Stop any existing containers
    docker-compose down --remove-orphans 2>/dev/null || true
    
    # Start the services
    docker-compose up -d
    
    if [[ $? -eq 0 ]]; then
        print_info "Monitoring stack started successfully!"
    else
        print_error "Failed to start monitoring stack!"
        exit 1
    fi
}

# Wait for services to be ready
wait_for_services() {
    print_info "Waiting for services to be ready..."
    
    # Wait for Prometheus
    print_info "Waiting for Prometheus..."
    timeout 60 bash -c 'until curl -s http://localhost:9090/-/ready > /dev/null 2>&1; do sleep 2; done' || {
        print_error "Prometheus failed to start!"
        exit 1
    }
    
    # Wait for Grafana
    print_info "Waiting for Grafana..."
    timeout 60 bash -c 'until curl -s http://localhost:3000/api/health > /dev/null 2>&1; do sleep 2; done' || {
        print_error "Grafana failed to start!"
        exit 1
    }
    
    print_info "All services are ready!"
}

# Import dashboard into Grafana
import_dashboard() {
    print_info "Importing dashboard into Grafana..."
    
    # Wait a bit more for Grafana to fully initialize
    sleep 5
    
    # Import the dashboard
    DASHBOARD_RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d @"$DASHBOARD_FILE" \
        http://admin:admin@localhost:3000/api/dashboards/db)
    
    if echo "$DASHBOARD_RESPONSE" | jq -e '.status == "success"' > /dev/null 2>&1; then
        DASHBOARD_URL=$(echo "$DASHBOARD_RESPONSE" | jq -r '.url')
        print_info "Dashboard imported successfully!"
        print_info "Dashboard URL: http://localhost:3000$DASHBOARD_URL"
    else
        print_warning "Dashboard import may have failed. Check Grafana logs."
        print_info "Dashboard should be available at: http://localhost:3000"
    fi
}

# Show service URLs
show_urls() {
    print_info "Service URLs:"
    echo "  Grafana:    http://localhost:3000 (admin/admin)"
    echo "  Prometheus: http://localhost:9090"
    echo "  AlertManager: http://localhost:9093"
    echo ""
    print_info "Dashboard file: $DASHBOARD_FILE"
}

# Show logs
show_logs() {
    if [[ "$1" == "--logs" ]]; then
        print_info "Showing container logs..."
        docker-compose logs -f
    fi
}

# Main execution
main() {
    print_info "Starting Grafana Dashboard Generator Test..."
    
    check_prerequisites
    backup_dashboard
    generate_dashboard
    validate_dashboard
    start_monitoring
    wait_for_services
    import_dashboard
    show_urls
    
    print_info "Test completed successfully!"
    print_info "Access Grafana at http://localhost:3000 with admin/admin"
    
    # Show logs if requested
    show_logs "$1"
}

# Cleanup function
cleanup() {
    if [[ "$1" == "--cleanup" ]]; then
        print_info "Cleaning up..."
        docker-compose down --remove-orphans -v
        print_info "Cleanup completed!"
        exit 0
    fi
}

# Handle arguments
if [[ "$1" == "--help" ]]; then
    echo "Usage: $0 [--logs] [--cleanup] [--help]"
    echo "  --logs:    Show container logs after startup"
    echo "  --cleanup: Stop and remove all containers and volumes"
    echo "  --help:    Show this help message"
    exit 0
fi

cleanup "$1"
main "$1"