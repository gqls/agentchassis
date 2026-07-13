#!/bin/bash
set -euo pipefail

# ── Registration ──────────────────────────────────────────────
# Two modes:
#   GITHUB_PAT set        → fetches a short-lived registration token via API (preferred)
#   GITHUB_REG_TOKEN set  → uses the token directly (manual, expires in 1 hour)

REPO_URL="https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}"

if [ -n "${GITHUB_PAT:-}" ]; then
    echo "Fetching registration token via PAT..."
    REG_TOKEN=$(curl -s -X POST \
        -H "Authorization: token ${GITHUB_PAT}" \
        -H "Accept: application/vnd.github+json" \
        "https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/actions/runners/registration-token" \
        | jq -r .token)

    if [ "$REG_TOKEN" = "null" ] || [ -z "$REG_TOKEN" ]; then
        echo "ERROR: Failed to fetch registration token. Check GITHUB_PAT permissions."
        exit 1
    fi
elif [ -n "${GITHUB_REG_TOKEN:-}" ]; then
    REG_TOKEN="${GITHUB_REG_TOKEN}"
else
    echo "ERROR: Set either GITHUB_PAT or GITHUB_REG_TOKEN"
    exit 1
fi

# Only configure if not already configured
if [ ! -f .runner ]; then
    echo "Configuring runner for ${REPO_URL}..."
    ./config.sh \
        --url "${REPO_URL}" \
        --token "${REG_TOKEN}" \
        --name "${RUNNER_NAME:-k8s-runner}" \
        --labels "${RUNNER_LABELS:-self-hosted,linux,x64}" \
        --unattended \
        --replace
fi

# ── Cleanup on exit ───────────────────────────────────────────
# Deregister when the pod stops so GitHub's runner list stays clean.
cleanup() {
    echo "Removing runner registration..."
    if [ -n "${GITHUB_PAT:-}" ]; then
        REMOVE_TOKEN=$(curl -s -X POST \
            -H "Authorization: token ${GITHUB_PAT}" \
            -H "Accept: application/vnd.github+json" \
            "https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/actions/runners/remove-token" \
            | jq -r .token)
        ./config.sh remove --token "${REMOVE_TOKEN}" || true
    else
        ./config.sh remove --token "${REG_TOKEN}" || true
    fi
}
trap cleanup EXIT SIGTERM SIGINT

# ── Run ───────────────────────────────────────────────────────
echo "Starting runner..."
./run.sh
