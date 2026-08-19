-- 493 — give every loop-nested create_work_item a per-item key suffix, so a
--       run's second finding stops silently deleting itself (bugs_open/321)
--
-- WHY. create_work_item builds item_key as '<item_key_prefix>_<domain>' — one key
-- per SITE (create_work_item_action.go:225-234). idx_swi_dedup is UNIQUE
-- (site_id, item_key) over non-terminal rows. A step that files N items per site
-- inside a loop therefore loses iterations 2..N to the collision, silently: the
-- loop reports success either way, and the evidence is split across llm_call_log
-- (the suggestions) and site_work_items (the items), so neither store alone shows
-- the gap. Measured on tool-suggester over its full history: 40 suggestions in
-- eight runs -> 11 work items, ~72% lost; the 2026-08-19 10:25 run alone was
-- 7 -> 1. It is a RACE, not a designed cap — a colliding key only blocks while
-- the earlier item is non-terminal, which is why the rate varies run to run.
--
-- The remedy exists in the same function and is simply unused: the optional
-- item_key_suffix_field config key (a dotted CollectedData path appended to the
-- key; configured-but-unresolved is a deliberate HARD ERROR — the fallback to the
-- site-wide key would silently reinstate exactly this bug, per the council rounds
-- recorded at create_work_item_action.go:236-259). tool-auditor's two loop steps
-- already use it (tool_data.page_id) — the proven idiom this migration copies.
--
-- SCOPE — the CLASS, not the instance. Fleet census 2026-08-19: of 20 live
-- create_work_item steps carrying item_key_prefix, exactly 6 sit inside a loop
-- (per-item filing; everywhere else the site-wide key is the intended dedupe).
-- Two already carry the suffix. This migration fixes the other FOUR, on three
-- agents:
--
--   tool-suggester.create_items_loop.create_novel_item    +suffix current_suggestion.function
--   tool-suggester.create_items_loop.create_library_item  +suffix current_suggestion.function
--   component-quality-auditor.create_regen_items.create_work_item
--                                                         +suffix current_component.component_id
--   internal-linker.create_items_loop.create_rewrite_item +suffix current_link.source_page
--
-- ...plus ONE flag: continue_on_error: true on tool-suggester's create_items_loop.
--
-- WHY THE FLAG SHIPS IN THE SAME MIGRATION (do not split them): the suffix's
-- unresolved-path hard error, alone, would abort the WHOLE expanded loop — the
-- loop is the only one of the four without the flag, none of its substeps carries
-- an error_step, and routeToErrorStepOrFail then falls to failWorkflow
-- (coordinator.go:3667-3681). One malformed suggestion -> 0 items, strictly worse
-- than today's partial loss. With the flag, a bad iteration is skipped AND
-- durably recorded: skipToNextLoopIteration persists {loop}_iter_{N}_error into
-- collected_data (loop_error_handler.go:141-149,185-188) and the aggregation
-- folds status:"error" into items_created (loop_actions.go:505-511), an output
-- field of the workflow. This also converges the one divergent loop onto the
-- estate norm — the other three loops in this class all declare it.
--
-- WHY EACH SUFFIX PATH IS SAFE (unresolved = hard error, so this was measured,
-- not assumed):
--   * current_suggestion.function — over ALL 60 historical tool-suggester
--     answers / 239 suggestions: 239 non-empty, 0 intra-answer duplicates.
--   * current_component.component_id and current_link.source_page — the SAME
--     path is already a hard-required spec_paths entry in the SAME step (an
--     unresolved spec_paths entry refuses the work item, same function, :286-300).
--     So a suffix that fails to resolve describes an iteration that already
--     hard-fails today: zero new failure modes. The pre-gate below asserts this
--     premise rather than trusting it.
--   component_id over function for CQA: it is the unique id; two components can
--   share a function. source_page for the linker: page-level granularity is
--   DELIBERATE — two links into one page dedupe to one rewrite item (two
--   concurrent rewrites of one page would clobber each other).
--
-- The two "latent" agents have never filed an item — but internal-linker's loop
-- has been dead only because of the broken conditional migration 490 (the
-- bugs_open/313 lane) fixes, so its collision goes LIVE when 490 lands. Sizing
-- from that lane: plan_links asks for 1-3 links per plan, so the collision costs
-- 1-2 of 3 — up to ~2/3 of that agent's entire output, from its first productive
-- day. Coordinated with the 313 session 2026-08-19 (proceed independently;
-- 490's gates and writes never touch the create_items_loop subtree, mine never
-- touch theirs, so apply order between 490 and 493 does not matter).
--
-- KNOWN, ACCEPTED consequences (measured before applying):
--   * Re-runs can now repeat a key. insertWorkItem's two-strike block brands the
--     THIRD item on a repeated (site_id, item_key) 'unresolved' — only 4 of 214
--     historical (site,function) pairs ever reached a third suggestion, and a
--     tool that failed to build twice SHOULD stop being retried.
--     recurrence_expected deliberately NOT set.
--   * Steady-state duplicate-rebuild waste ~10% (25/239 suggestions repeat a
--     function on their site across runs); the component layer is idempotent on
--     function (create_tool_component_action.go:225-246), so no duplicate pages.
--   * Build volume per suggester run rises from ~1 to up-to-8 items. Owner
--     decision 2026-08-19: no throttle; volume is bounded upstream (prompt caps
--     the answer at 8, loop max_iterations 10, dispatch max_items 5/pass, and
--     add_tool priority 120/130 sorts behind default-100 work).
--
-- Scoped by id, pre-state gated (subtree md5 + literal anchors), DO/RAISE with
-- ROW_COUNT asserts, post-verify re-reads INCLUDING non-clobber controls for the
-- concurrent 484 edits, snapshot first, rollback sidecar. Config-only, LIVE ON
-- APPLY, no roll needed. Not council-submitted: agent_definitions config is
-- outside the gate's platform/ internal/ pkg/ scope (484's recorded precedent);
-- the Go halves of this lane (detector + Warn) ARE submitted, separately.
--
-- NULL-direction analysis (asked of every migration since the 184 rounds): every
-- gate below is a positive-presence assertion (IS DISTINCT FROM the expected
-- literal), so a NULL — wrong path, renamed step, moved subtree — trips the
-- RAISE; no comparison can go silently green on absent data.
--
-- ROLLBACK: 493_loop_nested_item_key_suffixes_ROLLBACK.sql
--
-- Verify after apply (expect: 4 suffixes, 1 flag, 484 intact):
--   see the DO block's post-verify, and RUNBOOK_item_key_collisions.md for the
--   run-level disconfirming check (an N-suggestion answer must produce N items).

BEGIN;

SELECT snapshot_agent('tool-suggester',
  '493_loop_nested_item_key_suffixes: pre-update');
SELECT snapshot_agent('component-quality-auditor',
  '493_loop_nested_item_key_suffixes: pre-update');
SELECT snapshot_agent('internal-linker',
  '493_loop_nested_item_key_suffixes: pre-update');

DO $$
DECLARE
  ts_id  uuid;
  cqa_id uuid;
  il_id  uuid;
  n      int;
  v      text;
BEGIN
  ----------------------------------------------------------------------------
  -- PRE-GATE. Abort (rolling back the snapshots too) unless the live rows are
  -- exactly the shape this file was written against.
  ----------------------------------------------------------------------------

  -- Exactly one live row per type, ids pinned.
  SELECT id INTO STRICT ts_id  FROM agent_definitions
   WHERE type='tool-suggester' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  SELECT id INTO STRICT cqa_id FROM agent_definitions
   WHERE type='component-quality-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  SELECT id INTO STRICT il_id  FROM agent_definitions
   WHERE type='internal-linker' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF ts_id  <> 'c0756913-04b1-489d-86b4-9ec249dc804d' THEN
    RAISE EXCEPTION '493 pre-gate: tool-suggester live row id changed (%). Re-read the row before applying.', ts_id;
  END IF;
  IF cqa_id <> '9da24f98-4c57-493a-a15f-2d2bd35c3cf2' THEN
    RAISE EXCEPTION '493 pre-gate: component-quality-auditor live row id changed (%).', cqa_id;
  END IF;
  IF il_id  <> '93cffe67-baf4-4fb1-bec9-ba546fb24a54' THEN
    RAISE EXCEPTION '493 pre-gate: internal-linker live row id changed (%).', il_id;
  END IF;

  -- tool-suggester: the exact subtree this file rewrites, fingerprinted. Fails
  -- loudly iff ANYONE touched create_items_loop since this file was written,
  -- while leaving every other subtree (484's suggest_tools edits, 490's linker
  -- edits) free to move.
  SELECT md5(default_config #>> '{workflow,steps,create_items_loop}') INTO v
    FROM agent_definitions WHERE id = ts_id;
  IF v IS DISTINCT FROM '68889d441446cb03a9dec2968919eb3b' THEN
    RAISE EXCEPTION '493 pre-gate: tool-suggester create_items_loop subtree changed (md5 %). Re-derive the anchors; do NOT force.', v;
  END IF;

  -- component-quality-auditor: the premise that makes the suffix risk-free — the
  -- identical path is already hard-required via spec_paths — asserted, not trusted.
  SELECT default_config #>> '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config,spec_paths,component_id}'
    INTO v FROM agent_definitions WHERE id = cqa_id;
  IF v IS DISTINCT FROM 'current_component.component_id' THEN
    RAISE EXCEPTION '493 pre-gate: CQA spec_paths.component_id is % — the zero-new-failure-modes argument no longer holds; re-verify before applying.', COALESCE(v,'<absent>');
  END IF;
  SELECT default_config #>> '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config,item_key_suffix_field}'
    INTO v FROM agent_definitions WHERE id = cqa_id;
  IF v IS NOT NULL THEN
    RAISE EXCEPTION '493 pre-gate: CQA suffix already set (%) — double-apply, or another session got here first.', v;
  END IF;
  SELECT default_config #>> '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config,item_key_prefix}'
    INTO v FROM agent_definitions WHERE id = cqa_id;
  IF v IS DISTINCT FROM 'quality_regen' THEN
    RAISE EXCEPTION '493 pre-gate: CQA item_key_prefix is % (expected quality_regen).', COALESCE(v,'<absent>');
  END IF;

  -- internal-linker: same three, on its own subtree only (490 edits elsewhere).
  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config,spec_paths,page_name}'
    INTO v FROM agent_definitions WHERE id = il_id;
  IF v IS DISTINCT FROM 'current_link.source_page' THEN
    RAISE EXCEPTION '493 pre-gate: linker spec_paths.page_name is % — the zero-new-failure-modes argument no longer holds.', COALESCE(v,'<absent>');
  END IF;
  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config,item_key_suffix_field}'
    INTO v FROM agent_definitions WHERE id = il_id;
  IF v IS NOT NULL THEN
    RAISE EXCEPTION '493 pre-gate: linker suffix already set (%).', v;
  END IF;
  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config,item_key_prefix}'
    INTO v FROM agent_definitions WHERE id = il_id;
  IF v IS DISTINCT FROM 'internal_link' THEN
    RAISE EXCEPTION '493 pre-gate: linker item_key_prefix is % (expected internal_link).', COALESCE(v,'<absent>');
  END IF;

  ----------------------------------------------------------------------------
  -- UPDATES. id-scoped, leaf-path jsonb_set only; ROW_COUNT asserted.
  ----------------------------------------------------------------------------

  UPDATE agent_definitions SET
    default_config = jsonb_set(jsonb_set(jsonb_set(default_config,
      '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_novel_item,config,item_key_suffix_field}',
      '"current_suggestion.function"'),
      '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_library_item,config,item_key_suffix_field}',
      '"current_suggestion.function"'),
      '{workflow,steps,create_items_loop,config,continue_on_error}',
      'true'),
    updated_at = now()
  WHERE id = ts_id;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '493: tool-suggester UPDATE touched % rows', n; END IF;

  UPDATE agent_definitions SET
    default_config = jsonb_set(default_config,
      '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config,item_key_suffix_field}',
      '"current_component.component_id"'),
    updated_at = now()
  WHERE id = cqa_id;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '493: CQA UPDATE touched % rows', n; END IF;

  UPDATE agent_definitions SET
    default_config = jsonb_set(default_config,
      '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config,item_key_suffix_field}',
      '"current_link.source_page"'),
    updated_at = now()
  WHERE id = il_id;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '493: internal-linker UPDATE touched % rows', n; END IF;

  ----------------------------------------------------------------------------
  -- POST-VERIFY. Re-read every written leaf, plus non-clobber controls
  -- asserting the concurrent 484 edits survived this transaction untouched.
  ----------------------------------------------------------------------------

  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_novel_item,config,item_key_suffix_field}'
    INTO v FROM agent_definitions WHERE id = ts_id;
  IF v IS DISTINCT FROM 'current_suggestion.function' THEN
    RAISE EXCEPTION '493 post-verify: novel suffix reads %', COALESCE(v,'<absent>');
  END IF;
  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_library_item,config,item_key_suffix_field}'
    INTO v FROM agent_definitions WHERE id = ts_id;
  IF v IS DISTINCT FROM 'current_suggestion.function' THEN
    RAISE EXCEPTION '493 post-verify: library suffix reads %', COALESCE(v,'<absent>');
  END IF;
  SELECT default_config #>> '{workflow,steps,create_items_loop,config,continue_on_error}'
    INTO v FROM agent_definitions WHERE id = ts_id;
  IF v IS DISTINCT FROM 'true' THEN
    RAISE EXCEPTION '493 post-verify: continue_on_error reads %', COALESCE(v,'<absent>');
  END IF;
  -- non-clobber controls: 484's edits must still be present, byte-identical needles
  SELECT default_config #>> '{workflow,steps,suggest_tools,config,ai_service,max_tokens}'
    INTO v FROM agent_definitions WHERE id = ts_id;
  IF v IS DISTINCT FROM '6000' THEN
    RAISE EXCEPTION '493 post-verify NON-CLOBBER: suggest_tools max_tokens reads % (484 expected 6000)', COALESCE(v,'<absent>');
  END IF;
  IF (SELECT default_config #>> '{workflow,steps,suggest_tools,config,prompt_template}'
        FROM agent_definitions WHERE id = ts_id) NOT LIKE '%HARD LIMITS ON THE ANSWER%' THEN
    RAISE EXCEPTION '493 post-verify NON-CLOBBER: 484 HARD LIMITS block missing from suggest_tools prompt';
  END IF;

  SELECT default_config #>> '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config,item_key_suffix_field}'
    INTO v FROM agent_definitions WHERE id = cqa_id;
  IF v IS DISTINCT FROM 'current_component.component_id' THEN
    RAISE EXCEPTION '493 post-verify: CQA suffix reads %', COALESCE(v,'<absent>');
  END IF;

  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config,item_key_suffix_field}'
    INTO v FROM agent_definitions WHERE id = il_id;
  IF v IS DISTINCT FROM 'current_link.source_page' THEN
    RAISE EXCEPTION '493 post-verify: linker suffix reads %', COALESCE(v,'<absent>');
  END IF;

  RAISE NOTICE 'OK 493: 4 suffixes set, continue_on_error set, 484 controls intact';
END $$;

COMMIT;
