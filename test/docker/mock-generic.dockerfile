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

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy ONLY the directories we need for tests
# This prevents accidentally including unwanted directories
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
COPY platform ./platform
COPY test ./test
COPY configs ./configs

# Verify we have the right structure (no 'tests' directory)
RUN ls -la /workspace && \
    echo "Directories present:" && \
    ls -d */ | grep -E "(test|tests)" || echo "Only 'test' directory present (good!)"

# Verify the module structure
RUN echo "Go module:" && \
    go list -m && \
    echo "Test packages:" && \
    go list ./test/...

# Install Go testing tools
RUN go install github.com/onsi/ginkgo/v2/ginkgo@latest && \
    go install github.com/onsi/gomega/...@latest && \
    go install gotest.tools/gotestsum@latest

# Stage 2: Runtime
FROM golang:1.23-alpine

# Install runtime dependencies
RUN apk add --no-cache \
    bash \
    git \
    make \
    postgresql-client \
    curl \
    jq \
    ca-certificates

# Copy Go tools from builder
COPY --from=builder /go/bin/* /usr/local/bin/

# Create workspace
WORKDIR /workspace

# Copy only what we need from builder
COPY --from=builder /workspace /workspace

# Create directories for results
RUN mkdir -p /results /fixtures

# Copy and fix the test harness script
COPY test/scripts/test-harness.sh /usr/local/bin/test-harness
RUN chmod +x /usr/local/bin/test-harness && \
    sed -i 's/\r$//' /usr/local/bin/test-harness

# Environment variables
ENV GO_ENV=test \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    TEST_RESULTS_DIR=/results \
    GOTESTSUM_FORMAT=testname

ENTRYPOINT ["test-harness"]
CMD ["test"]