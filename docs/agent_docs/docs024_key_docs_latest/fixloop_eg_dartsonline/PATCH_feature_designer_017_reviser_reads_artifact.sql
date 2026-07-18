-- PATCH_feature_designer_017_reviser_reads_artifact.sql — make the reviser
-- roster-proof. 2026-07-18, feature-builder thread. Applies to clients_db.
--
-- WHY (reasoning-dataset thread's second 016 finding, relayed 2026-07-18):
-- threading each seat through repropose's input_fields + its own prompt section
-- is NOT idempotent under roster growth. On fix-proposer it has already failed:
-- 13 seats seeded, 6 referenced, so a revise round cannot see 54 percent of the
-- council's objections — and it silently recurs on seat 14. feature-designer is
-- currently complete (5/5/5) but is one seat from the same bug.
--
-- THE FIX: the reviser reads the council_report ARTIFACT once. council_decide
-- already persists every seat named in its review_fields into that artifact, so
-- adding a seat needs no prompt edit — and review_fields is self-enforcing (a
-- seat missing there already fails loudly by not counting toward the decision).
-- Three lists that must agree become one.
--
-- SURGICAL (the co-edited-row rule): jsonb_set on specific paths only.
-- Applied AFTER run 8e837814 completed — never mid-run.

BEGIN;

SELECT snapshot_agent('feature-designer', 'pre-update: 017 — reviser reads the council_report artifact instead of per-seat prompt sections');

UPDATE agent_definitions SET default_config = jsonb_set(
  default_config, '{workflow,steps,load_council_report}',
  jsonb_build_object(
    'action','query_database',
    'description','Roster-proof revise input: the full council_report artifact (every seat that voted), replacing per-seat prompt sections.',
    'output_field','council_report_row',
    'next_step','repropose',
    'config', jsonb_build_object(
      'output_format','object',
      'params', jsonb_build_array('input_data.fix_correlation_id'),
      'query','SELECT body FROM diagnosis_artifacts WHERE correlation_id = $1 AND kind = ''council_report'' ORDER BY created_at DESC LIMIT 1'
    )
  ), true)
WHERE type='feature-designer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions SET default_config = jsonb_set(
  default_config, '{workflow,steps,run_checks,next_step}', '"load_council_report"')
WHERE type='feature-designer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions SET default_config = jsonb_set(
  default_config, '{workflow,steps,repropose,config,input_fields}',
  '["spec_row","plan_persisted","council_report_row","check_results"]'::jsonb)
WHERE type='feature-designer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions SET default_config = jsonb_set(
  default_config, '{workflow,steps,repropose,config,prompt_template}', '"# PROMPT \u2014 REVISE the staged build plan\n\nA council reviewed your previous staged plan and asked for revision. Produce a NEW full plan (same staged-v1 JSON schema) that addresses every objection and covers every capability listed missing. You still write no code \u2014 you name stages and edits.\n\nThe SAME hard rules apply: only code_pointers paths or files this plan adds; stages are commits; caps (6 stages / 8 per stage / 24 total); seeds are files; platform not site data; minimal; grounded in the spec; every edit CHANGES something. CHECKLIST ACTS ARE A CLOSED SET: image_deploy | seed_apply | verify \u2014 NEVER invent a new act name. A pre-apply confirmation (e.g. \"confirm the image carries X\") is a verify entry ordered BEFORE the seed_apply; image_deploy is required only when the plan ships code edits.\n\n## The approved spec\n{{.spec_row.summary}}\n{{.spec_row.spec_text}}\n\n## Your previous plan\n{{.plan_persisted.plan_json}}\n\n## The council''s full report \u2014 EVERY reviewer''s verdict and objections\n\nThis is the council_report artifact for this round, verbatim. It contains one entry per seat that voted, whatever the roster is \u2014 read every objection in it, not just the first. Seats may be added between runs; anything in this report is binding on your revision.\n\n{{.council_report_row.body}}\n\n## Verification results (the reviewers'' own read-only queries, now answered)\n{{.check_results.results_text}}\n\nUse these results to SETTLE any objection that hinged on an unverified fact \u2014 cite them in grounded_in. If a result contradicts an edit, change or drop the edit; do not argue with the data.\n\n## Output \u2014 ONLY the staged-v1 plan JSON. Address the objections; do not merely restate the old plan.\n"'::jsonb)
WHERE type='feature-designer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

COMMIT;
