# bugs_open/358 phase 2 — the daily finding-code registry check.
#
# WHAT IT CHECKS. `agent_error_log` carries deliberately-written FINDING CODES: a
# detector's record of something it noticed and will not fix. 358 measured that
# most had no automated reader and are deleted at 30 days unresolved (14 if
# resolved — marking a row resolved makes it die FASTER). The fix is not a reader
# per code: some are legitimately human evidence, some are time-boxed
# instrumentation with an owner, and operational plumbing is correctly consumed by
# the generic newest-N diagnostic reads. It is that A CODE CANNOT ENTER THE ESTATE
# WITH NO DECLARED DISPOSITION AND NOBODY NOTICE.
#
# WHY AN IMAGE AND NOT A ConfigMap SCRIPT. Same argument as
# shared-output-fields-check: the rules are a Go mode with 30 tests, and a Python
# re-implementation would be a replica — a replica proves the replica works.
#
# WHY THE REGISTRY TRAVELS IN THE IMAGE, which is a real decision and not a copy
# of its siblings' habit. The freshest sibling landmine
# (component-source-vocabulary-check, 2026-08-22) says "the fix is a MOUNT, not a
# COPY, whenever the file is meant to change", and this registry IS meant to
# change: ruling the `unruled` backlog down is phase 3's whole deliverable. It is
# still copied, for two reasons that point the other way here:
#
#   1. A mounted ConfigMap goes stale INDEFINITELY — until a human remembers
#      `apply -k`. A copy goes stale only until the next fleet release, and this
#      service is in RELEASE_IMAGES precisely so that is hours, not for ever. The
#      estate's own stated remedy for a shipped file rotting is release
#      membership (live-declaration-drift-check.dockerfile).
#   2. `make build-*` builds from committed HEAD, so the copy is BY CONSTRUCTION
#      the reviewed version. A mount ships whatever is in the working tree at
#      `apply -k` time — which for a file whose whole job is to say which findings
#      are accepted means an unreviewed declaration could be applied with no diff
#      to review. That is the argument the three acks-file images already make.
#
# The residual — a declaration landing in the repo before the next release — is
# made VISIBLE rather than silent: the run summary states how many codes the
# registry it graded against declares, and the commit this binary was built from.
# A cluster row disagreeing with a local run on that number IS the staleness.
#
# ⚠ IT RUNS --no-source, and that is deliberate. Two of the mode's arms open the
# Go file a `consumed` entry names. This image has no repo, so without the flag
# all five `consumed` entries raise `reader-unreadable` and the job is RED EVERY
# DAY against a healthy registry (measured 2026-08-23: 5 findings, exit 1). Those
# arms grade the registry against source, both halves change only by commit, and
# they now run at commit time instead — scripts/check-finding-code-registry.sh,
# wired into .githooks/pre-commit. Every run says which arms it skipped.
#
# CONNECTS TO POSTGRES DIRECTLY — never kubectl exec. The ai-persona-app service
# account has no pods/exec RBAC in this namespace, so a kubectl-only tool fails
# there in a way that looks like a CLEAN RUN unless you are watching exit codes.
# See cmd/config-key-audit/fleetdb.go.
#
# IT REPORTS ON EVERY RUN, INCLUDING CLEAN ONES. A check that only speaks when it
# fails is INDISTINGUISHABLE FROM ONE THAT HAS STOPPED RUNNING, so a missing
# doc_notes row means THE JOB DID NOT RUN and can never be read as "nothing is
# wrong". The literal is in the writer (findingcodes.go), not guessed:
#   SELECT created_at, left(body, 600) FROM doc_notes
#    WHERE source = 'finding-code-registry-check' ORDER BY created_at DESC LIMIT 7;
#
# EXIT CODES: 0 every observed code is declared (the `unruled` BACKLOG is reported,
# not a finding — a check that fails from day one over a pre-existing backlog is a
# check that gets ignored); 1 an observed code is undeclared, a declaration does
# not hold up, the retention sweep disagrees with the registry, or the unruled
# backlog has grown past its cap; 2 the check could not run, which must never read
# as a pass — including a zero-row read of a table that is never empty in practice.
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG GIT_COMMIT=unknown
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-X github.com/gqls/agentchassis/pkg/buildinfo.GitCommit=${GIT_COMMIT}" \
    -o config-key-audit ./cmd/config-key-audit

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/config-key-audit /app/
# THE REGISTRY TRAVELS WITH THE BINARY — see the header for why this is a COPY and
# not a mount, and note the trap it is avoiding in the other direction: a check
# whose data file is left in the builder stage deploys clean, reports success at
# every layer and cannot run (LANDMINES, component-source-vocabulary-check
# v1.0.1326). Probe the FILE, not just the flag:
#   docker run --rm --entrypoint sh <image>:<tag> -c 'ls -l /app/finding_code_registry.json'
COPY --from=builder /app/docs/agent_docs/docs024_key_docs_latest/architecture_review/finding_code_registry.json /app/
RUN chown -R appuser:appgroup /app
USER appuser
CMD ["./config-key-audit", "--finding-codes", "--report", "--no-source", \
     "--registry", "/app/finding_code_registry.json"]
