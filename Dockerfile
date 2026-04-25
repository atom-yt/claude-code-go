# Gateway Dockerfile for claude-code-go
# Multi-stage build for optimized image size

# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build gateway binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o gateway ./cmd/gateway

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/gateway .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \n    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the gateway
CMD ["./gateway"]