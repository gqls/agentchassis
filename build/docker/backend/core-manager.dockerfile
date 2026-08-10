FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG GIT_COMMIT=unknown
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-X github.com/gqls/agentchassis/pkg/buildinfo.GitCommit=${GIT_COMMIT}" \
    -o core-manager ./cmd/core-manager

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/core-manager /app/
COPY configs/core-manager.yaml /app/configs/
RUN chown -R appuser:appgroup /app
USER appuser
CMD ["./core-manager", "-config", "configs/core-manager.yaml"]