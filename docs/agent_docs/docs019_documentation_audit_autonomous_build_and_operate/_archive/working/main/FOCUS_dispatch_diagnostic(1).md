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

### Q3. The "dispatch waits behind page work" issue from May 2026

STATUS_imagery_2026-05-12.md flagged this:
> Build-dispatch-loop isn't claiming triaged image items when page items are also in queue. Imagery items have been sitting in `triaged` for hours while page work flows.

This is a separate concern from the detected→triaged issue. It only manifests AFTER items are triaged, and it's about priority/concurrency within dispatch's claim logic. Worth investigating when imagery work starts queueing alongside other work.

The `load_work_items` action (referenced by dispatch's `load_items` step) is where this lives. Its config has `max_items: 5` and `item_pipeline: "build"`. Whether it prioritises by `priority` column, by `created_at`, or by item_type is in its Go source.

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
