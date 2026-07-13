#!/usr/bin/env bash
# Capture what `tnr connect` does server-side BEFORE ssh works.
# Runs an interactive connect under strace, writing to a file, and
# auto-exits the remote shell after a moment so nothing hangs.
# Usage: bash trace_connect.sh <instance_id>
set -uo pipefail
ID="${1:-0}"
OUT="/tmp/tnr_connect_trace.txt"

echo "==> Tracing 'tnr connect ${ID}' (execve + network syscalls) → ${OUT}"
echo "    Feeding 'exit' to the remote shell so it won't hang."
# -f follow forks; trace execve (what it runs) + connect (outbound sockets it opens)
# Feed 'exit\n' on stdin so the interactive ssh shell closes itself.
printf 'exit\n' | timeout 40 strace -f -e trace=execve,connect -o "${OUT}" tnr connect "${ID}" -y 2>&1 | head -20

echo
echo "==> execve lines (what tnr actually runs — ssh invocation + any helpers):"
grep -i 'execve' "${OUT}" | grep -iE 'ssh|tnr|proxy|sh"|python|node' | head -30
echo
echo "==> outbound connect() calls (Thunder API endpoints it hits before ssh):"
grep -i 'connect(' "${OUT}" | grep -vE '127.0.0.1|AF_UNIX|AF_NETLINK' | head -30
echo
echo "==> full trace saved at ${OUT} (read it if the greps miss something)"
