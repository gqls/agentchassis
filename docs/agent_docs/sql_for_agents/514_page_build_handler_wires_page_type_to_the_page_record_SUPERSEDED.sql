-- ═══════════════════════════════════════════════════════════════════════════
-- ⛔ SUPERSEDED 2026-08-21 — DO NOT RUN. Migration 515
--    (515_page_build_handler_plan_sections_declares_page_type.sql) is what
--    actually ships for this pair.
--
-- Two sessions of the staged_component_build lane worked pbh/`page_type`'s
-- disposition in parallel without seeing each other. Both aimed at the exact
-- same jsonb path (`{workflow,steps,plan_sections,config}` on
-- page-build-handler) and both were council-submitted before either applied —
-- this file under `Council-Submitted: 81a4fe27-8cbf-458f-8a55-c55698fbd6e3`,
-- 515 under `a452fc2a-160f-485c-949c-367c34c65df2`. Neither has applied to the
-- live row as of this write (checked: `schema_migrations` has no row for
-- either number, and the live `plan_sections.config` carries no `page_type`
-- key of any spelling).
--
-- ⚠ AND THIS FILE'S FIX IS GENUINELY THE WEAKER OF THE TWO, not merely
-- redundant — that is the reason to retire it rather than flip a coin. This
-- file wires a PLAIN key (`"page_type": "page_record.page_type"`), which
-- resolves via Strategy 0 only when `page_record.page_type` is present in the
-- tree and falls through to the whole-tree search when it is not. **515's own
-- measurement, which this file never ran, is that `page_record.page_type` is
-- ABSENT on 18 of 31 recent orchestrations** — so a plain wire fixes the
-- 13/31 minority and leaves the 18/31 majority exactly as exposed as they are
-- today, guessing a SIBLING page's type from the site's page list whenever
-- those siblings happen to agree (the instrument's own blind spot: agreeing
-- candidates write no conflict row, so the substitution is silent on top of
-- being wrong). 515 uses the `?` OPTIONAL-EXPLICIT marker instead — resolve
-- from the named path or be ABSENT, never the search — which closes the gap
-- in BOTH directions rather than one.
--
-- The tell, in hindsight: this file's own measurement asked "when both
-- `load_page_record.page_type` and `page_record.page_type` are present, do
-- they agree" (13/13, yes) and never asked "how often is the chosen path
-- present at all" — agreement and coverage are different questions, and only
-- the second one tells you what a plain wire actually protects. Logged in
-- WRONG_CALLS.
--
-- Renamed with an uppercase suffix so the migration runner's SIDECAR_RE
-- (`_[A-Z][A-Z0-9_]*\.sql$`, run-migrations.sh:65) excludes it from --apply
-- while still listing it, and so `council-scope.sh`'s own `_SUPERSEDED`
-- exclusion keeps it out of scope for any future review. It is NOT in
-- `schema_migrations`. If it were ever run by hand, its own guard would
-- refuse cleanly: it checks for an UNMARKED `page_type` key, which 515 never
-- writes (515 writes `page_type?`), so the two files' guards do not even see
-- each other — a real collision was possible had both applied, which is why
-- this is retired now rather than left to be resolved by which one landed
-- first.
--
-- ADDENDUM: the council round finished after retirement (2026-08-21 10:30:25Z,
-- REVISE, editquality gating). Worth recording since it is a second, distinct
-- lesson from this file's real, honest guard blocks below: **the objections
-- attack the submission's `sketch` field, which carried only the bare
-- `jsonb_set(...)` one-liner, not this file's actual guard/VERIFY blocks that
-- already answer them** (restructure check, already-wired check, sibling-
-- convention check, all present below). A reviewer sees the sketch, not the
-- file — an abbreviated sketch reads as an unguarded migration even when the
-- real one is not. Put the guard in the sketch next time, not a summary of it.
-- ═══════════════════════════════════════════════════════════════════════════

-- 514 — page-build-handler's `plan_sections` step WIRES `page_type` to the
--       page's own record, so the whole-tree search stops guessing it.
--       RFC_029 §10.13 step 5 precondition work. CONFIG ONLY — live on apply.
--
-- ============================================================================
-- WHAT THE PROBLEM IS, in plain terms
-- ============================================================================
-- `plan_sections` declares `page_type` as Optional (plan_sections_action.go:54)
-- and uses it to score the component selector's "site/page type relevance
-- bonuses" (plan_sections_action.go:968-973). page-build-handler's
-- `plan_sections` step does not wire it. So the whole-tree search runs, finds
-- every `page_type` key across the tree — the page's own record
-- (`load_page_record.page_type`), its alias (`page_record.page_type`), and 28
-- more inside `{ensure_site_record,site_record}.content_data.pages[N].page_type`
-- (one per page on the site) — and hands over the shallowest: always
-- `load_page_record.page_type`, in all 40 conflict rows logged since 08-16.
--
-- UNLIKE `bugs_open/330` or migration 512's `reason` case, the guessed value
-- HERE IS RIGHT, not a foreign substitute: `load_page_record` is the step that
-- loads THIS page's own record, and `page_record` (the step's own output_field
-- alias) carries the identical value by construction — [MEASURED 2026-08-21]
-- 13/13 orchestrations with either populated show `load_page_record.page_type`
-- and `page_record.page_type` agreeing exactly (they are two keys pointing at
-- one written value, not independent sources that happen to agree). So this is
-- NOT a case for declaring absence (contrast 512): the field should resolve,
-- and resolving it explicitly is what stops the guess without losing the value.
--
-- ============================================================================
-- WHY LEAVING IT TO THE SEARCH IS DANGEROUS, NOT MERELY UNTIDY
-- ============================================================================
-- Unlike `reason` in 512 (inert unless it equals one of three magic strings),
-- an ABSENT `page_type` here is NOT harmless. plan_sections_action.go:972-973:
--
--     pageType := inputs.Get("page_type")
--     if pageType == "" {
--         pageType = pageName // fall back to page name as page type
--     }
--
-- If RFC_029 step 5 ships before this pair is dispositioned, the conflict
-- refuses, `page_type` resolves to NOTHING, and the fallback substitutes the
-- PAGE'S NAME (e.g. "blog-post-my-article-slug") into a field the selector
-- scores as if it were a page TYPE (e.g. "blog-post"). That is a silent
-- degradation of the component selector's relevance bonuses, not a no-op — the
-- opposite disposition from 512's `reason`, and the reason both are in this
-- migration's header rather than assumed from the shape.
--
-- ============================================================================
-- THE RULE
-- ============================================================================
-- Same RFC_029 §9 D2 precondition as migration 512: "zero conflict WARNs
-- observed over the window, OR every observed field/caller pair given an
-- explicit mapping first." page-build-handler/`page_type` is live and
-- reawakened — 40 rows since 2026-08-16, 3 in the 24h to 2026-08-20 evening
-- (15:03/15:07/15:11Z), after reading quiet for two days. Declaring the wire
-- is this pair's disposition.
--
-- ============================================================================
-- THE FIX, AND WHY THIS SPELLING
-- ============================================================================
-- The live plan_sections step already wires its sibling field the SAME way:
-- `"page_name": "page_record.name"`. `page_type` gets the identical
-- convention, `"page_type": "page_record.page_type"` — not
-- `load_page_record.page_type`, which resolves to the same value (see the
-- agreement measurement above) but would introduce a second naming style onto
-- one step where every other wire uses the `page_record.*` alias.
--
-- With this key present, Strategy 0 resolves `page_type` directly
-- (action_inputs.go, dotted-path arm) and the LIVE step-1 prune (v1.0.1310)
-- removes it from what Strategy 1/2 request — so the whole-tree search is
-- never reached for this field on this step. No other reader of `page_type`
-- on this step exists to be starved; `site_type` is untouched by this file.
--
-- ROLLBACK: 514_page_build_handler_wires_page_type_to_the_page_record_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('page-build-handler', '514_page_build_handler_wires_page_type_to_the_page_record: pre-update');

-- GUARD: refuse unless the live row is the one this file was written against.
DO $$
DECLARE
    n    int;
    step jsonb;
    cfg  jsonb;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '514: expected exactly 1 live page-build-handler row, found %', n;
    END IF;

    SELECT default_config #> ARRAY['workflow','steps','plan_sections'] INTO step
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF step IS NULL THEN
        RAISE EXCEPTION '514: page-build-handler has no plan_sections step — the workflow has been restructured since 2026-08-21; re-derive this migration';
    END IF;
    IF step->>'action' <> 'plan_sections' THEN
        RAISE EXCEPTION '514: plan_sections runs %, not plan_sections — page_type would be read by a different spec', step->>'action';
    END IF;

    cfg := step->'config';

    IF cfg ? 'page_type' THEN
        RAISE EXCEPTION '514: plan_sections ALREADY wires page_type (%) — another session has applied this or an equivalent; do not overwrite it', cfg->'page_type';
    END IF;

    -- The premise is that the sibling field is wired to page_record.* — if that
    -- convention has changed, matching it blindly would introduce a mismatch,
    -- not fix one.
    IF cfg->>'page_name' IS DISTINCT FROM 'page_record.name' THEN
        RAISE EXCEPTION '514: plan_sections.page_name is % (want page_record.name) — the sibling convention has changed; re-derive the spelling for page_type', cfg->'page_name';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
           '{workflow,steps,plan_sections,config,page_type}',
           '"page_record.page_type"'::jsonb, true),
       updated_at = NOW()
 WHERE type = 'page-build-handler'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

-- VERIFY — a DO block that RAISEs, not a SELECT: ON_ERROR_STOP does not stop a
-- COMMIT on a non-empty result set (LANDMINES / RFC_006).
DO $$
DECLARE
    cfg    jsonb;
    leaked text;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','plan_sections','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg->>'page_type' IS DISTINCT FROM 'page_record.page_type' THEN
        RAISE EXCEPTION '514 VERIFY: plan_sections.page_type is %, want page_record.page_type', cfg->'page_type';
    END IF;

    -- jsonb_set is surgical; assert the step's five pre-existing keys survived.
    IF cfg->>'site_id'         IS DISTINCT FROM 'site_record.site_id'          THEN RAISE EXCEPTION '514 VERIFY: site_id did not survive: %', cfg::text; END IF;
    IF cfg->>'sections'        IS DISTINCT FROM 'spec_sections.sections'       THEN RAISE EXCEPTION '514 VERIFY: sections did not survive: %', cfg::text; END IF;
    IF cfg->>'page_name'       IS DISTINCT FROM 'page_record.name'             THEN RAISE EXCEPTION '514 VERIFY: page_name did not survive: %', cfg::text; END IF;
    IF cfg->>'error_step'      IS DISTINCT FROM 'mark_item_failed'             THEN RAISE EXCEPTION '514 VERIFY: error_step did not survive: %', cfg::text; END IF;
    IF cfg->>'work_item_id'    IS DISTINCT FROM 'input_data.work_item_id'      THEN RAISE EXCEPTION '514 VERIFY: work_item_id did not survive: %', cfg::text; END IF;
    IF cfg->>'section_facts'   IS DISTINCT FROM 'spec_sections.section_facts'  THEN RAISE EXCEPTION '514 VERIFY: section_facts did not survive: %', cfg::text; END IF;

    -- NEGATIVE CONTROL in the same transaction, RECURSIVE (sub_workflow-safe —
    -- this lane's own twice-bitten census trap, see migration 512): the other
    -- live caller of plan_sections (page-content-writer) must be untouched. It
    -- has no load_page_record/page_record step at all, so wiring page_type
    -- there would name a path that resolves to nothing — a different fix, not
    -- this one, and out of scope for this file.
    WITH RECURSIVE steps(type, path, step) AS (
        SELECT ad.type, s.key, s.value
          FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
         WHERE ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
        UNION ALL
        SELECT p.type, p.path || '.' || s.key, s.value
          FROM steps p, LATERAL jsonb_each(p.step->'config'->'sub_workflow'->'steps') s
    )
    SELECT string_agg(type || '.' || path, ', ') INTO leaked
      FROM steps
     WHERE step->>'action' = 'plan_sections'
       AND step->'config' ? 'page_type'
       AND type <> 'page-build-handler';
    IF leaked IS NOT NULL THEN
        RAISE EXCEPTION '514 VERIFY: the wire leaked to steps it was not meant for: %', leaked;
    END IF;

    RAISE NOTICE '514 OK: page-build-handler.plan_sections wires page_type -> page_record.page_type; five pre-existing keys intact; page-content-writer untouched';
END $$;

COMMIT;
