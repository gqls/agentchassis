# RUNBOOK — cta_dials_phone (bugs_open/299 slug + bugs_open/312)

Commands that were hard to get right, with their gotchas attached.

## Verify the bug on the served page (the false-pass trap)

```bash
curl -s https://preview.webdesign.uk/index.html | grep -o '<a[^>]*>[^<]*Brief Starter[^<]*</a>'
```
**Gotcha:** nav and footer link the tool CORRECTLY, so a page-wide grep for the correct URL
passes while the button stays broken. Assert on the anchor whose TEXT names a destination,
never on the URL's presence anywhere in the page. After the 08-18 rewrite the broken anchor's
text is "See how it works" — grep the cta-section block instead:
```bash
curl -s https://preview.webdesign.uk/index.html | grep -A3 'cta-btn-secondary'
```

## The stored pair (label vs url), per site

```sql
SELECT p.name, pc.slot_name, pc.updated_at,
       pc.content_data->>'secondary_cta' AS s_txt, pc.content_data->>'secondary_cta_url' AS s_url
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='webdesign.uk' AND pc.slot_name IN ('call-to-action','hero')
ORDER BY p.name, pc.position;
```

## Fleet census of CTA url scopes (sizes the class)

```sql
SELECT CASE WHEN v.url LIKE 'tel:%' THEN 'tel' WHEN v.url LIKE 'mailto:%' THEN 'mailto'
            WHEN v.url LIKE 'http%' OR v.url LIKE '//%' THEN 'external'
            WHEN v.url LIKE '#%' THEN 'anchor' WHEN v.url = '' THEN 'empty' ELSE 'page' END AS scope,
       count(*), count(DISTINCT s.domain)
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id,
LATERAL (VALUES (pc.content_data->>'cta_url'),(pc.content_data->>'primary_cta_url'),
                (pc.content_data->>'secondary_cta_url')) AS v(url)
WHERE v.url IS NOT NULL AND p.status='active' GROUP BY 1 ORDER BY 2 DESC;
-- 2026-08-18: page 1006 | empty 27 | tel 5 | external 2 | anchor 1 | mailto 1
```

## The 312 wiring proof (resolver output vs what rendered)

Find a page-content-writer orchestration and compare the two sides IN ITS OWN collected_data:

```sql
SELECT jsonb_path_query_array(collected_data->'resolved_links'->'response',
         '$.sections_ready[*] ? (@.name == "call-to-action").resolved_data') AS resolver_wrote,
       jsonb_path_query_array(collected_data->'sections_for_render',
         '$.sections_ready[*] ? (@.name == "call-to-action").resolved_data') AS render_used
FROM orchestration_states WHERE orchestration_id='05e3839d-8e18-4935-9c7e-3c6d741665d6';
```
**Gotchas:** (1) `left(jsonb,n)` does not exist — cast `(x)::text` first. (2) The parent holds
the child's result under BOTH `resolve_links` and `resolved_links`, each as `{response: …}`;
the config path reads `resolved_links`. (3) Retention: `resolved_links` rows go back only to
08-17 — do not claim "never in history", claim the 0/150 window.

The negative/positive control pair (path 1 has never matched; the real shape is normal):

```sql
SELECT count(*) AS runs,
       count(*) FILTER (WHERE collected_data->'resolved_links'->'response' ? 'link_resolution') AS path1_would_hit,
       count(*) FILTER (WHERE collected_data->'resolved_links'->'response' ? 'sections_ready') AS real_shape
FROM orchestration_states WHERE collected_data ? 'resolved_links';
-- 2026-08-18: 150 | 0 | 149
```

## The live select_sections config (the defective path)

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'select_sections')
FROM agent_definitions WHERE type='page-content-writer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Does a value reach the writer prompt? (the measurement trap)

```sql
-- WRONG (what I did first): counts the guidance SENTENCE, which contains the field name
SELECT count(*) FROM llm_call_log WHERE prompt_rendered LIKE '%_target_title%';
-- RIGHT: separate the phrase from a value-shaped occurrence
SELECT count(*) FILTER (WHERE prompt_rendered LIKE '%e.g. cta_target_title for cta_url%') AS guidance_text,
       count(*) FILTER (WHERE prompt_rendered ~ '(primary|secondary|cta)_target_title"?\s*[:=]') AS value_shaped
FROM llm_call_log WHERE created_at > now() - interval '36 hours' AND prompt_rendered LIKE '%_target_title%';
-- 2026-08-18: 179 guidance / 0 value-shaped (of 182)
```

## Ownership / collision checks before touching CTA machinery

```bash
./scripts/who-owns.py 248     # the page-scheme keep half — bugfix_248_authored_cta_destinations, ACTIVE
./scripts/who-owns.py 299     # this bug (by slug; number is ambiguous)
grep -n "cta_links_stale" docs/agent_docs/docs024_key_docs_latest/LANDMINES.md
```
**Gotcha:** `who-owns` reads COMMITS — a session mid-fix is invisible. `ListAgents` + live
`.jsonl` transcripts are the uncommitted check. The `bugfix 248` peer session exists.
