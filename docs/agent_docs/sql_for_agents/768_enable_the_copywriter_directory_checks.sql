-- 768 — enable the two copywriter directory checks on completeness-discovery-agent.
-- Same statement family as 434 (the finance pairs) and its predecessors 150/188/189/190/194/215.
--
-- WHY THIS FILE EXISTS: adding a kind to `directoryCheckProfiles` (commit 48bff098d) creates the
-- two checks in the binary, and `MissingDirectorySectionCheck.Name()` returns the profile's
-- SectionItemType — but a check the agent's `checks` array does not NAME is silently skipped by the
-- registry lookup, never an error. The council's editquality seat raised exactly this on
-- correlation 32c75bc5 ("registering a new item type for a discovery-check finding is NOT a
-- one-line change; the build only surfaces one of the two required follow-on edits"): the Go half
-- is covered by verifier_coverage_test.go (passing), the DB half is this migration and nothing
-- fails without it.
--
-- WHY SAFE TO APPLY BEFORE ANY PAGE EXISTS — the same double self-gate 434 relied on:
--   1. the site's opt-in flag `classification.content_features.copywriter_directory` — absent on
--      every site today, so no site is in scope; and
--   2. the register holding current claims of kind `copywriter` — populated 2026-09-04 (8 claims,
--      4 organisations), so the checks will have something to talk about when a site does opt in.
-- Net effect today: ZERO work items. Checks armed before the pilot builds, which is the order that
-- lets the pilot prove them.
--
-- IMAGE PRECONDITION: the profile ships in v1.0.1361 (cut 06c0b18f2; commit 48bff098d is an
-- ancestor — verified with git merge-base). Applying this BEFORE that image is live is harmless in
-- the same way 434 was: an unregistered check name is skipped, not an error. After the roll, confirm
-- the binary carries it rather than assuming:
--   SELECT pod_name, git_commit FROM service_binary_capabilities WHERE kind='build' AND pod_name LIKE 'agent-chassis-%';
--   git merge-base --is-ancestor 48bff098d <that commit>

BEGIN;

DO $g$
DECLARE arr jsonb; n int;
BEGIN
  SELECT s.v->'config'->'checks' INTO arr
    FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') s(k,v)
   WHERE ad.type='completeness-discovery-agent' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false
     AND ad.deleted_at IS NULL AND s.v->'config' ? 'checks';
  IF arr IS NULL THEN RAISE EXCEPTION '768 REFUSED: no checks array found on completeness-discovery-agent'; END IF;
  IF jsonb_typeof(arr) <> 'array' THEN RAISE EXCEPTION '768 REFUSED: checks is % not an array', jsonb_typeof(arr); END IF;
  IF arr ? 'missing_copywriter_directory_section' OR arr ? 'missing_copywriter_directory_page' THEN
    RAISE EXCEPTION '768 REFUSED: already applied'; END IF;
  -- the sibling pairs must be present, or this is not the array we think it is
  IF NOT (arr ? 'missing_model_directory_section' AND arr ? 'missing_mortgage_lender_directory_page') THEN
    RAISE EXCEPTION '768 REFUSED: the sibling directory checks are absent — wrong array or wrong agent'; END IF;
  SELECT jsonb_array_length(arr) INTO n;
  RAISE NOTICE '768: checks array has % entries before append', n;
END $g$;

UPDATE agent_definitions ad
   SET default_config = jsonb_set(ad.default_config,
         ARRAY['workflow','steps', (SELECT s.k FROM jsonb_each(ad.default_config->'workflow'->'steps') s(k,v) WHERE s.v->'config' ? 'checks' LIMIT 1), 'config','checks'],
         (SELECT s.v->'config'->'checks' FROM jsonb_each(ad.default_config->'workflow'->'steps') s(k,v) WHERE s.v->'config' ? 'checks' LIMIT 1)
           || '["missing_copywriter_directory_section","missing_copywriter_directory_page"]'::jsonb),
       updated_at = NOW()
 WHERE ad.type='completeness-discovery-agent' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;

DO $v$
DECLARE arr jsonb; n int;
BEGIN
  SELECT s.v->'config'->'checks' INTO arr
    FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') s(k,v)
   WHERE ad.type='completeness-discovery-agent' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false
     AND ad.deleted_at IS NULL AND s.v->'config' ? 'checks';
  IF NOT (arr ? 'missing_copywriter_directory_section') THEN RAISE EXCEPTION '768 VERIFY: section check not enabled'; END IF;
  IF NOT (arr ? 'missing_copywriter_directory_page') THEN RAISE EXCEPTION '768 VERIFY: page check not enabled'; END IF;
  -- nothing else was dropped
  IF NOT (arr ? 'missing_model_directory_section' AND arr ? 'missing_adoption_tracker_page'
          AND arr ? 'phantom_internal_links' AND arr ? 'orphan_pages') THEN
    RAISE EXCEPTION '768 VERIFY: pre-existing checks are missing — the append replaced instead of extending'; END IF;
  SELECT jsonb_array_length(arr) INTO n;
  RAISE NOTICE '768: checks array now has % entries', n;
  -- and no site is in scope yet, which is the safety claim this file makes
  SELECT count(*) INTO n FROM site_specs
   WHERE aspect='classification' AND is_current AND data->'content_features' ? 'copywriter_directory';
  IF n <> 0 THEN RAISE NOTICE '768: NOTE — % site(s) already carry the copywriter_directory opt-in; the checks are live for them immediately', n; END IF;
END $v$;

COMMIT;
