#!/usr/bin/env bash
# Thunder SSH probe — answers: does our public_key authenticate, what's the SSH port,
# and does DIRECT ssh work (vs needing Thunder's proxy)? Provisions, SSHes, deletes.
set -euo pipefail

TOKEN="a7383d1885fa832cc6f30674af5e327ddc9fd1ca0c912eb5e0ac1ce925b12596"
BASE="https://api.thundercompute.com:8443/v1"
AUTH=(-H "Authorization: Bearer ${TOKEN}")

WORK="$(mktemp -d)"
KEY="${WORK}/probe_ed25519"
echo "==> Generating ed25519 keypair (same type the adapter uses)"
ssh-keygen -t ed25519 -N "" -C "thunder-ssh-probe" -f "${KEY}" >/dev/null
PUB="$(cat "${KEY}.pub")"

echo "==> Creating instance with OUR public_key"
CREATE=$(curl -s "${AUTH[@]}" -H "Content-Type: application/json" \
  -X POST "${BASE}/instances/create" \
  -d "{\"gpu_type\":\"a100xl\",\"num_gpus\":1,\"cpu_cores\":4,\"disk_size_gb\":100,\"mode\":\"prototyping\",\"template\":\"base\",\"public_key\":\"${PUB}\"}")
echo "    create response: ${CREATE}"
ID=$(echo "${CREATE}" | python3 -c 'import sys,json;print(json.load(sys.stdin)["identifier"])')
echo "    identifier=${ID}"

cleanup() {
  echo "==> Deleting instance ${ID}"
  curl -s "${AUTH[@]}" -X POST "${BASE}/instances/${ID}/delete" >/dev/null && echo "    deleted"
  rm -rf "${WORK}"
}
trap cleanup EXIT

echo "==> Polling for RUNNING + ip + port"
IP=""; PORT=""
for i in $(seq 1 60); do
  LIST=$(curl -s "${AUTH[@]}" "${BASE}/instances/list")
  read -r STATUS IP PORT < <(echo "${LIST}" | python3 -c "
import sys,json
d=json.load(sys.stdin)
inst=d.get('${ID}',{})
print(inst.get('status',''), inst.get('ip',''), inst.get('port',''))
")
  echo "    [${i}] status=${STATUS} ip=${IP} port=${PORT}"
  if [ "${STATUS}" = "RUNNING" ] && [ -n "${IP}" ] && [ "${PORT}" != "0" ] && [ -n "${PORT}" ]; then
    break
  fi
  sleep 5
done

if [ "${STATUS}" != "RUNNING" ]; then echo "FAILED: never reached RUNNING"; exit 1; fi

echo "==> Attempting DIRECT ssh to ${IP}:${PORT} as ubuntu with our private key"
echo "    (give the box a few seconds for sshd)"
sleep 8
set +e
ssh -i "${KEY}" \
    -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
    -o ConnectTimeout=12 -o BatchMode=yes \
    -p "${PORT}" "ubuntu@${IP}" 'echo SSH_OK; whoami; nvidia-smi -L; python3 --version' 2>&1
RC=$?
set -e
echo "==> ssh exit code: ${RC}"
if [ "${RC}" -ne 0 ]; then
  echo "    DIRECT ssh failed. Trying common alt user 'root'..."
  set +e
  ssh -i "${KEY}" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
      -o ConnectTimeout=12 -o BatchMode=yes \
      -p "${PORT}" "root@${IP}" 'echo SSH_OK_ROOT; whoami' 2>&1
  echo "    root attempt exit code: $?"
  set -e
fi
echo "==> Probe done (instance will be deleted by trap)"
