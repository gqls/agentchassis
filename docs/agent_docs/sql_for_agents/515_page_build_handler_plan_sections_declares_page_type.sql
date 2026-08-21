-- 515 — page-build-handler's `plan_sections` step names WHERE `page_type` comes
--       from, with the `?` OPTIONAL-EXPLICIT marker: that path or nothing, never
--       the whole-tree search. RFC_029 §10.13 step 5 precondition work.
--       CONFIG ONLY — live on apply. NOT a _HOLD: see ORDERING below.
--
-- ============================================================================
-- WHAT THE PROBLEM IS, in plain terms
-- ============================================================================
-- `plan_sections` declares `page_type` as an Optional input
-- (plan_sections_action.go:54). page-build-handler's `plan_sections` step wires
-- nothing for it. So the resolver falls through to its last resort — collect
-- every key called `page_type` anywhere in the collected data and pick a winner.
--
-- What it finds, per the instrument (40 rows all-time, last 2026-08-20 15:11:38Z):
--   * `load_page_record.page_type` and `page_record.page_type` — the page's OWN
--     record. The right answer.
--   * 28 × `{ensure_site_record,site_record}.content_data.pages[N].page_type` —
--     the page_type of every OTHER page on the site.
--
-- Today's winner is `load_page_record.page_type`, i.e. correct — but only because
-- it sorts shallowest-first. Nothing declares that preference for this field.
--
-- ============================================================================
-- WHAT IT COSTS TODAY: A REAL WRONG-VALUE RISK, on the majority of runs
-- ============================================================================
-- [MEASURED 2026-08-21 ~10:3xZ, 31 live page-build-handler orchestrations]
--   page_record.page_type PRESENT : 13/31
--   page_record.page_type ABSENT  : 18/31
-- On the 13, the search has the right candidate available and picks it. On the
-- other 18 the page's own record is not in the tree at all, so the ONLY
-- candidates are the sibling-page array entries — and the search hands
-- `plan_sections` some OTHER page's page_type. If those siblings happen to agree
-- with each other, no conflict row is written and the substitution is SILENT
-- (the instrument's known blind spot: a row needs the candidates to differ).
--
-- So this is not merely step-5 hygiene. It is a live wrong-value path today, on
-- the majority of runs, and the fix improves current behaviour.
--
-- ============================================================================
-- AND UNDER STEP 5 IT WOULD GET WORSE IN THE OTHER DIRECTION
-- ============================================================================
-- RFC_029 §9 D2's flip makes CONFLICTING candidates resolve to nothing. On the
-- 13 good runs the candidates DO conflict (own record vs siblings), so the flip
-- would discard the correct value it is currently finding. This pair therefore
-- needs a declared mapping either way — it is one of the pairs the flip's own
-- precondition names ("every observed field/caller pair given an explicit
-- mapping first").
--
-- ============================================================================
-- WHY `?` AND NOT A PLAIN WIRE — the 18 misses are the whole point
-- ============================================================================
-- A plain `"page_type": "page_record.page_type"` resolves via Strategy 0 when
-- the path exists (13/31) and falls through to the whole-tree search when it
-- does not (18/31) — which is EXACTLY the case we most need to stop, because
-- that is where the only candidates are other pages. A plain wire would fix the
-- minority and leave the majority guessing.
--
-- The `?` marker (ecc419bd1, RFC_029 §10.15, CTS-060(5)) means: resolve this
-- field from the named path or leave it ABSENT — never the whole-tree search,
-- never the nested-object fallback, never the deprecated bridge. Unlike `!`, a
-- miss is not an error: an Optional field simply arrives absent.
--
-- ============================================================================
-- WHY ABSENCE IS SAFE HERE — read at the consumer, not assumed
-- ============================================================================
-- plan_sections_action.go:972-975:
--     pageType := inputs.Get("page_type")
--     if pageType == "" {
--         pageType = pageName // fall back to page name as page type
--     }
-- and the comment above it (:968-970): "site_type and page_type for the
-- component selector fallback path. If not provided (existing workflows), the
-- selector still works — it just scores without site/page type relevance
-- bonuses."
--
-- So on the 18 runs the value becomes the page's own NAME instead of another
-- page's type. That is strictly better than today: a locally-correct fallback
-- replacing a foreign value.
--
-- ============================================================================
-- WHY `page_record` AND NOT `load_page_record`
-- ============================================================================
-- Both keys exist because the coordinator stores a step result under the step
-- name AND under its declared `output_field`. `load_page_record`'s declared
-- output_field is `page_record`, so that is the CONTRACT; the step-name key is
-- an incidental alias, and it is only today's winner because "l" sorts before
-- "p". [MEASURED] the two agree on 8/8 of the most recent runs carrying them and
-- are co-present 13/13. Naming the declared contract is the stable choice.
--
-- Step order is satisfied: ensure_site_record -> load_page_record ->
-- check_page_found -> ... -> load_spec_sections -> plan_sections. The producer
-- runs first.
--
-- ============================================================================
-- ORDERING — why this is NOT a _HOLD file
-- ============================================================================
-- The `?` marker on the ExtractActionInputs surface only parses in a binary
-- carrying ecc419bd1. Applying this against an older binary would leave a
-- literal key `page_type?` that nothing reads — inert, but a silent no-op.
--
-- [VERIFIED 2026-08-21 ~10:2xZ] The parsing binary IS LIVE. `v1.0.1321` (both
-- chassis pods, up 2026-08-20T19:51Z) was built from `0483e7f4e`, and
-- `git merge-base --is-ancestor ecc419bd1 0483e7f4e` returns true.
--
-- How the stamp was established, because the usual routes both failed and the
-- next person will need this: the `build provenance` log line had scrolled (14 h
-- since pod start), and ecc419bd1 adds NO probeable symbol — no new string
-- literal in code (its two quotable phrases are inside COMMENTS, which Go
-- strips) and no new named function (all closures and locals). So the capability
-- probe this estate normally relies on has no target for this change. What
-- worked was a TARGETED fixed-string grep of the build stamp against candidate
-- commits, with a control:
--     kubectl -n ai-persona-system exec <pod> -- grep -aqF "<candidate-sha>" /proc/1/exe
--     kubectl -n ai-persona-system exec <pod> -- grep -aqF "deadbeef…deadbeef" /proc/1/exe   # MUST be absent
-- One sha per call. A 60-way alternation (`grep -aoE "a|b|c…"`) TIMES OUT at 2
-- minutes against the binary — do not try to batch it.
--
-- ROLLBACK: 515_page_build_handler_plan_sections_declares_page_type_ROLLBACK.sql
-- ============================================================================

BEGIN;

-- Snapshot the row before touching it, so the rollback has a source of truth
-- that does not depend on this file being read correctly. Uses the estate's own
-- helper rather than a hand-rolled INSERT: there is NO `snapshot_reason` column
-- (checked \d agent_definitions — the reason goes in `description`), and
-- snapshot_agent() is what every other migration uses.
SELECT snapshot_agent('page-build-handler',
                      '515_page_build_handler_plan_sections_declares_page_type: pre-update');

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','plan_sections','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '515: no live page-build-handler plan_sections step config — refusing to guess';
    END IF;

    -- The action must still be the one whose spec was read. If the step has been
    -- repointed, every measurement above is about a different action.
    IF (SELECT default_config #>> ARRAY['workflow','steps','plan_sections','action']
          FROM agent_definitions
         WHERE type = 'page-build-handler' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL) <> 'plan_sections' THEN
        RAISE EXCEPTION '515: page-build-handler plan_sections step no longer runs the plan_sections action — re-measure before applying';
    END IF;

    -- Idempotence, and refuse to fight a competing wire.
    IF cfg ? 'page_type?' THEN
        RAISE EXCEPTION '515: page_type? is already declared — already applied';
    END IF;
    IF cfg ? 'page_type' THEN
        RAISE EXCEPTION '515: an UNMARKED page_type wire already exists (%). Someone chose a plain wire deliberately; read their reason before converting it', cfg ->> 'page_type';
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(
               default_config,
               ARRAY['workflow','steps','plan_sections','config','page_type?'],
               '"page_record.page_type"'::jsonb,
               true),
           updated_at = NOW()
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
END $$;

-- ============================================================================
-- VERIFY — a DO block, not a SELECT. ON_ERROR_STOP does not abort a COMMIT on a
-- non-empty result set, so a verify block made of SELECTs cannot actually stop
-- a bad apply (LANDMINES / RFC_006).
-- ============================================================================
DO $$
DECLARE
    got text;
BEGIN
    SELECT default_config #>> ARRAY['workflow','steps','plan_sections','config','page_type?']
      INTO got
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF got IS DISTINCT FROM 'page_record.page_type' THEN
        RAISE EXCEPTION '515 VERIFY FAILED: page_type? reads % , expected page_record.page_type', COALESCE(got, '<null>');
    END IF;

    -- Snapshots land in agent_definitions_backup, NOT in agent_definitions:
    -- snapshot_agent() inserts there, and that is the table carrying
    -- snapshot_taken_at / snapshot_reason. (agent_definitions.is_snapshot is an
    -- older, separate convention — two rows use it. Checked both.)
    IF (SELECT count(*) FROM agent_definitions_backup
         WHERE type = 'page-build-handler'
           AND snapshot_reason LIKE '515_page_build_handler_plan_sections%') < 1 THEN
        RAISE EXCEPTION '515 VERIFY FAILED: no snapshot row in agent_definitions_backup';
    END IF;

    RAISE NOTICE '515 OK: plan_sections declares page_type? = page_record.page_type';
END $$;

COMMIT;
