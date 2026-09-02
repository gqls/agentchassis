-- 701_retype_357_population_by_adoption_HOLD.sql
--
-- bugs_open/357 phase 3, by OWNER DECISION "Option B" (2026-09-02): each of the
-- 22 mislabelled page_components rows — a whole working tool stored under the
-- shared hero component's identity — gets its OWN per-tool content_components
-- row whose html_template IS the stored rendered_html, byte for byte, with the
-- page row, the current plan element and pages.sections repointed at it in the
-- same transaction. No LLM regeneration, no byte moves anywhere.
--
-- ⚠ THIS FILE SUPERSEDES MIGRATION 578 FOR THIS POPULATION. 578 (retype onto
-- the shared adopted-fragment component) is the road not taken: applying BOTH
-- would have the two repairs fight over the same 22 rows. 578 stays in
-- sql_for_agents as the record of the alternative; do not apply it, and do not
-- edit it. Why Option B won: 578's safety depended on the UNTESTED Layer 2
-- identity-carry surviving a real rebuild (precondition 4, which the lane
-- proved could only ever be tested vacuously — HANDOFF_2026-08-26b Findings
-- 1-2). Adoption removes that dependency: once the declared template IS the
-- stored bytes, the regeneration that is 357's disaster arm (a 2KB hero band
-- minted over a 16KB tool) becomes a byte-identical no-op BY CONSTRUCTION.
--
-- Crib: lendzy's migration 693 (693_lendzy_adopt_three_orphan_tool_components
-- .sql) — the NULL-component_id arm of the same family (see 693's own note (b):
-- "adoption may be exactly their fix shape — CONTRIB'd to their session").
-- This file copies its guard/verify/doc_notes/work-item shapes. Where it
-- deviates, the deviation is stated here and argued in the lane's
-- notes_700_draft.md:
--
--   1. WRONG-id arm, not NULL-id: the guards assert each row still carries the
--      shared hero id 23f95f00-f293-466e-b43a-81791ea0fc6c with its pinned md5,
--      rather than component_id IS NULL.
--   2. PINNED CENSUS, not predicate-only selection. 693/578 selected by
--      predicate; here the owner's decision names these 22 rows, and two of the
--      measurements that license adoption (zero '{{' bindings; every body PASSES
--      toolTemplateValid through the real production function, both-way
--      controls, 2026-09-02) were taken against these exact bodies. A row
--      minted since the census has neither measurement, so the population
--      predicate is RE-RUN at apply time and the transaction ABORTS unless it
--      returns exactly the pinned set — growth means re-census and a new file,
--      never a silent widening.
--   3. PLANS ARE POPULATED HERE (lendzy's were empty), so the repoint is
--      three-legged: page_components.component_id/slot_name, the page's
--      site_plan_sections 'hero' element in the single current plan, and the
--      derived pages.sections copy (sync_pages overwrites pages.sections from
--      the plan — reconcile_site_plan_action.go:599 — so BOTH must move: the
--      plan for durability, pages.sections for immediate effect). No new
--      site_plans row is created and pages.built_from_plan_version is NOT
--      touched, so the drift reconciler (which compares built_from_plan_version
--      against the current site_plans.id) sees nothing — same stance as 693.
--   4. ALIGNMENT: plan element = slot_name = the new component's NAME (the
--      CLC-020 '<function>-<domainSlug>' string). Justification: section
--      resolution at rebuild has three paths and the STORED component_id wins
--      first (rerender_page_sections_action.go resolveComponent; plan_sections_
--      action.go loadComponentSchemasByID), so the repointed row resolves
--      directly; the plan element name is the writer-path key, and Layer 2
--      matches stored rows to incoming ones on SLOT equality. Making all three
--      the same string — a string that content_components_name_key guarantees
--      names exactly one component fleet-wide, unlike 'hero' — means every
--      name-keyed path resolves to the same row the id path resolves to, and
--      the splice match is trivial instead of resting on the untested carry.
--      (578 preserved slot='hero' because it did NOT move the plan and a lone
--      slot rename would have armed the re-append landmine; 693 set
--      slot=p.name but its plans were empty so nothing constrained it. Moving
--      plan element and slot TOGETHER, in one transaction, is what keeps them
--      matched here.)
--   5. RFC_036 §9.3 (fork, not insert): an active manual library tool
--      'tool-equity-release_pre_037' (a5236dec-c38e-46b5-8f05-4f7b43fa2f3f)
--      already claims function 'tool-equity-release', so mortgagecalculator's
--      adopted component for that page is born a FORK (forked_from = that id),
--      which also exits idx_cc_tool_function_unique's predicate. The other 21
--      insert with forked_from NULL. NOTE, measured 2026-09-02 while drafting:
--      the same site ALSO holds an unplaced deploy-path fork of that library
--      tool ('tool-equity-release_pre_037-mortgagecalculator-co-uk',
--      befacff0-…, 0 placements) — a second site copy is exactly the RFC_036
--      §11-addendum-2 tracked tail (the two fork producers do not recognise
--      each other's copies); it breaks nothing here and is flagged for the
--      council in the notes.
--   6. gamesdesign/tool-ttk-calculator HAS NO PLAN ROWS AT ALL [MEASURED
--      2026-09-02 at draft time — the site's single current plan, 44 rows,
--      created 2026-06-05, carries nothing for that page_name, while
--      pages.sections says ["hero","generic-text-block"]]. This CONTRADICTS
--      the lane's earlier "exactly one hero plan row per page" census line for
--      that one page. Handled explicitly: its pages.sections is repointed, the
--      plan leg is skipped, and the guard PINS the zero-plan-rows state so a
--      plan row appearing before apply aborts the run.
--   7. vetcomparison.uk's page is named 'index'; a mechanical CLC-020 function
--      would claim the fleet-wide tool function 'index' (idx_cc_tool_function_
--      unique has no site column), poisoning the name every site's homepage
--      shares. DECISION: semantic function 'tool-vet-comparison', name
--      'tool-vet-comparison-vetcomparison-uk'. The join stays honest because
--      this migration keys every step on the pinned pc_id census, never on
--      693's cc.function = p.name join — the naming is data in the census, not
--      a derivation. Trade-off argued in the notes.
--   8. RERENDERS ARE FILED FOR generic PAGES ONLY (16 of 22). The six
--      gamesdesign rebuild_policy='owned' rows are refused at assemble by the
--      owned-page guard (the 578 file's own finding: the pipeline refuses these
--      pages, which is also why a row repaired here STAYS repaired — they are
--      the most durable targets, not the riskiest). Filing rerenders for them
--      would mint six guaranteed failures. The filing SHAPE is 693's exactly.
--   9. NO component_versions rows are minted (693's stance). Verified at the
--      resolve path before drafting: loadComponentSchemasByID reads
--      content_components only, and no rerender/plan/save action references
--      component_versions — render does not require a version row.
--      page_components.component_version_id is a provenance stamp, left NULL.
--
-- LOSSLESSNESS LICENCE [MEASURED 2026-09-02]: all 22 stored bodies contain
-- ZERO '{{' template bindings, so the rendered form IS the template and there
-- is nothing to bind; and all 22 pass toolTemplateValid (the guard
-- loadComponentSchemasByID applies to component_level='tool' rows) through the
-- real production function, with a passing and a failing control in the same
-- run. Guard 3 re-checks the bindings claim at apply time.
--
-- ============================================================================
-- HOW TO APPLY — by hand, one transaction, never the runner (the _HOLD suffix
-- keeps SIDECAR_RE away), after council review:
--
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -v scope=pilot \
--     < 701_retype_357_population_by_adoption_HOLD.sql
--
--   -v scope=pilot      ONE row: mortgagecalculator.co.uk/tool-simple
--                       (the owner's chosen pilot; planned=1, generic, single row)
--   -v scope=remainder  the other 21 — REFUSES to run unless the pilot row is
--                       already repaired and verified in place
--   -v scope=all        all 22 in one transaction (only if the owner decides to
--                       skip the pilot; the burst is 16 rerenders at once)
--   (omitted)           defaults to PILOT — the smallest honest action, matching
--                       the lane plan, NOT all
--
-- The scope reaches the DO blocks as a transaction-local GUC (set_config …,
-- true), because psql does not interpolate :variables inside dollar-quoted
-- bodies. Deliberate side effect: run through anything that skips the psql
-- preamble and current_setting('m701.scope') raises, so the file refuses to
-- run outside its own harness.
--
-- ============================================================================
-- POST-CONDITIONS FOR THE OPERATOR (the migration cannot observe its async
-- outcome — run after the filed rerenders reach 'complete'):
--   1. Pilot pass condition (Option B's own, superseding Finding 2's): the
--      tool-simple rerender completes; the page still has EXACTLY ONE
--      page_components row (a count of 2 is the carry-forward landmine firing —
--      STOP, do not proceed to the remainder); md5(rendered_html) still
--      7873509b8087a15cc3b32120e746f9e5 (a byte-identical re-render also
--      passes — the template IS the bytes); component still
--      'tool-simple-mortgagecalculator-co-uk'.
--   2. At the ARTEFACT, with the invented-URL control (a parked domain 200s
--      every path): each repaired page still serves its tool markup — its
--      <input>/<canvas>/<script> counts, not just HTTP 200.
--   3. The closing metric: the population query (Guard 2's predicate) returns
--      0 after the full repair — bugs_open/357 closes at population = 0
--      VERIFIED AT THE ARTEFACTS (lane close-out bar §3), not at this commit.
--   4. The false required_fields_missing items about 'hero' on these pages
--      should stop being re-filed; they are left to their own verifiers, not
--      hand-closed (693's lesson (d)).
--
-- HOW TO READ THE SERVED PAGE (council round 1, debug_historian): send a
-- browser-like Accept header (a bare curl default can draw a different body
-- from the edge than a browser sees), wait out the publish lag before judging
-- (measured 11-97s across six publishes, webdesign lane 2026-09-02 — a
-- refusal inside that window is lag, not failure; re-read after it), and
-- assert the tool's own MARKERS (its <input>/<button>/<script>), never byte
-- equality against the DB row (head/chrome injection makes served != stored
-- by design).
--
-- ============================================================================
-- THE 357-KEYED LANDMINE, ENGAGED (council round 1, debug_historian HIGH).
-- LANDMINES.md carries: "a page whose only component row is an
-- adopted-fragment CANNOT be rebuilt — flagging it needs_rebuild crashes the
-- pod" (bugs_open/408, cv1.co.uk). That trap does NOT transfer to 701's rows,
-- and the difference is the mechanism, not hope:
--
--   * The cv1 crash chain NEEDS AN EMPTY RENDER: the shared adopted-fragment
--     template binds {{.body}}, which renders empty on a fresh write, so the
--     compile keeps 0 sections, page_html never exists, and the old
--     extractFieldValue recursed to a stack overflow (408 §4).
--   * 701's adopted templates carry ZERO template bindings — Guard 3 aborts
--     the transaction if that ever stops being true — so a rebuild of a
--     post-701 page renders the template TO ITS LITERAL BYTES: non-empty by
--     the same measurement that licenses the adoption. The no-content path is
--     unreachable, not merely unexercised.
--   * needs_rebuild flags on these pages are a REAL possibility, not
--     hypothetical (writers as of 2026-09-02: page_build_failure_guard.go:102
--     /:120, maintenance flagPagesForRebuild, the planner re-plan path,
--     store_generated_component_action.go:1359) — which is exactly why the
--     non-empty-render property above, not process discipline, is the guard.
--   * Belt and braces: bugs_open/408's fix is committed and council-approved
--     (6e2d4a039 + b8bf40694, corr 3918db52) — once an image >= those rolls,
--     even a genuinely absent content field skips cleanly instead of crashing.
--   * 701's OWN filings are page_rerender items (a path measured free of the
--     408 defect: zero extractFieldValue/skip_reason occurrences in
--     rerender_single_page_action.go, 2026-09-02), never needs_rebuild.
--
-- ============================================================================
-- PRIOR ART, READ AND DECLINED (council round 2, prior_art_librarian HIGH).
-- ConvertTemplateToInstanceScope (component_instance_conversion.go) and
-- ScopeToolBirthTemplate (tool_birth_instance_scope.go), routed via the
-- 'instance_scope_conversion' work-item type, are the bugs_open/283 programme:
-- they rewrite a SHARED template so element ids are namespaced per instance
-- ({{.InstanceID}}- prefixes on declared ids and every reference to them), so
-- that MULTIPLE instances of one component can share a page. Read at source
-- 2026-09-02; not used here, for a mechanism reason, not preference:
--   * Applying either would REWRITE THE SERVED BYTES (id attributes and every
--     script/CSS reference), which violates this migration's core invariant —
--     byte-for-byte adoption of markup that is LIVE and SERVING today.
--   * The defect class they close — same-page multi-instance id collisions —
--     cannot exist for these rows: each adopted component has exactly ONE
--     placement (its own page), recorded in the doc_note. DISCLOSED: a future
--     SECOND placement of an adopted component would re-open that class; that
--     is the pre-existing property of every adopted/owned tool on this estate,
--     not a state this migration creates.
--   * 693 (the crib) minted its three adopted components the same way, and its
--     own prior_art round was gated on a DIFFERENT question (an inert repair —
--     no rerender half), answered there by shipping the missing half; this
--     file carries its rerender half from round 1.
--
-- RFC_036 §9.3 LIVENESS — PROBED, NOT ASSUMED (council round 2,
-- prior_art_librarian medium). The future-forks story for the 21 library
-- claims depends on the §9.3 fork logic being IN THE RUNNING BINARY, which is
-- a deployed-artefact claim, not a source claim. Probed 2026-09-02 against
-- agent-chassis v1.0.1354 (/proc/1/exe): the §9.3 log literal "new component
-- forks from it" PRESENT (1), positive control "Field not found in path"
-- PRESENT, junk negative control ABSENT — the probe discriminates. Re-run
-- before applying if the fleet has rolled backward:
--   kubectl -n ai-persona-system exec <chassis-pod> -- \
--     grep -ac "new component forks from it" /proc/1/exe   # expect 1+
--
-- BACKUP TABLE vs THE HISTORY TRIGGERS (council round 2, reuse_agent). The
-- automatic net exists and fires on this migration's UPDATEs:
-- trg_page_component_artefact_archive_upd/_del → page_component_history
-- (verified in pg_trigger 2026-09-02). The dedicated backup table is kept
-- IN ADDITION because (a) the trigger archives page_components only — the
-- pages.sections and site_plan_sections pre-states have no trigger and are
-- half the rollback; (b) the rollback needs a KEYED, self-contained restore
-- source with the new_name mapping; (c) history rows are subject to pruning
-- (a history row with rendered_html NULL is a PRUNED row — STY-056), so
-- history alone is not a restore source one may rely on.
--
-- PRE-APPLY COLLISION CHECK (council round 2, guardian): this file was
-- RENUMBERED from 700 after another lane's committed 700 — the number-ledger
-- keys on filenames. Before applying:
--   ls docs/agent_docs/sql_for_agents/ | grep '^701'   # exactly this pair
--
-- THE 21 FLEET-WIDE FUNCTION CLAIMS ARE AN OWNER-LEVEL ACCEPTANCE (council
-- round 2, guardian missing): no register names an owner of the tool-function
-- namespace; naming follows CLC-020. This file is applied BY THE OWNER'S HAND,
-- and that application is the sign-off on the claims — stated here so the
-- acceptance is a decision, not a side effect.
--
-- BINDING-FREE IS GUARDED AT APPLY, NOT FOR EVER (council round 2,
-- bug_historian): template_closed is a quality-score field, not an edit lock
-- (compute_component_quality.go — measured, no standing invariant exists to
-- borrow, and minting a new detector inside a repair migration is the seam-
-- in-a-bug-patch shape this council rightly vetoes). A later deliberate
-- template edit that introduces bindings is a normal component change owned
-- by the normal flows; its worst rebuild outcome SINCE v1.0.1354 (the 408 fix
-- aboard, probed above) is an empty section render — a visible content
-- regression, not a pod crash. The three-way equality the repair establishes
-- is asserted per-row by the verify (slot_name = cc.name = plan element —
-- the direct answer to the bugs 095/039 wrong-slot shape).
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/bugfix_357_component_identity/
-- Rollback: 701_retype_357_population_by_adoption_HOLD_ROLLBACK.sql
-- ============================================================================

\if :{?scope}
\else
\set scope pilot
\endif

BEGIN;

SELECT set_config('m701.scope', :'scope', true);

-- ---------------------------------------------------------------------------
-- 0. Scope validation + the pinned census. Every later step keys on this
-- table; the VALUES are the owner's decision, measured 2026-09-02.
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  IF current_setting('m701.scope') NOT IN ('pilot', 'remainder', 'all') THEN
    RAISE EXCEPTION '701 ABORT: scope must be pilot | remainder | all, got %',
      current_setting('m701.scope');
  END IF;
END $$;

CREATE TEMP TABLE census_700 ON COMMIT DROP AS
SELECT v.pc_id::uuid, v.page_id::uuid, v.site_id::uuid, v.domain, v.page_name,
       v.md5_pinned, v.bytes_pinned::int, v.policy_pinned,
       v.has_plan_row::boolean, v.is_pilot::boolean,
       v.fork_parent::uuid, v.new_function, v.new_name,
       v.display_name, v.description,
       CASE current_setting('m701.scope')
         WHEN 'pilot'     THEN v.is_pilot::boolean
         WHEN 'remainder' THEN NOT v.is_pilot::boolean
         ELSE true
       END AS is_target
FROM (VALUES
  -- pc_id | page_id | site_id | domain | page_name | md5 | bytes | policy | has_plan_row | is_pilot | fork_parent | new_function | new_name | display_name | description
  ('403ac32c-3aeb-4865-9c5f-a631161b4ffa','da5eba0a-a352-4b22-a314-fb57095674d4','e33263f4-74f8-494f-b191-546845dbbddf','gamesdesign.co.uk','game-auto-battler','920c0c7048bbef9ac8e186ea54b64ec8','21654','generic',true,false,NULL,'game-auto-battler','game-auto-battler-gamesdesign-co-uk','Auto-battler simulator','Interactive auto-battler game-design simulator, adopted 2026-09-02 from the live build.'),
  ('db8f23cc-9638-4d8d-ad56-1c10ff86f601','d9a8e6e8-e40b-49df-b613-2fed24c1bc78','e33263f4-74f8-494f-b191-546845dbbddf','gamesdesign.co.uk','game-economy-simulator','3cf4b59e1fd7decd7b9423514b58fcbf','17539','generic',true,false,NULL,'game-economy-simulator','game-economy-simulator-gamesdesign-co-uk','Game economy simulator','Interactive in-game economy simulator for game designers, adopted 2026-09-02 from the live build.'),
  ('a611ff25-8bc3-4728-a48f-8808fd383075','7674c1ec-ebba-4bf9-a1db-c32980e16aea','e33263f4-74f8-494f-b191-546845dbbddf','gamesdesign.co.uk','game-p2p-networking','17bd0bfc51bf68516e2cf159c48aaaaa','21797','generic',true,false,NULL,'game-p2p-networking','game-p2p-networking-gamesdesign-co-uk','P2P networking simulator','Interactive peer-to-peer networking simulator for game designers, adopted 2026-09-02 from the live build.'),
  ('5fa68b6c-4706-463f-803a-92bf52d65b61','56af8679-1f7d-4da6-b148-f5727b16693d','e33263f4-74f8-494f-b191-546845dbbddf','gamesdesign.co.uk','game-pathfinding','231b6772c9afbc27176c87dba29e1774','20464','generic',true,false,NULL,'game-pathfinding','game-pathfinding-gamesdesign-co-uk','Pathfinding simulator','Interactive pathfinding algorithm simulator for game designers, adopted 2026-09-02 from the live build.'),
  ('15f1f798-51fb-41d0-8a07-18148b39a293','0f9ed454-1495-4f36-a809-9f4198816308','e33263f4-74f8-494f-b191-546845dbbddf','gamesdesign.co.uk','tool-drop-rate-simulator','77fde2ca8e4f2c765944b97b411b7577','15675','owned',true,false,NULL,'tool-drop-rate-simulator','tool-drop-rate-simulator-gamesdesign-co-uk','Drop rate simulator','Simulates loot drop rates over repeated runs, adopted 2026-09-02 from the live build.'),
  ('f5bbe031-b147-47e2-bb8f-ab7676af3526','0c437f83-69b4-4a0c-a7e8-2295ae112b45','e33263f4-74f8-494f-b191-546845dbbddf','gamesdesign.co.uk','tool-ehp-calculator','45497d6d72beb4595ffcaf245d7a3399','15451','owned',true,false,NULL,'tool-ehp-calculator','tool-ehp-calculator-gamesdesign-co-uk','EHP calculator','Calculates effective hit points from defensive stats, adopted 2026-09-02 from the live build.'),
  ('cbd6ec6d-ccb2-413f-a8c6-0b8c6f7be1b3','5117e035-53f7-4ee2-ae6c-b426307a75f6','e33263f4-74f8-494f-b191-546845dbbddf','gamesdesign.co.uk','tool-jump-physics','afd2b950514a50b79eb992c5a9303cbb','15572','owned',true,false,NULL,'tool-jump-physics','tool-jump-physics-gamesdesign-co-uk','Jump physics tool','Works out platformer jump physics parameters, adopted 2026-09-02 from the live build.'),
  ('4a376c04-b39f-4adf-a8f4-ec485339a2be','581d25fc-bce3-4ed8-ac5f-ee75ca2f9e10','e33263f4-74f8-494f-b191-546845dbbddf','gamesdesign.co.uk','tool-lanchester-sim','53686668a7ee36e565aafce0db108b7f','16718','owned',true,false,NULL,'tool-lanchester-sim','tool-lanchester-sim-gamesdesign-co-uk','Lanchester combat simulator','Simulates combat outcomes using the Lanchester equations, adopted 2026-09-02 from the live build.'),
  ('6bbc7bdb-f66c-4948-83a6-ca43d471299f','f2470f08-ea28-4ecf-8821-bbcc5ba129a3','e33263f4-74f8-494f-b191-546845dbbddf','gamesdesign.co.uk','tool-progression-architect','cc9f0458db754b5d7166814a9ccd151e','18711','owned',true,false,NULL,'tool-progression-architect','tool-progression-architect-gamesdesign-co-uk','Progression architect','Designs player progression and experience curves, adopted 2026-09-02 from the live build.'),
  ('77eaa64e-524d-401c-88c3-c0f485809351','da4a0559-4bfa-49b3-b5b9-4e0d8c5492fb','e33263f4-74f8-494f-b191-546845dbbddf','gamesdesign.co.uk','tool-ttk-calculator','b73b3913a53ea72794adc82c2348757d','16106','owned',false,false,NULL,'tool-ttk-calculator','tool-ttk-calculator-gamesdesign-co-uk','TTK calculator','Calculates time-to-kill from weapon and health stats, adopted 2026-09-02 from the live build.'),
  ('9aef5742-27f9-4602-a6b5-56f829e67a4f','fdb89c44-d2cc-4e94-989a-f405d584d9ed','62b5978e-4271-4589-8e00-4baebfc0447c','mortgagecalculator.co.uk','investor-index','7d57fd68cdd18957f3b9280594381dc0','12485','generic',true,false,NULL,'investor-index','investor-index-mortgagecalculator-co-uk','Investor tools index','Buy-to-let investor tools index page with interactive content, adopted 2026-09-02 from the live build.'),
  ('b6748524-a175-4fb2-838d-7666face0583','19833410-226c-441c-831b-b08fbf34cdf0','62b5978e-4271-4589-8e00-4baebfc0447c','mortgagecalculator.co.uk','tool-affordability','c8ad88f52d157a2c9e353ab2817cae27','20395','generic',true,false,NULL,'tool-affordability','tool-affordability-mortgagecalculator-co-uk','Affordability calculator','Mortgage affordability calculator, adopted 2026-09-02 from the live build.'),
  ('aa4b0d02-0a77-4f91-a1eb-1657b6d266a8','e4824872-3594-49d1-a7f4-f258fd0e821d','62b5978e-4271-4589-8e00-4baebfc0447c','mortgagecalculator.co.uk','tool-bridging-loan','1acb6bc5f8ac9df7dc7b3c66887f27a3','11621','generic',true,false,NULL,'tool-bridging-loan','tool-bridging-loan-mortgagecalculator-co-uk','Bridging loan calculator','Bridging loan cost calculator, adopted 2026-09-02 from the live build.'),
  ('6f07b479-605f-454c-b406-d245bc776e6a','4da41aed-9b3a-4bac-821f-a4ea8e19c2ff','62b5978e-4271-4589-8e00-4baebfc0447c','mortgagecalculator.co.uk','tool-equity-release','855b10916a10da28d00e4f3a81e702af','14164','generic',true,false,'a5236dec-c38e-46b5-8f05-4f7b43fa2f3f','tool-equity-release','tool-equity-release-mortgagecalculator-co-uk','Equity release calculator','Equity release calculator, adopted 2026-09-02 from the live build.'),
  ('f683b329-ace7-40a8-be15-9baa8c9a7b63','9325896e-9815-42c3-90ae-8a226aba0255','62b5978e-4271-4589-8e00-4baebfc0447c','mortgagecalculator.co.uk','tool-fee-analyser','c6453628c751b501d03a123dec78215d','12730','generic',true,false,NULL,'tool-fee-analyser','tool-fee-analyser-mortgagecalculator-co-uk','Fee analyser','Mortgage fee analyser, adopted 2026-09-02 from the live build.'),
  ('c92290c1-418e-4bc7-b710-a95f4dfecf15','979675ad-5371-4452-8988-a3433fb6e5d9','62b5978e-4271-4589-8e00-4baebfc0447c','mortgagecalculator.co.uk','tool-overpayment','e24864c58fdb29238ad3215726d738c1','11517','generic',true,false,NULL,'tool-overpayment','tool-overpayment-mortgagecalculator-co-uk','Overpayment calculator','Mortgage overpayment calculator, adopted 2026-09-02 from the live build.'),
  ('33c28099-77b4-4d03-83bb-ca3c054ab4d4','dee14ba2-9d0f-44c0-8b75-376fc3595cae','62b5978e-4271-4589-8e00-4baebfc0447c','mortgagecalculator.co.uk','tool-portfolio','c4832ed182531cca3a9213c1af42a078','22494','generic',true,false,NULL,'tool-portfolio','tool-portfolio-mortgagecalculator-co-uk','Portfolio calculator','Buy-to-let portfolio calculator, adopted 2026-09-02 from the live build.'),
  ('4e259f57-d490-4cd1-8614-04fae4a9be93','ca184e21-e067-4f7a-969e-8971fec8bbc8','62b5978e-4271-4589-8e00-4baebfc0447c','mortgagecalculator.co.uk','tool-rate-forecaster','c5b2dee7bb621d887030fd85d1bd9cf3','12770','generic',true,false,NULL,'tool-rate-forecaster','tool-rate-forecaster-mortgagecalculator-co-uk','Rate forecaster','Mortgage rate forecasting tool, adopted 2026-09-02 from the live build.'),
  ('01534267-0b2a-45c1-9f4d-a149cd86990e','1f59a7d6-6ff9-47d0-866e-df7e993de416','62b5978e-4271-4589-8e00-4baebfc0447c','mortgagecalculator.co.uk','tool-repayment','b3c7712c51fd7289c9627d92fb502a02','14554','generic',true,false,NULL,'tool-repayment','tool-repayment-mortgagecalculator-co-uk','Repayment calculator','Mortgage repayment calculator, adopted 2026-09-02 from the live build.'),
  ('cb406ec9-de30-4775-85ea-7bf6828b6c6f','6bdc8425-ff13-4a0b-bef0-fbfafda96992','62b5978e-4271-4589-8e00-4baebfc0447c','mortgagecalculator.co.uk','tool-simple','7873509b8087a15cc3b32120e746f9e5','9590','generic',true,true,NULL,'tool-simple','tool-simple-mortgagecalculator-co-uk','Simple mortgage calculator','Simple mortgage repayment calculator, adopted 2026-09-02 from the live build.'),
  ('46763ae8-17ab-494d-8d4f-1912d2648548','3d7d0d72-ff6b-4415-b786-a80e032e9827','62b5978e-4271-4589-8e00-4baebfc0447c','mortgagecalculator.co.uk','tool-stamp-duty','fb86cd1f97ad3937abd25e692cd1afbe','12734','generic',true,false,NULL,'tool-stamp-duty','tool-stamp-duty-mortgagecalculator-co-uk','Stamp duty calculator','Stamp duty calculator, adopted 2026-09-02 from the live build.'),
  ('788c88cf-fcb6-4808-8de3-afd39beb1d30','9fad89c1-ef22-4f15-a2f1-cd2b0a781378','72b9e3a6-872f-4528-a6d6-7f205ea60f4d','vetcomparison.uk','index','55b4348a6706b1a08302e4d821dfd193','11326','generic',true,false,NULL,'tool-vet-comparison','tool-vet-comparison-vetcomparison-uk','Vet practice comparison tool','Vet practice comparison tool serving the vetcomparison.uk homepage, adopted 2026-09-02 from the live build.')
) AS v(pc_id, page_id, site_id, domain, page_name, md5_pinned, bytes_pinned, policy_pinned, has_plan_row, is_pilot, fork_parent, new_function, new_name, display_name, description);

-- Census integrity: a typo above must die here, not downstream.
DO $$
DECLARE n int; d int; f int; p int;
BEGIN
  SELECT count(*), count(DISTINCT pc_id), count(DISTINCT new_name), count(*) FILTER (WHERE is_pilot)
    INTO n, d, f, p FROM census_700;
  IF n <> 22 OR d <> 22 OR f <> 22 OR p <> 1 THEN
    RAISE EXCEPTION '701 ABORT: census integrity — rows=% distinct_pc=% distinct_names=% pilots=% (want 22/22/22/1)', n, d, f, p;
  END IF;
  SELECT count(DISTINCT new_function) INTO f FROM census_700;
  IF f <> 22 THEN
    RAISE EXCEPTION '701 ABORT: census integrity — % distinct functions, want 22', f;
  END IF;
  RAISE NOTICE '701 scope=%: % target row(s) this run',
    current_setting('m701.scope'), (SELECT count(*) FROM census_700 WHERE is_target);
END $$;

-- ---------------------------------------------------------------------------
-- GUARD 1. RFC_036 §9.3: if a LIBRARY tool (component_level=tool, forked_from
-- NULL, is_active) claims one of these functions, the new row must be born a
-- FORK, and a bare INSERT dies on idx_cc_tool_function_unique (23505).
-- Exactly ONE such claim is expected — tool-equity-release_pre_037 — and the
-- census already carries its fork. Any OTHER claim, or that one going missing
-- or changing shape, means the world moved since 2026-09-02: abort.
-- TARGET rows only: on scope=remainder the pilot's own adopted component is,
-- by design, an active library claim on 'tool-simple' (forked_from NULL), and
-- must not trip this guard — functions are distinct across the 22, so a
-- non-target's claim can never collide with a target's INSERT.
-- ---------------------------------------------------------------------------
DO $$
DECLARE bad text; par record;
BEGIN
  SELECT string_agg(cc.function || ' (' || cc.id || ')', ', ') INTO bad
    FROM content_components cc
    JOIN census_700 c ON c.new_function = cc.function
   WHERE cc.component_level = 'tool' AND cc.forked_from IS NULL AND cc.is_active
     AND c.is_target AND c.fork_parent IS NULL;
  IF bad IS NOT NULL THEN
    RAISE EXCEPTION '701 ABORT: library tool row(s) claim function(s) not pinned as forks — this migration must fork, not insert (RFC_036 9.3): %', bad;
  END IF;

  SELECT id, function, component_level, forked_from, is_active INTO par
    FROM content_components WHERE id = 'a5236dec-c38e-46b5-8f05-4f7b43fa2f3f';
  IF NOT FOUND
     OR par.function <> 'tool-equity-release'
     OR par.component_level <> 'tool'
     OR par.forked_from IS NOT NULL
     OR NOT par.is_active THEN
    RAISE EXCEPTION '701 ABORT: pinned fork parent a5236dec-… is no longer an active library claim of tool-equity-release — re-measure before forking from it';
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- GUARD 2. The pinned census, Guard-2 style: re-run the bug's own population
-- predicate (rows bound to the shared hero 23f95f00-… whose stored bytes do
-- not contain hero's literal template prefix) and ABORT unless it returns
-- exactly the pinned set still awaiting repair; per-row, every censused fact
-- must still hold (md5, bytes, slot, position, rebuild_policy, plan shape,
-- sections shape). On scope=remainder the pilot row must instead be verified
-- already REPAIRED. ANY drift aborts the whole transaction, naming the row.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  scope text := current_setting('m701.scope');
  hero_prefix text;
  errs text[] := '{}';
  extra text;
  r record; pc record; pg record;
  plan_hero int; plan_hero_any int; plan_new int; plan_any int; n_hero int; n_new int;
  cc_id uuid;
BEGIN
  SELECT left(html_template, position('{{' in html_template) - 1) INTO hero_prefix
    FROM content_components WHERE id = '23f95f00-f293-466e-b43a-81791ea0fc6c';
  IF hero_prefix IS NULL OR length(hero_prefix) < 10 THEN
    RAISE EXCEPTION '701 ABORT: the shared hero component is missing or its template no longer opens with a literal prefix — the population predicate is meaningless; re-derive it';
  END IF;

  -- (a) The population must not have GROWN beyond the census: a fresh mint has
  -- neither the zero-bindings nor the toolTemplateValid measurement and must
  -- never be swept in. It would also mean the phase 2 birth-fix regressed.
  SELECT string_agg(pc2.id::text, ', ') INTO extra
    FROM page_components pc2
   WHERE pc2.component_id = '23f95f00-f293-466e-b43a-81791ea0fc6c'
     AND position(hero_prefix in pc2.rendered_html) = 0
     AND pc2.id NOT IN (SELECT pc_id FROM census_700);
  IF extra IS NOT NULL THEN
    RAISE EXCEPTION '701 ABORT: the population has GROWN since the 2026-09-02 census — unpinned mislabelled row(s): %. Re-census (and check the phase 2 birth-fix) before repairing anything', extra;
  END IF;

  -- (b) Per pinned row.
  FOR r IN SELECT * FROM census_700 ORDER BY domain, page_name LOOP
    SELECT * INTO pc FROM page_components WHERE id = r.pc_id;
    IF NOT FOUND THEN
      errs := errs || (r.domain || '/' || r.page_name || ': page_components row ' || r.pc_id || ' is GONE');
      CONTINUE;
    END IF;
    SELECT * INTO pg FROM pages WHERE id = r.page_id;
    IF NOT FOUND THEN
      errs := errs || (r.domain || '/' || r.page_name || ': pages row is GONE');
      CONTINUE;
    END IF;
    IF pg.rebuild_policy <> r.policy_pinned THEN
      errs := errs || (r.domain || '/' || r.page_name || ': rebuild_policy is ' || pg.rebuild_policy || ', censused ' || r.policy_pinned);
    END IF;

    SELECT count(*) INTO n_hero FROM jsonb_array_elements(pg.sections) e WHERE e.value = to_jsonb('hero'::text);
    SELECT count(*) INTO n_new  FROM jsonb_array_elements(pg.sections) e WHERE e.value = to_jsonb(r.new_name);
    SELECT count(*) FILTER (WHERE sps.component_name = 'hero' AND sps.ordering = 0 AND sps.component_version_id IS NULL),
           count(*) FILTER (WHERE sps.component_name = 'hero'),
           count(*) FILTER (WHERE sps.component_name = r.new_name AND sps.ordering = 0),
           count(*)
      INTO plan_hero, plan_hero_any, plan_new, plan_any
      FROM site_plan_sections sps
      JOIN site_plans sp ON sp.id = sps.plan_id AND sp.is_current
     WHERE sp.site_id = r.site_id AND sps.page_name = r.page_name;

    IF scope = 'remainder' AND r.is_pilot THEN
      -- The pilot must already be REPAIRED, exactly as this migration leaves it.
      SELECT id INTO cc_id FROM content_components
       WHERE name = r.new_name AND created_from = 'adopted' AND is_active;
      IF cc_id IS NULL OR pc.component_id IS DISTINCT FROM cc_id
         OR pc.slot_name IS DISTINCT FROM r.new_name
         OR md5(COALESCE(pc.rendered_html, '')) <> r.md5_pinned
         OR n_new <> 1 OR n_hero <> 0
         OR plan_new <> 1 OR plan_hero_any <> 0 THEN
        errs := errs || (r.domain || '/' || r.page_name || ': scope=remainder but the PILOT row is not in this migration''s repaired state (all three legs) — run scope=pilot first, or re-read the lane NOTES');
      END IF;
      CONTINUE;
    END IF;

    -- All other rows (and the pilot outside remainder) must be censused-exact.
    IF pc.component_id IS DISTINCT FROM '23f95f00-f293-466e-b43a-81791ea0fc6c'::uuid THEN
      errs := errs || (r.domain || '/' || r.page_name || ': component_id is ' || COALESCE(pc.component_id::text, 'NULL') || ', not the shared hero — someone has already acted; re-read before applying');
    END IF;
    IF pc.slot_name IS DISTINCT FROM 'hero' OR pc.position IS DISTINCT FROM 1 THEN
      errs := errs || (r.domain || '/' || r.page_name || ': slot/position is ' || COALESCE(pc.slot_name, 'NULL') || '/' || pc.position || ', censused hero/1');
    END IF;
    IF md5(COALESCE(pc.rendered_html, '')) <> r.md5_pinned
       OR length(COALESCE(pc.rendered_html, '')) <> r.bytes_pinned THEN
      errs := errs || (r.domain || '/' || r.page_name || ': stored bytes moved — md5 ' || md5(COALESCE(pc.rendered_html, '')) || ' (' || length(COALESCE(pc.rendered_html, '')) || ' B), censused ' || r.md5_pinned || ' (' || r.bytes_pinned || ' B)');
    END IF;
    IF position(hero_prefix in COALESCE(pc.rendered_html, '')) <> 0 THEN
      errs := errs || (r.domain || '/' || r.page_name || ': now MATCHES the hero prefix — no longer in the population; re-census');
    END IF;
    IF n_hero <> 1 OR n_new <> 0 THEN
      errs := errs || (r.domain || '/' || r.page_name || ': pages.sections has ' || n_hero || ' hero element(s) and ' || n_new || ' of the new name, censused 1/0');
    END IF;
    IF r.has_plan_row AND (plan_hero <> 1 OR plan_hero_any <> 1) THEN
      errs := errs || (r.domain || '/' || r.page_name || ': expected exactly 1 hero plan row, at ordering 0 with version NULL, in the current plan — found ' || plan_hero || ' at ordering 0 and ' || plan_hero_any || ' hero row(s) overall');
    END IF;
    IF NOT r.has_plan_row AND plan_any <> 0 THEN
      errs := errs || (r.domain || '/' || r.page_name || ': censused with ZERO plan rows (the tool-ttk-calculator exception) but the current plan now has ' || plan_any || ' — the design premise changed; re-measure');
    END IF;
  END LOOP;

  IF array_length(errs, 1) IS NOT NULL THEN
    RAISE EXCEPTION '701 ABORT: % drifted row(s) against the pinned census: %',
      array_length(errs, 1), array_to_string(errs, ' || ');
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- GUARD 3. The stored HTML must still carry no template bindings. This is the
-- measurement the whole approach rests on; if it has changed, adoption is no
-- longer lossless and this migration is wrong. (693's guard 3, verbatim in
-- spirit; satisfied population-wide on 2026-09-02, kept anyway.)
-- ---------------------------------------------------------------------------
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM page_components pc JOIN census_700 c ON c.pc_id = pc.id
   WHERE c.is_target AND pc.rendered_html LIKE '%{{%';
  IF n <> 0 THEN
    RAISE EXCEPTION '701 ABORT: % stored body/bodies now contain template bindings — promoting rendered_html to html_template is no longer lossless', n;
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- GUARD 4. content_components_name_key is TOTAL (no predicate, is_active does
-- not free it): every proposed name must be unclaimed by ANY row.
-- ---------------------------------------------------------------------------
DO $$
DECLARE bad text;
BEGIN
  SELECT string_agg(cc.name, ', ') INTO bad
    FROM content_components cc JOIN census_700 c ON c.new_name = cc.name
   WHERE c.is_target;
  IF bad IS NOT NULL THEN
    RAISE EXCEPTION '701 ABORT: component name(s) already exist (content_components_name_key is total): %', bad;
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- 1. BACKUP, before anything moves — 578's practice widened to the three legs.
-- One table: the whole page_components row as jsonb, the load-bearing columns
-- extracted for the rollback's exact restores, plus the pages.sections and
-- site_plan_sections pre-states. ON CONFLICT DO NOTHING is safe ONLY because
-- Guard 2 has just proven the live pre-state equals the pinned census, so a
-- re-run after a rollback stores byte-identical facts.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS page_components_backup_357b_20260902 (
  pc_id              uuid PRIMARY KEY,
  page_id            uuid        NOT NULL,
  site_id            uuid        NOT NULL,
  domain             text        NOT NULL,
  page_name          text        NOT NULL,
  pre_component_id   uuid        NOT NULL,
  pre_slot_name      text        NOT NULL,
  pre_position       int         NOT NULL,
  pre_md5            text        NOT NULL,
  pre_bytes          int         NOT NULL,
  pre_row            jsonb       NOT NULL,  -- to_jsonb(page_components.*), whole row
  pre_pages_sections jsonb       NOT NULL,
  pre_plan_row_id    uuid,                  -- NULL for tool-ttk-calculator (no plan row exists)
  pre_plan_row       jsonb,
  new_name           text        NOT NULL,
  new_function       text        NOT NULL,
  applied_scope      text        NOT NULL,
  applied_at         timestamptz NOT NULL DEFAULT now()
);

INSERT INTO page_components_backup_357b_20260902
    (pc_id, page_id, site_id, domain, page_name,
     pre_component_id, pre_slot_name, pre_position, pre_md5, pre_bytes,
     pre_row, pre_pages_sections, pre_plan_row_id, pre_plan_row,
     new_name, new_function, applied_scope)
SELECT c.pc_id, c.page_id, c.site_id, c.domain, c.page_name,
       pc.component_id, pc.slot_name, pc.position,
       md5(pc.rendered_html), length(pc.rendered_html),
       to_jsonb(pc.*), p.sections, sps.id, to_jsonb(sps.*),
       c.new_name, c.new_function, current_setting('m701.scope')
  FROM census_700 c
  JOIN page_components pc ON pc.id = c.pc_id
  JOIN pages p ON p.id = c.page_id
  LEFT JOIN site_plan_sections sps
    ON c.has_plan_row
   AND sps.page_name = c.page_name AND sps.component_name = 'hero' AND sps.ordering = 0
   AND sps.plan_id = (SELECT sp.id FROM site_plans sp WHERE sp.site_id = c.site_id AND sp.is_current)
 WHERE c.is_target
ON CONFLICT (pc_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. THE CHANGE. One component per tool, built FROM the page's own live HTML.
-- Shape is 693's: component_level=tool, section_type NULL (the
-- refuse_selector_invisible_section trigger exempts tool-level rows),
-- render_mode=template, category=interactive, created_from=adopted,
-- suitable_* empty. forked_from comes from the census — RFC_036 §9.3's fork
-- for tool-equity-release, NULL for the other 21.
-- ---------------------------------------------------------------------------
INSERT INTO content_components
    (name, display_name, description, html_template, function, section_type,
     component_level, category, render_mode, created_from, is_active,
     forked_from, suitable_site_types, suitable_page_types)
SELECT c.new_name, c.display_name, c.description, pc.rendered_html,
       c.new_function, NULL, 'tool', 'interactive', 'template', 'adopted', true,
       c.fork_parent, '[]'::jsonb, '[]'::jsonb
  FROM census_700 c
  JOIN page_components pc ON pc.id = c.pc_id
 WHERE c.is_target;

-- Repoint each page_components row at its own component, and rename the slot
-- to the component's NAME (the alignment argued in the header: plan element =
-- slot = cc.name, moved together in this one transaction). Bytes, position,
-- content_data, digests: untouched.
UPDATE page_components pc
   SET component_id = cc.id,
       slot_name    = c.new_name,
       updated_at   = NOW()
  FROM census_700 c
  JOIN content_components cc ON cc.name = c.new_name
 WHERE pc.id = c.pc_id AND c.is_target;

-- Repoint the plan element in the single current plan (durability: sync_pages
-- rewrites pages.sections FROM this). Skipped for tool-ttk-calculator, which
-- has no plan row — Guard 2 pinned that.
UPDATE site_plan_sections sps
   SET component_name = c.new_name
  FROM census_700 c, site_plans sp
 WHERE c.is_target AND c.has_plan_row
   AND sp.site_id = c.site_id AND sp.is_current
   AND sps.plan_id = sp.id AND sps.page_name = c.page_name
   AND sps.component_name = 'hero' AND sps.ordering = 0;

-- Repoint the derived pages.sections copy (immediate effect: the writer path
-- consumes this). Order-preserving element swap; Guard 2 proved exactly one
-- 'hero' element per page.
UPDATE pages p
   SET sections = (SELECT jsonb_agg(CASE WHEN e.value = to_jsonb('hero'::text)
                                         THEN to_jsonb(c.new_name)
                                         ELSE e.value END ORDER BY e.ord)
                     FROM jsonb_array_elements(p.sections) WITH ORDINALITY e(value, ord)),
       updated_at = NOW()
  FROM census_700 c
 WHERE p.id = c.page_id AND c.is_target;

-- ---------------------------------------------------------------------------
-- 3. DRIVE THE REBUILD — 693's filing shape exactly, for the GENERIC targets
-- only (rebuild_policy=owned pages are refused at assemble; filing theirs
-- would mint guaranteed failures — deviation 8 in the header). idx_swi_dedup
-- is UNIQUE on (site_id, item_key) excluding terminal statuses; the guard
-- turns a collision with a LIVE row into a readable abort instead of a 23505.
-- [MEASURED 2026-09-02: the only live page_rerender rows on these sites are 12
-- 'deferred' 2026-08-03 rows on OTHER mortgagecalculator pages — no overlap.]
-- Burst note: scope=all files 16 at once (11 mortgagecalculator, 4
-- gamesdesign, 1 vetcomparison), scope=remainder 15 — within the live
-- producer's observed batch size (20+ in one 'rerender-pages' batch,
-- 2026-09-01, per 693's council round), and priority 80 matches it.
-- ---------------------------------------------------------------------------
DO $$
DECLARE bad text;
BEGIN
  SELECT string_agg(swi.item_key, ', ') INTO bad
    FROM site_work_items swi
    JOIN census_700 c ON c.site_id = swi.site_id
   AND swi.item_key = 'page_rerender_' || c.page_name || '_' || c.site_id || '_assemble'
   WHERE c.is_target AND c.policy_pinned = 'generic'
     AND swi.status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
  IF bad IS NOT NULL THEN
    RAISE EXCEPTION '701 ABORT: live page_rerender item(s) already queued for target page(s) — a rebuild is already pending, do not stack another: %', bad;
  END IF;
END $$;

INSERT INTO site_work_items
    (site_id, item_type, item_key, status, handler_agent, priority, source, summary,
     page_id, created_by, spec)
SELECT c.site_id,
       'page_rerender',
       'page_rerender_' || c.page_name || '_' || c.site_id || '_assemble',
       'triaged',
       'page-rerender',
       80,
       'bugfix_357_component_identity lane (migration 701)',
       'Rerender page: ' || c.page_name,
       c.page_id,
       'bugfix_357_component_identity lane (migration 701)',
       jsonb_build_object(
           'domain',    c.domain,
           'page_id',   c.page_id::text,
           'filename',  ltrim(p.url, '/'),
           'page_name', c.page_name,
           'reason',    'component identity repaired by migration 701 (bugs_open/357 Option B) — hero mislabel retyped to a per-tool adopted component')
  FROM census_700 c
  JOIN pages p ON p.id = c.page_id
 WHERE c.is_target AND c.policy_pinned = 'generic';

-- ---------------------------------------------------------------------------
-- 4. Decision of record, one doc_notes row per site per run (a pilot run and a
-- remainder run each leave their own dated row — deliberate: the trail records
-- what actually happened, when). Registering these rows at
-- component_level='tool' pulls the pages into tool health/acceptance sweeps
-- from now on, WHICH IS INTENDED — same stance as 693.
-- ---------------------------------------------------------------------------
INSERT INTO doc_notes (subject_type, subject_key, site_id, body, categories, source, created_by)
SELECT 'decision', c.domain, c.site_id,
       'TOOL ADOPTION OF RECORD (migration 701, bugs_open/357 phase 3 — owner Option B decision of 2026-09-02, scope='
       || current_setting('m701.scope') || '): ' || count(*) || ' page_components row(s) on ' || c.domain
       || ' claimed the shared hero component while storing a whole working tool ('
       || string_agg(c.page_name, ', ' ORDER BY c.page_name)
       || '). Each now has its own content_components row ADOPTED byte-for-byte from its stored rendered_html '
       || '(created_from=adopted, component_level=tool, zero template bindings — lossless by measurement), and the '
       || 'page row, the current plan element and pages.sections are repointed to the new name in one transaction, '
       || 'so regeneration is safe BY CONSTRUCTION: the declared template IS the stored bytes. No LLM regeneration. '
       || CASE WHEN bool_or(c.fork_parent IS NOT NULL)
               THEN 'tool-equity-release is born a FORK of library row a5236dec-c38e-46b5-8f05-4f7b43fa2f3f (RFC_036 s9.3: a library tool claims the function — fork, not insert). '
               ELSE '' END
       || 'These pages now fall under tool health/acceptance sweeps, intentionally. Sibling shape: lendzy migration 693 '
       || '(the NULL-component_id arm of the same family). Evidence: bugfix_357_component_identity lane NOTES 2026-09-02.',
       '["bugfix-357","tool-adoption","component-identity"]'::jsonb,
       'bugfix_357_component_identity lane',
       'bugfix_357_component_identity lane (migration 701)'
  FROM census_700 c
 WHERE c.is_target
 GROUP BY c.domain, c.site_id;

-- ---------------------------------------------------------------------------
-- VERIFY, as DO/RAISE. A verify block of bare SELECTs CANNOT stop the COMMIT —
-- ON_ERROR_STOP ignores a non-empty result set. This block can.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  scope        text := current_setting('m701.scope');
  n_targets    int; n_generic int; n_forks_expected int;
  comps        int; forks int; repointed int;
  plans_expected int; plans int; sections_ok int; queued int;
  repaired_total int; pop int; expected_pop int;
BEGIN
  SELECT count(*), count(*) FILTER (WHERE policy_pinned = 'generic'),
         count(*) FILTER (WHERE fork_parent IS NOT NULL),
         count(*) FILTER (WHERE has_plan_row)
    INTO n_targets, n_generic, n_forks_expected, plans_expected
    FROM census_700 WHERE is_target;

  -- Adopted components: right shape, right fork parentage, and the template
  -- BYTE-EQUAL to the bytes still stored on the page row.
  SELECT count(*), count(*) FILTER (WHERE cc.forked_from IS NOT NULL)
    INTO comps, forks
    FROM census_700 c
    JOIN content_components cc ON cc.name = c.new_name
    JOIN page_components pc ON pc.id = c.pc_id
   WHERE c.is_target
     AND cc.component_level = 'tool' AND cc.is_active AND cc.created_from = 'adopted'
     AND cc.function = c.new_function
     AND cc.forked_from IS NOT DISTINCT FROM c.fork_parent
     AND cc.html_template = pc.rendered_html;
  IF comps <> n_targets OR forks <> n_forks_expected THEN
    RAISE EXCEPTION '701 VERIFY: expected % adopted tool component(s) (% forked), found % (% forked)',
      n_targets, n_forks_expected, comps, forks;
  END IF;

  -- Repointed rows, bytes UNCHANGED against the pinned census.
  SELECT count(*) INTO repointed
    FROM census_700 c
    JOIN page_components pc ON pc.id = c.pc_id
    JOIN content_components cc ON cc.id = pc.component_id
   WHERE c.is_target
     AND cc.name = c.new_name AND pc.slot_name = c.new_name
     AND pc.position = 1 AND md5(pc.rendered_html) = c.md5_pinned;
  IF repointed <> n_targets THEN
    RAISE EXCEPTION '701 VERIFY: expected % repointed page_components row(s) with md5 unchanged, found %',
      n_targets, repointed;
  END IF;

  -- Plan elements repointed (the ttk exception excluded by has_plan_row).
  SELECT count(*) INTO plans
    FROM census_700 c
    JOIN site_plans sp ON sp.site_id = c.site_id AND sp.is_current
    JOIN site_plan_sections sps ON sps.plan_id = sp.id
   AND sps.page_name = c.page_name AND sps.ordering = 0
   WHERE c.is_target AND c.has_plan_row AND sps.component_name = c.new_name;
  IF plans <> plans_expected THEN
    RAISE EXCEPTION '701 VERIFY: expected % repointed plan element(s), found %', plans_expected, plans;
  END IF;

  -- pages.sections repointed: exactly one new-name element, zero 'hero'.
  SELECT count(*) INTO sections_ok
    FROM census_700 c JOIN pages p ON p.id = c.page_id
   WHERE c.is_target
     AND (SELECT count(*) FROM jsonb_array_elements(p.sections) e WHERE e.value = to_jsonb(c.new_name)) = 1
     AND (SELECT count(*) FROM jsonb_array_elements(p.sections) e WHERE e.value = to_jsonb('hero'::text)) = 0;
  IF sections_ok <> n_targets THEN
    RAISE EXCEPTION '701 VERIFY: expected % pages.sections repoint(s), found %', n_targets, sections_ok;
  END IF;

  -- The rebuilds must actually be queued for the generic targets, or the plan
  -- repoint goes unexercised and the pilot proves nothing.
  SELECT count(*) INTO queued
    FROM site_work_items swi
    JOIN census_700 c ON c.site_id = swi.site_id
   AND swi.item_key = 'page_rerender_' || c.page_name || '_' || c.site_id || '_assemble'
   WHERE c.is_target AND c.policy_pinned = 'generic'
     AND swi.item_type = 'page_rerender'
     AND swi.source = 'bugfix_357_component_identity lane (migration 701)'
     AND swi.status = 'triaged' AND swi.page_id IS NOT NULL;
  IF queued <> n_generic THEN
    RAISE EXCEPTION '701 VERIFY: expected % queued rerender(s), found %', n_generic, queued;
  END IF;

  -- The founding metric, re-run fleet-wide: rows bound to the shared hero
  -- whose bytes do not contain hero's own literal prefix.
  SELECT count(*) INTO repaired_total
    FROM census_700 c
    JOIN page_components pc ON pc.id = c.pc_id
    JOIN content_components cc ON cc.id = pc.component_id
   WHERE cc.name = c.new_name;
  expected_pop := 22 - repaired_total;

  SELECT count(*) INTO pop
    FROM page_components pc, content_components hero
   WHERE hero.id = '23f95f00-f293-466e-b43a-81791ea0fc6c'
     AND pc.component_id = hero.id
     AND position(left(hero.html_template, position('{{' in hero.html_template) - 1) in pc.rendered_html) = 0;
  IF pop <> expected_pop THEN
    RAISE EXCEPTION '701 VERIFY: population predicate returns %, expected % — a row moved mid-flight; investigate before retrying',
      pop, expected_pop;
  END IF;

  IF expected_pop = 0 THEN
    RAISE NOTICE '701 OK (scope=%): % component(s) adopted (% forked), % row(s) repointed with bytes unchanged, % plan element(s) + % pages.sections repointed, % rerender(s) queued. population = 0 — bugs_open/357''s closing metric at the DB layer; the bug closes at population = 0 verified at the ARTEFACTS, not at this transaction.',
      scope, comps, forks, repointed, plans, sections_ok, queued;
  ELSE
    RAISE NOTICE '701 OK (scope=%): % component(s) adopted (% forked), % row(s) repointed with bytes unchanged, % plan element(s) + % pages.sections repointed, % rerender(s) queued. population now = % (the pilot leaves the remainder pinned for scope=remainder).',
      scope, comps, forks, repointed, plans, sections_ok, queued, expected_pop;
  END IF;
END $$;

COMMIT;
