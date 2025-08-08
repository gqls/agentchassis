# test/docker/test-monitor.dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .

RUN go build -o test-monitor ./test/tools/dashboard/test_dashboard.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/test-monitor /usr/local/bin/
EXPOSE 8090
ENTRYPOINT ["test-monitor"]