# Build stage
FROM golang:1.27-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Tidy up the dependencies
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Pin Alpine and bump OpenSSL past CVE-2026-14456 (fixed in 3.5.8-r0)
FROM alpine:3.22

RUN apk upgrade --no-cache \
    && apk add --no-cache \
        ca-certificates \
        'libcrypto3>=3.5.8-r0' \
        'libssl3>=3.5.8-r0'

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Change ownership to appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/v1/ || exit 1

# Run the application
CMD ["./main"]
