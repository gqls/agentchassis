# RUNBOOK — dartsonline.com traffic & affiliate-readiness

Commands that were hard to get right, with the gotcha attached. Change them HERE.

`SITE=5fe8785b-223d-41a3-88ee-c07187622381` · domain `dartsonline.com`
DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## 1. Applying a spec change

**Ordering is FORCED by the schema.** `idx_site_specs_current` is UNIQUE on
`(site_id, aspect) WHERE is_current` — the supersede must commit before the insert, or
the insert violates the index. Always in one transaction, always with a `bak_` table first.

```sql
BEGIN;
CREATE TABLE IF NOT EXISTS bak_darts_<aspect>_<yyyymmdd> AS
  SELECT * FROM site_specs WHERE site_id='<SITE>' AND aspect='<aspect>';
UPDATE site_specs SET is_current=false, superseded_at=now()
  WHERE site_id='<SITE>' AND aspect='<aspect>' AND is_current=true;
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by, notes)
  VALUES (...);
COMMIT;
```

To preserve unrelated keys, build the new `data` FROM the superseded row rather than
retyping it — `SELECT (b.data - 'deadkey') || jsonb_build_object(...) FROM site_specs b
WHERE ... AND b.is_current=false ORDER BY b.created_at DESC LIMIT 1`.

## 2. Checking a spec for banned content — per KEY, never over the blob

```sql
-- RIGHT: tells you WHICH key holds the hit
SELECT key FROM site_specs ss, jsonb_each(ss.data)
WHERE ss.site_id='<SITE>' AND ss.aspect='briefing' AND ss.is_current=true
  AND value::text ILIKE '%Portland%';
```

**Gotcha (cost real time 2026-07-29):** `WHERE data::text ILIKE '%stock%'` reports a hit
on a row whose only match is an `honesty_rails` entry saying *"Never claim to stock…"*.
A prohibition contains the word it prohibits. Once a fix adds negative instructions, a
whole-blob substring check can no longer tell "removed" from "forbidden". Go per-key.

## 3. Reading the live site without hanging the shell

```bash
curl -s --max-time 25 https://dartsonline.com/about.html -o /tmp/.../about.html -w "HTTP %{http_code} %{size_download}B\n"
for s in "Portland" "we stock" "we carry" "Red Dragon"; do
  printf "%-12s %s\n" "$s" "$(grep -ioF "$s" about.html | wc -l)"; done
```

**Two gotchas, both hit on 2026-07-29:**
- A bare `curl` with no `--max-time` hung past the 120 s tool timeout. Always cap it.
- `grep -oiE "[^<>]{0,80}(alt1|alt2|…)[^<>]{0,80}"` on a 28 KB page backtracks
  catastrophically and never returns. Use `grep -oF` counts, or strip tags in Python
  first and search the text.

## 4. Nav: the control surface is `pages`, not `site_nav_items`

`populate_nav_tables_action.go:147-150` **DELETEs every nav row for the site and rebuilds**
from `pages`. Hand-editing `site_nav_items` is wasted work. Set `pages.in_header`,
`nav_label`, `nav_order`, then trigger a nav rebuild.

Page selection filter is `status IN ('active','deployed','pending')`
(`populate_nav_tables_action.go:240-245`) — `archived` drops out. `archived` is the house
value (the only non-`active` status in the table fleet-wide).

The never-deployed prune happens LATER, at render, in `GetNavItems`
(`nav_tables.go:215-240`), and logs:
```
GetNavItems: dropped nav items whose target page has never been deployed
```
So a `planned` page flagged `in_header` is pruned at render but still sits in the nav
tables — and if the chrome is stale it still SERVES. Chrome is a stored artefact
(`bugs_open/117`): a page re-render does not regenerate it.

Fleet check for this class before assuming it is generic:
```sql
SELECT s.domain, count(*) FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.in_header AND p.deployed_at IS NULL AND p.status IN ('active','deployed','pending')
GROUP BY 1 ORDER BY 2 DESC;   -- 2026-07-29: dartsonline.com | 4, and nothing else
```

## 5. Which pages are REAL — plan vs pages table

The `pages` table accumulates orphans from superseded plans. The current plan is the
authority on what should exist:
```sql
SELECT spp.name, spp.role, spp.url FROM site_plan_pages spp
JOIN site_plans sp ON sp.id=spp.plan_id
WHERE sp.site_id='<SITE>' AND sp.is_current=true ORDER BY spp.name;
```
A `pages` row absent from that list and never deployed is an orphan — archive it, do not
build it. (2026-07-29: `shop`, `brands`, `guides` were exactly this, and all three were
flagged `in_header`.)

## 6. Work-item state and why `detected` items sit still

```sql
SELECT item_type, status, pipeline, handler_agent, count(*)
FROM site_work_items WHERE site_id='<SITE>'
  AND status NOT IN ('complete','cancelled','rejected','verified')
GROUP BY 1,2,3,4 ORDER BY 1;
```

The dispatch loop needs `status IN ('triaged','approved')` AND `pipeline='build'`
(`claim_work_item_action.go:102`, `load_work_item_actions.go:561`). Discovery checks write
`status='detected'`; `triage_detected_items` promotes them — but its only drivers are
`improvement-loop` / `site-review-agent` / `design-audit-agent`, and
`scheduled_tasks.improvement-sweep` has been `enabled=false` since **2026-05-02**.

**Do NOT "fix" this by changing a check's emitted status.** `bugs_open/083` owns the
question, has measured that routing is not the bottleneck (325 unread items in
`needs_human_review`, oldest 2026-03-15), and carries an explicit
*"Decision pending — do not act on this section"*. Site-scoped promotion is fine:

```sql
UPDATE site_work_items
SET status='triaged', triaged_at=now(),
    spec = jsonb_set(COALESCE(spec,'{}'::jsonb),'{original_pipeline}',to_jsonb(pipeline)),
    pipeline='build'
WHERE site_id='<SITE>' AND item_type='<one type>' AND status='detected';
```
Scope it to ONE item_type per statement — a blanket site-wide promotion sweeps types that
each need their own spend decision (`evaluate_tools`, `empty_section`, `capability_gap`).

## 7. Rebuilding one page (the shape proven on THIS site)

`needs_page` items with this exact shape completed end-to-end here (the `sale` builds,
2026-07-27/28):
```sql
INSERT INTO site_work_items
  (site_id, item_type, item_key, status, pipeline, priority, handler_agent, source, spec, created_by)
VALUES ('<SITE>','needs_page','<source>:<page>:<SITE>','triaged','build',50,
        'page-build-handler','<source>',
        jsonb_build_object('reason','<why>','plan_id','<current plan id>',
                           'page_name','<name>','page_role','<role>'),
        '<who>');
```
`build-pipeline-trigger` (120 s) claims it. `page_role` must match the plan's `role`
(`content`, `landing`, `blog-post`, `section-index`, `tool`, `entity-page`).

## 8. Contact details — two shapes, both needed

`sync_site_identity_action.go:103-110` writes/reads NESTED `identity.contact.email`;
`validate_page_content.go:1281-1288` reads FLAT `identity.data->>'email'` and
`->>'contact_email'`. Write both. `sites.email` wins over either (first in the COALESCE)
and is already correct here.

## 9. Darts RSS feeds — probe results 2026-07-29 (re-probe before trusting)

Every candidate fetched, parsed and recency-checked BEFORE any insert. Half the
plausible-looking list failed, which is the whole reason for the step — an inserted
dead feed just accrues `content_sources.error_count` and eventually trips
`all_sources_erroring`.

| feed | result |
|---|---|
| `https://dartsworld.com/feed/` | **USE** — 200, 10 items, newest same-day |
| `https://www.pdc.tv/rss.xml` | **USE** — 200, 20 items, newest same-day. The official PDC feed |
| `https://news.google.com/rss/search?q=darts+PDC&hl=en-GB&gl=GB&ceid=GB:en` | **USE as a wide net** — 200, 102 items. Aggregator, so lower credibility than the two above; expect triage to reject more of it |
| `https://www.live-darts.com/feed/` | REJECT — HTTP 403 (blocks our fetcher) |
| `https://www.dartsnews.com/feed/` | REJECT — HTTP 404 |
| `https://www.dartsorakel.com/rss` | REJECT — HTTP 404 |
| `https://www.skysports.com/rss/12040` | REJECT — 200 and 20 items, but it is the GENERIC Sky news feed: item titles empty, three occurrences of "darts" in the whole document. **A 200 with items is not evidence the feed is on-topic** — read the titles |

Probe command (counts `<item>` properly; `grep -c "<item"` miscounts against
`<itunes:*>` and self-closing forms):
```bash
curl -s -L --max-time 20 -o /tmp/f.xml -w "%{http_code}\n" "$URL"
grep -o "<item>" /tmp/f.xml | wc -l
grep -o "<pubDate>[^<]*" /tmp/f.xml | head -1
grep -o "<title>[^<]*" /tmp/f.xml | head -6      # confirm the SUBJECT, not just the count
```

## 10. Seeding order — the all-or-nothing trap

`seed_content_sources_action.go:92-111` skips seeding **entirely** if ANY active
`content_sources` row already exists for the site, and `:216-227` deliberately never
seeds `rss` or `scrape` ("requires manual URL config"). Put together:

**Insert the curated RSS rows BEFORE the first orchestrator run and the site never gets
its `news_search`/`api_news` sources at all.**

Correct order:
1. Write `classification.content_features.news_feed` (done 2026-07-29).
2. WAIT for a `content-feed-refresh` tick (6-hourly; check
   `SELECT last_completed_at FROM scheduled_tasks WHERE name='content-feed-refresh'`).
   That run seeds the search/api sources from `vertical_keywords`.
3. THEN insert the verified RSS rows from §9.
4. Two-pass delay: triage scores the PREVIOUS run's items, so nothing appears on the
   site until roughly two ticks (~12h) after arming. Not a failure.
