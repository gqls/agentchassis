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
# ambiguity). No ack file: a finding here is a template variable that resolves
# on no row, ever — always a defect, never a blessed exception. If a legitimate
# exception ever appears, add the ack mechanism the way
# shared-output-fields-check carries it (list travels IN the image, committed,
# reviewable), not as a ConfigMap.
#
# READ-ONLY BY CONSTRUCTION, and this is the property the owner asked about
# before scheduling it (2026-09-03): the mode's only two database touches are
# the fleet-export SELECT over agent_definitions and the doc_notes INSERT that
# --report performs. No Kafka, no object storage, no write to pages,
# page_components, site_components or any site artefact. It cannot cause a site
# to change.
#
# The fleet is read DIRECTLY from Postgres because PG_CLIENTS_HOST is set in the
# CronJob env — this image contains no kubectl and the service account has no
# pods/exec RBAC (see cmd/config-key-audit/fleetdb.go).
CMD ["./config-key-audit", "--template-input-fields", "--report"]
