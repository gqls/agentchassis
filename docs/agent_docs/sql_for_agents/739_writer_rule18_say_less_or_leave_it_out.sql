-- 739_writer_rule18_say_less_or_leave_it_out.sql
--
-- OWNER RULING 2026-09-03: when the writer has nothing verified to say, the fallback is
-- "say less (but keep it honest and user helpfulness focused) or leave it out" — NOT
-- "be general". His words, in the session that raised it: "We don't want vacuous content,
-- absolutely."
--
-- WHY THIS IS THE ONE LINE THAT MATTERED. `[MEASURED 2026-09-03]` 33 of 60 live sites carry
-- NO evidence_base at all, so on the majority of the estate the writer reaches this rule with
-- nothing verified about the business. Rule 18 then told it, in terms, that GENERAL was the
-- preferred answer — and general is exactly the register the owner reads as AI-written. The
-- planner already declines a too-thin PAGE, his 2026-08-25 ruling already declines the
-- values/testimonial SECTION, and the field spec already licenses "" for an optional field.
-- The writer was the one rung with no way out: a planned section had to be filled, and rule 18
-- named the filling.
--
-- ⚠ THE COUPLING, and it is why this migration touches TWO rules. Rule 19 RESTATES rule 18
-- ("keep preferring a general truth to a specific invention"). Change 18 alone and 19 keeps
-- instructing the exact behaviour 18 now forbids, in the same numbered list, four lines later.
-- Both anchors are therefore required to hit exactly once or this aborts.
--
-- NOT a licence to fabricate and NOT a licence to strip pages: "say less" is bounded by
-- usefulness to the reader, which is the half of his ruling a terse reading would drop.
-- The completeness/shrink/component floors are untouched by this migration and still refuse a
-- save that DROPS slots — so this changes what the writer says inside a section, not whether
-- the section exists. Making a declined section representable is separate work, not this file.

BEGIN;

SELECT snapshot_agent('page-content-writer',
  '739_writer_rule18_say_less_or_leave_it_out.sql: pre-ruling (rule 18 preferred "general")');

DO $m$
DECLARE cfg jsonb; tpl text; n int;
  old18 text := '18. It is ALWAYS better to be honest and general than specific and fabricated. A real visitor will trust a general statement of capability more than a fabricated testimonial.';
  new18 text := '18. When you have nothing verified to say, SAY LESS or LEAVE IT OUT. Do not pad with general statements of capability: a paragraph that would read the same on any company in this industry tells the reader nothing and is the thing to avoid. Never fabricate to fill a gap either — the choice is not between vague and invented. One short sentence that is true and useful to the person reading beats five that are true of anyone, and an omitted optional field beats a filled one carrying no information. Judge every sentence by what the reader gets from it.';
  old19 text := 'This does NOT soften rule 18: keep BEING straight with the reader, and keep preferring a general truth to a specific invention - just never LABEL it.';
  new19 text := 'This does NOT soften rule 18: keep BEING straight with the reader, and where you have nothing to say prefer saying less to saying something general - just never LABEL it.';
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN RAISE EXCEPTION '739: % active page-content-writer rows, want exactly 1 — duplicate-row landmine; find which row the runtime loads before touching either', n; END IF;

  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  tpl := cfg #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}';
  IF tpl IS NULL THEN RAISE EXCEPTION '739: generate_content has no prompt_template at the sub_workflow path — the loop shape changed, re-base'; END IF;

  -- Both anchors exactly once, or abort. A prompt this size accumulates near-duplicates;
  -- an anchor that matched twice would rewrite a passage nobody reviewed.
  IF (length(tpl) - length(replace(tpl, old18, ''))) / length(old18) <> 1 THEN
    RAISE EXCEPTION '739: rule 18 anchor not-exactly-once — template drifted, re-base';
  END IF;
  IF (length(tpl) - length(replace(tpl, old19, ''))) / length(old19) <> 1 THEN
    RAISE EXCEPTION '739: rule 19 cross-reference anchor not-exactly-once — template drifted, re-base';
  END IF;

  tpl := replace(tpl, old18, new18);
  tpl := replace(tpl, old19, new19);

  cfg := jsonb_set(cfg, '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}', to_jsonb(tpl));
  UPDATE agent_definitions SET default_config = cfg
   WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '739: update touched % rows, want 1', n; END IF;

  -- VERIFY on the loaded row, and RAISE rather than SELECT: a verify block of SELECTs cannot
  -- stop the COMMIT (ON_ERROR_STOP ignores a non-empty result) — the estate's own landmine.
  SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO tpl FROM agent_definitions
   WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position(new18 IN tpl) = 0 THEN RAISE EXCEPTION '739 VERIFY: new rule 18 absent from the loaded row'; END IF;
  IF position(new19 IN tpl) = 0 THEN RAISE EXCEPTION '739 VERIFY: new rule 19 absent from the loaded row'; END IF;
  IF position(old18 IN tpl) > 0 THEN RAISE EXCEPTION '739 VERIFY: old rule 18 still present'; END IF;
  IF position('preferring a general truth to a specific invention' IN tpl) > 0 THEN
    RAISE EXCEPTION '739 VERIFY: rule 19 still restates the repealed preference';
  END IF;
  RAISE NOTICE '739 verify: rules 18 and 19 both carry the say-less ruling on the single loaded row.';
END $m$;

INSERT INTO schema_migrations (filename, checksum, applied_by, notes)
VALUES ('739_writer_rule18_say_less_or_leave_it_out.sql', :'mig_checksum', 'copy_quality_two_stage session',
        'Owner ruling 2026-09-03: writer fallback is "say less or leave it out", not "be general". Rule 19 updated in the SAME migration because it restates rule 18.');

COMMIT;
