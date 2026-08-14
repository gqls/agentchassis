-- FILE: SQL_2026-08-14_restore_cta_urls_10_rows.sql  (bugs_open/268 repair)
--
-- THE 10 REGENERATION-LOSS ROWS, restored from page_component_history.
-- These are the only rows of the 217-row label-without-URL census that EVER
-- held a destination URL in an archived generation (split re-measured
-- 2026-08-14 ~16:50Z: 10 ever-held / 73 never-held / 134 no-history, of 217).
-- The other ~207 are the unresolved_cta never-resolved class — nothing to
-- restore, out of scope here (bugs_open/268 §11.1).
--
-- Every URL below is RECOVERED, NOT INVENTED: it is read from the newest
-- history generation of the same page_id+slot_name whose content_data carried
-- the key (the delete-archive trigger stores OLD, so these are the values the
-- live page actually served before the 268 drop destroyed them). Extraction
-- query in the lane RUNBOOK. Only keys the current row LACKS are merged —
-- `||` keeps everything else as-is.
--
-- WHY content_data AND NOT rendered_html: the platform's own rule (crib:
-- ai_site_selling_automation/SQL_2026-08-12d_restore_cta_urls.sql) — a
-- hand-patched rendered_html re-arms the same loss and trips
-- page_divergence_overwritten. Fix the data, then re-render (no LLM).
--
-- LOCKS, corrected 2026-08-14: webdesign.uk index/call-to-action is NOT
-- locked — the 08-12 repair locked index/HERO (repaired) but never touched
-- index/call-to-action (lost 08-11 13:43, before that lane's baseline, so it
-- was not in their repair set and not in their lock sweep). The handoff and
-- 268 §11.1 said it "sits LOCKED"; the live lock map refutes that. All 10
-- rows are unlocked; the locked_at guard below is belt-and-braces.
--
-- The fix (8f899cc8d, live v1.0.1298+) makes this restore durable: the next
-- content_rewrite carries these keys instead of destroying them. That
-- composition (fix-then-repair) is proven by an edit_live rewrite on one
-- repaired row afterwards — the permanence step in the lane HANDOFF §3.

BEGIN;

UPDATE page_components pc SET
  content_data = pc.content_data || v.urls,
  updated_at = now()
FROM pages p, (VALUES
  -- ai-agent-orchestration.com (2a8ebf9c-20a2-4c39-b191-840b012371da)
  ('2a8ebf9c-20a2-4c39-b191-840b012371da'::uuid, 'news',              'hero',           '{"cta_url":"/tools/password-entropy.html","primary_cta_url":"/contact.html","secondary_cta_url":"/services.html"}'::jsonb),
  -- dartsonline.com (5fe8785b-223d-41a3-88ee-c07187622381)
  ('5fe8785b-223d-41a3-88ee-c07187622381'::uuid, 'grip-styles',       'call-to-action', '{"primary_cta_url":"/tools/dart-weight-comparator/index.html","secondary_cta_url":"/tools/setup-builder/index.html"}'::jsonb),
  ('5fe8785b-223d-41a3-88ee-c07187622381'::uuid, 'grip-styles',       'hero',           '{"cta_url":"/tools/dart-weight-comparator/index.html","secondary_cta_url":"/tools/setup-builder/index.html"}'::jsonb),
  ('5fe8785b-223d-41a3-88ee-c07187622381'::uuid, 'index',             'call-to-action', '{"primary_cta_url":"/tools/dart-weight-comparator/index.html","secondary_cta_url":"/brands/index.html"}'::jsonb),
  ('5fe8785b-223d-41a3-88ee-c07187622381'::uuid, 'index',             'hero',           '{"cta_url":"/tools/dart-weight-comparator/index.html","secondary_cta_url":"/brands/index.html"}'::jsonb),
  -- idea.uk (1244516d-014d-421c-88c6-090bb1e9552a)
  ('1244516d-014d-421c-88c6-090bb1e9552a'::uuid, 'tool-funding-fit',  'hero',           '{"cta_url":"/tools/funding-fit/index.html","secondary_cta_url":"/report.html"}'::jsonb),
  ('1244516d-014d-421c-88c6-090bb1e9552a'::uuid, 'tool-patent-check', 'hero',           '{"cta_url":"/guides/patents/index.html","secondary_cta_url":"/report.html"}'::jsonb),
  -- vonc.com (9ec3b9ee-5b08-461b-b4f8-9e1e03579c74)
  ('9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'::uuid, 'archetypes',        'call-to-action', '{"primary_cta_url":"/tools/gauntlet/round.html","secondary_cta_url":"/tools/archetype-taster-quiz/index.html"}'::jsonb),
  ('9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'::uuid, 'archetypes',        'hero',           '{"cta_url":"/tools/gauntlet/round.html","secondary_cta_url":"/tools/archetype-taster-quiz/index.html"}'::jsonb),
  -- webdesign.uk (1fcfa4f3-ec80-4010-878b-b971cd46711f)
  ('1fcfa4f3-ec80-4010-878b-b971cd46711f'::uuid, 'index',             'call-to-action', '{"primary_cta_url":"/contact.html","secondary_cta_url":"tel:+44 (0) 7934 524 911"}'::jsonb)
) AS v(site_id, page_name, slot, urls)
WHERE pc.page_id = p.id
  AND p.site_id = v.site_id
  AND p.name = v.page_name
  AND p.status = 'active'
  AND pc.slot_name = v.slot
  AND pc.locked_at IS NULL;          -- never write through a lock

-- Verify: all 10 rows must now hold a destination key. DO/RAISE, not a bare
-- SELECT — ON_ERROR_STOP ignores a non-empty result (WRONG_CALLS class).
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
  FROM page_components pc JOIN pages p ON p.id=pc.page_id
  WHERE p.status='active'
    AND (pc.content_data ? 'cta_url' OR pc.content_data ? 'primary_cta_url')
    AND (p.site_id, p.name, pc.slot_name) IN (
      ('2a8ebf9c-20a2-4c39-b191-840b012371da'::uuid,'news','hero'),
      ('5fe8785b-223d-41a3-88ee-c07187622381'::uuid,'grip-styles','call-to-action'),
      ('5fe8785b-223d-41a3-88ee-c07187622381'::uuid,'grip-styles','hero'),
      ('5fe8785b-223d-41a3-88ee-c07187622381'::uuid,'index','call-to-action'),
      ('5fe8785b-223d-41a3-88ee-c07187622381'::uuid,'index','hero'),
      ('1244516d-014d-421c-88c6-090bb1e9552a'::uuid,'tool-funding-fit','hero'),
      ('1244516d-014d-421c-88c6-090bb1e9552a'::uuid,'tool-patent-check','hero'),
      ('9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'::uuid,'archetypes','call-to-action'),
      ('9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'::uuid,'archetypes','hero'),
      ('1fcfa4f3-ec80-4010-878b-b971cd46711f'::uuid,'index','call-to-action'));
  IF n <> 10 THEN
    RAISE EXCEPTION 'restore incomplete: % of 10 rows hold a destination key', n;
  END IF;
END $$;

COMMIT;
