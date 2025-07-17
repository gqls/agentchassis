FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o agent-chassis ./cmd/agent-chassis

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/agent-chassis /app/
COPY configs/agent-chassis.yaml /app/configs/
RUN chown -R appuser:appgroup /app
USER appuser
CMD ["./agent-chassis", "-config", "configs/agent-chassis.yaml"]