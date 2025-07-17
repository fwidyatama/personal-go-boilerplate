#!/bin/bash

# Application run script
# Usage: ./scripts/run.sh [dev|prod|docker]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Microservice Boilerplate${NC}"
    echo -e "${BLUE}================================${NC}"
}

# Function to check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go first."
        exit 1
    fi
}

# Function to check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
}

# Function to check if Docker Compose is installed
check_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
}

# Function to check if .env file exists
check_env_file() {
    if [ ! -f .env ]; then
        print_warning ".env file not found. Creating from example..."
        if [ -f env.example ]; then
            cp env.example .env
            print_status ".env file created from env.example"
        else
            print_error "env.example file not found. Please create a .env file manually."
            exit 1
        fi
    fi
}

# Function to run in development mode
run_dev() {
    print_status "Starting in development mode..."
    check_go
    check_env_file
    
    # Install dependencies
    print_status "Installing dependencies..."
    go mod download
    
    # Run with hot reload if air is available
    if command -v air &> /dev/null; then
        print_status "Starting with hot reload (air)..."
        air
    else
        print_warning "air not found. Installing air for hot reload..."
        go install github.com/air-verse/air@latest
        print_status "Starting with hot reload (air)..."
        air
    fi
}

# Function to run in production mode
run_prod() {
    print_status "Starting in production mode..."
    check_go
    check_env_file
    
    # Build the application
    print_status "Building application..."
    go build -o main ./cmd
    
    # Run the application
    print_status "Starting application..."
    ./main
}

# Function to run with Docker
run_docker() {
    print_status "Starting with Docker..."
    check_docker
    check_docker_compose
    check_env_file
    
    # Build and start all services
    print_status "Building and starting all services..."
    docker-compose up --build -d
    
    print_status "Services started successfully!"
    print_status "Application: http://localhost:8080"
    print_status "Health check: http://localhost:8080/api/v1/health"
    print_status "Kafka UI: http://localhost:8081"
    
    # Show logs
    print_status "Showing logs (Ctrl+C to stop)..."
    docker-compose logs -f
}

# Function to stop Docker services
stop_docker() {
    print_status "Stopping Docker services..."
    docker-compose down
    print_status "Services stopped successfully"
}

# Function to show status
show_status() {
    print_status "Checking service status..."
    
    # Check if application is running
    if curl -s http://localhost:8080/api/v1/health &> /dev/null; then
        print_status "Application is running: http://localhost:8080"
    else
        print_warning "Application is not running"
    fi
    
    # Check Docker services
    if command -v docker-compose &> /dev/null; then
        print_status "Docker services status:"
        docker-compose ps
    fi
}

# Main script logic
case "$1" in
    "dev")
        print_header
        run_dev
        ;;
    "prod")
        print_header
        run_prod
        ;;
    "docker")
        print_header
        run_docker
        ;;
    "stop")
        print_header
        stop_docker
        ;;
    "status")
        print_header
        show_status
        ;;
    "build")
        print_header
        print_status "Building application..."
        check_go
        go build -o main ./cmd
        print_status "Build completed successfully"
        ;;
    "test")
        print_header
        print_status "Running tests..."
        check_go
        go test -v ./...
        ;;
    "clean")
        print_header
        print_status "Cleaning build artifacts..."
        go clean
        rm -f main
        print_status "Clean completed"
        ;;
    *)
        print_header
        echo "Usage: $0 {dev|prod|docker|stop|status|build|test|clean}"
        echo ""
        echo "Commands:"
        echo "  dev     - Run in development mode with hot reload"
        echo "  prod    - Run in production mode"
        echo "  docker  - Run with Docker Compose (all services)"
        echo "  stop    - Stop Docker services"
        echo "  status  - Show service status"
        echo "  build   - Build the application"
        echo "  test    - Run tests"
        echo "  clean   - Clean build artifacts"
        echo ""
        echo "Examples:"
        echo "  $0 dev      # Start development server"
        echo "  $0 docker   # Start all services with Docker"
        echo "  $0 status   # Check service status"
        exit 1
        ;;
esac 