-- 752_d4b_governor_council_gate_stage_b_HOLD_ROLLBACK.sql — undo D4b stage B.
-- Removes the four gate steps from council-gate, restores start_step = load_schema_hint, and
-- RECREATES governor_withheld_runs + its view (752 dropped them) so 751's own rollback still
-- finds exactly what it expects. Refuses unless the row is in 752's post-apply shape.

BEGIN;

DO $$
DECLARE ss text; n int;
BEGIN
  SELECT default_config#>>'{workflow,start_step}',
         (SELECT count(*) FROM jsonb_object_keys(default_config#>'{workflow,steps}'))
    INTO ss, n
  FROM agent_definitions
  WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  ORDER BY version DESC LIMIT 1;
  IF ss IS NULL THEN RAISE EXCEPTION '752 ROLLBACK REFUSED: no live council-gate row.'; END IF;
  IF ss <> 'gate_spend_governor' OR n <> 48 THEN
    RAISE EXCEPTION '752 ROLLBACK REFUSED: row is not in 752''s shape (start_step=%, steps=%) — 752 not applied, or the roster changed since; investigate.', ss, n;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='governor_withheld_runs') THEN
    RAISE EXCEPTION '752 ROLLBACK REFUSED: governor_withheld_runs already exists — 752 did not drop it, or this already ran.';
  END IF;
  PERFORM snapshot_agent('council-gate', '752 ROLLBACK pre-apply');
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
  jsonb_set(default_config, '{workflow,steps}',
    (default_config#>'{workflow,steps}') - 'gate_spend_governor' - 'route_spend_governor' - 'note_withheld' - 'complete_withheld'),
  '{workflow,start_step}', '"load_schema_hint"'::jsonb)
WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Recreate what 752 dropped, verbatim from 751, so the 751 rollback's preconditions hold.
CREATE TABLE governor_withheld_runs (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  agent_type     text NOT NULL,
  correlation_id text,
  shed_level     int  NOT NULL,
  class          text,
  llm_bearing    boolean,
  request_topic  text,
  withheld_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_governor_withheld_runs_corr ON governor_withheld_runs (correlation_id)
  WHERE correlation_id IS NOT NULL;
CREATE INDEX idx_governor_withheld_runs_at ON governor_withheld_runs (withheld_at DESC);
CREATE VIEW governor_withheld_runs_recent AS
SELECT r.*, gs.shed_level AS current_shed_level, gc.enabled AS governor_enabled
FROM governor_withheld_runs r, governor_state gs, governor_config gc
WHERE gs.id = 1 AND gc.id = 1 AND r.withheld_at > now() - interval '7 days';

DO $$
DECLARE cfg jsonb; n int;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
  WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  ORDER BY version DESC LIMIT 1;
  IF cfg#>>'{workflow,start_step}' <> 'load_schema_hint' THEN RAISE EXCEPTION '752 ROLLBACK VERIFY: start_step not restored'; END IF;
  SELECT count(*) INTO n FROM jsonb_object_keys(cfg#>'{workflow,steps}');
  IF n <> 44 THEN RAISE EXCEPTION '752 ROLLBACK VERIFY: expected 44 steps, found %', n; END IF;
  IF md5(cfg#>>'{workflow}') <> '8dd74a5b042a7376a1e26fbf5db6ba00' THEN
    RAISE EXCEPTION '752 ROLLBACK VERIFY: workflow md5 % is not the pre-752 text — something else changed the row meanwhile; the four steps are gone but do not assume byte-identity', md5(cfg#>>'{workflow}');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='governor_withheld_runs') THEN
    RAISE EXCEPTION '752 ROLLBACK VERIFY: governor_withheld_runs not recreated';
  END IF;
  RAISE NOTICE '752 ROLLBACK OK: council-gate back to load_schema_hint / 44 steps, workflow md5 byte-identical to pre-752; governor_withheld_runs recreated.';
END $$;

COMMIT;
