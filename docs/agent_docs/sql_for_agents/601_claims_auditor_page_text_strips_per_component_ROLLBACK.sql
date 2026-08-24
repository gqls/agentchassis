-- ROLLBACK for 601_claims_auditor_page_text_strips_per_component.sql
-- Restores the 597-form query (aggregate-then-strip, cap 12000). Restores the extraction
-- DEFECT (Postgres greedy-first regex eats cross-component text) — recovery only.

BEGIN;

SELECT snapshot_agent('claims-auditor', '601_ROLLBACK: pre-revert');

DO $$
DECLARE q text; n int;
BEGIN
  SELECT default_config #>> '{workflow,steps,load_page_text,config,query}' INTO q
    FROM agent_definitions WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('ORDER BY pc.position, pc.slot_name' in q) = 0 THEN
    RAISE EXCEPTION '601 ROLLBACK: per-component form not present — 601 not applied';
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,load_page_text,config,query}',
           to_jsonb($Q$SELECT p.name, LEFT(regexp_replace(regexp_replace(regexp_replace(regexp_replace(string_agg(pc.rendered_html, ' '), '<style[^>]*>.*?</style>', ' ', 'gi'), '<script[^>]*>.*?</script>', ' ', 'gi'), '<[^>]*>', ' ', 'g'), '\s+', ' ', 'g'), 12000) AS page_text FROM pages p JOIN page_components pc ON pc.page_id = p.id WHERE p.site_id = $1 AND p.build_status IN ('deployed','active') AND pc.rendered_html IS NOT NULL AND pc.rendered_html <> '' AND pc.locked_at IS NULL GROUP BY p.id, p.name ORDER BY p.name$Q$::text),
           false),
         updated_at = now()
   WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '601 ROLLBACK: updated % rows', n; END IF;
  RAISE NOTICE '601 ROLLBACK OK.';
END $$;

COMMIT;
