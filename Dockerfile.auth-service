// FILE: Dockerfile.auth-service
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth-service ./cmd/auth-service

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/auth-service /app/
COPY configs/auth-service.yaml /app/configs/
RUN chown -R appuser:appgroup /app
USER appuser
CMD ["./auth-service", "-config", "configs/auth-service.yaml"]