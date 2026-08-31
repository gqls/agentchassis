-- 675_build_standard_carrier_row.sql
--
-- The best-in-class build standard's ONE carrier row (owner ruling 2026-08-25 #4 "the mission
-- for every site we build must be to make it the best in class"; BUILD GO relayed 2026-08-31,
-- copy_quality_two_stage rulings ledger). Mechanism in Go, words in one DB row — the house
-- voice's proven shape (voice_style_block, migration 240; platform/voicestyle generalised to
-- named blocks for this row). Injected as {{.build_standard}} by ExecuteLLMPromptAction; a
-- template opts in by naming it, everything else is unaffected.
--
-- SAFE TO APPLY ANY TIME: nothing reads this row until (a) the chassis roll carrying the
-- voicestyle generalisation AND (b) migration 676_HOLD adds the placeholder to a template.
--
-- WORDING PROVENANCE: verbatim from the classifier's "Build standard" block
-- (049_domain_research_classifier.sql:2593, confirmed byte-identical in the live rendered
-- prompt 2026-08-31) with ONE deliberate trim, recorded here: "...build on the reasoning
-- behind those choices rather than copying them" loses its trailing comparison ("build on the
-- reasoning behind those choices."). This lane's measured mechanism is that demonstration
-- phrases in writer-adjacent context are emitted as register, plans/briefs feed the writer,
-- and "rather than" is the production model's strongest tic (canary 2026-08-31: 6 on one page
-- with zero carriers). The meaning survives: "build on the reasoning" already excludes
-- copying. The block's "The bar is not X but Y" core is KEPT — it is the standard's force, and
-- the owner's own words; if he wants the trimmed clause restored, one live migration.

BEGIN;

DO $g$
BEGIN
  IF EXISTS (SELECT 1 FROM agent_default_configs WHERE config_name='build_standard_block') THEN
    RAISE EXCEPTION '675: build_standard_block already exists — another session got here first; read it before re-running anything';
  END IF;
END $g$;

INSERT INTO agent_default_configs (config_name, agent_type, config)
VALUES ('build_standard_block', 'platform', jsonb_build_object(
  'text',
  'BUILD STANDARD (applies to every site, regardless of inputs). Aim for best-in-class quality in this site''s field. The bar is not "competent template" but "stands comparison with the strongest sites in this vertical" — in the quality of the design, the clarity and usefulness of the writing, and the genuine value of any tools or content to the people who will actually use the site. When forming direction, consider what the best sites in this space do — how they position, what their design signals, how their copy reads, what earns return visits — and build on the reasoning behind those choices. Choose design and content that fit this specific industry and these objectives, not a generic house style. Favour fewer things done genuinely well over filler, and prefer interactive or visual elements where they aid understanding. Do what is most useful and interesting for the site''s visitors.',
  'source', '675_build_standard_carrier_row.sql; wording from 049_domain_research_classifier.sql Build standard block, one trim recorded in the header',
  'updated', '2026-08-31',
  'version', 1
));

DO $v$
DECLARE t text;
BEGIN
  SELECT config->>'text' INTO t FROM agent_default_configs WHERE config_name='build_standard_block';
  IF t IS NULL OR length(t) < 400 THEN RAISE EXCEPTION '675 VERIFY: block text missing or implausibly short (%)', COALESCE(length(t),0); END IF;
  IF position('best-in-class' IN t) = 0 THEN RAISE EXCEPTION '675 VERIFY: block lost its own subject'; END IF;
  IF t ~* 'rather than' THEN RAISE EXCEPTION '675 VERIFY: the recorded trim did not happen — a rather-than demonstration is in the carrier'; END IF;
  RAISE NOTICE '675 verify: carrier row present, % chars, trim confirmed.', length(t);
END $v$;

INSERT INTO schema_migrations (filename, checksum, applied_by, notes)
VALUES ('675_build_standard_carrier_row.sql', :'mig_checksum', 'copy_quality_two_stage session',
        'Best-in-class build standard carrier row (owner ruling 2026-08-25 #4, build go 2026-08-31). Inert until the voicestyle-generalisation roll + 676_HOLD opt-ins.');

COMMIT;
