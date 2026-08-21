-- 537_offer_analyser_cardinal_attribution_gate_HOLD_ROLLBACK.sql
--
-- Undoes 537: removes the gate step, restores set_audit_source -> write_offer_ordering,
-- points the write back at the raw model output, and strips the prompt rule.
-- Restores today's behaviour exactly (i.e. bugs_open/335 reopens).

BEGIN;

SELECT snapshot_agent('offer-analyser', '537_ROLLBACK: pre-revert');

-- Same duplicate-active-row guard as the forward migration: only the highest
-- version is ever loaded, so a second active row means an UPDATE here would
-- rewrite a row nobody is reading. [MEASURED 2026-08-21] four agent types on
-- this estate carry two; offer-analyser carries one.
DO $$
DECLARE total_active int;
BEGIN
    SELECT count(*) INTO total_active FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'offer-analyser';
    IF total_active <> 1 THEN
        RAISE EXCEPTION
          '537 ROLLBACK: offer-analyser has % active definition rows, expected 1 — decide which row is real before reverting', total_active;
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config #- '{workflow,steps,verify_ordering_cardinals}',
           '{workflow,steps,set_audit_source,next_step}',
           '"write_offer_ordering"'::jsonb,
           false),
         '{workflow,steps,write_offer_ordering,config,spec_data}',
         '"offer_analysis.result.ordering"'::jsonb,
         false),
       updated_at = now()
 WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
   AND type = 'offer-analyser';

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,run_offer_analysis,config,prompt}',
         to_jsonb(replace(
           default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt',
           ' ANY SPECIFIC QUANTITY you state in a point — a number, in digits or in words — MUST appear in the premise field you name in from_field. If that field states no quantity, write the point without one. A number that happens to be true of the site but is absent from the premise is exactly the failure this rule exists to stop: it arrives looking sourced, because from_field vouches for it. Quantities are checked mechanically after you answer, and a point whose number is not in its cited field is removed.',
           '')),
         false),
       updated_at = now()
 WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
   AND type = 'offer-analyser';

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM agent_definitions
                WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
                  AND type='offer-analyser'
                  AND default_config->'workflow'->'steps' ? 'verify_ordering_cardinals') THEN
        RAISE EXCEPTION '537 ROLLBACK verify: the gate step is still present';
    END IF;
    RAISE NOTICE '537 ROLLBACK OK: gate removed, path and prompt restored';
END $$;

COMMIT;
