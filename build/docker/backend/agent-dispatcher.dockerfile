# build/docker/backend/agent-dispatcher.dockerfile
#
# Build: docker build -t docker.io/aqls/agent-dispatcher:v1.0.0 -f build/docker/backend/agent-dispatcher.dockerfile .
# Push:  docker push docker.io/aqls/agent-dispatcher:v1.0.0

FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /dispatcher ./cmd/dispatcher/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /dispatcher /dispatcher

ENTRYPOINT ["/dispatcher"]
