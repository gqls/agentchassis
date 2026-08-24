-- 601_offer_analyser_acceptance_predicates_HOLD.sql
--
-- features_open/030 §10 v2(d). A finding's acceptance_test is prose, and nothing
-- checks it. This migration lets a finding carry, ALONGSIDE the prose, a small
-- structured predicate a machine can evaluate — and inserts the deterministic
-- gate that refuses any predicate it cannot prove is coupled to the finding.
--
-- WHY, MEASURED RATHER THAN ARGUED. [MEASURED 2026-08-24, live] of the 37
-- acceptance tests this agent has written, THREE sit on work items marked
-- `complete` — page rebuilt, deployed, commit sha in the item's result — while
-- their own stated criterion is refuted by a one-line check over the exact field
-- the test itself names:
--   * webdesign.co.uk / index (item created + completed 08-24): "the meta
--     description must state the zero-data or zero-account promise BEFORE any
--     catalogue count" — served meta opens "Sixty-three browser tools…". The
--     served <meta name="description"> was read at 16:5xZ and is byte-identical
--     to pages.meta_description.
--   * webdesign.co.uk / index (08-22, same page, same clause, different wording).
--   * robot-hands.com / gripper-catalog-index (08-24): "appears as a clickable
--     link in the site header navigation" — absent from the served header.
-- A fourth (webdesign / about, "must not contain 'curated'") is refuted too but
-- sits at wont_fix, so it is not a false green.
--
-- THE TWO RULES THE GATE ENFORCES, and why the feature is dangerous without them
-- (features_open/030 §10 states the trap; this is the answer to it):
--   1. REFUTE-ONLY. A predicate is a NECESSARY condition of the prose test, never
--      a sufficient one. Two thirds of live acceptance tests weld a checkable
--      clause to a judgement clause, and a green tick over the cheap half is
--      worse than the prose it replaced.
--   2. IT MUST REFUTE AT EMISSION or it is discarded. The finding says the page
--      is wrong today, so a condition that expresses the finding must fail today.
--      This is the only property checkable at the moment of writing, and it is
--      what stops the vacuous predicate — a needle present nowhere, a clause
--      already satisfied — being stored as honesty machinery and grading green
--      for ever. Discards are recorded per finding under
--      `acceptance_predicate_rejected`, never dropped silently.
--
-- ⚠ NOTHING AUTOMATED READS `acceptance_predicate` YET, and that is stated
-- rather than implied. The consumer this design expects is a COMPLETION-time
-- check ("the handler reported success — does the item's own predicate still
-- refute?") beside complete_work_item_no_change.go, whose own comment records
-- that grading acceptance_test needs "a producer-side contract change first".
-- This is that change, for one producer. The evaluator is exported
-- (EvaluateAcceptancePredicate) so that consumer is a call, not a rewrite. Until
-- it exists, the value here is the emission-time verdict stored on the item plus
-- the rejection record — and this is the same shape as bugs_closed/335's residual
-- 2, so whoever builds the consumer inherits its condition: the requirement
-- belongs in the ACTION, not in one call site.
--
-- ############################################################################
-- ##  _HOLD — DO NOT APPLY UNTIL A CHASSIS IMAGE CARRYING                    ##
-- ##  verify_acceptance_predicates_action.go HAS ROLLED.                     ##
-- ############################################################################
--
-- WHY THE HOLD, in one sentence: a step naming an action the binary does not
-- register is not a no-op — the workflow validator asks the registry, is told the
-- action is not local, concludes it must be remote, and REJECTS THE WHOLE
-- WORKFLOW with "requires a topic". Config is live on apply; Go is not. IMAGE
-- FIRST, THEN CONFIG. (537 records what that cost elsewhere: 49 items stamped
-- complete carrying a WORKFLOW_INVALID that proves nothing ran.)
--
-- THE `_HOLD` SUFFIX IS LOAD-BEARING: SIDECAR_RE (`_[A-Z][A-Z0-9_]*\.sql$`)
-- excludes this file from `--apply` while still listing it under "Sidecars
-- (hand-run only)". A banner alone would not hold it — the runner reads no
-- comments.
--
-- APPLY IT BY HAND, in this order:
--
--   1. PROBE THE CAPABILITY, NOT THE COMMIT — the registry key is a string
--      literal in the binary, so ask the binary whether it has the action, and
--      run BOTH lines because a probe without a control cannot tell "absent"
--      from "my grep is broken":
--        POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis \
--                -o jsonpath='{.items[0].metadata.name}')
--        kubectl -n ai-persona-system exec "$POD" -- \
--          grep -aq "verify_acceptance_predicates" /proc/1/exe && echo PRESENT-ok
--        kubectl -n ai-persona-system exec "$POD" -- \
--          grep -aq "verify_acceptance_predicates_NOPE" /proc/1/exe && echo CONTROL-FAILED
--      Expect PRESENT-ok and NO second line. Never `strings` (absent from these
--      images), never a discovery grep for "some 40-hex string" (it matches Go's
--      internal digit table), and never 40 zeros as the control (matches
--      everything). ⚠ RUN IT ON EVERY REPLICA: one image tag has shipped
--      several revisions before now (bugs_open/249).
--
--   2. THE FIRST QUERY AFTER APPLYING IS "WHAT DID I BREAK?", NOT "DID IT WORK?"
--        SELECT current_step, status, count(*) FROM orchestration_states
--         WHERE owner_agent_type='offer-analyser'
--           AND created_at > now() - interval '30 minutes'
--         GROUP BY 1,2;
--      ⚠ THE COLUMN IS `owner_agent_type`. There is no `agent_type` column here,
--      and the wrong name ERRORS rather than returning zero — so behind any
--      `2>/dev/null` the damage check reads as "nothing to see".
--      ⚠ A ZERO IS NOT EVIDENCE: terminal rows are reaped in ~24-48h, and
--      `owner_agent_type` names the ORCHESTRATION's owner. Corroborate with
--      llm_call_log keyed on `step_name`, NEVER on `agent_type` (a hand-fired run
--      lands under 'generic'):
--        SELECT created_at, step_name, success FROM llm_call_log
--         WHERE step_name='run_offer_analysis' ORDER BY created_at DESC LIMIT 5;
--
--   3. ONLY THEN look for the benefit, and look for it AT THE ARTEFACT. Fire ONE
--      agent — `scripts/fire-offer-analyser.sh <domain>`, NOT
--      run_improvement_sweep_once.sh, whose triage promotes every `detected` item
--      on the site (111 on webdesign.co.uk as of 08-22, including other lanes').
--      Then read what the run actually did with predicates:
--        SELECT wi.item_type, wi.spec->'acceptance_predicate' AS kept,
--               wi.spec->'acceptance_predicate_rejected' AS refused
--          FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
--         WHERE wi.spec->>'audit_source'='offer-analysis' AND s.domain='<domain>'
--           AND wi.created_at > now() - interval '1 hour';
--      ⚠ THE HONEST READING OF AN ALL-NULL RESULT IS "THE MODEL WROTE NONE",
--      NOT "THE GATE FAILED" — silence is the designed answer for most findings
--      and the key is deliberately absent from the OUTPUT skeleton, so zero
--      adoption on run 1 is a possible and acceptable outcome. What would be a
--      DEFECT is a kept predicate whose `verdict_at_emission` is anything other
--      than "refutes", or a rejection whose reason names a page that is on the
--      surface.
--      ⚠ And do NOT read "kept: 0, refused: 0" as "the gate works". A run
--      containing no predicates passes this gate trivially — the same mistake
--      that cost this lane two days of false confidence on 537's enforcement arm.
--
-- Rollback: 601_offer_analyser_acceptance_predicates_HOLD_ROLLBACK.sql.

BEGIN;

SELECT snapshot_agent('offer-analyser', '601_acceptance_predicates: pre-update');

-- Already-applied gate (the runner reads a RAISE containing 'already').
DO $$
DECLARE done int;
BEGIN
    SELECT count(*) INTO done FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'offer-analyser'
       AND default_config->'workflow'->'steps' ? 'verify_finding_predicates';
    IF done > 0 THEN
        RAISE EXCEPTION '601: already applied — offer-analyser already carries verify_finding_predicates';
    END IF;
END $$;

-- NEEDLE-GATE, AND THE ROW IT RESOLVES IS THE ROW THE UPDATES TOUCH.
-- Resolving the target ONCE into a temp table makes gate and write the same set
-- by construction (537's guardian-seat finding: four agent types carry TWO active
-- rows and only the higher version is ever loaded, so a gate on the broader
-- predicate can pass on the loaded row while the UPDATE rewrites both).
CREATE TEMP TABLE _601_target ON COMMIT DROP AS
SELECT id
  FROM agent_definitions
 WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
   AND type = 'offer-analyser'
   -- The chain this migration splices into, asserted end to end. 537 owns the
   -- first two links; if it has been rolled back or re-cut, this must not write.
   AND default_config->'workflow'->'steps'->'set_audit_source'->>'next_step' = 'verify_ordering_cardinals'
   AND default_config->'workflow'->'steps'->'verify_ordering_cardinals'->>'action' = 'verify_cited_cardinals'
   AND default_config->'workflow'->'steps'->'verify_ordering_cardinals'->>'next_step' = 'write_offer_ordering'
   AND default_config->'workflow'->'steps'->'write_offer_ordering'->>'next_step' = 'write_offer_findings'
   AND default_config->'workflow'->'steps'->'write_offer_findings'->>'action' = 'write_audit_findings'
   AND default_config->'workflow'->'steps'->'write_offer_findings'->'config'->>'findings_field' = 'offer_analysis.result.findings'
   AND default_config->'workflow'->'steps'->'write_offer_findings'->'config'->>'site_id' = 'site_record.site_id'
   AND default_config->'workflow'->'steps'->'run_offer_analysis'->>'output_field' = 'offer_analysis';

DO $$
DECLARE n int; total_active int;
BEGIN
    SELECT count(*) INTO n FROM _601_target;
    IF n <> 1 THEN
        RAISE EXCEPTION
          '601 needle-gate: expected exactly 1 offer-analyser whose chain is set_audit_source -> verify_ordering_cardinals -> write_offer_ordering -> write_offer_findings(findings_field=offer_analysis.result.findings), found % — re-derive against the live workflow before writing', n;
    END IF;

    SELECT count(*) INTO total_active FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'offer-analyser';
    IF total_active <> 1 THEN
        RAISE EXCEPTION
          '601 needle-gate: offer-analyser has % active definition rows, expected 1 — only the highest version is ever loaded, so decide which row is real before writing', total_active;
    END IF;
END $$;

-- PROMPT ANCHOR GATE. `replace()` on a missed anchor is a SILENT NO-OP that still
-- reports UPDATE 1 — the migration would ship the gate step with the model never
-- told the contract, and report success. Asserted to occur EXACTLY ONCE: zero
-- means the prompt has drifted, more than one means the rule would be spliced in
-- twice. Counted by length difference, which is exact and needs no regex.
DO $$
DECLARE prompt text; anchor text; occurrences int;
BEGIN
    anchor := 'Bad: "The homepage should be more benefit-led".';

    SELECT ad.default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt'
      INTO prompt
      FROM agent_definitions ad JOIN _601_target t ON t.id = ad.id;

    IF prompt IS NULL THEN
        RAISE EXCEPTION '601 prompt-anchor gate: run_offer_analysis carries no prompt';
    END IF;

    occurrences := (length(prompt) - length(replace(prompt, anchor, ''))) / length(anchor);
    IF occurrences <> 1 THEN
        RAISE EXCEPTION
          '601 prompt-anchor gate: the acceptance_test sentence occurs % time(s), expected exactly 1 — re-derive the anchor before splicing', occurrences;
    END IF;
END $$;

-- 1. Insert the gate step, rewire write_offer_ordering onto it, and point the
--    findings write at the CHECKED array rather than the raw model output.
--
--    ⚠ THE THIRD jsonb_set IS WHAT MAKES THE GATE BINDING. Without the second
--    the step exists and never runs; without the third it runs and the write
--    ignores it. Both halves have been shipped broken elsewhere on this estate.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(
             default_config,
             '{workflow,steps,verify_finding_predicates}',
             jsonb_build_object(
               'action', 'verify_acceptance_predicates',
               'config', jsonb_build_object(
                   'site_id',        'site_record.site_id',
                   'findings_field', 'offer_analysis.result.findings'),
               'next_step', 'write_offer_findings',
               'description', 'features_open/030 v2(d). Validates each finding''s optional acceptance_predicate against the page metadata it names, and keeps ONLY those that already refute the page as it stands — a predicate that cannot be shown to fail today cannot be shown to express this finding. Refusals are recorded per finding under acceptance_predicate_rejected. Never removes a finding and never fails the run: the prose finding is the valuable part. NO error_step by design, same as the write it guards.',
               'output_field', 'predicates_checked'),
             true),
           '{workflow,steps,write_offer_ordering,next_step}',
           '"verify_finding_predicates"'::jsonb,
           false),
         '{workflow,steps,write_offer_findings,config,findings_field}',
         '"predicates_checked.findings"'::jsonb,
         false),
       updated_at = now()
 WHERE id IN (SELECT id FROM _601_target);

-- 2. Tell the model the contract. Appended to the sentence that already defines
--    acceptance_test, so the optional structured half sits where the prose half
--    is defined rather than in a distant block.
--
--    ⚠ DELIBERATELY NOT ADDED TO THE `OUTPUT` SKELETON. Every key in that
--    skeleton reads as required — the prompt's own closing rule says so for the
--    ordering object — and this key must default to ABSENT. Silence is the
--    correct answer for most findings, so the emission arm is invited in prose
--    and never templated. The cost is that adoption may be zero on run 1; that
--    is the safe direction and it is measurable (see step 3 of the runbook).
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,run_offer_analysis,config,prompt}',
         to_jsonb(replace(
           default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt',
           'Bad: "The homepage should be more benefit-led".',
           $rule$Bad: "The homepage should be more benefit-led".

ACCEPTANCE PREDICATE — optional, per finding, and usually omitted. If a NECESSARY condition of your acceptance_test can be stated mechanically over ONE page's meta description or title, add "acceptance_predicate" to that finding. If it cannot, OMIT THE KEY ENTIRELY. Most findings will omit it, and that is the right answer: a predicate over a clause that needs judgement ("reads as a generic contact-us button", "consistent with the one-founder model") is worse than no predicate at all, because it grades green while the judgement clause is still unmet.

Three rules. (1) A predicate can only ever REFUTE. Satisfying it never means the test is met — your prose acceptance_test stays the authority. (2) It must be FALSE OF THE PAGE AS IT IS TODAY: your finding says this page is wrong now, so the condition you write must fail now. One that already holds is removed mechanically after you answer, with a note saying why. (3) Only these three shapes, only these two fields, and no other keys. Text is matched case-insensitively at word boundaries.

{"type": "text_absent", "page": "about", "field": "meta_description", "values": ["curated", "hand-picked"]} — refuted if any listed phrase is present.
{"type": "text_present", "page": "index", "field": "title", "values": ["Kubernetes", "Kafka", "Postgres"], "min": 2} — refuted if fewer than "min" of them are present ("min" defaults to 1).
{"type": "text_order", "page": "index", "field": "meta_description", "before": ["no account", "no upload"], "after": ["$cardinal"]} — refuted if none of "before" is present, or if anything in "after" appears earlier than all of them. The reserved word "$cardinal", valid only inside "after", means the first quantity of any kind, in digits or in words — "63" and "Sixty-three" both count.

"page" defaults to the finding's own page. You are shown every page's title and meta description in THE OFFER SURFACE above: write predicates over those only, never over body copy you have not been shown.$rule$)),
         false),
       updated_at = now()
 WHERE id IN (SELECT id FROM _601_target);

-- VERIFY. A DO/RAISE block, not a set of SELECTs: ON_ERROR_STOP does not fire on
-- a non-empty result set, so a verify block made of SELECTs cannot stop the
-- COMMIT.
DO $$
DECLARE
    cfg    jsonb;
    step   jsonb;
    prompt text;
    n_steps int;
    mentions int;
BEGIN
    SELECT ad.default_config INTO cfg
      FROM agent_definitions ad JOIN _601_target t ON t.id = ad.id;
    step   := cfg->'workflow'->'steps'->'verify_finding_predicates';
    prompt := cfg->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt';

    IF step IS NULL OR step->>'action' <> 'verify_acceptance_predicates' THEN
        RAISE EXCEPTION '601 verify: gate step missing or wrong action (%)', step->>'action';
    END IF;
    IF step->>'next_step' <> 'write_offer_findings' THEN
        RAISE EXCEPTION '601 verify: the gate does not hand on to write_offer_findings';
    END IF;
    IF step->'config'->>'findings_field' <> 'offer_analysis.result.findings'
       OR step->'config'->>'site_id' <> 'site_record.site_id' THEN
        RAISE EXCEPTION '601 verify: the gate reads the wrong inputs (%)', step->'config';
    END IF;

    -- REACHABLE, and BINDING.
    IF cfg->'workflow'->'steps'->'write_offer_ordering'->>'next_step' <> 'verify_finding_predicates' THEN
        RAISE EXCEPTION '601 verify: write_offer_ordering still bypasses the gate';
    END IF;
    IF cfg->'workflow'->'steps'->'write_offer_findings'->'config'->>'findings_field' <> 'predicates_checked.findings' THEN
        RAISE EXCEPTION '601 verify: the findings write still reads the UNCHECKED model output';
    END IF;

    SELECT count(*) INTO n_steps FROM jsonb_object_keys(cfg->'workflow'->'steps');
    IF n_steps <> 10 THEN
        RAISE EXCEPTION '601 verify: expected 10 workflow steps after the insert, found %', n_steps;
    END IF;

    -- The prompt half, asserted on three separate strings: the invitation, the
    -- refute-only rule and the reserved needle. One LIKE over one phrase would
    -- pass on a partial splice.
    IF prompt NOT LIKE '%ACCEPTANCE PREDICATE — optional, per finding%' THEN
        RAISE EXCEPTION '601 verify: the predicate contract did not splice into the prompt';
    END IF;
    IF prompt NOT LIKE '%can only ever REFUTE%' THEN
        RAISE EXCEPTION '601 verify: the refute-only rule is missing from the prompt';
    END IF;
    IF prompt NOT LIKE '%$cardinal%' THEN
        RAISE EXCEPTION '601 verify: the reserved needle is missing from the prompt';
    END IF;
    -- And the key must NOT be in the OUTPUT skeleton, where it would read as
    -- required. This is an assertion about what is ABSENT, which is the half a
    -- verify block usually forgets.
    -- Counted by length difference rather than regexp_count(): that function is
    -- PostgreSQL 15+, this idiom works everywhere, and the rest of this file
    -- already uses it. A version-dependent assertion inside a verify block fails
    -- the migration for the wrong reason.
    mentions := (length(prompt) - length(replace(prompt, 'acceptance_predicate', ''))) / length('acceptance_predicate');
    IF mentions <> 1 THEN
        RAISE EXCEPTION
          '601 verify: "acceptance_predicate" occurs % times in the prompt, expected exactly 1 (the contract paragraph) — it must NOT appear in the OUTPUT skeleton, where every key reads as required', mentions;
    END IF;

    RAISE NOTICE '601 OK: write_offer_ordering -> verify_finding_predicates -> write_offer_findings, the write reads predicates_checked.findings, and the prompt carries the contract once';
END $$;

COMMIT;
