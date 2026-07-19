-- PATCH_feature_designer_018_reframe_reads_artifact.sql — finish 016-finding-2
-- on feature-designer: the VETO path was left blind. 2026-07-19. clients_db.
--
-- WHY (verified live, 2026-07-19, this thread):
-- PATCH_017 made the REVISE path roster-proof (repropose reads the
-- council_report artifact) but wired the load step onto that branch only:
--
--     run_checks -> load_council_report -> repropose        <- fixed
--     check_reframe -> reframe                              <- still per-seat
--
-- so `reframe` still names its seats in input_fields and still renders two
-- prompt sections:
--
--     reframe.input_fields = [spec_row, plan_persisted,
--                             review_editquality, review_guardian]
--     designer seats       = editquality, guardian, bug_historian,
--                            guidelines, reuse_agent
--
-- 2 of 5 seats. The reframe LLM cannot see bug_historian, guidelines or
-- reuse_agent objections, and will not see seat 6 either. This is the exact
-- signature the fix-proposer patch closed on its own side, which recorded:
-- "reframe gains eleven seats it never saw (it referenced only edit-quality
-- and guardian)". PATCH_017's header claim that the designer was "currently
-- complete (5/5/5)" was true of the revise path and not of this one.
--
-- THE FIX — mirror fix-proposer's placement rather than invent a second shape.
-- On fix-proposer the load step sits BEFORE the branch, so every downstream
-- path inherits it:
--
--     council_decide -> load_council_reviews -> check_approved
--
-- Doing the same here fixes reframe without a second query step, and keeps the
-- two agents' council plumbing the same shape (the drift class 099 exists for).
-- The approved path then also runs one extra read-only query before `complete`
-- — harmless, and exactly what fix-proposer already does.
--
--     council_decide -> load_council_report -> check_approved   (new)
--     run_checks     -> repropose                               (restored)
--
-- repropose is UNCHANGED: council_report_row is populated earlier in the run
-- and collected_data carries it across the branch, so it still renders.
--
-- Correlation param verified, not assumed: 0NN_TRIGGER_feature_designer_v1.sh
-- line 31 sets input_data.fix_correlation_id (the feature correlation reuses
-- the fix-loop field name), which is what load_council_report already keys on.
--
-- SURGICAL (the co-edited-row rule): jsonb_set on specific paths only.
-- IDEMPOTENT: every write is a fixed value, so a second apply is a no-op.
-- Do not apply mid-run.

BEGIN;

SELECT snapshot_agent('feature-designer', 'pre-update: 018 — reframe reads the council_report artifact; load step moved ahead of the branch so both reviser paths inherit it');

-- 1. Move the load step ahead of the branch: council_decide -> load_council_report.
UPDATE agent_definitions SET default_config = jsonb_set(
  default_config, '{workflow,steps,council_decide,next_step}', '"load_council_report"')
WHERE type='feature-designer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 2. ...and on to the first router, so approve/reject/revise all pass through it.
UPDATE agent_definitions SET default_config = jsonb_set(
  default_config, '{workflow,steps,load_council_report,next_step}', '"check_approved"')
WHERE type='feature-designer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 3. Restore the revise path's own routing (017 had pointed it at the load step).
UPDATE agent_definitions SET default_config = jsonb_set(
  default_config, '{workflow,steps,run_checks,next_step}', '"repropose"')
WHERE type='feature-designer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 4. Describe the step for what it now is — both reviser paths, not just revise.
UPDATE agent_definitions SET default_config = jsonb_set(
  default_config, '{workflow,steps,load_council_report,description}',
  '"Roster-proof reviser input: the full council_report artifact (every seat that voted). Runs before the routers so BOTH revise (repropose) and veto (reframe) inherit it; adding a seat needs no prompt edit."')
WHERE type='feature-designer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 5. reframe: the two per-seat fields become the one artifact field.
UPDATE agent_definitions SET default_config = jsonb_set(
  default_config, '{workflow,steps,reframe,config,input_fields}',
  '["spec_row","plan_persisted","council_report_row"]'::jsonb)
WHERE type='feature-designer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 6. reframe prompt: one artifact section replaces the edit-quality/guardian pair.
--    Guardian-veto framing is preserved — the artifact names verdicts per seat,
--    so the prompt still says whose veto is the hard one.
UPDATE agent_definitions SET default_config = jsonb_set(
  default_config, '{workflow,steps,reframe,config,prompt_template}',
  '"# PROMPT — REFRAME after a council VETO\n\nThe council REJECTED your staged plan outright — the guardian judged it architecture-level change (or live-state mutation) dressed as a contained build. Do NOT resubmit the same shape: it will be vetoed again. Produce ONE of:\n\n(a) a STRICTLY NARROWER staged plan the guardian could accept — prefer the reviewer''s own recommended alternative if its review names one; fewer stages, smaller blast radius, every DB-side artifact a seed file; or\n\n(b) if no contained build exists at all, a plan whose only stages are the minimal safe preparatory steps, with risks stating plainly which decision the architecture review must take.\n\nSame staged-v1 schema and remaining hard rules: only code_pointers paths or files this plan adds; caps; grounded in the spec; every edit CHANGES something. CHECKLIST ACTS ARE A CLOSED SET: image_deploy | seed_apply | verify — never invent a new act name; a pre-apply confirmation is a verify entry ordered before seed_apply; image_deploy only when the plan ships code edits.\n\n## The approved spec\n{{.spec_row.summary}}\n{{.spec_row.spec_text}}\n\n## Your VETOED plan\n{{.plan_persisted.plan_json}}\n\n## The council''s reviews — EVERY seat that voted this round, verbatim\n{{.council_report_row.body}}\n\nThis is the council_report artifact: a JSON object with `decision`, `decided_by` and `reviews[]`, one entry per seat ({reviewer, verdict, objections[], missing[], notes}). Address EVERY objection from EVERY seat, not only the ones you find familiar — the roster grows, and a seat you have not seen before is not advisory noise. The guardian holds the only hard veto (read its notes for the recommended alternative); the rest are advisory, but an unaddressed advisory objection is what sends this round back again.\n\n## Output — ONLY the staged-v1 plan JSON."'::jsonb)
WHERE type='feature-designer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

COMMIT;
