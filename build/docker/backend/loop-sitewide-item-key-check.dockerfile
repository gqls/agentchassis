FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o config-key-audit ./cmd/config-key-audit

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/config-key-audit /app/
RUN chown -R appuser:appgroup /app
USER appuser
# --report writes ONE doc_notes row per run, clean or not, so "looked and found
# nothing" stays distinguishable from "stopped running" (bugs_open/140's
# ambiguity). No ack file: the fleet is clean since migration 493, and this
# check's findings have no known-and-accepted class — a finding here is always
# work lost silently (bugs_open/321), never a blessed exception. If a
# legitimate exception ever appears, add the ack mechanism the way
# shared-output-fields-check carries it (list travels IN the image, committed,
# reviewable), not as a ConfigMap.
#
# The fleet is read DIRECTLY from Postgres because PG_CLIENTS_HOST is set in the
# CronJob env — this image contains no kubectl and the service account has no
# pods/exec RBAC (see cmd/config-key-audit/fleetdb.go).
CMD ["./config-key-audit", "--loop-sitewide-item-keys", "--report"]
