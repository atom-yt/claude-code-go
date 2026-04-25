# Frontend Dockerfile for atom-ai-platform
# Multi-stage build for optimized image size

# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files first for better caching
COPY package*.json ./
RUN npm ci --only=production

# Copy source code
COPY . .

# Build the Next.js app
RUN npm run build

# Runtime stage
FROM node:18-alpine

WORKDIR /app

# Install dumb-init for proper signal handling
RUN apk add --no-cache dumb-init

# Copy public files
COPY --from=builder /app/public ./public

# Copy .next/standalone files
COPY --from=builder --chown=node:node /app/.next/standalone ./
COPY --from=builder --chown=node:node /app/.next/static ./.next/static

# Create non-root user
RUN addgroup -g 1001 -S nodejs && \n    adduser -S nodejs -u 1001 && \n    chown -R nodejs:nodejs /app

# Switch to non-root user
USER nodejs

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \n    CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

# Use dumb-init to handle signals properly
ENTRYPOINT ["dumb-init", "--"]

# Start the server
CMD ["node", "server.js"]