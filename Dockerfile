# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o paperless-uploader \
    ./cmd/paperless-uploader

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/paperless-uploader .

# Copy default config if exists
COPY --from=builder /app/config.yaml* ./

# Create necessary directories
RUN mkdir -p /app/consume /app/processed && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose any necessary ports (if applicable)
# EXPOSE 8080

# Set environment variables
ENV CONFIG_PATH=/app/config.yaml

# Run the application
ENTRYPOINT ["./paperless-uploader"]
