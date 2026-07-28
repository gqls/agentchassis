# RUNBOOK — `bugs_open/100` + `101`

Commands that were hard to get right, with the gotcha attached.

## Re-ground 101 (are the keys still inert in live definitions?)

```sql
SELECT ad.type, e.k AS step,
       (v->'config' ? 'max_pages') AS max_pages, (v->'config' ? 'follow_links') AS follow_links,
       (v->'config' ? 'extract_mode') AS extract_mode, (v->'config' ? 'fallback_url_field') AS fallback_url
FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v)
WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
  AND v->>'action' = 'scrape_web'
  AND (v->'config' ?| array['max_pages','follow_links','extract_mode','fallback_url_field']);
```

**Gotcha:** the three `deleted_at / is_snapshot / is_active` predicates are all load-bearing.
Without them the same query returns snapshot rows and reads as a much larger blast radius.

## Re-ground 100 — and the column that actually discriminates

```sql
SELECT count(*) AS total,
       count(*) FILTER (WHERE COALESCE(source_url,'')<>'') AS has_source_url,
       count(*) FILTER (WHERE raw_data ? 'source_url')     AS llm_claimed,
       max(collected_at) AS newest
FROM business_intel.data_observations;
```

**Gotcha:** `has_source_url` going non-zero is **not** the pass condition. `llm_claimed`
must stay **0**. If both rise, the fix is the rejected candidate 4 (ask the model for its
own provenance) and must be reverted — a populated column proves the column was written,
never *by what*.

## Is a config key actually read by any Go code?

```bash
grep -rn --include=*.go "<key>" . | grep -v '^./docs/'
```

**Gotcha — this is the recorded landmine, do not narrow it.** 101's own diagnostic was
`grep -rn "max_pages" ... | grep -i webscrape`, which is true about `webscrape` and false
about the fleet: `max_pages` is live and load-bearing for `select_representative_content`
and `validate_site_plan`. Run it unfiltered, then decide per (action, key) — never per key.
`docs/` is excluded because it holds vendored copies.

## Size the config surface before promising any fleet-wide validation

```sql
WITH steps AS (
  SELECT e.k AS step, v->>'action' AS action, v->'config' AS cfg
  FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v)
  WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
    AND v->'config' IS NOT NULL AND jsonb_typeof(v->'config')='object'
)
SELECT count(DISTINCT action) AS actions, count(*) AS steps,
       (SELECT count(*) FROM (SELECT DISTINCT s.action, ck.key FROM steps s, jsonb_object_keys(s.cfg) AS ck(key)) x) AS pairs
FROM steps;
```

**Gotcha:** `jsonb_typeof(v->'config')='object'` is required — some steps carry a *string*
config (a reference, not a literal; see the model-directory landmine), and
`jsonb_object_keys` errors on those rather than skipping them.

## Audit which live config keys no action declares

```bash
./scripts/audit-config-keys.sh          # undeclared (action,key) pairs, fleet-wide
```

Reads the live DB and compares against the declared allow-lists compiled into the binary
(`--json` for machine output). An (action, key) pair listed here is either a real inert key
or an action that has not opted in yet — the report does not distinguish, by design.

## Verify P1 without calling Firecrawl

```bash
go test ./internal/adapters/webscrape/providers/ -run TestScrapePayload -v
```

The assertion is on the **payload**, not on a live response: `only_main_content:false` must
produce a payload where `onlyMainContent` is **present and false**. Asserting on scraped
content instead would need Firecrawl and would still not distinguish "we sent false" from
"Firecrawl happened to keep the footer".
