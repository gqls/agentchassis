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
# bugs_open/309 candidate 2. Which ACTIVE component declares a data source that
# resolves nowhere for any site?
#
# CLC-018 shipped the BIRTH gate: a generated component whose input_schema names a
# `site_specs.<aspect>` no site has ever carried, an unregistered `query.*` name, or
# a prefix outside the vocabulary is refused at store_generated_component. The value
# would otherwise resolve to nothing, `on_missing` would default to skip_field, the
# key would be omitted, and a `{{if}}`-gated template would swallow the markup —
# producing a page that is complete-looking and data-less, indistinguishable from
# success at every stage. fundamentallyai.com's article index served six such cards,
# with zero links, for four months.
#
# WHY THE BIRTH GATE IS NOT ENOUGH, which is this image's whole reason to exist. The
# gate only fires on GENERATION. A component is routinely inserted or altered by a
# hand-written migration or by hand SQL, which never passes through the Go action at
# all — a standing LANDMINE entry, and precisely how the motivating component got
# there. MEASURED 2026-08-22: 69 fields across 17 active components still declare an
# unresolvable source, six of them live on 46 page instances, none of which the gate
# can ever see.
#
# WHY A CRONJOB AND NOT A REPO-SIDE TEST, which would need no image: at commit time a
# migration is unapplied, and `content_components` is routinely changed directly in
# the database with no commit at all. The only place the question has a true answer is
# the live table, on a clock. RFC_006's settled reasoning; single-owner-carriers-check
# records it, and component-fallback-check's header makes the same argument for the
# same table.
#
# WHY IT CALLS THE GUARD'S OWN FUNCTION rather than mirroring the rule in Python like
# the three older checks. The concept register asked for exactly this — CLC-018:
# "build it ON sourceVocabularyIssues, not on a second predicate, or they drift". A
# mirror can only ever DETECT drift; calling the function makes drift unrepresentable.
# The owner's 2026-08-14 preference for Python was answering a different objection —
# compiling INSIDE the job (git clone of a 262M repo + go mod download, uncertain
# egress) — which a pre-built image does not do.
#
# --report reads the library DIRECTLY from Postgres (PG_CLIENTS_HOST in the CronJob
# env; this image has no kubectl and the service account has no pods/exec RBAC in this
# namespace — see cmd/config-key-audit/fleetdb.go) and writes ONE doc_notes row per
# run, clean or not, so "looked and found nothing" stays distinguishable from "stopped
# running". It deliberately does NOT write to agent_error_log: bugs_open/358 measures
# that channel as write-only with a 30-day retention delete.
#
# EXIT CODES: 0 every live finding is grandfathered by the frozen baseline and
# unchanged; 1 a RED — a finding outside the baseline, a component baselined while
# DORMANT that has since been deployed, or a stale baseline entry; 2 the check could
# not run, which must never read as a pass.
#
# A RED IS NEVER A STANDING BACKLOG. It is always either a real new finding or a
# one-line trim of the baseline file. Read the doc_notes row it wrote before assuming
# the job itself is broken.
CMD ["./config-key-audit", "--component-source-vocabulary", "--report"]
