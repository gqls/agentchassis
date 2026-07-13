#!/usr/bin/env bash
# Probe v3 — the LAST SSH unknown: does OUR adapter-generated key authenticate?
# Run AFTER an adapter provision (not a manual tnr create). It:
#   1. takes the db_row_id (provisioning_id) of an adapter-provisioned instance
#   2. pulls OUR private key from the k8s secret thunder-ssh-<db_row_id>
#   3. resolves the SSH port via tnr connect --json (authoritative port)
#   4. waits for sshd (like tnr connect does), then SSHes as root with OUR key
#
# Usage: bash ssh_probe3_adapter_key.sh <db_row_id> <thunder_numeric_id>
#   e.g. bash ssh_probe3_adapter_key.sh a30437cf-1f38-40b9-b33a-dbeff1a4edb7 0
set -uo pipefail
DBID="${1:?need db_row_id (provisioning_id / k8s secret suffix)}"
TID="${2:?need thunder numeric id for tnr connect}"
NS="ai-persona-system"

echo "==> Pulling OUR adapter-generated private key from k8s secret"
SECRET="thunder-ssh-${DBID}"
KEY="$(mktemp)"
kubectl -n "${NS}" get secret "${SECRET}" -o jsonpath='{.data.private_key}' | base64 -d > "${KEY}"
chmod 600 "${KEY}"
echo "    secret=${SECRET}  (key bytes: $(wc -c < "${KEY}"))"

echo "==> Resolving authoritative SSH ip:port via tnr connect --json"
CONN=$(tnr connect "${TID}" --json -y 2>/dev/null)
echo "    ${CONN}"
IP=$(echo "${CONN}"   | python3 -c 'import sys,json;print(json.load(sys.stdin)["ip"])')
PORT=$(echo "${CONN}" | python3 -c 'import sys,json;print(json.load(sys.stdin)["port"])')
echo "    ip=${IP} port=${PORT}"

echo "==> Waiting for sshd to accept (like tnr connect's 'Waiting for SSH service')"
for i in $(seq 1 30); do
  if timeout 3 bash -c "cat < /dev/null > /dev/tcp/${IP}/${PORT}" 2>/dev/null; then
    echo "    [${i}] port open"; break
  fi
  echo "    [${i}] not yet"; sleep 3
done

echo "==> SSH as ROOT with OUR key"
ssh -i "${KEY}" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
    -o ConnectTimeout=12 -o BatchMode=yes -p "${PORT}" "root@${IP}" \
    'echo OUR_KEY_OK; whoami; nvidia-smi -L; python3 --version; nproc; df -h / | tail -1' 2>&1
echo "    ssh exit=$?"
rm -f "${KEY}"
echo "==> done. Remember to decommission the instance (via adapter decommission or tnr delete + reconcile)."
