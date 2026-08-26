-- 629_promoter_origin_door_holds_model_opinions.sql
--
-- bugs_open/405 candidate 1, the SQL half. The Go half (write_audit_findings
-- stamping spec.origin = 'model_opinion' into every finding's base spec) ships in
-- the same council submission; the two literals are pinned against each other by
-- TestOriginDoorLockstep, which reads THIS file.
--
-- WHAT IS WRONG (405 §1-§2, both peer-verified). detected-item-promoter's four
-- doors interrogate the HANDLER and the (item_type, handler) pair's lifetime
-- history; none asks what PRODUCED the row. So an LLM auditor's OPINION filed at
-- 'detected' rides a known-good door whose completions were earned by mechanical
-- defects: [MEASURED 2026-08-25] 27 opinion rows promoted 08-20..24 while
-- improvement-sweep was DISABLED. The 391 lane proved the axis cannot be DERIVED
-- from the row (created_by = a drifting name list; source='discovery' spans 27
-- creators; spec ? 'audit_source' over-blocks tool-acceptance-tier4 and
-- owner-request) - so the stamp is WRITTEN by the producer and this door READS it.
--
-- WHAT THIS DOES: four surgical, verbatim-anchored replaces on the live
-- pre_query (the 616 pattern - each anchor asserted to appear EXACTLY ONCE first,
-- so drift stops the migration instead of being silently edited):
--   1. scored gains  (COALESCE(wi.spec->>'origin','') <> 'model_opinion') AS origin_ok
--   2. candidates requires origin_ok
--   3. held's complement includes origin_ok
--   4. held's reason CASE names it: 'model opinion - release by hand or via record mode'
-- plus one sentence appended to the task's description.
--
-- WHAT IT DOES NOT DO: touch any existing row (the stamp exists on 0 rows today -
-- [MEASURED 2026-08-26] live and archive both 0 - so the door holds only FUTURE
-- rows whose producer stamps them); change record mode (deferred rows were never
-- the promoter's candidates); or promote anything it previously held.
--
-- VERIFY AFTER APPLY (405 §6, the SAFE two-direction form - the original recipe's
-- doc_notes target does not exist, and a synthetic PROMOTABLE control would be
-- DISPATCHED to a real handler, so:
--   direction 1 (held): insert one synthetic 'detected' row with a proven pair
--   AND spec.origin='model_opinion'; after >= 2 ticks assert it is STILL
--   'detected'; then close it BY HAND (status='cancelled', result noting it was
--   the 405 verification row).
--   direction 2 (the ticks ran): assert >= 1 NATURAL promotion (any row,
--   triaged_at in the same window, spec.original_pipeline stamped) - never a
--   synthetic control, which would dispatch real work.
--
-- NO ORDERING CONSTRAINT IS CLAIMED (owner ruling 2026-07-29). Config, live on
-- apply. The Go stamp is inert without this door; this door is inert until
-- stamped rows exist. Either order is safe; the door merely holds nothing until
-- the stamping binary rolls (it did: v1.0.1339 carries the write_audit_findings
-- seam, and the stamp rides the next roll after its commit).

DO $probe$
BEGIN
    IF EXISTS (SELECT 1 FROM scheduled_tasks WHERE name='detected-item-promoter' AND pre_query LIKE '%origin_ok%') THEN
        RAISE EXCEPTION '405/629: already applied - the origin door is present';
    END IF;
END $probe$;

BEGIN;

DO $edit$
DECLARE
    v_q  text;
    a1   text := $a$               ) AS floor_ok
        FROM site_work_items wi$a$;
    r1   text := $a$               ) AS floor_ok,
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
    a2   text := $a$        WHERE pipe_ok AND handler_ok AND known_good AND floor_ok
        ORDER BY created_at ASC$a$;
    r2   text := $a$        WHERE pipe_ok AND handler_ok AND known_good AND floor_ok AND origin_ok
        ORDER BY created_at ASC$a$;
    a3   text := $a$        WHERE NOT (pipe_ok AND handler_ok AND known_good AND floor_ok)$a$;
    r3   text := $a$        WHERE NOT (pipe_ok AND handler_ok AND known_good AND floor_ok AND origin_ok)$a$;
    a4   text := $a$               CASE WHEN NOT handler_ok  THEN 'handler not a live agent'$a$;
    r4   text := $a$               CASE WHEN NOT handler_ok  THEN 'handler not a live agent'
                    WHEN NOT origin_ok   THEN 'model opinion - release by hand or via record mode (bugs_open/405)'$a$;
BEGIN
    SELECT pre_query INTO v_q FROM scheduled_tasks WHERE name = 'detected-item-promoter';
    IF v_q IS NULL THEN RAISE EXCEPTION '405/629: detected-item-promoter not found'; END IF;

    IF (length(v_q) - length(replace(v_q, a1, ''))) / length(a1) <> 1 THEN RAISE EXCEPTION '405/629 drift: anchor 1 (scored floor_ok) not exactly once'; END IF;
    IF (length(v_q) - length(replace(v_q, a2, ''))) / length(a2) <> 1 THEN RAISE EXCEPTION '405/629 drift: anchor 2 (candidates) not exactly once'; END IF;
    IF (length(v_q) - length(replace(v_q, a3, ''))) / length(a3) <> 1 THEN RAISE EXCEPTION '405/629 drift: anchor 3 (held complement) not exactly once'; END IF;
    IF (length(v_q) - length(replace(v_q, a4, ''))) / length(a4) <> 1 THEN RAISE EXCEPTION '405/629 drift: anchor 4 (held CASE) not exactly once'; END IF;

    v_q := replace(v_q, a1, r1);
    v_q := replace(v_q, a2, r2);
    v_q := replace(v_q, a3, r3);
    v_q := replace(v_q, a4, r4);

    UPDATE scheduled_tasks
       SET pre_query = v_q,
           description = description || ' Door 5 (bugs_open/405, mig 629): rows stamped spec.origin=''model_opinion'' are HELD, never promoted - the stamp is written by write_audit_findings and read here.',
           updated_at = now()
     WHERE name = 'detected-item-promoter';
    RAISE NOTICE '405/629: origin door installed';
END $edit$;

DO $verify$
DECLARE v_q text; v_n int;
BEGIN
    SELECT pre_query INTO v_q FROM scheduled_tasks WHERE name = 'detected-item-promoter';
    v_n := (length(v_q) - length(replace(v_q, 'origin_ok', ''))) / length('origin_ok');
    IF v_n <> 4 THEN RAISE EXCEPTION '405/629 verify: origin_ok appears % times, want 4 (scored, candidates, held complement, held CASE)', v_n; END IF;
    IF position($x$(COALESCE(wi.spec->>'origin', '') <> 'model_opinion') AS origin_ok$x$ in v_q) = 0 THEN
        RAISE EXCEPTION '405/629 verify: the door predicate is not verbatim';
    END IF;
    IF position('model opinion - release by hand' in v_q) = 0 THEN
        RAISE EXCEPTION '405/629 verify: the held reason arm is missing';
    END IF;
    RAISE NOTICE '405/629: verified - origin_ok x4, predicate verbatim, held reason named';
END $verify$;

COMMIT;
