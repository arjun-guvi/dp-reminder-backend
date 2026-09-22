# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application (single binary that runs both server and worker)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/server ./main.go

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/server .

# Create non-root user
RUN adduser -D -g '' appuser

# Create log directory and set permissions
RUN mkdir -p /var/log/app && chown -R appuser:appuser /var/log/app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Set environment defaults
ENV PORT=8080
ENV GIN_MODE=release

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary (which now runs both server and worker as goroutines)
ENTRYPOINT ["/app/server"]
