# Runbook — gamesdesign.co.uk `index`: build/re-render pipeline fixes

## Context — what this runbook is for

This platform builds and maintains multi-page websites with a system of autonomous
**agents** — rows in the `agent_definitions` table that run as Kubernetes pods in the
`ai-persona-system` namespace, talk to each other over Kafka, and persist run state in
the `clients_db` Postgres database. Every unit of work (build a page, swap a hero
image, fix a broken link) is a row in `site_work_items`, claimed off a dispatch loop
and routed to a handler agent. Deployment is image-tag based: code ships via
GitHub → Backblaze into a new chassis image, then each agent's `image_tag` is bumped to
adopt it; workflow (`default_config`) changes are DB-only and take effect immediately.

This runbook tracks a connected series of fixes to the page build/re-render pipeline,
all of the same shape — **work that reports success but doesn't actually happen, or
work dropped/duplicated by key collisions.** It started from one symptom:
gamesdesign.co.uk's `index` page reported successful rebuilds for days while the live
page never changed. The investigation surfaced three structural faults, fixed in order
and documented as Parts 1–3 below:

- **Part 1 (shipped 2026-06-18)** — the chassis coordinator silently discarded a child
  agent's workflow result (when the child declared a singular `output_field`, or the
  result exceeded a size cap), substituting a stub that still reported
  `status: completed`; the resulting no-op save then rolled the work item up to
  `complete`. Fixed with a result-contract resolver and a loud failure on oversize.
- **Part 2 (deploying)** — an image landing, or section data that became resolvable, was
  being routed through a full LLM content rebuild (`needs_page` → page-build-handler →
  writer) when it only needed a field re-resolve and a section re-render. Added a no-LLM
  `rerender_page_sections` path and repointed those triggers to a `page_rerender` item.
- **Part 3 / Part B (next)** — work-item dedup keys (`item_key`) had drifted from their
  `item_type` across the various creators, so content rebuilds didn't dedup with each
  other and an adoption collision could silently drop work. Canonicalize the keys behind
  a shared builder.

Claude has no cluster or DB access — every SQL / `kubectl` / `git` step is run by the
operator and the output pasted back. Each part is self-contained: apply, deploy, verify,
roll back.

---

Purpose (Part 1): trigger one `index` rebuild and confirm the three changes deployed
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
- content rebuild = a `needs_page` item → `page-build-handler` (calls the writer; the
  path our fix touches). `page_rerender` → `page-rerender` and `needs_rerender` →
  `rerender-pages` are assemble-from-DB + deploy only — they cannot exercise the fix
  and will re-deploy stale content. See §3.

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

Use a **content** rebuild, which runs `page-build-handler` (the path that calls
`page-content-writer` and reads `page_content.response.sections_metadata` — the one
our fix touches). That path is driven by a **`needs_page`** item, handled by
`page-build-handler`. Do **not** use a `page_rerender` item: that routes to the
`page-rerender` agent, which only re-assembles existing `page_components` from the DB
and commits to git — no writer, no content generation — so it cannot exercise the fix
and will "succeed" while re-deploying stale content. (`needs_rerender` → `rerender-pages`
is likewise assemble-only.)

First look for an existing `needs_page` item for index to re-open:
```sql
SELECT id, status, source, attempt_count||'/'||max_attempts AS attempts, claimed_by,
       created_at, completed_at, item_key, spec->>'page_name' AS page_name
FROM site_work_items
WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
  AND item_type = 'needs_page'
  AND spec->>'page_name' = 'index'
ORDER BY created_at DESC;
```
If one exists, re-open the most recent (claimable status is **`triaged`** — confirmed
on this cluster; the `id`-subquery touches one row so terminal duplicates don't trip
the `(site_id, item_key)` dedup constraint):
```sql
UPDATE site_work_items
SET status = 'triaged',
    attempt_count = 0, claimed_by = NULL, claimed_at = NULL,
    completed_at = NULL, error = NULL, result = '{}'::jsonb, updated_at = now()
WHERE id = (
  SELECT id FROM site_work_items
  WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
    AND item_type = 'needs_page' AND spec->>'page_name' = 'index'
  ORDER BY created_at DESC LIMIT 1
)
RETURNING id, item_type, status, spec->>'page_name' AS page_name;
```
If none exists, raise one (manual-rebuild). `status='triaged'` so the loop claims it;
`item_key` is unique-per-active by the dedup index, so vary it per attempt:
```sql
INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary,
   spec, priority, status, created_by, handler_agent, item_key)
VALUES
  ('e33263f4-74f8-494f-b191-546845dbbddf', 'manual-rebuild', 'build', 'needs_page',
   'medium', 'Manual content rebuild of index (verify result-contract fix)',
   '{"page_name":"index","page_id":"6e988cc4-4898-4021-aa5e-2ab0271f9b75"}'::jsonb,
   50, 'triaged', 'runbook', 'page-build-handler',
   'needs_page_index_manual_20260619')
RETURNING id, item_type, status, item_key;
```
Keep the returned `id` — call it `WORK_ITEM_ID`.

Confirm the dispatch loop is running (it claims `triaged` items on a cycle; if idle,
the item just sits at `triaged`).

Alternative (let discovery raise the item the production way):
```sql
UPDATE pages SET build_status = 'needs_rebuild', updated_at = now()
WHERE id = '6e988cc4-4898-4021-aa5e-2ab0271f9b75';
```

Watch it get claimed (`triaged` → `claimed` → terminal):
```sql
SELECT status, claimed_by, attempt_count||'/'||max_attempts AS attempts,
       claimed_at, completed_at, left(error, 140) AS error
FROM site_work_items WHERE id = '<WORK_ITEM_ID>';
```
The `page-build-handler` orchestration this raises is the one §5 Stage A inspects.

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
| Work item stuck at `triaged` | 3 | Dispatch loop not claiming | confirm the loop is running |
| `complete` in ~40s, page re-deployed unchanged, `called_writer=f` | 3 | Ran `page_rerender`/`needs_rerender` (assemble-only), not a `needs_page` content rebuild | re-trigger via §3 `needs_page` |
| `Full result exceeded size limit` appears | any | Stale chassis image | re-roll the image; confirm tag |

0-row results aren't proof of absence until the query is confirmed — the child-side
path has no `.response` wrapper (that exists only on the parent), so querying the
parent path against a child row returns 0 rows as an artefact.

---

## 8. Rollback (SQL + shell)

The deploy is image-tag based:
```sql
SELECT type, image_tag FROM agent_definitions
WHERE type IN ('page-content-writer','page-build-handler');
```
Set `image_tag` back to `v1.0.1063` for those agents and re-roll, or redeploy the
previous chassis image under the new tag. Reverting restores the prior behaviour
(the silent stub returns), so only roll back if the new build is actively worse than
the known no-op.

---

## What "good" looks like for this run

Writer log shows `mode=flatten matched_key=output_field`; the parent
(`page-build-handler`) has `page_content.response.sections_metadata` as an array of 5;
all five `page_components.updated_at` advance past 2026-06-06 with changed `html_len`;
the `needs_page` item is `complete` **and** components changed; no `Full result
exceeded size limit` anywhere. If it stops earlier, the stage that failed names the
next thread — and unlike before, it stops with a reason in the logs rather than a
silent success.

---
---

# Part 2 — re-render path (Option Y): apply, deploy, verify

Purpose: ship and verify `rerender_page_sections` and the `page_rerender` routing,
so an image landing or now-resolvable section data re-renders a page **without** the
LLM writer, instead of going through `needs_page` → page-build-handler.

What changed (this is the change under test):
- NEW action `rerender_page_sections` (re-render all sections from stored
  `content_data` + fresh resolved fields; escalate NULL-`content_data` pages to a
  full `needs_page` rebuild).
- page-rerender `default_config` gains a gated pre-pass (`check_rerender_mode` →
  `rerender_sections` → `check_escalated` → `save_sections` → existing `render_page`).
- `flag_page_image_rebuild` and `reconcile_section_data` emit `page_rerender`
  (was `needs_page`), handler `page-rerender` (was `page-build-handler`).

Same IDs as Part 1 (site `e33263f4-74f8-494f-b191-546845dbbddf`, index page
`6e988cc4-4898-4021-aa5e-2ab0271f9b75`). DB steps = bare SQL at `clients_db=#`;
log/deploy steps = shell with `kubectl`.

## P2.0 The moves, in order (sequencing matters)

The workflow change references an action that exists only in the new image, so the
image must be live on page-rerender **before** the `default_config` update, and the
repointed creators (`flag`/`reconcile`) need the new image too. Do it in one window:

1. **Code:** add `rerender_page_sections_action.go`; register it in `registry.go`
   (`Handler: RerenderPageSectionsAction, Category: "site", IsLocal: true`); change
   `itemType`→`page_rerender` and `handlerAgent`→`page-rerender` in both
   `flag_page_image_rebuild_action.go` and `reconcile_section_data_action.go`.
2. **Build:** `go build ./...` (toolchain isn't in this environment — this is yours).
3. **Ship:** commit → GitHub Actions → Backblaze → new chassis image; note the tag.
4. **Bump `image_tag`** to the new tag on every def that runs the changed code —
   `page-rerender` (runs the new action), and the hosts that emit the repointed items
   (`image-build-handler` for `flag`, and whatever agent runs `reconcile_section_data`):
   ```sql
   SELECT type, image_tag FROM agent_definitions
   WHERE type IN ('page-rerender','image-build-handler');
   -- set image_tag to the new tag for each, then re-roll the deployments
   ```
5. **Apply the page-rerender workflow** (after the new image is live on page-rerender):
   ```sql
   UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(default_config, '{workflow,start_step}', '"check_rerender_mode"'),
         '{workflow,steps}',
         (default_config #> '{workflow,steps}') || '{
           "check_rerender_mode": {
             "action": "conditional",
             "config": {"condition": "input_data.spec.reason == ''image_landed'' OR input_data.spec.reason == ''section_data_resolved''", "then_step": "rerender_sections", "else_step": "render_page"},
             "description": "Re-render sections only for re-render reasons; else assemble stored HTML"
           },
           "rerender_sections": {
             "action": "rerender_page_sections",
             "config": {"target_site_id": "input_data.site_id", "page_name": "input_data.spec.page_name", "reason": "input_data.spec.reason"},
             "output_field": "rerender_sections", "next_step": "check_escalated",
             "description": "Re-render all sections from stored content_data + fresh resolved fields (no LLM)"
           },
           "check_escalated": {
             "action": "conditional",
             "config": {"condition": "rerender_sections.escalated == true", "then_step": "complete", "else_step": "save_sections"},
             "description": "If a section had no content_data, the page was escalated to the writer — stop"
           },
           "save_sections": {
             "action": "save_page_sections",
             "config": {"site_id_field": "rerender_sections.site_id", "page_name_field": "input_data.spec.page_name", "sections_metadata_field": "rerender_sections.sections_metadata"},
             "output_field": "sections_saved", "next_step": "render_page",
             "description": "Persist re-rendered sections, then assemble + deploy via the existing path"
           }
         }'::jsonb
       ),
       updated_at = now()
   WHERE type = 'page-rerender';
   ```
   (The `''` are escaped single quotes inside the SQL literal.) Confirm it took:
   ```sql
   SELECT default_config #>> '{workflow,start_step}' AS start_step,
          (default_config #> '{workflow,steps}') ? 'rerender_sections' AS has_rerender_step
   FROM agent_definitions WHERE type = 'page-rerender';
   -- expect: check_rerender_mode | t
   ```

## P2.1 Pre-flight — the load-bearing wiring fact

`rerender_sections` resolves `target_site_id` only from the explicit config path
`input_data.site_id` (the renamed input has no recursive fallback). So a claimed
`page_rerender` item must surface its `site_id` column at `input_data.site_id`. The
first test run proves it: if `rerender_sections` errors `missing required fields:
[target_site_id]`, the dispatch isn't putting `site_id` there — inspect a claimed
item's `collected_data` and adjust the config path to wherever site_id lands:
```sql
SELECT collected_data #>> '{input_data,site_id}' AS input_site_id
FROM orchestration_states
WHERE owner_agent_type = 'page-rerender'
  AND created_at > now() - interval '15 minutes'
ORDER BY created_at DESC LIMIT 1;
-- non-null → the wiring path is correct
```

## P2.2 Baseline (SQL)

Capture copy + resolved field + fingerprint for the test page (index):
```sql
SELECT position, slot_name,
       content_data->>'headline'    AS headline,      -- COPY: must NOT change on a re-render
       content_data->>'hero_url'    AS hero_url,       -- RESOLVED: may change (image test)
       length(rendered_html)        AS html_len,
       updated_at
FROM page_components
WHERE page_id = '6e988cc4-4898-4021-aa5e-2ab0271f9b75'
ORDER BY position;
```

## P2.3 Test 1 — happy path, direct insert (proves the no-LLM re-render runs)

Insert a `page_rerender` item for index. `site_id` column set so `input_data.site_id`
surfaces; status `triaged` so the dispatch loop claims it; vary `item_key` per attempt:
```sql
INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary,
   spec, priority, status, created_by, handler_agent, item_key)
VALUES
  ('e33263f4-74f8-494f-b191-546845dbbddf', 'runbook', 'build', 'page_rerender',
   'medium', 'Test re-render of index (image_landed path)',
   '{"reason":"image_landed","page_name":"index"}'::jsonb,
   50, 'triaged', 'runbook', 'page-rerender', 'page_rerender:index_test1')
RETURNING id, item_type, status, item_key;
```
Watch it claim → complete (`triaged` → `claimed` → `complete`):
```sql
SELECT status, claimed_by, completed_at, left(error,140) AS error,
       result->>'escalated' AS escalated, result->>'rerendered' AS rerendered,
       result->>'carried' AS carried
FROM site_work_items WHERE id = '<WORK_ITEM_ID>';
```
Logs (deployment `agent-page-rerender`):
```bash
kubectl -n ai-persona-system logs -l app=agent-page-rerender --since=15m --tail=-1 \
  | grep -E 'rerender_page_sections: done'           # expect rerendered=5 carried=0 (index has 5 sections)
kubectl -n ai-persona-system logs -l app=agent-page-rerender --since=15m --tail=-1 \
  | grep -E 'execute_llm_prompt|spawn.*page-content-writer'   # expect NOTHING — no LLM on this path
kubectl -n ai-persona-system logs -l app=agent-page-rerender --since=15m --tail=-1 \
  | grep -E 'CONTENT REGRESSION BLOCKED'             # expect NOTHING — copy is preserved
```
Verify (SQL) — copy unchanged, sections re-rendered, no full-rebuild spawned:
```sql
-- headline identical to baseline; updated_at advanced past the baseline timestamp
SELECT position, slot_name, content_data->>'headline' AS headline,
       length(rendered_html) AS html_len, updated_at
FROM page_components WHERE page_id = '6e988cc4-4898-4021-aa5e-2ab0271f9b75' ORDER BY position;

-- NO page-content-writer / page-build-handler orchestration for this run
SELECT owner_agent_type, count(*) FROM orchestration_states
WHERE created_at > now() - interval '15 minutes'
  AND owner_agent_type IN ('page-content-writer','page-build-handler')
GROUP BY owner_agent_type;
-- expect 0 rows (the re-render path neither rebuilds content nor calls the writer)
```
Pass: item `complete`, `escalated=false`, `rerendered=5`; headline unchanged;
`updated_at` advanced; no writer orchestration; no regression block; no LLM log.
(With no new asset, `html_len`/`hero_url` may be unchanged — Test 1 proves the path
runs and preserves copy; the actual asset swap is Test 2.)

**RESULT (2026-06-22): PASS.** Item `8717674e…` completed with no error. All five
sections' `updated_at` advanced 06-19 10:34 → 06-22 10:24 (so `save_sections` ran — the
re-render path, not assemble-only). Hero headline identical; only the hero's
`rendered_html` changed (+76 chars, consistent with the CTA now resolving), the other
four byte-identical (resolved data unchanged). This also retires P2.1 — `rerender_sections`
ran to completion, so `target_site_id` resolved from `input_data.site_id`; the wiring is
correct. P2.4–P2.7 remain.

## P2.4 Test 2 — real image-landed flow (proves the asset is picked up)

The production path: a page hero image is generated via `needs_imagery`, and when
`image-build-handler` stores + deploys the asset it calls `flag_page_image_rebuild`,
which now emits a `page_rerender`. Pick a page whose hero is still the build-time
fallback and trigger its imagery (or wait for a real one), then confirm the chain:
```sql
-- the flag emitted a page_rerender (NOT needs_page) routed to page-rerender
SELECT id, item_type, handler_agent, status, item_key,
       spec->>'reason' AS reason, spec->>'page_name' AS page_name, created_at
FROM site_work_items
WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
  AND source = 'image-build-handler'
  AND created_at > now() - interval '30 minutes'
ORDER BY created_at DESC;
-- expect item_type='page_rerender', handler_agent='page-rerender', reason='image_landed'
```
Then verify the hero's `rendered_html` (and `hero_url` in `content_data`) now points
at the generated asset, not `/assets/images/hero.jpg`:
```sql
SELECT slot_name, content_data->>'hero_url' AS hero_url,
       left(rendered_html, 220) AS hero_html_head
FROM page_components
WHERE page_id = '<the affected page_id>' AND slot_name = 'hero';
```
Pass: the `page_rerender` item completes, the hero now references the new asset URL,
and no page-content-writer orchestration ran for it.

## P2.5 Test 3 — section_data_resolved

For a page with an open `needs_section_data` whose query is now resolvable, running
`reconcile_section_data` (its host agent) emits a `page_rerender`
(`reason=section_data_resolved`). Verify the list section (`tool-list`/`guide-list`/
`game-list`) re-renders with the resolved items and the `needs_section_data` item
closes — again with no writer call.
```sql
SELECT id, item_type, handler_agent, spec->>'reason' AS reason, spec->>'page_name' AS page
FROM site_work_items
WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
  AND source = 'section-data-reconciler' AND created_at > now() - interval '30 minutes'
ORDER BY created_at DESC;
-- expect item_type='page_rerender', reason='section_data_resolved'
```

## P2.6 Test 4 — NULL content_data escalation (the self-heal path)

Find a page with a section whose `content_data` is NULL (a legacy page predating
structured-content capture):
```sql
SELECT p.name, count(*) FILTER (WHERE pc.content_data IS NULL) AS null_sections,
       count(*) AS total_sections
FROM pages p JOIN page_components pc ON pc.page_id = p.id
WHERE p.site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
GROUP BY p.name HAVING count(*) FILTER (WHERE pc.content_data IS NULL) > 0;
```
If one exists, raise a `page_rerender` for it (as in P2.3 but with that page_name).
Expect the action to **escalate**: it emits a `needs_page` rebuild and returns
`escalated=true` without re-rendering:
```sql
-- the re-render returned escalated
SELECT result->>'escalated' AS escalated, result->>'escalate_reason' AS reason
FROM site_work_items WHERE id = '<the page_rerender id>';   -- expect true / null_content_data

-- and a needs_page:<page> rebuild appeared (handled by page-build-handler, regenerates + backfills)
SELECT item_type, handler_agent, item_key, source, status
FROM site_work_items
WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
  AND item_key = 'needs_page:<page>' AND created_at > now() - interval '15 minutes';
```
After the `needs_page` rebuild completes, that page's `content_data` is populated, so
a subsequent `page_rerender` takes the light path. (If no NULL-`content_data` page
exists on this site, this test is N/A — note it and move on.)

## P2.7 Test 5 — backward-compat (plain page_rerender, no reason)

Existing `page_rerender` consumers (no `reason`, or a different one) must still
assemble-only. Insert one with no `reason`:
```sql
INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, spec, priority, status, created_by, handler_agent, item_key)
VALUES
  ('e33263f4-74f8-494f-b191-546845dbbddf', 'runbook', 'build', 'page_rerender', 'low',
   'Test plain re-render (assemble-only)', '{"page_name":"index"}'::jsonb,
   60, 'triaged', 'runbook', 'page-rerender', 'page_rerender:index_test5')
RETURNING id;
```
Verify `rerender_sections` did NOT run (no `rerender_page_sections: done` line for it)
and the page assembled via `render_page` as before — `check_rerender_mode` took
`else_step`. Confirms the gate doesn't disturb existing traffic.

## P2.8 Triage (symptom → cause → check)

| Symptom | Likely cause | Check |
|---|---|---|
| `rerender_sections` errors `missing required fields: [target_site_id]` | dispatch isn't surfacing `site_id` at `input_data.site_id` | P2.1 query; adjust the config path to where site_id lands |
| `page_rerender` item routed to page-build-handler | repoint not deployed (handler still `page-build-handler`) | confirm new `image_tag` live; confirm the two action files changed |
| Copy CHANGED after a re-render | the LLM path ran (wrong route) | confirm `reason` matched the gate and no page-content-writer orchestration; check `check_rerender_mode` condition |
| `page_rerender` keeps escalating to `needs_page` | `content_data` still NULL after the writer rebuild | the writer rebuild must backfill `content_data` — inspect that page-build-handler run (Part 1 §5) |
| `rerender_page_sections: done carried=N` (N>0) when copy exists | components couldn't be re-rendered (missing template / deferred) | check `carrying stored` log lines and the section's component function |
| Nothing happens | dispatch loop idle / item stuck `triaged` | confirm the loop is running |
| Two builds of one page | a stale `needs_page:<page>` and `page_rerender:<page>` both open (different work, OK) — but two `page_rerender` for one page is the Part B dedup gap | Part B (key canonicalization) — separate change |

## P2.9 Rollback (SQL + shell)

Workflow first (DB), then code (image). To revert the workflow:
```sql
UPDATE agent_definitions
SET default_config = jsonb_set(
      (default_config #- '{workflow,steps,check_rerender_mode}'
                       #- '{workflow,steps,rerender_sections}'
                       #- '{workflow,steps,check_escalated}'
                       #- '{workflow,steps,save_sections}'),
      '{workflow,start_step}', '"render_page"'),
    updated_at = now()
WHERE type = 'page-rerender';
```
Revert the two creators (`itemType`→`needs_page`, `handlerAgent`→`page-build-handler`)
and set `image_tag` back. Reverting the creators restores the prior behaviour: an
image landing again triggers a full writer rebuild (correct, just heavier). The new
action being present-but-unreferenced is harmless, so the action file need not be
removed to roll back.

## What "good" looks like for Part 2

`flag`/`reconcile` emit `page_rerender` (not `needs_page`); page-rerender's
`check_rerender_mode` routes those to `rerender_sections`; the page re-renders with
its copy unchanged and the new asset/section-data picked up; `page_components.updated_at`
advances; **no** page-content-writer orchestration and **no** `execute_llm_prompt` for
the run; **no** content-regression block. A legacy page with NULL `content_data`
escalates once to a `needs_page` rebuild and is light thereafter. Plain `page_rerender`
items (no reason) still assemble-only.

---
---

# Part 3 — item_key canonicalization (Part B): survey, decide, apply, verify

Status: **scoped, not built.** Separate commit, do **after** Part 2 is live. Two
decisions (P3.1) gate the build. The survey in P3.2 is runnable now — it confirms the
bugs are live and sizes the blast radius before any code is written.

Purpose: make `item_key` obey the contract `{item_type}:{target}` — one open row per
unit per site, prefix == item_type — so the `(site_id, item_key)` dedup index stops
(a) failing to collapse duplicates and (b) silently dropping work on collisions.

The two confirmed bugs:
- **Bug 1 — duplicate content builds.** `reconcile_site_plan` keyed `needs_page:<page>`
  while `flag`/`reconcile_section_data` keyed `page_rerender:<page>` despite being
  `needs_page`-type → same work, different keys, no dedup. **Part 2 already closes
  this** (those two now emit `page_rerender`-*type* items keyed `page_rerender:<page>`,
  so prefix == type). Part B only *verifies* it (P3.4 B4).
- **Bug 2 — collision drops work.** `apply_adoption_plan` keys BOTH
  `needs_content_page` AND `needs_tool_recreation` as `needs_page:<name>`. Different
  handlers; when names collide the unique index keeps one and drops the other to the
  wrong place. Part B fixes this.

## P3.1 Decisions (resolved 2026-06-21)

1. **`needs_page` vs `needs_content_page`** — two types, same `page-build-handler`, same
   effective work. Evidence from `apply_adoption_plan_action.go:653` + the doc-029 Phase-0
   comment: the adoption content item is keyed `needs_page:<name>` *on purpose*, so it
   co-dedups with a planner `needs_page` build of the same page.
   **DECIDED: Option B — keep content in the `needs_page` namespace**
   (`workItemKey("needs_page", page.Name)`), preserving that co-dedup. The "prefix == type"
   invariant carries one documented exception: `needs_content_page` shares the `needs_page`
   dedup namespace. (Rejected Option A — `needs_content_page:<name>` — which would have forced
   the planner + `reconcile_site_plan` to also emit `needs_content_page` to keep the co-dedup,
   a larger change touching more creators.)
2. **adoption-recreate vs reconcile-build of the same page** — **DECIDED: they co-dedup**, as
   a direct consequence of (1): both key the page as `needs_page:<page>`, so one open build
   serves both triggers (page-build-handler rebuilds from current state regardless of which
   fired). Known consequence: adoption sets `spec.mode="recreate"` while a planner build may
   not, so whichever row wins the `ON CONFLICT DO NOTHING` fixes the mode — acceptable (both
   produce the page); revisit only if recreate-mode and plan-build must diverge.

Both decisions are settled; the apply steps below reflect them.

## P3.2 Pre-flight survey + baseline (runnable now, before any change)

Confirm the dedup index predicate (which statuses count as "active"):
```sql
SELECT indexname, indexdef FROM pg_indexes
WHERE tablename = 'site_work_items' AND indexdef ILIKE '%item_key%';
```

Find every place key-prefix disagrees with item_type (the drift; Bug 2 shows here):
```sql
SELECT item_type,
       split_part(item_key, ':', 1) AS key_prefix,
       count(*) AS rows
FROM site_work_items
GROUP BY item_type, split_part(item_key, ':', 1)
ORDER BY (item_type <> split_part(item_key, ':', 1)) DESC, rows DESC;
-- rows where item_type <> key_prefix are the canonicalization targets.
-- expect needs_content_page and needs_tool_recreation showing key_prefix='needs_page'
```

The adoption pair specifically:
```sql
SELECT item_type, split_part(item_key, ':', 1) AS key_prefix, count(*)
FROM site_work_items
WHERE item_type IN ('needs_content_page','needs_tool_recreation')
GROUP BY 1, 2 ORDER BY 1, 2;
-- both at key_prefix='needs_page' → Bug 2 confirmed live
```

Keys that have served more than one type (the collision fingerprint — where a drop
could have happened):
```sql
SELECT site_id, item_key, array_agg(DISTINCT item_type) AS types, count(*) AS rows
FROM site_work_items
GROUP BY site_id, item_key
HAVING count(DISTINCT item_type) > 1
ORDER BY rows DESC;
-- e.g. needs_page:<name> mapping to {needs_content_page, needs_tool_recreation}
```

Keep these counts as the baseline — after deploy, new items must stop adding to the
mismatch rows.

## P3.3 The moves (gated on P3.1 + the code being written)

1. **Code — key-builder.** In `work_items_common.go` add a canonical builder used by
   both `insertWorkItem(workItem{…})` and the inline `INSERT … VALUES` creators, e.g.
   `func WorkItemKey(itemType, target string) string` returning `itemType + ":" + target`.
   Route all key construction through it so prefix == type by construction.
2. **Code — repoint adoption** (`apply_adoption_plan_action.go`). The work-item INSERT
   (lines 642–656) currently hardcodes `fmt.Sprintf("needs_page:%s", page.Name)` for both
   branches. Add `var itemKey string` to the declarations (620–621), set it in each branch of
   the `itemType` if-else (623–638), and replace the hardcoded key at line 655 with `itemKey`:
   - tool branch → `workItemKey("needs_tool_recreation", page.Name)` (the Bug 2 fix — its own
     namespace; different work on a different handler).
   - content branch → `workItemKey("needs_page", page.Name)` (Option B — keep the deliberate
     doc-029 co-dedup with planner `needs_page` builds).
   Comment both branches stating which namespace each uses and why.
3. `go build ./...` (yours — toolchain isn't in this environment).
4. Ship → new chassis image; bump `image_tag` on the agent that runs
   `apply_adoption_plan` (the adoption / plan-adoption handler) and any agent that runs
   the inline creators the helper now touches.
5. **Docs:** flip the caveat in `002_system_architecture.md` and
   `003_contracts_and_standards.md` from "route by `item_type`, never filter by `item_key`"
   to: "`item_key` prefix == `item_type`, with one documented exception — `needs_content_page`
   is keyed in the `needs_page` namespace so adoption-recreate co-dedups with planner builds —
   so filtering by key prefix is safe provided that pair is treated as a single namespace."
   Only **after** deploy — until then the old caveat is still true for historical rows.

No `default_config` / workflow change for Part B — this is action-code + docs only.

## P3.4 Tests

**B1 — Bug 2 fixed (the headline test).** Drive an adoption plan that produces both a
content page and a tool-recreation for the *same* name (or exercise the real adoption
flow that historically collided). After it runs, both rows must exist, distinct keys,
correct handlers:
```sql
SELECT item_type, handler_agent, item_key, status, created_at
FROM site_work_items
WHERE created_at > now() - interval '30 minutes'
  AND item_type IN ('needs_content_page','needs_tool_recreation')
ORDER BY created_at DESC;
-- expect TWO rows: needs_content_page:<name> (→ page-build-handler) AND
-- needs_tool_recreation:<name> (→ its tool handler). Pre-fix only one survived.
```

**B2 — invariant holds for new items.** Re-run the P3.2 prefix-mismatch survey, scoped
to post-deploy rows:
```sql
SELECT item_type, split_part(item_key, ':', 1) AS key_prefix, count(*)
FROM site_work_items
WHERE created_at > '<deploy timestamp>'
  AND item_type <> split_part(item_key, ':', 1)
GROUP BY 1, 2;
-- expect 0 rows
```

**B3 — legitimate dedup still works (regression).** The fix must not break collapsing
of genuine duplicates. Insert the same item twice and confirm one active row survives:
```sql
INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, spec, priority, status, created_by, handler_agent, item_key)
VALUES
  ('e33263f4-74f8-494f-b191-546845dbbddf','runbook','build','needs_content_page','low',
   'dedup test','{"page_name":"dedup_probe"}'::jsonb,90,'triaged','runbook','page-build-handler',
   'needs_content_page:dedup_probe')
ON CONFLICT DO NOTHING;
-- run the identical INSERT a second time, then:
SELECT count(*) FROM site_work_items
WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf'
  AND item_key='needs_content_page:dedup_probe'
  AND status NOT IN ('complete','failed','cancelled');   -- adjust to the index predicate from P3.2
-- expect 1. Delete the probe row afterwards.
```

**B4 — Part A's Bug 1 closure (cross-check).** Confirm `flag`/`reconcile_section_data`
items already carry the canonical key:
```sql
SELECT source, item_type, split_part(item_key, ':', 1) AS key_prefix, count(*)
FROM site_work_items
WHERE source IN ('image-build-handler','section-data-reconciler')
  AND created_at > now() - interval '1 day'
GROUP BY 1, 2, 3;
-- expect item_type='page_rerender' AND key_prefix='page_rerender' (prefix == type)
```

## P3.5 Triage

| Symptom | Likely cause | Check |
|---|---|---|
| Adoption still produces one row for a colliding name | repoint not deployed | confirm new `image_tag`; confirm `apply_adoption_plan` ~line 655 changed |
| New items still show prefix ≠ type | a creator bypasses the helper with an inline key | grep for inline `item_key` string-building not routed through `WorkItemKey` |
| Genuine duplicates now both persist | the helper changed the key so the index no longer matches | compare the produced key to the index predicate (P3.2) |
| A page builds twice (`page_rerender` + `needs_page`) | different units of work (a re-render and a full build) — both can be open legitimately | not a key bug; confirm they're genuinely different operations, not two of the same |

## P3.6 Rollback

Revert the `apply_adoption_plan` repoint (both types back to `needs_page:<name>`) and
the helper, set `image_tag` back, and revert the doc caveat. Reverting reintroduces
Bug 2 (the silent drop), so only roll back if the new build is worse. The helper being
present-but-unused is harmless, so it needn't be removed to roll back the repoint.

## What "good" looks like for Part B

Every creator routes keys through the builder; the P3.2 prefix-mismatch survey returns
zero for post-deploy rows; an adoption plan that names a content page and a tool the
same now yields two correctly-routed rows instead of one; genuine duplicates still
collapse to one open row; and the 002/003 caveat reads "prefix == item_type (with the
documented `needs_content_page` ⊂ `needs_page` namespace exception) — safe to filter."
Bug 1 stays closed by Part A (B4).

---
---

# Part 4 — interactive page rebuilt as content (game/tool dropped): cause confirmed

Status: **cause confirmed (2026-06-22), fix pending.** A different failure from a silent
no-op: the page rebuilt successfully but as the *wrong kind* of page — a working
interactive game was overwritten with a plain content layout. A `page_rerender` cannot
fix it (it re-renders the sections that exist; the game isn't among them), so this needs
a build, once the routing fix is in. Do not queue a `page_rerender` for these pages
expecting a game to appear.

**Blast radius — one page.** A site-wide sweep (game-/tool- pages, checking for
`<canvas>` / `game-container` / `tool-page` markup) shows only `game-pathfinding` lost its
interactivity (`has_interactive_markup = f`, 2 sections). Every other game/tool page still
has its interactive markup. `game-auto-battler` was in the *same* linking batch but
survived — its `deployed_at` stayed 06-13, so its link rebuild never re-deployed, whereas
pathfinding's ran the full writer-to-deploy chain. The survivors remain at risk until the
routing is fixed. (`game-jelly-invaders` is separately in `needs_rebuild` with no deploy —
not a casualty.)

**Confirmed mechanism** (`game-pathfinding`, page 56af8679-…):
- `page_component_history` shows the hero held the full A* game at 2026-06-14 **19:21:21**
  — `<canvas id="gridCanvas">`, Wall/Mud/Eraser brushes, Run/Reset, 18,449 bytes of
  `tool-page` markup. Real, working, interactive.
- At **20:07:52** that was overwritten by a `hero` + `generic-text-block` layout
  (`save_page_sections_overwrite`).
- The writer was the pathfinding `link_resolution_rebuild` (`46c2da91`): its
  `page-build-handler` orchestration `00615292` (20:04→20:08:36) spawned `page-content-writer`
  `393df4bf` (20:04→20:07:46), whose `page_html` came out as hero + generic content with no
  game; `page-rerender` `4b820f4b` then assembled and deployed the gameless page.
- A `link_resolution_rebuild` — whose own spec says *"preserve the existing copy; this
  rebuild exists to re-resolve internal links"* — is handled by `page-build-handler`, which
  runs the full `page-content-writer`. The writer regenerates from `plan_sections`, the plan
  has no knowledge of the tool, and the interactive component is discarded.

Note on attribution: the work-item's `completed_at` (20:08:37) is the *end* of its
orchestration; the destructive write happened mid-run inside the child writer (20:07:46–52).
The orchestration trace is authoritative, not the work-item timestamp — see the retro note
in NOTES.

`page-build-handler` structure (confirmed): one linear flow `ensure_site_record → … →
plan_sections → call/spawn_content_writer → save_sections → deploy_page`, with no branch on
`item_type`. It has a `load_existing_content` step and `check_has_ready_sections` /
`check_content_produced` conditionals, so a preserve-existing path may exist but isn't being
taken for this item type.

**Step-config result (2026-06-22) — no `item_type` branch.** The query below was run; every
step's `next_step` is unconditional and the only conditionals are `check_page_found`,
`check_has_ready_sections` (`section_plan.ready_count > 0`) and `check_content_produced`
(`page_content.response.skipped != true`) — none keyed on item type. So a
`link_resolution_rebuild` runs the identical full path as a fresh build
(`load_existing_content → load_spec_sections → plan_sections → spawn/call_content_writer →
validate → save_sections → deploy`); it is not a bypassed preserve-branch — there is none.
**Why the tool dies precisely:** the rebuild plan is built from the page SPEC
(`load_page_sections_from_spec` → `plan_sections`); the tool was attached as a section's
`rendered_html` but isn't in the spec, so the plan omits it, the writer produces only the
planned sections, and `save_page_sections` (DELETE+INSERT of the produced set) drops it.
`load_existing_content` runs but unconditionally proceeds to `load_spec_sections`, so it does
not prevent the regeneration — not the lever as wired. The query run:
```sql
SELECT key AS step, value->>'action' AS action,
       value->'config'->>'condition'   AS condition,
       value->>'next_step'   AS next_step,
       value->'config'->>'then_step' AS then_step,
       value->'config'->>'else_step' AS else_step
FROM agent_definitions ad,
     jsonb_each(ad.default_config #> '{workflow,steps}')
WHERE ad.type = 'page-build-handler'
  AND key IN ('ensure_site_record','load_page_record','check_page_found','load_existing_content',
              'load_spec_sections','check_has_ready_sections','plan_sections','call_content_writer',
              'spawn_content_writer','check_content_produced','save_sections')
ORDER BY step;
```

**P4.1 result (2026-06-22) — links are build-resolved CTA fields, not prose.** On the §5
pages `content_data` holds no `<a>` at all; the only internal link is the hero CTA in
`rendered_html` (guide-economy-basics: `cta_url = /contact.html`, a phantom — real contact is
`/contact/index.html`). `resolve_internal_links_action.go` confirms CTAs are resolved at BUILD
time into `section.resolved_data` (merged into stored `content_data`, then rendered); it is
explicitly "a build-time augmenter, not a rendered-HTML patcher," its caller is
page-content-writer, and `check_phantom_internal_links` deliberately routes page-surface link
fixes to **page-build-handler (a rebuild)**, NOT to the resolver. So the §5 routing was by
design — and two earlier assumptions are now corrected:
- The hero `cta_url` is a **schema-sourced field** (hero component schema: `cta_url` source
  `pages.contact`), resolved when contact lived at `/contact.html` and now stale (contact is
  `/contact/index.html`) — NOT a resolver hub (the resolver *excludes* contact). Confirmed
  2026-06-22: `cta_url = /contact.html` is stored in the hero's `content_data`. **This re-opens
  a question I over-closed last turn:** does the Part-2 `page_rerender` path re-resolve
  schema-sourced fields like this `cta_url`? If yes, a re-render fixes the phantom AND preserves
  the tool — reviving "route `link_resolution_rebuild` → `page_rerender`". Re-resolving a moved
  page reference is the *point* of that path, so it's plausible; the §5 batch chose
  page-build-handler only because it predates Part 2 (06-21), so that's not evidence. Resolve
  with P4.2 below.
- The clobber is not specific to `link_resolution_rebuild`: a tool's interactive content is a
  `page_component` section that ISN'T in the page spec, so ANY full rebuild
  (`needs_page` / `needs_content_page` / `content_rewrite` / admin "Regenerate Page" / this)
  plans from the spec, omits the tool, and the DELETE+INSERT drops it.

**Fix direction (revised) — layered, structural first.**
1. **Safety net (cheap, immediate, all triggers): interactivity-aware guard in
   `save_page_sections`** — block/escalate a save that replaces a `<canvas>` / `data-component`
   section with non-interactive content. Same shape as the Part-1 text guard, keyed on
   interactivity. Turns the silent clobber into a visible block now.
2. **Structural root: preserve interactive sections through a rebuild** — carry forward
   existing interactive/tool sections that aren't in the spec (extend `load_existing_content` /
   the save to merge rather than DELETE-all), or represent the tool in the spec as a
   non-regenerated section. Makes every rebuild non-destructive to tools while leaving CTA
   re-resolution working via the existing writer path — so `link_resolution_rebuild` →
   page-build-handler keeps doing its job and stops eating the game. No routing change.
3. **Efficiency (optional, later): no-LLM CTA re-resolution** — extend the Part-2 `page_rerender`
   path to re-run `resolve_internal_links`' hub logic for CTA fields, then route
   `link_resolution_rebuild` there. Reuses Part 2 + the resolver; avoids the LLM writer for a
   link-only fix. Only worth it once (2) makes things safe.

Recommended order: 1 (protect) → 2 (fix) → 3 (optimise). The earlier "repoint
`link_resolution_rebuild` → `page_rerender`" plan is **deferred, not dropped** — pending P4.2 it
may be the cleanest, no-LLM link fix (if `page_rerender` re-resolves the schema-sourced CTA).
Either way it only covers the link route; Layers 1–2 remain the baseline for the other rebuild
triggers.

**Open sub-questions (not blocking the fix):** (a) does `save_page_sections` honor
`page_components.locked_at`? If a rebuild carries forward locked sections, locking the tool
section is a zero-code interim mitigation — worth confirming (013 says execution agents process
explicit items regardless of lock, so probably not, but check). (b) `/contact.html` origin: the
resolver EXCLUDES contact from CTA targets, so it didn't set that — likely a hardcoded hero
default or a pre-resolver value, meaning the hero CTA may not be going through the resolver on
these pages. (c) stale duplicate `z_context/check_phantom_internal_links.go` still routes
page-link fixes to `internal-link-resolver` (the live `platform/...` copy routes to
page-build-handler) — delete it.

**Two guards worth adding regardless of the routing decision:**
1. **Stamp `source_item_id` into `page_component_history`.** The 20:07 overwrite wrote NULL
   `source_item_id`, which is why this had to be traced by timing + orchestration rather than
   by id. The save path should record the driving work item.
2. **Interactivity guard in the save path.** The Part 1 text-loss regression guard (>75% text
   drop) waved this through, because an 18KB game → 2KB text swap is mostly markup/JS loss,
   not prose. An overwrite that drops a `<canvas>` / `data-component` game from a section
   should be blocked or escalated, not silently saved.

**Remediation for the lost game (TBD, after the routing fix):** re-run the tool-recreation
for `game-pathfinding` to regenerate and reattach it. Sequence: size it (done — one page),
land the routing fix so it can't re-clobber, then re-create. No imminent linking batch, so
not urgent. Interacts with Part 3 (the same page carries the `needs_page:` mis-key on its
tool-recreation item).

## P4.1 — Diagnostic commands: link representation + routing (run these next)

Goal: answer the one question that picks the fix route — are internal links **resolved at
render time** or **baked into stored LLM prose**? — then locate where `link_resolution_rebuild`
items get their handler so the repoint is a known change. (Site `e33263f4-74f8-494f-b191-546845dbbddf`.)

**1. Link representation (SQL) — the deciding question.** Pick a content page that carries
inline internal links — a guide is a good bet; the §5 batch re-resolved links on
`guide-rng-design`, `guide-economy-basics`, `contact-index`. Compare where `<a href>` lives:
in stored `content_data` (the LLM prose → baked) vs only in `rendered_html` (→ applied at render):
```sql
-- resolve the page_id(s)
SELECT id, name, url FROM pages
WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
  AND name IN ('guide-rng-design','guide-economy-basics','contact-index')
ORDER BY name;

-- per section, where do internal links appear?
SELECT pc.position, pc.slot_name,
       (pc.content_data::text ~ '<a [^>]*href=') AS links_in_content_data,
       (pc.rendered_html      ~ '<a [^>]*href=') AS links_in_rendered_html,
       substring(pc.content_data->>'content' from '<a [^>]{0,90}') AS sample_content_data_link,
       substring(pc.rendered_html           from '<a [^>]{0,90}') AS sample_rendered_link
FROM page_components pc
WHERE pc.page_id = '<PAGE_ID from above>'
ORDER BY pc.position;
```
Read it:
- `links_in_content_data = t` (the prose itself holds `<a href="/...">`) → links are **baked
  into prose**. A plain re-render reproduces them unchanged, so it will NOT fix a moved link.
  Route → links-only rewrite over stored content (or the existing `internal-link-resolver`),
  preserving sections.
- `links_in_content_data = f` but `links_in_rendered_html = t` → links are **applied at
  render** (resolved field/token). Route → the Part 2 `page_rerender` path re-resolves them
  while preserving every section. Cheap, preferred route.
- Eyeball the samples: a literal path (`/games/pathfinding/index.html`) in content_data =
  baked; a token (`{{...}}`, `[[link:...]]`, a bare id) = resolved. Check more than one page
  before concluding — the writer may bake links on some section types and not others.

**2. What the `link_resolution_rebuild` items actually carry (SQL).**
```sql
SELECT id, handler_agent, item_key, source, status, spec, created_at, completed_at
FROM site_work_items
WHERE item_type = 'link_resolution_rebuild'
ORDER BY created_at DESC
LIMIT 10;
```
If `spec` only names the page (no specific links/anchors), the item is a generic "rebuild this
page" signal — which is why it landed on the full builder. If it carries link detail, a
links-only handler can consume it directly.

**3. Where the route is set + what the resolver does (repo grep — run in the code checkout).**
```bash
# who creates link_resolution_rebuild items and sets handler_agent (the §5 batch creator)
grep -rn "link_resolution_rebuild\|linking_rebuild" --include=*.go --include=*.sql .

# the existing links-only action — does it rewrite stored content in place, or feed the writer?
grep -rn "resolve_internal_links\|internal-link-resolver\|ResolveInternalLinks" --include=*.go .
# then read the action file it points to: does it UPDATE page_components.rendered_html /
# content_data directly (in-place, section-preserving), or produce input for page-content-writer?
```
Decision: if `internal-link-resolver` rewrites stored content **in place** (preserving
sections), the fix is to route `link_resolution_rebuild` to it instead of `page-build-handler`
— a one-line `handler_agent` change in the §5 creator. If it only feeds the writer, use the
`page_rerender` route (render-resolved links) or extend the resolver to operate on stored content.

**4. (Optional) logs — only if `internal-link-resolver` runs as a live agent.**
```bash
kubectl -n ai-persona-system get pods | grep -i link
kubectl -n ai-persona-system logs -l app=agent-internal-link-resolver --since=24h --tail=200 2>/dev/null | head -50
```

**Outcome — see the P4.1 result and revised (layered) fix direction in Part 4 above.** In
short: links are build-resolved CTA fields; the stale hero `cta_url` is a schema-sourced
`pages.contact` reference (`/contact.html`, now moved to `/contact/index.html`), and the
clobber affects *any* full rebuild. So the baseline fix is the interactivity guard (Layer 1) +
preserving interactive sections through the rebuild (Layer 2). Whether a `page_rerender` can
*also* fix the link (re-resolving the schema CTA) is an open, testable question — P4.2 below;
if it can, routing the link rebuild there becomes the no-LLM link fix. The two guards
(`source_item_id`; interactivity-aware save) and re-creating `game-pathfinding` follow regardless.

## P4.2 — Does `page_rerender` re-resolve a schema-sourced hero CTA? (decisive, low-risk)

The hero `cta_url` is sourced from `pages.contact` and is stale (`/contact.html`). If the Part-2
`page_rerender` path re-resolves it from the current contact page (`/contact/index.html`), then
routing `link_resolution_rebuild` → `page_rerender` fixes phantom CTAs with no LLM and preserves
every section. Test on `guide-economy-basics` — a NON-tool content page (hero + generic-text),
so re-rendering it risks nothing.

```sql
-- 1. baseline
SELECT slot_name, content_data->>'cta_url' AS cta_url, updated_at
FROM page_components WHERE page_id='8ed97fd2-2e33-4bdb-a24e-0f3badaca382' AND slot_name='hero';

-- 2. queue a section-re-rendering page_rerender for guide-economy-basics
INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary,
   spec, priority, status, created_by, handler_agent, item_key)
VALUES
  ('e33263f4-74f8-494f-b191-546845dbbddf',                 -- site_id (gamesdesign) — NOT the page_id
   'runbook', 'content', 'page_rerender', 'medium',
   'P4.2 diag: re-render guide-economy-basics to test CTA re-resolution',
   '{"reason":"image_landed","page_id":"8ed97fd2-2e33-4bdb-a24e-0f3badaca382","page_name":"guide-economy-basics"}'::jsonb,
   50, 'triaged', 'runbook', 'page-rerender',
   'page_rerender_guide-economy-basics_diag_p42')
RETURNING id, site_id, item_type, status, handler_agent, item_key, spec;
-- (pipeline is a non-gating category — claim is by status+site; 'content' fits a page-content
--  re-render. status='triaged' is the claimable state. item_key is unique — bump '_p42' to re-run
--  while a prior one is still non-terminal. track with:
--    SELECT id,status,claimed_by,left(error,160),completed_at FROM site_work_items
--    WHERE item_key='page_rerender_guide-economy-basics_diag_p42'; )

-- 3. after it completes, re-check
SELECT position, slot_name, content_data->>'cta_url' AS cta_url, updated_at,
       (rendered_html ~ '<a [^>]*href=') AS has_link
FROM page_components WHERE page_id='8ed97fd2-2e33-4bdb-a24e-0f3badaca382' ORDER BY position;
```
Read it:
- `cta_url` flips `/contact.html` → `/contact/index.html`, both sections present, `updated_at`
  advanced → **`page_rerender` re-resolves schema CTAs.** Promote Layer 3: route
  `link_resolution_rebuild` → `page_rerender` as the primary, no-LLM link fix. Still ship the
  Layer-1 guard; Layer 2 still wanted for non-link rebuilds.
- `cta_url` stays `/contact.html` → `page_rerender` does NOT re-resolve CTAs. The link fix needs
  the writer/resolver path; rely on Layer 2 to keep that rebuild non-destructive, plus the
  Layer-1 guard.

Alternative (no live test): read `rerender_page_sections_action.go` — does its field
re-resolution run the standard source resolver (incl. `pages.*` sources), or only query +
hero-image? That answers P4.2 directly.

---

## Where things stand — run order across the open parts

A quick index now that the runbook spans four parts (commands live in each part):

| Part | State | Immediate next | Commands |
|---|---|---|---|
| 1 — result contract | DONE (shipped 06-18, verified) | — | §1 |
| 2 — re-render path | Verified for `image_landed` (Test 1 PASS) | Run P2.4–P2.7; confirm live `/index.html` has all 5 sections | P2.3 (done) / P2.4–P2.7 |
| 3 — item_key canonicalization | Code prepared, NOT applied | Apply after Part 2 verifies, then P3.4 tests | P3.2–P3.4 |
| 4 — interactive page clobber | Cause confirmed; diagnostic done; **fix = interactivity guard + preserve interactive sections in rebuild** (not a routing repoint) | Implement Layer 1 guard, then Layer 2 carry-forward; re-create `game-pathfinding` | Part 4 (revised fix direction) |

Suggested order: **P4.1 now** (it's independent and decides the Part 4 fix), in parallel finish
**P2.4–P2.7**, then **apply Part 3** as its own commit, then implement the Part 4 routing fix +
guards and re-create the pathfinding game.
