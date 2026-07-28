-- 256 — experience_patterns ↔ content_components join check (brochure §4d step 1)
--
-- Every experience_patterns.section_types entry must name a LIVE component;
-- nothing reconciles the two (same defect class as bugs_open/109: declarations
-- nobody checks). Measured 2026-07-28: 4 of 9 patterns carried near-miss names
-- (hero-carousel vs hero-card-carousel etc) — exactly the four components the
-- brochure workstream built.
--
-- A component is named two ways in content_components: `function` AND
-- `section_type`. The §4d analysis originally joined on `function` alone —
-- this check accepts EITHER (active components only) and says which matched,
-- so a miss here is a real miss, not an artefact of picking the wrong column.
--
-- Run:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -f - < 256_experience_pattern_component_join_check.sql
--
-- Empty result = every declared section_type resolves. Any row = a pattern
-- declaring a component that does not exist (or is inactive) — fix the pattern
-- via the experience register's validating write path, NOT a bare UPDATE, and
-- check for a rename on the component side before assuming the pattern is wrong.

SELECT ep.name        AS pattern,
       ep.kind,
       ep.status      AS pattern_status,
       st.claimed     AS claimed_section_type,
       -- nearest live name either way, to catch the rename-vs-typo case at a glance
       (SELECT cc.function FROM content_components cc
         WHERE cc.is_active
           AND (cc.function ILIKE '%' || replace(st.claimed, '-', '%') || '%'
             OR cc.section_type ILIKE '%' || replace(st.claimed, '-', '%') || '%')
         ORDER BY length(cc.function) LIMIT 1) AS nearest_live_function
FROM experience_patterns ep,
     LATERAL (SELECT jsonb_array_elements_text(ep.section_types) AS claimed) st
WHERE NOT EXISTS (
        SELECT 1 FROM content_components cc
        WHERE cc.is_active
          AND (cc.function = st.claimed OR cc.section_type = st.claimed))
ORDER BY ep.name, st.claimed;
