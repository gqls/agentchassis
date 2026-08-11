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
# THE POINT OF SHIPPING THE GO BINARY rather than a Python re-implementation
# (council round 2, corr 3eb0d1f1, gating objection from the `reuse_agent`
# seat): this check must walk workflow steps EXACTLY as the runtime validator
# does. It now does, because it IS that code — validation.WalkSteps, the same
# function platform/validation/workflow.go enforces with. The first version
# carried a hand-written Python walk, copied from single-owner-carriers-check,
# and that is bugs_open/144's exact shape: two hand-written traversals go blind
# in the same direction and then agree with each other, which reads as
# correctness. The parity test that policed the two copies is deleted with the
# copy — the drift it guarded cannot occur now.
#
# --report reads the fleet DIRECTLY from Postgres (PG_CLIENTS_HOST in the
# CronJob env; this image has no kubectl and the service account has no
# pods/exec RBAC — see cmd/config-key-audit/fleetdb.go) and writes ONE doc_notes
# row per run, clean or not, so "looked and found nothing" stays distinguishable
# from "stopped running". Exit 1 on any carrier: a live definition carrying a
# retired key is refused by the validator on every message once the declaring
# binary rolls, so this is the last place it can be a report rather than an
# outage.
CMD ["./config-key-audit", "--removed-keys-in-use", "--report"]
