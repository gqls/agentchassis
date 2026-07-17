-- PATCH — add a STABILITY-PREFERENCE proviso to the fix-proposer council's
-- GUARDIAN reviewer (owner request 2026-07-17): as a preference, don't change
-- code that has been working for ages (orchestrator, Kafka/messaging, agent
-- spawning, core work-item dispatch); prefer a fix at a higher layer.
--
-- The guardian is the right home: it already judges (a) blast radius,
-- (b) architecture-change signals (incl. wire formats/message shapes),
-- (c) surface ownership. This adds (d) as a preference — object/steer to a
-- higher layer, reserving veto for a genuine architecture change (its existing
-- behaviour), NOT a new blanket veto on all core edits.
--
-- ██ SURGICAL BY DESIGN ██ — the guardian prompt is co-edited by another thread
-- (it gained a `code_checks` mechanism on 2026-07-17 that is NOT in the seat
-- migration files). A full-config reapply (the v6-v11 pattern) would clobber
-- that. So this touches ONLY review_guardian.config.prompt_template, via a
-- string replace() inside jsonb_set — every other step, code_checks, and the
-- relevance-filter wiring are left byte-identical. Idempotent (the WHERE guard
-- refuses if the proviso is already present).

BEGIN;

SELECT snapshot_agent('fix-proposer', 'pre-update: guardian stability-preference proviso (surgical; preserves code_checks + filter wiring)')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='fix-proposer' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,review_guardian,config,prompt_template}',
      to_jsonb(
        replace(
          default_config #>> '{workflow,steps,review_guardian,config,prompt_template}',
          'name the owning pipeline.',
          'name the owning pipeline. (d) STABILITY PREFERENCE — long-stable, load-bearing infrastructure (the orchestrator, Kafka/messaging, agent spawning, the core work-item dispatch) is battle-tested; PREFER a fix at a higher, less-foundational layer over editing it. An edit to this core is itself a strong architecture-change signal: object and ask whether the cause can be addressed above it, and reserve veto for a genuine architecture change to foundational plumbing dressed as a point fix.'
        )
      )
    ),
    updated_at = now()
WHERE type='fix-proposer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND (default_config #>> '{workflow,steps,review_guardian,config,prompt_template}') LIKE '%name the owning pipeline.%'
  AND (default_config #>> '{workflow,steps,review_guardian,config,prompt_template}') NOT LIKE '%(d) STABILITY PREFERENCE%';

COMMIT;

-- Rollback (manual): restore the pre-update snapshot from agent_definitions_backup
-- (snapshot_agent wrote it on apply), or re-run with the (d) clause stripped.
