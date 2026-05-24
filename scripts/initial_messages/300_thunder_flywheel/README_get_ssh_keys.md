Step 1 — provision

CORRELATION=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
echo "REQ=$REQ"
kubectl -n kafka run kcat-prov-$(date +%s) --rm -i --restart=Never --image=edenhill/kcat:1.7.1 -- \
kcat -P -c 1 -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
-t system.adapter.thunder.requests \
-H "correlation_id=$CORRELATION" -H "request_id=$REQ" -H "message_type=request" <<'JSON'
{"body":{"action":"provision_instance","gpu":"a100","mode":"prototyping","num_gpus":1,"reply_to_topic":"system.agent.generic.responses"}}
JSON

Step 2 — get the new db_row_id and thunder id (wait for "Reaper deadline set" which means INSERT succeeded):

kubectl -n ai-persona-system logs deploy/thunder-adapter --since=2m | grep -iE "Reaper deadline set|db_row_id|thunder_identifier|Provision complete|duplicate key"

Step 3 — the ubuntu-key test. Paste these lines DIRECTLY (not inside bash, not in a heredoc). Substitute the new db_row_id for DBID and the thunder id (probably 0):

DBID="<new-db_row_id-from-step-2>"
CONN=$(tnr connect 0 --json -y 2>/dev/null)
IP=$(echo "$CONN" | python3 -c 'import sys,json;print(json.load(sys.stdin)["ip"])')
PORT=$(echo "$CONN" | python3 -c 'import sys,json;print(json.load(sys.stdin)["port"])')
OURKEY=$(mktemp)
kubectl -n ai-persona-system get secret "thunder-ssh-${DBID}" -o jsonpath='{.data.private_key}' | base64 -d > "$OURKEY"
chmod 600 "$OURKEY"
ssh -i "$OURKEY" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=12 -o BatchMode=yes -p "$PORT" "ubuntu@$IP" 'echo OUR_KEY_AS_UBUNTU_OK; whoami; nvidia-smi -L'
echo "exit=$?"
rm -f "$OURKEY"
