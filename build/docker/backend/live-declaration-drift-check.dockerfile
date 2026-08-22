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
# NO ACKS FILE, deliberately — unlike commit-sha-exposure-check and
# optional-explicit-wires-check. A drifted declaration is never an acknowledgeable
# exception: either the live object is wrong or the declaration is, and one of the
# two gets fixed. An acks file here would be a supported way to record "yes, the
# database and the code disagree, and that is fine", which is the state this whole
# check exists to make impossible to hold quietly.
#
# THE DECLARATIONS ARE COMPILED IN. platform/livespec is Go, so the image and the
# spec cannot disagree — and because `make build-*` builds from committed HEAD, an
# unreviewed change to a declaration is unshippable rather than merely discouraged.
# The corollary is the ordinary one for this family: a STALE image is a stale
# SPEC, so this service must be in RELEASE_IMAGES (makefile) or its clean reports
# slowly stop meaning what they say.
#
# EXIT CODES — the middle one is the whole point:
#   0  every declaration matched its live object, and the line says how many were read
#   1  at least one live object has parted from its declaration
#   2  the check could not LOOK: no database, an unreadable probe, a NULL or
#      missing row, or an empty registry. Never a clean report. An auditor that
#      reports success when it could not read is the defect bugs_open/363 is about,
#      and reproducing it inside the fix would be the worst possible outcome.
CMD ["./config-key-audit", "--live-declaration-drift", "--report"]
