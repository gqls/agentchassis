# Add to your Dockerfile (or create a separate one)
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o workflow-monitor cmd/workflow-monitor/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/workflow-monitor .
ENTRYPOINT ["./workflow-monitor"]