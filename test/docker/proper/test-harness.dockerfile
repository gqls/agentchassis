# test/docker/test-harness.dockerfile
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    gcc \
    g++ \
    musl-dev \
    bash \
    curl

WORKDIR /workspace

# Copy the ENTIRE project
COPY go.mod go.sum ./
RUN go mod download

# Copy ALL source code
COPY . .

# Verify the module structure
RUN ls -la /workspace && \
    echo "Go module:" && \
    go list -m && \
    echo "Available packages:" && \
    go list ./...

# Install Go testing tools (if needed)
RUN go install github.com/onsi/ginkgo/v2/ginkgo@latest && \
    go install github.com/onsi/gomega/...@latest && \
    go install gotest.tools/gotestsum@latest

# BUILD THE GO HARNESS
RUN go build -o /usr/local/bin/test-harness ./test/cmd/harness/main.go

# Stage 2: Runtime
FROM golang:1.23-alpine

# Install runtime dependencies
RUN apk add --no-cache \
    bash \
    git \
    make \
    gcc \
    g++ \
    musl-dev \
    postgresql-client \
    curl \
    jq \
    ca-certificates

# Copy Go tools from builder
COPY --from=builder /go/bin/* /usr/local/bin/

# Copy the compiled harness
COPY --from=builder /usr/local/bin/test-harness /usr/local/bin/test-harness

# Create workspace with proper module structure
WORKDIR /workspace

# Copy the entire workspace for test execution
COPY --from=builder /workspace /workspace

# Verify setup
RUN cd /workspace && \
    go mod download && \
    echo "Module ready: $(go list -m)"

# Environment variables
ENV GO_ENV=test \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    TEST_RESULTS_DIR=/results

# The Go harness handles output directly
CMD ["/usr/local/bin/test-harness"]