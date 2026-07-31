-- 271_bug161_arm_monte_carlo_banned_patterns.sql
--
-- bugs_open/161 — STEP 3: arm detection, so the falsehood cannot come back silently.
--
-- ORDER IS LOAD-BEARING. Run this ONLY AFTER 270 (correct the register) and after
-- the 10 published components have been repaired. banned_claims are BLOCKER
-- severity: the bugs_open/149 C1 persistence floor REFUSES a save carrying one.
-- Arming this while the false copy was still published would have made 6 pages
-- unsaveable with the falsehood still live. The guard below enforces that order.
--
-- WHY A BANNED PATTERN IS THE RIGHT INSTRUMENT, measured rather than assumed.
-- All 10 offending components sit on page types in `editorialPageTypes`
-- (guide, game, blog-post), which are exempt from ScanUnregisteredNumbers — so
-- correcting the register flagged NOTHING (verified: 0 findings before, 0 after).
-- But claims.go states the complementary rule explicitly: "ScanBannedClaims runs
-- on every surface. A banned pattern is a human-authored record of a KNOWN
-- falsehood, not a heuristic, so it has no false-positive problem to protect
-- against". We now have a known falsehood, so this is exactly its case.
--
-- PRECISION, MEASURED against the complete live corpus (67 components, export row
-- count asserted 67/67, no base64 truncation stubs), 2026-07-31:
--
--   over the ORIGINAL copy:  fires on 10 components — precisely the 10 that
--                            asserted our tools run Monte Carlo trials; 18 findings
--   spared:                  tool-loot-table-balancer-guide / article-body, whose
--                            only Monte Carlo prose is general technique teaching
--                            ("Monte Carlo simulation tells you what actually
--                            happens...") and advice to the reader ("across at
--                            least 10,000 simulated players") — both TRUE
--   false positives:         0
--   over the REPAIRED copy:  0 findings across 67 components
--
-- The discriminator is deliberate: pattern 1 requires the NUMBER adjacent to
-- "Monte Carlo", and pattern 2 requires the count to be attributed to a query.
-- Teaching the technique and recommending it to the reader both remain sayable;
-- claiming OUR tools perform it does not. A bare /Monte\s*Carlo/ would have been
-- one character shorter to write and would have blocked the honest guide.
--
-- SCOPE: per-site. These go in gamesdesign.co.uk's own register, NOT the
-- fleet-wide set, because the falsehood is about this site's own tools. No other
-- site is affected.

BEGIN;

-- Guard 1: the register must already be corrected (270 applied).
-- Guard 2: no component may still carry the false claim, or this strands pages.
DO $$
DECLARE
    corrected int;
    still_false int;
BEGIN
    SELECT count(*) INTO corrected
    FROM site_specs
    WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
      AND aspect = 'evidence_base' AND is_current
      AND data->'facts' @> '[{"id":"gd-trials","claim":"maximum attempts modelled per query"}]'::jsonb;
    IF corrected <> 1 THEN
        RAISE EXCEPTION 'bug161: register is not in the 270-corrected state (found %). Apply 270 first.', corrected;
    END IF;

    SELECT count(*) INTO still_false
    FROM page_components pc JOIN pages p ON p.id = pc.page_id
    WHERE p.site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
      AND (pc.content_data::text ~* '10[,.]?000\s*Monte\s*Carlo'
           OR pc.content_data::text ~* 'Monte\s*Carlo\s+trials?\s+per\s+query'
           OR pc.rendered_html ~* '10[,.]?000\s*Monte\s*Carlo'
           OR pc.rendered_html ~* 'Monte\s*Carlo\s+trials?\s+per\s+query');
    IF still_false <> 0 THEN
        RAISE EXCEPTION 'bug161: % component(s) still assert the false claim. Repair the copy BEFORE arming, or these pages become unsaveable with the falsehood live.', still_false;
    END IF;
END $$;

UPDATE site_specs
SET data = jsonb_set(
        data,
        '{banned_claims}',
        COALESCE(data->'banned_claims', '[]'::jsonb) || jsonb_build_array(
            jsonb_build_object(
                'pattern', '10[,.]?000\s*Monte\s*Carlo',
                'reason',  'FALSE TECHNIQUE CLASS (bugs_open/161): neither drop-rate tool performs Monte Carlo simulation. Math.random count is 0 in both; the tuner is closed-form Math.pow(1-p,k) plus a CDF sized maxKills = Math.max(1, kph * hours) from the user''s own inputs, and the simulator computes exact binomial probability. The only real 10000 is Math.min(val, 10000), an input CLAMP on attempts modelled, not a trial count. Deliberately requires the NUMBER adjacent to "Monte Carlo" so it fires on the tool claim and spares legitimate teaching about the technique and advice to the reader. You may teach Monte Carlo and recommend it; you may never say OUR tools perform it.'
            ),
            jsonb_build_object(
                'pattern', 'Monte\s*Carlo\s+trials?\s+per\s+query',
                'reason',  'FALSE TECHNIQUE CLASS (bugs_open/161), phrasing variant that survives a number rewrite. "trials per query" attributes a trial count to OUR tool, which is the false part; the technique itself may still be discussed. Kept separate from the number pattern so a rewrite that drops the figure but keeps the attribution is still caught.'
            )
        )
    ),
    updated_at = now()
WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
  AND aspect = 'evidence_base' AND is_current;

-- Assert the outcome, and assert the patterns COMPILE by using them.
DO $$
DECLARE
    n_patterns int;
    self_test int;
BEGIN
    SELECT jsonb_array_length(data->'banned_claims') INTO n_patterns
    FROM site_specs
    WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
      AND aspect = 'evidence_base' AND is_current;
    IF n_patterns <> 2 THEN
        RAISE EXCEPTION 'bug161: expected 2 banned patterns, found %', n_patterns;
    END IF;

    -- A pattern that matches nothing is the failure mode LANDMINES.md records
    -- ("narrowing a detector can make it inert"). Prove both fire on the original
    -- sentences and spare the legitimate one.
    SELECT count(*) INTO self_test FROM (VALUES
        ('The drop-rate tuner runs 10,000 Monte Carlo trials per query.', true),
        ('Every query calculates 10,000 Monte Carlo trials to expose the limits.', true),
        ('Monte Carlo simulation tells you what actually happens across a distribution.', false),
        ('Run Monte Carlo simulations across at least 10,000 simulated players.', false)
    ) AS t(sentence, should_fire)
    WHERE (t.sentence ~* '10[,.]?000\s*Monte\s*Carlo'
           OR t.sentence ~* 'Monte\s*Carlo\s+trials?\s+per\s+query') <> t.should_fire;
    IF self_test <> 0 THEN
        RAISE EXCEPTION 'bug161: the banned patterns do not discriminate as measured (% mismatched case(s))', self_test;
    END IF;
END $$;

COMMIT;
