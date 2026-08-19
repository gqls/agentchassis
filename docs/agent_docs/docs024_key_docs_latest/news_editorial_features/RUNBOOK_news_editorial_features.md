# RUNBOOK — news editorial features

Every query and command this workstream had to get right, with its gotcha
attached. When one changes, change it **here**.

DB access is always:
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

---

## 1. The state-of-the-feed query

One query, everything that matters about where the pipeline stops:

```sql
SELECT count(*) AS total,
       count(*) FILTER (WHERE topics IS NOT NULL AND topics <> '[]'::jsonb) AS with_topics,
       count(*) FILTER (WHERE entity_ids IS NOT NULL AND array_length(entity_ids,1)>0) AS with_entities,
       count(*) FILTER (WHERE duplicate_of IS NOT NULL) AS grouped,
       count(*) FILTER (WHERE published_page_id IS NOT NULL) AS published
  FROM content_feed_items;
```

**Gotcha:** `array_length(x,1)` returns NULL for an empty array, not 0 — so the
`FILTER` must test `> 0` and tolerate NULL, which `array_length(...)>0` does
(NULL is not true). Testing `entity_ids IS NOT NULL` alone would count an empty
`{}` array as populated.

## 2. Proving a column has no writer

```bash
grep -rn "duplicate_of" --include=*.go platform/ internal/ pkg/
grep -rn "entity_ids"   --include=*.go platform/ internal/ pkg/
```

**Gotcha:** an empty grep is only evidence if the same grep shape finds something
when it should. Run a control in the same breath — `grep -rn "relevance_score"`
over the same paths returns hits, which proves the search is looking where the
writes live. Without that control an empty result is indistinguishable from a
mistyped path.

## 3. Proving an agent definition is absent

Always include a positive control **in the same query**, or a zero row count is
just as consistent with a broken filter:

```sql
SELECT type FROM agent_definitions WHERE type IN
 ('article-rewriter','feed-publisher','feed-lifecycle','news-analyst',
  'story-researcher','analysis-writer','visualization-renderer','data-chart-generator');
-- expect 0 rows

SELECT type FROM agent_definitions WHERE type='feed-triage' LIMIT 1;
-- expect 1 row  <- this is the control
```

**Gotcha:** the live roster query needs the snapshot guard, or you read a
historical row. The full-fidelity form is
`WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL`.

## 4. Proving a table is absent

```sql
SELECT to_regclass('public.event_timeline') IS NOT NULL AS event_timeline_exists,
       to_regclass('public.topic_packages') IS NOT NULL AS topic_packages_exists;
```

`to_regclass` returns NULL rather than raising, so this never aborts a
transaction the way `\d event_timeline` would.

## 5. Reading what a component actually is — never a Go grep

The chart renderers are **rows in `content_components`**, not code. A
`--include=*.go` grep will report them absent, which is exactly the mistake
`register/visualisation-and-charts.md` was written to stop.

```sql
SELECT name, component_level, render_mode, is_active
  FROM content_components
 WHERE name IN ('evidence-chart','evidence-timeseries','mechanism-flow')
    OR name ILIKE '%news%';
```

**Gotcha:** the homepage news component is named **`Latest News Feed`** (spaces,
title case), not `latest-news`. A kebab-case `IN (...)` list silently misses it.
Use `ILIKE '%news%'` when enumerating.

## 6. The triage prompt (where concept extraction actually happens)

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps')
  FROM agent_definitions
 WHERE type='feed-triage' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

The `topics` array comes back from this prompt's JSON and is written at
`platform/orchestration/actions/feed_triage_actions.go:245` (parsed at :204).

## 7. Firing the diagnosis loop for this workstream

```bash
SLUG=<slug> ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh "<symptom>"
```

**Gotchas, all of which cost time if missed:**
- The script prints **two** correlation ids. The intake one is *not* the key the
  artifacts are written under. Save the one labelled `RUN_CORRELATION_ID` — the
  dispatch loop mints it and stamps it back as `spec.dispatch_correlation_id`.
- Do **not** also publish by hand. `diagnose-pipeline-trigger` is enabled, so the
  loop claims the item within ~60s; a manual publish is a second full run on a
  correlation that cannot be joined to the first (`bugs_open/124`).
- The symptom must state a **mechanism** and **point at** tables/symbols without
  asserting counts — the loop fetches and cites the numbers itself.
- Poll with the run correlation, not the intake one:

```sql
SELECT current_step, status, updated_at FROM orchestration_states
 WHERE correlation_id::text='<RUN_CORRELATION_ID>' ORDER BY created_at;

SELECT status FROM site_work_items WHERE item_key='needs_diagnosis:<slug>';
```

- A code-only symptom (no site, no pages) makes the script warn *"nothing to key
  coverage on — dispatching blind"* and *"Subject: NONE — persist_note will
  SKIP"*. Both are expected for a platform-wide question, not failures. It means
  the verdict lands in `diagnosis_artifacts` rather than in a per-site
  `doc_notes` row.

## 8. Reading the verdict when it lands

```sql
SELECT body FROM doc_notes
 WHERE categories ? 'diagnosis'
 ORDER BY created_at DESC LIMIT 3;

SELECT * FROM diagnosis_artifacts
 WHERE correlation_id::text='<RUN_CORRELATION_ID>';
```

## 9. Checking the pools are still dormant

```sql
SELECT domain, status FROM sites WHERE status='pool' ORDER BY domain;
-- 17 rows, all *.internal, as of 2026-08-19
```

They are invisible to the fleet sweeps because those only walk deployed sites,
and to the news machinery because that only fires on classified sites with live
pages. Pools have neither. Verify rather than assume before arming one.
