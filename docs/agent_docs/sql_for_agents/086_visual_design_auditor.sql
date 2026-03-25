-- ============================================================================
-- Migration 086: Section Locking in Auditors + Audit Pass Cap
-- ============================================================================
-- Part A: Update audit agent data-loading queries to exclude locked components.
--         Currently the LLM sees locked sections' HTML and may re-report them.
--
-- Part B: Add audit_pass_count tracking to sites.settings and an
--         improvement-loop guard that stops after 3 passes.
--
-- Part C: Add "skip locked sections" instruction to audit prompts.
--
-- Components/sections that pass verification get locked (locked_at set).
-- Locked = auditor won't see them, won't report on them, won't generate
-- work items for them. This is the termination condition for the triage drain.
-- ============================================================================


-- ============================================================================
-- Part A: Fix data-loading queries to exclude locked components
-- ============================================================================

-- visual-design-auditor: load_design_context — exclude locked from index_samples
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_design_context,config,query}',
        to_jsonb(
                'SELECT sc.name as collection_name, sc.color_palette::text as palette, sc.typography::text as typo, LEFT(ct.css_content, 2000) as css_excerpt, (SELECT string_agg(scomp.slot_name || '':'' || LEFT(scomp.rendered_html, 800), ''|||'') FROM site_components scomp WHERE scomp.site_id = s.id) as component_samples, (SELECT string_agg(LEFT(pc.rendered_html, 600), ''|||'') FROM page_components pc JOIN pages p ON pc.page_id = p.id WHERE p.site_id = s.id AND p.name = ''index'' AND pc.rendered_html IS NOT NULL AND pc.locked_at IS NULL LIMIT 5) as index_samples FROM sites s LEFT JOIN style_collections sc ON s.style_collection_id = sc.id LEFT JOIN css_themes ct ON sc.css_theme_id = ct.id WHERE s.id = $1'::text
        )
                     ),
    updated_at = NOW()
WHERE type = 'visual-design-auditor' AND deleted_at IS NULL;


-- content-quality-auditor: load_page_content — exclude locked page_components
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_page_content,config,query}',
        to_jsonb(
                'SELECT p.name, LEFT(string_agg(pc.rendered_html, '' ''), 1000) as content_sample FROM pages p JOIN page_components pc ON pc.page_id = p.id WHERE p.site_id = $1 AND p.name IN (''index'', ''about'', ''services'', ''contact'') AND pc.rendered_html IS NOT NULL AND pc.locked_at IS NULL GROUP BY p.name ORDER BY p.name'::text
        )
                     ),
    updated_at = NOW()
WHERE type = 'content-quality-auditor' AND deleted_at IS NULL;


-- content-quality-auditor: check_empty_pages — don't count locked components as "empty"
-- A page with only locked components is complete, not empty.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_empty_pages,config,query}',
        to_jsonb(
                'SELECT p.name FROM pages p LEFT JOIN page_components pc ON pc.page_id = p.id AND pc.rendered_html IS NOT NULL AND pc.rendered_html != '''' WHERE p.site_id = $1 AND p.build_status IN (''deployed'', ''active'') GROUP BY p.name HAVING COUNT(pc.id) = 0 AND NOT EXISTS (SELECT 1 FROM page_components lpc WHERE lpc.page_id = p.id AND lpc.locked_at IS NOT NULL)'::text
        )
                     ),
    updated_at = NOW()
WHERE type = 'content-quality-auditor' AND deleted_at IS NULL;


-- ============================================================================
-- Part B: Audit pass tracking function
-- ============================================================================
-- Sites track audit_pass_count in settings. The improvement-loop checks
-- this before running and increments after completing.
-- After 3 passes the improvement-loop completes without running auditors.

CREATE OR REPLACE FUNCTION get_audit_pass_count(p_site_id UUID)
RETURNS INT AS $$
SELECT COALESCE(
               (settings->'maintenance_profile'->'audit_pass_count')::int,
               0
       )
FROM sites WHERE id = p_site_id;
$$ LANGUAGE sql STABLE;


CREATE OR REPLACE FUNCTION increment_audit_pass(p_site_id UUID)
RETURNS INT AS $$
DECLARE
v_current INT;
    v_new INT;
BEGIN
    v_current := get_audit_pass_count(p_site_id);
    v_new := v_current + 1;

UPDATE sites
SET settings = jsonb_set(
        COALESCE(settings, '{}'::jsonb),
        '{maintenance_profile,audit_pass_count}',
        to_jsonb(v_new)
               ),
    updated_at = NOW()
WHERE id = p_site_id;

RETURN v_new;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION reset_audit_passes(p_site_id UUID)
RETURNS VOID AS $$
UPDATE sites
SET settings = jsonb_set(
        COALESCE(settings, '{}'::jsonb),
        '{maintenance_profile,audit_pass_count}',
        '0'::jsonb
               ),
    updated_at = NOW()
WHERE id = p_site_id;
$$ LANGUAGE sql;


-- ============================================================================
-- Part C: Locked section counts in audit context
-- ============================================================================
-- Add a view that shows per-site locking progress — useful for operators
-- and for the improvement-loop guard.

CREATE OR REPLACE VIEW site_locking_progress AS
SELECT
    s.id as site_id,
    s.domain,
    COUNT(pc.id) as total_components,
    COUNT(pc.id) FILTER (WHERE pc.locked_at IS NOT NULL) as locked_components,
    COUNT(DISTINCT p.id) as total_pages,
    COUNT(DISTINCT p.id) FILTER (
        WHERE NOT EXISTS (
            SELECT 1 FROM page_components upc
            WHERE upc.page_id = p.id AND upc.locked_at IS NULL
              AND upc.rendered_html IS NOT NULL
        )
        AND EXISTS (
            SELECT 1 FROM page_components lpc
            WHERE lpc.page_id = p.id AND lpc.locked_at IS NOT NULL
        )
    ) as fully_locked_pages,
    get_audit_pass_count(s.id) as audit_passes,
    CASE
        WHEN COUNT(pc.id) = 0 THEN 'no_components'
        WHEN COUNT(pc.id) FILTER (WHERE pc.locked_at IS NOT NULL) = COUNT(pc.id) THEN 'fully_locked'
        WHEN get_audit_pass_count(s.id) >= 3 THEN 'max_passes_reached'
        ELSE 'in_progress'
END as locking_status
FROM sites s
JOIN pages p ON p.site_id = s.id AND p.status = 'active'
LEFT JOIN page_components pc ON pc.page_id = p.id AND pc.rendered_html IS NOT NULL
WHERE s.status = 'active'
GROUP BY s.id, s.domain;


-- ============================================================================
-- Verification
-- ============================================================================

SELECT proname FROM pg_proc
WHERE proname IN ('get_audit_pass_count', 'increment_audit_pass', 'reset_audit_passes')
ORDER BY proname;

-- Check the updated queries include locked_at filter
SELECT type,
       CASE
           WHEN default_config->'workflow'->'steps'->'load_design_context'->'config'->>'query' LIKE '%locked_at IS NULL%' THEN 'YES'
        ELSE 'NO'
END as design_excludes_locked
FROM agent_definitions
WHERE type = 'visual-design-auditor' AND deleted_at IS NULL;

SELECT type,
       CASE
           WHEN default_config->'workflow'->'steps'->'load_page_content'->'config'->>'query' LIKE '%locked_at IS NULL%' THEN 'YES'
        ELSE 'NO'
END as content_excludes_locked
FROM agent_definitions
WHERE type = 'content-quality-auditor' AND deleted_at IS NULL;

