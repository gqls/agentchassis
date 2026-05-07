#!/bin/bash
# ============================================================================
# 01_pull_dataset_from_postgres.sh
# ============================================================================
# Pulls a training dataset out of training_exports.rows into a local JSONL
# file ready for Unsloth consumption.
#
# Runs from the GPU VM. Needs kubectl access to the ai-persona-system cluster
# OR direct psql access to the clients_db (adjust connection args accordingly).
#
# Usage:
#   ./01_pull_dataset_from_postgres.sh <export_id_uuid> [output_path]
#
# Example:
#   ./01_pull_dataset_from_postgres.sh 146a9a12-c953-48eb-bf1f-c1856e5f13b7 \
#       /workspace/training_iter0.jsonl
# ============================================================================

set -euo pipefail

EXPORT_ID="${1:-}"
OUT="${2:-/workspace/training_iter0.jsonl}"

if [ -z "$EXPORT_ID" ]; then
    echo "Usage: $0 <export_id_uuid> [output_path]"
    echo ""
    echo "To list available exports:"
    echo "  kubectl -n ai-persona-system exec -i pod/postgres-clients-0 -- \\"
    echo "      psql -U clients_user -d clients_db -c \\"
    echo "      \"SELECT id, agent_type, rows_exported, created_at FROM training_exports.runs ORDER BY created_at DESC;\""
    exit 1
fi

mkdir -p "$(dirname "$OUT")"

echo "Pulling export $EXPORT_ID to $OUT"

kubectl -n ai-persona-system exec -i pod/postgres-clients-0 -- \
    psql -U clients_user -d clients_db -tA -c "
COPY (
    SELECT jsonb_build_object('messages', messages, 'metadata', metadata)
    FROM training_exports.rows
    WHERE export_id = '$EXPORT_ID'::uuid
    ORDER BY row_index
) TO STDOUT
" > "$OUT"

# Verify
LINES=$(wc -l < "$OUT")
VALID=$(jq -c . "$OUT" 2>/dev/null | wc -l)
SIZE=$(stat -c%s "$OUT" 2>/dev/null || stat -f%z "$OUT")

echo ""
echo "Dataset pulled:"
echo "  file:   $OUT"
echo "  lines:  $LINES"
echo "  valid:  $VALID"
echo "  bytes:  $SIZE"
echo ""

if [ "$LINES" != "$VALID" ]; then
    echo "WARNING: $((LINES - VALID)) lines failed jq parse"
    exit 1
fi

# Sanity — sample the first row's shape
echo "First record shape:"
head -1 "$OUT" | jq '{
    roles: [.messages[].role],
    assistant_keys: (try (.messages[-1].content | fromjson | keys | sort) catch ["FAILED_TO_PARSE"]),
    metadata_keys: (.metadata | keys | sort)
}'

echo ""
echo "Ready to train. Next:"
echo "  python 02_train_llama_3_3_70b.py --data $OUT --output /workspace/lora_iter0"
