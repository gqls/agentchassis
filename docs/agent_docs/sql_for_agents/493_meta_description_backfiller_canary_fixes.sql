-- 493 — meta-description-backfiller: two defects the FIRST CANARY RUN found
--       (bugs_open/320; the agent seeded by 488)
--
-- Both were found by running the thing once on a low-stakes site and reading the
-- result instead of the status. The run "succeeded": orchestration COMPLETED, no
-- error, no failed step. It had written nothing, and it would have gone on writing
-- nothing fleet-wide.
--
-- ── DEFECT 1: THE CONDITIONAL SILENTLY SKIPPED THE ONLY LLM STEP ─────────────
--
-- `load_pages_missing_meta` used `output_format: "array"`, and the gate after it
-- read `pages_missing_meta.count > 0`. Measured on the canary orchestration's own
-- collected_data: `pages_missing_meta` is an ARRAY of length **11** — the rows were
-- there — and the workflow still took `else_step` to `complete_nothing_to_do`.
--
-- `database_actions.go:129-145` is unambiguous about why:
--
--     if outputFormat == "array" { return results, nil }        // a BARE ARRAY
--     // "object" format: include metadata + flatten first row
--     result := map[string]interface{}{"rows": …, "count": len(results), …}
--
-- **`.count` exists only under `output_format: "object"`.** Against a bare array it
-- resolves to nothing, a numeric comparison against nothing is false, and the step
-- routes to else. Silently.
--
-- ⚠ THIS IS `bugs_open/313`, AND I WALKED INTO IT BY COPYING THE AGENT IT WAS FILED
-- AGAINST. `conditional_branch_action.go:54` records it in as many words: "an
-- unresolvable `candidate_pages.count > 0` silently skipped the agent's only LLM
-- step on every run for four months" — `candidate_pages` is `internal-linker`, which
-- is the workflow I modelled 488 on. Copying a live agent copies its bugs.
--
-- A census of live agents shows the same shape in at least eight more conditions
-- (`claimed.count`, `dispatchable.count`, `unswept_areas.count`, `batch.count`, …).
-- Whether each is broken depends on its own step's `output_format`; this migration
-- fixes only THIS agent, and the wider sweep is noted in `bugs_open/313`, not
-- silently widened here.
--
-- THE FIX IS AT THE DATA LEVEL, not at the guard. `output_format` becomes "object",
-- so `.count` genuinely resolves on the binary running today. Consequences carried
-- through in the same migration, because a half-applied rename is worse than the bug:
--   * the loop's `items_field` becomes `pages_missing_meta.rows`;
--   * the prompt's `{{range .pages_missing_meta}}` becomes `{{range .pages_missing_meta.rows}}`.
--
-- `fail_on_non_numeric: true` is ALSO set on the gate, and it is **INERT TODAY** —
-- stated rather than implied, because a config key the binary does not read looks
-- exactly like one it does. Probed on the live chassis v1.0.1315 with controls:
-- `fail_on_non_numeric` ABSENT, `else_step`/`then_step`/`save_page_meta_description`
-- PRESENT, a fake symbol ABSENT. It becomes a real guard on the next roll and turns
-- this whole class from a silent skip into a failed step. Setting it now means
-- nobody has to remember later.
--
-- ── DEFECT 2: THE MODEL WAS BEING SHOWN CSS ─────────────────────────────────
--
-- `content_sample` was `LEFT(string_agg(pc.rendered_html, ' '), 1200)` — raw stored
-- markup. On the canary's own rows that is a faceful of `<style>` blocks: the sample
-- for `tool-overpayment-calculator` begins
-- `<style>.hero-tool-section{--section-text:var(--color-primary-text…` and never
-- reaches a sentence inside 1200 characters.
--
-- So the writer would have been asked to describe a page from its CSS. It would have
-- produced something — that is the danger — and the copy gates cannot catch a
-- fluent, wrong sentence. Now the sample strips `<style>` and `<script>` blocks
-- WITH their contents, then tags, then unescapes the common entities and collapses
-- whitespace, and takes 1200 characters of what is left. The floor moves from
-- `length(raw markup) > 400` to `length(VISIBLE TEXT) > 400`, which is the thing
-- that was meant all along: a page with 20KB of CSS and one heading is not a page
-- with something to describe.
--
-- THE FLOOR ALSO DROPS 400 -> 200, CHOSEN FROM THE DISTRIBUTION RATHER THAN TASTE.
-- Across the 364 empty pages that have any components at all (of 407 empty overall —
-- 43 have no components and can never be described from their own content):
--
--     visible text > 400 chars : 291 pages
--     visible text > 200 chars : 327 pages
--     visible text > 120 chars : 330 pages
--     min 1 · 10th percentile 199 · median 1340
--
-- 200 is the knee: it admits 90% of the population, and going lower buys only three
-- more pages while starting to admit genuinely contentless ones. 400 measured on
-- VISIBLE text would have excluded loanzy.uk's own home page (314 chars) and its
-- glossary (271) — a homepage with no search description is exactly the outcome this
-- lane exists to stop, and the old raw-markup floor hid that behind CSS bulk.
--
-- Scoped by id, pre-state gated, DO/RAISE verify, snapshot first, rollback sidecar.
-- Config-only, LIVE ON APPLY. Not council-submitted: agent_definitions config is
-- outside the gate's platform/ internal/ pkg/ scope.
--
-- ROLLBACK: 493_meta_description_backfiller_canary_fixes_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('meta-description-backfiller',
  '493_canary_fixes: pre-update');

DO $$
DECLARE n int; fmt text; cond text; items text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'meta-description-backfiller'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '493: expected exactly 1 live meta-description-backfiller row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_pages_missing_meta,config,output_format}',
         default_config#>>'{workflow,steps,check_has_pages,config,condition}',
         default_config#>>'{workflow,steps,backfill_loop,config,items_field}'
    INTO fmt, cond, items
    FROM agent_definitions
   WHERE type = 'meta-description-backfiller'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF fmt IS DISTINCT FROM 'array' THEN
    RAISE EXCEPTION '493: output_format is %, expected the defective value ''array'' — already fixed or changed under me', fmt;
  END IF;
  IF cond IS DISTINCT FROM 'pages_missing_meta.count > 0' THEN
    RAISE EXCEPTION '493: the gate condition is %, not the one this migration reasons about', cond;
  END IF;
  IF items IS DISTINCT FROM 'written.result.descriptions' THEN
    RAISE EXCEPTION '493: loop items_field is %, unexpected', items;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config =
         -- 1. the query: strip style/script blocks, then tags, then collapse
         jsonb_set(
         jsonb_set(
         jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,load_pages_missing_meta,config,output_format}',
           to_jsonb('object'::text)
         ),
           '{workflow,steps,load_pages_missing_meta,config,query}',
           to_jsonb(
             'SELECT p.id, p.name, p.url, p.title, p.page_type, ' ||
             'LEFT(regexp_replace(' ||
             '  regexp_replace(' ||
             '    regexp_replace(' ||
             '      regexp_replace(string_agg(pc.rendered_html, '' ''), ''<(style|script)[^>]*>.*?</\1>'', '' '', ''gis''), ' ||
             '    ''<[^>]+>'', '' '', ''g''), ' ||
             '  ''&nbsp;|&amp;|&quot;|&#39;|&lt;|&gt;'', '' '', ''g''), ' ||
             '''\s+'', '' '', ''g''), 1200) AS content_sample ' ||
             'FROM pages p ' ||
             'JOIN page_components pc ON pc.page_id = p.id ' ||
             '  AND pc.rendered_html IS NOT NULL ' ||
             '  AND COALESCE(pc.slot_name, '''') NOT IN (''header'',''footer'',''head'') ' ||
             'WHERE p.site_id = $1 ' ||
             '  AND p.status = ''active'' ' ||
             '  AND COALESCE(p.meta_description, '''') = '''' ' ||
             'GROUP BY p.id, p.name, p.url, p.title, p.page_type ' ||
             -- the floor now measures VISIBLE TEXT, which is what it always meant
             'HAVING length(regexp_replace(' ||
             '  regexp_replace(' ||
             '    regexp_replace(string_agg(pc.rendered_html, '' ''), ''<(style|script)[^>]*>.*?</\1>'', '' '', ''gis''), ' ||
             '  ''<[^>]+>'', '' '', ''g''), ' ||
             '''\s+'', '' '', ''g'')) > 200 ' ||
             'ORDER BY p.name LIMIT 25'
           )
         ),
           -- 2. the loop now reads .rows
           '{workflow,steps,backfill_loop,config,items_field}',
           to_jsonb('written.result.descriptions'::text)  -- unchanged; the LLM output shape did not move
         ),
           -- 3. the gate gains the (currently inert) numeric guard
           '{workflow,steps,check_has_pages,config,fail_on_non_numeric}',
           to_jsonb(true)
         ),
       updated_at = now()
 WHERE type = 'meta-description-backfiller'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- 4. the prompt must iterate .rows now that the step returns an object
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,write_descriptions,config,prompt_template}',
         to_jsonb(
           replace(
             default_config#>>'{workflow,steps,write_descriptions,config,prompt_template}',
             '{{range .pages_missing_meta}}',
             '{{range .pages_missing_meta.rows}}'
           )
         )
       ),
       updated_at = now()
 WHERE type = 'meta-description-backfiller'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE cfg jsonb; q text; p text;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type = 'meta-description-backfiller'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cfg#>>'{workflow,steps,load_pages_missing_meta,config,output_format}' IS DISTINCT FROM 'object' THEN
    RAISE EXCEPTION '493 VERIFY: output_format did not become object';
  END IF;
  IF (cfg#>'{workflow,steps,check_has_pages,config,fail_on_non_numeric}')::text IS DISTINCT FROM 'true' THEN
    RAISE EXCEPTION '493 VERIFY: fail_on_non_numeric was not set';
  END IF;

  q := cfg#>>'{workflow,steps,load_pages_missing_meta,config,query}';
  IF position('<(style|script)' in q) = 0 THEN
    RAISE EXCEPTION '493 VERIFY: the query does not strip style/script blocks';
  END IF;

  p := cfg#>>'{workflow,steps,write_descriptions,config,prompt_template}';
  IF position('{{range .pages_missing_meta.rows}}' in p) = 0 THEN
    RAISE EXCEPTION '493 VERIFY: the prompt still iterates the bare field, which is now an object';
  END IF;
  -- and the OLD form must be gone, or the template renders nothing
  IF position('{{range .pages_missing_meta}}' in p) > 0 THEN
    RAISE EXCEPTION '493 VERIFY: the old range form survives — replace did not fire cleanly';
  END IF;

  -- the safety property from 488, re-asserted: this workflow still cannot overwrite
  IF cfg#>'{workflow,steps,backfill_loop,config,sub_workflow,steps,save_description,config}' ? 'overwrite_existing' THEN
    RAISE EXCEPTION '493 VERIFY: overwrite_existing appeared on the save step';
  END IF;

  RAISE NOTICE '493 OK: object output_format (so .count resolves), visible-text content_sample, prompt iterates .rows, numeric guard armed for the next roll';
END $$;

COMMIT;
