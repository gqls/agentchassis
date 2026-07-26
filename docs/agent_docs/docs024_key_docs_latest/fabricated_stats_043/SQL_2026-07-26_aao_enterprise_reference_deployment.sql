-- bugs_open/043 — the live residual the 2026-07-24 sweep missed.
--
-- ai-agent-orchestration.com/enterprise-reference-deployment.html (HTTP 200, live)
-- still serves the poisoned-spec figures that wave-2d corrected on /index only:
--
--   70+       agents in concurrent production
--   8         departments running isolated agent groups
--   (empty)   agent types under full audit coverage      <- renders "30+" from the
--                                                           2026-05-01 HTML; the stored
--                                                           value was emptied later and
--                                                           the page never re-rendered
--   Thousands concurrent agent instances under stable load
--   Minutes   to root cause from production alert
--
-- The component was last written 2026-05-01 and the page was never in the sweep, which
-- keyed on stat[_]?N_value and missed cardN_stat_value on a page nobody listed.
--
-- ── WHY EACH ONE CHANGES ───────────────────────────────────────────────────
-- Wave-2d's truth audit still holds for the COUNTS: "over 70 agents", "8 departments"
-- and "30+ agent types" are all TRUE and conservative (live: 175 active definitions,
-- 174 distinct types, and 8 is the platform's own named taxonomy). What is wrong here
-- is the FRAMING, and two of the five are on the site's own NEVER-STATE list:
--
--   card1  "in concurrent production"        -> concurrency is NOT measured. NEVER-STATE.
--   card2  "running isolated agent groups"   -> an architecture claim nothing evidences.
--   card3  value empty, "full audit coverage"-> "full coverage" is unmeasured; and the
--                                               empty required field was itself freezing
--                                               this page (see below).
--   card4  "Thousands ... concurrent"        -> the exact clause wave-2d removed from the
--                                               specs as untrue. NEVER-STATE.
--   card5  "Minutes to root cause"           -> a per-incident outcome nothing records.
--
-- Replacements come only from migration 218's registered facts for this site, each of
-- which carries the SQL that defines it. card5 has no honest replacement, so it is left
-- EMPTY — which is now a legal, rendering answer thanks to migration 217, and is the
-- whole point of that migration. Do NOT put a number back into card5.
--
-- ── THIS PAGE WAS A SECOND, UNFOUND INSTANCE OF bugs_open/073 ──────────────
-- `card3_stat_value` was stored EMPTY while `case-studies-grid` still marked it
-- required. Pre-217 that meant every re-render of this page escalated to the writer and
-- every rebuild died at the render gate — the page was frozen exactly as aao/index was,
-- and nobody had found it. Migration 217 unfroze it; this file gives it something true
-- to say when it next renders.
--
-- Content_data edits do NOT change the served page. A re-render is required, and the
-- build pipeline is currently down (bugs_open/029) — so this lands the truth in the data
-- and the page will pick it up on its next render. Verify LIVE, not from the DB.
--
-- Idempotent: sets values outright; re-running changes nothing.

\set ON_ERROR_STOP on

BEGIN;

DO $fix$
DECLARE
    pcid CONSTANT uuid := 'c4c3d2b4-bdf0-4c4e-827a-f688ed841ce5';
    n int;
BEGIN
    SELECT count(*) INTO n FROM page_components WHERE id = pcid;
    IF n <> 1 THEN
        RAISE EXCEPTION '043: page_component % not found — page_components.id is NOT stable across re-renders; re-key on (page, component, label) before applying', pcid;
    END IF;

    UPDATE page_components
       SET content_data = content_data
             || jsonb_build_object(
                  'card1_stat_value', '175',
                  'card1_stat_label', 'agent definitions in the production registry',
                  'card2_stat_value', '8',
                  'card2_stat_label', 'departments in the platform''s own agent taxonomy',
                  'card3_stat_value', '174',
                  'card3_stat_label', 'distinct agent types',
                  'card4_stat_value', '1,834',
                  'card4_stat_label', 'orchestrations run in the last 24 hours',
                  -- No measured per-incident figure exists. Empty is the honest answer
                  -- and, post-217, a rendering one: the stat simply does not appear.
                  'card5_stat_value', '',
                  'card5_stat_label', ''),
           updated_at = now()
     WHERE id = pcid;

    RAISE NOTICE '043: enterprise-reference-deployment stats re-pointed to registered facts; card5 left honestly empty';
END $fix$;

-- Verify: nothing unregistered survives, and card5 is empty rather than invented.
SELECT e.k || ' = [' || e.v || ']'
FROM page_components pc, LATERAL jsonb_each_text(pc.content_data) e(k,v)
WHERE pc.id = 'c4c3d2b4-bdf0-4c4e-827a-f688ed841ce5'
  AND e.k ~ 'card[1-5]_stat_(value|label)'
ORDER BY e.k;

COMMIT;
