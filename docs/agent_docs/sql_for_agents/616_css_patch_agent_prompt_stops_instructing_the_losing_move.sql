-- 616_css_patch_agent_prompt_stops_instructing_the_losing_move.sql
--
-- bugs_open/390, commit 1 of 3. The cheap half, live on apply, no image roll.
--
-- WHAT IS WRONG. css-patch-agent repairs a `contrast_failure` by appending a rule
-- to the site theme and is then marked `complete` because a rule was WRITTEN. On
-- the measured population that rule cannot take effect, and the prompt is the
-- reason: it currently says, verbatim,
--
--   "Rely on the cascade: an appended rule with the same or higher specificity
--    overrides the earlier declaration. Repeat the offending selector exactly as
--    it appears above (or more specifically) so your override wins."
--
-- Both sentences are false for this agent's actual situation, and following them
-- produces the losing rule every time:
--
--   * "as it appears above" - the offending selector does not appear above. The
--     model is handed `spec.selector`, the audit's ancestor-derived selector, and
--     the declaration that beats it is not in this stylesheet at all.
--   * "the same specificity overrides" - it does not, when the competitor is
--     LATER in source order, and the competitor always is. Page-level <style>
--     blocks live in `page_components.rendered_html` and `assemblePage`
--     (rerender_single_page_action.go:560-720, sections written at :700) emits
--     them inside <main>, always after the <link> in <head>. That ordering is
--     structural, on every page, and cannot be reordered from here.
--
-- THE MEASUREMENT [MEASURED 2026-08-25, and it could have come out otherwise].
-- 40 random completed findings with real (non-invented) selectors across 7 sites,
-- each page and theme fetched and the governing `color` declaration located:
--
--     winning declaration in the page block, OUT-SPECIFYING the filed selector  33
--     winning declaration in the page block, lower specificity                   6
--     winning declaration in this stylesheet                                     0
--     not located (likely over_image)                                            1
--
-- Near-uniform shape: filed `TAG.class` (0,1,1) against page `.section .class`
-- (0,2,0). Damage over site_work_items UNION site_work_items_archive: 75 of 151
-- real-selector (page, selector) pairings that ever reached `complete` were filed
-- AGAIN afterwards, and 97 re-filings carry byte-identical fg/bg. At the artefact:
-- noted.co.uk holds 16 appended contrast fixes for 5 distinct selectors.
--
-- WHY !important, IN AN ESTATE THAT AVOIDS IT. Because this agent's shape is the
-- one shape CSS's own escape hatch is for: an APPEND-ONLY writer to a file that is
-- always earlier in source order than its competitors and never sees them. The two
-- Go precedents avoid !important because each owns a surface this agent lacks -
-- fix_component_template_action.go:440-444 appends AFTER the slot's own <style> so
-- it wins on order, and emit_sprite_css_action.go:248 authors both competing rules.
-- Neither option is available here.
--
-- The premise is measured, not assumed [MEASURED 2026-08-25]: page-level colour
-- declarations already carrying !important, across 8 sites' homepages, are 0 of 812
-- (noted 0/157, idea 0/210, cookly 0/120, vonc 0/94, remortgagecalculator 0/78,
-- oufe 0/72, loanzy 0/42, loancash 0/39). So an appended important colour rule wins
-- today in every measured case. It is a STATE measurement and it expires - re-run
-- RUNBOOK section 5 before quoting it.
--
-- THIS IS THE INTERIM, NOT THE DESTINATION. Commit 2 of this lane teaches the
-- render audit to record WHICH declaration wins and where it lives, and commit 3
-- replaces the blanket instruction below with that measured requirement, so
-- !important is used only where importance parity actually demands it. Whoever
-- writes commit 3 must assert THIS file's text in its drift guard, not 318's.
--
-- WHAT THIS DOES NOT DO, so nobody reads it as more than it is:
--   * It does not help the ~56% of live inflow that never reaches the prompt: a
--     site with no linked theme parks at css_no_theme_198, and one on a shared or
--     short theme parks at css_base_integrity_guard_198 (migration 542's gate).
--   * It does not stop a repair being ERASED. Migration 543's persist_css_to_theme
--     rewrites css_themes.css_content byte-for-byte at the site's next design run,
--     taking every appended patch with it (agritec.uk: 5 repairs erased 2026-08-25
--     12:09:57, items still `complete`). Filed separately.
--   * It does not make `complete` mean the text became readable. That is the
--     completion-gate half, deliberately deferred by owner decision 2026-08-25.
--   * It does not repair the underlying token. A pale ink on a pale ground is a
--     palette defect; this wins the pairing the audit measured and suppresses no
--     evidence, because every other pairing the token breaks is still filed.
--
-- NO ORDERING CONSTRAINT IS CLAIMED (owner ruling 2026-07-29). This is config only
-- and is live on apply; it depends on no unshipped code and nothing depends on it.

-- ── COUNCIL ROUND 1: REVISE, and what it found (2026-08-25) ──────────────────
-- Correlation ef5f9a0d-48a4-468e-afb4-7b6a06520f7f. Gating objection from
-- debug_historian, severity HIGH, and guardian + editquality raised the same
-- thing independently — three seats on one issue, which is a real signal:
--
--   "The UPDATE (and the drift-guard's SELECT that feeds v_prompt) scope on
--    type='css-patch-agent' AND is_active AND is_snapshot=false AND deleted_at
--    IS NULL with no version tie-break. Four agent types carry TWO active
--    definition rows and only the higher version is ever loaded."
--
-- ⚠ THE LANDMINE IS REAL AND LIVE, and I checked it myself rather than taking
-- the verdict's word: exactly four types carry two active non-snapshot rows
-- today — chief-strategist, content-creator, content-creator-contact,
-- site-component-architect. **css-patch-agent is NOT one of them** (1 row,
-- version 1, id 661a27b9-9b09-4a9d-a737-34871dee3bd8), so the objection did not
-- bite here. It is still correct about the SHAPE: the guards assumed a single
-- row without asserting it, and the assumption would fail silently the day this
-- agent gains a second.
--
-- Second objection, low severity, and it is the better catch: the drift guard's
-- v_old literal and the UPDATE's dollar-quoted $old$ literal were TWO SEPARATELY
-- AUTHORED COPIES of the same string. Diverge by one character and the guard
-- passes while replace() no-ops — UPDATE 1, nothing changed. That is the same
-- class of defect as the length constant this file already records getting
-- wrong. **Both are now fixed by construction: the apply is ONE DO block, the
-- literals are declared once, and the row count is asserted before anything is
-- read or written.**
--
-- ⚠ HONEST RECORD OF WHAT WAS ACTUALLY APPLIED. The live apply happened at
-- ~16:47 BST on 2026-08-25, BEFORE this verdict landed, in the form WITHOUT the
-- row-count assertion and with the duplicated literal. Nothing went wrong — the
-- single-row precondition was true then and is true now, verified twice — but
-- the applied form was not this form, and re-running is a loud no-op via the
-- probe guard rather than a re-apply of the hardened version. Applying before
-- the verdict also made the council's own read-only checks report
-- `has_old_text: false`, i.e. the precondition reading as ABSENT because I had
-- already consumed it; prior_art_librarian asked for exactly that check to be
-- run before the replace executes. Worth knowing before anyone reads those
-- answers as evidence of drift.

-- Probe guard: tell the runner when this is already applied.
DO $probe$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'css-patch-agent'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #>> '{workflow,steps,plan_css_fix,config,prompt_template}'
              LIKE '%can never win on position%'
    ) THEN
        RAISE EXCEPTION '390/616: already applied - the prompt already carries the source-order correction';
    END IF;
END $probe$;

BEGIN;

-- README rule: every migration touching agent_definitions opens with a snapshot.
SELECT snapshot_agent('css-patch-agent',
  '616_css_patch_agent_prompt_stops_instructing_the_losing_move: pre-update');

-- == GUARD AND APPLY, IN ONE BLOCK =========================================
-- Deliberately ONE block, and that is the fix for the council's second
-- objection. The guard and the replace need the same two literals; declaring
-- them once and using the variables for both makes divergence unrepresentable
-- rather than merely detectable. Two separately-authored copies of one string is
-- how a guard passes while the replace no-ops.
DO $apply$
DECLARE
    v_old text := '- Rely on the cascade: an appended rule with the same or higher specificity overrides the earlier declaration. Repeat the offending selector exactly as it appears above (or more specifically) so your override wins.';
    v_new text := '- IMPORTANT: the declaration you must beat is usually NOT in the stylesheet above. Each page carries its own <style> block, and the platform emits that block inside <main>, always AFTER this file is linked in <head>. Your appended rule is therefore always earlier in source order, so it can never win on position: it wins only by higher specificity, or by !important.
- Measured across 40 completed repairs on 7 sites (2026-08-25, bugs_open/390): the winning declaration sat in the page block in 39 of them and in this stylesheet in 0, and it out-specified the audited selector in 33 - typically a section-scoped `.section-class .element-class` beating an audited `TAG.element-class`.
- So: repeat the audited selector exactly as the finding states it, and mark ONLY the single property you are correcting as !important, for example `color: #123456 !important`. No other property gets !important, and no other property changes.';
    v_rows   int;
    v_steps  jsonb;
    v_prompt text;
    v_at     int;
BEGIN
    -- ⚠ ROW-COUNT ASSERTION FIRST (council round 1, debug_historian, HIGH).
    -- Four agent types in this estate carry TWO active non-snapshot rows and
    -- only the higher version is ever loaded. css-patch-agent is not one of them
    -- today — but "not today" is not a guard, and every read and write below
    -- assumes exactly one row. Assert it rather than inherit it.
    SELECT count(*) INTO v_rows
      FROM agent_definitions
     WHERE type = 'css-patch-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '390/616: expected exactly ONE live css-patch-agent row, found % - this migration''s guards and its UPDATE both assume a single row, and a version tie-break is needed before either is safe', v_rows;
    END IF;

    SELECT default_config #> '{workflow,steps}'
      INTO v_steps
      FROM agent_definitions
     WHERE type = 'css-patch-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_steps IS NULL THEN
        RAISE EXCEPTION '390/616: no live css-patch-agent row found';
    END IF;
    IF v_steps #>> '{plan_css_fix,action}' <> 'execute_llm_prompt' THEN
        RAISE EXCEPTION '390/616 drift: plan_css_fix.action is %, expected execute_llm_prompt',
            v_steps #>> '{plan_css_fix,action}';
    END IF;
    -- 318's fix must still be in place: this migration is meaningless against a
    -- workflow that round-trips the whole document instead of appending.
    IF v_steps #>> '{save_css_to_db,config,params,1}' <> 'css_fix.result.css_added' THEN
        RAISE EXCEPTION '390/616 drift: save_css_to_db no longer appends css_added (318 undone?) - got %',
            v_steps #>> '{save_css_to_db,config,params,1}';
    END IF;

    v_prompt := v_steps #>> '{plan_css_fix,config,prompt_template}';
    IF v_prompt IS NULL THEN
        RAISE EXCEPTION '390/616 drift: plan_css_fix.config.prompt_template does not exist as a string on the live row (guardian round 1: confirm the path before claiming to replace text inside it)';
    END IF;
    IF position('"css_added"' in v_prompt) = 0 THEN
        RAISE EXCEPTION '390/616 drift: prompt no longer names the css_added output key';
    END IF;

    -- The sentence being replaced, asserted VERBATIM and asserted to appear
    -- EXACTLY ONCE. Once, because replace() rewrites every occurrence and a
    -- second would mean the prompt has a shape this migration has not read.
    --
    -- Deliberately NOT counted by dividing a length difference by a hand-typed
    -- constant: that constant is a second copy of the sentence's length, it
    -- drifts from the sentence silently, and getting it wrong turns this guard
    -- into one that passes for the wrong reason. Search twice instead - and the
    -- string searched for is the SAME VARIABLE the replace below uses, so the
    -- guard and the edit cannot disagree about what they are looking for.
    v_at := position(v_old in v_prompt);
    IF v_at = 0 THEN
        RAISE EXCEPTION '390/616 drift: the cascade bullet 318 planted is not present verbatim - the prompt has moved, and a blind replace here would be a silent no-op';
    END IF;
    IF position(v_old in substr(v_prompt, v_at + length(v_old))) > 0 THEN
        RAISE EXCEPTION '390/616 drift: the cascade bullet appears more than once - replace() would rewrite an occurrence this migration has not read';
    END IF;

    -- The edit. One key, one replace(), using the SAME literals the guard used.
    UPDATE agent_definitions
       SET default_config = jsonb_set(
               default_config,
               '{workflow,steps,plan_css_fix,config,prompt_template}',
               to_jsonb(replace(v_prompt, v_old, v_new)),
               false
           ),
           updated_at = NOW()
     WHERE type = 'css-patch-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '390/616: UPDATE touched % rows, expected exactly 1', v_rows;
    END IF;
END $apply$;

-- == VERIFY ==================================================================
-- A block of SELECTs cannot stop a COMMIT (ON_ERROR_STOP ignores a non-empty
-- result), so every post-condition here RAISEs.
DO $verify$
DECLARE
    v_steps    jsonb;
    v_prompt   text;
    v_dangling int;
BEGIN
    SELECT default_config #> '{workflow,steps}'
      INTO v_steps
      FROM agent_definitions
     WHERE type = 'css-patch-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    v_prompt := v_steps #>> '{plan_css_fix,config,prompt_template}';

    -- 1. The false instruction is gone.
    IF position('Repeat the offending selector exactly as it appears above' in v_prompt) > 0 THEN
        RAISE EXCEPTION '390/616 verify: the old cascade bullet survived the replace';
    END IF;

    -- 2. The correction is present. Both halves: the mechanism and the action.
    IF position('can never win on position' in v_prompt) = 0 THEN
        RAISE EXCEPTION '390/616 verify: the source-order correction is absent';
    END IF;
    IF position('mark ONLY the single property you are correcting as !important' in v_prompt) = 0 THEN
        RAISE EXCEPTION '390/616 verify: the actionable instruction is absent';
    END IF;

    -- 3. The output contract the rest of the workflow reads is untouched.
    IF position('"css_added"' in v_prompt) = 0
       OR position('"changes_summary"' in v_prompt) = 0 THEN
        RAISE EXCEPTION '390/616 verify: the JSON output contract was damaged by the edit';
    END IF;

    -- 4. The read-only stylesheet context and the template variables still render.
    IF position('{{.current_css.css_content}}' in v_prompt) = 0
       OR position('{{.input_data.spec.description}}' in v_prompt) = 0 THEN
        RAISE EXCEPTION '390/616 verify: a template variable was lost';
    END IF;

    -- 5. 546's post-condition, run again because this file edits the same row:
    --    every edge in the workflow must still resolve to a real step.
    SELECT count(*) INTO v_dangling
      FROM (
        SELECT e.v->>'next_step' AS tgt FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v ? 'next_step'
        UNION ALL
        SELECT e.v->>'error_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v ? 'error_step'
        UNION ALL
        SELECT e.v->'config'->>'then_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'then_step'
        UNION ALL
        SELECT e.v->'config'->>'else_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'else_step'
      ) AS edges
     WHERE tgt IS NOT NULL AND NOT (v_steps ? tgt);
    IF v_dangling > 0 THEN
        RAISE EXCEPTION '390/616 verify: % workflow edge(s) point at a step that does not exist', v_dangling;
    END IF;

    -- 6. 542/546's parks are still wired - this migration must not have disturbed them.
    IF v_steps #>> '{check_has_css,config,else_step}' <> 'mark_no_css'
       OR v_steps #>> '{check_saved,config,else_step}' <> 'mark_append_refused' THEN
        RAISE EXCEPTION '390/616 verify: 542/546 park wiring disturbed';
    END IF;

    RAISE NOTICE '390/616: verified - prompt corrected, output contract intact, 0 dangling edges, parks still wired';
END $verify$;

COMMIT;
