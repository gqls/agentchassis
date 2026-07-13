#################  first full sweep then the combined promote and verify
#!/bin/bash
# full_uk_sweep.sh - Sweep all UK postcode areas for vet practices
#
# Sends one area-sweep-orchestrator message per area code.
# Each orchestrator loads all districts for its area and dispatches discoverers.
#
# Usage:
#   bash full_uk_sweep.sh              # all 121 area codes
#   bash full_uk_sweep.sh 10           # first 10 area codes only
#   bash full_uk_sweep.sh 0 5          # all areas, 5s delay between
#   bash full_uk_sweep.sh 0 2 EH       # start from EH (Edinburgh)
#
# This can run for hours. Each area code dispatches 10-97 discoverers,
# each discoverer uses 1 search credit. Total: ~3,402 credits.
#
# The adapter's REQUEST_THROTTLE_MS handles rate limiting for web searches.

set -e

LIMIT=${1:-0}       # 0 = all area codes
DELAY=${2:-3}        # seconds between orchestrator messages
START_FROM=${3:-}    # optional: skip area codes before this one


LIMIT=0
DELAY=5
START_FROM=0
CLIENT_ID="00000000-0000-0000-0000-000000000001"
KAFKA_BOOTSTRAP="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
TOPIC="system.agent.generic.requests"
BATCH_NAME="full-sweep-$(date +%Y%m%d-%H%M%S)"

echo "========================================="
echo "Full UK Area Sweep"
echo "========================================="
echo "  Batch:      ${BATCH_NAME}"
echo "  Limit:      ${LIMIT} (0=all)"
echo "  Delay:      ${DELAY}s between areas"
echo "  Start from: ${START_FROM:-beginning}"
echo "========================================="
echo ""

# Step 1: Get area codes from database (sorted, with district counts)
echo "Querying area codes with unswept districts..."
AREA_DATA=$(kubectl exec -n ai-persona-system postgres-clients-0 -- \
    psql -U clients_user -d clients_db -t -A -c "
        SELECT area_code, COUNT(*) AS total,
               COUNT(*) FILTER (WHERE sweep_count = 0) AS unswept
        FROM business_intel.search_areas
        WHERE country = 'GB'
        GROUP BY area_code
        HAVING COUNT(*) FILTER (WHERE sweep_count = 0) > 0
        ORDER BY area_code
    ")

if [ -z "$AREA_DATA" ]; then
    echo "No unswept area codes found. Full sweep may already be complete."
    echo ""
    echo "Check status:"
    echo "  kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \\"
    echo "    \"SELECT COUNT(*) AS total, COUNT(*) FILTER (WHERE sweep_count > 0) AS swept,"
    echo "     COUNT(*) FILTER (WHERE sweep_count = 0) AS unswept FROM business_intel.search_areas WHERE country = 'GB';\""
    exit 0
fi

TOTAL_AREAS=$(echo "$AREA_DATA" | wc -l | tr -d ' ')
echo "Found $TOTAL_AREAS area codes with unswept districts."
echo ""

# Step 2: Start persistent kcat pod for sending
POD_NAME="sweep-sender-$$"
echo "Starting sender pod..."
kubectl -n kafka run "$POD_NAME" \
    --image=edenhill/kcat:1.7.1 \
    --restart=Never \
    --command -- sleep 86400 &

echo "Waiting for pod..."
kubectl -n kafka wait --for=condition=Ready pod/"$POD_NAME" --timeout=30s
echo ""

# Step 3: Send one orchestrator message per area code
COUNT=0
SKIPPING=true
if [ -z "$START_FROM" ]; then
    SKIPPING=false
fi

for ROW in $AREA_DATA; do
    AREA_CODE=$(echo "$ROW" | cut -d'|' -f1)
    TOTAL_DISTRICTS=$(echo "$ROW" | cut -d'|' -f2)
    UNSWEPT=$(echo "$ROW" | cut -d'|' -f3)

    # Handle start-from
    if [ "$SKIPPING" = true ]; then
        if [ "$AREA_CODE" = "$START_FROM" ]; then
            SKIPPING=false
        else
            continue
        fi
    fi

    COUNT=$((COUNT + 1))

    # Apply limit
    if [ "$LIMIT" -gt 0 ] && [ "$COUNT" -gt "$LIMIT" ]; then
        echo "Limit of $LIMIT reached."
        break
    fi

    ORCH_ID=$(cat /proc/sys/kernel/random/uuid)
    REQ_ID=$(cat /proc/sys/kernel/random/uuid)
    CORR_ID=$(cat /proc/sys/kernel/random/uuid)
    MSG_ID=$(cat /proc/sys/kernel/random/uuid)
    ORCH_NAME="${BATCH_NAME}-${AREA_CODE}"

    BODY="{\"action\":\"orchestrate\",\"config\":{\"agent_type\":\"area-sweep-orchestrator\"},\"input_data\":{\"limit\":0,\"country\":\"GB\",\"area_code\":\"${AREA_CODE}\"}}"

    echo "[$COUNT/$TOTAL_AREAS] ${AREA_CODE}: ${UNSWEPT}/${TOTAL_DISTRICTS} unswept districts"

    echo "$BODY" | kubectl exec -n kafka -i "$POD_NAME" -- kcat -P \
        -b "$KAFKA_BOOTSTRAP" \
        -t "$TOPIC" \
        -H "correlation_id=$CORR_ID" \
        -H "request_id=$REQ_ID" \
        -H "message_id=$MSG_ID" \
        -H "orchestration_id=$ORCH_ID" \
        -H "orchestration_name=$ORCH_NAME" \
        -H "step_name=start" \
        -H "client_id=$CLIENT_ID" \
        -H "message_type=request" \
        -H "action=orchestrate" \
        -H "from_agent_type=user" \
        -H "from_agent_id=cli-full-sweep" \
        -H "responses_topic=system.generic.responses"

    echo "  -> $ORCH_NAME"

    if [ "$COUNT" -lt "$TOTAL_AREAS" ]; then
        sleep "$DELAY"
    fi
done

echo ""
echo "========================================="
echo "Dispatched $COUNT area orchestrators."
echo "Each will dispatch discoverers for its unswept districts."
echo "========================================="

# Step 4: Cleanup
echo ""
echo "Cleaning up sender pod..."
kubectl -n kafka delete pod "$POD_NAME" --force 2>/dev/null || true

echo ""
echo "MONITORING:"
echo ""
echo "1. Sweep progress by area:"
echo "   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \\"
echo "     \"SELECT area_code, COUNT(*) AS total,"
echo "      COUNT(*) FILTER (WHERE sweep_count > 0) AS swept,"
echo "      COUNT(*) FILTER (WHERE sweep_count = 0) AS remaining,"
echo "      SUM(candidates_found) AS candidates"
echo "      FROM business_intel.search_areas WHERE country = 'GB'"
echo "      GROUP BY area_code ORDER BY area_code;\""
echo ""
echo "2. Overall progress:"
echo "   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \\"
echo "     \"SELECT COUNT(*) AS total_districts,"
echo "      COUNT(*) FILTER (WHERE sweep_count > 0) AS swept,"
echo "      COUNT(*) FILTER (WHERE sweep_count = 0) AS remaining,"
echo "      SUM(candidates_found) AS total_candidates"
echo "      FROM business_intel.search_areas WHERE country = 'GB';\""
echo ""
echo "3. Discovery candidates:"
echo "   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \\"
echo "     \"SELECT status, COUNT(*), COUNT(*) FILTER (WHERE website_url IS NOT NULL) AS have_url"
echo "      FROM business_intel.discovery_candidates GROUP BY status;\""
echo ""
echo "4. Errors:"
echo "   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \\"
echo "     \"SELECT orchestration_name, error FROM orchestration_states"
echo "      WHERE orchestration_name LIKE '${BATCH_NAME}%' AND status = 'ERROR';\""







####################################### promote and verify
#!/bin/bash
# promote_and_verify.sh - Promote discovery candidates then verify them
#
# Bridges the gap between sweep (writes to discovery_candidates) and
# verification (reads from businesses). Run this periodically while
# full_uk_sweep.sh is running.
#
# Usage:
#   bash promote_and_verify.sh          # promote all, verify all
#   bash promote_and_verify.sh 20       # promote all, verify first 20
#   bash promote_and_verify.sh 0 loop   # promote+verify in a loop every 5 min

set -e

VERIFY_LIMIT=${1:-0}
MODE=${2:-once}         # "once" or "loop"
LOOP_INTERVAL=300       # 5 minutes between loops
CLIENT_ID="00000000-0000-0000-0000-000000000001"
KAFKA_BOOTSTRAP="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
TOPIC="system.agent.generic.requests"

PG_CMD="kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db"

promote_candidates() {
    echo "=== PROMOTE: Moving candidates to businesses ==="

    $PG_CMD -c "
    DO \$\$
    DECLARE
        v_vertical_id UUID;
        v_candidate RECORD;
        v_business_id UUID;
        v_promoted INT := 0;
        v_skipped_dup INT := 0;
        v_skipped_nourl INT := 0;
        v_skipped_dir INT := 0;
        v_existing_id UUID;
    BEGIN
        SELECT id INTO v_vertical_id
        FROM business_intel.business_verticals WHERE slug = 'veterinary';

        IF v_vertical_id IS NULL THEN
            RAISE EXCEPTION 'Veterinary vertical not found';
        END IF;

        FOR v_candidate IN
            SELECT id, name, website_url, postcode, detected_group, is_independent
            FROM business_intel.discovery_candidates
            WHERE status = 'pending'
            ORDER BY created_at ASC
        LOOP
            -- Skip candidates without website_url (from directory listings)
            IF v_candidate.website_url IS NULL THEN
                UPDATE business_intel.discovery_candidates
                SET status = 'needs_enrichment', updated_at = NOW()
                WHERE id = v_candidate.id;
                v_skipped_nourl := v_skipped_nourl + 1;
                CONTINUE;
            END IF;

            -- Skip directory/junk titles
            IF v_candidate.name ILIKE '%vets near%'
               OR v_candidate.name ILIKE '%veterinarians near%'
               OR v_candidate.name ILIKE '%VETS DIRECTORY%'
               OR v_candidate.name ILIKE 'THE BEST%' THEN
                UPDATE business_intel.discovery_candidates
                SET status = 'dismissed',
                    notes = 'Auto-dismissed: directory listing title',
                    reviewed_at = NOW(), updated_at = NOW()
                WHERE id = v_candidate.id;
                v_skipped_dir := v_skipped_dir + 1;
                CONTINUE;
            END IF;

            -- Skip if website already in businesses
            SELECT id INTO v_existing_id
            FROM business_intel.businesses
            WHERE website_url ILIKE v_candidate.website_url || '%'
               OR website_url ILIKE 'https://www.' || REPLACE(v_candidate.website_url, 'https://', '') || '%'
            LIMIT 1;

            IF v_existing_id IS NOT NULL THEN
                UPDATE business_intel.discovery_candidates
                SET status = 'matched', matched_business_id = v_existing_id,
                    match_method = 'website_url', match_confidence = 0.95,
                    reviewed_at = NOW(), updated_at = NOW()
                WHERE id = v_candidate.id;
                v_skipped_dup := v_skipped_dup + 1;
                CONTINUE;
            END IF;

            -- Skip if another candidate with same URL already promoted
            SELECT promoted_business_id INTO v_existing_id
            FROM business_intel.discovery_candidates
            WHERE website_url = v_candidate.website_url
              AND status = 'promoted' AND promoted_business_id IS NOT NULL
            LIMIT 1;

            IF v_existing_id IS NOT NULL THEN
                UPDATE business_intel.discovery_candidates
                SET status = 'matched', matched_business_id = v_existing_id,
                    match_method = 'website_url_candidate_dup', match_confidence = 0.95,
                    reviewed_at = NOW(), updated_at = NOW()
                WHERE id = v_candidate.id;
                v_skipped_dup := v_skipped_dup + 1;
                CONTINUE;
            END IF;

            -- Insert into businesses
            INSERT INTO business_intel.businesses (
                name, website_url, postcode, country,
                group_name, is_independent,
                vertical_id, business_type,
                verification_status, is_active, created_at
            ) VALUES (
                v_candidate.name, v_candidate.website_url,
                v_candidate.postcode, 'GB',
                v_candidate.detected_group,
                COALESCE(v_candidate.is_independent, true),
                v_vertical_id, 'veterinary_practice',
                'pending', true, NOW()
            ) RETURNING id INTO v_business_id;

            UPDATE business_intel.discovery_candidates
            SET status = 'promoted', promoted_business_id = v_business_id,
                reviewed_at = NOW(), updated_at = NOW()
            WHERE id = v_candidate.id;

            v_promoted := v_promoted + 1;
        END LOOP;

        RAISE NOTICE 'Promoted: %, Dup: %, No URL: %, Directory: %',
            v_promoted, v_skipped_dup, v_skipped_nourl, v_skipped_dir;
    END \$\$;
    "
    echo ""
}

verify_promoted() {
    echo "=== VERIFY: Sending unverified businesses to verifier ==="

    LIMIT_CLAUSE=""
    if [ "$VERIFY_LIMIT" -gt 0 ] 2>/dev/null; then
        LIMIT_CLAUSE="LIMIT $VERIFY_LIMIT"
    fi

    BUSINESS_IDS=$($PG_CMD -t -A -c "
        SELECT b.id
        FROM business_intel.businesses b
        JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
        WHERE bv.slug = 'veterinary'
          AND b.is_active = true
          AND b.verification_status = 'pending'
          AND COALESCE(b.verification_count, 0) = 0
        ORDER BY b.created_at ASC
        ${LIMIT_CLAUSE}
    ")

    if [ -z "$BUSINESS_IDS" ]; then
        echo "No unverified businesses to process."
        return
    fi

    TOTAL=$(echo "$BUSINESS_IDS" | wc -l | tr -d ' ')
    echo "Found $TOTAL unverified businesses."

    BATCH_NAME="auto-verify-$(date +%Y%m%d-%H%M%S)"

    # Start persistent kcat pod
    POD_NAME="auto-verify-sender-$$"
    kubectl -n kafka run "$POD_NAME" \
        --image=edenhill/kcat:1.7.1 \
        --restart=Never \
        --command -- sleep 3600 &
    kubectl -n kafka wait --for=condition=Ready pod/"$POD_NAME" --timeout=30s

    COUNT=0
    for BIZ_ID in $BUSINESS_IDS; do
        COUNT=$((COUNT + 1))

        ORCH_ID=$(cat /proc/sys/kernel/random/uuid)
        REQ_ID=$(cat /proc/sys/kernel/random/uuid)

        echo "{\"action\":\"orchestrate\",\"config\":{\"agent_type\":\"vet-practice-verifier\"},\"input_data\":{\"business_id\":\"${BIZ_ID}\"}}" | \
        kubectl exec -n kafka -i "$POD_NAME" -- kcat -P \
            -b "$KAFKA_BOOTSTRAP" \
            -t "$TOPIC" \
            -H "correlation_id=$(cat /proc/sys/kernel/random/uuid)" \
            -H "request_id=$REQ_ID" \
            -H "message_id=$(cat /proc/sys/kernel/random/uuid)" \
            -H "orchestration_id=$ORCH_ID" \
            -H "orchestration_name=${BATCH_NAME}-${COUNT}" \
            -H "step_name=start" \
            -H "client_id=$CLIENT_ID" \
            -H "message_type=request" \
            -H "action=orchestrate" \
            -H "from_agent_type=user" \
            -H "from_agent_id=cli-auto-verify" \
            -H "responses_topic=system.generic.responses"

        echo "  [$COUNT/$TOTAL] $BIZ_ID"
        sleep 0.2
    done

    kubectl -n kafka delete pod "$POD_NAME" --force 2>/dev/null || true
    echo "Sent $COUNT verification requests."
    echo ""
}

show_status() {
    echo "=== STATUS ==="
    $PG_CMD -c "
        SELECT 'sweep' AS pipeline,
               COUNT(*) FILTER (WHERE sweep_count > 0) AS done,
               COUNT(*) FILTER (WHERE sweep_count = 0) AS remaining
        FROM business_intel.search_areas WHERE country = 'GB'
        UNION ALL
        SELECT 'candidates',
               COUNT(*) FILTER (WHERE status != 'pending'),
               COUNT(*) FILTER (WHERE status = 'pending')
        FROM business_intel.discovery_candidates
        UNION ALL
        SELECT 'verified',
               COUNT(*) FILTER (WHERE verification_status != 'pending'),
               COUNT(*) FILTER (WHERE verification_status = 'pending')
        FROM business_intel.businesses b
        JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
        WHERE bv.slug = 'veterinary' AND b.is_active = true;
    "
    echo ""
}

# Main
if [ "$MODE" = "loop" ]; then
    echo "Running in loop mode (every ${LOOP_INTERVAL}s). Ctrl+C to stop."
    echo ""
    while true; do
        echo "======= $(date) ======="
        promote_candidates
        verify_promoted
        show_status
        echo "Sleeping ${LOOP_INTERVAL}s..."
        sleep "$LOOP_INTERVAL"
    done
else
    promote_candidates
    verify_promoted
    show_status
fi