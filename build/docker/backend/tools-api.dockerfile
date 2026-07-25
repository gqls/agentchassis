# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /tools-api ./cmd/tools-api

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

COPY --from=builder /tools-api /tools-api

EXPOSE 8083

ENTRYPOINT ["/tools-api"]
