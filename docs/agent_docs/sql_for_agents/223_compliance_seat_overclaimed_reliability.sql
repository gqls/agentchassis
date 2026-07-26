-- ============================================================================
-- 223_compliance_seat_overclaimed_reliability.sql
--
-- Adds two content-integrity contracts to the COMPLIANCE / LEGAL EYE council
-- seat, and a fleet-wide writing rule to match.
--
-- WHY (owner direction, 2026-07-26, from a live catch on oufe.com)
--   oufe.com is a site whose entire positioning is epistemic honesty. Its
--   generated copy independently produced, and shipped live:
--       "Every factual claim is sourced to a named, dated primary document."
--       "A claim without a named, dated source does not appear here."
--       "If we can't trace a number to a document we hold, it doesn't appear here."
--       "This discipline is not a disclaimer. It's the method."
--   That is a promise of infallibility of process, and we cannot keep it. A
--   citation proves PROVENANCE, not CORRECTNESS: it cannot show we read the
--   source properly, chose the right passage, or that the source itself is right,
--   and it does nothing whatever about a model writing a plausible sentence
--   around a perfectly real link — the failure this estate has actually suffered.
--
--   NOTHING IN THE ESTATE LOOKS FOR THIS, and that is not an oversight in any one
--   component. Every layer we have — banned_claims, ScanUnregisteredNumbers, the
--   evidence_base whitelist, V5 citation re-verification — polices claims about
--   THIRD PARTIES, and almost all of them numeric. A promise about OUR OWN
--   accuracy is a claim about US, and a qualitative one. It is a different
--   failure class, it is invisible to every scanner, and the only control that
--   caught it was a human reading the live page.
--
--   So it goes where a human-equivalent reader already sits: the compliance seat.
--
-- ALSO ADDED: the illustration-not-authority posture (owner, same date).
--   Risk concentrates in asserting live facts about real named entities; value
--   concentrates in explaining mechanism. Those are separable — a mechanism can
--   be taught exactly using openly hypothetical figures, and hypothetical figures
--   cannot be wrong about anybody. So: lead with mechanism, and present the real
--   case as clearly-marked illustration ("a possibly inaccurate case study")
--   rather than as an authoritative account.
--
-- WHERE THIS IS APPLIED
--   Step 1 patches `fix-proposer`, which is the seat roster's source of truth.
--   Step 2 is the 099 mirror, run separately — DO NOT hand-patch council-gate
--   (CLAUDE.md: two hand-maintained rosters that must stay identical is exactly
--   the drift class this council reviews for):
--     python3 docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/099_SYNC_gate_roster.py --apply
--   Step 3 adds the matching writing rule to the content writers, so the rule
--   shapes copy at generation time and is not only caught at review time.
--
-- Idempotent: every UPDATE is a no-op once the marker text is present.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. fix-proposer :: review_compliance — two new contracts + two judge clauses
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,review_compliance,config,prompt_template}',
      to_jsonb(
        replace(
          replace(
            default_config->'workflow'->'steps'->'review_compliance'->'config'->>'prompt_template',

            -- (a) append two contracts after SCANNERS AND GATES
            '- SCANNERS AND GATES: claims-verification machinery (scans, banned-claim lists, evidence gates) must not be weakened or bypassed by a fix.',
            '- SCANNERS AND GATES: claims-verification machinery (scans, banned-claim lists, evidence gates) must not be weakened or bypassed by a fix.
- OVERCLAIMED RELIABILITY -- a promise about OUR OWN accuracy is a claim like any other, and it is the one class every scanner misses. banned_claims, the number scan, the evidence whitelist and citation re-verification all police claims about THIRD PARTIES, nearly all of them numeric; a claim about US, stated qualitatively, is invisible to all of them. Copy or a spec that says in effect "every figure is sourced", "a claim without a source does not appear here", "fully verified", "authoritative", "accurate", or "you can rely on this" is asserting an infallibility the platform cannot deliver: a citation proves PROVENANCE, not CORRECTNESS, and says nothing about a model writing a plausible sentence around a real link. The honest posture, and the one to require: we cite everything SO IT CAN BE CHECKED, and we can still be wrong -- the source can be wrong and our reading of it can be wrong. (Live catch, oufe.com 2026-07-26: generated copy shipped all four of those promises on a site whose whole positioning is honesty.)
- ILLUSTRATION, NOT AUTHORITY -- risk concentrates in asserting live facts about real named entities; value concentrates in explaining mechanism, and a mechanism can be stated exactly using openly hypothetical figures that cannot be wrong about anybody. Content about a real named entity should lead with the mechanism and present the real case as clearly-marked illustration -- a possibly inaccurate case study -- not as an authoritative account we would have to keep current to stay true. Any interactive tool must say plainly, where the reader will see it, that it can give a wrong answer, and a tool result should carry that caveat inside the result itself so it travels when the result is copied out.'
          ),

          -- (b) extend the judge list
          '(d) does it remove or obscure disclaimers or legal pages. If the fix does not touch user-facing content, claims machinery, or legal surface, approve.',
          '(d) does it remove or obscure disclaimers or legal pages; (e) does it introduce, or leave standing, a promise about our own accuracy, completeness or verification that the platform cannot keep (the overclaimed-reliability class -- remember no scanner will catch this, so this seat is the only control); (f) does it present a real named entity as an authoritative account rather than clearly-marked illustration, or ship a tool without a plain statement that it can be wrong. Where (e) or (f) bites, put the honest replacement wording in notes as a suggestion -- this seat is expected to propose the fix, not only to name the fault. If the fix does not touch user-facing content, claims machinery, or legal surface, approve.'
        )
      ),
      false
    ),
    updated_at = NOW()
WHERE type = 'fix-proposer'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'review_compliance'->'config'->>'prompt_template' NOT LIKE '%OVERCLAIMED RELIABILITY%';

-- ---------------------------------------------------------------------------
-- 2. Fleet-wide writing rule — catch it at generation, not only at review.
--    Mirrors 201's approach (never invent numbers), which is the closest prior
--    art: a rule appended to the writers' standing instructions.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
      to_jsonb(
        (default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template')
        || '

STRICT RULE -- NEVER PROMISE ACCURACY YOU CANNOT GUARANTEE.
Do not write any claim about the reliability, completeness or verification of this site or its content. Never write that every figure is sourced, that every claim is verified, that nothing unsourced appears here, that the content is authoritative or accurate, or that a reader can rely on it. Those are claims about us, they cannot be checked by any tool we run, and they are not true: a citation shows where something came from, it does not prove we read it correctly or that the source is right. If the section calls for a statement about method, say what we DO -- we name our sources and their dates so a reader can check them -- and, where it fits, say plainly that we can still be wrong. Where the copy describes an interactive tool, it must say the tool can give a wrong answer. Where it describes a real named company or case, present it as illustration of a mechanism rather than as a definitive account.'
      ),
      false
    ),
    updated_at = NOW()
WHERE type IN ('page-content-writer', 'content-writer')
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template' IS NOT NULL
  AND default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template' NOT LIKE '%NEVER PROMISE ACCURACY YOU CANNOT GUARANTEE%';

COMMIT;

-- Verify
--   SELECT type,
--          (default_config->'workflow'->'steps'->'review_compliance'->'config'->>'prompt_template'
--             LIKE '%OVERCLAIMED RELIABILITY%') AS has_contract
--     FROM agent_definitions
--    WHERE type IN ('fix-proposer','council-gate') AND is_active
--      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--   -- council-gate becomes true only AFTER the 099 mirror is applied.
--
--   SELECT type,
--          (default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'
--            ->'steps'->'generate_content'->'config'->>'prompt_template'
--             LIKE '%NEVER PROMISE ACCURACY%') AS has_rule
--     FROM agent_definitions WHERE type IN ('page-content-writer','content-writer')
--      AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- APPLIED 2026-07-26. Notes from the live run:
--   * `content-writer` keeps its prompt at
--       {workflow,steps,generate_content,config,prompt_template}
--     NOT under process_sections_loop.sub_workflow like `page-content-writer`.
--     Step 2's WHERE clause therefore matched page-content-writer only; the
--     content-writer half was applied separately with the same text at the
--     shallower path. Both now carry the rule — verify BOTH, by type, and do
--     not assume one path covers both writers.
--   * 099 mirror applied cleanly: drift == ['review_compliance'] exactly, 16
--     seats both sides, routing OK, snapshot taken. Verified afterwards that
--     each roster kept its OWN context block (fix-proposer: diagnosis;
--     council-gate: rationale) — the mirror swaps it, and a hand-patch would
--     have destroyed that.
-- ---------------------------------------------------------------------------
