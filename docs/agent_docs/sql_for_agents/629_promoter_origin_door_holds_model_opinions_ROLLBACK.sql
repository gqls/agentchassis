-- 629_..._ROLLBACK.sql - remove the origin door (the four inverse replaces).
-- NOTE what rolling back reinstates (bugs_open/405): LLM-audit opinions filed at
-- 'detected' with a proven pair are promoted and dispatched within 15 minutes.
DO $probe$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name='detected-item-promoter' AND pre_query LIKE '%origin_ok%') THEN
        RAISE EXCEPTION '405/629 ROLLBACK: not applied';
    END IF;
END $probe$;
BEGIN;
DO $undo$
DECLARE
    v_q text;
    r1  text := $a$               ) AS floor_ok,
               -- DOOR-CLOSER 3 (bugs_open/405, migration 629): PROVENANCE. Every other
               -- door interrogates the handler and the pair's history; none asked what
               -- PRODUCED the row, so LLM-audit opinions rode a door whose completions
               -- were earned by mechanical defects (27 promoted 08-20..24 with the
               -- sweep off). The stamp is written by write_audit_findings itself
               -- (workItemOriginModelOpinion - lockstep-pinned to this literal); a row
               -- carrying it is a JUDGEMENT: held for a person or record-mode's
               -- release recipe, never promoted by competence history.
               (COALESCE(wi.spec->>'origin', '') <> 'model_opinion') AS origin_ok
        FROM site_work_items wi$a$;
    a1  text := $a$               ) AS floor_ok
        FROM site_work_items wi$a$;
    r2  text := $a$        WHERE pipe_ok AND handler_ok AND known_good AND floor_ok AND origin_ok
        ORDER BY created_at ASC$a$;
    a2  text := $a$        WHERE pipe_ok AND handler_ok AND known_good AND floor_ok
        ORDER BY created_at ASC$a$;
    r3  text := $a$        WHERE NOT (pipe_ok AND handler_ok AND known_good AND floor_ok AND origin_ok)$a$;
    a3  text := $a$        WHERE NOT (pipe_ok AND handler_ok AND known_good AND floor_ok)$a$;
    r4  text := $a$               CASE WHEN NOT handler_ok  THEN 'handler not a live agent'
                    WHEN NOT origin_ok   THEN 'model opinion - release by hand or via record mode (bugs_open/405)'$a$;
    a4  text := $a$               CASE WHEN NOT handler_ok  THEN 'handler not a live agent'$a$;
BEGIN
    SELECT pre_query INTO v_q FROM scheduled_tasks WHERE name = 'detected-item-promoter';
    IF position(r1 in v_q) = 0 OR position(r2 in v_q) = 0 OR position(r3 in v_q) = 0 OR position(r4 in v_q) = 0 THEN
        RAISE EXCEPTION '405/629 ROLLBACK drift: the door is not verbatim as installed - undo by hand';
    END IF;
    v_q := replace(replace(replace(replace(v_q, r1, a1), r2, a2), r3, a3), r4, a4);
    UPDATE scheduled_tasks
       SET pre_query = v_q,
           description = replace(description, ' Door 5 (bugs_open/405, mig 629): rows stamped spec.origin=''model_opinion'' are HELD, never promoted - the stamp is written by write_audit_findings and read here.', ''),
           updated_at = now()
     WHERE name = 'detected-item-promoter';
    IF EXISTS (SELECT 1 FROM scheduled_tasks WHERE name='detected-item-promoter' AND pre_query LIKE '%origin_ok%') THEN
        RAISE EXCEPTION '405/629 ROLLBACK verify: origin_ok survived';
    END IF;
    RAISE NOTICE '405/629 ROLLBACK: door removed';
END $undo$;
COMMIT;
