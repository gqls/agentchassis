-- 618_..._ROLLBACK.sql - put vertical-exemplar-researcher back to the strictly
-- linear chain with no error_steps and no floor.
--
-- The inverse edit, not a snapshot restore: 618 touched exactly seven keys, and a
-- snapshot restore would also discard anything another lane has legitimately
-- changed on this row since. The snapshot is the row stamped
-- '618_vertical_exemplar_researcher_routes_refused_crawls_and_floors_on_content: pre-update'.
--
-- NOTE what rolling back reinstates (bugs_open/376): one refused exemplar kills the
-- whole research stage, and a succeeds-but-empty crawl proceeds uncounted. Roll
-- back only to unblock something else, and say so.
DO $probe$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'vertical-exemplar-researcher'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #> '{workflow,steps}' ? 'check_exemplar_floor'
    ) THEN
        RAISE EXCEPTION '376/618 ROLLBACK: not applied - check_exemplar_floor is absent';
    END IF;
END $probe$;

BEGIN;
SELECT snapshot_agent('vertical-exemplar-researcher', '618_..._ROLLBACK: pre-restore');

DO $undo$
DECLARE
    v_id    uuid;
    v_steps jsonb;
    v_n     int;
BEGIN
    SELECT id, default_config #> '{workflow,steps}' INTO v_id, v_steps
      FROM agent_definitions
     WHERE type = 'vertical-exemplar-researcher'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     ORDER BY version DESC LIMIT 1;
    IF v_steps #>> '{format_exemplar_3,next_step}' IS DISTINCT FROM 'check_exemplar_floor' THEN
        RAISE EXCEPTION '376/618 ROLLBACK drift: format_exemplar_3.next_step is %, not check_exemplar_floor', v_steps #>> '{format_exemplar_3,next_step}';
    END IF;
    FOR v_n IN 1..3 LOOP
        v_steps := v_steps #- ARRAY[format('crawl_exemplar_%s', v_n), 'error_step'];
        v_steps := jsonb_set(v_steps, ARRAY[format('crawl_exemplar_%s', v_n), 'description'],
            to_jsonb(format('Shallow crawl of exemplar %s (front page + direct links)', v_n)), true);
    END LOOP;
    v_steps := jsonb_set(v_steps, '{format_exemplar_3,next_step}', to_jsonb('synthesise'::text), true);
    v_steps := v_steps - 'check_exemplar_floor' - 'record_exemplar_floor' - 'insufficient_exemplars';
    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config, '{workflow,steps}', v_steps, false), updated_at = now()
     WHERE id = v_id;
    IF (SELECT count(*) FROM jsonb_object_keys(v_steps)) <> 12 THEN
        RAISE EXCEPTION '376/618 ROLLBACK verify: expected 12 steps after restore, got %', (SELECT count(*) FROM jsonb_object_keys(v_steps));
    END IF;
    RAISE NOTICE '376/618 ROLLBACK: restored linear chain on %', v_id;
END $undo$;
COMMIT;
