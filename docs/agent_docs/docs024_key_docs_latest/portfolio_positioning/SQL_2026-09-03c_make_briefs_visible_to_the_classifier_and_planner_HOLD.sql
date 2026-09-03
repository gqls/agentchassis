-- 2026-09-03 — make a brief-writer mission_brief VISIBLE to its two consumers.
--
-- ⚠ _HOLD: NOT APPLIED. Written and evidenced; awaiting the owner's word, because copyonline.co.uk
-- is mid-build and his standing instruction is not to change a running build.
--
-- THE DEFECT (bugs_open/453 + this lane's CONTRIB). Exactly two live agents read the brief and both
-- read `mission_brief.text`; nothing reads it any other way:
--   domain-research-classifier.classify_and_extract  — decides what the site IS
--   build-site-planner.plan_site                     — decides what pages it HAS
-- A brief-writer mission_brief is a structured object with NO `text` key, so both guards open on the
-- parent and print a missing child: the prompt asserts "an owner mission exists and is the primary
-- source", then renders `<no value>`.
--
-- EVIDENCE, all at the rendered artefact (llm_call_log), with controls:
--   advertise / designblog / seotools / websitepromotion plan_site prompts: "## Mission" then <no value>
--   advertise / seotools / copyonline classify_and_extract prompts:        heading then <no value>
--   gamedesign.uk (its brief HAS .text):  renders its real mission        <- working control
--   boxingonline.com (no mission_brief):  renders no block at all          <- negative control
-- 7 of 23 current mission_brief specs lack .text and all 7 are brief-writer output.
--
-- WHY THIS IS THE RIGHT FIX AND NOT A WORKAROUND: the consumers' contract is `.text`; the working
-- control proves a brief carrying it renders in both steps. The alternative — teaching both templates
-- to render the object — is a shared-seam change belonging to 453, not to one lane's data.
--
-- WHAT IT DOES: supersedes the current mission_brief with an identical object PLUS a `text` key
-- holding a prose rendering of the brief. Nothing existing is altered, so a consumer reading any
-- other key is unaffected.
--
-- ⚠ AFTER APPLYING, PROVE IT RAN — do not repeat 734's mistake of calling a config change verified:
--   1. fire or await one classification, then
--   2. read the rendered prompt and confirm the mission text appears where <no value> was:
--      SELECT substring(prompt_rendered from position('## Pre-Defined Mission' in prompt_rendered) for 300)
--        FROM llm_call_log WHERE agent_type='domain-research-classifier'
--         AND prompt_rendered LIKE '%Domain: <domain>%' ORDER BY created_at DESC LIMIT 1;
--   3. the planner's needle is '## Mission', NOT '## Pre-Defined Mission' — searching for the wrong
--      heading is how this lane briefly mis-diagnosed the planner as a separate defect.
--
-- SET :dom BEFORE RUNNING, e.g.  psql -v dom="'copyonline.co.uk'" -f this_file

BEGIN;

DO $g$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain = :dom AND ss.aspect='mission_brief' AND ss.is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'REFUSED: expected exactly 1 current mission_brief for %, found %', :dom, n; END IF;
  SELECT count(*) INTO n FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain = :dom AND ss.aspect='mission_brief' AND ss.is_current AND (ss.data ? 'text');
  IF n <> 0 THEN RAISE EXCEPTION 'REFUSED: this brief already carries a text key — nothing to do'; END IF;
END $g$;

UPDATE site_specs ss SET is_current=false, superseded_at=now()
  FROM sites s WHERE s.id=ss.site_id AND s.domain = :dom
   AND ss.aspect='mission_brief' AND ss.is_current;

-- The prose rendering is built from the brief's own fields, so it cannot drift from the object.
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by, notes)
SELECT prev.site_id, 'mission_brief',
       prev.data || jsonb_build_object('text',
         concat_ws(E'\n\n',
           'PROPOSITION: '   || COALESCE(prev.data->>'proposition',''),
           'AUDIENCE: '      || COALESCE(prev.data->>'audience',''),
           'DIFFERENTIATION: '|| COALESCE(prev.data->>'differentiation',''),
           'STANCE: '        || COALESCE(prev.data->>'stance',''),
           'MUST NOT: '      || COALESCE((SELECT string_agg('- '||value, E'\n') FROM jsonb_array_elements_text(prev.data->'must_nots')),''),
           'PLANNED PAGES: ' || COALESCE((SELECT string_agg('- '||(it->>'name')||' ('||COALESCE(it->>'kind','')||', '||COALESCE(it->>'priority','')||'): '||COALESCE(it->>'what',''), E'\n') FROM jsonb_array_elements(prev.data->'content_plan') it),'')
         )),
       'brief-visibility-fix', 'portfolio_positioning', true, 'portfolio_positioning',
       'Adds a text key so the brief reaches classify_and_extract and plan_site, which both read mission_brief.text and rendered <no value> without it (bugs_open/453 CONTRIB). Object otherwise unchanged.'
  FROM site_specs prev JOIN sites s ON s.id=prev.site_id
 WHERE s.domain = :dom AND prev.aspect='mission_brief' AND prev.superseded_at IS NOT NULL
 ORDER BY prev.superseded_at DESC LIMIT 1;

DO $v$
DECLARE n int; t text;
BEGIN
  SELECT count(*) INTO n FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain = :dom AND ss.aspect='mission_brief' AND ss.is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'VERIFY: % current rows', n; END IF;
  SELECT ss.data->>'text' INTO t FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain = :dom AND ss.aspect='mission_brief' AND ss.is_current;
  IF t IS NULL OR length(t) < 200 THEN RAISE EXCEPTION 'VERIFY: text key missing or implausibly short (%)', COALESCE(length(t),0); END IF;
  SELECT count(*) INTO n FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain = :dom AND ss.aspect='mission_brief' AND ss.is_current AND (ss.data ? 'proposition') AND (ss.data ? 'content_plan');
  IF n <> 1 THEN RAISE EXCEPTION 'VERIFY: the original object was not preserved'; END IF;
END $v$;
COMMIT;

-- ============================================================================
-- ⚠ ADDENDUM 2026-09-03 ~17:35Z — THIS FILE IS STEP 1 OF 2, NOT THE WHOLE FIX.
--
-- I told the owner earlier that the repair was "one well-evidenced change". That was true when the
-- consumers had not yet run. They have now, and a blind read is PERSISTED, so restoring visibility
-- does not undo it. Correcting the record here rather than leaving the file to imply otherwise.
--
-- MEASURED SEQUENCE on copyonline.co.uk (site 3d965325-519a-4515-b79f-50c886954a80), from
-- site_work_items created_at/updated_at, all times UTC 2026-09-03:
--   15:55:06 → 15:56:45  needs_composition   completed NOT READY: missing identity+classification;
--                                            it queued a backfill classifier and closed itself.
--   15:56:41 → 16:58:07  needs_domain_research (backfill) — the classifier ran BLIND here.
--   16:54:55 → (claimed) needs_vertical_research — reading the blind classification NOW.
--   no resolved_composition spec has EVER been written; no plan exists.
--
-- SO THE DAMAGE IS UPSTREAM OF THE PLANNER AND STILL SPREADING. Applying step 1 alone hands the
-- planner a correct brief AND the wrong classification side by side, because plan_site's
-- input_fields include site_specs, which carries `classification`. It would be an improvement and
-- not a repair.
--
-- STEP 2 — RE-RUN THE CLASSIFIER once the brief is visible, and check the artefact, not the status:
--   the current classification says category=hub, tags marketplace/community-platform/tool-portal,
--   inferred (its own `reasoning` says so) from the old Drupal 7 rules page. A correct re-run should
--   not say "no mission brief was supplied" — that sentence is the pass/fail test, and it is written
--   by the agent itself into site_specs.classification.data->>'reasoning'.
--   Do NOT hand-edit the classification to match the brief. That would put this lane's opinion where
--   the pipeline's output belongs, and the estate's rule is that the framework writes the content.
--
-- STEP 3, only if the owner wants the tool set corrected: the tools spec (16:01:52Z) was written by
--   the same blind path — its own reasoning says "Without existing pages loaded, I'm inferring from
--   the domain" — and produced six suggestions, five of them now built and deployed, none of them
--   among the brief's four. Retiring those pages is a separate, owner-level decision, not part of
--   this fix, and three of the five are the seotools duplicates already awaiting his ruling.
--
-- STILL NOT APPLIED. Nothing above changes the permission position.
-- ============================================================================
