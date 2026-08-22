-- 550_bugfix_342_arm_chrome_absent_required_record_ROLLBACK.sql
--
-- Removes record_absent_required_fields from every live render_site_components
-- step config. Unset means today's-before-550 behaviour byte for byte (the key
-- is read fail-open by recordAbsentRequiredFields), so removal is the complete
-- rollback; the snapshots taken by 550 remain available for a full restore.

BEGIN;

DO $$
DECLARE
    r record;
    n integer := 0;
BEGIN
    FOR r IN
        SELECT a.id, s.k AS step
        FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s(k,v)
        WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
          AND s.v->>'action' = 'render_site_components'
          AND s.v->'config' ? 'record_absent_required_fields'
    LOOP
        UPDATE agent_definitions
        SET default_config = default_config #- ARRAY['workflow','steps', r.step, 'config', 'record_absent_required_fields'],
            version    = version + 1,
            updated_at = now()
        WHERE id = r.id;
        n := n + 1;
    END LOOP;
    RAISE NOTICE 'ROLLBACK 550: removed the key from % step(s)', n;
END $$;

DO $$
DECLARE
    n_left integer;
BEGIN
    SELECT count(*) INTO n_left
    FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s(k,v)
    WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
      AND s.v->>'action' = 'render_site_components'
      AND s.v->'config' ? 'record_absent_required_fields';
    IF n_left > 0 THEN
        RAISE EXCEPTION 'ROLLBACK 550: % step(s) still carry the key.', n_left;
    END IF;
END $$;

COMMIT;
