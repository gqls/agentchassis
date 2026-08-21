-- 537_offer_analyser_cardinal_attribution_gate_HOLD.sql
--
-- bugs_open/335. The offer-analyser wrote, at RANK 1 on leopardessconsulting.co.uk,
-- "the same stack that runs eight live sites built by this team", stamped
-- from_field: "trust_threshold". The true count is 23; the number came off a page
-- meta_description carried in the offer surface; and the cited premise field
-- contains no number at all. from_field is this lane's own honesty machinery, so
-- the field built to prove sourcing vouched for a number the premise never held.
--
-- This migration does two things: it inserts the deterministic gate step between
-- the analysis and the write, and it tells the model the rule in the prompt. Both
-- halves are wanted — the prompt reduces how often it happens, the gate makes it
-- unrepresentable.
--
-- ############################################################################
-- ##  _HOLD — DO NOT APPLY UNTIL A CHASSIS IMAGE CARRYING                    ##
-- ##  verify_cited_cardinals_action.go HAS ROLLED.                           ##
-- ##  The gate commit is d79e4243c.                                          ##
-- ############################################################################
--
-- WHY THE HOLD. A step naming an action the binary does not register is not a
-- no-op: the workflow validator asks the registry whether the action is local,
-- is told "no", concludes it must be remote, and rejects the whole workflow with
-- "requires a topic" (registry_parity_test.go records what that cost the
-- fix_forced_text_colors path — 49 items stamped complete carrying a
-- WORKFLOW_INVALID that proves nothing ran). Config is live on apply; Go is not.
-- IMAGE FIRST, THEN CONFIG.
--
-- The `_HOLD` suffix is load-bearing: SIDECAR_RE (`_[A-Z][A-Z0-9_]*\.sql$`)
-- excludes this file from `--apply` while STILL listing it under "Sidecars
-- (hand-run only)". A banner alone would not hold it — the runner does not read
-- comments.
--
-- APPLY IT BY HAND, in this order:
--
--   1. Confirm the running chassis registers the action. Ask the ARTEFACT, not
--      git, and per SERVICE rather than per fleet:
--        kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 \
--          | grep -m1 'build provenance'
--        git merge-base --is-ancestor d79e4243c <the stamped sha> && echo SAFE-TO-APPLY
--
--      An EMPTY grep means "the startup line has scrolled", NOT "unstamped".
--      Fall back to the binary probe and ALWAYS run a control in the same breath
--      (a plausible fake sha that must be ABSENT, and one that must be PRESENT):
--        kubectl -n ai-persona-system exec <pod> -- grep -aq "<sha>" /proc/1/exe
--      Never `strings` (absent from these images) and never a discovery grep for
--      "some 40-hex string" (it matches Go's internal digit table and returns the
--      same wrong answer on every service).
--
--      BETTER STILL, PROBE THE CAPABILITY RATHER THAN THE COMMIT. What actually
--      matters is not "which sha built this" but "does this binary REGISTER the
--      action" — and the registry map key is a string literal in the binary, so
--      it can be asked directly. Run BOTH lines; the second is the control, and
--      a probe without one cannot tell "absent" from "my grep is broken":
--        kubectl -n ai-persona-system exec <pod> -- \
--          grep -aq "verify_cited_cardinals" /proc/1/exe      && echo PRESENT-ok
--        kubectl -n ai-persona-system exec <pod> -- \
--          grep -aq "verify_cited_cardinals_NOPE" /proc/1/exe && echo CONTROL-FAILED
--      Expect PRESENT-ok and NO second line. The control is a plausible name that
--      must be ABSENT — never 40 zeros, which matches everything.
--
--   2. THE FIRST QUERY AFTER APPLYING IS "WHAT DID I BREAK?", NOT "DID IT WORK?".
--      An unregistered action name fails the whole offer-analyser workflow, not
--      just this step:
--        SELECT current_step, status, count(*) FROM orchestration_states
--         WHERE agent_type='offer-analyser'
--           AND created_at > now() - interval '30 minutes'
--         GROUP BY 1,2;
--
--   3. ONLY THEN look for the benefit, and look for it AT THE ARTEFACT. The
--      re-proof is a re-run against leopardessconsulting.co.uk, whose premise is
--      unchanged and whose pages still carry the phrase, so the input that
--      produced the defect is still live:
--        SELECT lw->>'rank', lw->>'from_field', lw->>'point'
--          FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
--               jsonb_array_elements(ss.data->'lead_with') lw
--         WHERE ss.aspect='offer_ordering' AND ss.is_current
--           AND s.domain='leopardessconsulting.co.uk'
--         ORDER BY (lw->>'rank')::int;
--      POSITIVE: no cardinal in any surviving point that is absent from its cited
--      field; the removed one recorded under data->'dropped_unsourced'.
--      NEGATIVE CONTROL, and it is not optional: re-run webdesign.co.uk and
--      robot-hands.com and confirm they KEEP their specifics — "sixty-three
--      tools", "six actuation types", "2-3 technical articles per month".
--      ⚠ Do NOT use gaswholesalers.com as the negative control, although
--      bugs_open/335 proposes it: [MEASURED 2026-08-21] none of its six points
--      contains a cardinal at all, so it passes any rule including one that bans
--      every numeral. It cannot discriminate and it will read as a clean pass.
--
--   4. ⚠ APPLYING THIS DOES NOT REPAIR leopardess. The false rank-1 is already
--      persisted and stays is_current until a run REPLACES it: write_site_spec
--      deep-merges, and an array takes the scalar-overwrite arm, so only a
--      SUCCESSFUL re-run rewrites lead_with. The improvement-sweep was disabled
--      on 2026-08-17 (owner cost control), so nothing will re-run this site on its
--      own. Coordinate with the leopardess lane before firing B4 at that site —
--      it is holding this lane's findings pending an owner design report.
--
-- WHY on_violation = "drop" HERE, when the action defaults to "fail". Failing the
-- run writes nothing, and "nothing" leaves the previous row is_current — so on the
-- one site that actually carries the defect, fail-mode would report a working gate
-- while the false claim stayed live. It would also lose the findings, which are
-- written by the step AFTER this one. Drop writes the surviving points, removes the
-- offending one, and records what it removed under data->'dropped_unsourced', so
-- the artefact carries its own account rather than quietly shrinking. Drop still
-- refuses to write an EMPTY lead_with.
--
-- Rollback: 537_offer_analyser_cardinal_attribution_gate_HOLD_ROLLBACK.sql.

BEGIN;

SELECT snapshot_agent('offer-analyser', '537_cardinal_attribution_gate: pre-update');

-- Already-applied gate (the runner reads a RAISE containing 'already').
DO $$
DECLARE done int;
BEGIN
    SELECT count(*) INTO done FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'offer-analyser'
       AND default_config->'workflow'->'steps' ? 'verify_ordering_cardinals';
    IF done > 0 THEN
        RAISE EXCEPTION '537: already applied — offer-analyser already carries verify_ordering_cardinals';
    END IF;
END $$;

-- NEEDLE-GATE, AND THE ROW IT RESOLVES IS THE ROW THE UPDATES TOUCH.
--
-- ⚠ THE TEMP TABLE IS THE POINT, NOT TIDINESS. The first cut of this migration
-- asserted the step shape in a DO block and then ran the UPDATEs against the
-- BROADER predicate (type + is_active + not-snapshot + not-deleted). Those are
-- DIFFERENT SETS, and this estate has the landmine that makes the difference
-- bite: [MEASURED 2026-08-21] four agent types carry TWO active definition rows
-- (content-creator, content-creator-contact, chief-strategist,
-- site-component-architect), and only the higher version is ever loaded. Had
-- offer-analyser been one of them the gate would have passed on the loaded row
-- while the UPDATE silently rewrote BOTH — corrupting the row nobody reads,
-- where nothing would ever surface it. offer-analyser has exactly one active row
-- today, so the old form would not have misfired; that is luck, not a guard.
-- Caught by the council gate's guardian seat (corr 9a8f1283, round 1).
--
-- Resolving the target ONCE into a temp table makes gate and write the same set
-- by construction, so they cannot drift apart in a later edit either.
CREATE TEMP TABLE _537_target ON COMMIT DROP AS
SELECT id
  FROM agent_definitions
 WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
   AND type = 'offer-analyser'
   AND default_config->'workflow'->'steps'->'set_audit_source'->>'next_step' = 'write_offer_ordering'
   AND default_config->'workflow'->'steps'->'write_offer_ordering'->>'action' = 'write_site_spec'
   AND default_config->'workflow'->'steps'->'write_offer_ordering'->'config'->>'spec_data' = 'offer_analysis.result.ordering'
   AND default_config->'workflow'->'steps'->'write_offer_ordering'->'config'->>'aspect' = 'offer_ordering'
   AND default_config->'workflow'->'steps'->'run_offer_analysis'->>'output_field' = 'offer_analysis';

DO $$
DECLARE n int; total_active int;
BEGIN
    SELECT count(*) INTO n FROM _537_target;
    IF n <> 1 THEN
        RAISE EXCEPTION
          '537 needle-gate: expected exactly 1 offer-analyser whose path is set_audit_source -> write_offer_ordering (write_site_spec, spec_data=offer_analysis.result.ordering), found % — re-derive against the live workflow', n;
    END IF;

    -- Stated separately so the failure names its own cause. A second active row
    -- that does NOT match the shape above is not something to write through
    -- blind: it is either a stale definition or a migration someone else is
    -- part-way through.
    SELECT count(*) INTO total_active FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'offer-analyser';
    IF total_active <> 1 THEN
        RAISE EXCEPTION
          '537 needle-gate: offer-analyser has % active definition rows, expected 1 — only the highest version is ever loaded, so decide which row is real before writing', total_active;
    END IF;
END $$;

-- PROMPT ANCHOR GATE. `replace()` on a missed anchor is a SILENT NO-OP that still
-- reports UPDATE 1 — the migration would ship the gate with the rule never added
-- to the prompt, and report success. So the anchor is asserted to occur EXACTLY
-- ONCE before anything is written: zero means it has drifted (another session
-- edited the prompt), and more than one means `replace()` would splice the rule
-- in twice. Counted by length difference, which is exact and needs no regex.
DO $$
DECLARE prompt text; anchor text; occurrences int;
BEGIN
    anchor := 'naming which premise field it came from, so a later reader can check it.';

    SELECT ad.default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt'
      INTO prompt
      FROM agent_definitions ad JOIN _537_target t ON t.id = ad.id;

    IF prompt IS NULL THEN
        RAISE EXCEPTION '537 prompt-anchor gate: run_offer_analysis carries no prompt';
    END IF;

    occurrences := (length(prompt) - length(replace(prompt, anchor, ''))) / length(anchor);
    IF occurrences <> 1 THEN
        RAISE EXCEPTION
          '537 prompt-anchor gate: the from_field sentence occurs % time(s), expected exactly 1 — re-derive the anchor before splicing', occurrences;
    END IF;
END $$;

-- 1. Insert the gate step, rewire set_audit_source onto it, and point the write
--    at the CHECKED object rather than the raw model output.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(
             default_config,
             '{workflow,steps,verify_ordering_cardinals}',
             jsonb_build_object(
               'action', 'verify_cited_cardinals',
               'config', jsonb_build_object(
                   'object_field', 'offer_analysis.result.ordering',
                   'items_key',    'lead_with',
                   'source_field', 'premise.strategy',
                   'text_key',     'point',
                   'citation_key', 'from_field',
                   'on_violation', 'drop',
                   'dropped_key',  'dropped_unsourced'),
               'next_step', 'write_offer_ordering',
               'description', 'bugs_open/335. Every cardinal in a lead_with point must appear in the premise field the point cites. Drop mode: the offending point is removed and recorded under dropped_unsourced, because a failed run would write nothing and leave the previous, false ordering is_current. NO error_step by design, same as the write it guards.',
               'output_field', 'ordering_checked'),
             true),
           '{workflow,steps,set_audit_source,next_step}',
           '"verify_ordering_cardinals"'::jsonb,
           false),
         '{workflow,steps,write_offer_ordering,config,spec_data}',
         '"ordering_checked.object"'::jsonb,
         false),
       updated_at = now()
 WHERE id IN (SELECT id FROM _537_target);

-- 2. Tell the model the rule as well as enforcing it. Appended to the sentence
--    that already introduces from_field, so the rule sits where the field is
--    defined rather than in a distant constraints block.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,run_offer_analysis,config,prompt}',
         to_jsonb(replace(
           default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt',
           'naming which premise field it came from, so a later reader can check it.',
           'naming which premise field it came from, so a later reader can check it. ANY SPECIFIC QUANTITY you state in a point — a number, in digits or in words — MUST appear in the premise field you name in from_field. If that field states no quantity, write the point without one. A number that happens to be true of the site but is absent from the premise is exactly the failure this rule exists to stop: it arrives looking sourced, because from_field vouches for it. Quantities are checked mechanically after you answer, and a point whose number is not in its cited field is removed.')),
         false),
       updated_at = now()
 WHERE id IN (SELECT id FROM _537_target);

-- VERIFY. A DO/RAISE block, not a SELECT: ON_ERROR_STOP does not fire on a
-- non-empty result set, so a verify block made of SELECTs cannot stop the COMMIT.
DO $$
DECLARE
    step   jsonb;
    prompt text;
BEGIN
    SELECT ad.default_config->'workflow'->'steps'->'verify_ordering_cardinals',
           ad.default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt'
      INTO step, prompt
      FROM agent_definitions ad JOIN _537_target t ON t.id = ad.id;

    IF step IS NULL OR step->>'action' <> 'verify_cited_cardinals' THEN
        RAISE EXCEPTION '537 verify: gate step missing or wrong action (%)', step->>'action';
    END IF;
    IF step->'config'->>'on_violation' <> 'drop' THEN
        RAISE EXCEPTION '537 verify: on_violation is %, expected drop', step->'config'->>'on_violation';
    END IF;
    IF step->>'next_step' <> 'write_offer_ordering' THEN
        RAISE EXCEPTION '537 verify: gate does not hand on to write_offer_ordering';
    END IF;

    -- The two rewires are the ones that make the gate REACHABLE and BINDING.
    -- Without the first the step exists and never runs; without the second it
    -- runs and the write ignores it. Both have been shipped broken elsewhere.
    IF (SELECT ad.default_config->'workflow'->'steps'->'set_audit_source'->>'next_step'
          FROM agent_definitions ad JOIN _537_target t ON t.id = ad.id) <> 'verify_ordering_cardinals' THEN
        RAISE EXCEPTION '537 verify: set_audit_source still bypasses the gate';
    END IF;
    IF (SELECT ad.default_config->'workflow'->'steps'->'write_offer_ordering'->'config'->>'spec_data'
          FROM agent_definitions ad JOIN _537_target t ON t.id = ad.id) <> 'ordering_checked.object' THEN
        RAISE EXCEPTION '537 verify: the write still persists the UNCHECKED model output';
    END IF;

    IF prompt NOT LIKE '%MUST appear in the premise field you name in from_field%' THEN
        RAISE EXCEPTION '537 verify: the prompt rule did not splice';
    END IF;

    RAISE NOTICE '537 OK: gate spliced, set_audit_source -> verify_ordering_cardinals -> write_offer_ordering, write reads ordering_checked.object, prompt rule present';
END $$;

COMMIT;
