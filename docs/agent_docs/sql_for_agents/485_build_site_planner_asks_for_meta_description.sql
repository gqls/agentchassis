-- 485 — build-site-planner: ASK for a meta description (bugs_open/320, mechanism M1)
--
-- WHY. `pages.meta_description` is the one-sentence summary Google prints under a
-- page's title, and on this estate it is also the excerpt the blog-listing
-- component renders in each card. Measured 2026-08-19: **407 of 731 active pages
-- (55.7%) across 26 of 27 sites have none.**
--
-- THE CAUSE IS AN OMISSION IN THIS PROMPT, NOT A BUG IN THE WRITER. The page
-- object in `plan_site`'s `Return JSON:` template asks for exactly:
--
--   name, title, page_type, nav_label, nav_order, in_header, in_footer, sections
--
-- and NOTHING else. Meanwhile `upsertPage` (site_db_actions.go:1131) does:
--
--   metaDescription := datahelpers.GetStringField(page, "meta_description", "")
--
-- i.e. it asks the plan for a key the plan was never told to produce, and takes
-- the empty string. So every plan-built page is born with no description and
-- always has been — `content` pages are 81.4% empty, `landing` 85.3%.
--
-- ⚠ `default_config::text ILIKE '%meta_description%'` ON THIS AGENT RETURNS TRUE
-- AND MEANS NOTHING. The planner already contains the string, in the
-- `load_existing_pages` step, which SELECTs the column. Matching the string proves
-- the agent mentions the field, not that it is asked to write one. That grep is
-- why this went unnoticed; read the output schema, not the census.
--
-- WHY THE PROMPT AND NOT A GO COMPOSER. Owner ruling 2026-08-06 — the framework
-- writes the content, a session does not. A Go format string is right for tool
-- pages, where the sentence really is mechanical (`composedToolMetaDescription`),
-- and wrong for arbitrary content pages, which are most of the 407. The planner
-- already authors `title` and `nav_label` for these pages; a description is the
-- same authorial job, one field along.
--
-- PAIRS WITH THE GO HALF, AND THE ORDER DOES NOT MATTER. Migration M2's fix
-- (site_db_actions.go, same bug) changes the upsert's conflict clause to
-- `COALESCE(NULLIF(EXCLUDED.meta_description,''), pages.meta_description)`, so a
-- blank can no longer destroy an existing description. Applying THIS file before
-- that ships is safe: a non-blank incoming value has always won, and this file
-- only makes incoming values non-blank. Applying it after is equally safe.
--
-- WHAT IT DOES NOT DO. It does not fill the 407 pages that already exist. Those
-- are plan-managed for 295 of them (a later replan now carries a description) and
-- 112 are not, which is what `save_page_meta_description` (register SEO-004) and
-- its workflow are for.
--
-- Scoped by id, pre-state gated, DO/RAISE verify, snapshot first, rollback
-- sidecar — the shape 445/446/484 used. Config-only, LIVE ON APPLY, no roll.
-- Not council-submitted: agent_definitions config is outside the gate's
-- platform/ internal/ pkg/ scope.
--
-- ROLLBACK: 485_build_site_planner_asks_for_meta_description_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('build-site-planner',
  '485_build_site_planner_asks_for_meta_description: pre-update');

DO $$
DECLARE n int; p text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '485: expected exactly 1 live build-site-planner row, found % — a second active row would make the id-scoped UPDATE a silent no-op', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
    INTO p
    FROM agent_definitions WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

  IF p IS NULL THEN
    RAISE EXCEPTION '485: plan_site.config.prompt_template is NULL — wrong step or the workflow has changed under me, refusing';
  END IF;
  IF position('"title": "Page Title | Site Name",' in p) = 0 THEN
    RAISE EXCEPTION '485: the page-object title line is not present verbatim — the Return JSON template has changed under me, refusing to insert blind';
  END IF;
  IF position('You have FINAL SAY on architecture.' in p) = 0 THEN
    RAISE EXCEPTION '485: anchor "You have FINAL SAY on architecture." missing — refusing to insert the rule blind';
  END IF;
  IF position('"meta_description"' in p) > 0 THEN
    RAISE EXCEPTION '485: already applied (the page object already names meta_description) — refusing to double-apply';
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,plan_site,config,prompt_template}',
         to_jsonb(
           replace(
             replace(
               default_config#>>'{workflow,steps,plan_site,config,prompt_template}',
               $old$      "title": "Page Title | Site Name",$old$,
               $new$      "title": "Page Title | Site Name",
      "meta_description": "One sentence, 120-155 characters, saying what a visitor gets from this page.",$new$
             ),
             $anchor$You have FINAL SAY on architecture.$anchor$,
             $rule$EVERY page needs a `meta_description`. It is what a search engine prints under the page title, and on listing pages it is shown as the card excerpt, so it is read by people far more often than the page itself. Write it as a promise to a visitor, not as a description of the build: 120-155 characters, one sentence, plain English, no site name suffix, no build or generator wording. If a page is genuinely too thin to describe, that is a signal the page should not be planned.

You have FINAL SAY on architecture.$rule$
           )
         )
       ),
       updated_at = now()
 WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b'
   AND type = 'build-site-planner'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
    INTO p
    FROM agent_definitions WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

  -- Post-conditions. These are DO/RAISE and not bare SELECTs on purpose: a verify
  -- block made of SELECTs cannot stop the COMMIT (ON_ERROR_STOP ignores a
  -- non-empty result), which is a trap this estate has already paid for.
  IF position('"meta_description": "One sentence, 120-155 characters' in p) = 0 THEN
    RAISE EXCEPTION '485 VERIFY: the page object still does not carry meta_description — the replace did not fire';
  END IF;
  IF position('EVERY page needs a `meta_description`.' in p) = 0 THEN
    RAISE EXCEPTION '485 VERIFY: the authoring rule was not inserted';
  END IF;
  IF position('You have FINAL SAY on architecture.' in p) = 0 THEN
    RAISE EXCEPTION '485 VERIFY: the anchor line was consumed rather than preserved';
  END IF;
  -- The anchor must appear exactly once: the rule text re-emits it, so a botched
  -- replace that duplicated the block would show up here and nowhere else.
  IF (length(p) - length(replace(p, 'You have FINAL SAY on architecture.', ''))) / length('You have FINAL SAY on architecture.') <> 1 THEN
    RAISE EXCEPTION '485 VERIFY: anchor appears more than once — the prompt has been duplicated';
  END IF;
  RAISE NOTICE '485 OK: build-site-planner now asks for a meta description on every page';
END $$;

COMMIT;
