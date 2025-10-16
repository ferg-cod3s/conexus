# Multi-stage build for Conexus MCP server
# Stage 1: Builder
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo \
    -ldflags '-extldflags "-static" -s -w' \
    -o conexus ./cmd/conexus

# Stage 2: Runtime
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs

# Create non-root user
RUN addgroup -g 1000 conexus && \
    adduser -D -u 1000 -G conexus conexus

# Create necessary directories
RUN mkdir -p /app /data /config && \
    chown -R conexus:conexus /app /data /config

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=conexus:conexus /build/conexus .

# Switch to non-root user
USER conexus

# Set default environment variables
ENV CONEXUS_HOST=0.0.0.0 \
    CONEXUS_PORT=8080 \
    CONEXUS_DB_PATH=/data/conexus.db \
    CONEXUS_ROOT_PATH=/data/codebase \
    CONEXUS_CHUNK_SIZE=512 \
    CONEXUS_CHUNK_OVERLAP=50 \
    CONEXUS_LOG_LEVEL=info \
    CONEXUS_LOG_FORMAT=json

# Expose MCP server port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the server
CMD ["./conexus"]
