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
# THE ACK LIST TRAVELS IN THE IMAGE, and the path in the repo stays the single
# source of truth — there is no second copy to drift.
#
# It is this check's own definition of "already known", so it must be versioned
# with the code that reads it. A ConfigMap ack list could be edited to silence a
# finding with no diff anybody reviews, which is the quiet unreviewed clearing
# this whole RFC is about one level up. `make build-shared-output-fields-check`
# builds from committed HEAD via `git archive`, so this COPY structurally cannot
# carry an uncommitted acknowledgement.
#
# Banking a real fix is therefore: fix the shape (or add a REASONED line to the
# ack file), commit, rebuild. Visible, reviewable, revertable.
COPY --from=builder /app/scripts/shared_output_fields_ack.txt /app/
RUN chown -R appuser:appgroup /app
USER appuser
# --report writes ONE doc_notes row per run, clean or not, so "looked and found
# nothing" stays distinguishable from "stopped running". --ack makes the check a
# RATCHET: green while only the two known pairs reproduce, exit 1 on a NEW one.
#
# Note the argument order deliberately puts --report BEFORE --ack: that is the
# exact shape the old positional parser silently ignored the ack file in, and
# TestParseSharedOutputArgs pins it. Nothing here needs to know that, which is
# the point of having fixed it rather than documented it.
#
# The fleet is read DIRECTLY from Postgres because PG_CLIENTS_HOST is set in the
# CronJob env — this image contains no kubectl and the service account has no
# pods/exec RBAC (see cmd/config-key-audit/fleetdb.go).
CMD ["./config-key-audit", "--shared-output-fields", "--report", "--ack", "/app/shared_output_fields_ack.txt"]
