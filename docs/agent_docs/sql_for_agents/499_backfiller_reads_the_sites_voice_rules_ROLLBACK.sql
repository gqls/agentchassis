-- 499 ROLLBACK — stop telling the writer the site's banned phrases
--
-- ⚠ THIS DOES NOT WEAKEN ANY GATE. The refusal in save_page_meta_description is Go
-- and is untouched either way. What rolling back restores is the state where the
-- writer is refused AFTER writing rather than told the rules before — i.e. it
-- re-arms the permanent hourly retry on the 9 sites that carry a voice gate.
--
-- Restores from 499's own snapshot rather than un-splicing, because the forward
-- migration rewired a step chain and edited a prompt; reversing that by string
-- replacement is how a half-rewired workflow happens.

BEGIN;

DO $$
DECLARE snap_id uuid; n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='meta-description-backfiller' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '499 ROLLBACK: expected exactly 1 live row, found %', n;
  END IF;

  SELECT id INTO snap_id FROM agent_definitions
   WHERE type='meta-description-backfiller' AND COALESCE(is_snapshot,false)=true
     AND description LIKE '%499_voice_rules: pre-update%'
   ORDER BY created_at DESC LIMIT 1;
  IF snap_id IS NULL THEN
    RAISE EXCEPTION '499 ROLLBACK: no 499 pre-update snapshot — refusing to guess at the previous config';
  END IF;

  UPDATE agent_definitions live
     SET default_config = snap.default_config, updated_at = now()
    FROM agent_definitions snap
   WHERE snap.id = snap_id
     AND live.type='meta-description-backfiller'
     AND live.is_active AND COALESCE(live.is_snapshot,false)=false AND live.deleted_at IS NULL;
END $$;

DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='meta-description-backfiller' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg#>>'{workflow,steps,ensure_site_record,next_step}' IS DISTINCT FROM 'load_pages_missing_meta' THEN
    RAISE EXCEPTION '499 ROLLBACK VERIFY: the chain was not restored';
  END IF;
  IF position('voice_rules' in cfg#>>'{workflow,steps,write_descriptions,config,prompt_template}') > 0 THEN
    RAISE EXCEPTION '499 ROLLBACK VERIFY: the prompt still references voice_rules';
  END IF;
  RAISE NOTICE '499 ROLLBACK OK — the writer is no longer told the rules; the hourly retry on gated sites is re-armed';
END $$;

COMMIT;
