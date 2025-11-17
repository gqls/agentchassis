# build/docker/backend/Dockerfile.git-adapter
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files (you'll need to create these)
# COPY go.mod go.sum ./
# RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o git-adapter cmd/git-adapter/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/git-adapter .

# Health check endpoint (optional)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/bin/sh", "-c", "ps aux | grep git-adapter | grep -v grep || exit 1"]

# Run the binary
CMD ["./git-adapter"]