#!/bin/bash

# Database migration script
# Usage: ./scripts/migrate.sh [up|down|status|create]

set -e

# Configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-microservice_db}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-password}
DB_SSL_MODE=${DB_SSL_MODE:-disable}

# Connection string
DB_URL="host=$DB_HOST port=$DB_PORT user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME sslmode=$DB_SSL_MODE"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if goose is installed
check_goose() {
    if ! command -v goose &> /dev/null; then
        print_error "goose is not installed. Please install it first:"
        echo "go install github.com/pressly/goose/v3/cmd/goose@latest"
        exit 1
    fi
}

# Function to check database connectivity
check_db_connection() {
    print_status "Checking database connection..."
    if ! goose -dir migrations postgres "$DB_URL" status &> /dev/null; then
        print_error "Cannot connect to database. Please check your configuration."
        exit 1
    fi
    print_status "Database connection successful"
}

# Main script logic
case "$1" in
    "up")
        print_status "Running database migrations up..."
        check_goose
        check_db_connection
        goose -dir migrations postgres "$DB_URL" up
        print_status "Migrations completed successfully"
        ;;
    "down")
        print_warning "Rolling back database migrations..."
        check_goose
        check_db_connection
        read -p "Are you sure you want to rollback migrations? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            goose -dir migrations postgres "$DB_URL" down
            print_status "Migrations rolled back successfully"
        else
            print_status "Migration rollback cancelled"
        fi
        ;;
    "status")
        print_status "Checking migration status..."
        check_goose
        check_db_connection
        goose -dir migrations postgres "$DB_URL" status
        ;;
    "create")
        if [ -z "$2" ]; then
            print_error "Migration name is required"
            echo "Usage: $0 create <migration_name>"
            exit 1
        fi
        print_status "Creating new migration: $2"
        check_goose
        goose -dir migrations create "$2" sql
        print_status "Migration file created successfully"
        ;;
    "reset")
        print_warning "Resetting database (this will drop all data)..."
        check_goose
        check_db_connection
        read -p "Are you sure you want to reset the database? This will delete all data! (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            goose -dir migrations postgres "$DB_URL" reset
            print_status "Database reset successfully"
        else
            print_status "Database reset cancelled"
        fi
        ;;
    *)
        echo "Usage: $0 {up|down|status|create|reset}"
        echo ""
        echo "Commands:"
        echo "  up      - Run all pending migrations"
        echo "  down    - Rollback the last migration"
        echo "  status  - Show migration status"
        echo "  create  - Create a new migration file"
        echo "  reset   - Reset database (drop all data)"
        echo ""
        echo "Environment variables:"
        echo "  DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD, DB_SSL_MODE"
        exit 1
        ;;
esac 