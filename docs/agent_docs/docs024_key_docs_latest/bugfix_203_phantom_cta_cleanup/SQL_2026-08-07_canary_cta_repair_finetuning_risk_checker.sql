-- bugfix 203 cleanup — CANARY: repair one misdirected CTA through the framework
--
-- WHY THIS SHAPE. The owner authorised route (2) from PLAN D7: page-content-writer with
-- spec.mode='edit_live' (bugs_open/178's edit-not-regenerate channel). The spec keys below
-- are copied from the one live emitter of this item type,
-- create_tool_cross_link_items.go:242-279 — including the precedent for naming an exact URL
-- in `suggestion` ("use that URL exactly as written, do not alter or invent one"), which is
-- what keeps this framework-legitimate rather than hand-authored content: the URL is derived
-- from a real pages row, and the FRAMEWORK writes the copy.
--
-- status='triaged' + handler_agent='page-build-handler' is what the dispatcher selects on
-- (load_work_item_actions.go:653 — `wi.status IN ('triaged','approved')`).
-- page_id is set in BOTH the column and the spec, per the hand-made-work-item landmine.
--
-- TARGET (all ids read live 2026-08-07):
--   site   finetuning.uk                        1368e337-dd1d-4799-bbb3-8221a1b79bcc
--   page   /guides/tool-ai-data-risk-checker-guide.html  856e2b44-49e1-4abb-a1eb-13df784d1f32
--   name   tool-ai-data-risk-checker-guide      (page_type blog-post)
--
-- BEFORE STATE (pinned so the diff is assertable, not impressionistic):
--   hero            2908 bytes  html_md5 dd767cfba85fe90e4b1e482f2209c146  cdata_md5 3acb498efb1b21577c59b23571029763  phantom=YES
--   article-body   12440 bytes  html_md5 b958d6249c605f5b29fd21b65d11605c  cdata_md5 b89e4456c146f92b4db87a0186d60456  phantom=no
--   call-to-action  2443 bytes  html_md5 b2d1e81a8a7210999b3419dacb9bf4c1  cdata_md5 350421879d01f0b308b6e20501881e48  phantom=no
--   hero content_data has cta_text="Run the Risk Checker" and NO cta_url — the phantom href
--   was fabricated at render time, which is why the fix removes it and cannot re-point it.
--   SERVED page confirmed carrying the defect at 08:0xZ 2026-08-07:
--     <a href="/contact.html" class="btn btn-primary">Run the Risk Checker
--
-- THE TARGET URL EXISTS (this is what makes repair, not deletion, the right outcome):
--   finetuning.uk/tools/tool-ai-data-risk-checker.html  (pages.page_type='tool')
--
-- ROLLBACK: this INSERT only queues work; it changes no page. To abandon before dispatch,
--   UPDATE site_work_items SET status='cancelled' WHERE item_key='cta_repair:tool-ai-data-risk-checker-guide:hero:1368e337-dd1d-4799-bbb3-8221a1b79bcc';

BEGIN;

INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary, spec,
    page_id, priority, handler_agent, status, created_by, item_key
) VALUES (
    '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
    'bugs_open/203_cleanup',
    'build',
    'content_rewrite',
    'medium',
    'Hero CTA "Run the Risk Checker" points at /contact.html — re-aim at the tool page it names',
    jsonb_build_object(
        'page',      'tool-ai-data-risk-checker-guide',
        'page_name', 'tool-ai-data-risk-checker-guide',
        'page_id',   '856e2b44-49e1-4abb-a1eb-13df784d1f32',
        'description',
            'The hero call-to-action button is labelled "Run the Risk Checker" but its link '
            || 'points at /contact.html, which is not the risk checker. The destination was '
            || 'fabricated by a render-time default that has since been fixed (bugs_open/203), '
            || 'so the button now has no destination at all and would disappear on a plain '
            || 'rebuild. The tool it names does exist and the button should point at it.',
        'suggestion',
            'Set the hero section''s cta_url to /tools/tool-ai-data-risk-checker.html — use '
            || 'that URL exactly as written, do not alter or invent one. Keep the existing '
            || 'button label "Run the Risk Checker" as it is. Do NOT rewrite, reword, shorten '
            || 'or restructure any prose on this page: the headline, subheadline, article body '
            || 'and closing call-to-action must survive byte-for-byte. The only change wanted '
            || 'is the hero button''s destination.',
        'acceptance_test',
            'The hero section''s content_data contains cta_url = "/tools/tool-ai-data-risk-checker.html", '
            || 'the hero still shows the label "Run the Risk Checker", and the article-body and '
            || 'call-to-action sections are unchanged.',
        'source',           'bugs_open/203_cleanup',
        'work_item_type',   'content_rewrite',
        'max_fix_attempts', 1,
        -- opts this item into load_current_section_content (bugs_open/178): the writer is
        -- handed the page's own current rendered_html to EDIT rather than nothing to
        -- fabricate a replacement from. Target is an existing, already-built page.
        'mode', 'edit_live'
    ),
    '856e2b44-49e1-4abb-a1eb-13df784d1f32',
    110,
    'page-build-handler',
    'triaged',
    'bugfix-203-cleanup-session',
    'cta_repair:tool-ai-data-risk-checker-guide:hero:1368e337-dd1d-4799-bbb3-8221a1b79bcc'
);

-- Verify exactly one row is queued, in a dispatchable state, carrying edit_live.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM site_work_items
     WHERE item_key = 'cta_repair:tool-ai-data-risk-checker-guide:hero:1368e337-dd1d-4799-bbb3-8221a1b79bcc'
       AND status = 'triaged' AND handler_agent = 'page-build-handler'
       AND spec->>'mode' = 'edit_live' AND page_id = '856e2b44-49e1-4abb-a1eb-13df784d1f32';
    IF n <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 dispatchable edit_live row, found %', n;
    END IF;
END $$;

COMMIT;
