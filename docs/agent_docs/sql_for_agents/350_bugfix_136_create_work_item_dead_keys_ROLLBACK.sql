-- ============================================================================
-- 350_bugfix_136_create_work_item_dead_keys_ROLLBACK.sql
--
-- Reverses 350: puts both dead keys back. Behaviour-neutral in both
-- directions — neither key is resolved by any strategy in
-- ExtractActionInputs, which is why 350 removed them.
-- ============================================================================

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,create_review_item,config,spec_fields}',
      '["draft","grounding_audit","registration","input_data"]'::jsonb
    ),
    updated_at = now()
WHERE type='grounded-explainer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,request_claims_review,config,domain}',
      '"site_record.domain"'::jsonb
    ),
    updated_at = now()
WHERE type='claims-auditor' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

COMMIT;
