# RUNBOOK — bug 392

Every query below was run first-hand on 2026-08-25 and returned the figure quoted in PLAN/NOTES.
Gotchas are attached to the command, not kept separately.

## 1. The instrument — and why the obvious one is wrong

⚠ **Three instruments disagree. State which you used or your number means nothing.**

- `page_components.rendered_html` includes template nav/hero/CTA links, so a page whose writer
  emitted nothing still reads 2–3. Fleet: 140/737 zero.
- `page_components.content_data` `href="/` is the **writer's prose anchors**. Fleet: 411/736 zero.
- Structured link fields (`cta_url`, `link_url`, `hero_url`) are real links but carry no `href=`,
  so the regex above misses them **by design** — they are chosen by a different mechanism.

The prose-anchor census, by page type:

```sql
WITH pl AS (
  SELECT p.id, p.page_type,
         sum((SELECT count(*) FROM regexp_matches(coalesce(pc.content_data::text,''),
                                                  'href=\\?"/[^"]*','g'))) AS cd_links
  FROM pages p JOIN page_components pc ON pc.page_id=p.id
  WHERE p.status='active' AND p.build_status='deployed'
  GROUP BY p.id, p.page_type)
SELECT page_type, count(*) FILTER (WHERE coalesce(cd_links,0)=0) AS zero, count(*) AS total
FROM pl GROUP BY page_type ORDER BY zero DESC;
```

⚠ **`\\?` is load-bearing** — inside `content_data::text` the quote may be backslash-escaped.
⚠ **`page_components` is keyed by `page_id`; `site_components` is keyed by `site_id` and has no
`page_id` column at all.** Joining the wrong one returns zero rows and reads as "no components".

Owned-page split of the gated population (the second gate, 48 of 48):

```sql
WITH pl AS (
  SELECT p.id, coalesce(p.rebuild_policy,'(null)') AS pol,
         sum((SELECT count(*) FROM regexp_matches(coalesce(pc.content_data::text,''),
                                                  'href=\\?"/[^"]*','g'))) AS cd
  FROM pages p JOIN page_components pc ON pc.page_id=p.id
  WHERE p.status='active' AND p.build_status='deployed'
    AND p.page_type IN ('blog-post','guide','content')
  GROUP BY p.id, p.rebuild_policy)
SELECT pol, count(*) FILTER (WHERE coalesce(cd,0)=0) AS zero, count(*) AS total
FROM pl GROUP BY pol;
```

## 2. The finding-code rows, and the join the bug file missed

```sql
SELECT id, occurred_at, site_id, orchestration_id, work_item_id, domain,
       resolved, context
FROM agent_error_log WHERE error_code='LINK_CONTEXT_UNAVAILABLE' ORDER BY occurred_at;
```

⚠ **`context` holds only `site_id` — but the ROW carries `orchestration_id` as a first-class
column**, filled by `actionJoinIdentity` (`log_action_error.go:99-112`). That is the exact join,
and the bug file does not mention it. Resolve it:

```sql
SELECT orchestration_id,
       collected_data->'input_data'->'current_page'->>'id'   AS page_id,
       collected_data->'input_data'->'current_page'->>'name' AS page_name,
       collected_data->'input_data'->'current_page'->>'url'  AS url,
       collected_data->'input_data'->>'domain'               AS domain
FROM orchestration_states WHERE orchestration_id IN (…);
```

⚠ **The primary key is `orchestration_id`, not `id`** — `SELECT id FROM orchestration_states`
errors, it does not return null.
⚠ `page_id` is NOT at `input_data.page_id`; it is at `input_data.current_page.id`.
⚠ This join is best-effort over time: orchestration rows are reaped, the log row lives 365 days.
That asymmetry is why the writer should record `page_name` in `context` itself.

## 3. Which route actually re-runs the writer

**Read the live agent definitions, not the seed files** — CLAUDE.md: *the seed is not the system*.

```sql
SELECT type,
       (default_config->'workflow'->'steps')::text ILIKE '%page-content-writer%' AS spawns_writer,
       (default_config->'workflow'->'steps')::text ILIKE '%edit_live%'          AS knows_edit_live
FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND type IN ('page-build-handler','page-rerender');
-- page-build-handler | t | t
-- page-rerender      | f | f
```

```sql
SELECT type FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND (default_config->'workflow'->'steps')::text ILIKE '%prepare_link_context%';
-- page-content-writer   (exactly one, fleet-wide)
```

**Conclusion: a `page_rerender` item cannot restore prose links.** Bug 392's candidate 1 is wrong
as written.

## 4. Repair-route health, before you rely on it

```sql
SELECT handler_agent, status, count(*) FROM site_work_items
WHERE item_type='content_rewrite' AND created_at > now() - interval '14 days'
GROUP BY 1,2 ORDER BY 3 DESC;
-- page-build-handler: complete 93 | wont_fix 53 | failed 45 | needs_human_review 21
```
~21% fail or are refused. Any plan that assumes a filed item equals a repaired page is wrong.

## 5. Is the framework we are building into actually running?

```sql
SELECT source, count(*), max(created_at) FROM site_work_items
WHERE created_at > now() - interval '3 days' GROUP BY 1 ORDER BY 2 DESC LIMIT 8;
-- discovery | 659 | (minutes ago)
```
⚠ Detection existing is not detection running (MEMORY: *"detection works; SCHEDULE and DISPATCH
do not"*). Re-run this before assuming a new discovery check will ever fire.

## 6. Canary induction — the safe route

`resolveLinkContextSiteID` (`prepare_link_context_action.go:296-311`): *"An explicit config value
wins"*. An unparseable `site_id` in the step config lands in the same `default:` arm as a real
timeout → `dbFailure`, `dbConsulted=false`, `degraded=true`. The writer's `collected_data`
fallback is empty (0/26, `:84-90`) so it will not mask the degrade.

⚠ **Do not edit the shared `page-content-writer` definition in place** — it is fleet-wide for the
duration. Dispatch at a short-lived clone and delete it after.
⚠ **Do not use `LOCK TABLE pages`** to force a real timeout: it stalls every query touching
`pages` fleet-wide for up to the pgbouncer `query_timeout` (~130s).
⚠ **Do not hand-insert a synthetic row**: it proves the reader, not the chain, and pollutes a
code whose rows are evidence.

Canary candidates with today's prose-link counts (a page that HAS links makes the before-state
non-vacuous): `pool-energy-utilities.internal` about (2) / faq (3); cookly.uk
`/guides/tool-takeaway-cost-comparison-guide.html` (3); garden-tools.uk `/about.html` (3).

## 7. Verifying at the artefact

Use `scripts/probe-page-url.sh` — it reads the recorded `pages.url` (it structurally cannot
compose one) and runs both per-domain controls. ⚠ Never compose a URL from `pages.name`: that
mistake filed bug 387 and has now happened four times.
