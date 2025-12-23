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
    && rm -rf /var/lib/apt/lists/*

# Install Go tools (pinned versions compatible with Go 1.23)
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0 \
    && go install github.com/air-verse/air@v1.61.0

# Keep container running for devcontainer
CMD ["sleep", "infinity"]

# ===========================================
# Build Stage - for production
# ===========================================
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main cmd/api/main.go

# ===========================================
# Production Stage - minimal runtime image
# ===========================================
FROM alpine:latest AS production

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /root

USER appuser

EXPOSE 8080

CMD ["./main"]