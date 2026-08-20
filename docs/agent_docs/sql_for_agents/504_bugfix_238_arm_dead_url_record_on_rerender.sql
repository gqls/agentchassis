-- FILE: docs/agent_docs/sql_for_agents/504_bugfix_238_arm_dead_url_record_on_rerender.sql
--
-- bugs_open/238 — arm the RECORD-ONLY half of the dead-URL guard (PBP-040) on
-- the one live step whose action is `rerender_page_sections`.
--
-- WHY THIS FILE EXISTS, AND WHY IT IS NOT A _HOLD.
--
-- 238's detection half shipped as Go on 2026-08-11 and has been ARMED NOWHERE
-- ever since. Measured 2026-08-19/20: ZERO active agent_definitions rows carry
-- `record_dead_url_controls` or `refuse_dead_url_controls`, and ZERO
-- `dead_url_control` work items have ever been filed. The register still called
-- the Go half "INERT until the chassis image rolls"; it has rolled many times
-- (live since v1.0.1291, re-verified on v1.0.1317 below), so the only thing
-- standing between this estate and its detection was a config key nobody set.
--
-- 380_bugfix_238_arm_dead_url_guard_HOLD.sql is `_HOLD` because it arms the
-- REFUSAL, which can block a rebuild. This file arms the RECORD-ONLY emit on the
-- re-render path: it files a work item and changes no disposition, so there is no
-- blocking behaviour to sequence and no reason to keep it out of the runner.
--
-- ORDER CHECK (the thing 380's header is emphatic about, applied to this file).
-- A config key naming behaviour the running binary does not have is a no-op that
-- LOOKS applied. Verified before writing, at the artefact and with controls in
-- both directions:
--   image  docker.io/aqls/agent-chassis:v1.0.1317
--   label  org.opencontainers.image.revision = 2d13d530d2943831641ff6e51e4c92d8eb4b6c10
--   git merge-base --is-ancestor 51f56d0c9 2d13d530d   -> PASS (the guard's commit)
--   control+  the stamp is trivially its own ancestor  -> PASS
--   control-  today's HEAD is NOT an ancestor          -> PASS (so the probe discriminates)
-- Do NOT use `strings` for this (absent from the debian-slim images; three wrong
-- readings in one day, bugs_open/249). Read the stamp of THIS service, not the
-- fleet tag.
--
-- WHAT IT DOES. Sets `record_dead_url_controls: true` on `page-rerender`'s
-- `rerender_sections` step. The consumer is
-- rerender_page_sections_action.go's `recordDeadURLControls(params.StepConfig.Config)`;
-- when armed, a section that RENDERED but left a src=/href= attribute empty is
-- appended to `resolution.DeadURLSlots` and files a page+slot-keyed
-- `dead_url_control` item via `emitSectionDeadControlItem`. Record-only on this
-- path by design: the re-render MERGES stored ⊕ fresh and is the repair vehicle,
-- so a re-render that refused on the state it was dispatched to fix would
-- deadlock its own remedy.
--
-- ⚠ THE STEP IS AT THE TOP LEVEL HERE, unlike 380's. page-content-writer nests
-- its render steps inside the `process_sections_loop` sub_workflow; page-rerender
-- does not. Read live 2026-08-20 — all ten of its steps sit at
-- `{workflow,steps,<name>}`, and the one carrying this action is
-- `rerender_sections`. A jsonb path copied from 380 would find nothing here and
-- read as "no such step", which is the census trap this bug family keeps hitting.
--
-- COUNTED IN THE UNIT OF THE CHANGE (WRONG_CALLS 2026-08-11). The question is how
-- many STEPS run this action, not how many agents mention it. A text match on the
-- definition returns three agent types (`page-rerender`, `fix-proposer`,
-- `council-gate`); a jsonb traversal of `$.**.steps` for
-- `action = 'rerender_page_sections'` returns exactly ONE step, fleet-wide. The
-- other two mention the string in prompt/scope text, not as a step action. The
-- verify block below re-derives that count at apply time and demands that every
-- such step carries the flag, so a second one appearing later fails this file
-- loudly rather than shipping partial coverage under a complete label.
--
-- DECLARATION FIRST, THEN ARM. `record_dead_url_controls` was UNDECLARED on
-- RerenderPageSectionsInputSpec until commit bb6600e48 (2026-08-20). Arming an
-- undeclared key does not break the step — that spec is CheckConfig, not
-- StrictConfig, verified at the deciding arm (platform/validation/workflow.go:185-195)
-- — but it makes a live working setting read as "keys this action does not read,
-- silently ignored at execution", whose stated fix is to DELETE it. The same
-- omission on a StrictConfig spec took the fleet's page-publishing path down for
-- 33 minutes on 2026-08-19 (WRONG_CALLS, migration 494). The declaration is Go,
-- so it is inert until the next roll: between now and then the config report
-- names this key as unrecognised. That is cosmetic and it is the stated cost of
-- arming detection today rather than after a roll.
--
-- ✅ RESOLVED 2026-08-20 (same day): bb6600e48 shipped on v1.0.1319 (revision
-- 447f3a8a8, merge-base with controls both ways). The interim window lasted one
-- build. The key is now declared on the action that reads it, so the config
-- report no longer names a live working setting as unrecognised.
--
-- WHAT IT COSTS. One work item per (page, slot, dead-field-set) on any re-render
-- of a section that renders an empty src=/href=. `insertWorkItem`'s dedup
-- (idx_swi_dedup-matched ON CONFLICT) and two-strike anti-churn label bound the
-- volume. Items are born `needs_human_review` with NO handler — deliberately:
-- nothing automated can invent a missing image or destination. No disposition
-- changes; nothing is refused; no page is blocked.
--
-- ⚠ ONE CHURN EDGE, stated because the close-out has to look for it: the item key
-- embeds the SORTED dead-field list, so a PARTIAL repair changes the key and can
-- mint a second item while the first is still parked. Bounded, not zero.
--
-- GATED FIELDS ARE NOT COVERED BY THIS, and that is the honest limit of the whole
-- detection half: `missingBareFields` walks root-scope actions only, so a
-- `{{if .card1_link_url}}` guard keeps its field out of the report. For the gated
-- class the only signal is plan_sections' `STRUCTURAL_KEY_CARRY_MISS` finding
-- (28 rows in agent_error_log, 2026-08-11 -> 08-17, no consumer until 498/499).
-- A gated field fails MORE quietly than an ungated one.
--
-- DISJOINT FROM bugs_open/312 AND ITS HELD 477. 477 edits page-content-writer's
-- `select_sections` field paths; this file edits page-rerender's
-- `rerender_sections` step config. Different agent, different step, different key.
-- Nothing here requires 477, and applying 477 later neither needs nor undoes it.
--
-- NOT A NEW FLOOR IN save_page_sections (bugs_open/178's standing stop sign):
-- this arms an existing render-path emit and adds no guard to that function.
--
-- ⚠ THIS FILE WAS 497 FOR ABOUT AN HOUR, AND THE COUNCIL SUBMISSION SAYS 497.
-- Council correlation 8a2aab7c-2ffa-469d-bb55-ce5a11126613 was dispatched while
-- the file was still numbered 497, so its payload names `497_bugfix_238_arm_dead_url_
-- record_on_rerender.sql`. Same file, same content, renumbered to 504 — nothing
-- about the change moved. Recorded here rather than left to be noticed because a
-- verdict that names a path which no longer exists reads as a stale review.
--
-- WHY it moved: I read the next free number (496 highest, so 497), then spent an
-- hour writing, and by the time the file landed two OTHER lanes had taken 497 AND
-- 498 — and 499-503 besides. The lesson is not "pick a bigger number", it is that
-- on this tree a migration number read more than a few minutes ago is stale: take
-- it immediately before you name the file, and re-check before you commit.
-- Logged in WRONG_CALLS.md 2026-08-20. Note 497 and 498 each still carry TWO
-- lanes' files today (not mine to fix, but do not trust `ls | tail -1`).
--
-- Rollback: 504_bugfix_238_arm_dead_url_record_on_rerender_ROLLBACK.sql.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('page-rerender', '504_bugfix_238_arm_dead_url_record: pre-update');

\echo '=== BEFORE: the rerender_sections step config ==='
SELECT jsonb_pretty(default_config #> '{workflow,steps,rerender_sections,config}')
  FROM agent_definitions
 WHERE type = 'page-rerender' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ⚠ DUPLICATE-ROW GUARD. Four agent types on this estate carry TWO active
-- definition rows and only the HIGHER VERSION is loaded at runtime, so an
-- `UPDATE ... WHERE type = '<x>'` can silently touch a stale duplicate, or both,
-- while a verify block that re-reads "the row for this type" confirms whichever
-- one it happens to find. Measured for page-rerender 2026-08-20: exactly 1 active
-- row. This guard exists so that if a second appears, the file refuses instead of
-- half-applying.
DO $$
DECLARE
    v_rows  int;
    v_armed boolean;
BEGIN
    SELECT count(*) INTO v_rows
      FROM agent_definitions
     WHERE type = 'page-rerender' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '238/504: expected exactly 1 live page-rerender row, found % — this type now carries duplicates and only the HIGHEST VERSION is loaded at runtime; target that row by id, do not update by type', v_rows;
    END IF;

    SELECT (default_config #>> '{workflow,steps,rerender_sections,config,record_dead_url_controls}')::boolean
      INTO v_armed
      FROM agent_definitions
     WHERE type = 'page-rerender' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_armed IS TRUE THEN
        RAISE EXCEPTION '238/504: already applied — record_dead_url_controls is already true';
    END IF;
END $$;

-- ⚠ PRE-FLIGHT: assert the path EXISTS and holds this action, BEFORE any
-- jsonb_set runs. `jsonb_set(..., create_missing := true)` on a wrong path
-- inserts a whole new branch and reports success — arming nothing while every
-- downstream reader says "armed". The council's editquality seat gated 380's
-- round 2 on exactly this.
DO $$
DECLARE
    v_action text;
BEGIN
    SELECT default_config #>> '{workflow,steps,rerender_sections,action}'
      INTO v_action
      FROM agent_definitions
     WHERE type = 'page-rerender' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_action IS DISTINCT FROM 'rerender_page_sections' THEN
        RAISE EXCEPTION '238/504: rerender_sections is % at the expected path (want rerender_page_sections) — the workflow shape moved; re-derive the path before arming', COALESCE(v_action, '(absent)');
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           -- Top level, NOT inside a sub_workflow. See the header: 380's nested
           -- path belongs to page-content-writer and finds nothing here.
           '{workflow,steps,rerender_sections,config,record_dead_url_controls}',
           'true'::jsonb,
           true),
       updated_at = now()
 WHERE type = 'page-rerender' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

\echo '=== AFTER ==='
SELECT jsonb_pretty(default_config #> '{workflow,steps,rerender_sections,config}')
  FROM agent_definitions
 WHERE type = 'page-rerender' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- VERIFY BY COUNTING, FLEET-WIDE, not by reading back the path just written.
--
-- Three disciplines, all borrowed from 380 because they were all learned the hard
-- way there. (1) A jsonb comparison with `=`/`<>` sits GREEN for ever when the key
-- is ABSENT (NULL <> 'true' is NULL, not TRUE), so every comparison here is
-- `IS DISTINCT FROM` / `IS NOT TRUE`, which treat absence as failure. (2) Reading
-- back the one path you wrote proves only that you wrote it; it cannot see a step
-- you never armed — so this counts ALL steps running this action and demands they
-- ALL carry the flag. (3) The traversal is asserted non-empty first: a zero count
-- from a wrong path is indistinguishable from a clean pass otherwise.
--
-- Scoped to every live definition, NOT to page-rerender, deliberately: the point
-- is to catch a SECOND agent acquiring this step later.
DO $$
DECLARE
    v_steps int;
    v_armed int;
BEGIN
    SELECT count(*) FILTER (WHERE k.value->>'action' = 'rerender_page_sections'),
           count(*) FILTER (WHERE k.value->>'action' = 'rerender_page_sections'
                              AND (k.value->'config'->>'record_dead_url_controls')::boolean IS TRUE)
      INTO v_steps, v_armed
      FROM agent_definitions ad,
           LATERAL jsonb_path_query(ad.default_config, 'strict $.**.steps') AS steps,
           LATERAL jsonb_each(steps) AS k
     WHERE ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL;

    IF v_steps = 0 THEN
        RAISE EXCEPTION '238/504: found ZERO rerender_page_sections steps fleet-wide — the traversal is wrong, so a green result here would mean nothing; aborting';
    END IF;
    IF v_armed <> v_steps THEN
        RAISE EXCEPTION '238/504: % of % rerender_page_sections step(s) armed — a step exists that this file does not know about; add it rather than shipping partial coverage the report will call complete', v_armed, v_steps;
    END IF;
    RAISE NOTICE '238/504: dead-URL RECORD-ONLY emit ARMED on ALL % rerender_page_sections step(s) fleet-wide', v_steps;
END $$;

-- NEGATIVE CONTROL, in the same transaction as the positive one.
--
-- This file must NOT arm the refusal. If a future edit reaches for the wrong key
-- or the wrong agent, the two halves stop being separately decidable and 380's
-- owner-facing choice is made silently by this file instead. So assert the
-- refusal is still unset everywhere: the check that could come out otherwise.
DO $$
DECLARE
    v_refusals int;
BEGIN
    SELECT count(*)
      INTO v_refusals
      FROM agent_definitions ad,
           LATERAL jsonb_path_query(ad.default_config, 'strict $.**.steps') AS steps,
           LATERAL jsonb_each(steps) AS k
     WHERE ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
       AND (k.value->'config'->>'refuse_dead_url_controls')::boolean IS TRUE;

    IF v_refusals <> 0 THEN
        RAISE EXCEPTION '238/504: % step(s) carry refuse_dead_url_controls=true — this file arms the RECORD half only, and 380 (the refusal) is gated on the damaged set being drained first; something else armed it', v_refusals;
    END IF;
    RAISE NOTICE '238/504: negative control OK — the refusal remains unarmed (380 stays the deliberate, separate decision)';
END $$;

COMMIT;

-- ---------------------------------------------------------------------------
-- WATCH after arming.
--
--   SELECT count(*), max(created_at) FROM site_work_items
--    WHERE item_type = 'dead_url_control';
--
-- Baseline is ZERO fleet-wide, all history. The first row is therefore also this
-- class's first real frequency measurement.
--
-- ⚠ A SUSTAINED ZERO HAS TWO READINGS AND THEY MUST BE DISCRIMINATED: either
-- nothing is re-rendering a damaged section (entirely plausible — the improvement
-- loop and the discovery rotations are both paused on cost grounds,
-- bugs_open/230), or the flag did not take. This file's own verify block settles
-- the second, so do not read silence as evidence about the first without it.
--
-- THE DEMAND CONTROL, if the count is still zero after 498/499's conversions run:
-- dispatch one `page_rerender` at ai-agent-orchestration.com /index.html in 379's
-- shape (`spec.reason = 'section_data_resolved'`). That page's `case-studies-grid`
-- has five card*_image_url fields BARE inside src= and absent from content_data —
-- it serves five `src=""` today, verified at the artefact 2026-08-19 — so
-- `missingBareFields` returns non-empty and this emit MUST fire. A traffic-free
-- control cannot tell "armed and quiet" from "armed and blind"
-- (docs024_key_docs_latest: a post-fix zero needs a demand control).
--
-- Its cost, stated: the re-render re-ships damage that is already live. It cannot
-- worsen the page — the re-render path merges stored ⊕ fresh and structurally
-- cannot lose a key — and it writes one more STRUCTURAL_KEY_CARRY_MISS row, which
-- is more signal, not less.
