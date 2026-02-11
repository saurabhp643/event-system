# Backend Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the rest of the application
COPY backend/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server ./backend

# Final stage - minimal image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the built binary
COPY --from=builder /app/bin/server .
COPY --from=builder /app/backend/config.yaml .

# Create data directory for SQLite
RUN mkdir -p /app/data

EXPOSE 8080

# Run the server
CMD ["./server"]
