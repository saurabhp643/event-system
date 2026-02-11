# Multi-stage build for full-stack application

# Stage 1: Build Go backend
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the rest of the backend application
COPY backend/ ./

# Build the Go application from current directory (.)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server .

# Stage 2: Build React frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app

# Copy package.json and package-lock.json
COPY frontend/package*.json ./

# Install dependencies
RUN npm ci

# Copy frontend source code
COPY frontend/ ./

# Build the frontend
RUN npm run build

# Stage 3: Final production image with Nginx
FROM nginx:alpine

# Install required packages
RUN apk --no-cache add ca-certificates curl

# Copy nginx configuration for proxying
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Create directories for backend and frontend
RUN mkdir -p /app/backend /app/data

# Copy Go binary from backend-builder
COPY --from=backend-builder /app/bin/server /app/backend/server
COPY --from=backend-builder /app/config.yaml /app/backend/config.yaml

# Copy built frontend files from frontend-builder
COPY --from=frontend-builder /app/dist /usr/share/nginx/html

# Copy entrypoint script
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Expose ports (Nginx serves on 80, backend on 8080 internally)
EXPOSE 8080

# Set working directory
WORKDIR /app

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:80/health || curl -f http://localhost:8080/health || exit 1

# Entrypoint script manages both Nginx and Go backend
ENTRYPOINT ["/app/entrypoint.sh"]
