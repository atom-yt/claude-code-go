# Frontend Dockerfile for atom-ai-platform
# Multi-stage build for optimized image size

# Build stage
FROM node:18-alpine AS builder

# Install wget for health checks in runtime stage
RUN apk add --no-cache wget

WORKDIR /app

# Copy package files first for better caching
COPY package*.json ./
# Install all dependencies including devDependencies for building
RUN npm ci

# Copy source code
COPY . .

# Build the Next.js app
RUN npm run build

# Runtime stage
FROM node:18-alpine

WORKDIR /app

# Install wget for health checks and dumb-init for proper signal handling
RUN apk add --no-cache wget dumb-init

# Copy public files
COPY --from=builder /app/public ./public

# Copy .next/standalone files
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

# Expose port
EXPOSE 3000

# Health check - use netstat to check if port 3000 is listening
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD netstat -tlnp 2>/dev/null | grep :3000 || ss -tlnp 2>/dev/null | grep :3000 || exit 1

# Use dumb-init to handle signals properly
ENTRYPOINT ["dumb-init", "--"]

# Start the server
CMD ["node", "server.js"]