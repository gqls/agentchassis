/ FILE: Dockerfile.image-generator-adapter
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o image-generator-adapter ./cmd/image-generator-adapter

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/image-generator-adapter /app/
COPY configs/image-adapter.yaml /app/configs/
RUN chown -R appuser:appgroup /app
USER appuser
CMD ["./image-generator-adapter", "-config", "configs/image-adapter.yaml"]