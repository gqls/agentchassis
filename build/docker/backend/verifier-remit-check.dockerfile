FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o verifier-remit-check ./cmd/verifier-remit-check

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/verifier-remit-check /app/
RUN chown -R appuser:appgroup /app
USER appuser
# NO FLAGS. The writing path (file a work item, close answered ones, record the
# run in doc_notes) is the DEFAULT, and --dry-run is what suppresses it.
#
# That inversion is deliberate. The sibling detector this job is modelled on
# shipped with its writing behind a flag and was inert-by-omission until a council
# seat noticed — "confirm it's actually invoked somewhere, or the detector is
# inert-by-omission the same way the field is inert-by-design". A flag the CronJob
# must remember to pass is a silence waiting to happen; a flag it must remember NOT
# to pass cannot silence anything.
#
# The Go binary is linked against the LIVE verifier registry, which is the whole
# reason this is not a Python job with a mirrored list of which item types have
# verifiers and which declare a remit: that list would go stale exactly when a new
# verifier lands, i.e. at the only moment the check matters.
CMD ["./verifier-remit-check"]
