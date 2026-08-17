-- ============================================================================
-- VOICE_2026-08-17_banned_phrases_ready.sql
--
-- ⚠ PREPARED, NOT APPLIED. Written 2026-08-17 while the Anthropic endpoint was
-- capped (`ai_endpoint_health.healthy=f`, "regain access on 2026-09-01"), which
-- is why it was not run: the edits themselves are inert DB writes, but they only
-- reach the served page through a `page_rerender`, and a rerender that escalates
-- to the LLM writer during the cap fails mid-page. Apply BOTH halves together,
-- once the endpoint is healthy again:
--     SELECT endpoint_url, healthy, error FROM ai_endpoint_health;
--
-- SCOPE: only the two banned phrases that need no owner ruling —
--   \bearns its keep\b   "owner 2026-07-18: overused phrase, not the owner's voice"
--   \bhonest(ly)?\b      "owner 2026-07-18: overused; show the honesty, do not label it"
-- It deliberately does NOT touch \btrust(ed|worthy|s)?\b, which is 138 of the
-- site's 145 served banned-phrase hits and is an OWNER QUESTION, not a copy fix
-- (it currently flags the site's own product name, "The AI Vendor Trust
-- Checklist"). See RUNNING_NOTES 2026-08-17 for the measurement.
--
-- ── THE TRAP THIS FILE EXISTS TO AVOID ──────────────────────────────────────
-- `use-cases-list.use_cases` is `source: site_specs.portfolio.use_cases`, so the
-- "earns its keep" sentence is NOT editable in page_components.content_data —
-- an edit there reads back correctly, passes the escalation gate, and is
-- silently reverted by the very rerender you fire to publish it (LANDMINES, "A
-- `site_specs.<aspect>.<path>`-sourced field…"). The aspect is the authority.
-- One edit there fixes BOTH /how-it-works.html and /use-cases.html, which is
-- also why the same sentence appears on two pages.
-- `generic-text-block.content` is `source: llm`, so that one IS a content_data
-- edit. Verified 2026-08-17 with the landmine's own source query.
--
-- Also note: the `client`/`status` fields below carry "honest" and do NOT appear
-- on the served page today, so a served-HTML sweep cannot see them and a
-- page_components census cannot either. They are fixed here because the aspect
-- is the authority and a template change would surface them.
-- ============================================================================

BEGIN;

-- ── backups (site rule: bak_* before ANY change) ────────────────────────────
CREATE TABLE IF NOT EXISTS bak_leo_portfolio_voice_20260817 AS
  SELECT * FROM site_specs
   WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND aspect='portfolio' AND is_current;

CREATE TABLE IF NOT EXISTS bak_leo_insights_pc_20260817 AS
  SELECT * FROM page_components WHERE id='7425fde8-3469-4e55-bf5d-25a13958d212';

-- ── 1. portfolio.use_cases[2] — "earns its keep" + two "honest" ─────────────
-- Indexed by KEY, not by position: the array order is not ours to rely on.
UPDATE site_specs s
   SET data = jsonb_set(s.data, '{use_cases}', (
         SELECT jsonb_agg(
           CASE WHEN e->>'title' = 'Taking a repetitive process off a person'
                THEN e
                     || jsonb_build_object('summary', replace(e->>'summary',
                          'These are the jobs where an agent earns its keep, because',
                          'These are the jobs worth giving to an agent, because'))
                     || jsonb_build_object('client', replace(e->>'client',
                          'This is the honest starting point.', 'This is the starting point.'))
                     || jsonb_build_object('status', replace(e->>'status',
                          'This is the honest starting point.', 'This is the starting point.'))
                ELSE e END ORDER BY ord)
         FROM jsonb_array_elements(s.data->'use_cases') WITH ORDINALITY t(e, ord)))
 WHERE s.site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND s.aspect='portfolio' AND s.is_current;

-- ── 2. /insights.html generic-text-block — one self-labelling "honestly" ────
-- "named honestly" -> "named": the strongest form of the owner's own rule
-- (show it, do not label it). Shape after replacement is grammatical — no
-- double comma, no dangling article — which is the check the fleet_copy_quality
-- CONTRIB says to make instead of only checking the word is gone.
UPDATE page_components
   SET content_data = jsonb_set(content_data, '{content}',
         to_jsonb(replace(content_data->>'content',
           'they need the failure modes named honestly, because',
           'they need the failure modes named, because')))
 WHERE id='7425fde8-3469-4e55-bf5d-25a13958d212';

-- ── verification INSIDE the transaction — raises, so it can actually stop the
-- COMMIT. A block of bare SELECTs cannot: ON_ERROR_STOP ignores a non-empty
-- result (LANDMINES / RFC_006). Induce it once by inverting a count.
DO $$
DECLARE n_keep int; n_honest int; n_pc int;
BEGIN
  SELECT count(*) INTO n_keep FROM site_specs
   WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND aspect='portfolio' AND is_current
     AND data::text ~* '\mearns its keep';
  SELECT count(*) INTO n_honest FROM site_specs
   WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND aspect='portfolio' AND is_current
     AND data::text ~* '\mhonest';
  SELECT count(*) INTO n_pc FROM page_components
   WHERE id='7425fde8-3469-4e55-bf5d-25a13958d212' AND content_data::text ~* '\mhonest';
  IF n_keep <> 0 OR n_honest <> 0 OR n_pc <> 0 THEN
    RAISE EXCEPTION 'voice edits did not take: earns_its_keep=% honest_in_aspect=% honest_in_pc=%',
      n_keep, n_honest, n_pc;
  END IF;
  -- demand control: the rows must still exist and still be non-trivial, or the
  -- three zeros above would also be satisfied by having deleted everything.
  IF (SELECT length(data::text) FROM site_specs
        WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND aspect='portfolio' AND is_current) < 500 THEN
    RAISE EXCEPTION 'portfolio aspect is implausibly small after the edit — refusing';
  END IF;
END $$;

COMMIT;

-- ── THEN, and only then, publish. TWO pages read the aspect, so both rerender.
--   ./docs/leopardessconsulting/scripts/rerender_page_safe.sh \
--       4851f6fc-71cf-4160-a270-e03d6d3e0732 leopardessconsulting.co.uk use-cases
--   ./docs/leopardessconsulting/scripts/rerender_page_safe.sh \
--       4851f6fc-71cf-4160-a270-e03d6d3e0732 leopardessconsulting.co.uk how-it-works
--   ./docs/leopardessconsulting/scripts/rerender_page_safe.sh \
--       4851f6fc-71cf-4160-a270-e03d6d3e0732 leopardessconsulting.co.uk insights
--
-- ── VERIFY AT THE SERVED PAGE (a COMPLETED rerender is not a repaired page):
--   for p in /use-cases.html /how-it-works.html /insights.html; do
--     printf '%s ' "$p"
--     curl -s "https://leopardessconsulting.co.uk$p?cb=$(date +%s%N)" \
--       | grep -ciE '\bearns its keep\b|\bhonest(ly)?\b'
--   done                                   # expect 0 0 0
--   # CONTROL — the sweep must be able to find something, or the zeros are blind:
--   curl -s "https://leopardessconsulting.co.uk/use-cases.html?cb=$(date +%s%N)" \
--     | grep -ci 'repetitive process'      # expect >= 1
--
-- ── NOT INCLUDED, on purpose:
--   * the two ARCHIVED pages that carry "honest" in content_data
--     (/case-study-data-pipeline-companies-house.html generic-text-block
--      7bed62c3…, /case-study-tool-generation-pipeline.html contact-block
--      24430c18…). They serve 404 today; bugs_open/266 says archived pages get
--      rebuilt by four producers, so they are worth clearing eventually, but
--      not under a voice pass whose scope is what visitors read.
--   * /tools/process-automation-scorer/index.html — "Answer honestly rather
--     than optimistically" is an instruction to the USER about their own
--     answers, not the site labelling itself. Same judgement the fleet_copy_
--     quality CONTRIB §4 applied to "dishonest" meaning "unfair". Left alone.
