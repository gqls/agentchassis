-- FILE: SQL_2026-08-12b_copy_migration_canaries.sql
--
-- webdesign.uk £149 copy migration, stage 2 of 2: the page copy itself.
-- Stage 1 (SQL_2026-08-12_evidence_base_149.sql) put the £149 register live and
-- armed the bans; this hands the pages to the framework's own writer.
--
-- THROUGH THE FRAMEWORK, NOT BY HAND (owner ruling 2026-08-04). These are
-- ordinary `content_rewrite` work items on the build pipeline, the same rows
-- apply_gap_plan_action.go writes, so page-build-handler picks them up with no
-- special path and every gate the pipeline applies still applies.
--
-- `spec.mode = 'edit_live'` IS LOAD-BEARING, not decoration (bugs_open/178).
-- Without it, load_current_section_content is a pass-through, the writer gets
-- the guidance text and NOTHING of the page's current prose, and it fabricates
-- a replacement section that satisfies the instruction while dropping most of
-- what was there. That was measured at 4,439 -> 1,806 chars on one page, one
-- paragraph in three surviving. A copy migration that gutted these pages would
-- look like a successful migration in every status column.
--
-- TWO CANARIES FIRST, DELIBERATELY, and they are chosen to DISAGREE:
--   faq   — the heaviest page (8 of the 36 claim findings: £75, deposit,
--           refund x2, 14 days x2, rounds of changes x2, "We handle the setup",
--           plus "refunds" in its hero). If edit_live holds here it holds
--           anywhere.
--   index — the lightest (1 finding, the phrase "the refund window"). Proves
--           the minimal-edit case does not rewrite a page that barely needed
--           touching.
-- The remaining three (how-it-works, what-you-get,
-- tool-website-brief-starter-guide) are held until these two are read at the
-- artefact. The guide page especially: it is long-form prose carrying internal
-- links, and this estate has already measured a writer silently dropping 5 of
-- 13 links from a guide while reporting success.
--
-- QUEUE POSITION, checked before writing: webdesign.uk has ZERO dispatchable
-- build items (0 triaged, 0 approved — 92 complete, 8 cancelled, 5 awaiting
-- human review, 2 failed, 2 blocked, 1 wont_fix). So these are first in their
-- own site's queue rather than behind the fleet's ~700-item backlog, which is
-- fairness-ordered per site.
--
-- ROLLBACK: UPDATE site_work_items SET status='cancelled' WHERE item_key LIKE
-- 'copy_migration_149_%' AND site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f';

BEGIN;

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, page_id, priority,
   handler_agent, status, created_by, item_key, pipeline, triaged_at)
SELECT
  p.site_id,
  'ai-site-selling-automation',
  'content_rewrite',
  'high',
  'Migrate ' || p.name || ' to the £149 offer: the deposit, the refund window and the two-rounds term are retired',
  jsonb_build_object(
    'page_name', p.name,
    'mode', 'edit_live',
    'source', 'ai-site-selling-automation-2026-08-12',
    'content_guidance',
    'The offer this site sells has changed, and this page states terms we will not honour. '
    || 'Edit the existing sections in place. Keep the structure, the headings, the register and '
    || 'everything that is not about the commercial terms, and change only what the new terms require. '
    || 'Do not add sections and do not remove sections. '
    || E'\n\n'
    || 'REMOVE COMPLETELY, because these no longer exist: the £75 deposit; the fourteen-day acceptance '
    || 'or refund window; a refund of any kind, including the phrase "the refund window"; and "two rounds '
    || 'of revisions". They belonged to the previous offer and were retired by the owner on 2026-08-11. '
    || E'\n\n'
    || 'STATE INSTEAD, using only the verified facts you have been given: the site costs £149 in total and '
    || 'no VAT is added; the customer pays after they have seen the finished site on a private preview link '
    || 'and approved it; we do not offer refunds; one set of changes is included in the price and anything '
    || 'after that is charged as work; we build only a few sites at a time and submissions close when we '
    || 'are full; the site is built by AI and the page says so plainly rather than implying otherwise; the '
    || 'customer gets a private preview link and then a ZIP of the finished site to host wherever they '
    || 'like; hosting and the domain are not included and stay with the customer, with hosting by us and a '
    || 'manual domain transfer available as optional paid extras. '
    || E'\n\n'
    || 'ONE PHRASING RULE THAT WILL BLOCK THE PAGE IF YOU GET IT WRONG: write "we do not offer refunds" or '
    || '"we never offer refunds". Do NOT write "there is no refund". The claims gate treats a bare "no" as '
    || 'an intensifier rather than a negation, so the second form reads to it as a refund promise and fails '
    || 'the page.'
    || CASE p.name
         WHEN 'faq' THEN E'\n\n'
           || 'THIS PAGE SPECIFICALLY: the answer to "What about the domain and hosting?" currently says '
           || '"We handle the setup as part of getting the site live". That is now false. The customer keeps '
           || 'their own domain and their own DNS and hosts the site themselves. The question "How many '
           || 'rounds of changes do I get?" should become a question about changes, not rounds, and its '
           || 'answer is one set. The hero mentions refunds in its list of what the page covers; that word '
           || 'has to go with the rest.'
         WHEN 'index' THEN E'\n\n'
           || 'THIS PAGE SPECIFICALLY: the only retired term here is step 3 of the how-it-works summary, '
           || 'which says the customer can "decline it within the refund window". Replace that step with '
           || 'what actually happens now: they ask for their changes, or they approve the site and pay. '
           || 'Everything else on this page is about how the service works and should survive unchanged. '
           || 'The home page currently states no price at all; £149 belongs on it.'
         ELSE ''
       END
  ),
  p.id,
  20,                              -- ahead of the site's own default 35s
  'page-build-handler',
  'triaged',
  'ai-site-selling-automation-2026-08-12',
  'copy_migration_149_' || p.name,
  'build',
  now()
FROM pages p
WHERE p.site_id = '1fcfa4f3-ec80-4010-878b-b971cd46711f'
  AND p.status = 'active'
  AND p.name IN ('faq', 'index');

-- Assert, and ABORT if wrong. RAISE rather than a SELECT: ON_ERROR_STOP does
-- not stop a COMMIT on a non-empty result, so a verify block of SELECTs cannot
-- actually hold the transaction.
DO $$
DECLARE n int; n_editlive int; n_archived int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND item_key LIKE 'copy_migration_149_%' AND status='triaged';
  IF n <> 2 THEN RAISE EXCEPTION 'expected 2 canary items, found %', n; END IF;

  -- mode=edit_live is the whole safety property. If it is missing the writer
  -- replaces the page instead of editing it, and nothing downstream says so.
  SELECT count(*) INTO n_editlive FROM site_work_items
   WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND item_key LIKE 'copy_migration_149_%' AND spec->>'mode' = 'edit_live';
  IF n_editlive <> 2 THEN
    RAISE EXCEPTION 'edit_live missing on % of 2 items — the writer would fabricate replacements', 2 - n_editlive;
  END IF;

  -- The archived index-rejected page shares the url /index.html and must never
  -- be handed to the writer. p.status='active' is what excludes it; this
  -- asserts the filter did its job rather than trusting that it did.
  SELECT count(*) INTO n_archived FROM site_work_items w
    JOIN pages pg ON pg.id = w.page_id
   WHERE w.item_key LIKE 'copy_migration_149_%' AND pg.status <> 'active';
  IF n_archived <> 0 THEN RAISE EXCEPTION 'a non-active page was queued for rewrite (% items)', n_archived; END IF;
END $$;

COMMIT;
