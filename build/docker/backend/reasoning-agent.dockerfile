# Dockerfile for the reasoning-agent service

# --- Build Stage ---
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies first, leveraging Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

# Build the reasoning-agent binary
# The output path is /app/reasoning-agent
RUN CGO_ENABLED=0 GOOS=linux go build -v -o reasoning-agent ./cmd/reasoning-agent

# --- Final Stage ---
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/reasoning-agent .
COPY configs/reasoning-agent.yaml /app/configs/

# Expose the port the service might use for health checks (if any)
# EXPOSE 8080

# The command to run when the container starts
ENTRYPOINT ["./reasoning-agent"]