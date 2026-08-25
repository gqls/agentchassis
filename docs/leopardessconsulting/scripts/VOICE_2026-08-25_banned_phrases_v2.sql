-- ============================================================================
-- VOICE_2026-08-25_banned_phrases_v2.sql — SUPERSEDES VOICE_2026-08-17_…_ready.sql
--
-- Why v2 exists: the 08-17 file, run on 2026-08-25, REFUSED ITSELF (exit 3) —
-- correctly — for two expiries its own verification caught:
--   1. its page_components UPDATE pinned an id (7425fde8…) that a rerender had
--      re-minted (UPDATE 0) — the §1.1 lesson, learned the day after it was written;
--   2. the portfolio aspect had GROWN two new "honest" sentences (use_cases
--      elements 4 and 5 summaries) since 08-17, so `honest_in_aspect` was 1
--      after its single-element edit.
-- The DO/RAISE rolled everything back; nothing applied. This file fixes ALL
-- occurrences content-addressed (no element titles, no pinned ids).
--
-- Context: owner ruling 2026-08-25 DROPPED the trust rule (applied separately,
-- bak_leo_voice_20260825). The "honest — demonstrate it, never label it" and
-- "earns its keep" rules STAND; this file clears the copy that violates them.
-- ============================================================================

BEGIN;

-- ── backups (site rule: bak_* before ANY change) ────────────────────────────
CREATE TABLE IF NOT EXISTS bak_leo_portfolio_voice_20260825 AS
  SELECT * FROM site_specs
   WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND aspect='portfolio' AND is_current;

CREATE TABLE IF NOT EXISTS bak_leo_insights_pc_20260825 AS
  SELECT pc.* FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE p.site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732'
     AND p.name='insights' AND pc.slot_name='generic-text-block';

-- ── 1. portfolio.use_cases — every string field of every element, one pass ──
-- Content-addressed: the four exact phrases, wherever they occur. Non-string
-- values pass through untouched; element order preserved.
UPDATE site_specs s
   SET data = jsonb_set(s.data, '{use_cases}', (
     SELECT jsonb_agg(cleaned ORDER BY ord) FROM (
       SELECT ord, (
         SELECT jsonb_object_agg(key,
           CASE WHEN jsonb_typeof(e->key) = 'string' THEN to_jsonb(
             replace(replace(replace(replace(e->>key,
               'These are the jobs where an agent earns its keep, because',
               'These are the jobs worth giving to an agent, because'),
               'This is the honest starting point.',
               'This is the starting point.'),
               'the honest answer is a system built specifically for them',
               'the answer is a system built specifically for them'),
               'so the honest next step would be understanding',
               'so the next step would be understanding'))
           ELSE e->key END)
         FROM jsonb_object_keys(e) key) AS cleaned
       FROM jsonb_array_elements(s.data->'use_cases') WITH ORDINALITY t(e, ord)
     ) x))
 WHERE s.site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND s.aspect='portfolio' AND s.is_current;

-- ── 2. /insights.html generic-text-block — addressed by PAGE + SLOT, never id
UPDATE page_components pc
   SET content_data = jsonb_set(content_data, '{content}',
         to_jsonb(replace(content_data->>'content',
           'they need the failure modes named honestly, because',
           'they need the failure modes named, because')))
  FROM pages p
 WHERE p.id = pc.page_id
   AND p.site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732'
   AND p.name='insights' AND pc.slot_name='generic-text-block';

-- ── verification INSIDE the transaction — raises, so it can stop the COMMIT
DO $$
DECLARE n_keep int; n_honest int; n_pc int; asp_len int;
BEGIN
  SELECT count(*) INTO n_keep FROM site_specs
   WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND aspect='portfolio' AND is_current
     AND data::text ~* '\mearns its keep';
  SELECT count(*) INTO n_honest FROM site_specs
   WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND aspect='portfolio' AND is_current
     AND data::text ~* '\mhonest';
  SELECT count(*) INTO n_pc FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE p.site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732'
     AND p.name='insights' AND pc.slot_name='generic-text-block'
     AND pc.content_data::text ~* '\mhonestly';
  IF n_keep <> 0 OR n_honest <> 0 OR n_pc <> 0 THEN
    RAISE EXCEPTION 'voice edits did not take: earns_its_keep=% honest_in_aspect=% honestly_in_pc=%',
      n_keep, n_honest, n_pc;
  END IF;
  -- demand controls: the rows must still exist, be non-trivial, and still carry
  -- known content — or the zeros above are satisfiable by deletion.
  SELECT length(data::text) INTO asp_len FROM site_specs
   WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND aspect='portfolio' AND is_current;
  IF asp_len < 500 THEN
    RAISE EXCEPTION 'portfolio aspect implausibly small (%) after the edit — refusing', asp_len;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM site_specs
    WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND aspect='portfolio' AND is_current
      AND data::text LIKE '%repetitive process%') THEN
    RAISE EXCEPTION 'demand control failed: known phrase missing from aspect — refusing';
  END IF;
  IF (SELECT jsonb_array_length(data->'use_cases') FROM site_specs
    WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND aspect='portfolio' AND is_current) <> 5 THEN
    RAISE EXCEPTION 'use_cases element count changed — refusing';
  END IF;
END $$;

COMMIT;

-- ── THEN publish: use-cases + how-it-works read the aspect; insights has the pc edit.
--   ./docs/leopardessconsulting/scripts/rerender_page_safe.sh \
--       4851f6fc-71cf-4160-a270-e03d6d3e0732 leopardessconsulting.co.uk use-cases
--   (repeat for how-it-works, insights)
-- ── VERIFY AT THE SERVED PAGE (with the positive control):
--   for p in /use-cases.html /how-it-works.html /insights.html; do
--     printf '%s ' "$p"; curl -s "https://leopardessconsulting.co.uk$p?cb=$(date +%s%N)" \
--       | grep -ciE '\bearns its keep\b|\bhonest(ly)?\b'; done   # expect 0 0 0
--   curl -s "https://leopardessconsulting.co.uk/use-cases.html?cb=$(date +%s%N)" \
--     | grep -ci 'repetitive process'                             # expect >= 1
