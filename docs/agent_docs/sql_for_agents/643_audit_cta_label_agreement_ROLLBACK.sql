-- ROLLBACK for 643 — disarm the write-time CTA label/destination audit.
--
-- Removes the key rather than setting it false, so the row returns to exactly
-- its pre-643 shape and the code's own default (OFF) governs. Safe at any time:
-- the pass only records, so disarming loses instrumentation and changes no
-- content, no markup and no destination.
BEGIN;

UPDATE agent_definitions
SET default_config = default_config
        #- '{workflow,steps,save_sections,config,audit_cta_label_agreement}'
        #- '{workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections,config,audit_cta_label_agreement}'
        #- '{workflow,steps,build_items_loop,config,sub_workflow,steps,save_sections,config,audit_cta_label_agreement}',
    updated_at = NOW()
WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config::text LIKE '%audit_cta_label_agreement%';

DO $$
DECLARE
    armed integer;
BEGIN
    SELECT count(*) INTO armed
    FROM agent_definitions a,
         LATERAL jsonb_path_query(a.default_config,
                 'strict $.**.audit_cta_label_agreement ? (@ == true)') x
    WHERE a.is_active AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL;
    IF armed <> 0 THEN
        RAISE EXCEPTION 'rollback left % steps armed', armed;
    END IF;
END $$;

COMMIT;
