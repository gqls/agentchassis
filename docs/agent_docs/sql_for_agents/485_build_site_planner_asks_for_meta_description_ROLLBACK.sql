-- 485 ROLLBACK — remove the meta_description field and rule from build-site-planner
--
-- Reverses 485_build_site_planner_asks_for_meta_description.sql by removing the
-- two inserted fragments, rather than by restoring a snapshot: another session
-- may have edited this prompt for an unrelated reason since, and a snapshot
-- restore would silently revert their work too.
--
-- ⚠ ROLLING THIS BACK RE-ARMS bugs_open/320's mechanism M1 — plan-built pages go
-- back to being born with no meta description. Only do it if the planner is
-- producing bad descriptions, and prefer editing the wording to removing the ask.

BEGIN;

SELECT snapshot_agent('build-site-planner',
  '485_ROLLBACK: pre-revert');

DO $$
DECLARE n int; p text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '485 ROLLBACK: expected exactly 1 live build-site-planner row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
    INTO p FROM agent_definitions WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

  IF p IS NULL OR position('"meta_description": "One sentence, 120-155 characters' in p) = 0 THEN
    RAISE EXCEPTION '485 ROLLBACK: 485 is not applied (page object does not carry the field) — nothing to revert';
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
               $new$
      "meta_description": "One sentence, 120-155 characters, saying what a visitor gets from this page.",$new$,
               $old$$old$
             ),
             $rule$EVERY page needs a `meta_description`. It is what a search engine prints under the page title, and on listing pages it is shown as the card excerpt, so it is read by people far more often than the page itself. Write it as a promise to a visitor, not as a description of the build: 120-155 characters, one sentence, plain English, no site name suffix, no build or generator wording. If a page is genuinely too thin to describe, that is a signal the page should not be planned.

You have FINAL SAY on architecture.$rule$,
             $anchor$You have FINAL SAY on architecture.$anchor$
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
    INTO p FROM agent_definitions WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

  IF position('"meta_description"' in p) > 0 THEN
    RAISE EXCEPTION '485 ROLLBACK VERIFY: the field is still present';
  END IF;
  IF position('EVERY page needs a `meta_description`.' in p) > 0 THEN
    RAISE EXCEPTION '485 ROLLBACK VERIFY: the rule is still present';
  END IF;
  IF position('You have FINAL SAY on architecture.' in p) = 0 THEN
    RAISE EXCEPTION '485 ROLLBACK VERIFY: the anchor line was destroyed rather than preserved';
  END IF;
  RAISE NOTICE '485 ROLLBACK OK — the planner no longer asks for a meta description (M1 is re-armed)';
END $$;

COMMIT;
