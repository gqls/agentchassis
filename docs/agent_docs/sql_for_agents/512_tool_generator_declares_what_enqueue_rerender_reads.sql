-- 512 — tool-generator's `enqueue_rerender` step DECLARES what it reads, so the
--       whole-tree search stops guessing a `reason` for it.
--       RFC_029 §10.13 step 5 precondition work. CONFIG ONLY — live on apply.
--
-- ============================================================================
-- WHAT THE PROBLEM IS, in plain terms
-- ============================================================================
-- When a step's action declares an input field and nothing in the step config
-- says where that field comes from, the resolver's last resort is a whole-tree
-- search: it collects every key of that name anywhere in the collected data and
-- picks a winner. `create_rerender_items` declares `reason` as Optional
-- (create_rerender_items_action.go:36-47). tool-generator's `enqueue_rerender`
-- step does not wire it. So the search runs, finds 42 keys called `reason`
-- scattered through the site spec, and hands the step the shallowest one:
--
--     load_brand_context.specs.classification.content_features.news_feed.reason
--
-- which is the site's news-feed classification note. [MEASURED 2026-08-20 17:2xZ]
-- Its live value is prose — "Phase 2 news section. Retuned 2026-07-27 on the
-- owner directive to aim at the industry and ..." — on all four most recent runs.
--
-- ============================================================================
-- WHAT IT COSTS TODAY: NOTHING, AND THAT IS THE POINT
-- ============================================================================
-- `reason` only changes behaviour in this action when it equals one of three
-- values (create_rerender_items_action.go:216-231): `section_data_resolved`,
-- `image_landed` (both also needing component_id) or `cta_links_stale`.
-- Anything else leaves stampReason false and keyReason empty — identical to
-- carrying no reason at all, which is what a newly created tool page wants
-- (assemble-only, the tool-birth deploy gap this step exists to close).
--
-- So the substituted prose is inert BY LUCK OF ITS VALUE, not by design. Nothing
-- stops a future spec carrying a `reason` of `cta_links_stale` somewhere in that
-- tree, at which point a tool-birth rerender silently changes render mode. This
-- file converts accidental inertness into a declared absence.
--
-- ============================================================================
-- THE RULE
-- ============================================================================
-- RFC_029 §9 D2's flip (conflicting whole-tree candidates resolve to NOTHING)
-- is gated on its own stated precondition, quoted from findFieldRecursive:
-- "zero conflict WARNs observed over the window, OR every observed field/caller
-- pair given an explicit mapping first". `tool-generator`/`reason` is one of the
-- pairs still logging (30 rows all-time, 16 in the 24 h to 2026-08-20 17:20Z,
-- last 17:08Z, against 16 tool-generator runs in the same window). Declaring the
-- step's read list is this pair's disposition.
--
-- ============================================================================
-- HOW THIS CASE MEASURES AGAINST IT — the blast radius, arm by arm
-- ============================================================================
-- The action reads exactly four fields through `inputs`: site_id, domain, reason,
-- component_id (grep `inputs.Get` in create_rerender_items_action.go); pages_field
-- and the three *_field keys are read from `config` DIRECTLY, never through the
-- resolver, so they do not need to be extracted at all.
--
-- With input_fields = ["site_id","domain"]:
--   Strategy 0  unchanged — still resolves domain, site_id and the three *_field
--               config dot-paths exactly as today.
--   Strategy 1  requests withoutResolved(["site_id","domain"]) = EMPTY, so
--               `reason` and `component_id` are never searched. THIS IS THE
--               WHOLE CHANGE. An empty pruned list is an anticipated case, not
--               an edge one (action_inputs.go, the step-1 prune's own comment).
--   Strategy 3  spec.Deprecated is empty for this action — nothing.
--   nested-obj  its parents (current_page/rerender_pages/site_record/input_data)
--               are not spec fields here, so result.Values never holds them.
--   Strategy 4  no config["reason"], no config["component_id"] — nothing.
--   Strategy 5  non-string scalars only — nothing.
--   Strategy 6  the pages_field spec Default stands, as today.
--
-- If Strategy 0 ever FAILS to resolve site_id or domain, both stay in the pruned
-- list and the search still rescues them — this file removes no robustness for a
-- field the action actually consumes.
--
-- `component_id` also stops resolving, and that is inert for the same measured
-- reason: it is read only inside `if scoped` (line 235), and `scoped` requires a
-- magic `reason` this step will now never have. It has no other reader.
--
-- `input_fields` is a FRAMEWORK step-config key (frameworkStepConfigKeys,
-- action_inputs.go:212-230), so adding it draws no unknown-key report even though
-- CreateRerenderItemsInputSpec sets CheckConfig: true.
--
-- Precedent: migration 483 declared input_fields on html-developer-chunked's three
-- generate_html steps for step 3 of the same RFC, in this shape.
--
-- EXPECTED OBSERVABLE EFFECT: none on rendered output. The
-- RESOLVER_CONFLICTING_CANDIDATES rows for tool-generator/`reason` stop; the
-- tool-generator/`related_pages` class MUST keep firing (it is a different
-- defect, bugs_open/330, and it is also the proof the instrument is still alive);
-- and no `component_id` class may appear.
--
-- ROLLBACK: 512_tool_generator_declares_what_enqueue_rerender_reads_ROLLBACK.sql
--
-- ============================================================================
-- COUNCIL: APPROVED round 1, corr 2bd7fb37-cac2-409a-8452-50a7ed933467
-- (2026-08-20 17:33:23Z), 3 advisory objections, none high-severity. All three
-- dispositioned here rather than in a doc nobody opens next to this file:
--
--  1. THE MULTI-ACTIVE-VERSION LANDMINE (editquality medium, debug_historian
--     medium, guardian low): "UPDATE agent_definitions WHERE type=<x>" — four
--     types carry TWO active rows and only the higher `version` loads, so the
--     write can land on the stale one. MEASURED 2026-08-20 17:4xZ: tool-generator
--     has exactly ONE live row (version 1, id 1bca62f6-…), so it is not one of
--     the four. The guard above already asserts `count(*) = 1` and RAISEs
--     otherwise — deliberately, rather than the `ORDER BY version DESC LIMIT 1`
--     the seats suggested: on a migration, refusing and making a human look is
--     the better failure than silently choosing a row. Nothing changed in
--     response; the check the seats asked for is the one that was already here.
--  2. TOP-LEVEL-ONLY NEGATIVE CONTROL (guardian low): UPHELD and FIXED — the
--     in-transaction control is now RECURSIVE (see the VERIFY block). The claim
--     it supports was already true (recursive count == top-level count == 3
--     steps), but it was true unverifiably.
--  3. UNVERIFIED PRECEDENT (prior_art_librarian medium): 483's existence and
--     shape, and findFieldRecursive's precondition comment, were quoted from a
--     first-hand read, not from another document —
--     docs/agent_docs/sql_for_agents/483_html_developer_chunked_declares_current_page.sql
--     is 3 jsonb_set calls adding input_fields to three generate_html steps with
--     guard + DO/RAISE VERIFY + negative control, i.e. the shape claimed; and
--     `input_fields` long predates this file (frameworkStepConfigKeys,
--     action_inputs.go:212-230, commented "read by ExtractActionInputs for every
--     action"). The seat's remit is absence-claims; this is neither an absence
--     claim nor new machinery.
--
--  bug_historian's low objection — "this closes one of four live classes and
--  leaves findFieldRecursive generic and armed" — is CORRECT and is the point:
--  the disposition is per field/caller pair BY DESIGN (that is what the flip's
--  precondition asks for), and the other three classes are named with owners in
--  the lane NOTES. Nothing here should be read as evidence the resolver is safe.
-- ============================================================================

BEGIN;

-- GUARD: refuse unless the live row is the one this file was written against.
DO $$
DECLARE
    n    int;
    step jsonb;
    cfg  jsonb;
    k    text;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'tool-generator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '512: expected exactly 1 live tool-generator row, found %', n;
    END IF;

    SELECT default_config #> ARRAY['workflow','steps','enqueue_rerender'] INTO step
      FROM agent_definitions
     WHERE type = 'tool-generator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF step IS NULL THEN
        RAISE EXCEPTION '512: tool-generator has no enqueue_rerender step — the workflow has been restructured since 2026-08-20; re-derive this migration';
    END IF;
    IF step->>'action' <> 'create_rerender_items' THEN
        RAISE EXCEPTION '512: enqueue_rerender runs %, not create_rerender_items — input_fields would be read by a different spec', step->>'action';
    END IF;

    cfg := step->'config';

    IF cfg ? 'input_fields' THEN
        RAISE EXCEPTION '512: enqueue_rerender ALREADY carries input_fields (%) — another session has applied this or an equivalent; do not overwrite it', cfg->'input_fields';
    END IF;

    -- The premise of this file is that `reason` is UNWIRED. If a session has
    -- since wired it, Strategy 0 resolves it, the search never runs, and this
    -- edit is answering a question nobody is asking any more.
    IF cfg ? 'reason' THEN
        RAISE EXCEPTION '512: enqueue_rerender now WIRES reason (%) — the premise of this migration is gone; re-read the class before applying', cfg->'reason';
    END IF;

    -- The two fields being declared must still be wired where the analysis found
    -- them; a declaration listing a field the step cannot resolve would move the
    -- guess rather than remove it.
    FOREACH k IN ARRAY ARRAY['site_id','domain'] LOOP
        IF NOT (cfg ? k) THEN
            RAISE EXCEPTION '512: enqueue_rerender no longer wires % — declaring it would hand the field back to the whole-tree search', k;
        END IF;
    END LOOP;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
           '{workflow,steps,enqueue_rerender,config,input_fields}',
           '["site_id","domain"]'::jsonb, true),
       updated_at = NOW()
 WHERE type = 'tool-generator'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

-- VERIFY — a DO block that RAISEs, not a SELECT: ON_ERROR_STOP does not stop a
-- COMMIT on a non-empty result set (LANDMINES / RFC_006).
DO $$
DECLARE
    cfg       jsonb;
    leaked    text;
BEGIN
    SELECT default_config #> ARRAY['workflow','steps','enqueue_rerender','config'] INTO cfg
      FROM agent_definitions
     WHERE type = 'tool-generator' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg->'input_fields' <> '["site_id","domain"]'::jsonb THEN
        RAISE EXCEPTION '512 VERIFY: enqueue_rerender.input_fields is %, want ["site_id","domain"]', cfg->'input_fields';
    END IF;

    -- jsonb_set is surgical; assert it rather than trusting it. Every key the
    -- step carried before must still be there with its value.
    IF cfg->>'domain'          IS DISTINCT FROM 'input_data.domain'      THEN RAISE EXCEPTION '512 VERIFY: domain did not survive: %', cfg::text; END IF;
    IF cfg->>'site_id'         IS DISTINCT FROM 'site_record.site_id'    THEN RAISE EXCEPTION '512 VERIFY: site_id did not survive: %', cfg::text; END IF;
    IF cfg->>'error_step'      IS DISTINCT FROM 'complete'               THEN RAISE EXCEPTION '512 VERIFY: error_step did not survive: %', cfg::text; END IF;
    IF cfg->>'page_id_field'   IS DISTINCT FROM 'create_result.page_id'  THEN RAISE EXCEPTION '512 VERIFY: page_id_field did not survive: %', cfg::text; END IF;
    IF cfg->>'filename_field'  IS DISTINCT FROM 'create_result.page_url' THEN RAISE EXCEPTION '512 VERIFY: filename_field did not survive: %', cfg::text; END IF;
    IF cfg->>'page_name_field' IS DISTINCT FROM 'create_result.function' THEN RAISE EXCEPTION '512 VERIFY: page_name_field did not survive: %', cfg::text; END IF;

    -- Still UNWIRED, and still not searched-for by a back door: the point of the
    -- file is that `reason` resolves to nothing here, not that it resolves to
    -- something else.
    IF cfg ? 'reason' THEN
        RAISE EXCEPTION '512 VERIFY: a reason key appeared during this transaction: %', cfg->'reason';
    END IF;
    IF cfg->'input_fields' @> '["reason"]'::jsonb THEN
        RAISE EXCEPTION '512 VERIFY: reason is in the declared list — that asks the search for it again';
    END IF;

    -- NEGATIVE CONTROL in the same transaction: the other two live
    -- create_rerender_items steps (nav-updater, rerender-pages) must be
    -- untouched. A wider WHERE would pass every assertion above identically.
    --
    -- RECURSIVE, and that is the council's objection made mechanical (guardian
    -- seat, corr 2bd7fb37, round 1 — advisory, low): a control built on a bare
    -- `jsonb_each(default_config->'workflow'->'steps')` is BLIND to a step nested
    -- in a `config.sub_workflow.steps`, which is this estate's twice-bitten census
    -- trap (lane RUNBOOK, "the jsonb_each(steps) census is TOP-LEVEL ONLY").
    -- Measured 2026-08-20 17:4xZ: the recursive count equals the top-level count
    -- (the same 3 steps), so the claim was TRUE — but it was true unverifiably,
    -- and a file that reads as a template should not teach the blind form.
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
     WHERE step->>'action' = 'create_rerender_items'
       AND step->'config' ? 'input_fields'
       AND type <> 'tool-generator';
    IF leaked IS NOT NULL THEN
        RAISE EXCEPTION '512 VERIFY: the declaration leaked to steps it was not meant for: %', leaked;
    END IF;

    RAISE NOTICE '512 OK: tool-generator.enqueue_rerender declares [site_id, domain]; reason and component_id are no longer searched for; nav-updater and rerender-pages untouched';
END $$;

COMMIT;
