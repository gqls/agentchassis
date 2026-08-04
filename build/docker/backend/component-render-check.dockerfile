FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o component-render-check ./cmd/component-render-check

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/component-render-check /app/
# The baseline is go:embed-ed INTO THE BINARY (see loadBaseline) — nothing to COPY.
#
# It is the check's own definition of "already known", so it is made inseparable
# from the code that interprets it. A baseline in a ConfigMap, or COPYed in from
# elsewhere, could be edited to silence a finding with no diff anybody reviews —
# the quiet unreviewed clearing this whole lane exists to close. Banking a real
# fix is therefore: --write-baseline, commit the diff, rebuild. Visible and
# revertable. `make build-component-render-check` builds from committed HEAD, so
# the image can never carry an uncommitted baseline.
RUN chown -R appuser:appgroup /app
USER appuser
CMD ["./component-render-check", "--compare", "--report"]
