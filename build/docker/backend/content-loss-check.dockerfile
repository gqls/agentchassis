FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o content-loss-check ./cmd/content-loss-check

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/content-loss-check /app/
RUN chown -R appuser:appgroup /app
USER appuser
# NO FLAGS — the verifier-remit-check inversion, kept deliberately: the writing
# path (file CONTENT_KEY_LOSS findings, resolve healed ones, record the run in
# doc_notes) is the DEFAULT, and --dry-run is what suppresses it. A flag the
# CronJob must remember to pass is a silence waiting to happen; a flag it must
# remember NOT to pass cannot silence anything.
#
# The binary is its own reader (bugs_open/355 A3, same-commit rule): the
# detector's findings expire from agent_error_log (mig 466 retention), so the
# durable record is the doc_notes heartbeat this writes on EVERY run — clean
# runs included — plus the live-damage census it re-derives from current state.
CMD ["./content-loss-check"]
