# ===========================================
# Development Stage - for devcontainer
# ===========================================
FROM golang:1.23-bookworm AS development

WORKDIR /workspace

# Install development tools
RUN apt-get update && apt-get install -y \
    git \
    curl \
    postgresql-client \
    redis-tools \
    && rm -rf /var/lib/apt/lists/*

# Install Go tools (pinned versions compatible with Go 1.23)
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0 \
    && go install github.com/air-verse/air@v1.61.0

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum* ./
RUN go mod download && go mod verify

# Keep container running for devcontainer
CMD ["sleep", "infinity"]

# ===========================================
# Build Stage - for worker
# ===========================================
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the worker application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o worker cmd/worker/worker.go

# ===========================================
# Production Stage - minimal runtime image
# ===========================================
FROM alpine:latest AS production

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/worker .

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /root

USER appuser

CMD ["./worker"]

