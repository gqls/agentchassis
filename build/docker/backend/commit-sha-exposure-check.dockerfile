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
# SHIPS THE ACKS FILE WITH THE BINARY, exactly as optional-explicit-wires-check
# does and for the same reason: the acks file is this check's own definition of
# a permitted exception, so an image built from a working tree could bake in an
# UNREVIEWED acknowledgement — a silenced finding with no diff to review.
# `make build-*` builds from committed HEAD, which makes that unrepresentable.
COPY --from=builder /app/docs/agent_docs/docs024_key_docs_latest/architecture_review/commit_sha_exposure_acks.json /app/
RUN chown -R appuser:appgroup /app
USER appuser
# The STANDING form of migration 537's apply-time guard (bugs_closed/334).
# 537 wired build-dispatch-loop's mark_complete to the handler's own reply —
# `commit_sha?`: that path or ABSENCE, never the whole-tree search. Absence is
# the contract, so a NEW commit-producing handler that does not expose
# `response.commit_sha` via its complete step's result_mapping will simply
# never record result.commit_sha: no error, no log, no row. The guard proved
# the estate ready ONCE, at apply time; the handler population is per-item and
# dynamic. This asks the same question daily.
#
# --report reads the fleet and the two set queries DIRECTLY from Postgres
# (PG_CLIENTS_HOST in the CronJob env; no kubectl in this image, no pods/exec
# RBAC on the service account — see cmd/config-key-audit/fleetdb.go) and writes
# ONE doc_notes row per run, clean or not, so "looked and found nothing" stays
# distinguishable from "stopped running".
#
# EXIT CODES: 0 every commit-producing live handler exposes response.commit_sha
# or carries a reasoned exception in the acks file; 1 at least one does
# neither; 2 the check could not run — including the three blind states
# (empty producers set, empty handlers set, zero exposing agents fleet-wide),
# each of which would otherwise print a clean report over a dead query.
CMD ["./config-key-audit", "--commit-sha-exposure", "--report", "--acks", "/app/commit_sha_exposure_acks.json"]
