# FILE: Dockerfile.platform
# Base image with common dependencies
FROM golang:1.21-alpine AS base
RUN apk add --no-cache git ca-certificates
WORKDIR /app
