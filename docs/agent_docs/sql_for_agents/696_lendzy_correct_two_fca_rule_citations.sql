-- 696_lendzy_correct_two_fca_rule_citations.sql
--
-- OWNER DECISION 2026-09-02 ("please fix both"): correct the two wrong FCA rule
-- citations in lendzy's stored copy, at every layer that stores them, and drive
-- the re-renders that put the correction on the wire.
--
-- THE TWO ERRORS (evidence: lendzy lane NOTES (g); every quote verified against
-- the handbook text on 2026-09-02):
--   * The two-rollover limit is attributed to CONC 6.7.17. That is the
--     DEFINITIONS rule ("In CONC 6.7.18 R to CONC 6.7.23 R 'refinance' means to
--     extend..."). The limit is CONC 6.7.23: "A firm must not refinance
--     high-cost short-term credit (other than by exercising forbearance) on
--     more than two occasions."
--   * The two-attempt continuous-payment-authority limit is attributed to
--     CONC 6.7.23 — the refinance rule above. It is CONC 7.6.12: "...if it has
--     done so ... on two previous occasions and those previous payment requests
--     have been refused."
-- Both SUBSTANTIVE claims are true. Only the rule numbers move.
--
-- SCOPE, FROM A FLEET CENSUS OF ALL FOUR STORAGE SURFACES [MEASURED 2026-09-02]
-- (page_components.content_data, page_components.rendered_html,
--  site_components.rendered_html, site_specs current, content_components
--  html_template):
--   lendzy.co.uk: 10x 'CONC 6.7.23' (all CPA-context — every occurrence's
--     surrounding text was read, none assumed) across 9 content_data rows and
--     their rendered mirrors; 6x 'CONC 6.7.17' (all rollover-context) across 4
--     content_data rows, 7 in rendered mirrors (the tool page adds one); the
--     content_direction SPEC carries 2x the CPA error — the writing instructions
--     themselves, i.e. the source that would RE-PLANT the error on any future
--     rewrite (the bugs_closed/414 lesson: fix the spec, not just the output);
--     and the tool template 1fbbd1da (rollover checker) carries 1.
--   loancash.co.uk: the SAME error propagated via a tool fork — component
--     6525121a (tool-rollover-limit-checker-lendzy-co-uk-loancash-co-uk,
--     1x 6.7.17) and its page's rendered_html (1x), serving today at
--     https://loancash.co.uk/tools/rollover-limit-checker/index.html
--     [MEASURED: http grep 6.7.17=1]. Included because it is OUR sentence,
--     propagated by OUR fork mechanism, and leaving a known-wrong legal citation
--     serving on a sister site while fixing the original is not defensible.
--   Nothing else fleet-wide stores either token.
--
-- EVERY occurrence is in the exact form 'CONC 6.7.17' / 'CONC 6.7.23' (prefix
-- counts equal bare counts on every surface), so those are the replacement
-- tokens; a bare '6.7.23' is never touched. The tool templates carry ~51 '{{'
-- bindings — a literal token replace cannot touch a binding, which is why this
-- is a replace and not a rewrite. No LLM rewrite is used anywhere: the change
-- is two verbatim token substitutions, and an LLM rewrite risks prose drift the
-- owner did not authorise.
--
-- ORDER MATTERS AND IS LOAD-BEARING: the CPA replace (6.7.23 -> 7.6.12) runs
-- FIRST, while the census guarantees every stored 'CONC 6.7.23' is the CPA
-- error; the rollover replace (6.7.17 -> 6.7.23) runs SECOND and mints the
-- legitimate 6.7.23s. Reversed, the rollover fix would be immediately destroyed
-- by the CPA replace.
--
-- rendered_html is corrected in the same pass as content_data: serving only
-- updates on deploy, so this alone changes nothing on the wire — but it means a
-- carry path (a rerender that fails and ships stored HTML, the bugs 260 class)
-- carries the CORRECTED bytes, not the wrong ones.
--
-- APPROVED round 1 (corr bb352ee8, 2026-09-02 15:34Z) with advisories, ACTED ON:
--   * editquality (version-pin concern): VERIFIED INAPPOSITE at the code — the
--     renderer loads html_template LIVE from content_components
--     (loadContentComponentsByID, v3_site_actions.go:5247-5252, no
--     component_versions join); page_components.component_version_id is a
--     PROVENANCE stamp carried through saves (RFC_046), not a render source. So
--     the template edit is what a re-render renders, and a rerender cannot
--     resurrect the stale citation from a snapshot. (The cited "ships NOTHING"
--     landmine is about a deployment overlay, not SQL template edits.)
--   * render_guardian (unrecognised rerender reason -> assemble-mode): CORRECT,
--     and safe BY CONSTRUCTION here — assemble-mode re-embeds stored
--     rendered_html, which this migration corrects in the same transaction, so
--     either rerender mode ships corrected bytes.
--   * debug_historian (no physical backup): backup table added below
--     (bak_696_citation_surgery — every touched page_components row's prior
--     content_data/rendered_html, kept like bak_670/bak_farmer_cull).
--   * debug_historian (artefact acceptance not committed as a step): the
--     POST-APPLY CHECK below is that step. Run it after the 11 rerenders
--     complete; do not report this migration done without it.
--
-- APPLY: psql -v ON_ERROR_STOP=1 -f <this file>  (single wrapped transaction;
-- an aborted apply leaves nothing half-done).
--
-- POST-APPLY CHECK (the artefact is the acceptance, DB counts are not):
--   for u in rollover-rules.html cant-pay.html check-your-loan.html \
--            continuous-payment-authority.html how-to-complain.html \
--            your-rights.html index.html; do
--     curl -s "https://lendzy.co.uk/$u" | grep -c 'CONC 6\.7\.17'   # every one: 0
--   done
--   curl -s https://lendzy.co.uk/rollover-rules.html | grep -c 'CONC 6\.7\.23'  # >=1
--   curl -s https://lendzy.co.uk/cant-pay.html       | grep -c 'CONC 7\.6\.12'  # >=1
--   curl -s https://lendzy.co.uk/tools/rollover-limit-checker/index.html | grep -c 'CONC 6\.7\.23'   # 1, and inputs unchanged (8)
--   curl -s https://loancash.co.uk/tools/rollover-limit-checker/index.html | grep -c 'CONC 6\.7\.23' # 1
--   plus the invented-URL 404 control per RUNBOOK 1.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/lendzy_co_uk/ (NOTES (g), PLAN A/B)
-- Rollback: 696_..._ROLLBACK.sql (enumerated, because after this migration a
-- blanket reverse replace could not tell a minted 6.7.23 from a CPA one)

BEGIN;

-- ---------------------------------------------------------------------------
-- GUARDS: pin the exact pre-state the census measured. Any drift aborts.
-- ---------------------------------------------------------------------------
DO $$
DECLARE c int; r int;
BEGIN
  -- G1: lendzy content_data rollover error: 6 occurrences across 4 rows
  SELECT COALESCE(sum((length(pc.content_data::text)-length(replace(pc.content_data::text,'CONC 6.7.17','')))/11),0), count(*)
    INTO c, r
    FROM pages p JOIN page_components pc ON pc.page_id=p.id
   WHERE p.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76' AND pc.content_data::text LIKE '%CONC 6.7.17%';
  IF c <> 6 OR r <> 4 THEN RAISE EXCEPTION '696 ABORT G1: lendzy content_data CONC 6.7.17 = % across % rows (expected 6/4)', c, r; END IF;

  -- G2: lendzy content_data CPA error: 10 occurrences across 9 rows
  SELECT COALESCE(sum((length(pc.content_data::text)-length(replace(pc.content_data::text,'CONC 6.7.23','')))/11),0), count(*)
    INTO c, r
    FROM pages p JOIN page_components pc ON pc.page_id=p.id
   WHERE p.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76' AND pc.content_data::text LIKE '%CONC 6.7.23%';
  IF c <> 10 OR r <> 9 THEN RAISE EXCEPTION '696 ABORT G2: lendzy content_data CONC 6.7.23 = % across % rows (expected 10/9)', c, r; END IF;

  -- G3: lendzy rendered_html: 7x 6.7.17 (5 rows), 10x 6.7.23 (9 rows)
  SELECT COALESCE(sum((length(pc.rendered_html)-length(replace(pc.rendered_html,'CONC 6.7.17','')))/11),0), count(*)
    INTO c, r
    FROM pages p JOIN page_components pc ON pc.page_id=p.id
   WHERE p.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76' AND pc.rendered_html LIKE '%CONC 6.7.17%';
  IF c <> 7 OR r <> 5 THEN RAISE EXCEPTION '696 ABORT G3a: lendzy rendered_html CONC 6.7.17 = % across % rows (expected 7/5)', c, r; END IF;
  SELECT COALESCE(sum((length(pc.rendered_html)-length(replace(pc.rendered_html,'CONC 6.7.23','')))/11),0), count(*)
    INTO c, r
    FROM pages p JOIN page_components pc ON pc.page_id=p.id
   WHERE p.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76' AND pc.rendered_html LIKE '%CONC 6.7.23%';
  IF c <> 10 OR r <> 9 THEN RAISE EXCEPTION '696 ABORT G3b: lendzy rendered_html CONC 6.7.23 = % across % rows (expected 10/9)', c, r; END IF;

  -- G4: the spec: exactly one current content_direction row carrying exactly 2
  SELECT COALESCE(sum((length(ss.data::text)-length(replace(ss.data::text,'CONC 6.7.23','')))/11),0), count(*)
    INTO c, r
    FROM site_specs ss
   WHERE ss.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76' AND ss.aspect='content_direction' AND ss.is_current;
  IF c <> 2 OR r <> 1 THEN RAISE EXCEPTION '696 ABORT G4: content_direction CONC 6.7.23 = % across % current rows (expected 2/1)', c, r; END IF;

  -- G5: the two tool templates, one occurrence each
  SELECT count(*) INTO r FROM content_components cc
   WHERE cc.id IN ('1fbbd1da-a467-468d-99dd-7e56cfeb78d9','6525121a-1d06-44b5-bd16-e551d45167b2')
     AND (length(cc.html_template)-length(replace(cc.html_template,'CONC 6.7.17','')))/11 = 1;
  IF r <> 2 THEN RAISE EXCEPTION '696 ABORT G5: expected both tool templates to carry exactly 1x CONC 6.7.17, matching rows = %', r; END IF;

  -- G6: loancash rendered_html: exactly the one tool page row, 1 occurrence
  SELECT COALESCE(sum((length(pc.rendered_html)-length(replace(pc.rendered_html,'CONC 6.7.17','')))/11),0), count(*)
    INTO c, r
    FROM pages p JOIN page_components pc ON pc.page_id=p.id
   WHERE p.site_id='ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND pc.rendered_html LIKE '%CONC 6.7.17%';
  IF c <> 1 OR r <> 1 THEN RAISE EXCEPTION '696 ABORT G6: loancash rendered_html CONC 6.7.17 = % across % rows (expected 1/1)', c, r; END IF;

  -- G7: COMPLETENESS — nothing else, anywhere, stores either token. A row born
  -- since the census would silently survive a scoped fix; abort instead.
  SELECT count(*) INTO r FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE (pc.content_data::text LIKE '%CONC 6.7.17%' OR pc.content_data::text LIKE '%CONC 6.7.23%'
       OR pc.rendered_html    LIKE '%CONC 6.7.17%' OR pc.rendered_html    LIKE '%CONC 6.7.23%')
     AND p.site_id NOT IN ('8ff093d5-1f19-453b-9439-a10379bbcd76','ee4a8199-4f5b-4e2e-88ce-01e600721b74');
  IF r <> 0 THEN RAISE EXCEPTION '696 ABORT G7a: % page_components row(s) OUTSIDE the two censused sites carry a token — re-census', r; END IF;
  SELECT count(*) INTO r FROM site_components sc
   WHERE sc.rendered_html LIKE '%CONC 6.7.17%' OR sc.rendered_html LIKE '%CONC 6.7.23%';
  IF r <> 0 THEN RAISE EXCEPTION '696 ABORT G7b: % site_components row(s) carry a token — the census said zero', r; END IF;
  SELECT count(*) INTO r FROM content_components cc
   WHERE (cc.html_template LIKE '%CONC 6.7.17%' OR cc.html_template LIKE '%CONC 6.7.23%')
     AND cc.id NOT IN ('1fbbd1da-a467-468d-99dd-7e56cfeb78d9','6525121a-1d06-44b5-bd16-e551d45167b2');
  IF r <> 0 THEN RAISE EXCEPTION '696 ABORT G7c: % other component template(s) carry a token — the census said two', r; END IF;
  SELECT count(*) INTO r FROM site_specs ss
   WHERE ss.is_current AND (ss.data::text LIKE '%CONC 6.7.17%' OR ss.data::text LIKE '%CONC 6.7.23%')
     AND NOT (ss.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76' AND ss.aspect='content_direction');
  IF r <> 0 THEN RAISE EXCEPTION '696 ABORT G7d: % other current spec(s) carry a token — the census said one', r; END IF;
END $$;

-- ---------------------------------------------------------------------------
-- 0. PHYSICAL BACKUP of every row about to be touched (debug_historian
-- advisory). Kept after success, like bak_670 / bak_farmer_cull_content_data.
-- ---------------------------------------------------------------------------
CREATE TABLE bak_696_citation_surgery AS
SELECT now() AS backed_up_at, pc.id AS page_component_id, p.name AS page_name,
       s.domain, pc.content_data, pc.rendered_html
  FROM pages p JOIN sites s ON s.id=p.site_id JOIN page_components pc ON pc.page_id=p.id
 WHERE pc.content_data::text LIKE '%CONC 6.7.17%' OR pc.content_data::text LIKE '%CONC 6.7.23%'
    OR pc.rendered_html    LIKE '%CONC 6.7.17%' OR pc.rendered_html    LIKE '%CONC 6.7.23%';

CREATE TABLE bak_696_component_templates AS
SELECT now() AS backed_up_at, id, name, html_template
  FROM content_components
 WHERE id IN ('1fbbd1da-a467-468d-99dd-7e56cfeb78d9','6525121a-1d06-44b5-bd16-e551d45167b2');

-- ---------------------------------------------------------------------------
-- 1. CPA FIRST: CONC 6.7.23 -> CONC 7.6.12 (every stored 6.7.23 is the CPA
--    error at this point — that is exactly what G2/G3b/G4/G7 pinned).
-- ---------------------------------------------------------------------------
UPDATE page_components pc
   SET content_data = replace(pc.content_data::text,'CONC 6.7.23','CONC 7.6.12')::jsonb,
       updated_at = NOW()
  FROM pages p
 WHERE pc.page_id=p.id AND p.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76'
   AND pc.content_data::text LIKE '%CONC 6.7.23%';

UPDATE page_components pc
   SET rendered_html = replace(pc.rendered_html,'CONC 6.7.23','CONC 7.6.12'),
       updated_at = NOW()
  FROM pages p
 WHERE pc.page_id=p.id AND p.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76'
   AND pc.rendered_html LIKE '%CONC 6.7.23%';

-- The spec is SUPERSEDED, not edited: history is the point of site_specs. A CTE
-- so the INSERT can only ever read the row this statement just flipped — a
-- SELECT over superseded rows could resurrect an OLDER version's data.
WITH old AS (
  UPDATE site_specs SET is_current=false, superseded_at=NOW()
   WHERE site_id='8ff093d5-1f19-453b-9439-a10379bbcd76' AND aspect='content_direction' AND is_current
  RETURNING *
)
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
SELECT old.site_id, old.aspect,
       replace(old.data::text,'CONC 6.7.23','CONC 7.6.12')::jsonb,
       old.source, old.source_agent,
       'lendzy_co_uk lane (migration 696)',
       true, old.pinned,
       'Supersedes '||old.id||': CPA rule citation corrected CONC 6.7.23 -> CONC 7.6.12 (owner decision 2026-09-02). The wrong number in the WRITING INSTRUCTIONS is what would re-plant the error on any rewrite.'
  FROM old;

-- ---------------------------------------------------------------------------
-- 2. ROLLOVER SECOND: CONC 6.7.17 -> CONC 6.7.23 (mints the legitimate 6.7.23s)
-- ---------------------------------------------------------------------------
UPDATE page_components pc
   SET content_data = replace(pc.content_data::text,'CONC 6.7.17','CONC 6.7.23')::jsonb,
       updated_at = NOW()
  FROM pages p
 WHERE pc.page_id=p.id
   AND p.site_id IN ('8ff093d5-1f19-453b-9439-a10379bbcd76','ee4a8199-4f5b-4e2e-88ce-01e600721b74')
   AND pc.content_data::text LIKE '%CONC 6.7.17%';

UPDATE page_components pc
   SET rendered_html = replace(pc.rendered_html,'CONC 6.7.17','CONC 6.7.23'),
       updated_at = NOW()
  FROM pages p
 WHERE pc.page_id=p.id
   AND p.site_id IN ('8ff093d5-1f19-453b-9439-a10379bbcd76','ee4a8199-4f5b-4e2e-88ce-01e600721b74')
   AND pc.rendered_html LIKE '%CONC 6.7.17%';

UPDATE content_components
   SET html_template = replace(html_template,'CONC 6.7.17','CONC 6.7.23'),
       updated_at = NOW()
 WHERE id IN ('1fbbd1da-a467-468d-99dd-7e56cfeb78d9','6525121a-1d06-44b5-bd16-e551d45167b2');

-- ---------------------------------------------------------------------------
-- 3. DRIVE THE RE-RENDERS (the 693 round-2 lesson: a repair that fixes storage
--    and queues nothing satisfies none of its own acceptance criteria).
-- ---------------------------------------------------------------------------
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items wi
   WHERE wi.item_key IN (
     SELECT 'page_rerender_'||p.name||'_'||p.site_id||'_assemble'
       FROM pages p
      WHERE (p.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76' AND p.name IN
              ('rollover-rules','cant-pay','check-your-loan','continuous-payment-authority',
               'how-to-complain','index','your-rights',
               'tool-cpa-cancellation-checker-guide','tool-rollover-limit-checker-guide',
               'tool-rollover-limit-checker'))
         OR (p.site_id='ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND p.name='tool-rollover-limit-checker'))
     AND wi.status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
  IF n <> 0 THEN
    RAISE EXCEPTION '696 ABORT: % live page_rerender item(s) already queued for these pages — do not stack', n;
  END IF;
END $$;

INSERT INTO site_work_items
    (site_id, item_type, item_key, status, handler_agent, priority, source, summary,
     page_id, created_by, spec)
SELECT p.site_id, 'page_rerender',
       'page_rerender_'||p.name||'_'||p.site_id||'_assemble',
       'triaged', 'page-rerender', 80,
       'lendzy_co_uk lane (migration 696)',
       'Rerender page: '||p.name,
       p.id,
       'lendzy_co_uk lane (migration 696)',
       jsonb_build_object('domain', s.domain, 'page_id', p.id::text,
                          'filename', ltrim(p.url,'/'), 'page_name', p.name,
                          'reason', 'FCA rule citation corrected by migration 696 (owner decision 2026-09-02)')
  FROM pages p JOIN sites s ON s.id=p.site_id
 WHERE (p.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76' AND p.name IN
         ('rollover-rules','cant-pay','check-your-loan','continuous-payment-authority',
          'how-to-complain','index','your-rights',
          'tool-cpa-cancellation-checker-guide','tool-rollover-limit-checker-guide',
          'tool-rollover-limit-checker'))
    OR (p.site_id='ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND p.name='tool-rollover-limit-checker');

-- ---------------------------------------------------------------------------
-- VERIFY (DO/RAISE — bare SELECTs cannot stop a COMMIT)
-- ---------------------------------------------------------------------------
DO $$
DECLARE c int; r int;
BEGIN
  -- The wrong token is EXTINCT on every surface, fleet-wide.
  SELECT count(*) INTO r FROM page_components pc
   WHERE pc.content_data::text LIKE '%CONC 6.7.17%' OR pc.rendered_html LIKE '%CONC 6.7.17%';
  IF r <> 0 THEN RAISE EXCEPTION '696 VERIFY: CONC 6.7.17 survives in % page_components row(s)', r; END IF;
  SELECT count(*) INTO r FROM content_components WHERE html_template LIKE '%CONC 6.7.17%';
  IF r <> 0 THEN RAISE EXCEPTION '696 VERIFY: CONC 6.7.17 survives in % template(s)', r; END IF;
  SELECT count(*) INTO r FROM site_specs WHERE is_current AND data::text LIKE '%CONC 6.7.17%';
  IF r <> 0 THEN RAISE EXCEPTION '696 VERIFY: CONC 6.7.17 survives in % current spec(s)', r; END IF;

  -- lendzy content_data: exactly 10x 7.6.12 and exactly 6x 6.7.23 (all rollover)
  SELECT COALESCE(sum((length(pc.content_data::text)-length(replace(pc.content_data::text,'CONC 7.6.12','')))/11),0) INTO c
    FROM pages p JOIN page_components pc ON pc.page_id=p.id
   WHERE p.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76';
  IF c <> 10 THEN RAISE EXCEPTION '696 VERIFY: lendzy content_data CONC 7.6.12 = % (expected 10)', c; END IF;
  SELECT COALESCE(sum((length(pc.content_data::text)-length(replace(pc.content_data::text,'CONC 6.7.23','')))/11),0) INTO c
    FROM pages p JOIN page_components pc ON pc.page_id=p.id
   WHERE p.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76';
  IF c <> 6 THEN RAISE EXCEPTION '696 VERIFY: lendzy content_data CONC 6.7.23 = % (expected 6, all rollover)', c; END IF;

  -- the spec: new current row carries 2x 7.6.12 and 0x 6.7.23
  SELECT count(*) INTO r FROM site_specs ss
   WHERE ss.site_id='8ff093d5-1f19-453b-9439-a10379bbcd76' AND ss.aspect='content_direction' AND ss.is_current
     AND (length(ss.data::text)-length(replace(ss.data::text,'CONC 7.6.12','')))/11 = 2
     AND ss.data::text NOT LIKE '%CONC 6.7.23%'
     AND ss.created_by='lendzy_co_uk lane (migration 696)';
  IF r <> 1 THEN RAISE EXCEPTION '696 VERIFY: corrected content_direction row not found as the single current one', r; END IF;

  -- both tool templates now cite 6.7.23
  SELECT count(*) INTO r FROM content_components cc
   WHERE cc.id IN ('1fbbd1da-a467-468d-99dd-7e56cfeb78d9','6525121a-1d06-44b5-bd16-e551d45167b2')
     AND (length(cc.html_template)-length(replace(cc.html_template,'CONC 6.7.23','')))/11 = 1;
  IF r <> 2 THEN RAISE EXCEPTION '696 VERIFY: expected both tool templates to carry 1x CONC 6.7.23, matching = %', r; END IF;

  -- 11 rerenders queued
  SELECT count(*) INTO r FROM site_work_items
   WHERE source='lendzy_co_uk lane (migration 696)' AND item_type='page_rerender' AND status='triaged'
     AND page_id IS NOT NULL;
  IF r <> 11 THEN RAISE EXCEPTION '696 VERIFY: expected 11 queued rerenders, found %', r; END IF;

  RAISE NOTICE '696 OK: CONC 6.7.17 extinct; 10 CPA cites now 7.6.12; spec superseded; 2 tool templates corrected (lendzy + the loancash fork); 11 rerenders queued';
END $$;

COMMIT;
