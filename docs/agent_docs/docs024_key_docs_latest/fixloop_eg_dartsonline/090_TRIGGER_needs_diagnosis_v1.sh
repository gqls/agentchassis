#!/usr/bin/env bash
# 090_TRIGGER_needs_diagnosis_v1.sh — F0.1c: the ONE documented way a bug goes
# into the diagnosis→fix loop.
#
# It does two things, in order:
#   1. writes a durable `needs_diagnosis` work item (pipeline='diagnose') —
#      the intake RECORD, queryable long after the pods are reaped;
#   2. fires the diagnose-orchestrator envelope of 084, carrying the SAME
#      correlation_id, so the item, the bundles in diagnosis_artifacts, and the
#      terminal doc_notes row all join on one key.
#
# Extends drafts/084_TRIGGER_diagnose_v1.sh (the canonical envelope). It does
# not replace it: 084 remains the bare manual trigger for ad-hoc runs with no
# intake record. Q-B kept both, deliberately.
#
# ── THREE THINGS THE DESIGN GOT WRONG UNTIL THE SCHEMA WAS READ (2026-07-09) ──
#
# 1. site_id is NOT NULL. Q-B said "site-less code bugs ride null-site". They
#    cannot: the column forbids it, AND LoadWorkItemsAction parses site_id as a
#    required uuid and filters `WHERE wi.site_id = $1`, so a NULL-site item
#    would never be loaded even if the column allowed it.
#
# 2. The platform already solved this. `system.internal`
#    (eac60db8-b032-432b-b36d-76f37632045d, sites.status='system') is an
#    existing pseudo-site carrying platform-wide work — the maintenance-pipeline
#    component_quality_scan items live there today. We anchor to it rather than
#    invent a second mechanism. Reuse before recreate.
#
# 3. EVERY needs_diagnosis item anchors to system.internal — even when the bug
#    IS about a real site. The site under diagnosis travels in spec.site_id /
#    spec.runtime_site, never in the item's own site_id column. Why: the live
#    build-dispatch-loop's load_items step is configured with ONLY
#    {site_id, max_items} — it has NO item_pipeline filter — so an item parked
#    on a real site would be claimed by that site's next build dispatch and
#    handed to whatever handler_agent it names. Anchoring to system.internal
#    keeps the diagnose namespace physically off every per-site dispatch loop,
#    without editing a workflow the builder thread owns.
#
# 4. The item parks at status='awaiting_diagnosis', NOT 'detected'.
#    triage_detect_items promotes `WHERE site_id = $1 AND status='detected'`
#    with no pipeline filter, and REWRITES pipeline to 'build' — so 'detected'
#    would let a diagnose item be laundered into the build queue. No existing
#    sweep selects 'awaiting_diagnosis', so it is inert by construction.
#    See 0NN_diagnose_dispatch_loop.sql for the full inertness table.
#
# 5. max_attempts=1. claimed-item-timeout resets any claim older than 40 minutes
#    back to 'triaged' — which WOULD expose a slow diagnosis to the build
#    dispatcher. With max_attempts=1 that reset lands on 'failed' (terminal)
#    instead. It is also the right semantics: a 26-minute LLM loop should not
#    silently auto-retry.
#
# Automatic dispatch: `diagnose-dispatch-loop` + the `diagnose-pipeline-trigger`
# scheduled task (0NN_diagnose_dispatch_loop.sql) claim awaiting_diagnosis items
# on a 60s tick. The task ships DISABLED — enable it deliberately. Until then,
# and for any ad-hoc run, THIS SCRIPT is the dispatcher.
#
# Usage:
#   ./090_TRIGGER_needs_diagnosis_v1.sh "symptom text — keep free of \" and \\"
#
#   # REF defaults to the current branch; override only to pin a different one
#   SLUG=guides-nav RUNTIME_SITE=dartsonline.com \
#     ./090_TRIGGER_needs_diagnosis_v1.sh "nav links to a guides section with no content"
#
# ── 6. PRE-DISPATCH COVERAGE CHECK (added 2026-07-16, multi_session_coordination)
#    Many sessions work this cluster concurrently. On 2026-07-16 a diagnosis run
#    was refuted mid-flight because another session's json-leak-fix-retry batch
#    repaired the evidence underneath it — that batch was five minutes old and
#    visible in site_work_items at dispatch time. A live-state check is also a
#    snapshot: checking the pod does not check the queue. So before writing the
#    intake, this script asks "does any OPEN work item already touch this
#    target?" and REFUSES to dispatch on a hit unless FORCE=1.
#    Coverage semantics are copied (not redesigned) from silentCoverageClause,
#    platform/orchestration/actions/diagnose_silent_check_action.go: status NOT
#    IN ('complete','cancelled','rejected'). 'complete' does NOT cover — a
#    completed remedy with the violation still observable is the
#    remediation-ineffective signal, and that IS a valid diagnosis target.
#    Coverage key for code-only diagnoses (no site, no pages): the SEED_SCOPE
#    file set, at file granularity — the symbol part after ':' is ignored.
#
# Env:
#   SLUG           item_key suffix (default: derived from the symptom)
#   REF            git ref the diagnosis reads. DEFAULT = the current branch
#                  (changed from 'main' 2026-07-19 — main is routinely hundreds
#                  of commits behind here, and diagnosing a stale tree returns a
#                  confident wrong answer). Never the literal "HEAD". Refuses if
#                  the ref is not on origin, or if the branch cannot be
#                  determined — it does NOT fall back.
#   RUNTIME_SITE   domain of the site under diagnosis (runtime evidence tier)
#   SITE_ID        uuid of the site under diagnosis (goes in spec, not site_id)
#   SUBJECT_TYPE   tool | pipeline   } together, these open the tools chat's
#   SUBJECT_KEY    <function> | <pipeline> } persist_note subject gate (their 3b)
#   SEED_SCOPE     comma-separated path[:Symbol] entries for iteration 1
#                  (emitted into the envelope as a JSON ARRAY — a bare string
#                   parses to nil in ExtractStringListHelper and does nothing)
#   PAGES          comma-separated page names (pages.name slugs) or /urls the
#                  symptom is about — drives the page-keyed coverage probe,
#                  matched across ALL sites (a diagnosis can span sites: the
#                  2026-07-16 incident hit finetuning.uk AND gamesdesign.co.uk).
#                  No spaces around commas, no quotes in entries.
#   FORCE          1 = print coverage findings but dispatch anyway (default 0)
#   RECENT_WINDOW  site-level probe: how recently an open item must have been
#                  touched to block (default '2 hours'; older open items are
#                  listed FYI only — a parked backlog is not in-flight work)
#   DISPATCH       1 (default) fires the Kafka envelope; 0 records the item only
set -euo pipefail

SYMPTOM="${1:?usage: $0 \"symptom text\"}"

# The pseudo-site every platform-wide work item is anchored to. Not a magic
# number: SELECT id FROM sites WHERE domain='system.internal'.
SYSTEM_SITE_ID='eac60db8-b032-432b-b36d-76f37632045d'

TARGET_AGENT_TYPE='diagnose-orchestrator'
OWNER="${OWNER:-gqls}"
REPO="${REPO:-agentchassis}"
# REF — the git ref the diagnosis reads. Still NEVER the literal "HEAD" (user
# decision 2026-07-02); the loop fetches from the REMOTE, so it must be a ref
# the remote knows.
#
# Default changed from 'main' to the CURRENT BRANCH (owner decision 2026-07-19).
# Why: on this repo main is routinely hundreds of commits behind the working
# branch. On 2026-07-19 origin/main carried a 2-entry ctaFieldNames map while
# the working branch carried 6 and was 345 commits ahead — a diagnosis taken on
# the old default would have examined a tree in which the reported bug barely
# existed, and returned a confident wrong answer. A cited diagnosis of the wrong
# tree is worse than no diagnosis, because it reads as authoritative.
#
# Both checks below REFUSE rather than fall back. A silent fallback to main is
# precisely the failure this change exists to remove.
if [ -z "${REF:-}" ]; then
  if REPO_ROOT_FOR_REF=$(git rev-parse --show-toplevel 2>/dev/null); then
    REF=$(cd "$REPO_ROOT_FOR_REF" && git rev-parse --abbrev-ref HEAD 2>/dev/null || true)
  fi
  if [ -z "${REF:-}" ] || [ "$REF" = "HEAD" ]; then
    echo "REFUSING TO DISPATCH: cannot determine the current branch" >&2
    echo "  (detached HEAD, or not run from inside a git checkout)." >&2
    echo "  Set it explicitly to a ref the remote has:  REF=<branch> $0 \"<symptom>\"" >&2
    exit 2
  fi
fi

# A ref that exists only locally cannot be fetched by the loop.
if REPO_ROOT_FOR_REF=$(git rev-parse --show-toplevel 2>/dev/null); then
  if ! (cd "$REPO_ROOT_FOR_REF" && git ls-remote --exit-code --heads origin "$REF" >/dev/null 2>&1); then
    echo "REFUSING TO DISPATCH: ref '$REF' does not exist on origin." >&2
    echo "  The diagnosis clones from the REMOTE, not this working tree." >&2
    echo "  Push the branch, or set REF to a ref the remote already has." >&2
    exit 2
  fi
  # Advisory: the remote ref can lag your working tree, and the loop reads the
  # remote. Commits you have made but not pushed are INVISIBLE to the diagnosis.
  AHEAD_OF_REMOTE=$(cd "$REPO_ROOT_FOR_REF" && \
                    git rev-list --count "origin/${REF}..HEAD" 2>/dev/null || echo 0)
  if [ "${AHEAD_OF_REMOTE:-0}" -gt 0 ]; then
    echo "   ADVISORY (not blocking): local HEAD is ${AHEAD_OF_REMOTE} commit(s) ahead of"
    echo "   origin/${REF} — the diagnosis reads origin/${REF}, so those commits are"
    echo "   INVISIBLE to it. Push first if the symptom depends on them."
  fi
fi
RUNTIME_SITE="${RUNTIME_SITE:-}"
SITE_ID="${SITE_ID:-}"
SUBJECT_TYPE="${SUBJECT_TYPE:-}"
SUBJECT_KEY="${SUBJECT_KEY:-}"
SEED_SCOPE="${SEED_SCOPE:-}"
PAGES="${PAGES:-}"
FORCE="${FORCE:-0}"
RECENT_WINDOW="${RECENT_WINDOW:-2 hours}"
DISPATCH="${DISPATCH:-1}"
CLIENT_ID='demo_client'

PSQL=(kubectl exec -i -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1)

# item_key drives idx_swi_dedup (site_id, item_key) WHERE status not terminal —
# so re-running with the same SLUG while an intake is still open is a no-op
# rather than a duplicate. That is the intended idempotency.
if [ -z "${SLUG:-}" ]; then
  SLUG=$(printf '%s' "$SYMPTOM" | tr '[:upper:]' '[:lower:]' \
         | tr -cs 'a-z0-9' '-' | cut -c1-40 | sed 's/^-//; s/-$//')
fi
ITEM_KEY="needs_diagnosis:${SLUG}"

# ── 0. pre-dispatch coverage check (header note 6) ───────────────────────────
# Every probe excludes our own ITEM_KEY: re-running the same intake is the
# documented idempotency (idx_swi_dedup + ON CONFLICT DO NOTHING), not a
# collision. Probes fail closed: if the DB is unreachable, we do not dispatch.
COVERAGE_HITS=0

run_coverage_probe() { # $1 label, $2 sql — prints hits, counts them
  local label="$1" sql="$2" rows
  rows=$(printf '%s' "$sql" | "${PSQL[@]}" -t -A -F' | ')
  if [ -n "$rows" ]; then
    echo ""
    echo "!! COVERED — ${label}:"
    printf '%s\n' "$rows" | sed 's/^/     /'
    COVERAGE_HITS=$((COVERAGE_HITS + 1))
  fi
}

# comma-separated -> SQL 'a','b' list. Entries must not contain single quotes.
sql_list() { printf "'%s'" "$(printf '%s' "$1" | sed "s/,/','/g")"; }

echo "-- pre-dispatch coverage check (FORCE=$FORCE) ..."

# resolve the target site if we have one (for the site-level probe only —
# the page probe deliberately searches ALL sites)
TARGET_SITE_ID="$SITE_ID"
if [ -z "$TARGET_SITE_ID" ] && [ -n "$RUNTIME_SITE" ]; then
  TARGET_SITE_ID=$(printf '%s' \
    "SELECT id FROM sites WHERE domain = $(sql_list "$RUNTIME_SITE");" \
    | "${PSQL[@]}" -t -A)
fi

if [ -n "$PAGES" ]; then
  PAGES_SQL=$(sql_list "$PAGES")
  run_coverage_probe "open work items on these pages (any site)" "
    SELECT s.domain || ' ' || p.name, w.item_key, w.status, w.created_by,
           to_char(w.created_at AT TIME ZONE 'UTC','YYYY-MM-DD HH24:MI') || 'Z'
    FROM site_work_items w
    JOIN pages p ON p.site_id = w.site_id
    JOIN sites s ON s.id = p.site_id
    WHERE (p.name IN (${PAGES_SQL}) OR p.url IN (${PAGES_SQL}))
      AND w.status NOT IN ('complete','cancelled','rejected')
      AND w.item_key IS DISTINCT FROM '${ITEM_KEY}'
      AND (w.page_id = p.id
           OR w.spec->>'page_id' = p.id::text
           OR w.item_key LIKE '%:' || p.name
           OR w.item_key LIKE '%:' || p.name || ':%')
    ORDER BY 5 DESC;"
elif [ -n "$TARGET_SITE_ID" ]; then
  # no pages named: block on recently-touched open items for the target site,
  # list (don't block on) the older open backlog
  run_coverage_probe "open items on the target site touched in the last ${RECENT_WINDOW}" "
    SELECT w.item_key, w.item_type, w.status, w.created_by,
           to_char(GREATEST(w.created_at, w.updated_at) AT TIME ZONE 'UTC','YYYY-MM-DD HH24:MI') || 'Z'
    FROM site_work_items w
    WHERE w.site_id = '${TARGET_SITE_ID}'
      AND w.status NOT IN ('complete','cancelled','rejected')
      AND w.item_key IS DISTINCT FROM '${ITEM_KEY}'
      AND GREATEST(w.created_at, w.updated_at) > now() - interval '${RECENT_WINDOW}'
    ORDER BY 5 DESC;"
  OLDER_OPEN=$(printf '%s' "
    SELECT count(*) FROM site_work_items w
    WHERE w.site_id = '${TARGET_SITE_ID}'
      AND w.status NOT IN ('complete','cancelled','rejected')
      AND w.item_key IS DISTINCT FROM '${ITEM_KEY}'
      AND GREATEST(w.created_at, w.updated_at) <= now() - interval '${RECENT_WINDOW}';" \
    | "${PSQL[@]}" -t -A)
  [ "${OLDER_OPEN:-0}" != "0" ] && \
    echo "   FYI: ${OLDER_OPEN} older open item(s) on this site (not blocking; not touched in ${RECENT_WINDOW})"
fi

if [ -n "$SEED_SCOPE" ]; then
  SEED_FILES=$(printf '%s' "$SEED_SCOPE" | tr ',' '\n' | cut -d: -f1 | paste -sd, -)
  run_coverage_probe "open diagnose items sharing a seed_scope file" "
    SELECT w.item_key, w.status, w.created_by,
           to_char(w.created_at AT TIME ZONE 'UTC','YYYY-MM-DD HH24:MI') || 'Z', e.entry
    FROM site_work_items w,
         LATERAL jsonb_array_elements_text(COALESCE(w.spec->'seed_scope','[]'::jsonb)) e(entry)
    WHERE w.pipeline = 'diagnose'
      AND w.status NOT IN ('complete','cancelled','rejected')
      AND w.item_key IS DISTINCT FROM '${ITEM_KEY}'
      AND split_part(e.entry, ':', 1) IN ($(sql_list "$SEED_FILES"))
    ORDER BY 4 DESC;"
  # advisory only: uncommitted local edits to the seed files mean a session
  # (possibly this one) is mid-change; the diagnosis reads REF, not this tree
  if git rev-parse --show-toplevel >/dev/null 2>&1; then
    DIRTY=$(cd "$(git rev-parse --show-toplevel)" && \
            git status --porcelain -- $(printf '%s' "$SEED_FILES" | tr ',' ' ') 2>/dev/null || true)
    if [ -n "$DIRTY" ]; then
      echo ""
      echo "   ADVISORY (not blocking): seed files have uncommitted local edits —"
      printf '%s\n' "$DIRTY" | sed 's/^/     /'
      echo "   the diagnosis will read ${REF}, not this working tree."
    fi
  fi
fi

if [ -z "$PAGES" ] && [ -z "$TARGET_SITE_ID" ] && [ -z "$SEED_SCOPE" ]; then
  echo "   WARNING: nothing to key coverage on (no PAGES, no site, no SEED_SCOPE) — dispatching blind."
fi

if [ "$COVERAGE_HITS" -gt 0 ]; then
  if [ "$FORCE" = "1" ]; then
    echo ""
    echo "FORCE=1 — dispatching despite ${COVERAGE_HITS} coverage hit(s) above."
  else
    cat <<EOF

=========================================
REFUSING TO DISPATCH: open work already touches this target (see above).
Another session may be fixing what you are about to diagnose — that exact
collision cost a diagnosis run on 2026-07-16 (correlation 781ea4f7).
If you have looked at the items above and still want the run:
  FORCE=1 <same command>
=========================================
EOF
    exit 1
  fi
else
  echo "   coverage: clear."
fi

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)

# ── the envelope, built once and shared by the item spec and the Kafka message ─
INPUT_DATA="{\"symptom\":\"$SYMPTOM\",\"owner\":\"$OWNER\",\"repo\":\"$REPO\",\"ref\":\"$REF\""
[ -n "$RUNTIME_SITE" ] && INPUT_DATA="${INPUT_DATA},\"runtime_site\":\"$RUNTIME_SITE\""
[ -n "$SITE_ID" ]      && INPUT_DATA="${INPUT_DATA},\"site_id\":\"$SITE_ID\""
if [ -n "$SEED_SCOPE" ]; then
  # comma-separated -> JSON array. ExtractStringListHelper takes []interface{}
  # or []string only; a bare "a,b" string yields nil and the seed is ignored.
  SCOPE_JSON=$(printf '%s' "$SEED_SCOPE" | awk -F, '{for(i=1;i<=NF;i++){printf "%s\"%s\"", (i>1?",":""), $i}}')
  INPUT_DATA="${INPUT_DATA},\"seed_scope\":[${SCOPE_JSON}]"
fi
if [ -n "$SUBJECT_TYPE" ] && [ -n "$SUBJECT_KEY" ]; then
  INPUT_DATA="${INPUT_DATA},\"subject_type\":\"$SUBJECT_TYPE\",\"subject_key\":\"$SUBJECT_KEY\""
fi
INPUT_DATA="${INPUT_DATA},\"correlation_id\":\"$CORRELATION_ID\"}"

echo "========================================="
echo "needs_diagnosis intake  (pipeline='diagnose')"
echo "========================================="
echo "  Symptom:       ${SYMPTOM}"
echo "  item_key:      ${ITEM_KEY}"
echo "  Anchor site:   system.internal  (target site travels in spec)"
echo "  Repo:          ${OWNER}/${REPO} @ ${REF}   (explicit — never HEAD)"
[ -n "$RUNTIME_SITE" ] && echo "  Under diagnosis: ${RUNTIME_SITE}"
[ -z "${RUNTIME_SITE}${SITE_ID}" ] && echo "  Anchor:        NONE (code-only) — needs the load_runtime error_step migration"
if [ -n "$SUBJECT_TYPE" ] && [ -n "$SUBJECT_KEY" ]; then
  echo "  Subject:       ${SUBJECT_TYPE}/${SUBJECT_KEY}   (persist_note WRITES a doc_notes row, post-3b)"
else
  echo "  Subject:       NONE — persist_note will SKIP (by design; do not guess)"
fi
echo "  Correlation:   ${CORRELATION_ID}"
echo "========================================="
echo "SAVE: CORRELATION_ID=${CORRELATION_ID}"
echo ""

# ── 1. the durable intake record ──────────────────────────────────────────────
# ON CONFLICT DO NOTHING pairs with idx_swi_dedup: a second intake for a still-open
# slug is silently ignored. rows_written tells you which happened.
echo "-- writing needs_diagnosis work item ..."
"${PSQL[@]}" <<SQL
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key, max_attempts
) VALUES (
    '${SYSTEM_SITE_ID}', '090_TRIGGER_needs_diagnosis', 'diagnose', 'needs_diagnosis',
    'medium', \$sum\$${SYMPTOM}\$sum\$,
    \$json\$${INPUT_DATA}\$json\$::jsonb,
    50, '${TARGET_AGENT_TYPE}', 'awaiting_diagnosis', '090_TRIGGER_needs_diagnosis', '${ITEM_KEY}',
    1   -- see header note 5: makes the 40-minute claim reset terminal, not 'triaged'
)
ON CONFLICT DO NOTHING;

SELECT count(*) AS open_intakes_for_this_slug
FROM site_work_items
WHERE site_id = '${SYSTEM_SITE_ID}' AND item_key = '${ITEM_KEY}'
  AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved');
SQL

if [ "$DISPATCH" != "1" ]; then
  echo ""
  echo "DISPATCH=0 — intake recorded, no Kafka message sent."
  exit 0
fi

# ── 2. dispatch, on the SAME correlation_id ───────────────────────────────────
echo ""
echo "-- dispatching to ${TARGET_AGENT_TYPE} ..."
kubectl -n kafka run -i --rm "kcat-diagnose-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=needs-diagnosis-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"$TARGET_AGENT_TYPE"},"input_data":$INPUT_DATA}
JSON

# bugs_open/030: the queries below return 0 rows until the dispatch reaches the
# head of the single generic requests lane — indistinguishable from a drop, and
# it has already ended one investigation with a TODO. Print the lane depth so
# nobody re-fires. Advisory and fail-soft; never let it break a dispatch.
QUEUE_REPORT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && git rev-parse --show-toplevel 2>/dev/null || true)/scripts/dispatch-queue-depth.sh"
if [[ -x "$QUEUE_REPORT" ]]; then
  echo ""
  "$QUEUE_REPORT" || true
fi

cat <<EOF

=========================================
intake recorded and dispatched.
=========================================

1) Follow the run:
   kubectl -n ai-persona-system logs -f -l agent-type=diagnose-agent --tail=500 | grep '$CORRELATION_ID'

2) Orchestration state (by correlation id, never by created_at):
   SELECT status, current_step, EXTRACT(EPOCH FROM (NOW() - last_activity))::int AS since_s,
          substring(COALESCE(error,''),1,200) AS err
   FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid ORDER BY created_at;

3) FETCH THE BUNDLES (F0.1 — this is the egress route; needs the chassis image
   carrying the diagnose_assemble_bundle write-through):
   SELECT iteration, length(body) AS bytes, metadata->>'symbol_count' AS symbols,
          metadata->>'truncated' AS truncated, created_at
   FROM diagnosis_artifacts
   WHERE correlation_id = '$CORRELATION_ID' AND kind = 'bundle'
   ORDER BY iteration;

   -- one bundle to a file:
   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db \\
     -t -A -c "SELECT body FROM diagnosis_artifacts WHERE correlation_id='$CORRELATION_ID' AND kind='bundle' AND iteration=1" \\
     > bundle_iter1.md

4) The intake record, and closing it by hand until a diagnose dispatch loop exists:
   SELECT status, created_at, completed_at FROM site_work_items WHERE item_key = '$ITEM_KEY';
   UPDATE site_work_items SET status='complete', completed_at=now()
   WHERE item_key = '$ITEM_KEY' AND status NOT IN ('complete','verified');

Timing: minutes, not seconds (repo tarball fetch + up to 5 verdict iterations).
EOF
