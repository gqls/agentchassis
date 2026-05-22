#!/usr/bin/env bash
# Probe v2 — tests whether DIRECT ssh works once we use the RIGHT user (root)
# and a freshly-resolved port, AND whether OUR key works vs Thunder's own key.
# Uses the EXISTING manual instance (no new provision). Pass the instance id.
set -uo pipefail
ID="${1:-0}"
TOKEN="a7383d1885fa832cc6f30674af5e327ddc9fd1ca0c912eb5e0ac1ce925b12596"
BASE="https://api.thundercompute.com:8443/v1"

echo "==> What does tnr connect report right now (authoritative port)?"
CONN=$(tnr connect "${ID}" --json -y 2>/dev/null)
echo "${CONN}"
IP=$(echo "${CONN}"   | python3 -c 'import sys,json;print(json.load(sys.stdin)["ip"])')
PORT=$(echo "${CONN}" | python3 -c 'import sys,json;print(json.load(sys.stdin)["port"])')
TKEY=$(echo "${CONN}" | python3 -c 'import sys,json;print(json.load(sys.stdin)["key_file"])')
echo "    ip=${IP} port=${PORT} thunder_key=${TKEY}"

echo
echo "==> Test 1: direct ssh as ROOT with THUNDER's own key (baseline — should work)"
ssh -i "${TKEY}" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
    -o ConnectTimeout=12 -o BatchMode=yes -p "${PORT}" "root@${IP}" \
    'echo THUNDER_KEY_OK; whoami; nvidia-smi -L' 2>&1
echo "    exit=$?"

echo
echo "==> Test 2: does the instance also have OUR adapter-style key installed?"
echo "    (only meaningful if this instance was made WITH a public_key — the manual"
echo "     'tnr create' was NOT, so this should FAIL, confirming manual instances"
echo "     only carry Thunder's key. The adapter sends public_key, so adapter"
echo "     instances differ — test that separately via a real adapter provision.)"
echo "    Listing authorized_keys on the box (via Thunder key) to see what's installed:"
ssh -i "${TKEY}" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
    -o ConnectTimeout=12 -o BatchMode=yes -p "${PORT}" "root@${IP}" \
    'echo "--- /root/.ssh/authorized_keys ---"; cat /root/.ssh/authorized_keys 2>/dev/null; echo "--- ubuntu user? ---"; id ubuntu 2>&1; ls -la /home/ 2>/dev/null' 2>&1
echo "    exit=$?"
