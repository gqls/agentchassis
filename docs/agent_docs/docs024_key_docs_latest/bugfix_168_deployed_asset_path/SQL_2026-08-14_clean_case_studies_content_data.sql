-- Remove the unsupported figure "75,061 orchestration state records." from
-- leopardessconsulting.co.uk /case-studies, page_components 9c9aaed8.
--
-- WHY: the site's evidence_base fact C4-orchestration-state-records carries
-- value 2578 with tolerance `gte` ("published snapshots supported up to the
-- live count"). 75,061 is ~29x that, so the page overclaims. Two independent
-- platform findings say so: claims_unverified item e713613f (unregistered
-- number) and stale_evidence item 3a5419a1 (same fact, drifted down).
--
-- ACT: minimal deletion of one sentence, per the owner's 2026-08-06 ruling
-- ("minimal deletion is not writing") and the check's own prescribed fix
-- ("unregistered numbers need either an evidence_base fact row ... or removal
-- from the copy ... cleared in page_components.content_data, NOT only in the
-- rendered HTML"). No connective repair is needed: the sentences are
-- independent, so the deletion leaves "...spawn other agents. Eight live
-- sites are built and maintained this way...".
--
-- GUARD: every assertion is DO/RAISE, never a bare SELECT. A verify block of
-- SELECTs cannot stop a COMMIT (ON_ERROR_STOP ignores a non-empty result) —
-- LANDMINE, see MEMORY a-one-off-deletion-is-not-a-class-fix. Set
-- :expect_delta to a wrong value to prove the guard aborts a real UPDATE.
\set ON_ERROR_STOP on

BEGIN;

-- psql does NOT interpolate :vars inside a dollar-quoted body, so the expected
-- delta travels via a GUC set out here, where interpolation does happen.
SET LOCAL app.expect_delta = :expect_delta;

DO $$
DECLARE
  v_pc      uuid := '9c9aaed8-d13f-4825-afdb-e449bfe8f92b';
  v_needle  text := '75,061 orchestration state records. ';
  v_delta   int  := current_setting('app.expect_delta')::int;
  v_title   text;
  v_old     text;
  v_new     text;
  v_old_cd  text;
  v_new_cd  text;
  v_rows    int;
BEGIN
  SELECT content_data #>> '{case_studies,3,title}',
         content_data #>> '{case_studies,3,results}',
         content_data::text
    INTO v_title, v_old, v_old_cd
    FROM page_components WHERE id = v_pc;

  IF v_old IS NULL THEN
    RAISE EXCEPTION 'ABORT: path {case_studies,3,results} not readable on %', v_pc;
  END IF;

  -- Identify the element by what it SAYS, not only by its index: an index is a
  -- position another session's edit could move underneath us.
  IF v_title IS DISTINCT FROM 'The platform that built this website' THEN
    RAISE EXCEPTION 'ABORT: case_studies[3] is %, not the expected entry', COALESCE(v_title,'<null>');
  END IF;

  IF position(v_needle in v_old) = 0 THEN
    RAISE EXCEPTION 'ABORT: needle not present in results — already edited, or the text moved';
  END IF;

  v_new := replace(v_old, v_needle, '');

  UPDATE page_components
     SET content_data = jsonb_set(content_data, '{case_studies,3,results}', to_jsonb(v_new)),
         updated_at   = now()
   WHERE id = v_pc;
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 1 THEN
    RAISE EXCEPTION 'ABORT: updated % row(s), expected exactly 1', v_rows;
  END IF;

  SELECT content_data::text INTO v_new_cd FROM page_components WHERE id = v_pc;

  -- The claim must be gone from the WHOLE component, not just the slot we aimed at.
  IF position('75,061' in v_new_cd) > 0 THEN
    RAISE EXCEPTION 'ABORT: 75,061 survives elsewhere in content_data';
  END IF;

  -- Exactly the needle's length, and nothing else, may have gone.
  IF length(v_old_cd) - length(v_new_cd) <> v_delta THEN
    RAISE EXCEPTION 'ABORT: content_data shrank by % chars, expected %',
      length(v_old_cd) - length(v_new_cd), v_delta;
  END IF;

  -- And structurally: the new document must equal the old one with ONLY that
  -- one path replaced. Catches a half-applied or wider edit that happens to
  -- have the right length.
  IF jsonb_set(v_old_cd::jsonb, '{case_studies,3,results}', to_jsonb(v_new)) <> v_new_cd::jsonb THEN
    RAISE EXCEPTION 'ABORT: content_data differs in more than the one field';
  END IF;

  RAISE NOTICE 'BEFORE: %', v_old;
  RAISE NOTICE 'AFTER : %', v_new;
  RAISE NOTICE 'OK: one field, -% chars, 1 row.', v_delta;
END $$;

COMMIT;
