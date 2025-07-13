// FILE: Dockerfile.core-manager
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o core-manager ./cmd/core-manager

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/core-manager /app/
COPY configs/core-manager.yaml /app/configs/
RUN chown -R appuser:appgroup /app
USER appuser
CMD ["./core-manager", "-config", "configs/core-manager.yaml"]