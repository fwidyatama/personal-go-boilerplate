# Microservice Boilerplate

A production-ready microservice boilerplate built with Go, following Clean Architecture principles and featuring event-driven architecture with Kafka, Redis caching, and PostgreSQL database.

## 🚀 Features

- **Clean Architecture** - Domain, UseCase, Repository, Handler layers
- **Event-Driven Architecture** - Kafka producer/consumer with event handling
- **Caching** - Redis integration with TTL and cache invalidation
- **Database** - PostgreSQL with RAW SQL queries (no ORM)
- **Migrations** - Goose migration system with up/down support
- **Configuration** - Environment-based configuration with Viper
- **Logging** - Structured JSON logging with Logrus, context-aware, and request_id propagation
- **Health Checks** - Database, Redis, and Kafka connectivity monitoring
- **Docker** - Multi-stage builds with security best practices
- **Testing** - Unit tests with mocks and integration tests
- **API Documentation** - Swagger/OpenAPI annotations
- **Monitoring** - Prometheus metrics endpoint
- **Graceful Shutdown** - Proper cleanup and resource management

## 📋 Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL (for local development)
- Redis (for local development)
- Kafka (for local development)

## 🛠️ Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd go-boilerplate
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment**
   ```bash
   cp env.example .env
   # Edit .env file with your configuration
   ```

4. **Generate .env from config.yaml (optional, recommended for Docker Compose sync)**
   ```bash
   ./scripts/generate_env.sh
   # .env will be generated from config.yaml
   ```

5. **Install development tools (optional)**
   ```bash
   make install-tools
   ```

## 🚀 Quick Start

### Option 1: Docker Compose (Recommended)

Start all services with Docker Compose:

```bash
# Generate .env from config.yaml (if not already)
./scripts/generate_env.sh

make docker-compose-up
```

This will start:
- PostgreSQL database
- Redis cache
- Kafka with Zookeeper
- Kafka UI (http://localhost:8081)
- Application (http://localhost:8080)

### Option 2: Local Development

1. **Start dependencies**
   ```bash
   docker-compose up -d postgres redis zookeeper kafka
   ```

2. **Run migrations**
   ```bash
   ./scripts/migrate.sh up
   ```

3. **Start application**
   ```bash
   make dev
   # or
   go run cmd/main.go
   ```

## 📚 API Documentation

Once the application is running, you can access:

- **API Base URL**: http://localhost:8080/api/v1
- **Health Check**: http://localhost:8080/api/v1/health
- **Metrics**: http://localhost:8080/api/v1/metrics
- **Swagger UI**: http://localhost:8080/swagger/

### Example API Calls

```bash
# Create a user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'

# Get all users
curl http://localhost:8080/api/v1/users

# Get user by ID
curl http://localhost:8080/api/v1/users/1

# Update user
curl -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Updated",
    "email": "john.updated@example.com",
    "password": "newpassword123"
  }'

# Delete user
curl -X DELETE http://localhost:8080/api/v1/users/1
```

## 🗄️ Database Migrations

The project uses Goose for database migrations:

```bash
# Run migrations
./scripts/migrate.sh up

# Check migration status
./scripts/migrate.sh status

# Rollback last migration
./scripts/migrate.sh down

# Create new migration
./scripts/migrate.sh create add_new_table

# Reset database (WARNING: deletes all data)
./scripts/migrate.sh reset
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run benchmarks
make bench

# Run integration tests
go test -tags=integration ./...
```

## 🐳 Docker

### Build Docker Image

```bash
# Generate .env from config.yaml (if not already)
./scripts/generate_env.sh

# Build image
make docker-build

# Run container
make docker-run

# Build and run with Docker Compose
make docker-compose-up
```

### Production Deployment

```bash
# Build for production
docker build -t go-boilerplate:latest .

# Run with production environment
docker run -p 8080:8080 \
  -e APP_ENV=production \
  -e DB_HOST=your-db-host \
  -e REDIS_HOST=your-redis-host \
  go-boilerplate:latest
```

## 📁 Project Structure

```
go-boilerplate/
├── cmd/
│   └── main.go                 # Application entry point, manual dependency injection
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── domain/
│   │   └── user.go             # Domain entities and interfaces
│   ├── handler/
│   │   ├── user_handler.go     # HTTP handlers
│   │   ├── health_handler.go   # Health check handlers
│   │   └── auth_handler.go     # Auth handlers
│   ├── middleware/
│   │   └── request_id.go       # Gin middleware for request_id propagation
│   ├── repository/
│   │   └── user_repository.go  # Data access layer
│   └── usecase/
│       └── user_usecase.go     # Business logic layer
├── pkg/
│   ├── database/
│   │   └── database.go         # Database connection management
│   └── logger/
│       └── logger.go           # Context-aware JSON logging
├── migrations/
│   ├── 00001_create_users_table.sql
│   ├── 00002_add_user_indexes.sql
│   └── embed.go
├── scripts/
│   ├── migrate.sh              # Database migration script
├── swagger-ui/
│   └── ...                     # Swagger UI static files
├── config.yaml                 # Main configuration file
├── Dockerfile                  # Docker build file
├── Makefile                    # Build and dev scripts
└── README.md                   # This file
```

## 📝 Architecture Notes

- **Manual Dependency Injection**: Semua dependency (config, DB, logger, service, usecase, handler, Gin, middleware) diinisialisasi dan dihubungkan secara eksplisit di `main.go`. Tidak ada lagi penggunaan wire/go-cloud/wire.
- **Logger**: Semua log aplikasi (handler, usecase, repo) menggunakan context-aware logger yang otomatis menyertakan `request_id` dan output JSON. Log HTTP request juga sudah JSON.
- **Middleware**: Middleware request_id akan meng-inject request_id ke context dan logger di setiap request.

## 🔒 Security

- **Non-root Docker container** for security
- **Environment-based configuration** for secrets
- **Input validation** on all endpoints
- **SQL injection protection** with parameterized queries
- **HTTPS support** (configure in production)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run tests and ensure they pass
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:
1. Check the documentation
2. Search existing issues
3. Create a new issue with detailed information 