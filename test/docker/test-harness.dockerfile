# test/docker/test-harness.dockerfile
# Go-focused test harness for Kubernetes environments

# Stage 1: Go builder with test tools
FROM golang:1.21-alpine AS builder

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

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Install Go testing tools
RUN go install github.com/onsi/ginkgo/v2/ginkgo@latest && \
    go install github.com/onsi/gomega/...@latest && \
    go install gotest.tools/gotestsum@latest && \
    go install github.com/vektra/mockery/v2@latest && \
    go install github.com/golang/mock/mockgen@latest && \
    go install github.com/rakyll/gotest@latest && \
    go install github.com/axw/gocov/gocov@latest && \
    go install github.com/AlekSi/gocov-xml@latest && \
    go install github.com/matm/gocov-html/cmd/gocov-html@latest && \
    go install github.com/jstemmer/go-junit-report/v2@latest

# Copy source code
COPY . .

# Build test binaries
RUN go test -c -o /test-runner ./test/cmd/runner

# Stage 2: Kafka tools
FROM confluentinc/cp-kafka:7.5.0 AS kafka-tools

# Stage 3: Final test harness
FROM golang:1.21-alpine

# Install runtime dependencies
RUN apk add --no-cache \
    bash \
    git \
    make \
    postgresql-client \
    mysql-client \
    redis \
    curl \
    wget \
    jq \
    vim \
    tmux \
    htop \
    ca-certificates \
    netcat-openbsd

# Install k6 for load testing
RUN wget -q https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-linux-amd64.tar.gz && \
    tar -xzf k6-v0.47.0-linux-amd64.tar.gz && \
    mv k6-v0.47.0-linux-amd64/k6 /usr/local/bin/ && \
    rm -rf k6-v0.47.0-linux-amd64*

# Install kafkacat/kcat
RUN wget -q https://github.com/edenhill/kcat/releases/download/1.7.1/kcat-1.7.1-linux-amd64 -O /usr/local/bin/kcat && \
    chmod +x /usr/local/bin/kcat && \
    ln -s /usr/local/bin/kcat /usr/local/bin/kafkacat

# Copy Kafka tools
COPY --from=kafka-tools /usr/bin/kafka-* /usr/local/bin/
COPY --from=kafka-tools /usr/share/java/kafka /usr/share/java/kafka

# Copy Go test tools from builder
COPY --from=builder /go/bin/* /usr/local/bin/
COPY --from=builder /test-runner /usr/local/bin/test-runner

# Create directories
RUN mkdir -p /workspace /results /fixtures /scripts /tools

# Copy test files
COPY test/scripts /scripts/
COPY test/fixtures /fixtures/
COPY test/tools /tools/
COPY test /workspace/test/

# Copy the main test harness script
COPY test/scripts/test-harness.sh /usr/local/bin/test-harness
RUN chmod +x /usr/local/bin/test-harness
RUN chmod +x /usr/local/bin/test-harness

# Create Kafka test helper
COPY --from=builder /usr/local/bin/run-kafka-test /usr/local/bin/run-kafka-test

# Create database test helper
COPY --from=builder /usr/local/bin/run-db-test /usr/local/bin/run-db-test

# Set up workspace
WORKDIR /workspace

# Environment variables
ENV GO_ENV=test \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    TEST_RESULTS_DIR=/results \
    TEST_FIXTURES_DIR=/fixtures \
    TEST_TIMEOUT=30m \
    GOTESTSUM_FORMAT=testname

# Default entrypoint
ENTRYPOINT ["test-harness"]
CMD ["test"]