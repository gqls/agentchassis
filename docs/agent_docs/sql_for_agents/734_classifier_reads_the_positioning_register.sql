-- 734_classifier_reads_the_positioning_register.sql
--
-- RFC_037 (owner ruling 2026-08-19) — "feed the register entry to the classifier" —
-- and the owner's direct instruction 2026-09-03: "please fix the classifier to read the register."
--
-- TWO changes to domain-research-classifier, both config-only and live on apply:
--
--  (A) THE PRE-EXISTING DEFECT, found while wiring (B), credited to the bugs_open/445 lane whose
--      87%-unmatchable-tags finding it explains. `classify_and_extract.config.input_fields` is a
--      strict allow-list (ai_actions.go extractDataForAiAgent -> datahelpers.ExtractFields), and
--      `layout_taxonomy` was NOT in it — although the `read_layout_taxonomy` step runs, populates
--      collected_data correctly, and the prompt references {{.layout_taxonomy.industry_tags}}.
--      So the taxonomy was fetched and dropped at the template boundary. MEASURED at the rendered
--      artefact (llm_call_log id a01d9e76, 2026-09-02 20:15:25Z), the model was shown:
--          Current library tags (match these when they describe this site ...):
--          null
--          The library currently has <no value> active layouts. If no existing tag fits ...
--      i.e. told to match a null list, then told to coin. 188 of 216 emitted terms match no
--      layout. This adds `layout_taxonomy` to the allow-list. The step has been live since it was
--      built and has never once delivered its payload: treat the first classification after this
--      as NEW behaviour, not a restoration.
--
--  (B) THE REGISTER INPUT (RFC_037). A new `read_positioning_register` step (query_database,
--      reusing existing machinery — no Go change) between read_layout_taxonomy and
--      classify_and_extract, whose output_field `positioning_register` carries ONE column `block`:
--      either a rendered advisory prose block, or the empty string.
--
-- INERT BY CONSTRUCTION, as the ruling requires ("a site whose entry has not been written yet must
-- be unaffected, and the change must never fail closed on a missing row"). The block is '' unless
-- the site has a register row that is (1) not exclude_from_build, (2) has a non-empty proposition,
-- and (3) does not self-declare "direction unassigned" — a placeholder that sits on 6 domains
-- including webdesign.uk, where it would actively mislead. The query always returns exactly one
-- row, so the template variable can never render <no value>. VERIFIED before writing this file:
-- copyonline.co.uk 2,602 B, gamedesign.uk 1,677 B, webdesign.uk/farmerinsurance.uk/advertise.co.uk/
-- seotools.co.uk all EMPTY.
--
-- ADVISORY, not binding (ruling answer 3): the block says so in its own words, and defers to a
-- Pre-Defined Mission. There is no post-classification collision check here.
--
-- COVERAGE, stated honestly: the register holds 194 domains but only 17 have a `sites` row today,
-- and none of the four 2026-09-02 remakes has an entry. The estate-wide inventory the widened
-- ruling asks for is still an open ask on the owner; this change is what makes an entry MATTER
-- once written, so coverage is now a data task rather than an engineering one.
--
-- ORDERING: the bugs_open/445 lane holds two further edits to this same prompt_template (removing
-- the layout names from the tag examples, and rewording the coining sentence). They have agreed to
-- sequence BEHIND this migration. This file anchors on text they are not touching.
--
-- Apply: psql -f THIS FILE ONLY. Companion ROLLBACK alongside.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='domain-research-classifier' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '734 REFUSED: expected exactly 1 active domain-research-classifier row, found %', n;
  END IF;
  PERFORM snapshot_agent('domain-research-classifier',
                         '734_classifier_reads_the_positioning_register.sql: pre-update');
END $$;

DO $do$
DECLARE
  cfg jsonb; tpl text; newtpl text; n int;
  ifs_before int; ends_before int; elses_before int; vars_before int;
  anchor text := $a734${{if .site_specs.specs.site_archetype}}## Adoption Reference — STRONGEST signal$a734$;
  block  text := $b734${{.positioning_register.block}}
{{if .site_specs.specs.site_archetype}}## Adoption Reference — STRONGEST signal$b734$;
  q text := $q734$WITH me AS (
  SELECT r.* FROM positioning_register r
   JOIN sites s ON lower(s.domain) = lower(r.domain)
   WHERE s.id = $1
     AND COALESCE(r.exclude_from_build, false) = false
     AND COALESCE(r.proposition, '') <> ''
     AND r.proposition NOT ILIKE '%direction unassigned%'
   ORDER BY r.updated_at DESC NULLS LAST, r.created_at DESC
   LIMIT 1
), nbr AS (
  SELECT string_agg(
           format('- %s: %s',
                  COALESCE(NULLIF(n->>'domain',''), n->>'code'),
                  COALESCE(NULLIF(n->>'rule',''), 'no boundary rule recorded')),
           E'\n' ORDER BY n->>'code') AS lines
    FROM me, jsonb_array_elements(me.neighbours) n
), mn AS (
  SELECT string_agg('- ' || value, E'\n') AS lines
    FROM me, jsonb_array_elements_text(me.must_nots)
)
SELECT COALESCE((
  SELECT format(
E'## Portfolio position (ADVISORY)\n\n'
'This domain has a recorded position in the owner''s portfolio register. Treat it as strong '
'guidance on WHERE THIS SITE SITS relative to its siblings, so the estate does not publish two '
'sites with the same proposition. It is advisory, not a constraint: weigh it against the research '
'below, and where a Pre-Defined Mission is present the mission wins.\n\n'
'Recorded position: %s\n%s%s%s%s%s\n'
'Differentiate on the boundaries above. Do not adopt a sibling''s proposition, audience or tool set, '
'and do not describe this site by what it is NOT — state its own ground positively.',
    me.proposition,
    COALESCE('Audience: ' || NULLIF(me.audience,'') || E'\n', ''),
    COALESCE('Mode: ' || NULLIF(me.mode,'') || E'\n', ''),
    COALESCE('Stance: ' || NULLIF(me.stance,'') || E'\n', ''),
    COALESCE(E'\nThis site must NOT:\n' || (SELECT lines FROM mn) || E'\n', ''),
    COALESCE(E'\nSibling sites in this portfolio, and the boundary with each:\n' || (SELECT lines FROM nbr) || E'\n', E'\nNo sibling boundaries are recorded for this domain.\n')
  ) FROM me
), '') AS block;$q734$;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='domain-research-classifier' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  tpl := cfg #>> '{workflow,steps,classify_and_extract,config,prompt_template}';
  IF tpl IS NULL THEN RAISE EXCEPTION '734: classify_and_extract prompt_template not found'; END IF;

  -- anchor must be unique
  n := (length(tpl) - length(replace(tpl, anchor, ''))) / length(anchor);
  IF n <> 1 THEN RAISE EXCEPTION '734: anchor found % times, expected 1', n; END IF;

  -- the variable must not already be present (idempotence guard)
  IF position('{{.positioning_register.block}}' in tpl) > 0 THEN
    RAISE EXCEPTION '734: prompt already carries the positioning_register variable — already applied?';
  END IF;

  ifs_before   := (length(tpl) - length(replace(tpl, '{{if ',   ''))) / length('{{if ');
  ends_before  := (length(tpl) - length(replace(tpl, '{{end}}', ''))) / length('{{end}}');
  elses_before := (length(tpl) - length(replace(tpl, '{{else}}',''))) / length('{{else}}');
  vars_before  := (length(tpl) - length(replace(tpl, '{{.',     ''))) / length('{{.');

  newtpl := replace(tpl, anchor, block);

  -- exactly one new template variable, no new control structures
  IF (length(newtpl) - length(replace(newtpl, '{{.', ''))) / length('{{.') <> vars_before + 1 THEN
    RAISE EXCEPTION '734: expected exactly one new template variable';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{if ',   ''))) / length('{{if ')   <> ifs_before
     OR (length(newtpl) - length(replace(newtpl, '{{end}}', ''))) / length('{{end}}') <> ends_before
     OR (length(newtpl) - length(replace(newtpl, '{{else}}',''))) / length('{{else}}') <> elses_before THEN
    RAISE EXCEPTION '734: template if/else/end balance changed';
  END IF;

  cfg := jsonb_set(cfg, '{workflow,steps,classify_and_extract,config,prompt_template}',
                   to_jsonb(newtpl), false);

  -- (A) + (B): the allow-list gains BOTH the dropped taxonomy and the new register
  cfg := jsonb_set(cfg, '{workflow,steps,classify_and_extract,config,input_fields}',
                   '["input_data","search_results","scraped_data","site_specs","layout_taxonomy","positioning_register"]'::jsonb,
                   false);

  -- (B) the new step, and the rewire
  cfg := jsonb_set(cfg, '{workflow,steps,read_positioning_register}', jsonb_build_object(
           'action',       'query_database',
           'description',  'RFC_037: the site''s own portfolio-register entry and its recorded sibling boundaries, as advisory prose. Empty string when the site has no usable entry, so the classifier is unaffected.',
           'config',       jsonb_build_object('query', q, 'output_format', 'object'),
           'output_field', 'positioning_register',
           'next_step',    'classify_and_extract'
         ), true);
  cfg := jsonb_set(cfg, '{workflow,steps,read_layout_taxonomy,next_step}',
                   '"read_positioning_register"'::jsonb, false);

  UPDATE agent_definitions SET default_config = cfg, updated_at = now()
   WHERE type='domain-research-classifier' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '734: updated % rows, expected exactly 1', n; END IF;
END $do$;

-- Verify (DO/RAISE: a block of bare SELECTs cannot stop a COMMIT)
DO $$
DECLARE cfg jsonb; fields jsonb; tpl text;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='domain-research-classifier' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  fields := cfg #> '{workflow,steps,classify_and_extract,config,input_fields}';
  IF NOT (fields ? 'layout_taxonomy') THEN
    RAISE EXCEPTION '734 VERIFY: layout_taxonomy missing from input_fields — defect (A) not fixed';
  END IF;
  IF NOT (fields ? 'positioning_register') THEN
    RAISE EXCEPTION '734 VERIFY: positioning_register missing from input_fields';
  END IF;
  IF NOT (fields ? 'input_data' AND fields ? 'search_results' AND fields ? 'scraped_data' AND fields ? 'site_specs') THEN
    RAISE EXCEPTION '734 VERIFY: an original input field was dropped';
  END IF;

  IF cfg #>> '{workflow,steps,read_positioning_register,action}' IS DISTINCT FROM 'query_database' THEN
    RAISE EXCEPTION '734 VERIFY: read_positioning_register step missing or wrong action';
  END IF;
  IF cfg #>> '{workflow,steps,read_positioning_register,output_field}' IS DISTINCT FROM 'positioning_register' THEN
    RAISE EXCEPTION '734 VERIFY: output_field must be positioning_register or the prompt variable renders <no value>';
  END IF;
  IF cfg #>> '{workflow,steps,read_positioning_register,next_step}' IS DISTINCT FROM 'classify_and_extract' THEN
    RAISE EXCEPTION '734 VERIFY: new step does not lead to classify_and_extract';
  END IF;
  IF cfg #>> '{workflow,steps,read_layout_taxonomy,next_step}' IS DISTINCT FROM 'read_positioning_register' THEN
    RAISE EXCEPTION '734 VERIFY: read_layout_taxonomy was not rewired — the new step is unreachable';
  END IF;

  tpl := cfg #>> '{workflow,steps,classify_and_extract,config,prompt_template}';
  IF position('{{.positioning_register.block}}' in tpl) = 0 THEN
    RAISE EXCEPTION '734 VERIFY: prompt does not reference the register block';
  END IF;
  IF position('## Adoption Reference — STRONGEST signal' in tpl) = 0 THEN
    RAISE EXCEPTION '734 VERIFY: the adoption block went missing — refuse';
  END IF;
  IF position('{{.layout_taxonomy.industry_tags' in tpl) = 0 THEN
    RAISE EXCEPTION '734 VERIFY: the layout taxonomy reference went missing — refuse';
  END IF;
END $$;

COMMIT;
