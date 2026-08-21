-- 542_css_patch_agent_base_integrity_refusal.sql
--
-- bugs_open/198, the arm that three incident waves walked through and migration
-- 318 did not close. 318 made the LLM fragment unrepresentable: the model returns
-- only `css_added`, the DB append is monotonic, `deploy_css` ships the row as
-- returned. That fix works exactly as built. What it ASSUMED is that
-- `css_themes.css_content` holds the site's stylesheet.
--
-- ON MANY SITES IT DOES NOT, AND THE APPEND THEN CONVERGES IN THE WRONG DIRECTION.
-- `assets/css/styles.css` has TWO producers that never read each other:
--
--   * webdesign-agent  → `render_css_from_spec` builds the file from the FK'd
--     palette/layout/typography rows + design_spec + css_snippets and git-commits
--     it. It never reads and never writes `css_themes.css_content`.
--   * css-patch-agent  → appends its patch rules to `css_themes.css_content` and
--     git-commits THE WHOLE ROW over that same file.
--
-- `install_site_composition_action.go:342-370` inserts theme rows with
-- `css_content = ''` deliberately ("the renderer reads composition via FKs"), so a
-- normally-composed site is BORN with an empty row while git serves 17–26KB. The
-- first css-patch dispatch then appends ~100 bytes to '' and deploys the result
-- wholesale: the DB writer cannot shrink, the FILE shrinks ~98%, and every
-- `--color-*` definition vanishes. Measured incidents: relojistas 2026-08-04;
-- idea.uk/noted/dartsonline/vonc/cookly/oufe 08-17..08-19;
-- remortgagecalculator.uk + loanzy.uk 08-21 (17,403 → 68 → 136 bytes, arithmetic
-- in the bug file proving the base was EMPTY, not truncated).
--
-- It is self-amplifying: with `:root` gone the render audit measures the wreckage,
-- files more `contrast_failure` items, and the promoter routes them straight back
-- to the agent that caused the damage. loancash took 11 items in 8 minutes; oufe
-- nine successive clobber commits in one day. Every run reported success.
--
-- WHAT THIS MIGRATION DOES — two independent defects, both config-only.
--
-- (A) A BASE-INTEGRITY GATE, because the existing one cannot see this.
--     `check_has_css` tests `current_css.css_content != null`. An EMPTY STRING is
--     not null, so it passes, and that is the arm every wave went through. The new
--     `check_base_integrity` gate is numeric and refuses two shapes:
--
--       css_len >= 4096      — the base is big enough to BE a stylesheet.
--       site_count <= 1      — this theme row is not shared with another site.
--
--     4096 BYTES (octet_length, not length: LANDMINES "length() on stored HTML is
--     CHARACTERS; a file's size is BYTES"). The number is census-derived, not
--     chosen: measured 2026-08-21 across every linked theme row, healthy rows run
--     13,650–26,917 bytes and every clobbered or stub row ever observed is
--     ≤ 2,381. 4096 sits in the empty middle with ~3× margin either way. It also
--     catches the shape the fleet backfill could NOT repair — three sites carrying
--     a 1,649-byte bare `:root` palette block and no layout rules, which deploys a
--     page where every variable resolves and every layout rule is gone.
--
--     `site_count <= 1` closes the shared-row door: `professional-dark` is one row
--     linked by finetuning.uk AND gaswholesalers.com, which serve DIFFERENT files
--     (13,988 vs 20,271 bytes). No backfill can make that row true for both, so a
--     patch there would push one site's CSS onto the other. It is refused until a
--     human splits the themes.
--
--     `fail_on_non_numeric: true` is deliberate. Without it a missing `css_len` —
--     i.e. this migration's own query edit not having landed — silently routes
--     EVERY run to the else arm and reads as "the guard is working". With it, that
--     state fails the step loudly instead. (Registered opt-in, live in the running
--     binary: pod-grepped 2026-08-21, 3 hits per replica, negative control 0.)
--
-- (B) THE COMPLETIONS STOP LYING. Every terminal here is a success-labelled
--     `complete_workflow`, so the parent dispatch loop's `complete_work_item`
--     stamps `complete` on the row whatever happened — the `complete_no_css` arm
--     minted 11 `complete`s for loancash.co.uk while doing nothing at all, and a
--     `complete_error` does the same for a real failure. A refusal that reads as a
--     repair is worse than no guard, because it also suppresses the evidence: any
--     census of unfixed contrast findings taken over these rows is a floor, not a
--     count (`bugs_open/296` §10.4).
--
--     So each non-success exit now stamps the item BEFORE its terminal, using
--     `update_work_item_status`. The dispatch loop's completion guard
--     (`load_work_item_actions.go`, WII-003) skips a row already sitting in
--     needs_human_review/failed/…, so the pre-stamp wins and the terminal becomes
--     a no-op on the ledger.
--
--       mark_base_unsafe  → needs_human_review, result_fields.parked_by =
--                           'css_base_integrity_guard_198'  (a DECISION: the base
--                           must be repaired by a human or by a webdesign render)
--       mark_no_css       → needs_human_review, parked_by = 'css_no_theme_198'
--                           (no style_collection at all — unpatchable by design)
--       mark_step_failed  → failed, no literal error_message so the routed
--                           `__step_error` is what gets recorded. `failed` and NOT
--                           needs_human_review on purpose: a genuine step error
--                           should enter the promoter floor's denominator and go
--                           through the shared retry ladder, which is what
--                           `update_work_item_status` does for that status.
--
--     The `parked_by` markers exist so the unpark sweep is EXACT. A parked row
--     holds its dedup key (needs_human_review is not in the dedup exclusion list),
--     so the same finding cannot re-file while parked — which is the desired
--     no-balloon behaviour, and also means restoring them to `detected` after a
--     backfill is the only route back. That query is in the workstream RUNBOOK.
--
-- (C) `deploy_css` gains `file_shrink_floor: 0.5`, the opt-in defence-in-depth
--     guard shipping as Go in the same task (git-adapter refuses a commit that
--     replaces an existing ≥2048-byte file with less than half its bytes). The key
--     is INERT until both the chassis and git-adapter images roll: the running
--     GitCommitAction reads config keys ad-hoc and ignores unknown ones, and an
--     old adapter drops the unknown JSON field at unmarshal. Both orders are safe
--     and no ordering constraint is claimed (owner ruling 2026-07-29 retired that
--     condition). Registered as DGH-008.
--
-- WHAT THIS DOES NOT DO, so nobody reads it as more than it is:
--   * It does not fix a site whose file is already clobbered — that is a restore,
--     and the fleet backfill of every empty row was completed 2026-08-21 by
--     another lane.
--   * It does not stop the agent authoring a patch that cannot match (the
--     `H3.H3` / `p.P` uppercased-tag-as-class family, three sites' evidence). That
--     is candidate 6 and a separate task.
--   * It does not guard webdesign-agent's own deploy — regeneration may honestly
--     shrink. Migration 543 makes that producer WRITE the row it never wrote,
--     which is what removes the divergence at source.
--
-- CONFIG IS LIVE IMMEDIATELY ON APPLY. There are 8 contrast items queued against
-- one site at time of writing, so this closes the door in front of live traffic.

-- Probe guard: tell the runner when this is already applied.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'css-patch-agent'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #> '{workflow,steps}' ? 'check_base_integrity'
    ) THEN
        RAISE EXCEPTION '198/542: already applied — check_base_integrity already exists';
    END IF;
END $$;

BEGIN;

-- README rule: every migration touching agent_definitions opens with a snapshot.
SELECT snapshot_agent('css-patch-agent',
  '542_css_patch_agent_base_integrity_refusal: pre-update');

-- ── DRIFT GUARD ────────────────────────────────────────────────────────────────
-- Assert the EXACT shape this migration rewires. A concurrent session may have
-- changed any of it since this file was written; a rewire against a shape that
-- moved would silently orphan a step or route an arm to the wrong terminal.
DO $$
DECLARE
    v_steps  jsonb;
    v_query  text;
BEGIN
    SELECT default_config #> '{workflow,steps}'
      INTO v_steps
      FROM agent_definitions
     WHERE type = 'css-patch-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_steps IS NULL THEN
        RAISE EXCEPTION '198/542: no live css-patch-agent row found';
    END IF;

    -- The edges being rewired, each asserted at its current value.
    IF v_steps #>> '{check_has_css,config,then_step}' <> 'plan_css_fix' THEN
        RAISE EXCEPTION '198/542 drift: check_has_css.then_step is %, expected plan_css_fix',
            v_steps #>> '{check_has_css,config,then_step}';
    END IF;
    IF v_steps #>> '{check_has_css,config,else_step}' <> 'complete_no_css' THEN
        RAISE EXCEPTION '198/542 drift: check_has_css.else_step is %, expected complete_no_css',
            v_steps #>> '{check_has_css,config,else_step}';
    END IF;
    IF v_steps #>> '{load_current_css,error_step}' <> 'complete_no_css' THEN
        RAISE EXCEPTION '198/542 drift: load_current_css.error_step is %, expected complete_no_css',
            v_steps #>> '{load_current_css,error_step}';
    END IF;
    IF v_steps #>> '{plan_css_fix,error_step}' <> 'complete_error' THEN
        RAISE EXCEPTION '198/542 drift: plan_css_fix.error_step is %, expected complete_error',
            v_steps #>> '{plan_css_fix,error_step}';
    END IF;
    IF v_steps #>> '{save_css_to_db,error_step}' <> 'complete_error' THEN
        RAISE EXCEPTION '198/542 drift: save_css_to_db.error_step is %, expected complete_error',
            v_steps #>> '{save_css_to_db,error_step}';
    END IF;
    IF v_steps #>> '{deploy_css,error_step}' <> 'complete_error' THEN
        RAISE EXCEPTION '198/542 drift: deploy_css.error_step is %, expected complete_error',
            v_steps #>> '{deploy_css,error_step}';
    END IF;
    IF v_steps #>> '{deploy_css,action}' <> 'git_commit' THEN
        RAISE EXCEPTION '198/542 drift: deploy_css.action is %, expected git_commit',
            v_steps #>> '{deploy_css,action}';
    END IF;

    -- 318's fix must still be in place — this migration builds on top of it and
    -- would be meaningless against a workflow that still round-trips the document.
    IF v_steps #>> '{save_css_to_db,config,params,1}' <> 'css_fix.result.css_added' THEN
        RAISE EXCEPTION '198/542 drift: save_css_to_db no longer appends css_added (318 undone?) — got %',
            v_steps #>> '{save_css_to_db,config,params,1}';
    END IF;

    -- The query being replaced, asserted verbatim: the replacement below only
    -- ADDS two columns, so if the FROM/WHERE has moved, the replacement is wrong.
    v_query := v_steps #>> '{load_current_css,config,query}';
    IF v_query <> 'SELECT ct.id::text as theme_id, ct.css_content, ct.name as theme_name FROM css_themes ct JOIN style_collections sc ON sc.css_theme_id = ct.id JOIN sites s ON s.style_collection_id = sc.id WHERE s.id = $1::uuid' THEN
        RAISE EXCEPTION '198/542 drift: load_current_css.query is not the expected text — got: %', v_query;
    END IF;
END $$;

-- ── (A) the query learns to measure its own base ───────────────────────────────
-- Two added columns, same JOIN. octet_length = BYTES. site_count counts every site
-- whose style_collection points at this theme row, which is how the shared-row
-- case becomes visible to a condition.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_current_css,config,query}',
         to_jsonb(
           'SELECT ct.id::text as theme_id, ct.css_content, ct.name as theme_name, '
           || 'octet_length(ct.css_content) AS css_len, '
           || '(SELECT count(*) FROM sites s2 JOIN style_collections sc2 ON s2.style_collection_id = sc2.id WHERE sc2.css_theme_id = ct.id) AS site_count '
           || 'FROM css_themes ct '
           || 'JOIN style_collections sc ON sc.css_theme_id = ct.id '
           || 'JOIN sites s ON s.style_collection_id = sc.id '
           || 'WHERE s.id = $1::uuid'
         )
       ),
       updated_at = NOW()
 WHERE type = 'css-patch-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── (A) the gate itself, plus the rewire that puts it in the path ──────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,check_base_integrity}',
           jsonb_build_object(
             'action', 'conditional_branch',
             'description',
               'bugs_open/198: refuse to patch unless the css_themes base is a real stylesheet '
               || '(>= 4096 BYTES, octet_length) and is not shared with another site. An empty or '
               || 'stub row passes check_has_css (empty string is not null) and its append then '
               || 'deploys wholesale over a 17-26KB file, wiping every --color-* definition. '
               || 'Floor is census-derived: healthy rows 13,650-26,917 B, every clobbered row <= 2,381 B.',
             'config', jsonb_build_object(
               'condition',           'current_css.css_len >= 4096 AND current_css.site_count <= 1',
               'fail_on_non_numeric', true,
               'then_step',           'plan_css_fix',
               'else_step',           'mark_base_unsafe'
             )
           )
         ),
         '{workflow,steps,check_has_css,config,then_step}',
         to_jsonb('check_base_integrity'::text)
       ),
       updated_at = NOW()
 WHERE type = 'css-patch-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── (B) the three honest-status steps ──────────────────────────────────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(
             default_config,
             '{workflow,steps,mark_base_unsafe}',
             jsonb_build_object(
               'action', 'update_work_item_status',
               'description',
                 'bugs_open/198: park the item as a DECISION, not a repair. Stamped before the '
                 || 'terminal so the dispatch loop''s completion guard (WII-003) leaves it alone.',
               'next_step', 'complete_refused',
               'config', jsonb_build_object(
                 'status',        'needs_human_review',
                 'error_message',
                   'css-patch refused (bugs_open/198): the linked css_themes row is below 4096 bytes '
                   || 'or is shared by more than one site, so appending a patch and deploying the row '
                   || 'would replace the live stylesheet wholesale. Repair the base first — backfill '
                   || 'the row from the deployed styles.css, or let a webdesign-agent render persist it '
                   || '(migration 543) — or split a shared theme into per-site rows. Then sweep items '
                   || 'carrying result->>''parked_by'' = ''css_base_integrity_guard_198'' back to detected.',
                 'result_fields', jsonb_build_object('parked_by', 'css_base_integrity_guard_198')
               )
             )
           ),
           '{workflow,steps,mark_no_css}',
           jsonb_build_object(
             'action', 'update_work_item_status',
             'description',
               'bugs_open/198: the no-stylesheet exit used to mint complete (loancash.co.uk: 11 items '
               || 'in 8 minutes, every one complete, nothing done). Park it instead.',
             'next_step', 'complete_no_css',
             'config', jsonb_build_object(
               'status',        'needs_human_review',
               'error_message',
                 'css-patch could not run (bugs_open/198): this site has no linked css_themes row via '
                 || 'style_collections, so there is no stylesheet for the agent to patch. The finding '
                 || 'may still be real — it needs a different repair path, not this handler.',
               'result_fields', jsonb_build_object('parked_by', 'css_no_theme_198')
             )
           )
         ),
         '{workflow,steps,mark_step_failed}',
         jsonb_build_object(
           'action', 'update_work_item_status',
           'description',
             'bugs_open/198: a real step failure used to ride complete_error, a success-labelled '
             || 'complete_workflow, and read as complete. Record it as failed — through the shared '
             || 'retry ladder — with the routed __step_error as the message. No literal error_message '
             || 'on purpose: omitting it is what makes the action record the actual step error.',
           'next_step', 'complete_error',
           'config', jsonb_build_object(
             'status',        'failed',
             'result_fields', jsonb_build_object('failed_by', 'css_patch_agent_198')
           )
         )
       ),
       updated_at = NOW()
 WHERE type = 'css-patch-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── (B) the refusal terminal ───────────────────────────────────────────────────
-- A separate terminal from complete_no_css so the two refusal reasons stay
-- distinguishable in the orchestration record.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,complete_refused}',
         jsonb_build_object(
           'action', 'complete_workflow',
           'description', 'Refused: the css_themes base is not safe to deploy (bugs_open/198)',
           'config', jsonb_build_object(
             'output_fields',    jsonb_build_array(),
             'success_message',  'CSS patch refused — unsafe stylesheet base (bugs_open/198); item parked for review'
           )
         )
       ),
       updated_at = NOW()
 WHERE type = 'css-patch-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── (B) rewire every non-success exit through its stamping step ────────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(
             jsonb_set(
               jsonb_set(
                 default_config,
                 '{workflow,steps,check_has_css,config,else_step}',
                 to_jsonb('mark_no_css'::text)
               ),
               '{workflow,steps,load_current_css,error_step}',
               to_jsonb('mark_no_css'::text)
             ),
             '{workflow,steps,plan_css_fix,error_step}',
             to_jsonb('mark_step_failed'::text)
           ),
           '{workflow,steps,save_css_to_db,error_step}',
           to_jsonb('mark_step_failed'::text)
         ),
         '{workflow,steps,deploy_css,error_step}',
         to_jsonb('mark_step_failed'::text)
       ),
       updated_at = NOW()
 WHERE type = 'css-patch-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── (C) opt deploy_css into the git-writer shrink floor (inert until both rolls)─
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,deploy_css,config,file_shrink_floor}',
         to_jsonb(0.5::numeric)
       ),
       updated_at = NOW()
 WHERE type = 'css-patch-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── VERIFY ─────────────────────────────────────────────────────────────────────
-- DO/RAISE, not a SELECT: ON_ERROR_STOP ignores a non-empty result set, so a
-- verify block made of SELECTs cannot stop the COMMIT.
DO $$
DECLARE
    v_steps jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps}'
      INTO v_steps
      FROM agent_definitions
     WHERE type = 'css-patch-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    -- the gate exists, with the exact condition and the loud-on-missing-field flag
    IF v_steps #>> '{check_base_integrity,config,condition}'
       <> 'current_css.css_len >= 4096 AND current_css.site_count <= 1' THEN
        RAISE EXCEPTION '198/542 verify: check_base_integrity.condition wrong — got %',
            v_steps #>> '{check_base_integrity,config,condition}';
    END IF;
    IF (v_steps #> '{check_base_integrity,config,fail_on_non_numeric}') <> 'true'::jsonb THEN
        RAISE EXCEPTION '198/542 verify: fail_on_non_numeric not set on check_base_integrity';
    END IF;

    -- the query actually measures, or the gate above is testing a field that
    -- never arrives (which fail_on_non_numeric turns into a loud failure, but the
    -- correct state is that it arrives)
    IF position('css_len' in (v_steps #>> '{load_current_css,config,query}')) = 0
       OR position('site_count' in (v_steps #>> '{load_current_css,config,query}')) = 0 THEN
        RAISE EXCEPTION '198/542 verify: load_current_css does not select css_len/site_count';
    END IF;
    IF position('octet_length' in (v_steps #>> '{load_current_css,config,query}')) = 0 THEN
        RAISE EXCEPTION '198/542 verify: css_len must be measured in BYTES (octet_length)';
    END IF;

    -- every edge landed
    IF v_steps #>> '{check_has_css,config,then_step}' <> 'check_base_integrity' THEN
        RAISE EXCEPTION '198/542 verify: check_has_css.then_step not rewired';
    END IF;
    IF v_steps #>> '{check_has_css,config,else_step}' <> 'mark_no_css' THEN
        RAISE EXCEPTION '198/542 verify: check_has_css.else_step not rewired';
    END IF;
    IF v_steps #>> '{load_current_css,error_step}' <> 'mark_no_css' THEN
        RAISE EXCEPTION '198/542 verify: load_current_css.error_step not rewired';
    END IF;
    IF v_steps #>> '{plan_css_fix,error_step}' <> 'mark_step_failed'
       OR v_steps #>> '{save_css_to_db,error_step}' <> 'mark_step_failed'
       OR v_steps #>> '{deploy_css,error_step}' <> 'mark_step_failed' THEN
        RAISE EXCEPTION '198/542 verify: an error_step was not rewired to mark_step_failed';
    END IF;

    -- the stamping steps exist and carry their markers
    IF v_steps #>> '{mark_base_unsafe,config,status}' <> 'needs_human_review'
       OR v_steps #>> '{mark_base_unsafe,config,result_fields,parked_by}' <> 'css_base_integrity_guard_198' THEN
        RAISE EXCEPTION '198/542 verify: mark_base_unsafe wrong';
    END IF;
    IF v_steps #>> '{mark_no_css,config,status}' <> 'needs_human_review'
       OR v_steps #>> '{mark_no_css,config,result_fields,parked_by}' <> 'css_no_theme_198' THEN
        RAISE EXCEPTION '198/542 verify: mark_no_css wrong';
    END IF;
    IF v_steps #>> '{mark_step_failed,config,status}' <> 'failed' THEN
        RAISE EXCEPTION '198/542 verify: mark_step_failed wrong';
    END IF;
    -- mark_step_failed must NOT carry a literal error_message: its absence is what
    -- makes the action record the routed __step_error instead of a fixed string.
    IF v_steps #> '{mark_step_failed,config}' ? 'error_message' THEN
        RAISE EXCEPTION '198/542 verify: mark_step_failed must not set error_message';
    END IF;

    -- every stamping step reaches a real terminal
    IF v_steps #>> '{mark_base_unsafe,next_step}' <> 'complete_refused'
       OR NOT (v_steps ? 'complete_refused') THEN
        RAISE EXCEPTION '198/542 verify: complete_refused missing or unreachable';
    END IF;
    IF v_steps #>> '{mark_no_css,next_step}' <> 'complete_no_css' THEN
        RAISE EXCEPTION '198/542 verify: mark_no_css does not reach complete_no_css';
    END IF;
    IF v_steps #>> '{mark_step_failed,next_step}' <> 'complete_error' THEN
        RAISE EXCEPTION '198/542 verify: mark_step_failed does not reach complete_error';
    END IF;

    -- the opt-in floor
    IF (v_steps #> '{deploy_css,config,file_shrink_floor}') <> '0.5'::jsonb THEN
        RAISE EXCEPTION '198/542 verify: deploy_css.file_shrink_floor not 0.5 — got %',
            v_steps #> '{deploy_css,config,file_shrink_floor}';
    END IF;

    -- 318 must survive this migration untouched
    IF position('patched_css' in (v_steps::text)) > 0 THEN
        RAISE EXCEPTION '198/542 verify: patched_css reappeared — 318 undone';
    END IF;

    RAISE NOTICE '198/542: verified — base-integrity gate live, three exits stamp before their terminal, floor opted in at 0.5';
END $$;

COMMIT;
