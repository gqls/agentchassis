#!/usr/bin/env bash
#
# kafka-orphan-topic-sweep.sh — delete orphaned `job.*` Kafka topics using a
# PER-TOPIC liveness test, not a fleet-wide idleness test.
#
# WHY THIS EXISTS (bugs_open/240)
# ------------------------------------------------------------------------------
# Every kafka-go client in this estate shares `platform/kafka`'s producer
# transport, which leaves `MetadataTopics` blank. kafka-go documents that as
# "metadata information of all topics in the cluster will be retrieved", and it
# re-issues that request on rand[0, MetadataTTL) — a ~3s mean — in a background
# goroutine, forever. So the cost of every idle client scales with the TOTAL
# TOPIC COUNT of the cluster. At 24,998 topics / 50,100 partitions it pushed
# kafka-scheduler (128Mi, the fleet's smallest limit) into a permanent OOM
# crashloop and cut fleet orchestration volume by ~90%.
#
# 24,087 of those topics were ephemeral per-step `job.*` topics that nothing
# deletes. The existing reaper (agent-job-cleanup) only deletes on a tick where
# the WHOLE fleet is idle, then only topics already orphaned on the previous
# tick. A live platform is essentially never idle, so that branch never fires.
#
# WHY NOT JUST DELETE THEM ALL (bugs_closed/071)
# ------------------------------------------------------------------------------
# That is precisely the bug this estate already had. Before 2026-07-26 the
# cleanup deleted every `job.*` topic every 10 minutes because its guard never
# matched, killing agents mid-run: a response produced after its topic was
# deleted auto-recreates the topic, but the consumer group's offsets die with the
# old one, so the reply is produced and never consumed. Long-running spawned
# agents were near-guaranteed to die.
#
# 071's fix made the guard fleet-wide and conservative. This script is the third
# option: keep a real safety test, but make it PER TOPIC so it can run while the
# fleet is busy.
#
# THE TEST
# ------------------------------------------------------------------------------
# A `job.<corr8>-<orch8>-<agent>-<step>.{requests,responses}` topic is DELETABLE
# only if its `corr8` matches no protected correlation, where protected means:
#
#   * an orchestration_states row that is NOT terminal
#     (status NOT IN COMPLETED/FAILED/CANCELLED), OR
#   * any orchestration_states row touched within $PROTECT_WINDOW (default 6h).
#
# The second clause is the belt-and-braces one and it is what makes this safe to
# run mid-flight: a run that finished 30 seconds ago is still protected, because
# a late response may yet land on its topic.
#
# Note orchestration_states retains ~2 days. A topic whose row has been purged
# classifies as orphaned — which is correct: that run cannot come back.
#
# VERIFY THE TEST BEFORE YOU TRUST IT. The script prints a positive control:
# a known-live correlation MUST appear in the protected set. If the protected
# count is 0, something is wrong with the DB read — refuse rather than delete.
#
# USAGE
#   ./scripts/kafka-orphan-topic-sweep.sh            # dry run, prints counts only
#   ./scripts/kafka-orphan-topic-sweep.sh --apply    # actually deletes
#   PROTECT_WINDOW='24 hours' ./scripts/kafka-orphan-topic-sweep.sh
#
set -uo pipefail

NS_KAFKA="${NS_KAFKA:-kafka}"
NS_APP="${NS_APP:-ai-persona-system}"
KAFKA_POD="${KAFKA_POD:-personae-kafka-cluster-combined-pool-prod-0}"
PG_POD="${PG_POD:-postgres-clients-0}"
PROTECT_WINDOW="${PROTECT_WINDOW:-6 hours}"
APPLY=0
[ "${1:-}" = "--apply" ] && APPLY=1

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

echo "=== kafka orphan topic sweep $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="
echo "protect window: $PROTECT_WINDOW   mode: $([ $APPLY -eq 1 ] && echo APPLY || echo DRY-RUN)"

# --- 1. every topic -----------------------------------------------------------
kubectl -n "$NS_KAFKA" exec "$KAFKA_POD" -- \
  bin/kafka-topics.sh --bootstrap-server localhost:9092 --list 2>/dev/null \
  | grep '^job\.' | sort -u > "$WORK/job_topics.txt"
TOTAL=$(wc -l < "$WORK/job_topics.txt")
echo "job.* topics found: $TOTAL"
if [ "$TOTAL" -eq 0 ]; then echo "nothing to do"; exit 0; fi

# --- 2. protected correlations ------------------------------------------------
kubectl -n "$NS_APP" exec -i "$PG_POD" -- \
  psql -U clients_user -d clients_db -At -c "
    SELECT DISTINCT left(correlation_id::text,8)
    FROM orchestration_states
    WHERE status NOT IN ('COMPLETED','FAILED','CANCELLED')
       OR updated_at > now() - interval '$PROTECT_WINDOW';" 2>/dev/null \
  | grep -E '^[0-9a-f]{8}$' | sort -u > "$WORK/protected.txt"
PROT=$(wc -l < "$WORK/protected.txt")
echo "protected correlations: $PROT"

# A zero here means the DB read failed or returned nothing. Deleting on that
# basis would delete the topics of every live run — refuse. (This is the
# fail-safe 071 did not have: an empty guard result must never mean "delete
# everything", which is exactly how that bug behaved.)
if [ "$PROT" -eq 0 ]; then
  echo "REFUSING: zero protected correlations. Either the DB read failed or the"
  echo "fleet has been idle for $PROTECT_WINDOW. Investigate before deleting."
  exit 1
fi

# --- 3. classify --------------------------------------------------------------
awk -F'[.-]' '{print $2"\t"$0}' "$WORK/job_topics.txt" > "$WORK/by_corr.tsv"
awk 'NR==FNR{p[$1];next} ($1 in p){print $2 > KEEP; next} {print $2 > DEL}' \
    KEEP="$WORK/keep.txt" DEL="$WORK/delete.txt" \
    "$WORK/protected.txt" "$WORK/by_corr.tsv"
touch "$WORK/keep.txt" "$WORK/delete.txt"
KEEP_N=$(wc -l < "$WORK/keep.txt"); DEL_N=$(wc -l < "$WORK/delete.txt")
echo "KEEP (protected): $KEEP_N"
echo "DELETE (orphaned): $DEL_N"

# --- 4. positive control ------------------------------------------------------
# The classifier must protect SOMETHING if anything is live. A keep-count of 0
# while protected correlations exist means the corr8 extraction is broken (e.g.
# the topic name format changed) — that would silently reclassify live topics as
# orphans, so refuse.
if [ "$KEEP_N" -eq 0 ]; then
  echo "REFUSING: $PROT correlations are protected but 0 topics matched them."
  echo "The topic-name format has probably changed; re-check the corr8 field"
  echo "position in step 3 before trusting this classification."
  exit 1
fi

if [ $APPLY -eq 0 ]; then
  echo
  echo "DRY RUN — nothing deleted. Sample of what WOULD go:"
  head -5 "$WORK/delete.txt"
  echo "Re-run with --apply to delete $DEL_N topics."
  exit 0
fi

# --- 5. delete ----------------------------------------------------------------
# One --delete call per topic; kafka-topics.sh accepts a list but a single bad
# name fails the batch, and a partial sweep is fine (the next run finishes it).
OK=0; FAIL=0
while read -r TOPIC; do
  [ -z "$TOPIC" ] && continue
  if kubectl -n "$NS_KAFKA" exec "$KAFKA_POD" -- \
       bin/kafka-topics.sh --bootstrap-server localhost:9092 \
       --delete --topic "$TOPIC" >/dev/null 2>&1; then
    OK=$((OK+1))
  else
    FAIL=$((FAIL+1))
  fi
  if [ $(( (OK+FAIL) % 250 )) -eq 0 ]; then
    echo "  ... $((OK+FAIL))/$DEL_N (ok=$OK fail=$FAIL)"
  fi
done < "$WORK/delete.txt"

echo "deleted: $OK   failed: $FAIL"
echo "remaining job.* topics:"
kubectl -n "$NS_KAFKA" exec "$KAFKA_POD" -- \
  bin/kafka-topics.sh --bootstrap-server localhost:9092 --list 2>/dev/null \
  | grep -c '^job\.'
