-- 764 — domain-research-classifier and build-site-planner render a mission_brief / roadmap_brief that
-- carries no `text` key, instead of printing <no value> under a guard that already opened.
-- bugs_open/453 (the 453 lane's SUMMARY: "the fix lives at the opposite end of the system and belongs
-- to whoever takes it" — taken by portfolio_positioning, 2026-09-03 evening).
--
-- COUNCIL ROUND 1 (888e7319, 2026-09-03 19:47Z): REVISE — gating objection from bug_historian. Every
-- objection answered below with evidence, and the file revised; resubmitted on the SAME correlation.
--
--  * editquality (funcmap per action): classify_and_extract, plan_site AND domain-strategist's
--    analyze_strategy ALL run under `execute_llm_prompt` [MEASURED at agent_definitions]. That action
--    renders through datahelpers.RenderPromptTemplate (ai_actions.go:328), which parses with
--    datahelpers.PromptTemplateFuncs() — the funcmap that registers toJSON (template_context_contract.go);
--    promptTemplateRenderingActions["execute_llm_prompt"] = nil, i.e. no per-action root differences.
--    The proof was REWRITTEN to import the real package and call RenderPromptTemplate + ScanMissingValues
--    (PRC-003's own scan): portfolio_positioning/tplproof/proof_test.go, `go test -tags tplproof`.
--    Result: object case ScanMissingValues occurrences planner 3→2, classifier 4→3, the ORIGINAL report
--    attributing `site_specs.specs.mission_brief.text` and the fixed one not; prose case byte-identical;
--    no-brief case unchanged. PASS.
--  * editquality + prior_art (dual active rows): every UPDATE/SELECT here is scoped
--    `is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL` and the guard REFUSES unless
--    exactly 1 row matches per type; the reviewers' own check read 1 active row, version {1}, for both.
--  * bug_historian (completeness census walked only top-level steps): re-run BOTH ways — the corpus's
--    prescribed nested walk `jsonb_path_query(default_config,'$.**.steps')` over every active agent, and a
--    regex over the WHOLE default_config::text at any depth. Both return exactly the four expressions in
--    the two steps named here; the classifier's third prompt_template (review_mission_alignment) carries
--    none; 49 other `{{…\.text}}` reads fleet-wide sit under different roots (schema_hint, decisions,
--    experience_*) written by their own producers — a different class. Re-verify with the query in the
--    council submission's grounded_in.
--  * bug_historian (per-call-site patch; loud failure; write-side guard): the read-side loud failure the
--    seat asks for is BUILT — PRC-003 (681b0ee65, 2026-09-03 17:33Z, NOT live till the roll) names,
--    strips and escalates <no value> inside RenderPromptTemplate. It makes the hole VISIBLE; it cannot
--    make the brief visible. 764 is its complement: once both are live, a brief renders as its object and
--    any remaining hole is named in the log. The config-side guard is WFA-024 (453's lint). A write-side
--    normaliser (every brief carries `text`) is a producer contract and is deliberately not bundled here.
--    Why not domain-strategist's whole-blob shape: these two templates render classification, identity,
--    briefing and strategy SEPARATELY under headings; `{{.site_specs}}` would duplicate all of it. The
--    fallback applies the same principle — toJSON of the object — at the aspect level.
--  * guardian (two pipelines in one edit; in-flight runs): the council plan now carries one edit per
--    agent (classification/onboarding pipeline; build/planning pipeline). In-flight orchestrations are
--    unaffected by construction — LANDMINES "Editing a live agent's config CANNOT reach an in-flight
--    orchestration — it carries its own snapshot (orchestration_states.workflow_plan)"; the change
--    reaches the NEXT spawn of each agent.
--  * prior_art (the same-day LANDMINES correction): that correction — written by this lane — NARROWED the
--    mechanism (a third consumer, domain-strategist, DOES read the structured brief, via a whole-blob
--    render) and left the premise intact: these two templates render <no value> for a keyless brief,
--    measured at llm_call_log with gamedesign.uk (has text) as positive and boxingonline.com (no brief)
--    as negative control. The strategist's success is the reference implementation this fix mirrors.
--  * reuse_agent: the harness now reuses the production renderer rather than re-implementing it.
--
-- COUNCIL: submitted 2026-09-03 ~20:1xZ, SUBMISSION_CORR 888e7319-01ae-4371-846d-76fe227a1ebc (run orch 33a3f8e0…).
--   Verdict: SELECT body FROM doc_notes WHERE categories ? 'council-gate' AND body LIKE '%888e7319%' ORDER BY created_at DESC LIMIT 1;
-- _HOLD: applied BY HAND after the council verdict and at a moment of the owner's choosing — it changes
-- what two shared agents see on every run. Nothing about ordering; the guard checks DRIFT (md5 pins).
--
-- THE DEFECT, measured: exactly two live agents read a brief; both spell it
--   {{if .site_specs.specs.mission_brief}} … {{.site_specs.specs.mission_brief.text}} … {{end}}
-- (and the same pair for roadmap_brief). Under text/template's default missingkey the guard opens on
-- the parent and the child prints "<no value>". 7 of 23 current mission_brief specs lack `text`
-- [MEASURED 2026-09-03 17:44Z], from THREE producers (brief-writer ×5, manual ×1, an operator ×1).
-- domain-strategist renders {{.site_specs}} whole and reads the same brief fine — the working reference.
-- Cost on one site: copyonline's classifier recorded "no mission brief was supplied" and classified the
-- site from its previous Drupal install; ten pages built, none of the brief's thirty.
--
-- THE CHANGE — four expressions, two templates, the complete fleet-wide set [MEASURED 18:22Z]:
--   {{.site_specs.specs.<aspect>.text}}
--     → {{if .site_specs.specs.<aspect>.text}}{{.site_specs.specs.<aspect>.text}}{{else}}{{toJSON .site_specs.specs.<aspect>}}{{end}}
-- Prose briefs (gamedesign.uk has `.text`) render byte-identically; object briefs render as indented
-- JSON via the SAME toJSON the planner already uses for {{toJSON .site_specs.specs.strategy}}
-- (datahelpers template funcmap, template_context_contract.go:156). `{{if .x.text}}` on a map lacking
-- the key is falsy under default missingkey — that is the behaviour we already observe, now used on purpose.
--
-- PROVEN UNDER THE REAL ENGINE BEFORE WRITING THIS FILE (not a shape check — migration 734 asserted shape
-- through 17 guards and broke the classifier fleet-wide for 4h22m): both full templates pulled from the
-- live rows, the replacements applied, parsed and executed with text/template + toJSON, three contexts
-- each (brief with text / brief object without text / no brief) plus a control on the UNMODIFIED template.
-- Result 2026-09-03 ~20:0xZ: prose case delta 0 and byte-same block; object case whole-render <no value>
-- delta EXACTLY −1 and the object's sentinel present; no-brief case renders no Mission block; control
-- reproduces the defect. Harness: portfolio_positioning/tplproof/main.go (go run . after pulling the two
-- templates to <type>.tpl). 8/8.
--
-- ⚠ AFTER APPLYING, MAKE IT RUN ONCE — the proof above is the template engine, not the fleet:
--   1. fire one classification for a site whose brief lacks `text` (copyonline.co.uk, site 3d965325…),
--   2. read orchestration_states.current_step/status for that run — it must COMPLETE,
--   3. read the rendered prompt: the ## Pre-Defined Mission block must carry the JSON object, and
--   4. the NEW site_specs.classification.data->>'reasoning' must NOT say the brief was missing.
--   Then the same for the planner when it next runs. Do not claim this works before step 4.

BEGIN;
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions WHERE type='domain-research-classifier' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN RAISE EXCEPTION '764 REFUSED: expected exactly 1 active domain-research-classifier, found %', n; END IF;
  SELECT count(*) INTO n FROM agent_definitions WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN RAISE EXCEPTION '764 REFUSED: expected exactly 1 active build-site-planner, found %', n; END IF;
  PERFORM snapshot_agent('domain-research-classifier', '764_classifier_and_planner_render_the_brief_object_when_it_has_no_text_HOLD.sql: pre-update');
  PERFORM snapshot_agent('build-site-planner', '764_classifier_and_planner_render_the_brief_object_when_it_has_no_text_HOLD.sql: pre-update');
END $$;

DO $do$
DECLARE
  rec record; tpl text; newtpl text; n int; ifs int; ends int; elses int; vars int;
  a_m text := '{{.site_specs.specs.mission_brief.text}}';
  r_m text := '{{if .site_specs.specs.mission_brief.text}}{{.site_specs.specs.mission_brief.text}}{{else}}{{toJSON .site_specs.specs.mission_brief}}{{end}}';
  a_r text := '{{.site_specs.specs.roadmap_brief.text}}';
  r_r text := '{{if .site_specs.specs.roadmap_brief.text}}{{.site_specs.specs.roadmap_brief.text}}{{else}}{{toJSON .site_specs.specs.roadmap_brief}}{{end}}';
BEGIN
  FOR rec IN SELECT * FROM (VALUES
      ('domain-research-classifier','classify_and_extract','7d8494ede7be6f8900511d4659d0f4db'),
      ('build-site-planner','plan_site','a73e9d499a507dec39e7cfd233b83735')) AS t(atype, step, pin)
  LOOP
    SELECT default_config #>> ('{workflow,steps,'||rec.step||',config,prompt_template}')::text[] INTO tpl
      FROM agent_definitions WHERE type=rec.atype AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF tpl IS NULL THEN RAISE EXCEPTION '764: % prompt_template not found at step %', rec.atype, rec.step; END IF;
    IF md5(tpl) <> rec.pin THEN RAISE EXCEPTION '764 REFUSED (drift): % prompt_template md5 % <> pinned % — re-read the template and re-pin', rec.atype, md5(tpl), rec.pin; END IF;
    IF position('{{toJSON .site_specs.specs.mission_brief}}' in tpl) > 0 THEN RAISE EXCEPTION '764: already applied on %', rec.atype; END IF;
    n := (length(tpl) - length(replace(tpl, a_m, ''))) / length(a_m);
    IF n <> 1 THEN RAISE EXCEPTION '764: % mission anchor found % times, expected 1', rec.atype, n; END IF;
    n := (length(tpl) - length(replace(tpl, a_r, ''))) / length(a_r);
    IF n <> 1 THEN RAISE EXCEPTION '764: % roadmap anchor found % times, expected 1', rec.atype, n; END IF;
    ifs   := (length(tpl) - length(replace(tpl, '{{if ',   ''))) / length('{{if ');
    ends  := (length(tpl) - length(replace(tpl, '{{end}}', ''))) / length('{{end}}');
    elses := (length(tpl) - length(replace(tpl, '{{else}}',''))) / length('{{else}}');
    vars  := (length(tpl) - length(replace(tpl, '{{.',     ''))) / length('{{.');
    newtpl := replace(replace(tpl, a_m, r_m), a_r, r_r);
    IF length(newtpl) <> length(tpl) + (length(r_m)-length(a_m)) + (length(r_r)-length(a_r)) THEN
      RAISE EXCEPTION '764: % unexpected length delta — an anchor matched more than intended', rec.atype; END IF;
    -- each replacement adds exactly one {{if , one {{else}}, one {{end}}, and no new {{. (the {{if . and {{toJSON . forms do not match '{{.')
    IF (length(newtpl) - length(replace(newtpl, '{{if ',   ''))) / length('{{if ')   <> ifs + 2
       OR (length(newtpl) - length(replace(newtpl, '{{end}}', ''))) / length('{{end}}') <> ends + 2
       OR (length(newtpl) - length(replace(newtpl, '{{else}}',''))) / length('{{else}}')<> elses + 2
       OR (length(newtpl) - length(replace(newtpl, '{{.',     ''))) / length('{{.')      <> vars THEN
      RAISE EXCEPTION '764: % if/else/end/var balance moved by other than the intended +2/+2/+2/+0', rec.atype; END IF;
    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config, ('{workflow,steps,'||rec.step||',config,prompt_template}')::text[], to_jsonb(newtpl)),
           updated_at = now()
     WHERE type=rec.atype AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    GET DIAGNOSTICS n = ROW_COUNT;
    IF n <> 1 THEN RAISE EXCEPTION '764: % updated % rows, expected exactly 1', rec.atype, n; END IF;
  END LOOP;
END $do$;

DO $$
DECLARE rec record; tpl text; n int;
BEGIN
  FOR rec IN SELECT * FROM (VALUES ('domain-research-classifier','classify_and_extract'),('build-site-planner','plan_site')) AS t(atype, step) LOOP
    SELECT default_config #>> ('{workflow,steps,'||rec.step||',config,prompt_template}')::text[] INTO tpl
      FROM agent_definitions WHERE type=rec.atype AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    n := (length(tpl) - length(replace(tpl, '{{toJSON .site_specs.specs.mission_brief}}', ''))) / length('{{toJSON .site_specs.specs.mission_brief}}');
    IF n <> 1 THEN RAISE EXCEPTION '764 VERIFY: % mission fallback present % times, expected 1', rec.atype, n; END IF;
    n := (length(tpl) - length(replace(tpl, '{{toJSON .site_specs.specs.roadmap_brief}}', ''))) / length('{{toJSON .site_specs.specs.roadmap_brief}}');
    IF n <> 1 THEN RAISE EXCEPTION '764 VERIFY: % roadmap fallback present % times, expected 1', rec.atype, n; END IF;
    IF position('{{if .site_specs.specs.mission_brief}}' in tpl) = 0 THEN RAISE EXCEPTION '764 VERIFY: % outer mission guard lost', rec.atype; END IF;
  END LOOP;
  IF (SELECT count(*) FROM agent_definitions_backup WHERE snapshot_reason LIKE '764_%' AND type IN ('domain-research-classifier','build-site-planner')) < 2 THEN
    RAISE EXCEPTION '764 VERIFY: pre-image snapshots missing from agent_definitions_backup'; END IF;
END $$;
COMMIT;
