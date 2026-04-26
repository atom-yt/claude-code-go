# Backend Dockerfile for atom-ai-platform
# Multi-stage build for optimized image size

# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /workspace

# Copy entire project for replace dependencies
COPY . ./

WORKDIR /workspace/backend

# Set GOPROXY to use Chinese mirror
ENV GOPROXY=https://goproxy.cn,direct

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /workspace/backend/server .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the server
CMD ["./server"]