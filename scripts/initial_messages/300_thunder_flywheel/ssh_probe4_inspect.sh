#!/usr/bin/env bash
# Probe v4 — our key was rejected as root. Get on the box via THUNDER's key
# (the one tnr connect uses, which works) and inspect WHERE our public key landed
# and which users exist. Then we know whether ssh_exec needs a different user,
# or whether Thunder isn't honouring our public_key at all.
#
# Usage: bash ssh_probe4_inspect.sh <thunder_numeric_id> <db_row_id>
set -uo pipefail
TID="${1:?need thunder numeric id}"
DBID="${2:?need db_row_id (for our key comparison)}"

CONN=$(tnr connect "${TID}" --json -y 2>/dev/null)
IP=$(echo "${CONN}"   | python3 -c 'import sys,json;print(json.load(sys.stdin)["ip"])')
PORT=$(echo "${CONN}" | python3 -c 'import sys,json;print(json.load(sys.stdin)["port"])')
TKEY=$(echo "${CONN}" | python3 -c 'import sys,json;print(json.load(sys.stdin)["key_file"])')
echo "==> Thunder key: ${TKEY}  target: root@${IP}:${PORT}"

# Our public key (what the adapter sent) — derive from the secret's private half
OURKEY=$(mktemp)
kubectl -n ai-persona-system get secret "thunder-ssh-${DBID}" -o jsonpath='{.data.private_key}' | base64 -d > "${OURKEY}"
chmod 600 "${OURKEY}"
OURPUB=$(ssh-keygen -y -f "${OURKEY}" 2>/dev/null)
echo "==> Our public key (adapter-generated, what we TRIED to auth with):"
echo "    ${OURPUB}"

echo
echo "==> Logging in via THUNDER's key to inspect the box:"
ssh -i "${TKEY}" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
    -o ConnectTimeout=12 -o BatchMode=yes -p "${PORT}" "root@${IP}" bash -s <<'REMOTE'
echo "=== whoami / users ==="
whoami
echo "--- /etc/passwd users with shells ---"
grep -E '/(bash|sh)$' /etc/passwd | cut -d: -f1,6,7
echo "=== root authorized_keys ==="
cat /root/.ssh/authorized_keys 2>/dev/null || echo "(none)"
echo "=== ubuntu authorized_keys (if user exists) ==="
cat /home/ubuntu/.ssh/authorized_keys 2>/dev/null || echo "(no ubuntu user / no keys)"
echo "=== any other home dirs with authorized_keys ==="
for d in /home/*; do [ -f "$d/.ssh/authorized_keys" ] && echo "$d:" && cat "$d/.ssh/authorized_keys"; done
echo "=== sshd: which users/auth allowed ==="
grep -iE 'PermitRootLogin|PubkeyAuthentication|AllowUsers|PasswordAuthentication' /etc/ssh/sshd_config 2>/dev/null | grep -v '^#'
REMOTE
echo "    (remote inspect exit=$?)"
rm -f "${OURKEY}"
echo
echo "==> COMPARE: is OUR public key (above) present in any authorized_keys on the box?"
echo "    If yes but under a non-root user → ssh_exec must use THAT user."
echo "    If our key is NOWHERE → Thunder isn't honouring public_key; we must capture Thunder's key."
