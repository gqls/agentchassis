FROM golang:1.24-alpine AS builder
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

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9090/health || exit 1

# Expose health check port
EXPOSE 9090

# Set default environment variables
ENV SERVICE_ENVIRONMENT=production \
    SERVICE_LOGGING_LEVEL=info \
    S3_USE_SSL=true \
    S3_REGION=us-east-1

CMD ["./image-generator-adapter", "-config", "configs/image-adapter.yaml"]