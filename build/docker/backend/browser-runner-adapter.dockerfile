# FILE: build/docker/backend/browser-runner-adapter.dockerfile
#
# Browser-runner adapter (Tier-4 headless acceptance runner, P0).
#
# DELIBERATE DIFFERENCE from the other adapter images: the runtime is
# debian-slim with Chromium + the Playwright driver baked in, not alpine.
# Chromium's glibc/font/lib dependencies rule alpine out, and installing the
# browser at container start would add ~30s and a network dependency to every
# pod restart. Expect a ~1.2GB image — that is the cost of a real browser and
# is normal for Playwright images.
#
# The driver and browsers are installed as root into shared world-readable
# paths (XDG_CACHE_HOME + PLAYWRIGHT_BROWSERS_PATH), then the container runs
# as the unprivileged appuser with the same env vars pointing at them —
# playwright-go's playwright.Run() resolves both from the env at runtime.
# Keep the CLI version pinned to the go.mod version of playwright-go.

FROM golang:1.24-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o browser-runner-adapter ./cmd/browser-runner-adapter
# The playwright CLI (same module version as the library) installs the driver
# + Chromium in the runtime stage. Import path is the module's DECLARED path —
# v0.6100.0's go.mod still says mxschmitt (upstream release accident); this
# version is REQUIRED because earlier tags download the driver from the
# decommissioned playwright.azureedge.net CDN (404s since 2026). v0.6100.0
# fetches the driver npm package from registry.npmjs.org and a Node runtime
# from nodejs.org/dist instead — both live.
RUN go build -o /playwright-cli github.com/mxschmitt/playwright-go/cmd/playwright

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

# Shared, world-readable homes for the Playwright driver and browsers.
# playwright-go v0.6100.0 resolves the DRIVER directory from $HOME/.cache on
# Linux (getDefaultCacheDirectory in run.go — it IGNORES XDG_CACHE_HOME), so
# HOME is the knob, not XDG_CACHE_HOME. Docker ENV persists across the later
# `USER appuser` line, so install (as root) and runtime (as appuser) resolve
# the SAME /pw-home/.cache/ms-playwright-go path. Browsers honour
# PLAYWRIGHT_BROWSERS_PATH as usual.
ENV HOME=/pw-home \
    PLAYWRIGHT_BROWSERS_PATH=/pw-browsers

COPY --from=builder /playwright-cli /usr/local/bin/playwright-cli
# --with-deps pulls Chromium's apt dependencies; the browsers land in
# /pw-browsers and the driver (node runtime + npm package) in
# /pw-home/.cache/ms-playwright-go/<version>.
RUN playwright-cli install --with-deps chromium && \
    rm -rf /var/lib/apt/lists/*

# appuser owns its HOME (Chromium writes crashpad/config scratch there);
# browsers stay root-owned, world-readable — nothing writes into them.
RUN groupadd -r appgroup && useradd -r -g appgroup -m appuser && \
    chown -R appuser:appgroup /pw-home && \
    chmod -R a+rX /pw-browsers
WORKDIR /app
COPY --from=builder /app/browser-runner-adapter /app/
COPY configs/browser-runner-adapter.yaml /app/configs/
RUN chown -R appuser:appgroup /app
USER appuser
CMD ["./browser-runner-adapter", "-config", "configs/browser-runner-adapter.yaml"]
