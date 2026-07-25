# RUNBOOK — brochure component library / fundamentallyai.com

Every command here had a gotcha attached. The gotcha is the point — when one
changes, change it HERE, not in your scrollback.

Site: `fundamentallyai.com`, site_id `199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`,
current plan_id `81741260-6447-492c-bf98-4b3c185f8e7b`.

DB access (the only route):
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

---

## Rebuilding a page (the whole minefield)

**Do not re-queue a historic work item.** `stale-work-item-reaper` (hourly) parks
any `triaged` build item whose **`created_at`** is 48h+ old — row age, not time
waiting (`bugs_open/070`). A five-day-old row is born eligible, so your
minutes-old request gets stamped `[stale: triaged 48h+]` and parked. Observed
here 12 minutes after a re-queue.

**Insert a fresh row instead.** `unresolved` is in `idx_swi_dedup`'s terminal set,
so the same `item_key` inserts cleanly beside the parked one.

```sql
INSERT INTO site_work_items (
  site_id, item_type, item_key, status, pipeline, summary, spec,
  handler_agent, source, created_by, attempt_count, max_attempts, created_at, updated_at)
SELECT site_id, item_type, item_key, 'triaged', pipeline,
       'Rebuild <page>: <why, in words a stranger can act on>',
       spec, handler_agent, 'operator:<workstream>', 'operator:<workstream>',
       0, max_attempts, NOW(), NOW()
FROM site_work_items WHERE item_key='needs_page:<page>'
  AND site_id=(SELECT id FROM sites WHERE domain='fundamentallyai.com') LIMIT 1;
```

- **`created_by` is NOT NULL with no default** — a copy-INSERT that omits it dies
  with 23502. Name it.
- Also set `pages.build_status='needs_rebuild'` for a full rebuild.
- A hand INSERT **bypasses** the Go-side two-strike suppression in
  `insertWorkItem`. Deliberate; know you are doing it.
- With fresh rows, queueing several at once is safe — the reaper was the whole
  reason batches used to park. Verified 2026-07-25 (contact claimed in ~4 min
  alongside three siblings).

**A page with no plan sections cannot build**, and fails fast rather than
loudly: `plan_sections` → `check_has_ready_sections` → `mark_no_ready_sections`
in ~38s, no LLM spend, item → `needs_human_review`. Check first:

```sql
SELECT page_name, count(*) FROM site_plan_sections
WHERE plan_id='81741260-6447-492c-bf98-4b3c185f8e7b' GROUP BY 1 ORDER BY 1;
```

Placing sections (this is also what makes a component survive future rebuilds):

```sql
INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name) VALUES
 ('<plan>','<page>',0,'hero'), ('<plan>','<page>',1,'generic-text-block'), ...;
```
`idx_site_plan_sections_key` is UNIQUE(plan_id, page_name, ordering) — to insert
in the middle, shift high (+100), insert, bring down (−99). An in-place
`ordering+1` collides.

---

## Republishing a page after a DIRECT data edit (no content regeneration)

**Do not use `049b_deploy_single_page.sh`.** Its bare `action=orchestrate`
envelope hits the kubectl-run stdin race and can fail **silently** — no
orchestration row, no work item, no log. Four calls on 2026-07-25 produced zero
rows, and a completed link fix sat unpublished.

**`scripts/republish_page_086.sh` (the `086` direct-orchestrator envelope,
`action=process` + inline workflow) DOES work** — verified 2026-07-25 on the
homepage: orchestration row COMPLETED, live page clean. Budget ~2 minutes for the
row to appear and don't conclude it failed before then; my first check at +45s
found nothing and I nearly declared it dead.

**The route that needs no Kafka envelope at all is the work-item queue** — prefer
it when the queue is draining (check: has anything been claimed in the last ~10
minutes?), because it leaves a durable, inspectable row:

```sql
INSERT INTO site_work_items (
  site_id, item_type, item_key, status, pipeline, summary, spec,
  handler_agent, source, created_by, attempt_count, max_attempts, created_at, updated_at)
VALUES ((SELECT id FROM sites WHERE domain='fundamentallyai.com'),
  'page_rerender',
  'page_rerender_<page>_199733a8-ac9c-4c30-b2ce-65ecdac6f3bd_assemble_<tag>',
  'triaged','build','Republish <page>: <why>',
  jsonb_build_object('domain','fundamentallyai.com','page_id','<page_id>',
                     'filename','<path/index.html>','page_name','<page>'),
  'page-rerender','operator:<workstream>','operator:<workstream>',0,3,NOW(),NOW());
```

`item_type='page_rerender'` + handler `page-rerender` re-renders from **stored
`content_data`** (no LLM). Do NOT use `needs_page` for this — that routes into
the full LLM pipeline and rewrites all the copy (016b §9,
"`spec.reason` does not make `needs_page` scoped").

**Verify by payload, never by the printed correlation id:**
```sql
SELECT status, current_step, created_at FROM orchestration_states
WHERE initial_request_data->'input_data'->>'page_id'='<page_id>'
ORDER BY created_at DESC LIMIT 1;
```

---

## Internal-link census (the check the gate does but does not keep)

The site is **`.html`-based**: `/capabilities` 404s, `/capabilities.html` is 200.
Extension-less internal hrefs are broken, always.

```sql
WITH hrefs AS (
  SELECT p.name AS page, unnest(regexp_matches(pc.rendered_html, 'href="(/[^"#?]*)"', 'g')) AS href
  FROM pages p JOIN sites s ON s.id=p.site_id JOIN page_components pc ON pc.page_id=p.id
  WHERE s.domain='fundamentallyai.com'
)
SELECT href, string_agg(DISTINCT page, ', ') AS on_pages, count(*) AS n
FROM hrefs
WHERE NOT EXISTS (SELECT 1 FROM pages p2 JOIN sites s2 ON s2.id=p2.site_id
                  WHERE s2.domain='fundamentallyai.com' AND p2.url = hrefs.href)
GROUP BY href ORDER BY href;
```
Zero rows = every internal link resolves. **Gotcha in my first version:** I
filtered the target lookup on `build_status='deployed'`, which mislabelled
`/contact` as an invented page when `/contact.html` was serving 200 — the page
row was `needs_rebuild` while the artefact was live. Don't filter on build_status
when asking "does this URL exist".

Recovering what the gate found on a build that **passed** (warnings are not
persisted — `bugs_open/071`; this dies with `collected_data` at ~24h):
```sql
SELECT jsonb_pretty(jsonb_agg(DISTINCT jsonb_build_object(
         'type', i->>'type', 'sev', i->>'severity', 'value', i->>'value')))
FROM orchestration_states os,
     jsonb_array_elements(os.collected_data->'validation_result'->'issues') i
WHERE os.orchestration_id::text LIKE '<oid-prefix>%';
```

Blocker/error detail on a build that **failed** (this one IS persisted):
```sql
SELECT occurred_at, jsonb_pretty(context) FROM agent_error_log
WHERE domain='fundamentallyai.com' AND error_code='CONTENT_VALIDATION_BLOCKER_DETAIL'
ORDER BY occurred_at DESC LIMIT 1;
```
`agent_error_log` has **`occurred_at`, not `created_at`**.

---

## Editing rendered content safely

Replace **quoted exact** strings, not bare paths: `'"/contact"'` → `'"/contact.html"'`.
The leading `"/` stops `/contact` also matching `/contact.html` (→ `.html.html`)
and `/review-council` matching `/multi-agent-review-council`. Apply to
`rendered_html` **and** `content_data`, or the next re-render undoes it. Snapshot
first: `CREATE TABLE bak_pc_<slug>_<date> AS SELECT pc.* FROM page_components pc …`.

Full worked example: `sql/`-adjacent script used on 2026-07-25 for the link
repair (14 replacements, post-check included).

## Shell traps hit on this workstream

- **`grep -rn` goes silent-binary** when another session leaves a NUL byte in the
  tree — it found nothing for a string that was there. **Use `git grep`.**
- `printf` breaks on rendered HTML containing `%` and `)` — build SQL with
  `echo`/`cat` heredocs only.
- The Bash tool's working directory **persists between calls**. A relative `cd`
  that worked last call fails this call; `cd X && cat >> f << EOF` then silently
  writes nothing (the heredoc is consumed, the `&&` short-circuits). Use absolute
  paths for appends.
- Backticks inside `git commit -m` execute in bash — use `-F -` with a heredoc.
