-- 621_park_work_items_verb.sql
--
-- THE PARK VERB — bugs_open/396.
--
-- WHY THIS EXISTS, measured rather than assumed. Four lanes each needed to hold a
-- site's work queue still while they rebuilt that site. The platform offered no
-- way to do it, so each improvised the same hand-written UPDATE at a psql prompt:
--
--   mortgagecalculator_couk_adoption/HANDOFF_2026-08-03_continue_here.md:81-90
--     UPDATE site_work_items SET status='deferred', updated_at=NOW()
--      WHERE site_id=(...) AND status IN ('triaged','approved') AND id NOT IN (...)
--
-- That statement sets `status` and `updated_at` and NOTHING ELSE, which is exactly
-- why its 38 rows carry no provenance and why, twenty-two days later, one of them
-- blocked bugs_open/328's dispatch with a 23505 that reads as "already queued" and
-- means "queued and abandoned". Other lanes' hand-parks (loancalculator_rebuild_thread,
-- apis-uk-bees-lane) DID stamp themselves and are still fully attributable — so the
-- estate has grown SIX competing ad-hoc conventions for one act:
--   spec.parked_by (migration 389) · spec.deferred_reason · spec.not_dispatchable ·
--   result.deferred_by/-_reason/-_from_status · result.repair_284 (migration 442) ·
--   a reason appended to created_by.
--
-- THE FIX IS A MISSING VERB, NOT A BROKEN WRITER. No Go path produces the bad shape
-- (verified: no `UPDATE site_work_items … SET status='deferred'` exists in the repo,
-- and all six Go writers of that status pair it with an empty handler_agent). The
-- rows are made by people, so the remedy has to be reachable by people — at a psql
-- prompt, which is where every one of the four parks was performed. A Go action
-- would not have been called by any of them.
--
-- ── THE STRUCTURAL PROPERTY THAT MAKES THIS A FIX AND NOT A CONVENTION ──
--
-- p_parked_by, p_parked_reason and p_release_condition have NO DEFAULTS. You cannot
-- call this function without saying who you are, why, and what would release it.
-- The stamp is unskippable because it is an ARGUMENT, not a habit. A comment is not
-- a control on a tree this many sessions share (RFC_010 §2's own reasoning).
--
-- ── WHY `result` AND NOT `spec` — read before "improving" this ──
--
-- refreshOpenWorkItem (load_work_item_actions.go, `refreshOpenWorkItemSQL`) does
-- `SET summary = $3, spec = $4::jsonb` — it REPLACES spec wholesale whenever a
-- refreshOnConflict producer re-detects the same item_key. A spec-based park stamp
-- is therefore destroyed by exactly the re-detection a parked row is most exposed
-- to. `result` is merged by the estate's main status writer
-- (`result = COALESCE(result,'{}'::jsonb) || $3::jsonb`, v3_site_actions.go:6306)
-- and is untouched by the refresh. Migration 442 already stamps result.repair_284
-- for the same reason. This is why the two hand-conventions diverged and why the
-- `result` one is the durable half.
--
-- ── THE UNSAFE SIDE IS OFF BY DEFAULT (owner ruling 2026-08-02, RFC_010 §2) ──
--
-- p_apply DEFAULT false. The bare call is a DRY RUN that returns the plan and
-- changes nothing. Writing requires p_apply => true, typed by the caller, in the
-- statement a reviewer reads.
--
-- ── WHAT IT REFUSES, AND WHY EACH GUARD EARNS ITS LINE ──
--
--   * a NULL site_id            — a fleet-wide park is THE damage mode.
--   * a blank who/why/release   — that is the defect this file exists to close.
--   * more rows than p_max_rows — a typo must not silently sweep a site. The
--                                 largest real park on record is 60; the default
--                                 cap is 50 and the caller must raise it knowingly.
--   * status = 'claimed'        — a handler may be acting on that row right now.
--   * any terminal status       — parking a closed row would resurrect it into
--                                 idx_swi_dedup's kept set and block re-filing.
--
-- unpark_work_items restores each row to the status it was parked FROM, recorded
-- per row, and will only release rows carrying YOUR p_parked_by. That is deliberate:
-- as of 2026-08-25 sixty parked rows carry a live release condition set by another
-- lane ("un-park after rebuild verify"), and a blanket release would fire work
-- somebody is deliberately holding.
--
-- ROLLBACK: see 621_park_work_items_verb_ROLLBACK.sql (drops both functions; it does
-- NOT un-park anything, because rows parked through the verb are stamped and can be
-- released deliberately with unpark_work_items).

BEGIN;

-- ---------------------------------------------------------------------------
-- park_work_items — hold a site's dispatchable queue, with provenance.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION public.park_work_items(
    p_site_id           uuid,
    p_parked_by         text,
    p_parked_reason     text,
    p_release_condition text,
    p_item_types        text[] DEFAULT NULL,
    p_exclude_ids       uuid[] DEFAULT NULL,
    p_max_rows          integer DEFAULT 50,
    p_apply             boolean DEFAULT false
) RETURNS jsonb
LANGUAGE plpgsql
AS $function$
DECLARE
    v_parkable   text[] := ARRAY['triaged','approved','detected'];
    v_candidates bigint;
    v_parked     bigint := 0;
    v_by_type    jsonb;
    v_domain     text;
BEGIN
    IF p_site_id IS NULL THEN
        RAISE EXCEPTION 'park_work_items: p_site_id is required — a fleet-wide park is the damage mode this function exists to prevent';
    END IF;
    IF COALESCE(btrim(p_parked_by),'') = '' THEN
        RAISE EXCEPTION 'park_work_items: p_parked_by is required — name the lane or session doing the parking (bugs_open/396: 52 rows exist that nobody can attribute)';
    END IF;
    IF COALESCE(btrim(p_parked_reason),'') = '' THEN
        RAISE EXCEPTION 'park_work_items: p_parked_reason is required — say WHY, in a sentence the next reader will understand without your context';
    END IF;
    IF COALESCE(btrim(p_release_condition),'') = '' THEN
        RAISE EXCEPTION 'park_work_items: p_release_condition is required — say what would make these releasable ("after the rebuild verify", "when bugs_open/NNN closes"). A park with no stated end is how 20-day residue happens';
    END IF;
    IF p_max_rows IS NULL OR p_max_rows < 1 THEN
        RAISE EXCEPTION 'park_work_items: p_max_rows must be >= 1';
    END IF;

    SELECT domain INTO v_domain FROM sites WHERE id = p_site_id;
    IF v_domain IS NULL THEN
        RAISE EXCEPTION 'park_work_items: no site with id %', p_site_id;
    END IF;

    CREATE TEMP TABLE _park_candidates ON COMMIT DROP AS
    SELECT w.id, w.item_type, w.status
    FROM site_work_items w
    WHERE w.site_id = p_site_id
      AND w.status = ANY(v_parkable)
      AND (p_item_types  IS NULL OR w.item_type = ANY(p_item_types))
      AND (p_exclude_ids IS NULL OR NOT (w.id = ANY(p_exclude_ids)));

    SELECT count(*) INTO v_candidates FROM _park_candidates;

    SELECT COALESCE(jsonb_object_agg(t.item_type, t.n), '{}'::jsonb) INTO v_by_type
    FROM (SELECT item_type, count(*) AS n FROM _park_candidates GROUP BY 1) t;

    IF v_candidates > p_max_rows THEN
        RAISE EXCEPTION 'park_work_items: % candidate rows on % exceeds p_max_rows=% — re-run with a narrower p_item_types, or raise the cap deliberately. Breakdown: %',
            v_candidates, v_domain, p_max_rows, v_by_type;
    END IF;

    IF NOT p_apply THEN
        RETURN jsonb_build_object(
            'applied', false, 'domain', v_domain, 'would_park', v_candidates,
            'by_item_type', v_by_type,
            'note', 'DRY RUN — nothing changed. Re-run with p_apply => true to write.');
    END IF;

    UPDATE site_work_items w
       SET status = 'deferred',
           result = COALESCE(w.result, '{}'::jsonb) || jsonb_build_object(
                      'parked_by',          p_parked_by,
                      'parked_reason',      p_parked_reason,
                      'parked_from_status', w.status,
                      'parked_at',          to_char(now(),'YYYY-MM-DD"T"HH24:MI:SSOF'),
                      'release_condition',  p_release_condition,
                      'parked_by_verb',     'park_work_items')
      FROM _park_candidates c
     WHERE w.id = c.id
       AND w.status = c.status;   -- lost-update guard: skip anything that moved under us
    GET DIAGNOSTICS v_parked = ROW_COUNT;

    RETURN jsonb_build_object(
        'applied', true, 'domain', v_domain,
        'candidates', v_candidates, 'parked', v_parked,
        'raced_and_skipped', v_candidates - v_parked,
        'by_item_type', v_by_type,
        'release_with',
          format('SELECT unpark_work_items(%L::uuid, %L, p_apply => true);', p_site_id, p_parked_by));
END;
$function$;

-- ---------------------------------------------------------------------------
-- unpark_work_items — release YOUR park, restoring each row's own prior status.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION public.unpark_work_items(
    p_site_id   uuid,
    p_parked_by text,
    p_max_rows  integer DEFAULT 200,
    p_apply     boolean DEFAULT false
) RETURNS jsonb
LANGUAGE plpgsql
AS $function$
DECLARE
    v_restorable text[] := ARRAY['triaged','approved','detected'];
    v_candidates bigint;
    v_unparked   bigint := 0;
    v_bad        bigint;
    v_by_status  jsonb;
    v_domain     text;
BEGIN
    IF p_site_id IS NULL THEN
        RAISE EXCEPTION 'unpark_work_items: p_site_id is required';
    END IF;
    IF COALESCE(btrim(p_parked_by),'') = '' THEN
        RAISE EXCEPTION 'unpark_work_items: p_parked_by is required — you release YOUR OWN park. As of 2026-08-25, 60 parked rows carry another lane''s live release condition ("un-park after rebuild verify"); a blanket release would fire work somebody is deliberately holding';
    END IF;

    SELECT domain INTO v_domain FROM sites WHERE id = p_site_id;
    IF v_domain IS NULL THEN
        RAISE EXCEPTION 'unpark_work_items: no site with id %', p_site_id;
    END IF;

    CREATE TEMP TABLE _unpark_candidates ON COMMIT DROP AS
    SELECT w.id, w.result->>'parked_from_status' AS restore_to
    FROM site_work_items w
    WHERE w.site_id = p_site_id
      AND w.status  = 'deferred'
      AND w.result->>'parked_by' = p_parked_by
      AND w.result ? 'parked_from_status';

    SELECT count(*) INTO v_candidates FROM _unpark_candidates;
    SELECT count(*) INTO v_bad FROM _unpark_candidates WHERE NOT (restore_to = ANY(v_restorable));

    IF v_bad > 0 THEN
        RAISE EXCEPTION 'unpark_work_items: % row(s) record a parked_from_status outside %s — refusing rather than guessing', v_bad, v_restorable;
    END IF;
    IF v_candidates > p_max_rows THEN
        RAISE EXCEPTION 'unpark_work_items: % rows exceeds p_max_rows=% — raise the cap deliberately', v_candidates, p_max_rows;
    END IF;

    SELECT COALESCE(jsonb_object_agg(t.restore_to, t.n), '{}'::jsonb) INTO v_by_status
    FROM (SELECT restore_to, count(*) AS n FROM _unpark_candidates GROUP BY 1) t;

    IF NOT p_apply THEN
        RETURN jsonb_build_object(
            'applied', false, 'domain', v_domain, 'would_unpark', v_candidates,
            'restoring_to', v_by_status,
            'note', 'DRY RUN — nothing changed. Re-run with p_apply => true to write.');
    END IF;

    UPDATE site_work_items w
       SET status = c.restore_to,
           result = (COALESCE(w.result,'{}'::jsonb)
                     - 'parked_by' - 'parked_reason' - 'parked_from_status'
                     - 'parked_at' - 'release_condition' - 'parked_by_verb')
                    || jsonb_build_object('unparked_at',
                         to_char(now(),'YYYY-MM-DD"T"HH24:MI:SSOF'),
                         'unparked_by', p_parked_by)
      FROM _unpark_candidates c
     WHERE w.id = c.id
       AND w.status = 'deferred';
    GET DIAGNOSTICS v_unparked = ROW_COUNT;

    RETURN jsonb_build_object(
        'applied', true, 'domain', v_domain,
        'candidates', v_candidates, 'unparked', v_unparked,
        'raced_and_skipped', v_candidates - v_unparked,
        'restored_to', v_by_status);
END;
$function$;

-- ---------------------------------------------------------------------------
-- GUARDS — assert the post-conditions inside the transaction, and INDUCE the
-- refusals rather than asserting they exist. A verify block of bare SELECTs
-- cannot stop a COMMIT (ON_ERROR_STOP ignores a non-empty result), so this is a
-- DO block with RAISE, and every guard below is proven to FIRE, not merely to be
-- present. A guard nobody has watched fail is a guard you are assuming.
-- ---------------------------------------------------------------------------
DO $guard$
DECLARE
    v_site    uuid;
    v_fired   int := 0;
    v_res     jsonb;
BEGIN
    IF to_regprocedure('public.park_work_items(uuid,text,text,text,text[],uuid[],integer,boolean)') IS NULL THEN
        RAISE EXCEPTION '621: park_work_items did not get created';
    END IF;
    IF to_regprocedure('public.unpark_work_items(uuid,text,integer,boolean)') IS NULL THEN
        RAISE EXCEPTION '621: unpark_work_items did not get created';
    END IF;

    SELECT id INTO v_site FROM sites ORDER BY created_at LIMIT 1;
    IF v_site IS NULL THEN
        RAISE EXCEPTION '621: no sites row to test against';
    END IF;

    -- (1) blank parked_by must be refused
    BEGIN
        PERFORM park_work_items(v_site, '', 'why', 'when');
        RAISE EXCEPTION '621: GUARD DID NOT FIRE — blank p_parked_by was accepted';
    EXCEPTION WHEN raise_exception THEN
        IF SQLERRM LIKE '%GUARD DID NOT FIRE%' THEN RAISE; END IF;
        v_fired := v_fired + 1;
    END;

    -- (2) blank release condition must be refused
    BEGIN
        PERFORM park_work_items(v_site, 'guard-test', 'why', '   ');
        RAISE EXCEPTION '621: GUARD DID NOT FIRE — blank p_release_condition was accepted';
    EXCEPTION WHEN raise_exception THEN
        IF SQLERRM LIKE '%GUARD DID NOT FIRE%' THEN RAISE; END IF;
        v_fired := v_fired + 1;
    END;

    -- (3) NULL site must be refused
    BEGIN
        PERFORM park_work_items(NULL, 'guard-test', 'why', 'when');
        RAISE EXCEPTION '621: GUARD DID NOT FIRE — NULL p_site_id was accepted';
    EXCEPTION WHEN raise_exception THEN
        IF SQLERRM LIKE '%GUARD DID NOT FIRE%' THEN RAISE; END IF;
        v_fired := v_fired + 1;
    END;

    -- (4) unknown site must be refused
    BEGIN
        PERFORM park_work_items('00000000-0000-0000-0000-000000000000'::uuid,
                                'guard-test', 'why', 'when');
        RAISE EXCEPTION '621: GUARD DID NOT FIRE — unknown site was accepted';
    EXCEPTION WHEN raise_exception THEN
        IF SQLERRM LIKE '%GUARD DID NOT FIRE%' THEN RAISE; END IF;
        v_fired := v_fired + 1;
    END;

    -- (5) unpark with a blank parked_by must be refused
    BEGIN
        PERFORM unpark_work_items(v_site, '');
        RAISE EXCEPTION '621: GUARD DID NOT FIRE — unpark accepted a blank p_parked_by';
    EXCEPTION WHEN raise_exception THEN
        IF SQLERRM LIKE '%GUARD DID NOT FIRE%' THEN RAISE; END IF;
        v_fired := v_fired + 1;
    END;

    IF v_fired <> 5 THEN
        RAISE EXCEPTION '621: expected 5 guards to fire, got %', v_fired;
    END IF;

    -- (6) POSITIVE CONTROL — a well-formed DRY RUN must succeed, return
    --     applied=false, and change nothing. Without this, the five refusals
    --     above are equally consistent with a function that refuses everything.
    SELECT park_work_items(v_site, 'guard-test', 'positive control',
                           'immediately — this is a dry run') INTO v_res;
    IF COALESCE((v_res->>'applied')::boolean, true) <> false THEN
        RAISE EXCEPTION '621: POSITIVE CONTROL FAILED — a well-formed dry run did not return applied=false: %', v_res;
    END IF;
    IF NOT (v_res ? 'would_park') THEN
        RAISE EXCEPTION '621: POSITIVE CONTROL FAILED — dry run did not report would_park: %', v_res;
    END IF;

    -- (7) NEGATIVE CONTROL on the writer — the dry run must not have parked anything.
    IF EXISTS (SELECT 1 FROM site_work_items
                WHERE result->>'parked_by' = 'guard-test') THEN
        RAISE EXCEPTION '621: DRY RUN WROTE — rows carry parked_by=guard-test';
    END IF;

    RAISE NOTICE '621 OK: 5 refusals induced, dry run returned a plan, nothing written.';
END;
$guard$;

COMMIT;
