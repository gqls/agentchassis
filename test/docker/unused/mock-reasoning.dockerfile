# test/docker/mock-generic.dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 1. Copy only the go.mod and go.sum files to leverage caching
COPY go.mod go.sum ./

# 2. Download dependencies
RUN go mod download

# 3. Copy the rest of the source code
COPY . .

# 4. Build the specific mock agent binary
# The source path is relative to the project root (the build context)
RUN go build -o /app/mock-agent ./test/docker/mock-generic-agent/main.go

# 5. Final, small image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/mock-agent /usr/local/bin/
ENTRYPOINT ["mock-agent"]