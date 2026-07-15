# FOCUS: Dispatch Loop and the `detected → triaged → claimed` State Machine

**Status:** Architecture understood. Workarounds documented. Future work optional, listed at the end.

**Supersedes:** an earlier version of this doc from 2026-05-14 that hypothesised a "missing detected→triaged transition." That hypothesis was wrong; the transition exists. See "Evidence trail" below for what was actually found.

---

## TL;DR

`detected` is a valid intermediate state, not a bug.

```
discovery emits  →  detected
                       ↓
                    (design-audit-agent runs visual + content
                     auditors, then calls triage_detected_items)
                       ↓
                    triaged
                       ↓
                    (dispatch loop claims; partial indexes
                     idx_swi_handler and idx_swi_site_pending
                     target this status)
                       ↓
                    claimed
                       ↓
                    (handler runs; marks complete or failed
                     directly or via the new mark_work_item_complete
                     step)
                       ↓
                    complete / failed
```

The 8 `needs_imagery` items stuck in `detected` on robot-hands.com (2026-05-14) are NOT a dispatch failure. They're items that haven't been through the audit/triage pass yet. Dispatch is correctly ignoring them because they aren't yet triaged.

---

## Evidence trail

What's in the code:

1. **`triage_detected_items` action is registered** at `registry.go:722`:
   ```go
   "triage_detected_items": {
       Handler:     TriageDetectedItemsAction,
       Category:    "maintenance",
       Description: "Promote detected discovery items to triaged with target domain",
       IsLocal:     true,
   },
   ```

2. **Index definitions explicitly target triaged/approved** (`site_work_items`):
   ```
   idx_swi_handler        WHERE status = ANY (ARRAY['triaged','approved'])
   idx_swi_site_pending   WHERE status = ANY (ARRAY['triaged','approved'])
   ```
   These exist to serve dispatch's "find claimable items" query. They're the strongest evidence of what status dispatch reads.

3. **`site_admin_handlers.go` confirms `detected` is intentional.** Two relevant patterns:
   - Line 455: admin-created items are inserted directly at `status='triaged'` because they bypass discovery — they don't need triaging.
   - Line 749: there's an explicit endpoint for operator-driven `status='triaged'` transitions.

What's in the prior session transcript (April 2026, recovered from `/mnt/transcripts/`):

4. **`design-audit-agent` workflow includes a `triage` step** that calls `triage_detected_items`:
   ```json
   {
     "steps": {
       "ensure_site_record":     {..., "next_step": "spawn_visual_auditor"},
       "spawn_visual_auditor":   {..., "next_step": "call_visual_auditor"},
       "call_visual_auditor":    {..., "next_step": "spawn_content_auditor"},
       "spawn_content_auditor":  {..., "next_step": "call_content_auditor"},
       "call_content_auditor":   {..., "next_step": "triage"},
       "triage": {
         "action": "triage_detected_items",
         "config": {"site_id": "site_record.site_id", "target_domain": "build"},
         "next_step": "complete"
       }
     }
   }
   ```

5. **STATUS_imagery_2026-05-12.md establishes both states as routine:**
   > The improvement loop / dispatch loop is currently OFF (user request, noise reduction). Variant items sit `triaged` waiting.

   Items routinely sit in `triaged` because dispatch is off — and they routinely transition from `detected → triaged` via audit. Both states are valid intermediate steps.

---

## Why our 8 items are stuck

Sequence of events this session (2026-05-14):

1. User triggered `design-discovery-agent` directly via kcat at 18:54 UTC.
2. The agent's `unfulfilled_imagery_plan` check (Phase 2G.4) emitted 8 work items at `status='detected'`. This is correct behaviour — that's the output state for discovery.
3. `design-audit-agent` was NOT triggered afterwards.
4. The dispatch loop continued running every ~60s, correctly ignoring `detected` items (its query targets `triaged`/`approved`).
5. Items sit indefinitely in `detected`.

There's no automated coupling between discovery and audit. They're separate triggers with separate workflows. A site can have items in `detected` for arbitrary duration without anything being wrong with dispatch or any other component.

---

## How to unstick the queue

Three options, ordered by alignment with the architecture:

### Option A — Run `design-audit-agent` (recommended)

Most aligned with the system's intended flow. The audit agent will run visual + content auditors first (which may emit additional findings), then call `triage_detected_items` which promotes all detected items on the site to triaged. Dispatch claims them on its next tick.

Trigger the same way as `design-discovery-agent`:

```bash
set -u
AGENT_TYPE="design-audit-agent"
SITE_ID="00ff3af5-dad8-4770-9f70-3edc267a3c92"
DOMAIN="robot-hands.com"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

kubectl -n kafka run -i --rm "kcat-audit-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H message_type=request \
  -H client_id=$CLIENT_ID \
  -H action=orchestrate \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON
```

Side effect: audit findings may add to the queue. Acceptable in a normal cycle but if you only want to unstick the current 8 without inviting more work, use option B or C.

### Option B — Call `triage_detected_items` directly

Bypasses audit. Cheapest if you only want the existing 8 items moved. Requires constructing an ad-hoc workflow that just runs that one action. Not the normal pattern; only useful as a manual operator step.

### Option C — Direct SQL

For a one-shot operator intervention:

```sql
UPDATE site_work_items
   SET status = 'triaged',
       triaged_at = now()
 WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
   AND item_type = 'needs_imagery'
   AND status = 'detected';
```

Then watch the next dispatch tick (~60s) — items should transition `triaged → claimed`. Within 2-3 minutes the icon item should be processing through image-build-handler.

After option A or C, dispatch demonstrates end-to-end functionality for `needs_imagery` work items. Worth doing if only to close the loop verification.

---

## What this is NOT

- **NOT a dispatch bug.** Dispatch is working correctly.
- **NOT a missing state transition.** The transition exists; it's owned by audit, not dispatch.
- **NOT specific to `needs_imagery`.** Any item type emitted by a discovery check sits in `detected` until audit runs.
- **NOT a `handler_agent` issue.** The handler routing is set correctly on emission.

---

## Open architectural questions (not blocking)

These are real questions, but answering them isn't required to operate the system today. They're parked here for future consideration.

### Q1. Should discovery agents auto-triage their own emissions for low-risk item types?

Pro: closes the discovery→triaged gap immediately, no audit-agent dependency. Items like `needs_imagery` (cost-bounded, plan-locked, deletion-cheap) probably don't need audit review before generation.

Con: removes a centralised review point. If discovery auto-triages, audit findings are the only things that go through `triage_detected_items` — and audit findings may genuinely benefit from human or LLM review before becoming actionable.

A reasonable middle ground: per-check declarative flag (`auto_triage_emissions: true|false`) on the DiscoveryCheck interface. Most checks set true; a few (the ones that genuinely need review, e.g., audit findings) keep the manual triage step.

### Q2. Is there a scheduler that runs `design-audit-agent` automatically?

There's a `scheduled_tasks` table referenced by `build-dispatch-loop` (the `notify_scheduler` step at the end of dispatch's workflow). If similar entries exist for audit, then items DO get triaged automatically on a cadence — just not within seconds of being emitted.

Worth checking next time someone's near the data:

```sql
SELECT name, last_started_at, last_completed_at,
       EXTRACT(EPOCH FROM (now() - COALESCE(last_completed_at, last_started_at)))::int AS seconds_since_run
FROM scheduled_tasks
ORDER BY name;
```

If `design-audit-agent` is scheduled but its `last_completed_at` is stale, that's a different issue (scheduler stuck). If it's not scheduled at all, then triaging is operator-driven.

### Q3. The dispatcher architecture — one-site-per-tick, NOT-EXISTS-blocked (researched 2026-05-15)

STATUS_imagery_2026-05-12.md flagged this in passing:
> Build-dispatch-loop isn't claiming triaged image items when page items are also in queue. Imagery items have been sitting in `triaged` for hours while page work flows.

The original hypothesis was "priority race within dispatch's claim logic." That hypothesis turned out to be wrong. The actual mechanism is a two-stage selection chain with a hard exclusion clause. Reasoning trail follows.

**Reasoning trail (2026-05-15):**

After running audit+triage successfully, the 7 remaining `needs_imagery` items were `triaged`, `pipeline='build'`, `priority=1`, with `attempt_count=0`. Yet 30+ minutes of dispatch ticks went by without claiming them. Initial hypothesis was the pipeline mismatch (Q4); fixing that didn't unblock. Lock check (`sites.locked_at`) returned NULL — not locked. So neither the field-routing nor the site-lock explanation held.

The dispatch trace `had_items: false` repeatedly for our site_id meant **the dispatcher was being invoked but finding nothing in the queue for our site**. That doesn't fit either of "dispatcher is broken" or "priority race" — it fits "dispatcher was invoked with a site_id where there's nothing to claim". So the question shifted to: **what invokes the dispatcher, and how does it choose which site to invoke it for?**

Two layers of indirection were involved. The `scheduled_tasks` table has a row `build-pipeline-trigger` running every 30s. That triggers an agent called `build-pipeline-trigger` (not the dispatcher itself). That agent's workflow contains a `find_dispatchable_site` step which contains the actual selection SQL:

```sql
SELECT DISTINCT ON (wi.site_id) wi.site_id::text, s.domain
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.status IN ('triaged', 'approved')
  AND wi.attempt_count < wi.max_attempts
  AND NOT EXISTS (
    SELECT 1 FROM site_work_items active
    WHERE active.site_id = wi.site_id
      AND active.status = 'claimed'
  )
ORDER BY wi.site_id, wi.priority ASC
LIMIT 1
```

Running this query manually returned `eac60db8-...` (system.internal), not robot-hands.com. So robot-hands.com is being **excluded by the selection criteria**, not just losing the `LIMIT 1` race. The most plausible exclusion is the `NOT EXISTS` clause — if any item on a site is in `status='claimed'`, the entire site is excluded from dispatch until that item clears.

**What this means architecturally:**

The dispatch chain is:

```
scheduled_tasks (every 30s)
  ↓
build-pipeline-trigger agent
  ├─ seed_queue       (handle new sites entering from build_queue)
  ├─ find_dispatchable_site  (the SQL above — one site per tick)
  ├─ check_has_site   (conditional: stop if none)
  └─ spawn build-dispatch-loop for that one site
      ↓
      build-dispatch-loop (now scoped to one site)
        ├─ load_items   (up to 5 items, pipeline='build', from this site)
        ├─ claim_work_item    (per item)
        └─ spawn_agent  (the handler)
```

This is a one-site-per-30s throughput cap. With many sites in queue, each site gets attention infrequently. Within a site, the inner dispatch loop processes 5 items per invocation. So the system can sustain ~5 items per site per 30s when it's that site's turn — but only one site is "its turn" at any given time.

**The NOT EXISTS clause is an absolute blocker, not a deprioritiser.** A single stuck `claimed` item on a site excludes the entire site from dispatch consideration until it clears or is reset. There's no fall-through. This is by design — it prevents racing claim attempts on a site that's already mid-execution — but it makes stuck claims a system-stopping condition for that site rather than a queue-position-degrading one.

**Why our site is excluded right now (hypothesis, still to confirm):** the audit run earlier claimed 2 `needs_design_review` items. If either is still `claimed` (rather than having transitioned to complete/failed), it's blocking dispatch for the whole site.

**Confirm with:**

```sql
SELECT id, item_type, status, claimed_at, claimed_by, attempt_count, max_attempts,
       EXTRACT(EPOCH FROM (now() - claimed_at))::int AS seconds_claimed,
       LEFT(COALESCE(error, ''), 200) AS err
FROM site_work_items
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND status = 'claimed';
```

If anything is genuinely stuck (claimed > 15 minutes ago, no orchestration making progress), reset to triaged:

```sql
UPDATE site_work_items
   SET status = 'triaged',
       claimed_at = NULL,
       claimed_by = NULL,
       attempt_count = attempt_count + 1
 WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
   AND status = 'claimed'
   AND claimed_at < now() - interval '15 minutes';
```

**Two secondary findings worth recording:**

1. **`build-pipeline-trigger` doesn't write `orchestration_states`** (or writes under a different owner_agent_type). The 34-type enumeration of `owner_agent_type` doesn't include it. This makes the trigger's decisions harder to trace post-hoc. Either the trigger should write state, or there should be a different audit trail. Worth doing because dispatch decisions are operationally important.
2. **`ORDER BY wi.site_id, wi.priority ASC` only orders within DISTINCT ON groups.** There's no outer `ORDER BY` across sites. So the "first" site picked depends on Postgres's scan order — effectively arbitrary among eligible sites. Multiple eligible sites compete fairly chaotically. A deterministic outer order (oldest-stuck-first, highest-pending-priority-first, round-robin) would improve fairness.

**Architectural improvements worth considering (not urgent):**

- **Auto-reset stuck claims.** Watchdog query that resets `claimed` items where `claimed_at > now() - interval '15 minutes'` (or some configurable threshold). Prevents single stuck handlers from indefinitely blocking a site.
- **Outer ORDER BY** in find_dispatchable_site for fairness. Currently arbitrary; should probably be `ORDER BY oldest-pending-item ASC` so sites that have waited longest go first.
- **build-pipeline-trigger writes orchestration_states** so the decision trail is auditable.

The first one is the cheapest and highest-leverage. Easy SQL job, no architectural change.

### Q4. What is the `pipeline` field actually for? (Surfaced 2026-05-15)

**Discovered:** After running the audit/triage cycle for robot-hands.com on 2026-05-15, the 8 `needs_imagery` items got promoted to `triaged` but **still sat unclaimed** for 11+ minutes despite being boosted to priority 1. The dispatch query `has_items=false` came back every tick. Reason: items were `pipeline='design'`; dispatch loads `item_pipeline: "build"` only. Pipeline mismatch.

**Immediate unblock:** UPDATE-d the 7 stuck items to `pipeline='build'`. (Items still didn't move — that's Q3's separate problem.)

**Where 'design' came from:** the `unfulfilled_imagery_plan` discovery check (added in Phase 2G.4) emitted them at `pipeline='design'`. Either explicitly set in the check's INSERT, or the parent discovery agent's helper defaulted to 'design' rather than the schema's 'build' default. Either way, a bug.

**Historical context (added after user input):** the column was originally called `domain` and was renamed to `pipeline`. That explains why the audit triage step uses `target_domain: "build"` rather than `target_pipeline: "build"` — the action's config keyword wasn't updated when the schema was renamed. The naming is semantically stale but functionally fine.

**Research into what the field is for:**

| Source | What it tells us |
|---|---|
| Schema | `pipeline TEXT NOT NULL DEFAULT 'build'` — required column, 'build' if nothing sets it |
| Every grep hit in production code | Only ever filters/sets `pipeline = 'build'` |
| Dispatch config (`build-dispatch-loop`) | `item_pipeline: "build"` — explicit filter |
| Audit's triage step config | `target_domain: "build"` — stale terminology from before the rename; writes `pipeline='build'` |
| `build-pipeline-trigger`'s find_dispatchable_site query | `WHERE wi.status IN ('triaged','approved')` — does NOT filter on pipeline! Pipeline filter only happens inside the dispatch loop's `load_items`, not at site-selection level |
| Values in the wild | `'build'`, `'maintenance'` (system.internal has `component_quality_scan` items at `pipeline='maintenance'`). Two known values; only one wired to a dispatcher. |
| `design-dispatch-loop` | Does not exist. Only `build-dispatch-loop` is wired up. |

**What the field is for, summarised:** a coarse routing label that allows multiple pipeline-specific dispatchers to coexist. Shape is built for it; the wiring isn't. The `maintenance` items on system.internal are dormant for the same reason our `design` items were — no `maintenance-dispatch-loop` exists to consume them.

**Why this is fragile:** the field duplicates information that `handler_agent` already encodes. Every `needs_imagery` item necessarily routes to `image-build-handler` — that's a build-side handler — so `pipeline='build'` is derivable. Letting the work item declare `pipeline` separately means two facts that must be kept in sync, with nothing enforcing it. We hit the divergence on day one. The `maintenance` items reveal there's a third inconsistency dormant on a different site.

**Decision reached (2026-05-15, with user):** leave the field as a soft, currently-unused routing label. Two concrete changes:

1. **Discovery checks INSERT `pipeline='build'`** — match the schema default and the only-currently-wired dispatcher. Specifically, fix `unfulfilled_imagery_plan` to not declare pipeline (let the schema default apply) or to explicitly write `'build'`.
2. **Loosen the dispatcher to accept any value in the field** — change `build-dispatch-loop`'s `load_items.item_pipeline: "build"` to either accept a list, accept `*`, or drop the filter entirely. Pipeline becomes informational, not load-bearing.

The future-flexibility benefit is preserved (you can still introduce a `maintenance-dispatch-loop` and have it filter to `pipeline='maintenance'`). The fragility is removed because no discovery check has to know which pipeline its handler belongs to.

**Not implemented yet.** The immediate unblock (manual UPDATE to `pipeline='build'`) keeps the loop closing while these changes are designed and shipped. The `unfulfilled_imagery_plan` check should be fixed before the next discovery check is written so the same bug doesn't get baked into Phase 4/5/6 emissions.

**Worth recording:** the `maintenance` value's existence on system.internal means there's already real evidence of multi-pipeline-by-name. If the system grows to need real isolation between pipelines (different cadences, concurrency limits, error tolerances), the latent design is already there. The current state is "the shape exists, the wiring doesn't" — entirely fine to keep as-is until there's an operational need.

---

## Workarounds in use

For loops blocked on `detected → triaged`: trigger `design-audit-agent` to run the triage step (option A above).

For loops blocked on dispatch not claiming triaged items: manual `kcat` trigger of the handler directly. Pattern documented in `016_debugging_guide.md` section 9, with the psql-jsonb-builder script preferred:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -t -A -c "
SELECT jsonb_build_object(
  'action', 'orchestrate',
  'config', jsonb_build_object('agent_type', 'image-build-handler'),
  'input_data', jsonb_build_object(
    'site_id', '<site_id>',
    'domain', '<domain>',
    'work_item_id', '<work_item_id>',
    'item_type', 'needs_imagery',
    'spec', spec::jsonb
  )
)::text
FROM site_work_items
WHERE id = '<work_item_id>';
" > /tmp/trigger.json
```

Then pipe `/tmp/trigger.json` to kcat with standard headers.

Bookkeeping: the `mark_work_item_complete` step added in `phase_2g_followup_mark_work_item_complete.sql` correctly marks the work item complete when manually triggered, provided `input_data.work_item_id` is set. So manual triggers don't create the orphan `detected` problem any more.

---

## Cross-references

- `/PLAN_imagery_loop_closure.md` — broader plan; this dispatch-loop topic flagged under Open Items.
- `016_debugging_guide.md` section 9 — handler-side manual trigger workarounds.
- `registry.go:722` — `triage_detected_items` action registration (canonical evidence).
- `site_admin_handlers.go:455, :749` — admin-driven triaged-state creation and explicit transitions.
- `STATUS_imagery_2026-05-12.md` — historical context that should have informed the original hypothesis but didn't.
- `phase_2g_followup_mark_work_item_complete.sql` and `phase_2g_followup_mark_work_item_failed.sql` — handler-side bookkeeping (independent of this issue but completes the lifecycle).

---

## Lesson learned

The earlier version of this doc made the classic mistake of inferring system behaviour from one source (the partial-index definitions) without searching for upstream writers. Indexes told me where dispatch reads from; they didn't tell me where items get written to that state. A 30-second grep for `triage` would have surfaced both the registry entry and the design-audit-agent workflow that calls it.

This is reflected in `016_debugging_guide.md` Section 0 (Assumption Checklist) item 6 — "Sibling functions in the same file are the canonical pattern" — but it applies more broadly. When something looks like a missing transition, search the whole codebase for the relevant verb (`triage`, `promote`, `claim`, `complete`) before concluding the writer doesn't exist.
