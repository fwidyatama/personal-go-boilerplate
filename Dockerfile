# Build stage
FROM golang:1.24.4-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Final stage
FROM alpine:latest

# Install ca-certificates, netcat, and wget for HTTPS requests and health checks
RUN apk --no-cache add ca-certificates tzdata netcat-openbsd wget

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy migrations if they exist
COPY --from=builder /app/migrations ./migrations/

# Copy Swagger UI static files for API documentation (if present)
COPY --from=builder /app/swagger-ui ./swagger-ui/

# Copy OpenAPI YAML file for Swagger UI (if present, from docs directory)
COPY --from=builder /app/docs/openapi.yaml ./docs/openapi.yaml

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Run the application
CMD ["./main"]