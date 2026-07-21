# RUNBOOK — bug 025 acceptance test (live rebuild)

Proves setting `pages.content_direction` steers the generated copy. Run on the
**vonc.com** test site (`about` page) 2026-07-21.

Page: `vonc.com/about` — site_id `9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`,
page_id `a28abcd7-186b-4a33-9b89-5d7bfd727012`.

## 1. Set a distinctive, greppable direction + baseline check
Marker chosen with **no quotes/apostrophes** so the SQL literal stays clean; the
phrase is off-theme for vonc so a natural occurrence is near-impossible.
```sql
UPDATE pages SET content_direction = '{
  "instruction": "In the main body paragraph of this page, weave in this exact sentence verbatim: Quite simply, it began with one stubborn question.",
  "format": "plain, confident, first-person plural; short paragraphs",
  "avoid": ["fabricated statistics", "invented client names"]
}'::jsonb, updated_at=now()
WHERE id='a28abcd7-186b-4a33-9b89-5d7bfd727012';
```
Baseline must be 0:
```sql
SELECT count(*) FILTER (WHERE rendered_html ILIKE '%stubborn question%')
FROM page_components WHERE page_id='a28abcd7-186b-4a33-9b89-5d7bfd727012';
```

## 2. Trigger the rebuild
`build-pipeline-trigger` (scheduled_task, every 120s) only fires a site whose
pre_query finds a **`triaged`, pipeline=`build`** work item — setting
`build_status='needs_rebuild'` alone does NOT create one (the site-work-orchestrator's
WriteBuildItemsAction does). So create the work item directly; the dispatch loop then
runs `page-build-handler` → `load_page_record` (my code) → `current_page` → writer.
Unique item_key to avoid touching the normal `needs_page:about` dedup slot.
```sql
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary,
  spec, page_id, priority, handler_agent, status, created_by, item_key, batch_id,
  attempt_count, max_attempts)
VALUES ('9ec3b9ee-...', 'planner','build','needs_content_page','high',
  'Build page: about (bug025 acceptance test)',
  '{"page_name":"about","page_id":"a28abcd7-...","name":"about","page_type":"content",
    "site_id":"9ec3b9ee-...","url":"/about"}'::jsonb,
  'a28abcd7-...', 10, 'page-build-handler','triaged','bug025-acceptance-test',
  'bug025:needs_page:about', gen_random_uuid(), 0, 3);
```
Work item created: `61de62d8-a1c5-4947-9a32-c3021d741625`.

## 3. Watch (background monitor scratchpad/mon025.sh)
Claim path: `triaged` → `claimed` → terminal. Dispatch loop claims status IN
('triaged','approved') and checks handler_agent is registered.

## 4. Verify — the PASS condition
```sql
SELECT slot_name, left(regexp_replace(rendered_html,'<[^>]+>',' ','g'),200)
FROM page_components WHERE page_id='a28abcd7-...'
  AND rendered_html ILIKE '%stubborn question%';
```
PASS = the marker sentence appears in a regenerated section (it was absent at
baseline). That proves the column value reached the writer prompt and steered the
copy — the exact behaviour bug 025 said did not exist. Verify against the SAVED
SECTION, never a `complete` status.

## 5. Cleanup after PASS
```sql
UPDATE pages SET content_direction=NULL WHERE id='a28abcd7-...';  -- leave the test page clean
```
(The marker copy will wash out on the next natural rebuild; optionally rebuild once
more with content_direction NULL to restore the original tone.)
