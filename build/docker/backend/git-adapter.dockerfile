# build/docker/backend/Dockerfile.git-adapter

# --- Builder Stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Copy all go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the git-adapter binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /app/git-adapter ./cmd/git-adapter/main.go

# --- Final Stage ---
FROM gcr.io/distroless/static-debian11

WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /app/git-adapter /app/git-adapter

# Copy the config file
# This assumes you have a 'configs' dir in your build context
COPY configs/git-adapter.yaml /app/configs/git-adapter.yaml

# Set the entrypoint
ENTRYPOINT ["/app/git-adapter", "-config", "/app/configs/git-adapter.yaml"]