-- FILE: SQL_2026-08-12c_copy_migration_remaining_three.sql
--
-- webdesign.uk £149 copy migration: the remaining three pages, released after
-- both canaries were read at the SERVED artefact (not at their status).
--
-- WHAT THE CANARIES ESTABLISHED, measured:
--   faq   4,512 -> 5,339 bytes body (+18%), scan 8 findings -> 0, live at
--         16:45:03Z with £149 x4, "we do not offer refunds", "built by AI".
--   index 9,617 -> 9,713 bytes, scan 1 finding -> 0, live at 16:55:29Z.
--   Neither page was gutted, so `edit_live` is doing what bugs_open/178 says
--   it does, and the writer chose the mandated refund phrasing unprompted by
--   anything except writer_block and the ban's own reason.
--
-- THE GUIDE PAGE IS THE RISKY ONE, and it is protected MECHANICALLY, not by a
-- sentence in the brief. `tool-website-brief-starter-guide` is long-form prose
-- whose article-body carries four internal links. This estate has already
-- measured a writer dropping 5 of 13 guide links while reporting success, and
-- has ruled that a prose instruction to preserve a SET is not followed
-- (48dcd2eda). So before this runs, the page's own `required_links` are
-- declared in pages.content_direction and gate_page_links.py is proven live
-- against them:
--   python3 …/loanandmortgagecalculator_couk/gate_page_links.py --domain webdesign.uk
--     -> ok, all 4 required links present
--   …--self-test  -> FAIL on an impossible link, "CONTROL OK"
-- Run the same gate AFTER these complete. A pass that was never shown capable
-- of failing is not a pass.
--
-- SYNC LAG, so the next reader does not misread it as a broken deploy: the box
-- pulls on a timer. index.html was `deployed` in the DB at 16:50:44Z and was
-- still serving the 02:22 file at 16:52 — it appeared at 16:55:29Z. Allow ~5
-- minutes between `deployed` and the artefact, and check last-modified rather
-- than refetching the body.
--
-- ROLLBACK: UPDATE site_work_items SET status='cancelled' WHERE item_key LIKE
-- 'copy_migration_149_%' AND status IN ('triaged','approved');

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
         WHEN 'how-it-works' THEN E'\n\n'
           || 'THIS PAGE SPECIFICALLY: the "what happens first" section ends with the whole retired offer in '
           || 'one run of sentences: the fourteen days, the decline-for-a-refund, the £75 deposit, and two '
           || 'rounds of revisions. Replace that run with what happens now: the customer gets the preview '
           || 'link, asks for their one set of changes, then approves and pays, and receives the site as a '
           || 'ZIP to host themselves. Everything before it, about the two questions and the three or four '
           || 'days, is still true and should survive.'
         WHEN 'what-you-get' THEN E'\n\n'
           || 'THIS PAGE SPECIFICALLY: it states "Two rounds of revisions are included in the fixed price", '
           || 'refers to "the 14-day acceptance window described on the How It Works page", and points at '
           || '"full terms, including price and refund conditions". All three are now wrong. This is the '
           || 'page that must be explicit about what is NOT included: hosting, the domain, and any ongoing '
           || 'service. Say what the customer receives (a preview link, then a ZIP of the finished site) '
           || 'and what they do not.'
         WHEN 'tool-website-brief-starter-guide' THEN E'\n\n'
           || 'THIS PAGE SPECIFICALLY: it is a guide, not a sales page, and only its commercial paragraph '
           || 'is wrong. That paragraph currently gives the fourteen days, the price-back-minus-a-£75-deposit, '
           || 'and two rounds of revisions. Replace ONLY those terms. '
           || 'PRESERVE EVERY INTERNAL LINK THE PAGE ALREADY HAS: /faq.html, /how-it-works.html, '
           || '/what-you-get.html and /tools/website-brief-starter/index.html must all still be linked when '
           || 'you are done. This is checked mechanically after the rewrite, not read for.'
         ELSE ''
       END
  ),
  p.id,
  20,
  'page-build-handler',
  'triaged',
  'ai-site-selling-automation-2026-08-12',
  'copy_migration_149_' || p.name,
  'build',
  now()
FROM pages p
WHERE p.site_id = '1fcfa4f3-ec80-4010-878b-b971cd46711f'
  AND p.status = 'active'
  AND p.name IN ('how-it-works', 'what-you-get', 'tool-website-brief-starter-guide');

DO $$
DECLARE n int; n_editlive int; n_guide_links int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND item_key LIKE 'copy_migration_149_%' AND status='triaged';
  IF n <> 3 THEN RAISE EXCEPTION 'expected 3 new items triaged, found %', n; END IF;

  SELECT count(*) INTO n_editlive FROM site_work_items
   WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND item_key LIKE 'copy_migration_149_%' AND status='triaged'
     AND spec->>'mode' = 'edit_live';
  IF n_editlive <> 3 THEN
    RAISE EXCEPTION 'edit_live missing on % of 3 — the writer would fabricate replacements', 3 - n_editlive;
  END IF;

  -- The guide's link set must be DECLARED before its rewrite is queued, or the
  -- gate that protects it has nothing to assert and passes vacuously.
  SELECT jsonb_array_length(content_direction->'required_links') INTO n_guide_links
    FROM pages WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND name='tool-website-brief-starter-guide';
  IF COALESCE(n_guide_links,0) < 4 THEN
    RAISE EXCEPTION 'guide page required_links not declared (got %) — gate_page_links would pass vacuously', COALESCE(n_guide_links,-1);
  END IF;
END $$;

COMMIT;
