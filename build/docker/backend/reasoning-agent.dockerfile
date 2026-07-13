# --- Build Stage ---
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o reasoning-agent ./cmd/reasoning-agent

# --- Final Stage ---
FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the correct working directory
WORKDIR /app

# Copy the binary and config into the working directory
COPY --from=builder /app/reasoning-agent .
COPY configs/reasoning-agent.yaml ./configs/

# Set correct ownership and user
RUN chown -R appuser:appgroup /app
USER appuser

# Tell the application where to find its config
CMD ["./reasoning-agent", "-config", "configs/reasoning-agent.yaml"]