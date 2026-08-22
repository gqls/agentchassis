# RUNBOOK — bugfix 114 imagery wiring

Every query/command that was hard to get right, with its gotcha attached. Change it
HERE when it changes.

DB prefix throughout:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## The one query that made the bug legible

Content-hero assets per site against components actually wired to one. This is the
class census — run it before and after any fix.

```sql
SELECT s.domain,
       count(DISTINCT a.asset_key) AS content_hero_assets,
       count(DISTINCT CASE WHEN pc.content_data->>'hero_url' LIKE '%content-hero-%'
                           THEN pc.id END) AS wired_components
FROM assets a
JOIN sites s ON s.id = a.site_id
LEFT JOIN pages p ON p.site_id = s.id
LEFT JOIN page_components pc ON pc.page_id = p.id
WHERE a.purpose = 'content_hero' AND a.status = 'active'
GROUP BY 1 ORDER BY 2 DESC;
```

**Gotcha:** `wired_components` counts components on the site, not components on the
page the asset was made for. It is a *class* signal, not a per-asset verdict. For a
per-asset verdict join on the key convention (below).

## Per-page verdict — asset to component, by the resolver's own key rule

Mirrors `imageryplan.ContentHeroKey` exactly: `content_hero_` + page name with `-`→`_`.

```sql
SELECT p.name,
       a.created_at AS asset_created,
       a.status     AS asset_status,
       pc.updated_at AS component_rendered,
       round(EXTRACT(EPOCH FROM (pc.updated_at - a.created_at))) AS secs_after_asset,
       CASE WHEN pc.content_data->>'hero_url' LIKE '%content-hero-%'
            THEN 'WIRED' ELSE 'fallback' END AS outcome
FROM pages p
JOIN page_components pc ON pc.page_id = p.id
JOIN content_components cc ON cc.id = pc.component_id AND cc.name = 'hero'
JOIN assets a ON a.site_id = p.site_id
             AND a.asset_key = 'content_hero_' || replace(p.name, '-', '_')
WHERE p.site_id = '<site>' AND p.page_type = 'tool'
ORDER BY pc.updated_at;
```

**Gotcha:** `secs_after_asset` is the column that killed the race theory — do not drop
it. Without it "the asset landed late" stays a live hypothesis for ever.

## Ruling the plan routes in or out

`ensureAssets` tries plan page-scope hero (route 1), then Lane B by key (route 2), then
plan site-scope hero (route 3), then section imagery, then `sites.content_data`. Before
blaming route 2, prove 1 and 3 are empty:

```sql
SELECT spi.scope, spi.scope_ref, spi.key, spi.kind
FROM site_plan_imagery spi
JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current
WHERE sp.site_id = '<site>';
```

**Gotcha:** the `is_current` join is load-bearing. Without it you read superseded plans
and conclude a route was available when it was not.

## Entity-link census

```sql
SELECT (entity_type IS NULL) AS null_link, count(*),
       min(created_at)::date, max(created_at)::date
FROM assets GROUP BY 1;

SELECT purpose, (entity_type IS NULL) AS null_link, count(*)
FROM assets WHERE created_at > now() - interval '14 days'
GROUP BY 1,2 ORDER BY 3 DESC;
```

**Gotcha:** the all-time split understates it. The 14-day cut is what shows that *only*
`card` gets a link — i.e. that this is current behaviour, not history.

## The poisoned site-wide default

```sql
SELECT s.domain, s.content_data->>'hero_url'
FROM sites s WHERE s.content_data->>'hero_url' IS NOT NULL ORDER BY 1;
```

Then probe what it actually serves — the DB value is not the artefact:

```bash
for d in <domains>; do
  printf '%s %s\n' "$(curl -s -o /dev/null -m 8 -w '%{http_code}' "https://$d/assets/images/hero.jpg")" "$d"
done
```

**Gotcha:** a 200 here does not mean the site is fine — it means the generic hero exists.
The bug is the page showing the generic one while its own image sits unreferenced. Check
what the page references too:

```bash
curl -s -m 10 "https://<domain>/" | grep -oE "(src=\"[^\"]*hero[^\"]*\"|url\([^)]*hero[^)]*\))" | sort | uniq -c
```

**Gotcha:** heroes are frequently CSS `background-image`, invisible to an `<img src=`
census. The `url(...)` alternative above is not optional.

## Discovery rotation health (why convergence never happened)

```sql
SELECT agent_type, count(*), min(last_selected_at)::date, max(last_selected_at)::date
FROM site_discovery_rotation GROUP BY 1 ORDER BY 1;
```

**Gotcha:** the column is `last_selected_at`. `last_run_at` and `swept_at` do not exist;
guessing them costs a round trip each. The table is `(site_id, agent_type,
last_selected_at)` and nothing else.

The daily watcher's own words:

```sql
SELECT created_at::date, substr(body,1,400) FROM doc_notes
WHERE subject_key = 'site-discovery-staleness' ORDER BY created_at DESC LIMIT 3;
```

## Work-item queue state

```sql
SELECT status, count(*) FROM site_work_items
WHERE spec->>'reason' = 'image_landed' GROUP BY 1;

SELECT item_type, handler_agent, status, substr(summary,1,60),
       result->'revalidation'->>'verdict', result->'revalidation'->>'reason'
FROM site_work_items
WHERE spec->>'reason' = 'image_landed' AND status = 'needs_human_review'
ORDER BY created_at;
```

**Gotcha:** `site_work_items` has **no `attempts` column** (it is `attempt_count`), and
`orchestration_states` has **no `id` or `agent_type`** — the keys are
`orchestration_id`, `correlation_id`, `owner_agent_type`, `site_id`. Three separate
round trips were lost to this; `\d <table>` first, as CLAUDE.md says.

## Firing a diagnosis run

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh "<symptom>"
```

**⚠ Gotcha, cost one wasted run:** the symptom is interpolated into a `$json$`-quoted
SQL literal **unescaped**. A double quote anywhere in the symptom — e.g. writing
`assets["hero"]` to name a map key — aborts the intake with
`ERROR: invalid input syntax for type json … Token "hero" is invalid`. **Write symptoms
with no double quotes at all**; name map keys in prose instead.

Then take the **run** correlation, not the intake one:

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<RUN_CORRELATION_ID>';
```

Budget ~30 minutes: the council/diagnosis itself takes 2–5, the dispatch queues behind
the fleet. A missing row is latency, not a dropped dispatch — do not retry on it.
