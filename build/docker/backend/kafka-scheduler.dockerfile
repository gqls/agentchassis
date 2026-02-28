# FILE: build/docker/backend/kafka-scheduler.dockerfile
# Build the kafka-scheduler binary from the agent-chassis repo.
# Uses the same Go module so it can import platform/kafka.
#
# Build: docker build -f build/docker/backend/kafka-scheduler.dockerfile -t aqls/kafka-scheduler:v1.0.1 .
# Push:  docker push aqls/kafka-scheduler:v1.0.1

FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kafka-scheduler ./cmd/scheduler

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/kafka-scheduler /app/
RUN chown -R appuser:appgroup /app
USER appuser
EXPOSE 8080
CMD ["./kafka-scheduler"]
