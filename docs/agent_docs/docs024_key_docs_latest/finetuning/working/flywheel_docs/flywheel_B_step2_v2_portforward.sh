#!/bin/bash
# ============================================================================
# Flywheel B — Step 2 (v2): real content + real embeddings + real retrieval
# ============================================================================
# Uses kubectl port-forward rather than spawning a curl pod per embedding.
# All 5 embeddings go through a single background port-forward tunnel.
#
# Requires: kubectl, jq, curl, bash. Runs from local machine.
# Transaction is rolled back at the end — nothing persists.
# ============================================================================

set -e

NS=ai-persona-system
LOCAL_PORT=21434                    # uncommon, almost certainly free
PG_DEPLOY=deploy/postgres-clients

# ── Deps ──
for cmd in kubectl jq curl; do
    if ! command -v $cmd &> /dev/null; then
        echo "ERROR: '$cmd' is required. Install it and retry."
        exit 1
    fi
done

# ── Port-forward to ollama-adapter ──
echo "Starting port-forward: localhost:$LOCAL_PORT → ollama-adapter:11434 ..."
kubectl -n $NS port-forward svc/ollama-adapter $LOCAL_PORT:11434 > /tmp/pf.log 2>&1 &
PF_PID=$!

cleanup() {
    if kill -0 $PF_PID 2>/dev/null; then
        kill $PF_PID 2>/dev/null || true
        wait $PF_PID 2>/dev/null || true
    fi
}
trap cleanup EXIT INT TERM

# Wait for tunnel to be ready (up to 10s)
for i in $(seq 1 20); do
    if curl -s -o /dev/null -w '%{http_code}' --max-time 1 \
        "http://localhost:$LOCAL_PORT/api/tags" 2>/dev/null | grep -q '200'; then
        echo "  tunnel up"
        break
    fi
    sleep 0.5
    if [ $i -eq 20 ]; then
        echo "ERROR: port-forward didn't become ready in 10s"
        cat /tmp/pf.log
        exit 1
    fi
done

# ── Helper: fetch embedding via the local tunnel ──
fetch_embedding() {
    local text="$1"
    local payload
    payload=$(jq -nc --arg p "$text" '{model:"nomic-embed-text", prompt:$p}')
    curl -s --max-time 30 \
        -X POST "http://localhost:$LOCAL_PORT/api/embeddings" \
        -H 'Content-Type: application/json' \
        -d "$payload" \
        | jq -c '.embedding'
}

# ── Content pieces ──
C1="French Bulldogs are brachycephalic breeds with shortened skulls. They are prone to Brachycephalic Obstructive Airway Syndrome (BOAS) which affects up to 50 percent of the breed and often requires surgical intervention such as soft palate resection to improve breathing."
C2="Labrador Retrievers are popular family dogs known for their friendly temperament and high energy levels. They require regular exercise and respond well to positive reinforcement training methods."
C3="A grand piano produces sound through strings struck by felt hammers when keys are pressed. The instrument has 88 keys covering over seven octaves and is the standard for classical concert performance."
C4="Electric vehicles use rechargeable lithium-ion batteries to power an electric motor. Modern EVs can achieve ranges exceeding 300 miles on a single charge and charge fully overnight at home."

QUERY="dog breed airway problems breathing difficulty"

echo ""
echo "Fetching embeddings ..."
E1=$(fetch_embedding "$C1");   echo "  [1/5] French Bulldog content"
E2=$(fetch_embedding "$C2");   echo "  [2/5] Labrador content"
E3=$(fetch_embedding "$C3");   echo "  [3/5] Piano content"
E4=$(fetch_embedding "$C4");   echo "  [4/5] EV content"
EQ=$(fetch_embedding "$QUERY"); echo "  [5/5] Query embedded"

# ── Sanity check before sending to postgres ──
for var_name in E1 E2 E3 E4 EQ; do
    val="${!var_name}"
    if [ -z "$val" ] || [ "$val" = "null" ]; then
        echo "ERROR: $var_name is empty/null. Ollama likely returned an error."
        exit 1
    fi
done

echo ""
echo "All 5 embeddings received. Running insert + similarity query ..."
echo ""

# ── Run SQL via postgres-clients exec ──
kubectl -n $NS exec -i $PG_DEPLOY -- \
    psql -U clients_user -d clients_db <<EOF
BEGIN;

INSERT INTO knowledge_base (collection, title, content, embedding, embedding_model, metadata) VALUES
    ('flywheel_b_real', 'French Bulldog health',  '$C1', '$E1'::vector, 'nomic-embed-text', '{"test":true,"topic":"dog_brachycephalic"}'::jsonb),
    ('flywheel_b_real', 'Labrador temperament',   '$C2', '$E2'::vector, 'nomic-embed-text', '{"test":true,"topic":"dog_breed_general"}'::jsonb),
    ('flywheel_b_real', 'Grand piano',            '$C3', '$E3'::vector, 'nomic-embed-text', '{"test":true,"topic":"music"}'::jsonb),
    ('flywheel_b_real', 'Electric vehicles',      '$C4', '$E4'::vector, 'nomic-embed-text', '{"test":true,"topic":"transport"}'::jsonb);

\echo '=== Ranking for query: "dog breed airway problems breathing difficulty" ==='
\echo ''

SELECT title,
       LEFT(content, 55) as preview,
       ROUND((1 - (embedding <=> '$EQ'::vector))::numeric, 4) as similarity
FROM knowledge_base
WHERE collection = 'flywheel_b_real'
ORDER BY embedding <=> '$EQ'::vector;

ROLLBACK;
EOF

echo ""
echo "Done. Expected: 'French Bulldog health' top, music + transport bottom."
