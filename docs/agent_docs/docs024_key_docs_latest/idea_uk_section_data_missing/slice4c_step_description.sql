-- Slice 4c follow-up: the image-build-handler flag_rebuild step description said
-- page-only; align it with the deployed behaviour. Run WITH or AFTER the 4c code
-- deploy (cosmetic drift before that is harmless). Guarded + idempotent.
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
  '{workflow,steps,flag_rebuild,description}',
  to_jsonb('Imagery landed: flag the page needs_rebuild and emit needs_page so plan_sections re-resolves the now-present asset. Page scope uses scope_ref directly; section scope (scope_ref "<page>:<ordinal>") maps to its page prefix. Still a no-op for site scope (logo). error_step is complete so a failure here does not fail the asset workflow.'::text)),
  updated_at = now()
WHERE type = 'image-build-handler' AND is_active = true
  AND default_config->'workflow'->'steps'->'flag_rebuild'->>'description' LIKE 'Page-scoped imagery%'
RETURNING default_config->'workflow'->'steps'->'flag_rebuild'->>'description' AS new_description;
-- Expect: UPDATE 1; the RETURNING text begins 'Imagery landed:'.
