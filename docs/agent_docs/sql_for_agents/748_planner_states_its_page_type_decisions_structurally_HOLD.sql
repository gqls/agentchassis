-- 748_planner_states_its_page_type_decisions_structurally_HOLD.sql
--
-- bugs_open/428, phase 2. Adds ONE optional field to build-site-planner's
-- plan_site output — `page_type_decisions` — so the planner's reason for not
-- planning a recommended page_type becomes machine-readable, and specifically so
-- that a deferral naming a PRODUCER can be checked against whether that producer
-- is running.
--
-- ⚠ _HOLD: DO NOT APPLY UNTIL A CHASSIS CARRYING THE READER HAS ROLLED.
-- The reader is reconcileRecommendedPageTypes
-- (platform/orchestration/actions/recommended_type_reconciliation.go, commits
-- eee40b554 + 91173c6d7, council corr ee84dfb6-2574-4744-8f06-18eed91fea49).
-- Applying this early would ask the planner to emit a field nothing reads —
-- which is, precisely, the defect bugs_open/428 is about (migration 687 obliged
-- the planner to justify omissions in `strategy_notes`, and NOTHING READS
-- strategy_notes), reproduced one migration along. The reader degrades
-- gracefully without this field, so there is no rush and no ordering risk in
-- waiting: absent the field it falls back to a substring read of
-- strategy_notes, records reason_source='strategy_notes', and every other arm of
-- the check works unchanged.
--
-- WHAT IT DOES NOT DO, deliberately. It does not ask the planner to say whether
-- a type is PRESENT, and the reader does not believe it if it does: presence is
-- computed from `pages` in Go, every time. That split is the whole design — the
-- model states intent, the framework verifies the fact. It matters because the
-- motivating case is a planner asserting presence falsely while complying with
-- 687: gamedesign.uk, call 7b3bffdd-64dc-4a97-bb00-7633aa7271f8, 2026-09-03
-- 10:40:15Z, wrote "All four types are present" on a plan carrying zero
-- blog-post and zero blog-index pages. A field the reader trusted would have
-- carried that falsehood straight through.
--
-- It also does not touch 687's own sentence's REQUIREMENT, only points at the
-- new field beside it: the licensed final say stays exactly as 687 left it.
--
-- SHARED ROW, FOUR LANES. build-site-planner's prompt_template is edited by this
-- lane, the 444 listing lane (720), the 450 tool lane (729, unapplied) and the
-- designblog lane (730/731). Following 729's precedent, the verify block below
-- refuses if any neighbour's surface has gone missing, so a migration that
-- silently ate another lane's sentence cannot look like a clean apply.
-- ⚠ While this is applied, 687's own text is unchanged but the JSON block is
-- longer; any ROLLBACK of a neighbour that anchors on the block must be checked.
--
-- Apply: psql -f THIS FILE ONLY. Then record it:
--   ./scripts/migration/run-migrations.sh --record-only <this file> --note "..."

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '748 REFUSED: expected exactly 1 active build-site-planner row, found %', n;
  END IF;
  PERFORM snapshot_agent('build-site-planner', '748_planner_states_its_page_type_decisions_structurally_HOLD.sql: pre-update');
END $$;

DO $do$
DECLARE
  tpl text; newtpl text; n int;
  ifs int; ends int; elses int; vars int;
  anchor_a text := $A748$  "strategy_notes": "any notes on how you used or diverged from the strategy/roadmap",$A748$;
  repl_a   text := $R748$  "strategy_notes": "any notes on how you used or diverged from the strategy/roadmap",
  "page_type_decisions": [
    {"page_type": "one of the recommended_page_types you did NOT include in pages", "decision": "deferred|rejected|substituted", "reason": "the real, specific reason for THIS type", "deferred_to": "the name of the mechanism you expect to supply it later, or omit this key if there is none"}
  ],$R748$;
  anchor_b text := $B748$An omitted named type with no per-type reason in strategy_notes is a gap, not a decision.$B748$;
  repl_b   text := $S748$An omitted named type with no per-type reason in strategy_notes is a gap, not a decision. Record the same decisions in the `page_type_decisions` array as well, one entry per omitted type: prose is for a human, that array is what the build system reads. If your reason is that something else will supply the pages later, name that mechanism in `deferred_to` — it is checked against whether that mechanism is actually running, so a deferral to something dormant is caught rather than believed. Do NOT use these entries to claim a type is present: whether a type is in the plan is read from `pages`, never from what you say about it.$S748$;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '748: prompt_template not found'; END IF;

  IF position('page_type_decisions' in tpl) > 0 THEN
    RAISE EXCEPTION '748: already applied — page_type_decisions present';
  END IF;

  n := (length(tpl) - length(replace(tpl, anchor_a, ''))) / length(anchor_a);
  IF n <> 1 THEN RAISE EXCEPTION '748: strategy_notes JSON anchor found % times, expected 1', n; END IF;
  n := (length(tpl) - length(replace(tpl, anchor_b, ''))) / length(anchor_b);
  IF n <> 1 THEN RAISE EXCEPTION '748: 687 gap-not-a-decision anchor found % times, expected 1', n; END IF;

  -- Template-language balance: this migration adds PROSE and JSON only, so every
  -- Go template construct must survive untouched. A migration on this row that
  -- unbalanced {{if}}/{{end}} would break every plan on the estate.
  ifs   := (length(tpl) - length(replace(tpl, '{{if ',   ''))) / length('{{if ');
  ends  := (length(tpl) - length(replace(tpl, '{{end}}', ''))) / length('{{end}}');
  elses := (length(tpl) - length(replace(tpl, '{{else}}',''))) / length('{{else}}');
  vars  := (length(tpl) - length(replace(tpl, '{{.',     ''))) / length('{{.');

  newtpl := replace(replace(tpl, anchor_a, repl_a), anchor_b, repl_b);

  IF length(newtpl) <> length(tpl)
       + (length(repl_a) - length(anchor_a))
       + (length(repl_b) - length(anchor_b)) THEN
    RAISE EXCEPTION '748: unexpected length delta — an anchor matched more than intended';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{.',     ''))) / length('{{.')      <> vars
     OR (length(newtpl) - length(replace(newtpl, '{{if ',   ''))) / length('{{if ')   <> ifs
     OR (length(newtpl) - length(replace(newtpl, '{{end}}', ''))) / length('{{end}}') <> ends
     OR (length(newtpl) - length(replace(newtpl, '{{else}}',''))) / length('{{else}}')<> elses THEN
    RAISE EXCEPTION '748: template variable or if/else/end balance changed';
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(newtpl), false),
         updated_at = now()
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '748: updated % rows, expected exactly 1', n; END IF;
END $do$;

-- Verify: this migration's own two surfaces, AND every neighbouring lane's, so
-- that eating another lane's sentence cannot read as a clean apply (729's rule).
DO $$
DECLARE tpl text; n int;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  n := (length(tpl) - length(replace(tpl, 'page_type_decisions', ''))) / length('page_type_decisions');
  IF n <> 2 THEN RAISE EXCEPTION '748 VERIFY: page_type_decisions present % times, expected 2 (JSON block + rule text)', n; END IF;
  IF position('it is checked against whether that mechanism is actually running' in tpl) = 0 THEN
    RAISE EXCEPTION '748 VERIFY: the deferred_to check sentence is missing';
  END IF;

  -- Neighbours (each must survive unchanged):
  IF position('name that exact page_type in strategy_notes' in tpl) = 0 THEN
    RAISE EXCEPTION '748 VERIFY: migration 687 (this lane, per-type reason) damaged';
  END IF;
  IF position('Validation holds back listing pages whose item source resolves to zero' in tpl) = 0 THEN
    RAISE EXCEPTION '748 VERIFY: migration 720 (bugs_open/444 listing gate rule 3) damaged';
  END IF;
  IF position('NO LATER EDITORIAL PASS RUNS' in tpl) = 0 THEN
    RAISE EXCEPTION '748 VERIFY: migrations 730/731 (designblog lane, rule 20) damaged';
  END IF;
  IF position('apply the Directory rule above' in tpl) = 0 THEN
    RAISE EXCEPTION '748 VERIFY: migration 433 (directory rule 18) damaged';
  END IF;
  IF position('Imagery Block rules above' in tpl) = 0 THEN
    RAISE EXCEPTION '748 VERIFY: migration 718 (imagery surface, rule 13) damaged';
  END IF;

  RAISE NOTICE '748 OK: the planner now states its page-type decisions structurally, and every neighbour surface is intact.';
END $$;

COMMIT;
