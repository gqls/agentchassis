#!/bin/bash
# ============================================================================
# 003_backfill_jsonl_to_postgres.sh
# ============================================================================
# One-off loader: read the 1,957-row JSONL we already pulled from a chassis
# pod, insert as a named seed export into training_exports.
#
# Why: we already spent time running the export, retrieving the file, and
# validating it. Re-running via the new orchestrator would produce an
# identical-but-slightly-larger dataset (more llm_call_log rows accumulated).
# Seeding the existing file preserves the dataset we validated.
#
# After this runs:
#   - training_exports.runs has a new row with source_notes='backfilled from JSONL 2026-04-23'
#   - training_exports.rows has 1,957 rows referencing that export_id
#
# This is NOT idempotent — running it twice creates two seed exports. Rename
# or move the JSONL file after running to avoid accidental duplication.
# ============================================================================

set -euo pipefail

JSONL_PATH="${1:-./page_content_writer_iter0.jsonl}"
POD="${POD:-pod/postgres-clients-0}"
NAMESPACE="${NAMESPACE:-ai-persona-system}"

if [ ! -f "$JSONL_PATH" ]; then
    echo "JSONL file not found: $JSONL_PATH"
    echo "Usage: $0 [path-to-jsonl]"
    exit 1
fi

LINE_COUNT=$(wc -l < "$JSONL_PATH")
VALID_COUNT=$(jq -c . "$JSONL_PATH" 2>/dev/null | wc -l)
FILE_SIZE=$(stat -c%s "$JSONL_PATH")

echo "============================================================"
echo "JSONL backfill"
echo "============================================================"
echo "  file:     $JSONL_PATH"
echo "  lines:    $LINE_COUNT"
echo "  valid:    $VALID_COUNT"
echo "  bytes:    $FILE_SIZE"
echo "============================================================"

if [ "$LINE_COUNT" != "$VALID_COUNT" ]; then
    echo "ERROR: $((LINE_COUNT - VALID_COUNT)) lines failed jq parse. Fix before seeding."
    exit 1
fi

# Extract metadata from the first row to pick filter values for the runs entry
FIRST=$(head -1 "$JSONL_PATH")
AGENT_TYPE=$(echo "$FIRST" | jq -r '.metadata.agent_type')
STEP_NAME=$(echo "$FIRST" | jq -r '.metadata.step_name')
MODEL=$(echo "$FIRST" | jq -r '.metadata.model')

echo "  agent_type: $AGENT_TYPE"
echo "  step_name:  $STEP_NAME"
echo "  model:      $MODEL"
echo ""

# ── Step 1: insert the runs row, get back the export_id ─────────────────────
# Build the INSERT statement as a single-line string. Use psql -c with -tA
# to get tuple-only output. Pipe through head -1 to guarantee only the first
# line of output (INSERT returning clauses can sometimes emit the RETURNING
# value and a status line depending on psql version/flags).

INSERT_SQL="INSERT INTO training_exports.runs (agent_type, step_name, model_filter, format, export_version, size_bytes, source_notes, rows_seen, rows_exported, completed_at) VALUES ('$AGENT_TYPE', '$STEP_NAME', '$MODEL', 'chatml', '1', $FILE_SIZE, 'Backfilled from JSONL retrieved via kubectl exec cat on $(date +%Y-%m-%d)', $LINE_COUNT, $LINE_COUNT, NOW()) RETURNING id::text"

EXPORT_ID=$(kubectl -n "$NAMESPACE" exec -i "$POD" -- \
    psql -U clients_user -d clients_db -tA -c "$INSERT_SQL" \
    | head -1 \
    | tr -d '[:space:]')

if [ -z "$EXPORT_ID" ]; then
    echo "ERROR: failed to create runs row"
    exit 1
fi

# Sanity-check it looks like a UUID
if ! echo "$EXPORT_ID" | grep -qE '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'; then
    echo "ERROR: returned export_id doesn't look like a UUID: '$EXPORT_ID'"
    exit 1
fi

echo "Created runs row: export_id=$EXPORT_ID"
echo ""

# ── Step 2: stream rows via COPY FROM STDIN ─────────────────────────────────
# For each JSONL line, emit a TSV row (export_id, row_index, messages_json, metadata_json).
# Use `\copy` with FREEZE OFF, CSV format with tab delimiter for robust escaping
# of embedded tabs/newlines.
#
# jq produces each row as: {export_id, row_index, messages, metadata}
# We reformat with awk/printf to build the COPY stream.

ROW_INDEX=0
TEMP_TSV=$(mktemp)
trap 'rm -f "$TEMP_TSV"' EXIT

echo "Preparing bulk-insert stream..."
jq -c --arg eid "$EXPORT_ID" '
    [$eid, (input_line_number - 1),
     (.messages | @json),
     (.metadata | @json)]
    | @tsv
' "$JSONL_PATH" > "$TEMP_TSV"

PREPARED=$(wc -l < "$TEMP_TSV")
echo "Prepared $PREPARED rows for COPY"
echo ""

# ── Step 3: pipe TSV into psql \copy ────────────────────────────────────────

echo "Running COPY..."
kubectl -n "$NAMESPACE" exec -i "$POD" -- \
    psql -U clients_user -d clients_db -c "\\copy training_exports.rows (export_id, row_index, messages, metadata) FROM STDIN WITH (FORMAT text, DELIMITER E'\t')" \
    < "$TEMP_TSV"

echo ""

# ── Step 4: verify ──────────────────────────────────────────────────────────

echo "Verification:"
kubectl -n "$NAMESPACE" exec -i "$POD" -- \
    psql -U clients_user -d clients_db -c "
SELECT r.id as export_id,
       r.agent_type,
       r.step_name,
       r.rows_exported,
       r.size_bytes,
       r.source_notes,
       COUNT(rw.id) as actual_rows
FROM training_exports.runs r
LEFT JOIN training_exports.rows rw ON rw.export_id = r.id
WHERE r.id = '$EXPORT_ID'::uuid
GROUP BY r.id, r.agent_type, r.step_name, r.rows_exported, r.size_bytes, r.source_notes;"

echo ""
echo "Done. export_id=$EXPORT_ID"
