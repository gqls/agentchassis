-- ============================================================================
-- 203_link_resolver_sections_optional.sql — bugs_open/068 fix candidate B
--
-- Move `sections` from required to optional in internal-link-resolver's
-- input_contract. The Go action (resolve_internal_links_action.go:132-135,151)
-- is nil-safe on sections — missing => empty loop => empty sections_ready, no
-- error — so the contract was the ONLY fatal element. Effect: page-rebuild's
-- writer children (which structurally cannot supply section_plan — this writer
-- generation selects its own sections) stop dying at extraction; the
-- page-build-handler path is unchanged (it always supplies sections).
--
-- Evidence + runtime caller comparison: bugs_open/068. Same failure 2026-07-16
-- x2 (unfiled) and 2026-07-24 (about-commercial pilot).
--
-- REVERT (original contract verbatim):
--   UPDATE agent_definitions SET input_contract = '{
--     "optional": ["page_type", "page_name"],
--     "required": ["site_id", "sections"],
--     "description": "site_id and the section_plan sections_ready list; page_type/page_name refine destination choice (own-hub exclusion)."
--   }'::jsonb
--   WHERE type='internal-link-resolver' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- ============================================================================

DO $$
DECLARE
  cur jsonb;
BEGIN
  SELECT input_contract INTO cur
  FROM agent_definitions
  WHERE type='internal-link-resolver' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF cur IS NULL THEN
    RAISE EXCEPTION 'no active internal-link-resolver row found';
  END IF;
  IF NOT (cur->'required' ? 'sections') THEN
    RAISE NOTICE 'sections already NOT required — nothing to do (contract: %)', cur;
    RETURN;
  END IF;

  UPDATE agent_definitions
  SET input_contract = jsonb_set(
        jsonb_set(cur, '{required}', (cur->'required') - 'sections'),
        '{optional}', (cur->'optional') || '["sections"]'::jsonb)
      || '{"description": "site_id required; sections (the section_plan sections_ready list) OPTIONAL — absent means the nil-safe resolver returns empty sections_ready (rebuild-path callers have no section_plan; bugs_open/068). page_type/page_name refine destination choice (own-hub exclusion)."}'::jsonb
  WHERE type='internal-link-resolver' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  RAISE NOTICE 'internal-link-resolver contract updated: sections required -> optional';
END $$;

-- verify
SELECT jsonb_pretty(input_contract)
FROM agent_definitions
WHERE type='internal-link-resolver' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
