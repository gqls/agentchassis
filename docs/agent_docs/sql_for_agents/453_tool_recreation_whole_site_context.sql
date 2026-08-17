-- 453 — tool-recreation-handler: show the analyst the WHOLE site, one line per
-- page (bugs_open/297)
--
-- WHY. `load_related_context` ends `ORDER BY p.nav_order LIMIT 10`. Measured
-- 2026-08-17: 25 sites, **19 of them hold more pages than the cap**; median
-- population 26, worst 107 (webdesign.co.uk). The rows become the "Other Pages
-- on This Site" block of the `analyze_tool` prompt, so the model deciding how a
-- recreated tool fits its site saw 10 of 26 at the median site and 10 of 107 at
-- the worst — and which 10 was an accident of MENU POSITION (`nav_order`), not
-- a judgement of relevance. A row cap feeding an LLM is silent by construction
-- (bugs_open/275's mechanism, register LCO-009).
--
-- WHY THE CAP CAN SIMPLY GO — measured, and it inverts 275's remedy. The prompt
-- renders one short line per row (`- {{.name}} ({{.page_type}}): {{.title}}`).
-- Column extremes across all 727 pages: name <= 66 chars, title <= 144,
-- page_type <= 16 — nothing needs bounding, so no left()/marker machinery.
-- Whole-population rendered block at the WORST site: 8,810 chars (~2.2k tokens)
-- against 735 today, inside a prompt that already embeds the original page's
-- full raw HTML. `rr.summary` is selected and never rendered, and is nearly
-- empty estate-wide (21 of 727 pages, max 48 chars) — kept for shape (275's
-- `category` reasoning: dropping a column a future consumer might read is
-- scope creep for a negligible saving).
--
-- SECOND DEFECT CLOSED IN THE SAME EDIT: the plain LEFT JOIN on
-- research_results has no one-row guarantee and one page ALREADY has two
-- 'adoption_page' rows (page 0747e2fc…, the `index` of site 00ff3af5…,
-- nav_order 1) — today's prompt on that site lists `index` TWICE inside the
-- visible 10. The LATERAL takes the newest research row per page, so a page
-- contributes exactly one line and the result is exactly the site's page
-- population. The inner LIMIT 1 is the fetch-one idiom — outside the
-- silent-cap class by LCO-009's stated design (n=1 excluded; its end-anchored
-- regex ignores subquery LIMITs, both arms vindicated on live cases).
-- Indexed: idx_research_page + idx_research_created.
--
-- NULLS LAST is deliberate. `research_results.created_at` is NULLABLE, and a
-- NULL sorts FIRST under plain DESC — so an untimestamped row would win the
-- "newest" tie and pin a page to it for ever. Measured 2026-08-17: 0 of 21
-- adoption_page rows are NULL today, so it costs nothing now and makes the bad
-- state unreachable rather than merely unlikely.
--
-- Validated read-only before writing this file: proposed text returns 107 rows
-- = population on the worst site, and 40 = population on the fan-out site with
-- the duplicate gone.
--
-- Consumers of load_related_context's output (`related_pages`): the
-- `analyze_tool` LLM step of this same workflow, nothing else — verified in
-- the live config (input_fields + template) 2026-08-17. Row shape unchanged:
-- name, title, page_type, summary.
--
-- Config-only: no image dependency, LIVE ON APPLY. Scoped by id with a
-- TYPE-COUNT pre-state gate (the shape bugs_open/275's council round asked
-- for: an id-scoped UPDATE on a shadowed sibling row silently no-ops, so gate
-- on count(*) by type, not by id).
--
-- APPLY (own file only — the runner's --apply takes every pending file):
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
--     -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/453_tool_recreation_whole_site_context.sql
--   ./scripts/migration/run-migrations.sh --record-only docs/agent_docs/sql_for_agents/453_tool_recreation_whole_site_context.sql --note "..."
--
-- ROLLBACK: 453_tool_recreation_whole_site_context_ROLLBACK.sql, or restore
-- the snapshot this file takes
-- (snapshot_agent note '453_tool_recreation_whole_site_context: pre-update').

BEGIN;

SELECT snapshot_agent('tool-recreation-handler',
  '453_tool_recreation_whole_site_context: pre-update');

-- Pre-state gate: refuse unless (a) exactly ONE live row exists for the type —
-- a second would make the id-scoped UPDATE below a silent no-op on the loaded
-- row — and (b) the step still carries the text this file was written against.
DO $$
DECLARE q text; n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'tool-recreation-handler'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '453: expected exactly 1 live tool-recreation-handler row, found % — an id-scoped UPDATE would silently miss the loaded one', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_related_context,config,query}' INTO q
    FROM agent_definitions
   WHERE type = 'tool-recreation-handler'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF q IS NULL OR position('LIMIT 10' in q) = 0 THEN
    RAISE EXCEPTION '453: pre-state has no LIMIT 10 — someone has already changed this step; re-read it before applying: %', q;
  END IF;
  IF position('LEFT JOIN research_results rr ON' in q) = 0 THEN
    RAISE EXCEPTION '453: pre-state does not carry the plain research_results join this file replaces: %', q;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_related_context,config,query}',
         to_jsonb('SELECT p.name, p.title, p.page_type, rr.summary FROM pages p LEFT JOIN LATERAL (SELECT r.summary FROM research_results r WHERE r.page_id = p.id AND r.result_type = ''adoption_page'' ORDER BY r.created_at DESC NULLS LAST LIMIT 1) rr ON true WHERE p.site_id = $1 AND p.name != $2 ORDER BY p.nav_order'::text)
       ),
       updated_at = now()
 WHERE id = '8701375f-81f7-4d92-ba39-c85f8489dada'
   AND type = 'tool-recreation-handler'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Post-state verify. DO/RAISE, never a bare SELECT: ON_ERROR_STOP ignores a
-- non-empty result, so a verify block of SELECTs cannot stop the COMMIT.
DO $$
DECLARE q text; p jsonb;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_related_context,config,query}',
         default_config#>'{workflow,steps,load_related_context,config,params}'
    INTO q, p
    FROM agent_definitions
   WHERE type = 'tool-recreation-handler'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  -- no MULTI-ROW limit survives anywhere in the text (the LATERAL's LIMIT 1 is
  -- the fetch-one idiom and is the only LIMIT allowed)
  IF q ~* 'LIMIT[[:space:]]+([2-9][0-9]*|1[0-9]+)' THEN
    RAISE EXCEPTION '453: a multi-row LIMIT survives in the query: %', q;
  END IF;
  -- the de-dup is in place
  IF position('LEFT JOIN LATERAL' in q) = 0 OR position('ORDER BY r.created_at DESC NULLS LAST LIMIT 1' in q) = 0 THEN
    RAISE EXCEPTION '453: the one-row-per-page LATERAL is missing — the fan-out door is open: %', q;
  END IF;
  -- presentation order preserved
  IF position('ORDER BY p.nav_order' in q) = 0 THEN
    RAISE EXCEPTION '453: nav_order ordering was lost: %', q;
  END IF;
  -- the two bindings are untouched
  IF p IS NULL OR jsonb_array_length(p) <> 2
     OR p->>0 <> 'site_record.site_id' OR p->>1 <> 'page_record.name' THEN
    RAISE EXCEPTION '453: load_related_context params wrong: %', p;
  END IF;
END $$;

COMMIT;
