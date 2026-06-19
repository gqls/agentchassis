# Runbook — rebuild gamesdesign.co.uk `index` and verify the shipped changes

Purpose: trigger one `index` rebuild and confirm the three changes deployed
2026-06-18 behave, end to end.

What this exercises:
1. **Flatten contract** — `page-content-writer`'s singular `output_field` now flattens
   into the response, so `page_content.response.sections_metadata` resolves for
   `page-build-handler` (was dropped → silent no-op).
2. **Contract logging** — `resolveResultSpec` emits a resolved-contract line, a
   deprecation Warn per old key name (a census for the rename follow-up), and a
   conflict Warn when more than one key is present.
3. **Silent-success hardening** — an undeliverable (oversize) result now fails
   loudly (`error_unrecoverable` + `agent_error_log`) instead of a `status:"completed"`
   stub.

Read this first: a rebuild can legitimately stop at one of several stages. This
site's history shows all three of: a silent stub (06-13), **loud** `save_sections`
"content regression blocked" failures (06-14), and `deploy_page` `call_agent`
timeouts (06-06/06-08). The change above fixes the silent stub and makes the
others diagnosable — it does not by itself guarantee the page updates. §5 walks
the pipeline stage by stage so you can see where it stops.

No `perfect`/final claims here — this is a check, run it and read the stages.

---

## 1. Preconditions

- The chassis image carrying the fix is rolled out, i.e. the agents are **off**
  `v1.0.1063`. Confirm:
  ```bash
  kubectl -n ai-persona-system get deploy -o wide | grep -E 'page-content-writer|page-build-handler'
  # IMAGE column should show the new chassis tag, not v1.0.1063
  ```
- A psql helper (used throughout):
  ```bash
  PG() { kubectl exec -n ai-persona-system -i postgres-clients-0 -- \
         psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 "$@"; }
  ```
- Log access. Agent deployments are named `agent-<agent-type>`. If your label key
  isn't `app`, adjust the selector; `kubectl -n ai-persona-system get pods | grep page-content-writer`
  finds the pods.
  ```bash
  WLOG() { kubectl -n ai-persona-system logs -l app=agent-page-content-writer --prefix --since=20m --tail=-1 "$@"; }
  HLOG() { kubectl -n ai-persona-system logs -l app=agent-page-build-handler   --prefix --since=20m --tail=-1 "$@"; }
  ```

---

## 2. Capture the baseline (before)

Resolve the ids and snapshot what exists, so "did anything change" is answerable.

```bash
# page_id + site_id for the index page
PG -c "SELECT id AS page_id, site_id, build_status, last_built_at, deployed_at, version
       FROM pages WHERE url ILIKE '%gamesdesign%' AND name='index';"
```

Set `PAGE_ID` and `SITE_ID` from that row, then:

```bash
# component fingerprint — the real 'did it change' signal
PG -c "SELECT position, slot_name, build_status,
              left(content_hash,12) AS hash12,
              length(rendered_html)  AS html_len,
              updated_at, deploy_commit
       FROM page_components WHERE page_id='$PAGE_ID' ORDER BY position;"

# the most recent page_rerender work item (this is what we'll re-open)
PG -c "SELECT id, item_type, status, attempt_count||'/'||max_attempts AS attempts,
              claimed_by, error, created_at, completed_at, item_key
       FROM site_work_items
       WHERE site_id='$SITE_ID' AND item_type='page_rerender'
       ORDER BY created_at DESC LIMIT 3;"

# recent failures on this site's build path (read these BEFORE the run)
PG -c "SELECT occurred_at, agent_type, step_name, action, left(error,90) AS error
       FROM agent_error_log
       WHERE agent_type IN ('page-build-handler','page-content-writer')
       ORDER BY occurred_at DESC LIMIT 10;"
```

Note the baseline `max(updated_at)` and the per-position `hash12`. The deployed
page was written 2026-06-06 16:59 (5 sections, ~34k) and has not moved since.

---

## 3. Trigger the rebuild

Re-open the existing `page_rerender` work item rather than fabricating one — its
`spec` is already the shape the dispatch loop expects. The `LIMIT 1` subquery
targets exactly one row (the dedup index forbids two active rows with the same
`item_key`).

```bash
PG -c "
UPDATE site_work_items
SET status='approved',          -- the status your build-dispatch-loop claims
    attempt_count=0, claimed_by=NULL, claimed_at=NULL,
    completed_at=NULL, error=NULL, result='{}'::jsonb, updated_at=now()
WHERE id = (
  SELECT id FROM site_work_items
  WHERE site_id='$SITE_ID' AND item_type='page_rerender'
  ORDER BY created_at DESC LIMIT 1
)
RETURNING id, item_type, status, item_key;"
```

Two things to confirm for your environment, both quick:
- **Claimable status.** The pending indexes are on `status IN ('triaged','approved')`.
  `approved` is the furthest-along claimable state; if your dispatch query claims
  `triaged` instead, use that. Check with `grep -rn "FROM site_work_items" platform/orchestration/actions/*work_item*`.
- **The dispatch loop is running.** It claims `approved`/`triaged` items on a cycle.
  If it isn't active, the item will just sit at `approved`.

Alternative (lets discovery create the item the production way): flag the page and
wait for the next discovery sweep.
```bash
PG -c "UPDATE pages SET build_status='needs_rebuild', updated_at=now() WHERE id='$PAGE_ID';"
```

Confirm it got claimed (status should move `approved` → `claimed` → terminal):
```bash
PG -c "SELECT status, claimed_by, attempt_count||'/'||max_attempts AS attempts,
              claimed_at, completed_at, left(error,120) AS error
       FROM site_work_items WHERE id='<WORK_ITEM_ID_FROM_RETURNING>';"
```

---

## 4. Watch the logs (during)

The writer's contract resolution — the heart of fix #1:
```bash
WLOG | grep -E 'resolveResultSpec: resolved result contract' | tail -5
# expect a line with mode=flatten and matched_key=output_field

WLOG | grep -E 'extractWorkflowResult: result built' | tail -5
# expect mode=flatten and size_bytes in the tens-of-thousands (~80k), field_count>1

WLOG | grep -E 'result_from field not found' | tail -5
# expect NOTHING — if present, the writer's complete step lost its field name
```

The deprecation census (fix #2) — across all agents, not just the writer:
```bash
kubectl -n ai-persona-system logs -l app --prefix --since=20m --tail=-1 2>/dev/null \
  | grep -E 'resolveResultSpec: deprecated complete-step key in use' \
  | grep -oE 'deprecated_key[":= ]+[a-z_]+' | sort | uniq -c
# this is the list to feed the rename migration (output_field/output_fields/output)

kubectl -n ai-persona-system logs -l app --prefix --since=20m --tail=-1 2>/dev/null \
  | grep -E 'multiple result-contract keys present' | tail -5
# expect NOTHING — if present, an agent has two contract keys; resolve it
```

The silent-success guard (fix #3) — should stay quiet on this run:
```bash
WLOG | grep -E 'result exceeds delivery cap|undeliverable — notifying parent of FAILURE' | tail -5
# expect NOTHING for the writer (flatten bounds it well under the cap)

kubectl -n ai-persona-system logs -l app --prefix --since=20m --tail=-1 2>/dev/null \
  | grep -E 'Full result exceeded size limit' | tail -5
# expect NOTHING anywhere — that stub string is removed; any hit means a stale image
```

The save and deploy steps on the handler:
```bash
HLOG | grep -E 'save_sections|save_page_sections|content regression|deploy_page' | tail -20
```

---

## 5. Verify the result (after) — stage by stage

Run these in order; the first one that fails tells you where the pipeline stopped.

**Stage A — the writer delivered a flat result to the parent.** Find the run, then
read the parent's collected_data at the flat path.
```bash
PG -c "SELECT orchestration_id, owner_agent_type, status, current_step, created_at
       FROM orchestration_states
       WHERE owner_agent_type IN ('page-build-handler','page-content-writer')
         AND created_at > now() - interval '30 minutes'
       ORDER BY created_at DESC;"
```
For the `page-build-handler` orchestration_id from that list:
```bash
PG -c "SELECT
         jsonb_typeof(collected_data #> '{page_content,response,sections_metadata}')      AS sm_type,
         jsonb_array_length(collected_data #> '{page_content,response,sections_metadata}') AS sm_count,
         (collected_data #> '{page_content,response,page_html}') IS NOT NULL               AS has_html,
         collected_data #>> '{page_content,response,message}'                              AS stub_message
       FROM orchestration_states WHERE orchestration_id='<PARENT_ORCH_ID>';"
```
- `sm_type='array'`, `sm_count` = number of sections, `has_html=true` → **flatten works**.
- `sm_type` null → flatten did not populate; the writer response didn't carry
  `sections_metadata` (look at the writer run, Stage A is the fix's direct target).
- `stub_message` non-null → a stub slipped through (should be impossible post-fix;
  means a stale image).

**Stage B — the save was attempted, and passed or was blocked (loudly).**
```bash
PG -c "SELECT occurred_at, step_name, action, left(error,140) AS error
       FROM agent_error_log
       WHERE agent_type='page-build-handler' AND occurred_at > now() - interval '30 minutes'
       ORDER BY occurred_at DESC;"
```
- No `save_sections` error + components changed (Stage C) → **save succeeded**.
- `content regression blocked: new content has N chars vs M existing` → the save was
  reached (so Stage A is fine) but the new content is thinner than what's live. This
  is the 06-14 failure mode: a loud, correct block, not a silent no-op. Next thread
  is **why the writer produced thin content** (candidate: the writer's research/content
  sub-agents — see the research-agent `sources` mapping and content-reviewer follow-ups
  in the notes — not this fix). Do not disable the guard to force it through.

**Stage C — components actually changed.** Re-run the Stage-2 fingerprint and diff:
```bash
PG -c "SELECT position, slot_name, left(content_hash,12) AS hash12,
              length(rendered_html) AS html_len, updated_at, deploy_commit
       FROM page_components WHERE page_id='$PAGE_ID' ORDER BY position;"
```
- `updated_at` advanced past 2026-06-06 **and** `hash12` differs from baseline →
  content was rewritten and saved. `updated_at` unchanged → nothing was saved (read
  Stage B for the reason).

**Stage D — the work item completed on a real save, not a false complete.**
```bash
PG -c "SELECT status, completed_at, result, left(error,140) AS error
       FROM site_work_items WHERE id='<WORK_ITEM_ID>';"
```
- The distinguishing test from the old bug: `status='complete'` is only meaningful
  **together with Stage C showing changed components**. `complete` + unchanged
  components = the old false-complete; that combination should no longer occur.
- `status='failed'`/`needs_human_review` with an error → it failed loudly (acceptable;
  read the error and Stage B).

**Stage E — deploy (optional).** The page row and deploy.
```bash
PG -c "SELECT build_status, last_built_at, deployed_at, version FROM pages WHERE id='$PAGE_ID';"
```
- `deploy_page`/`call_agent` timeouts in Stage B's error list → the build saved but
  the deploy step timed out (06-06/06-08 mode). Check the deploy agent and the
  GitHub→Backblaze path separately.

---

## 6. Verify the silent-success hardening

Fix #3 is a safety net that should rarely fire now. Verify it passively, plus an
optional active test.

Passive — the stub is gone and the failure path is the only oversize outcome:
```bash
# no stub rows anywhere recent
PG -c "SELECT count(*) AS stub_rows
       FROM orchestration_states
       WHERE updated_at > now() - interval '1 day'
         AND collected_data::text LIKE '%Full result exceeded size limit%';"
# expect 0

# if any oversize DID occur, it logged a fatal error naming the largest field
PG -c "SELECT occurred_at, agent_type, step_name, left(error,160) AS error
       FROM agent_error_log
       WHERE occurred_at > now() - interval '1 day'
         AND error ILIKE '%exceeds the%delivery cap%'
       ORDER BY occurred_at DESC;"
# 0 rows on a healthy run; any row here is an agent that needs a result contract
```

Active (optional, dev/staging only — do not do this in prod): the cap is a
compile-time const (`MaxResultSizeBytes`), so the clean way to exercise the failure
path is to lower it temporarily in a dev build, run any agent, and confirm it
emits `result exceeds delivery cap` → `notifyParentOfFailure` → an
`error_unrecoverable` response and a `CHILD_ORCHESTRATION_FAILED` `agent_error_log`
entry, with the page **not** marked complete. Restore the const after.

---

## 7. If the page still doesn't update — triage

| Symptom (from the stages above) | Stage | Likely cause | Next check |
|---|---|---|---|
| Writer log shows `mode=fallback`, not flatten | A | The writer's `complete` step lost its `output_field` | `grep "complete" 023_page_content_writer_agent.sql`; check live agent_definitions complete config |
| `sm_type` null on the parent | A | Writer response didn't carry `sections_metadata` | Read the writer orchestration's own `collected_data`; confirm its compile step produced sections_metadata |
| `content regression blocked` in errors | B | New content thinner than live | Writer sub-agent data flow (research `sources` mapping, content-reviewer); not this fix |
| `save_sections` other error | B | Save action failure | Read full `error` text; `save_page_sections_action.go` |
| `complete` work item but components unchanged | C/D | Would be the old false-complete | Confirm the running image is the new tag, not `v1.0.1063` |
| `deploy_page`/`call_agent` timed out | E | Deploy step / downstream agent | Deploy agent logs; GitHub Actions → Backblaze |
| Work item stuck at `approved` | 3 | Dispatch loop not claiming, or wrong status | Confirm the loop is running and the claim status matches §3 |
| `Full result exceeded size limit` appears | any | Stale chassis image still deployed | Re-roll the image; confirm tag |

Reminder on 0-row results: a query returning 0 rows is not proof of absence until
the query itself is confirmed — the child-side path has no `.response` wrapper
(that exists only on the parent), so querying the parent path against a child row
returns 0 rows as an artefact.

---

## 8. Rollback

The deploy is image-tag based. To revert, point the affected agents back at the
prior chassis tag:
```bash
PG -c "SELECT type, image_tag FROM agent_definitions
       WHERE type IN ('page-content-writer','page-build-handler');"   -- confirm column names
# then set image_tag back to v1.0.1063 for those agents and re-roll,
# OR redeploy the previous chassis image under the new tag.
```
Reverting restores the prior behaviour (the silent stub returns), so only roll back
if the new build is actively worse than the known no-op.

---

## What "good" looks like for this run

Writer log shows `mode=flatten matched_key=output_field`; the parent's
`page_content.response.sections_metadata` is a populated array; `page_components`
`updated_at` advances past 2026-06-06 with changed hashes; the `page_rerender` work
item is `complete` **and** components changed; no `Full result exceeded size limit`
anywhere. If it stops earlier, the stage that failed names the next thread — and
unlike before, it stops with a reason in the logs rather than a silent success.
