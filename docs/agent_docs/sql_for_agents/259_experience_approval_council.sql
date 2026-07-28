-- ============================================================================
-- 259_experience_approval_council.sql — per-experience approval, the gate the
-- register's lifecycle has been blocked at since 2026-07-24.
--
-- WHY THIS EXISTS
--   The owner ruled on 2026-07-24 that approval is PER EXPERIENCE, via a
--   council. It was never built. The consequence, live until now: every entry is
--   `draft`, so every per-site fork is `proposed`, so nothing can reach
--   `verified` or `proven` at all. The lifecycle had no first step.
--
-- WHAT IT DECIDES, AND WHAT IT CANNOT
--   It returns a verdict on ONE register entry and persists it as a
--   `council_report` in diagnosis_artifacts, using the same deterministic
--   decision action the code council uses (`diagnose_council_decide`:
--   veto → rejected, any objection → revise, else approved). It does NOT write
--   the entry's status — nothing here can set `approved`. Applying a verdict is
--   a separate action, deliberately, so that this council can be run and read
--   before it is given the power to change anything.
--
-- THE SEATS, AND WHY EACH ONE
--   Every seat below exists because of something that actually went wrong in
--   this workstream in the last three days, not because it sounded like a good
--   category:
--
--   observable_outcome  — the harvest's founding correction: a contract must be
--     able to say what is NOT a control. A clause with no observable outcome is
--     how "there is a button here" gets recorded as behaviour.
--
--   honesty (HARD VETO) — the register exists to stop a page claiming something
--     it has not got. This seat holds the veto because an entry that legitimises
--     a dishonest surface is worse than no entry: it would be copied.
--
--   checkability — the sharpest lesson of 2026-07-28. `at_least_two_cards`
--     asserted "at least one". `rows_rendered` asserted only that the container
--     existed. A check whose NAME promises more than it asserts is worse than a
--     missing check, because it reads as covered. This seat's whole job is to
--     catch that.
--
--   deferral_honesty — deferral is the escape hatch that keeps an unassertable
--     clause in the record instead of deleting it. That is only worth having if
--     it cannot become a way to defer everything hard and ship an entry that
--     asserts nothing. Migration 230 makes the extreme case unrepresentable;
--     this seat judges the middle.
--
--   prior_art — nine entries in, the duplicate risk is real and the register's
--     own case rests on clauses recurring across DIFFERENT components. An entry
--     that is a variant of an existing one should reference the invariant, not
--     restate it.
--
-- INPUT (from the trigger, which reads the STORED row, not the harvest file):
--   input_data: { fix_correlation_id, pattern_name, entry_json, register_summary }
--
-- KNOWN GAP, stated rather than hidden: the council reviews the entry as it
-- stood at SUBMISSION. If the entry changes mid-round the verdict describes the
-- old one. The demotion rule in write_experience_pattern limits the damage (a
-- contract change returns an approved entry to draft), but the apply step must
-- still re-check the entry's updated_at before acting on a verdict. That check
-- belongs in the apply action, which is not in this file.
-- ============================================================================

BEGIN;

INSERT INTO agent_definitions (
  id, type, display_name, description, category, status, is_active, default_config
)
SELECT
  gen_random_uuid(),
  'experience-approval-council',
  'Experience approval council',
  'Reviews ONE experience-register entry and returns a verdict (approved/revise/rejected) using the same deterministic decision action as the code council. Five seats: observable outcome, honesty (hard veto), checkability, deferral honesty, prior art. It records a verdict; it does not change the entry.',
  'documentation',
  'experimental',
  true,
  $cfg${
  "workflow": {
    "start_step": "review_observable_outcome",
    "processing_mode": "orchestrator",
    "timeout_seconds": 900,
    "steps": {

      "review_observable_outcome": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {"provider": "anthropic", "model": "claude-sonnet-5", "max_tokens": 6000, "api_key_env_var": "ANTHROPIC_API_KEY"},
          "temperature": 0.0,
          "output_format": "json",
          "tolerate_truncation": true,
          "error_step": "complete_invalid",
          "input_fields": ["input_data"],
          "prompt_template": "# Council reviewer: OBSERVABLE OUTCOME\n\nYou judge ONE entry in a library of reusable user-experience contracts. You change nothing; you judge.\n\nAn entry describes what a component or journey must DO, in terms specific enough that a machine could check it. Your single question for every clause: **could a machine observe whether this happened?**\n\nJudge:\n(a) every clause in `contract` names an `outcome` that is observable — a page loaded, an element appeared, an address changed. \"The user understands the offer\" is not observable. \"There is a button\" is not an outcome at all.\n(b) a clause that NAVIGATES names a destination ROLE, never a URL. A concrete address in a base entry is a value that re-applies on every render and cannot be overridden per site.\n(c) clauses that say what must NOT happen (in `states`, `must_not`) are as important as the positive ones — this library exists because 'a control that cannot do anything must not be presented as one' was reinvented in six places. An entry that only describes success is incomplete.\n(d) an entry with an EMPTY contract is legitimate ONLY if its behaviour is genuinely not visitor-driven and lives in `automatic_triggers` — check, do not assume.\n\nVerdicts: approve, object (fixable), veto (never — you hold no veto).\n\n## The entry\n{{.input_data.entry_json}}\n\n## Output — ONLY this JSON\n{\"reviewer\": \"observable_outcome\", \"verdict\": \"approve|object\", \"objections\": [{\"edit\": 1, \"problem\": \"...\", \"severity\": \"low|medium|high\"}], \"missing\": [], \"notes\": \"...\"}"
        },
        "next_step": "review_honesty",
        "output_field": "review_observable_outcome",
        "description": "Seat 1 — every clause names an outcome a machine could observe; navigation names a role, not a URL."
      },

      "review_honesty": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {"provider": "anthropic", "model": "claude-sonnet-5", "max_tokens": 6000, "api_key_env_var": "ANTHROPIC_API_KEY"},
          "temperature": 0.0,
          "output_format": "json",
          "tolerate_truncation": true,
          "error_step": "complete_invalid",
          "input_fields": ["input_data"],
          "prompt_template": "# Council reviewer: HONESTY (hard-veto holder)\n\nYou judge ONE entry in a library of reusable user-experience contracts. This library exists to stop a page claiming something it has not got. An entry that LEGITIMISES a dishonest surface is worse than no entry, because it will be copied to every site that adopts it. That is why you hold the veto.\n\nVeto if the entry would permit, or fail to forbid, any of these:\n(a) a control that cannot do anything being presented as a control — a placeholder href, a tab stop, a rendered button with nowhere to go;\n(b) a number, count, or status rendered as fact when it is not true by construction — this platform has a live history of fabricated participation counts and invented figures;\n(c) a progress indicator, timer, or verdict that advances on anything other than a real result;\n(d) a failure path that looks like a success path — a degraded state that renders as though it worked. A failure path is where dishonesty enters a page, so an entry whose behaviour can fail and which names NO degraded state is incomplete, and you should object.\n\nObject (do not veto) where the entry is merely silent on something that could be added. Veto where it would actively license a false impression.\n\nIf you veto, name in notes the smallest change that would make it honest — a veto with no route out is a dead end.\n\n## The entry\n{{.input_data.entry_json}}\n\n## Output — ONLY this JSON\n{\"reviewer\": \"honesty\", \"verdict\": \"approve|object|veto\", \"objections\": [{\"edit\": 1, \"problem\": \"...\", \"severity\": \"low|medium|high\"}], \"missing\": [], \"notes\": \"...\"}"
        },
        "next_step": "review_checkability",
        "output_field": "review_honesty",
        "description": "Seat 2 — HARD VETO. Would this entry license a page to claim something it has not got?"
      },

      "review_checkability": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {"provider": "anthropic", "model": "claude-sonnet-5", "max_tokens": 6000, "api_key_env_var": "ANTHROPIC_API_KEY"},
          "temperature": 0.0,
          "output_format": "json",
          "tolerate_truncation": true,
          "error_step": "complete_invalid",
          "input_fields": ["input_data"],
          "prompt_template": "# Council reviewer: CHECKABILITY\n\nYou judge whether this entry's TESTS actually assert what its CLAUSES claim. This seat exists because of three real defects found on 2026-07-28, all of which passed every other check:\n\n  * a test named `at_least_two_cards` that asserted \"at least one\" — the platform reads no minimum, so the count was silently ignored;\n  * a test named `rows_rendered` that asserted only that the CONTAINER existed — which another test already asserted. Two tests, one fact, and the informative name was on the empty one;\n  * a test referring to a value nothing defined, which would have typed the placeholder's own name into the input and passed.\n\nSo your question is never \"is there a test?\" It is **\"does this test fail when the clause is violated?\"**\n\nJudge:\n(a) NAME vs ASSERTION — for every check, does its id promise more than its type and fields deliver? A check whose name describes a quantity, an ordering, or an attribute, but whose type only confirms presence, is the defect above. Say so explicitly.\n(b) REDUNDANCY — do two checks assert the same fact? One of them is decoration.\n(c) COVERAGE — for each contract clause and each `must_not`, is there a check that would fail if it were violated? Name the clauses with no such check.\n(d) THE NEGATIVE CLAUSES — 'must not be a control', 'must not render', 'must not advance' are the most valuable and the hardest to check. If they are covered only by a deferred check, say so plainly: that is honest, but it means the entry's most important rule is unverified.\n\nBeing unable to check something is NOT a reason to object, provided it is declared. Claiming to check something you do not, is.\n\nVerdicts: approve, object. You hold no veto.\n\n## The entry\n{{.input_data.entry_json}}\n\n## Output — ONLY this JSON\n{\"reviewer\": \"checkability\", \"verdict\": \"approve|object\", \"objections\": [{\"edit\": 1, \"problem\": \"...\", \"severity\": \"low|medium|high\"}], \"missing\": [], \"notes\": \"...\"}"
        },
        "next_step": "review_deferral_honesty",
        "output_field": "review_checkability",
        "description": "Seat 3 — does each test FAIL when its clause is violated? Catches a name that promises more than the assertion delivers."
      },

      "review_deferral_honesty": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {"provider": "anthropic", "model": "claude-sonnet-5", "max_tokens": 6000, "api_key_env_var": "ANTHROPIC_API_KEY"},
          "temperature": 0.0,
          "output_format": "json",
          "tolerate_truncation": true,
          "error_step": "complete_invalid",
          "input_fields": ["input_data"],
          "prompt_template": "# Council reviewer: DEFERRAL HONESTY\n\nThis library lets an entry mark a check DEFERRED when the platform cannot execute it. That exists so an unassertable clause stays in the record instead of being deleted — deleting it is how a pattern comes to look fully checked when its most important rule is not checked at all.\n\nIt is therefore also the obvious way to game the library: defer everything hard, and ship an entry that asserts nothing while looking thorough. The database already refuses to approve an entry with ZERO executable checks. Your job is the middle ground it cannot judge.\n\nJudge:\n(a) is each deferral's reason SPECIFIC and TRUE — does it name the missing capability, or is it vague ('not supported yet')? A deferral without a real reason is a check nobody wrote.\n(b) is anything deferred that our platform CAN in fact do? The executable check types are: selector_exists, selector_count, interaction, asset_loads, page_status_ok (static tier); plus no_horizontal_overflow and no_console_errors in a real browser. A clause deferred when one of these would cover it is an evasion.\n(c) RATIO AND IMPORTANCE — count executable versus deferred, and say which side the entry's MOST IMPORTANT clause falls on. An entry with nine trivial passing checks and its central rule deferred is weaker than one with two real ones.\n(d) is a deferred check WRITTEN OUT properly — would it be runnable the day the capability lands, or is it a placeholder that would need designing from scratch?\n\nVerdicts: approve, object. You hold no veto.\n\n## The entry\n{{.input_data.entry_json}}\n\n## Register context\n{{.input_data.register_summary}}\n\n## Output — ONLY this JSON\n{\"reviewer\": \"deferral_honesty\", \"verdict\": \"approve|object\", \"objections\": [{\"edit\": 1, \"problem\": \"...\", \"severity\": \"low|medium|high\"}], \"missing\": [], \"notes\": \"...\"}"
        },
        "next_step": "review_prior_art",
        "output_field": "review_deferral_honesty",
        "description": "Seat 4 — is each deferral specific, true, and not covering something we could actually check?"
      },

      "review_prior_art": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {"provider": "anthropic", "model": "claude-sonnet-5", "max_tokens": 6000, "api_key_env_var": "ANTHROPIC_API_KEY"},
          "temperature": 0.0,
          "output_format": "json",
          "tolerate_truncation": true,
          "error_step": "complete_invalid",
          "input_fields": ["input_data"],
          "prompt_template": "# Council reviewer: PRIOR ART AND REUSE\n\nThe register's whole case is that CLAUSES recur even when components do not. So the failure mode you guard against is not two similar entries — it is the same rule written out twice instead of referenced once.\n\nJudge, against the register contents supplied below:\n(a) DUPLICATION — is this entry a restatement of an existing one? Name it. A genuine variant is fine; an entry that differs only in wording is not.\n(b) INVARIANTS — the library holds shared rules separately, and entries reference them by name in `requires_invariant`. Does this entry RESTATE a rule that already exists as an invariant instead of referencing it? That is the drift this library was built to end, so say so.\n(c) COMPOSITION — a journey should reference the component contracts it depends on (`requires_component_contract`) rather than repeating their clauses.\n(d) NAMING — is the name accurate and specific? A name that overstates what the entry does is a defect: entries are selected by name.\n\nDo NOT claim an absence you cannot see. You are given the register's current contents; if something is not there, say 'not present in the supplied register contents', never 'does not exist'.\n\nVerdicts: approve, object. You hold no veto.\n\n## The entry\n{{.input_data.entry_json}}\n\n## The register's current contents\n{{.input_data.register_summary}}\n\n## Output — ONLY this JSON\n{\"reviewer\": \"prior_art\", \"verdict\": \"approve|object\", \"objections\": [{\"edit\": 1, \"problem\": \"...\", \"severity\": \"low|medium|high\"}], \"missing\": [], \"notes\": \"...\"}"
        },
        "next_step": "council_decide",
        "output_field": "review_prior_art",
        "description": "Seat 5 — duplication, restated invariants, and a name that overstates what the entry does."
      },

      "council_decide": {
        "action": "diagnose_council_decide",
        "config": {
          "fix_correlation_id": "{{.input_data.fix_correlation_id}}",
          "max_rounds": 3,
          "hard_veto_from": ["honesty"],
          "review_fields": [
            "review_observable_outcome",
            "review_honesty",
            "review_checkability",
            "review_deferral_honesty",
            "review_prior_art"
          ],
          "error_step": "complete_invalid"
        },
        "next_step": "complete",
        "output_field": "council_result",
        "description": "The SAME deterministic decision action the code council uses: veto -> rejected, any objection -> revise, else approved. Persists a council_report. Reusing it rather than writing a second decision rule is deliberate — two rules would drift and then disagree while sharing a vocabulary."
      },

      "complete": {
        "action": "complete_workflow",
        "config": {"output_fields": ["council_result"]}
      },

      "complete_invalid": {
        "action": "complete_workflow",
        "config": {"output_fields": ["council_result"]}
      }
    }
  }
}$cfg$::jsonb
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions
  WHERE type = 'experience-approval-council'
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
);

DO $guard$
DECLARE
    cfg jsonb; steps jsonb; seats text[]; s text;
BEGIN
    SELECT default_config INTO cfg FROM agent_definitions
     WHERE type = 'experience-approval-council'
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF cfg IS NULL THEN
        RAISE EXCEPTION '259: the council definition was not created';
    END IF;
    steps := cfg->'workflow'->'steps';

    seats := ARRAY['review_observable_outcome','review_honesty','review_checkability',
                   'review_deferral_honesty','review_prior_art'];
    FOREACH s IN ARRAY seats LOOP
        IF steps->s IS NULL THEN
            RAISE EXCEPTION '259: seat % is missing', s;
        END IF;
        -- A seat whose output_field is not in review_fields is a seat that runs,
        -- costs money, and is never counted. That is worse than not seating it.
        IF NOT (steps->'council_decide'->'config'->'review_fields' @> to_jsonb(s)) THEN
            RAISE EXCEPTION '259: seat % runs but is not in review_fields — its opinion would be discarded', s;
        END IF;
    END LOOP;

    IF NOT (steps->'council_decide'->'config'->'hard_veto_from' @> '"honesty"'::jsonb) THEN
        RAISE EXCEPTION '259: the honesty seat must hold the veto — it is the only one whose failure mode is uncorrectable downstream';
    END IF;

    IF steps->'council_decide'->>'action' <> 'diagnose_council_decide' THEN
        RAISE EXCEPTION '259: the decision must use the shared deterministic action, not a second rule';
    END IF;
END
$guard$;

COMMIT;

-- Verify
SELECT type, status, is_active,
       jsonb_array_length(default_config->'workflow'->'steps'->'council_decide'->'config'->'review_fields') AS seats,
       default_config->'workflow'->'steps'->'council_decide'->'config'->'hard_veto_from' AS veto_holders
FROM agent_definitions
WHERE type = 'experience-approval-council'
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Rollback (hand-run):
--   UPDATE agent_definitions SET is_active=false, deleted_at=now()
--    WHERE type='experience-approval-council' AND COALESCE(is_snapshot,false)=false;
