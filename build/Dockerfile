# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main cmd/api/main.go

# Final stage - minimal runtime image
FROM alpine:latest

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