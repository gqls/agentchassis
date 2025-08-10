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

# Install Go testing tools
RUN go install github.com/onsi/ginkgo/v2/ginkgo@latest && \
    go install github.com/onsi/gomega/...@latest && \
    go install gotest.tools/gotestsum@latest

# Stage 2: Runtime
FROM golang:1.23-alpine

# Install runtime dependencies INCLUDING coreutils for stdbuf
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
    ca-certificates \
    coreutils

# Copy Go tools from builder
COPY --from=builder /go/bin/* /usr/local/bin/

# Create workspace with proper module structure
WORKDIR /workspace

# Copy the entire module (including source code)
COPY --from=builder /workspace /workspace

# Copy test harness script
COPY test/scripts/test-harness.sh /usr/local/bin/test-harness
RUN chmod +x /usr/local/bin/test-harness && \
    sed -i 's/\r$//' /usr/local/bin/test-harness

# Verify setup
RUN cd /workspace && \
    go mod download && \
    echo "Module ready: $(go list -m)"

# Environment variables
ENV GO_ENV=test \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    TEST_RESULTS_DIR=/results \
    GOTEST_FLAGS="-v -count=1"

# Use stdbuf for unbuffered output
CMD ["stdbuf", "-o0", "-e0", "/usr/local/bin/test-harness"]