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
# bugs_open/316. Which capped `query_database` step picks its work by a STATIC
# sort over a candidate set that the CLOCK refills?
#
# content-feed-trigger.find_news_sites ended `ORDER BY s.domain LIMIT 5` over the
# news-feed sites whose sources were due. The runs are 6-hourly and the sort is
# stable, so the same five names won every time more than five were in
# contention: measured over five consecutive cap-hitting runs, the alphabetically
# LAST eligible site was selected ZERO times while continuously due, reaching
# 419% of its own configured cadence.
#
# WHY THIS IS NOT ALREADY COVERED BY LCO-009, which watches the same step.
# That check counts rows and warns when a result reaches its ceiling. The row
# count is IDENTICAL whether the ordering is fair or not, so it reported "hit its
# cap" truthfully on every run for days while the defect went unseen. This check
# reads the ORDER BY.
#
# WHY A CRONJOB AND NOT A REPO-SIDE TEST, which would need no image at all: at
# commit time a migration is unapplied, and config on this platform is routinely
# changed directly in the database with no commit. The only place the question
# has a true answer is live `agent_definitions`, on a clock. That is RFC_006's
# settled reasoning, and single-owner-carriers-check's docstring records it.
#
# --report reads the fleet DIRECTLY from Postgres (PG_CLIENTS_HOST in the CronJob
# env; this image has no kubectl and the service account has no pods/exec RBAC —
# see cmd/config-key-audit/fleetdb.go) and writes ONE doc_notes row per run,
# clean or not, so "looked and found nothing" stays distinguishable from
# "stopped running".
#
# EXIT CODES: 0 no capped step picks clock-replenished work by a static sort
# (including zero capped steps); 1 at least one does; 2 the check could not run,
# which must never read as a pass — that includes a fleet export that decodes to
# zero agents, which is refused rather than reported as a clean estate.
CMD ["./config-key-audit", "--capped-schedule-ordering", "--report"]
