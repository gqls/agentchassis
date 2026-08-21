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
# SHIPS THE ACKS FILE WITH THE BINARY, for the reason shared-output-fields-check
# ships its ack list and component-render-check ships its baseline: the acks file
# is this check's own definition of "already known", so an image built from a
# working tree could bake in an UNREVIEWED acknowledgement — a silenced finding
# with no diff to review. `make build-*` builds from committed HEAD, which makes
# that unrepresentable rather than merely discouraged.
COPY --from=builder /app/docs/agent_docs/docs024_key_docs_latest/architecture_review/optional_explicit_wire_acks.json /app/
RUN chown -R appuser:appgroup /app
USER appuser
# RFC_029 §10.15 / CTS-060(5). The `?` OPTIONAL-EXPLICIT marker resolves a field
# from its declared path or leaves it ABSENT — deliberately with no error, no log
# and no row, because absence IS the contract. Nothing at runtime can tell an
# intended absence from a mistyped path, so adoption is gated offline: every live
# `?` wire must carry an entry in the acks file stating what was checked
# DOWNSTREAM (that the consuming code treats absent as a safe no-op, not as an
# empty value it will render or persist).
#
# --report reads the fleet DIRECTLY from Postgres (PG_CLIENTS_HOST in the CronJob
# env; this image has no kubectl and the service account has no pods/exec RBAC —
# see cmd/config-key-audit/fleetdb.go) and writes ONE doc_notes row per run,
# clean or not, so "looked and found nothing" stays distinguishable from
# "stopped running".
#
# EXIT CODES: 0 every live wire acknowledged (including zero wires — the marker
# is opt-in); 1 a live unacknowledged wire; 2 the check could not run, which must
# never read as a pass. That third case includes the blind-green feed shape: a
# fleet export that decodes but walks no steps is refused rather than reported as
# a clean estate.
CMD ["./config-key-audit", "--optional-explicit-wires", "--report", "--acks", "/app/optional_explicit_wire_acks.json"]
