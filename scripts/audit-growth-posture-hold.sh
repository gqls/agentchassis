#!/usr/bin/env bash
# audit-growth-posture-hold.sh — hand-run twin of the growth-posture-hold-check CronJob.
#
# Runs THE SAME check.py the cluster runs (one file, no copy to drift), through
# `kubectl exec` instead of the in-cluster DSN, and writes NO doc_notes row unless you
# pass --write. Exit 0 clean, 1 findings, 2 refused to look.
#
#   scripts/audit-growth-posture-hold.sh              # report to stdout
#   scripts/audit-growth-posture-hold.sh --days 14    # a different threshold for this run
#   scripts/audit-growth-posture-hold.sh --write      # also record the doc_notes row
#   scripts/audit-growth-posture-hold.sh --self-test  # fixtures only, no cluster
#
# What it reads and why: register WDS-020; the check's own docstring.
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
exec python3 "$here/deployments/kustomize/services/growth-posture-hold-check/base/check.py" --local "$@"
