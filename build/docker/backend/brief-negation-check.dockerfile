FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o brief-negation-check ./cmd/brief-negation-check

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/brief-negation-check /app/
RUN chown -R appuser:appgroup /app
USER appuser
# NO FLAGS. Writing (file a finding, close answered ones, record the run in
# doc_notes) is the DEFAULT; --dry-run is what suppresses it. A flag the CronJob
# must remember to PASS is a silence waiting to happen — the sibling detector
# shipped that way and was inert by omission until a council seat noticed.
#
# A GO IMAGE, NOT A PYTHON ConfigMap SCRIPT (the choice its Python siblings made),
# for one reason: half the question is what the WRITER'S PROMPT reads, and the
# other half is the define-by-negation scanner. Both are compiled-in Go
# (datahelpers.ScanDefineByNegation, and the {{.site_specs.specs.…}} parse), and a
# mirrored copy of either would go stale exactly when the prompt or the scanner
# changes — which is the only moment this check matters.
CMD ["./brief-negation-check"]
