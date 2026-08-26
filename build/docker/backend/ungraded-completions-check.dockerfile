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
# SHIPS THE ACKS FILE WITH THE BINARY, exactly as commit-sha-exposure-check and
# optional-explicit-wires-check do and for the same reason: the acks file is
# this check's own definition of a permitted exception, so an image built from
# a working tree could bake in an UNREVIEWED acknowledgement — a silenced
# finding with no diff to review. `make build-*` builds from committed HEAD,
# which makes that unrepresentable.
COPY --from=builder /app/docs/agent_docs/docs024_key_docs_latest/architecture_review/no_change_unreadable_acks.json /app/
RUN chown -R appuser:appgroup /app
USER appuser
# The commissioned reader for NO_CHANGE_GATE_UNREADABLE_RESULT (bugs_open/393,
# owner decision 4, 2026-08-25). A row of that code means a rostered
# abstain-policy item_type completed UNGRADED: its handler's result shape has
# drifted from the counters noChangeGates declares. The first instance sat
# unread for 11 days and was found by a census accident; this asks daily, and
# any item_type not acknowledged (with its diagnosis) in the acks file fails
# the run — a drifting handler becomes a finding the morning after it first
# appears, instead of never.
#
# --report reads agent_error_log DIRECTLY from Postgres (PG_CLIENTS_HOST in the
# CronJob env; no kubectl in this image — see cmd/config-key-audit/fleetdb.go)
# and writes ONE doc_notes row per run, clean or not, so "looked and found
# nothing" stays distinguishable from "stopped running".
#
# EXIT CODES: 0 every drifting item_type is acknowledged (or none exists);
# 1 a NEW drifting item_type; 2 the check could not run — including an
# agent_error_log that reads as EMPTY, which means the read went blind, never
# that the fleet stopped erring.
CMD ["./config-key-audit", "--ungraded-completions", "--report", "--acks", "/app/no_change_unreadable_acks.json"]
