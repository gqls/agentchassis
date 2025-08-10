# test/docker/test-harness-from-test-dir.dockerfile
FROM golang:1.23-alpine AS deps

WORKDIR /workspace

# Copy go.mod and go.sum from parent directory
COPY ../go.mod ../go.sum ./
RUN go mod download

# Separate stage for test tools
FROM golang:1.23-alpine AS tools
RUN go install gotest.tools/gotestsum@latest

# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /workspace

# Copy downloaded modules from deps stage
COPY --from=deps /go/pkg /go/pkg

# Copy source files from parent directory
COPY ../go.mod ../go.sum ./
COPY . ./test
COPY ../pkg ./pkg
COPY ../platform ./platform
COPY ../internal ./internal

# Final minimal runtime
FROM golang:1.23-alpine

RUN apk add --no-cache bash git postgresql-client

WORKDIR /workspace

# Copy modules
COPY --from=deps /go/pkg /go/pkg

# Copy source
COPY --from=builder /workspace ./

# Copy tools
COPY --from=tools /go/bin/gotestsum /usr/local/bin/

# Copy test script
COPY scripts/test-harness.sh /usr/local/bin/test-harness
RUN chmod +x /usr/local/bin/test-harness && \
    sed -i 's/\r$//' /usr/local/bin/test-harness

ENV CGO_ENABLED=0 \
    GOPATH=/go \
    PATH=/go/bin:/usr/local/go/bin:$PATH

ENTRYPOINT ["test-harness"]
CMD ["test"]