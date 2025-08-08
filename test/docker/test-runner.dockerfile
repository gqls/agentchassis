# test/docker/test-runner.dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build test binary
RUN go test -c -o test-runner ./test/cmd/runner

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy test runner and fixtures
COPY --from=builder /app/test-runner .
COPY --from=builder /app/test/fixtures ./fixtures

ENTRYPOINT ["./test-runner"]