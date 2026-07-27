-- ============================================================================
-- 227_correct_overclaim_contract_scanner_claim.sql
--
-- CORRECTS A FALSE STATEMENT I PUT INTO A LIVE COUNCIL SEAT IN MIGRATION 223.
--
-- 223 told the compliance reviewer, in the contract text and again in judge
-- clause (e):
--     "it is the one class every scanner misses … invisible to all of them"
--     "remember no scanner will catch this, so this seat is the only control"
--
-- **Both are wrong.** `ScanBannedClaims` (datahelpers/claims.go:284-325) is a
-- bare case-insensitive regex over prose blocks — no number extraction, no
-- `businessClaimContextRe`, no `isExcludedNumber`; those gate only
-- `ScanUnregisteredNumbers` (claims.go:365,369). A `banned_claims` pattern
-- catches whatever you write into it, about anyone, numeric or not. Live
-- registers already carry purely qualitative patterns ("leaderboard", "live
-- now", "price target", "years of experience").
--
-- The true statement, and the one the seat now carries:
--   the scanner CAN catch this class, but only on a site where somebody wrote a
--   pattern for it — and as of 2026-07-26 only **5 of 15 live sites carry a
--   single banned_claims pattern**, with no mechanism to define a set once and
--   apply it fleet-wide. So in practice the class is usually uncaught, which is
--   a COVERAGE gap, not a capability gap.
--
-- WHY THIS MATTERS ENOUGH TO SPEND A MIGRATION ON
--   A reviewer told "no scanner will catch this, you are the only control" will
--   reason differently from one told "a scanner can catch it if the site is
--   armed — check whether it is". The first invites the seat to substitute for a
--   mechanism; the second invites it to ask for the mechanism. The second is
--   what we want, and it is also true.
--
--   It is a poor look to ship an overclaim-detection contract that itself
--   contains an unverified claim stated with confidence. The failure mode this
--   seat exists to catch is exactly the one I committed writing it.
--
-- Seated on fix-proposer; run the 099 mirror afterwards, as with 223:
--   python3 docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/099_SYNC_gate_roster.py --apply
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,review_compliance,config,prompt_template}',
      to_jsonb(
        replace(
          replace(
            default_config->'workflow'->'steps'->'review_compliance'->'config'->>'prompt_template',

            -- the contract paragraph's false premise
            'a promise about OUR OWN accuracy is a claim like any other, and it is the one class every scanner misses. banned_claims, the number scan, the evidence whitelist and citation re-verification all police claims about THIRD PARTIES, nearly all of them numeric; a claim about US, stated qualitatively, is invisible to all of them.',
            'a promise about OUR OWN accuracy is a claim like any other, and it is the class most often left uncovered. The deterministic scanner CAN catch it -- banned_claims is an unrestricted case-insensitive regex over prose, so it matches whatever patterns a site has been given, numeric or not -- but as of 2026-07-26 only 5 of 15 live sites carry a single banned_claims pattern, and there is no mechanism to define a pattern set once and apply it fleet-wide. So treat this as a COVERAGE question, not an impossibility: where a plan introduces or leaves standing a claim of this kind, ask whether the affected site is actually armed for it, and say so.'
          ),

          -- the judge clause's false premise
          '(the overclaimed-reliability class -- remember no scanner will catch this, so this seat is the only control)',
          '(the overclaimed-reliability class -- a banned_claims pattern can catch this on an armed site, so do not assume a scanner covered it and do not assume none could; where it matters, ask for the pattern)'
        )
      ),
      false
    ),
    updated_at = NOW()
WHERE type = 'fix-proposer'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'review_compliance'->'config'->>'prompt_template' LIKE '%one class every scanner misses%';

COMMIT;

-- Verify (expect false on both, for BOTH rosters, after the mirror runs)
--   SELECT type,
--     (default_config->'workflow'->'steps'->'review_compliance'->'config'->>'prompt_template'
--        LIKE '%every scanner misses%')            AS still_false_1,
--     (default_config->'workflow'->'steps'->'review_compliance'->'config'->>'prompt_template'
--        LIKE '%no scanner will catch this%')      AS still_false_2,
--     (default_config->'workflow'->'steps'->'review_compliance'->'config'->>'prompt_template'
--        LIKE '%5 of 15 live sites%')              AS has_correction
--   FROM agent_definitions
--    WHERE type IN ('fix-proposer','council-gate') AND is_active
--      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
