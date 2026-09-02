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
# SHIPS THE ACKS FILE WITH THE BINARY, exactly as ungraded-completions-check and
# commit-sha-exposure-check do and for the same reason: the acks file is this
# check's own definition of a permitted exception, so an image built from a
# working tree could bake in an UNREVIEWED acknowledgement — a silenced finding
# with no diff to review. `make build-*` builds from committed HEAD, which makes
# that unrepresentable.
COPY --from=builder /app/docs/agent_docs/docs024_key_docs_latest/architecture_review/render_truncation_acks.json /app/
RUN chown -R appuser:appgroup /app
USER appuser
# The commissioned reader for RENDER_AUDIT_TRUNCATED (bugs_open/394, owner
# decision 4, 2026-08-25).
#
# WHAT THE ROW MEANS, AND WHY IT NEEDED A READER RATHER THAN AN ALARM. The render
# audit caps a site's sweep at max_pages. Until the coverage cursor it took the
# SAME prefix every run, so pages past the cap had never been audited and never
# would be — bugs_closed/242 made that loud and nothing read the signal. The
# cursor changed what the row ASSERTS: under coverage_mode=cursor a truncation
# row is healthy pagination, this run's window, and alarming on the code itself
# would now be wrong. Telling the three meanings apart is the whole job.
#
# THE ARMS: (a) coverage_mode absent or "prefix" from render-audit-agent, the
# opted-in rotating caller — the migration-660 flip regressed or the pod predates
# the cursor; (b) two consecutive cursor rows at the same window_first — a
# stalled cursor, which looks healthy from every other angle; (c) a caller
# neither opted in nor acknowledged. design-critique-agent is acked at birth
# (manual sampler, no cadence, its 8-page prefix is plausibly the intended
# vision-critique sample).
#
# DORMANT GROUPS ARE REPORTED, NOT JUDGED. A site that has STOPPED truncating
# freezes its last row for ever; judging it would make this check red on day one
# and every day after. Measured relative to the fleet's newest row, so it is a
# pure function of the data. Dormant groups are counted and NAMED, so "0
# findings" cannot quietly become "0 findings among the groups I still look at".
#
# --report reads agent_error_log DIRECTLY from Postgres (PG_CLIENTS_HOST in the
# CronJob env; no kubectl in this image) and writes ONE doc_notes row per run,
# clean or not, so "looked and found nothing" stays distinguishable from
# "stopped running".
#
# EXIT CODES: 0 every truncation row is accounted for (or none exists); 1 a
# finding; 2 the check could not run — including an agent_error_log that reads as
# EMPTY, which means the read went blind, never that the fleet stopped erring.
CMD ["./config-key-audit", "--render-truncation", "--report", "--acks", "/app/render_truncation_acks.json"]
