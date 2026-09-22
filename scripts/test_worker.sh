#!/bin/bash

# Test script for the worker
# Usage: ./scripts/test_worker.sh [command]
#
# Commands:
#   run-once    - Run notification check once
#   worker      - Run worker continuously
#   combined    - Run server + worker (default Docker mode)
#   create-test - Create a test payment record
#   list        - List pending payments
#   clear-lock  - Clear the notification lock in Redis
#   status      - Get worker status via API
#   test-api    - Test notification via API (dry-run)
#   help        - Show this help

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

# Load environment variables
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

case "${1:-help}" in
    run-once)
        print_header "Running Notification Check Once"
        print_info "This will check for due payments and send notifications"
        echo ""
        go run main.go -notifications-once
        print_success "Notification check completed"
        ;;
    
    worker)
        print_header "Starting Worker (Continuous Mode)"
        print_info "Worker will check for payments every NOTIFICATION_INTERVAL_MINUTES"
        print_info "Press Ctrl+C to stop"
        echo ""
        go run main.go -worker
        ;;
    
    combined)
        print_header "Starting Server + Worker (Docker Mode)"
        print_info "Both HTTP server and worker will run together"
        print_info "Press Ctrl+C to stop"
        echo ""
        go run main.go
        ;;
    
    create-test)
        print_header "Creating Test Payment"
        
        # Default values
        TITLE="${2:-Test Payment}"
        AMOUNT="${3:-100.00}"
        DAYS="${4:-1}"
        
        # Calculate due date (macOS and Linux compatible)
        if [[ "$OSTYPE" == "darwin"* ]]; then
            DUE_DATE=$(date -u -v+${DAYS}d +"%Y-%m-%dT%H:%M:%SZ")
        else
            DUE_DATE=$(date -u -d "+${DAYS} days" +"%Y-%m-%dT%H:%M:%SZ")
        fi
        
        print_info "Title: $TITLE"
        print_info "Amount: $AMOUNT"
        print_info "Due Date: $DUE_DATE"
        echo ""
        
        # Create payment via API
        RESPONSE=$(curl -s -X POST http://localhost:8080/payments \
            -H "Content-Type: application/json" \
            -d "{
                \"title\": \"$TITLE\",
                \"amount\": $AMOUNT,
                \"currency\": \"USD\",
                \"payment_type\": \"one_time\",
                \"due_date\": \"$DUE_DATE\",
                \"recipient_name\": \"Test Recipient\",
                \"recipient_email\": \"test@example.com\",
                \"notification_channels\": [\"email\"]
            }")
        
        if echo "$RESPONSE" | grep -q '"id"'; then
            print_success "Payment created successfully"
            echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
        else
            print_error "Failed to create payment"
            echo "$RESPONSE"
        fi
        ;;
    
    list)
        print_header "Listing Pending Payments"
        echo ""
        curl -s http://localhost:8080/payments | python3 -m json.tool 2>/dev/null || curl -s http://localhost:8080/payments
        echo ""
        ;;
    
    clear-lock)
        print_header "Clearing Notification Lock"
        print_info "Removing distributed lock from Redis"
        
        # Build Redis URL
        LOCK_KEY="${REDIS_PREFIX}:lock:payment-notifications"
        
        if command -v redis-cli &> /dev/null; then
            # Parse Redis URI
            REDIS_HOST=$(echo $REDIS_URI | cut -d':' -f1)
            REDIS_PORT=$(echo $REDIS_URI | cut -d':' -f2)
            
            if [ -n "$REDIS_PASSWORD" ]; then
                redis-cli -h $REDIS_HOST -p $REDIS_PORT -a "$REDIS_PASSWORD" --no-auth-warning DEL "$LOCK_KEY" 2>/dev/null
            else
                redis-cli -h $REDIS_HOST -p $REDIS_PORT DEL "$LOCK_KEY" 2>/dev/null
            fi
            print_success "Lock cleared: $LOCK_KEY"
        else
            print_error "redis-cli not found"
            print_info "Install redis-tools or manually delete key: $LOCK_KEY"
        fi
        ;;
    
    status)
        print_header "Worker Status (via API)"
        echo ""
        curl -s http://localhost:8080/notifications/status | python3 -m json.tool 2>/dev/null || \
            curl -s http://localhost:8080/notifications/status
        echo ""
        ;;
    
    test-api)
        print_header "Testing Notification via API (Dry-Run)"
        FORCE="${2:-false}"
        echo ""
        curl -s -X POST http://localhost:8080/notifications/test \
            -H "Content-Type: application/json" \
            -d "{\"dry_run\": true, \"force\": $FORCE}" | python3 -m json.tool 2>/dev/null || \
            curl -s -X POST http://localhost:8080/notifications/test \
                -H "Content-Type: application/json" \
                -d "{\"dry_run\": true, \"force\": $FORCE}"
        echo ""
        ;;
    
    trigger)
        print_header "Trigger Notifications (Actually Send)"
        print_info "This will actually send notifications!"
        echo ""
        read -p "Are you sure? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            curl -s -X POST http://localhost:8080/notifications/trigger \
                -H "Content-Type: application/json" | python3 -m json.tool 2>/dev/null || \
                curl -s -X POST http://localhost:8080/notifications/trigger
        else
            print_info "Aborted"
        fi
        ;;
    
    test-notification)
        print_header "Full Notification Test"
        
        print_info "Step 1: Checking worker status..."
        echo ""
        curl -s http://localhost:8080/notifications/status | python3 -m json.tool
        
        echo ""
        print_info "Step 2: Dry-run notification test..."
        echo ""
        curl -s -X POST http://localhost:8080/notifications/test \
            -H "Content-Type: application/json" \
            -d '{"dry_run": true}' | python3 -m json.tool
        
        echo ""
        print_info "Step 3: Listing pending payments..."
        echo ""
        curl -s http://localhost:8080/payments | python3 -m json.tool
        ;;
    
    docker-up)
        print_header "Starting Docker Environment"
        print_info "Starting MongoDB, Redis, and API with Worker"
        echo ""
        docker-compose up -d
        print_success "Services started"
        print_info "Logs: docker-compose logs -f api"
        ;;
    
    docker-down)
        print_header "Stopping Docker Environment"
        docker-compose down
        print_success "Services stopped"
        ;;
    
    docker-logs)
        docker-compose logs -f api
        ;;
    
    help|*)
        echo ""
        echo "Worker Test Script"
        echo ""
        echo -e "${YELLOW}Usage:${NC} ./scripts/test_worker.sh [command]"
        echo ""
        echo -e "${YELLOW}Commands:${NC}"
        echo "  run-once        Run notification check once (good for cron)"
        echo "  worker          Run worker continuously"
        echo "  combined        Run server + worker (Docker mode)"
        echo "  status          Get worker status via API"
        echo "  test-api        Test notification via API (dry-run) [force=true|false]"
        echo "  trigger         Actually trigger notifications (requires auth)"
        echo "  create-test     Create a test payment [title] [amount] [days]"
        echo "  list            List pending payments"
        echo "  clear-lock      Clear the notification lock in Redis"
        echo "  test-notification  Full test: status + dry-run + list"
        echo "  docker-up       Start Docker services"
        echo "  docker-down     Stop Docker services"
        echo "  docker-logs     View API logs"
        echo "  help            Show this help"
        echo ""
        echo -e "${YELLOW}API Endpoints:${NC}"
        echo "  GET  /notifications/status  - Get worker status"
        echo "  POST /notifications/test   - Dry-run notification test"
        echo "  POST /notifications/trigger - Actually send (auth required)"
        echo ""
        echo -e "${YELLOW}Examples:${NC}"
        echo "  ./scripts/test_worker.sh status"
        echo "  ./scripts/test_worker.sh test-api"
        echo "  ./scripts/test_worker.sh test-api true    # Force already-sent"
        echo "  ./scripts/test_worker.sh create-test 'Rent' 1500 3"
        echo "  curl -s http://localhost:8080/notifications/status | jq ."
        echo ""
        ;;
esac
