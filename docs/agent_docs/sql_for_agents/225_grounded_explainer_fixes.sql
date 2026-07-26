-- ============================================================================
-- 225_grounded_explainer_fixes.sql — three fixes from the lane's first live run
--
-- The first run of `grounded-explainer` (oufe.com, Part 26A cross-class cram
-- down, corr 8896cc75) acquired and machine-verified **14 citations from
-- legislation.gov.uk and Court of Appeal commentary** — the statutory majority,
-- both cram-down conditions, the definition of the relevant alternative, the
-- Adler appeal. The acquisition half worked exactly as designed.
--
-- Then two defects, one of them the more interesting kind.
--
-- FIX 1 — THE COMPOSER NEVER SAW THE FACTS.
--   `verify_and_register_citations` returns a REGISTRATION SUMMARY:
--       {"registered": ["CIT-a7f91f88754d560", ...]}
--   — opaque ids, not claims. The composer prompt interpolated {{.registration}}
--   and therefore received a list of identifiers with no content whatsoever.
--
--   **And it behaved correctly.** It refused to assert anything, returned
--   `sources_used: []`, and listed as unverifiable gaps exactly the facts that
--   had just been registered: "the statutory majority required for a class to be
--   treated as approving a plan", "the precise legal definition of the no worse
--   off test", "whether cross-class cramdown is available at all". Every one of
--   those was sitting in evidence_base, quote-verified, at that moment.
--
--   That is the design working. A writer starved of facts produced honest gaps
--   instead of confident law. The bug cost a run; the alternative failure mode —
--   a writer that fills gaps from memory — costs a wrong statement of statute on
--   a live page. Recorded because the failure being SAFE is the point.
--
--   Fix: read the register back after registration and hand the composer the
--   facts themselves.
--
-- FIX 2 — THE AUDIT'S FAILURE DESTROYED THE WORK IT WAS AUDITING.
--   `audit_grounding` hit its 6000-token cap:
--       response truncated: stop_reason=max_tokens (output_tokens=6000 reached
--       the configured cap, 0 chars recovered)
--   and, having no error_step, failed the whole orchestration. The composed
--   draft — 6,397 characters, recoverable only by hand out of collected_data —
--   was lost from the workflow's own output.
--
--   Two things wrong, fixed separately:
--     (a) the cap was too small for the job (a sentence-by-sentence pass over a
--         6k-character draft plus 14 facts). Raised to 12000, and the prompt now
--         asks only for the problems rather than an enumeration of every
--         sentence.
--     (b) **a check that fails must not destroy the thing it was checking.**
--         An error_step now routes a failed audit straight to the review item,
--         flagged as unaudited so a human reads it more carefully — not less.
--         Same discipline as tool-generator's doc steps, which carry
--         `error_step: complete` so a docs failure can never fail tool creation.
--
--   This is the `output_tokens == max_tokens means the completion was CUT` rule
--   from CLAUDE.md, met in a new place: not an artifact truncated into the DB,
--   but a *verification* step whose truncation took a good draft down with it.
--
-- FIX 3 — the audit reads the register too, for the same reason as fix 1.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- 1. new step: read the register back, between registration and composition ---
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,load_evidence}',
      jsonb_build_object(
        'action', 'read_site_spec',
        'config', jsonb_build_object('site_id', 'site_record.site_id'),
        'next_step', 'compose_explainer',
        'output_field', 'evidence',
        'description', 'Read the register back. verify_and_register returns ids, not facts — the composer needs the claims themselves.'
      ),
      true
    ),
    updated_at = NOW()
WHERE type = 'grounded-explainer' AND deleted_at IS NULL;

-- route registration -> load_evidence -> compose
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,verify_and_register,next_step}',
      '"load_evidence"'::jsonb, false
    ),
    updated_at = NOW()
WHERE type = 'grounded-explainer' AND deleted_at IS NULL;

-- 2. composer: read the facts, not the receipt --------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,compose_explainer,config,input_fields}',
      '["input_data","evidence","registration","site_record"]'::jsonb, false
    ),
    updated_at = NOW()
WHERE type = 'grounded-explainer' AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,compose_explainer,config,prompt_template}',
      to_jsonb(replace(
        default_config->'workflow'->'steps'->'compose_explainer'->'config'->>'prompt_template',
        'These survived machine verification — each quote was found in the live source:

{{.registration}}',
        'These survived machine verification — each quote was re-fetched and found in the live source. Every one carries the exact wording of its source, so you can state what it says precisely:

{{.evidence.specs.evidence_base.facts}}

If that list is empty or contains no usable claim, say so in "gaps" and write only mechanism — do NOT fall back on what you remember.'
      )), false
    ),
    updated_at = NOW()
WHERE type = 'grounded-explainer' AND deleted_at IS NULL;

-- 3. audit: bigger cap, reads the facts, and cannot take the draft down with it
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,audit_grounding,config,ai_service,max_tokens}',
      '12000'::jsonb, false
    ),
    updated_at = NOW()
WHERE type = 'grounded-explainer' AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,audit_grounding,config,input_fields}',
      '["draft","evidence","input_data"]'::jsonb, false
    ),
    updated_at = NOW()
WHERE type = 'grounded-explainer' AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,audit_grounding,config,error_step}',
      '"create_review_item"'::jsonb, true
    ),
    updated_at = NOW()
WHERE type = 'grounded-explainer' AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config, '{workflow,steps,audit_grounding,config,prompt_template}',
      to_jsonb(replace(
        replace(
          default_config->'workflow'->'steps'->'audit_grounding'->'config'->>'prompt_template',
          '## The verified facts that were available to the writer
{{.registration}}',
          '## The verified facts that were available to the writer
{{.evidence.specs.evidence_base.facts}}'
        ),
        'Return ONLY:',
        'BE BRIEF. Report ONLY the problems — do not enumerate the sentences that are fine, and do not restate the draft. At most 12 ungrounded items; if there are more, report the 12 worst and say so in notes. A long audit that gets truncated protects nobody.

Return ONLY:'
      )), false
    ),
    updated_at = NOW()
WHERE type = 'grounded-explainer' AND deleted_at IS NULL;

COMMIT;

-- Verify
--   SELECT default_config->'workflow'->'steps'->'verify_and_register'->>'next_step' AS after_verify,
--          default_config->'workflow'->'steps'->'audit_grounding'->'config'->>'error_step' AS audit_error_step,
--          default_config->'workflow'->'steps'->'audit_grounding'->'config'->'ai_service'->>'max_tokens' AS audit_cap,
--          (default_config->'workflow'->'steps'->'compose_explainer'->'config'->>'prompt_template'
--             LIKE '%evidence_base.facts%') AS composer_reads_facts
--     FROM agent_definitions WHERE type='grounded-explainer' AND deleted_at IS NULL;
