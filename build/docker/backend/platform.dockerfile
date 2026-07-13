# Base image with common dependencies
FROM golang:1.24-alpine AS base
RUN apk add --no-cache git ca-certificates
WORKDIR /app
