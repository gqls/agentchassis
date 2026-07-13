# build/docker/backend/remote-job-spawner.dockerfile
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /remote-job-spawner ./cmd/remote-job-spawner/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /remote-job-spawner /remote-job-spawner

ENTRYPOINT ["/remote-job-spawner"]
