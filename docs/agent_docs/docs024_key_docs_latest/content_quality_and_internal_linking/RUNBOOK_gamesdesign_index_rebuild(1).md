# Runbook — rebuild gamesdesign.co.uk `index` and verify the shipped changes

Purpose: trigger one `index` rebuild and confirm the three changes deployed
2026-06-18 behave, end to end.

What this exercises:
1. **Flatten contract** — `page-content-writer`'s singular `output_field` now
   flattens into the response, so `page_content.response.sections_metadata` resolves
   for `page-build-handler` (was dropped → silent no-op).
2. **Contract logging** — `resolveResultSpec` logs the resolved contract, a
   deprecation Warn per old key name (a census for the rename follow-up), and a
   conflict Warn when more than one key is present.
3. **Silent-success hardening** — an undeliverable (oversize) result now fails
   loudly (`error_unrecoverable` + `agent_error_log`) instead of a `status:"completed"`
   stub.

A rebuild can legitimately stop at one of several stages. This site's history has
all three of: a silent stub (06-13), **loud** `save_sections` "content regression
blocked" failures (06-14), and `deploy_page` timeouts (06-06/08). The change fixes
the stub and makes the others diagnosable — it does not by itself guarantee the
page updates. §5 walks the pipeline stage by stage.

**Where each step runs:** DB steps are bare SQL — paste them at your `clients_db=#`
prompt. Log and deploy steps run in a shell with `kubectl`. No wrapper functions.

**Known IDs for this site (resolved 2026-06-18):**
- site_id `e33263f4-74f8-494f-b191-546845dbbddf`  (`sites.domain` = gamesdesign.co.uk)
- index page_id `6e988cc4-4898-4021-aa5e-2ab0271f9b75`  (`name` = `index`, `url` = `/index.html`)
- index rerender item_key `page_rerender_index_e33263f4-74f8-494f-b191-546845dbbddf` (confirm in §3)

For a different page: `pages.url` is a path (`/index.html`), not the domain, so
don't filter pages by domain. Resolve via the site, then the page name:
```sql
SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk';
SELECT id, name, url, build_status FROM pages WHERE site_id = '<site_id>' AND name = 'index';
```

---

## 1. Precondition — the fix is actually rolled out (shell)

```bash
kubectl -n ai-persona-system get deploy -o wide | grep -E 'page-content-writer|page-build-handler'
# the IMAGE column must show the new chassis tag, not v1.0.1063
```

---

## 2. Capture the baseline (SQL)

The component fingerprint is the reliable change signal. `content_hash` is empty on
this page, so compare `updated_at` and `rendered_html` length — not the hash. Note
too that the page row's `deployed_at` (06-15) has moved without the components
(06-06): trust `page_components.updated_at`, not the page row.

```sql
SELECT position, slot_name, build_status,
       length(rendered_html) AS html_len,
       updated_at, deploy_commit
FROM page_components
WHERE page_id = '6e988cc4-4898-4021-aa5e-2ab0271f9b75'
ORDER BY position;
```
Baseline to beat: 5 rows, all `updated_at` = 2026-06-06 16:59, `html_len`
hero 2426 / tool-list 8951 / guide-list 7513 / game-list 8116 / system-stats 7369
(~34k total).

Recent failures on this site's build path (note the column is `error_message`):
```sql
SELECT occurred_at, agent_type, step_name, action, severity,
       left(error_message, 110) AS error_message
FROM agent_error_log
WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
ORDER BY occurred_at DESC
LIMIT 12;
```

---

## 3. Trigger the rebuild (SQL)

First confirm the **index** work item exists and see its state. Filter by `item_key`
— do not rely on `LIMIT`, because this site has one `page_rerender` item per page
from the same batch and `index` sorts below the first several:
```sql
SELECT id, status, attempt_count||'/'||max_attempts AS attempts, claimed_by,
       created_at, completed_at, item_key, spec->>'page_name' AS page_name
FROM site_work_items
WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
  AND item_type = 'page_rerender'
  AND item_key = 'page_rerender_index_e33263f4-74f8-494f-b191-546845dbbddf'
ORDER BY created_at DESC;
```
If that returns rows, re-open the most recent one. The `item_key` filter targets
index specifically; the `id`-subquery touches exactly one row, so the older
duplicate index rows (all terminal) don't trip the `(site_id, item_key)` dedup
constraint when this one goes non-terminal:
```sql
UPDATE site_work_items
SET status = 'approved',        -- the status your build-dispatch-loop claims
    attempt_count = 0, claimed_by = NULL, claimed_at = NULL,
    completed_at = NULL, error = NULL, result = '{}'::jsonb, updated_at = now()
WHERE id = (
  SELECT id FROM site_work_items
  WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
    AND item_type = 'page_rerender'
    AND item_key = 'page_rerender_index_e33263f4-74f8-494f-b191-546845dbbddf'
  ORDER BY created_at DESC
  LIMIT 1
)
RETURNING id, item_type, status, item_key;
```
Keep the `id` it returns — call it `WORK_ITEM_ID` below.

Two things to confirm, both quick:
- **Claimable status.** The pending indexes are on `status IN ('triaged','approved')`.
  `approved` is the furthest-along claimable state; if your dispatch query claims
  `triaged`, use that. Check: `grep -rn "FROM site_work_items" platform/orchestration/actions/*work_item*`.
- **The dispatch loop is running.** It claims on a cycle; if it isn't active the
  item just sits at `approved`.

If the `SELECT` returned **no** index row, queue it the production way instead and
let discovery raise the item:
```sql
UPDATE pages SET build_status = 'needs_rebuild', updated_at = now()
WHERE id = '6e988cc4-4898-4021-aa5e-2ab0271f9b75';
```

Watch it get claimed (`approved` → `claimed` → terminal):
```sql
SELECT status, claimed_by, attempt_count||'/'||max_attempts AS attempts,
       claimed_at, completed_at, left(error, 140) AS error
FROM site_work_items WHERE id = '<WORK_ITEM_ID>';
```

---

## 4. Watch the logs (shell)

Agent deployments are named `agent-<agent-type>`. If your label key isn't `app`,
adjust the selector, or find the pod with
`kubectl -n ai-persona-system get pods | grep page-content-writer` and pass the pod
name to `logs` instead.

The writer's contract resolution — the heart of fix #1:
```bash
kubectl -n ai-persona-system logs -l app=agent-page-content-writer --since=20m --tail=-1 \
  | grep -E 'resolveResultSpec: resolved result contract'
# expect a line with mode=flatten and matched_key=output_field

kubectl -n ai-persona-system logs -l app=agent-page-content-writer --since=20m --tail=-1 \
  | grep -E 'extractWorkflowResult: result built'
# expect mode=flatten, field_count>1, a size_bytes value well under 900000

kubectl -n ai-persona-system logs -l app=agent-page-content-writer --since=20m --tail=-1 \
  | grep -E 'result_from field not found'
# expect NOTHING — if present, the writer's complete step lost its field name
```

The deprecation census + conflict check (fix #2), across all agents:
```bash
kubectl -n ai-persona-system logs -l app --since=20m --tail=-1 2>/dev/null \
  | grep -E 'deprecated complete-step key in use' \
  | grep -oE 'deprecated_key[":= ]+[a-z_]+' | sort | uniq -c
# the list to feed the rename migration (output_field / output_fields / output)

kubectl -n ai-persona-system logs -l app --since=20m --tail=-1 2>/dev/null \
  | grep -E 'multiple result-contract keys present'
# expect NOTHING
```

The silent-success guard (fix #3) — should stay quiet on this run:
```bash
kubectl -n ai-persona-system logs -l app --since=20m --tail=-1 2>/dev/null \
  | grep -E 'result exceeds delivery cap|undeliverable — notifying parent of FAILURE'
# expect NOTHING for the writer (flatten keeps it well under the cap)

kubectl -n ai-persona-system logs -l app --since=20m --tail=-1 2>/dev/null \
  | grep -E 'Full result exceeded size limit'
# expect NOTHING anywhere — that stub string is removed; a hit means a stale image
```

The handler's save/deploy steps:
```bash
kubectl -n ai-persona-system logs -l app=agent-page-build-handler --since=20m --tail=-1 \
  | grep -E 'save_sections|save_page_sections|content regression|deploy_page'
```

---

## 5. Verify the result, stage by stage (SQL)

Run in order; the first failing stage tells you where it stopped.

**Stage A — the writer delivered a flat result to the parent.** Find the run:
```sql
SELECT orchestration_id, owner_agent_type, status, current_step, created_at
FROM orchestration_states
WHERE owner_agent_type IN ('page-build-handler','page-content-writer')
  AND created_at > now() - interval '30 minutes'
ORDER BY created_at DESC;
```
For the `page-build-handler` orchestration_id, read the flat path:
```sql
SELECT
  jsonb_typeof(collected_data #> '{page_content,response,sections_metadata}')       AS sm_type,
  jsonb_array_length(collected_data #> '{page_content,response,sections_metadata}')  AS sm_count,
  (collected_data #> '{page_content,response,page_html}') IS NOT NULL                AS has_html,
  collected_data #>> '{page_content,response,message}'                              AS stub_message
FROM orchestration_states WHERE orchestration_id = '<PARENT_ORCH_ID>';
```
- `sm_type='array'`, `sm_count=5`, `has_html=true` → **flatten works**.
- `sm_type` null → flatten didn't populate; the writer response didn't carry
  `sections_metadata` (look at the writer's own run — this is the fix's direct target).
- `stub_message` non-null → a stub slipped through (post-fix that means a stale image).

**Stage B — the save was attempted, and passed or was blocked loudly.**
```sql
SELECT occurred_at, step_name, action, severity, left(error_message, 150) AS error_message
FROM agent_error_log
WHERE agent_type = 'page-build-handler'
  AND occurred_at > now() - interval '30 minutes'
ORDER BY occurred_at DESC;
```
- No `save_sections` error + components changed (Stage C) → **save succeeded**.
- `content regression blocked: new content has N chars vs M existing` → the save was
  reached (so Stage A is fine), but new content is thinner than what's live — the
  06-14 mode. This is a loud, correct block, not a silent no-op. The next thread is
  **why the writer produced thin content** (the writer's research/content sub-agents
  — research `sources` mapping, content-reviewer — not this fix). Don't disable the
  guard to force it through.

**Stage C — components actually changed.** Re-run the §2 fingerprint and compare:
```sql
SELECT position, slot_name, length(rendered_html) AS html_len, updated_at, deploy_commit
FROM page_components
WHERE page_id = '6e988cc4-4898-4021-aa5e-2ab0271f9b75'
ORDER BY position;
```
- `updated_at` advanced past 2026-06-06 16:59 (use this, the hash is empty) and
  `html_len` moved → content was rewritten and saved. Unchanged → nothing saved
  (read Stage B for the reason).

**Stage D — the work item completed on a real save, not a false complete.**
```sql
SELECT status, completed_at, result, left(error, 150) AS error
FROM site_work_items WHERE id = '<WORK_ITEM_ID>';
```
- The distinguishing test from the old bug: `status='complete'` is only meaningful
  **with** Stage C showing changed components. `complete` + unchanged components is
  the old false-complete and should no longer occur.
- `status='failed'`/`needs_human_review` with an error → it failed loudly
  (acceptable; read the error and Stage B).

**Stage E — deploy (optional).**
```sql
SELECT build_status, last_built_at, deployed_at, version
FROM pages WHERE id = '6e988cc4-4898-4021-aa5e-2ab0271f9b75';
```
- `deploy_page` timeouts in Stage B's list → the build saved but the deploy step
  timed out (06-06/08 mode). Check the deploy agent and the GitHub→Backblaze path.

---

## 6. Verify the silent-success hardening (SQL)

Fix #3 is a safety net that should rarely fire now. Verify passively:
```sql
-- no stub rows anywhere recent
SELECT count(*) AS stub_rows
FROM orchestration_states
WHERE updated_at > now() - interval '1 day'
  AND collected_data::text LIKE '%Full result exceeded size limit%';
-- expect 0

-- if any oversize DID occur, it logged a fatal error naming the largest field
SELECT occurred_at, agent_type, severity, error_code, left(error_message, 160) AS error_message
FROM agent_error_log
WHERE occurred_at > now() - interval '1 day'
  AND (error_code = 'CHILD_ORCHESTRATION_FAILED' OR error_message ILIKE '%delivery cap%')
ORDER BY occurred_at DESC;
-- 0 rows on a healthy run; any row is an agent that needs a result contract
```

Active test (dev/staging only — not prod): the cap is a compile-time const
(`MaxResultSizeBytes`), so the clean way to exercise the failure path is to lower it
in a dev build, run any agent, and confirm `result exceeds delivery cap` →
`notifyParentOfFailure` → an `error_unrecoverable` response and a
`CHILD_ORCHESTRATION_FAILED` `agent_error_log` row (`severity='fatal'`), with the
page **not** marked complete. Restore the const afterwards.

---

## 7. If the page still doesn't update — triage

| Symptom (from the stages) | Stage | Likely cause | Next check |
|---|---|---|---|
| Writer log shows `mode=fallback`, not flatten | A | Writer `complete` step lost its `output_field` | live `agent_definitions` complete config; `023_page_content_writer_agent.sql` |
| `sm_type` null on the parent | A | Writer response didn't carry `sections_metadata` | the writer orchestration's own `collected_data` |
| `content regression blocked` | B | New content thinner than live | writer sub-agent data flow (research `sources`, content-reviewer); not this fix |
| Other `save_sections` error | B | Save action failure | full `error_message`; `save_page_sections_action.go` |
| `complete` work item but components unchanged | C/D | Would be the old false-complete | confirm the running image is the new tag, not `v1.0.1063` |
| `deploy_page` timed out | E | Deploy step / downstream agent | deploy agent logs; GitHub Actions → Backblaze |
| Work item stuck at `approved` | 3 | Dispatch loop not claiming, or wrong status | confirm the loop is running and the claim status matches §3 |
| `Full result exceeded size limit` appears | any | Stale chassis image | re-roll the image; confirm tag |

0-row results aren't proof of absence until the query is confirmed — the child-side
path has no `.response` wrapper (that exists only on the parent), so querying the
parent path against a child row returns 0 rows as an artefact.

---

## 8. Rollback (SQL + shell)

The deploy is image-tag based:
```sql
SELECT type, image_tag FROM agent_definitions
WHERE type IN ('page-content-writer','page-build-handler');   -- confirm the agent-type column name
```
Set `image_tag` back to `v1.0.1063` for those agents and re-roll, or redeploy the
previous chassis image under the new tag. Reverting restores the prior behaviour
(the silent stub returns), so only roll back if the new build is actively worse than
the known no-op.

---

## What "good" looks like for this run

Writer log shows `mode=flatten matched_key=output_field`; the parent's
`page_content.response.sections_metadata` is an array of 5; all five
`page_components.updated_at` advance past 2026-06-06 with changed `html_len`; the
index `page_rerender` item is `complete` **and** components changed; no `Full result
exceeded size limit` anywhere. If it stops earlier, the stage that failed names the
next thread — and unlike before, it stops with a reason in the logs rather than a
silent success.
