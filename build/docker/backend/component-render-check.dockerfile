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
# The baseline travels IN THE IMAGE, deliberately.
#
# It is the check's own definition of "already known", so it must be versioned
# with the code that interprets it — a baseline in a ConfigMap could be edited
# to silence a finding with no diff anybody reviews, which is the failure mode
# this whole lane exists to close. Baking it in means banking a fix is a commit
# plus a rebuild: visible, reviewable, revertable. `make build-component-render-check`
# builds from committed HEAD, so the image can never carry an uncommitted baseline.
COPY docs/agent_docs/docs024_key_docs_latest/bugfix_140_contact_info_fabrication/baseline/component_render_check_baseline.json /app/baseline.json
RUN chown -R appuser:appgroup /app
USER appuser
CMD ["./component-render-check", "--baseline", "/app/baseline.json", "--report"]
