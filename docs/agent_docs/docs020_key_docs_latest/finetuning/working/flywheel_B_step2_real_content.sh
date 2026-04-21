#!/bin/bash
# ============================================================================
# Flywheel B — Step 2: Real content + real embeddings + real retrieval
# ============================================================================
# Still bypasses the chassis / workflow layer. Proves the infrastructure
# triangle (Ollama → pgvector → similarity search) works on real text, not
# just synthetic vectors.
#
# Requires: kubectl, jq, bash. Runs from local machine.
# Transaction is rolled back at the end — nothing persists.
# ============================================================================

set -e

NS=ai-persona-system
OLLAMA_SVC="ollama-adapter.ai-persona-system.svc.cluster.local:11434"

# Check dependencies
if ! command -v jq &> /dev/null; then
    echo "ERROR: jq is required. Install with 'brew install jq' or 'apt install jq'"
    exit 1
fi

# ── Helper: fetch embedding for a single string via a throwaway curl pod ──
fetch_embedding() {
    local text="$1"
    # Unique pod name per call so we can run in quick succession
    local pod_name="embed-probe-$(date +%s)-$RANDOM"
    local payload
    payload=$(jq -nc --arg p "$text" '{model:"nomic-embed-text", prompt:$p}')

    kubectl -n $NS run "$pod_name" --rm -i --quiet --restart=Never \
        --image=curlimages/curl:latest -- \
        -s -X POST "http://$OLLAMA_SVC/api/embeddings" \
        -H 'Content-Type: application/json' \
        -d "$payload" \
        | jq -c '.embedding'
}

# ── Four pieces of content — one brachycephalic-dog topic, others unrelated ──
C1="French Bulldogs are brachycephalic breeds with shortened skulls. They are prone to Brachycephalic Obstructive Airway Syndrome (BOAS) which affects up to 50 percent of the breed and often requires surgical intervention such as soft palate resection to improve breathing."
C2="Labrador Retrievers are popular family dogs known for their friendly temperament and high energy levels. They require regular exercise and respond well to positive reinforcement training methods."
C3="A grand piano produces sound through strings struck by felt hammers when keys are pressed. The instrument has 88 keys covering over seven octaves and is the standard for classical concert performance."
C4="Electric vehicles use rechargeable lithium-ion batteries to power an electric motor. Modern EVs can achieve ranges exceeding 300 miles on a single charge and charge fully overnight at home."

QUERY="dog breed airway problems breathing difficulty"

echo "Fetching embeddings (5 pods will spin up in sequence, ~30s total)..."
E1=$(fetch_embedding "$C1");  echo "  [1/5] French Bulldog content"
E2=$(fetch_embedding "$C2");  echo "  [2/5] Labrador content"
E3=$(fetch_embedding "$C3");  echo "  [3/5] Piano content"
E4=$(fetch_embedding "$C4");  echo "  [4/5] EV content"
EQ=$(fetch_embedding "$QUERY"); echo "  [5/5] Query embedded"

# Sanity-check the embeddings look right before we send SQL
echo ""
echo "Embedding sizes (first 80 chars of each):"
echo "  C1: $(echo "$E1" | head -c 80)..."
echo "  EQ: $(echo "$EQ" | head -c 80)..."

# ── Build and run SQL ──
echo ""
echo "Inserting and querying..."
echo ""

kubectl -n $NS exec -i deploy/postgres-clients -- \
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
echo "Done. Expected: 'French Bulldog health' highest similarity, music and transport lowest."
