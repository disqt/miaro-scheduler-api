#!/bin/bash
set -e

# Miaro Scheduler API - Build and Deploy Script
# Usage: ./scripts/deploy.sh [build|deploy|restart|stop|logs|status]

SERVICE_NAME="miaro-scheduler-api"
BINARY_NAME="main"
INSTALL_DIR="/opt/miaro-scheduler-api"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Run tests
run_tests() {
    log_info "Running tests..."
    go test ./... -v
    log_info "Tests passed!"
}

# Build the Go binary
build() {
    log_info "Building Go binary..."
    CGO_ENABLED=0 go build -o $BINARY_NAME .
    log_info "Binary built: ./$BINARY_NAME"
}

# Deploy to server (requires sudo)
deploy() {
    run_tests
    build

    log_info "Stopping service..."
    sudo systemctl stop $SERVICE_NAME || true

    log_info "Copying binary to $INSTALL_DIR..."
    sudo mkdir -p $INSTALL_DIR
    sudo cp $BINARY_NAME $INSTALL_DIR/
    sudo cp -r templates $INSTALL_DIR/

    log_info "Starting service..."
    sudo systemctl start $SERVICE_NAME
    sudo systemctl enable $SERVICE_NAME

    log_info "Deployment complete!"
    show_status
}

# Restart service
restart() {
    log_info "Restarting $SERVICE_NAME..."
    sudo systemctl restart $SERVICE_NAME
    log_info "Service restarted."
    show_status
}

# Stop service
stop() {
    log_info "Stopping $SERVICE_NAME..."
    sudo systemctl stop $SERVICE_NAME
    log_info "Service stopped."
}

# Show logs
show_logs() {
    sudo journalctl -u $SERVICE_NAME -f
}

# Show status
show_status() {
    echo ""
    log_info "Service status:"
    sudo systemctl status $SERVICE_NAME --no-pager || true
    echo ""
    log_info "Health check:"
    curl -s http://localhost:8081/health 2>/dev/null && echo "" || log_warn "Service not responding"
}

# Show help
show_help() {
    echo "Miaro Scheduler API - Build and Deploy Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  test      Run tests"
    echo "  build     Build binary locally"
    echo "  deploy    Full deploy (test + build + install + restart service)"
    echo "  restart   Restart the systemd service"
    echo "  stop      Stop the systemd service"
    echo "  logs      Show service logs (follow)"
    echo "  status    Show service status and health"
    echo "  help      Show this help message"
    echo ""
    echo "Service: $SERVICE_NAME"
    echo "Install directory: $INSTALL_DIR"
}

# Main
case "${1:-help}" in
    test)
        run_tests
        ;;
    build)
        build
        ;;
    deploy)
        deploy
        ;;
    restart)
        restart
        ;;
    stop)
        stop
        ;;
    logs)
        show_logs
        ;;
    status)
        show_status
        ;;
    help|*)
        show_help
        ;;
esac
