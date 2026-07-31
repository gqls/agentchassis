-- 270_bug161_correct_gamesdesign_monte_carlo_fact.sql
--
-- bugs_open/161 — STEP 1 of the fix: correct the FALSE registered fact.
--
-- WHAT IS WRONG. gamesdesign.co.uk's evidence_base carries fact `gd-trials`,
-- "Monte Carlo trials per query" = 10000, sourced to "the figure is hard-coded in
-- the shipped drop-rate tool JavaScript". Neither drop-rate tool performs any
-- Monte Carlo simulation and neither contains any randomness at all
-- (`Math.random` count 0 in both, measured 2026-07-31 against
-- page_components.rendered_html with the export length asserted against the
-- column's own length()):
--
--   tool-drop-rate-tuner      closed-form Math.pow(1-p,k) survival + a CDF array
--                             indexed by kill count, optional hard pity cap.
--                             Its own doc comment: "Cumulative distribution
--                             modelled via geometric distribution with optional
--                             hard pity cap". Its only `10000`-ish literal is a
--                             `pity > 100000` bound.
--   tool-drop-rate-simulator  says "binomial" 6x; its single 10000 is
--                             `return Math.min(val, 10000)` — an INPUT CLAMP on
--                             attempts, not a trial count. Untouched since
--                             2026-06-05, i.e. seven weeks BEFORE this fact was
--                             registered, so the fact was false on arrival.
--
-- So the NUMBER is real and its stated MEANING is wrong. 10,000 is the maximum
-- attempts the tools will model. That is exactly the reading another session
-- reached independently when it repaired the homepage stat card on 2026-07-30
-- ("Max Attempts Modelled" / "computes exact binomial probabilities for up to ten
-- thousand") — that repair was right, and it corrected the copy without
-- correcting the register, which is why the register has been contradicting the
-- repaired page ever since.
--
-- WHY THIS IS NOT MERELY A TYPO. The register is BOTH the whitelist injected into
-- the page-content-writer prompt (`writer_block`, headed "NUMBERS (state only
-- these...)") AND the authority `numberSupported` checks published numbers
-- against. So this row instructed writers to state the falsehood and then vouched
-- for it, disarming the prose scan, ScanStatClaims and the bugs_open/149 C1
-- persistence floor in one go — all three call that one function.
--
-- WHAT THIS FILE DOES NOT DO, deliberately:
--   * It does NOT add the banned_claims patterns. Those are BLOCKER severity, so
--     arming them while the false copy is still published would make 6 pages
--     UNSAVEABLE with the falsehood still live. Copy repair comes first.
--   * It does NOT touch published copy. 10 components still assert the false
--     claim; that is step 2 and it is a separate, reviewed change.
--
-- MEASURED EFFECT OF THIS FILE ALONE: ZERO new findings. Verified with
-- cmd/claimscan over the complete live corpus (67 components, export row count
-- asserted 67/67, no truncation stubs) against this exact corrected register:
-- "0 finding(s) across 67 component(s)" — identical to the current register's
-- baseline. The reason is that all 10 offending components sit on page types in
-- `editorialPageTypes` (guide, game, blog-post), which are exempt from
-- ScanUnregisteredNumbers. So this step strands nothing and flags nothing; it
-- stops the writer being INSTRUCTED to state the falsehood, and stops the engine
-- vouching for it. Detection of the already-published copy needs the banned
-- patterns, because ScanBannedClaims runs on every surface.
--
-- CONTEXT_TERMS: "monte carlo" and "simulation" are REMOVED and "trial" is NOT
-- kept. Keeping "trial" would leave "10,000 Monte Carlo trials" inside a window
-- that still matches this fact, so the engine would go on treating the false
-- sentence as supported — which is the whole defect. "attempt"/"modelled"/"max"
-- keep the homepage's honest "Max Attempts Modelled" stat card supported
-- (verified: that window contains "Attempts").
--
-- Supersede-then-insert rather than UPDATE, following writeRefreshedEvidenceBase
-- and because idx_site_specs_current is UNIQUE on (site_id, aspect) WHERE
-- is_current — so the two statements must be one transaction, in this order.
-- The old row is kept as history, which is the point: this register was wrong for
-- seven days and that fact should remain readable.

BEGIN;

-- Guard: refuse to run if the row is not in the state this file was written for.
-- A concurrent session may already have corrected it.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
    FROM site_specs
    WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
      AND aspect = 'evidence_base' AND is_current
      AND data->'facts' @> '[{"id":"gd-trials","claim":"Monte Carlo trials per query"}]'::jsonb;
    IF n <> 1 THEN
        RAISE EXCEPTION 'bug161: expected exactly 1 current evidence_base row still carrying the false gd-trials claim, found %. Someone else may have fixed it — re-read before applying.', n;
    END IF;
END $$;

-- 1. supersede the wrong row (kept as history)
UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
  AND aspect = 'evidence_base' AND is_current;

-- 2. insert the corrected register
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, pinned, notes)
SELECT
    s.site_id,
    'evidence_base',
    jsonb_set(
        jsonb_set(
            s.data,
            '{facts}',
            (
                SELECT jsonb_agg(
                    CASE WHEN f->>'id' = 'gd-trials' THEN
                        jsonb_build_object(
                            'id',            'gd-trials',
                            'kind',          'metric',
                            'claim',         'maximum attempts modelled per query',
                            'value',         10000,
                            'tolerance',     'exact',
                            'context_terms', jsonb_build_array('attempt', 'modelled', 'max'),
                            'source',        jsonb_build_object(
                                'artifact',
                                'return Math.min(val, 10000) — the input clamp on attempts in tool-drop-rate-simulator; the tuner''s CDF is built to the same bound. NOT a trial count: neither tool samples (Math.random count 0 in both, 2026-07-31). The tuner is closed-form Math.pow(1-p,k) + CDF; the simulator computes exact binomial probability.'
                            ),
                            'verified_at',   '2026-07-31',
                            'writer_line',   '{value} maximum attempts modelled per query in the drop-rate tools, using exact probability rather than sampling'
                        )
                    ELSE f END
                    ORDER BY ord
                )
                FROM jsonb_array_elements(s.data->'facts') WITH ORDINALITY AS t(f, ord)
            )
        ),
        '{writer_block}',
        to_jsonb(
            'NUMBERS (state only these, with their listed meaning; as of 2026-07-31):' || E'\n' ||
            '- 11 interactive design tools live, all client-side and free' || E'\n' ||
            '- 10,000 maximum attempts modelled per query in the drop-rate tools (the tools compute EXACT probability — a closed-form binomial/geometric distribution — they do NOT run Monte Carlo trials or any random sampling. 10,000 is the input clamp on attempts, Math.min(val, 10000).)' || E'\n' ||
            '- 4 configurable inputs in the drop-rate tuner: drop chance, kills per hour, pity timer, target hours' || E'\n' ||
            '- 10 guides & articles live (5 blog posts + 5 guides)' || E'\n' ||
            'NOT TRACKED / DOES NOT EXIST, NEVER STATE: user counts, accuracy-gap percentages, PRD figures (no tool on this site implements PRD), economy model/archetype counts, pity parameters other than the four listed, and — added 2026-07-31, bugs_open/161 — Monte Carlo trials, random simulations, or any sampling technique attributed to these tools: they are exact analytic calculators. You may discuss Monte Carlo as a general technique and recommend it to the reader; you may never say OUR tools perform it.'
        )
    ),
    'manual',
    NULL,
    'session-2026-07-31-bug161-register-correction',
    s.pinned,
    'bugs_open/161: gd-trials claimed "Monte Carlo trials per query" and cited shipped tool JavaScript that contains no randomness at all (Math.random count 0 in both drop-rate tools). The number 10000 is real (Math.min(val,10000) attempts clamp); its stated meaning was false, and it was false when registered on 2026-07-24 — the simulator component was last written 2026-06-05 and never changed. Because the register is both the writer whitelist and the gate authority, this one row caused the claim and then vouched for it. Superseded row kept as history. Does NOT arm banned_claims and does NOT repair the 10 published components; both are separate steps.'
FROM site_specs s
WHERE s.site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
  AND s.aspect = 'evidence_base'
  AND s.is_current = false
  AND s.superseded_at IS NOT NULL
ORDER BY s.superseded_at DESC
LIMIT 1;

-- 3. assert the outcome before committing
DO $$
DECLARE
    n_current int; still_wrong int; wb_wrong int;
BEGIN
    SELECT count(*) INTO n_current FROM site_specs
    WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf' AND aspect='evidence_base' AND is_current;
    IF n_current <> 1 THEN
        RAISE EXCEPTION 'bug161: expected exactly 1 current row after the write, found %', n_current;
    END IF;

    SELECT count(*) INTO still_wrong FROM site_specs
    WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf' AND aspect='evidence_base' AND is_current
      AND (data->'facts')::text ~* 'monte carlo';
    IF still_wrong <> 0 THEN
        RAISE EXCEPTION 'bug161: the corrected facts[] still mentions Monte Carlo';
    END IF;

    -- the writer_block MUST still mention Monte Carlo, in the NEVER STATE clause
    SELECT count(*) INTO wb_wrong FROM site_specs
    WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf' AND aspect='evidence_base' AND is_current
      AND (data->>'writer_block') ~ 'never say OUR tools perform it';
    IF wb_wrong <> 1 THEN
        RAISE EXCEPTION 'bug161: the writer_block prohibition did not land';
    END IF;

    -- all four facts must survive: this corrects one, it does not drop any
    IF (SELECT jsonb_array_length(data->'facts') FROM site_specs
        WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf' AND aspect='evidence_base' AND is_current) <> 4 THEN
        RAISE EXCEPTION 'bug161: fact count changed — this file must correct one fact, never drop one';
    END IF;
END $$;

COMMIT;
