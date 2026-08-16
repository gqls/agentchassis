-- FILE: SQL_2026-08-15c_fundamentallyai_register_tool_count.sql
--
-- Register the interactive-tool count as an evidence_base fact, so the build
-- gate can tell a CORRECTION from a FABRICATION and stop refusing a true number.
--
-- WHY: the nav_drift rebuild that would ship the Tools menu entry died with
--   content validation failed: 0 blockers, 1 errors
--   unregistered_stat  hero-tool.stat_one_value  value "5"
-- The stored and served page both say "3" ("Three interactive tools"). The
-- rebuild REGENERATED the hero and proposed 3 -> 5; the gate refused the write
-- because no fact registers the count. Five is correct — the page has been
-- undercounting itself. The gate was right to refuse: without a registered fact
-- it cannot distinguish a correction from an invention.
--
-- HOW THE GATE MATCHES (read at source, not assumed —
-- datahelpers/claims.go:962 numberSupported + claims_stats.go:86 Window):
--   window   = Label + " " + Value + " " + Detail  ->  "Interactive tools 5"
--   a fact is CONSIDERED only if one of its context_terms is a substring of the
--   lowercased window; then tolerance decides.
-- So context_terms MUST match that window or the fact is skipped silently and
-- nothing changes. "interactive tool" matches "interactive tools 5".
--
-- WHY context_terms IS DELIBERATELY NARROW: a bare "tools" term would also
-- license any other stat on this site whose label happens to contain "tools",
-- turning one registered count into a blanket permission. Scope it to the claim
-- it actually backs.
--
-- WHY tolerance = "exact" AND NOT "gte": with gte, numberSupported passes any
-- val <= 5, so the current understating "3" would validate and the page would
-- keep undercounting for ever. The tool count is precisely knowable, unlike
-- F1-live-sites (gte, because sites launch and the copy states a floor). Exact
-- is what makes the stale "3" fail and the correction land.
--
-- writer_block is regenerated from facts[] by refresh_evidence_base because this
-- site has writer_block_managed: true. The line is appended here as well so the
-- two agree immediately rather than only after the next refresher run; if the
-- generator words it differently, its version wins on the next refresh and that
-- is fine.

BEGIN;

DO $chk$
DECLARE n int; v int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND aspect='evidence_base' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 current evidence_base, found %', n; END IF;

  -- the fact must not already exist, in any form
  SELECT count(*) INTO n FROM site_specs,
       LATERAL jsonb_array_elements(data->'facts') f
   WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND aspect='evidence_base' AND is_current
     AND (f->>'id' = 'F14-interactive-tools'
          OR lower(f->>'claim') LIKE '%interactive tool%');
  IF n <> 0 THEN RAISE EXCEPTION 'an interactive-tool fact already exists (n=%) — read it before adding a second', n; END IF;

  -- the value must still be TRUE at write time, not merely true when authored
  SELECT count(*) INTO v FROM pages
   WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND status='active'
     AND page_type='tool' AND name <> 'tools';
  IF v <> 5 THEN
    RAISE EXCEPTION 'the live tool count is now %, not 5 — do NOT register a stale number; re-derive the fact', v;
  END IF;
END $chk$;

-- supersede, matching this estate's convention: history stays readable
UPDATE site_specs SET is_current=false, superseded_at=now()
 WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND aspect='evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, notes, is_current)
SELECT
  '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd',
  'evidence_base',
  jsonb_set(
    jsonb_set(prev.data, '{facts}', (prev.data->'facts') || jsonb_build_array(
      jsonb_build_object(
        'id','F14-interactive-tools',
        'kind','metric',
        'claim','interactive tools published on this site, free to run in the browser (excludes the /tools index page and the companion guides)',
        'value', 5,
        'source', jsonb_build_object(
          'sql','SELECT count(*) FROM pages WHERE site_id=''199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'' AND status=''active'' AND page_type=''tool'' AND name <> ''tools'''),
        'tolerance','exact',
        'verified_at','2026-08-15',
        'writer_line','five interactive tools, free to run in the browser (live count {value}; an EXACT count — do not round it or state a floor)',
        'context_terms', jsonb_build_array('interactive tool')
      ))),
    '{writer_block}',
    to_jsonb(
      replace(
        prev.data->>'writer_block',
        E'\nCAPABILITIES (assert without inventing numbers):',
        E'\n- five interactive tools, free to run in the browser (an EXACT count — do not round it or state a floor)\n\nCAPABILITIES (assert without inventing numbers):'
      ))
  ),
  'owner 2026-08-15: register the tool count so the build gate stops refusing a true correction (3 -> 5) and the Tools nav entry can ship',
  'brochure_contrast_front_thread',
  'brochure_contrast_front_thread',
  'Adds F14-interactive-tools (value 5, tolerance exact, context_terms ["interactive tool"]). The site published "Three interactive tools" while serving five; a rebuild proposed the correction and validate_page_content refused it as an unregistered_stat. Narrow context_terms deliberately: a bare "tools" term would license any stat labelled with that word. Exact tolerance deliberately: gte would let the stale 3 keep validating.',
  true
FROM (SELECT data FROM site_specs
       WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND aspect='evidence_base'
         AND NOT is_current ORDER BY superseded_at DESC LIMIT 1) prev;

DO $post$
DECLARE n int; wb text;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND aspect='evidence_base' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'post: % current evidence_base rows, expected 1', n; END IF;

  -- the new fact must be present AND the pre-existing ones must have SURVIVED
  SELECT count(*) INTO n FROM site_specs,
       LATERAL jsonb_array_elements(data->'facts') f
   WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND aspect='evidence_base' AND is_current;
  IF n <> 16 THEN RAISE EXCEPTION 'post: % facts, expected 16 (15 existing + 1 new) — an existing fact was LOST', n; END IF;

  SELECT count(*) INTO n FROM site_specs,
       LATERAL jsonb_array_elements(data->'facts') f
   WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND aspect='evidence_base' AND is_current
     AND f->>'id'='F14-interactive-tools' AND (f->>'value')::int = 5
     AND f->>'tolerance'='exact'
     AND f->'context_terms' ? 'interactive tool';
  IF n <> 1 THEN RAISE EXCEPTION 'post: F14 missing or malformed'; END IF;

  -- the writer_block edit must have FIRED, not silently matched nothing
  SELECT data->>'writer_block' INTO wb FROM site_specs
   WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND aspect='evidence_base' AND is_current;
  IF wb NOT LIKE '%five interactive tools%' THEN
    RAISE EXCEPTION 'post: writer_block anchor did not match — the replace() was a silent no-op';
  END IF;
  IF wb NOT LIKE '%CAPABILITIES (assert without inventing numbers):%' THEN
    RAISE EXCEPTION 'post: writer_block CAPABILITIES section lost by the replace()';
  END IF;
END $post$;

SELECT jsonb_array_length(data->'facts') AS facts,
       length(data->>'writer_block') AS writer_block_len, created_by
  FROM site_specs
 WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND aspect='evidence_base' AND is_current;

COMMIT;

-- NEXT: the blocked nav_drift item is at needs_human_review with attempt_count 0.
-- Re-arm it (status back to 'triaged') so the build loop retries the rebuild now
-- that the stat is registered, then VERIFY AT THE SERVED PAGE, not the item:
--   curl -s https://fundamentallyai.com/index.html | grep -c 'href="/tools.html"'
--   curl -s https://fundamentallyai.com/tools.html | grep -o 'Interactive tools[^<]*'
-- Expect the nav link to appear AND the stat to read 5.
