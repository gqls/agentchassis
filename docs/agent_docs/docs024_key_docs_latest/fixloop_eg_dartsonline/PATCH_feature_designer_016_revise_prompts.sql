-- PATCH_feature_designer_016_revise_prompts.sql — fix bugs_open/016 for feature-designer.
-- 2026-07-18, feature-builder thread. Applies to clients_db.
--
-- THE DEFECT (bugs_open/016, found by the experience-loop thread): prompt
-- template data is unwrapped by ExtractFields → UnwrapDeep, which strips the
-- {type,result} wrapper. So in a PROMPT TEMPLATE, {{.review_X.result}} is a
-- lookup for a "result" key that no longer exists — a json-output step renders
-- it as "<no value>", SILENTLY. feature-designer's repropose (5 refs) and
-- reframe (2 refs) therefore fed the reviser NO objections at all: it revised
-- blind, seeing only its previous plan and "address every objection".
--
-- PROVEN LIVE HERE: run 3b084712 (2026-07-18) burned all three revise rounds
-- and escalated with the bug-historian's objection UNCHANGED across every
-- round — the reviser never saw it. The plan still improved factually because
-- {{.check_results.results_text}} is CORRECT (results_text is a field ON the
-- unwrapped value, not the stripped wrapper) — which is exactly why the
-- failure looked like stubbornness rather than a broken feedback loop.
--
-- THE FIX: in prompt templates ONLY, {{.review_X.result}} → {{.review_X}}.
-- Config dot-paths (review_fields, check_fields, plan_field) read RAW
-- collected_data and keep .result — they are correct and untouched here.
--
-- SURGICAL BY CONSTRUCTION (the rule this thread's own council taught it, and
-- CLAUDE.md's council-gate note): agent_definitions rows are CO-EDITED. This
-- patch jsonb_set's exactly two leaf paths and leaves every other step, key
-- and byte identical — rather than reconstructing default_config from a seed
-- file whose view of the live row is necessarily partial.

BEGIN;

SELECT snapshot_agent('feature-designer', 'pre-update: bugs_open/016 — revise/reframe prompts dropped reviewer output');

-- repropose.prompt_template: strip .result from the review_* references only.
UPDATE agent_definitions SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,repropose,config,prompt_template}',
    to_jsonb(regexp_replace(
        default_config->'workflow'->'steps'->'repropose'->'config'->>'prompt_template',
        '\{\{\.(review_[a-z_]+)\.result\}\}', '{{.\1}}', 'g'))
)
WHERE type = 'feature-designer'
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- reframe.prompt_template: same.
UPDATE agent_definitions SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,reframe,config,prompt_template}',
    to_jsonb(regexp_replace(
        default_config->'workflow'->'steps'->'reframe'->'config'->>'prompt_template',
        '\{\{\.(review_[a-z_]+)\.result\}\}', '{{.\1}}', 'g'))
)
WHERE type = 'feature-designer'
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

COMMIT;

-- Verification (expect 0, 0, then true, true):
--   SELECT (SELECT count(*) FROM regexp_matches(default_config->'workflow'->'steps'->'repropose'->'config'->>'prompt_template','\{\{\.review_[a-z_]+\.result\}\}','g')) AS repropose_broken,
--          (SELECT count(*) FROM regexp_matches(default_config->'workflow'->'steps'->'reframe'->'config'->>'prompt_template','\{\{\.review_[a-z_]+\.result\}\}','g')) AS reframe_broken,
--          position('{{.review_bug_historian}}' in default_config->'workflow'->'steps'->'repropose'->'config'->>'prompt_template')>0 AS fixed_ref_present,
--          position('{{.check_results.results_text}}' in default_config->'workflow'->'steps'->'repropose'->'config'->>'prompt_template')>0 AS check_results_untouched
--   FROM agent_definitions WHERE type='feature-designer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--
-- Rollback: restore the pre-update snapshot from agent_definitions_backup.
