# FILE: build/docker/backend/content-creator-agent.dockerfile
# A dedicated Dockerfile for building the content creator agent service.

# Stage 1: Build the application
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project context
COPY . .

# Build the specific service binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/content-creator-agent ./cmd/content-creator-agent

# Stage 2: Create the final small image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Create a non-root user for security
RUN addgroup -S appgroup && adduser -S -G appgroup appuser

WORKDIR /app

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/content-creator-agent /app/content-creator-agent

# Copy its specific configuration file
COPY configs/content-creator-agent.yaml /app/configs/content-creator-agent.yaml

RUN chown appuser:appgroup /app/content-creator-agent

USER appuser

# The command to run the service, pointing to its own config
CMD ["./content-creator-agent", "-config", "configs/content-creator-agent.yaml"]