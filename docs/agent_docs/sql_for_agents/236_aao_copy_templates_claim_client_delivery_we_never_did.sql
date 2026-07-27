-- 236_aao_copy_templates_claim_client_delivery_we_never_did.sql
--
-- ══ FOUND WHILE FINISHING 235, AND IT IS A DIFFERENT KIND OF PROBLEM ═══════
-- 235 fixed the superseded agent figure in five spec aspects. A final sweep for
-- the same figure across ALL aspects found two more — and neither is a stale
-- number. Both are **copy templates teaching the writer to claim client work
-- that does not exist.** Bumping "70+" to "170+" here would have made a false
-- claim bigger, which is why this is a separate file and not part of 235.
--
--   cta_strategy (index page CTA subtext, verbatim):
--     "We have shipped 70+ agent systems on Kubernetes, Kafka, and Postgres.
--      We know exactly where pipelines break."
--
--   content_standards (the worked EXAMPLE of a credibility sentence):
--     "The same architecture runs 70+ agents in production today across
--      financial services and logistics environments."
--
-- ── THE PREMISE, CHECKED RATHER THAN ASSUMED ───────────────────────────────
--   SELECT count(*) FILTER (WHERE domain NOT LIKE 'pool-%') FROM sites;   -- 15
--   SELECT count(*) FROM sites s WHERE EXISTS (SELECT 1 FROM pages p
--     WHERE p.site_id=s.id AND p.build_status='deployed');                -- 14
--
-- All 15 are this platform's own properties: ai-agent-orchestration, dartsonline,
-- finetuning, fundamentallyai, gamesdesign, gaswholesalers, idea.uk, leopardess,
-- oufe, relojistas, robot-hands, system.internal, vetcomparison, vonc, webdesign.
--
-- So "shipped 70+ agent systems" overstates delivered client systems by roughly
-- seventy times, and **not one** of those sites is a financial-services or
-- logistics deployment. These are not exaggerations of a true thing; they are
-- claims about a business we do not have.
--
-- ── WHY TEMPLATES ARE THE WORST PLACE FOR THIS ─────────────────────────────
-- A false sentence on a page is one false sentence. A false sentence in
-- `content_standards` is an EXAMPLE the writer is told to imitate, and in
-- `cta_strategy` it is canonical copy assigned to a page. They regenerate the
-- claim by design, on every page that uses them — which is 235's finding
-- (a spec is a SOURCE) applied to a claim that was never true in the first place.
--
-- ── WHAT THIS WRITES, AND WHAT IT DELIBERATELY DOES NOT ────────────────────
-- The minimum honest edit: keep each sentence's PURPOSE, replace the claim with
-- the one that is true and registered (170+ agents in production, aao-agent-
-- definitions = 175, gte, with its query attached). No new marketing claim is
-- invented, no sector is named, and nothing is said about clients.
--
-- **This is copy, so it is flagged for the owner rather than treated as
-- settled.** If the intended claim was about client delivery, the honest
-- version needs a real number the owner can attest, and neither exists today.

BEGIN;

DO $fix$
DECLARE
    v_site uuid;
    n int;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain='ai-agent-orchestration.com';
    IF v_site IS NULL THEN RAISE EXCEPTION '236: no site row'; END IF;

    WITH upd AS (
        UPDATE site_specs ss
        SET data = replace(replace(ss.data::text,
              'We have shipped 70+ agent systems on Kubernetes, Kafka, and Postgres.',
              'We run 170+ agents in production on Kubernetes, Kafka, and Postgres.'),
              'The same architecture runs 70+ agents in production today across financial services and logistics environments.',
              'The same architecture runs 170+ agents in production today across the sites this platform builds and operates.'
            )::jsonb,
            updated_at = now()
        WHERE ss.site_id = v_site AND ss.is_current
          AND ss.aspect IN ('cta_strategy','content_standards')
        RETURNING 1
    )
    SELECT count(*) INTO n FROM upd;
    IF n <> 2 THEN RAISE EXCEPTION '236: expected to update 2 spec rows, updated %', n; END IF;

    -- Ban the CLASS, so it cannot come back through a different phrasing.
    -- Patterns are compiled by GO (RE2): \b not \m. Verified to compile, and
    -- verified not to match the replacement copy, before seeding.
    UPDATE site_specs ss
    SET data = jsonb_set(ss.data, '{banned_claims}',
          (ss.data->'banned_claims') || $bc$[
            {"pattern": "shipped\\s+[0-9]{1,4}\\s*\\+?\\s*(agent|ai|multi-agent)\\s*systems?",
             "reason": "2026-07-27: a client-DELIVERY claim. Every site this platform runs is its own property (15 sites, 14 with deployed pages, checked); no count of delivered client systems exists at any value. Was live in the cta_strategy index subtext as 'We have shipped 70+ agent systems'."},
            {"pattern": "across financial services and logistics",
             "reason": "2026-07-27: names client SECTORS we have no deployment in. Was the worked example of a credibility sentence in content_standards, i.e. a template teaching the writer to imitate it."}
          ]$bc$::jsonb),
        updated_at = now()
    WHERE ss.site_id = v_site AND ss.aspect='evidence_base' AND ss.is_current
      AND ss.data ? 'banned_claims';
END $fix$;

-- ── Post-conditions: absence of the CLAIMS, not presence of my edits ───────
DO $post$
DECLARE
    v_shipped int; v_sector int; v_bans int; v_facts int;
BEGIN
    SELECT count(*) FILTER (WHERE ss.data::text ~* 'shipped\s+[0-9]{1,4}\s*\+?\s*(agent|ai|multi-agent)\s*systems?'),
           count(*) FILTER (WHERE ss.data::text ~* 'across financial services and logistics')
      INTO v_shipped, v_sector
    FROM site_specs ss JOIN sites s ON s.id=ss.site_id
    WHERE s.domain='ai-agent-orchestration.com' AND ss.is_current AND ss.aspect <> 'evidence_base';

    IF v_shipped <> 0 THEN RAISE EXCEPTION '236: % spec(s) still claim delivered client systems', v_shipped; END IF;
    IF v_sector  <> 0 THEN RAISE EXCEPTION '236: % spec(s) still name client sectors', v_sector; END IF;

    SELECT jsonb_array_length(ss.data->'banned_claims'), jsonb_array_length(ss.data->'facts')
      INTO v_bans, v_facts
    FROM site_specs ss JOIN sites s ON s.id=ss.site_id
    WHERE s.domain='ai-agent-orchestration.com' AND ss.aspect='evidence_base' AND ss.is_current;

    IF v_bans  <> 10 THEN RAISE EXCEPTION '236: expected 10 banned_claims, found %', v_bans; END IF;
    IF v_facts <> 7  THEN RAISE EXCEPTION '236: facts changed to % — expected 7', v_facts; END IF;

    RAISE NOTICE '236 OK: client-delivery claims removed from both templates; % banned_claims, % facts', v_bans, v_facts;
END $post$;

COMMIT;
