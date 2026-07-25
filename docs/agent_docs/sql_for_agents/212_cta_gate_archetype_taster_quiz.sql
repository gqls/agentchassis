-- 212_cta_gate_archetype_taster_quiz.sql — 2026-07-25, cta_link_integrity (bugs_open/023)
--
-- The last hardcoded placeholder CTA in the active library, found by the standing
-- lint (scripts/check_cta_gates.py) immediately after migration 211 swept the
-- 156 field-bound anchors. Worth recording HOW it was found: 211's own worklist
-- came from a field parse, which by construction cannot see an anchor whose href
-- is a literal '#'. The lint was written to report that second shape, and on its
-- first run it produced one component the sweep had missed. The tool emitting the
-- distinction is what caught it — not a re-reading of the same list.
--
--   archetype-taster-quiz   <a class="atq-cta-primary" id="atq-result-cta" href="#">
--                             Explore the full assessment ...
--
-- A button that always renders and never goes anywhere: 023's exact shape, and
-- the same shape migration 179 removed from tool-guide-intro (#guide-start).
--
-- NOT a JS-driven control (that check is why the lint excludes some '#' anchors):
-- the id 'atq-result-cta' occurs 4 times in this component and NOTHING assigns
-- its href — two are CSS rules for .atq-result-cta-group, one is the wrapping
-- div, one is the anchor itself. The component's inline <script> never touches
-- it, and no other active component references the id. Contrast
-- provocations-archive-list, whose href="#" anchor carries [data-archive-template]
-- + hidden and IS a JS clone source — gating that one would break its archive.
--
-- 0 placements today, so this is prophylactic — which is the point. It is
-- dormant library stock, and bugs_open/045 is the case of dormant stock being
-- adopted onto a live page and shipping its frozen defaults (a Bayesian ranker's
-- labels onto an unrelated tool). Its live sibling tool-archetype-taster-quiz was
-- already gated by 211.
--
-- input_schema is currently '{}' — this component's template is otherwise fully
-- static. Adding one renderer/optional field is safe: plan_sections_action.go:999
-- normalises a missing schema to an empty map and walks fields uniformly; an
-- optional field with on_missing:skip_field adds no requirement to any build.
--
-- Config change: LIVE IMMEDIATELY. ROLLBACK: bak_cta_gate_atq_20260725.

\set ON_ERROR_STOP on
BEGIN;

CREATE TABLE bak_cta_gate_atq_20260725 AS
SELECT * FROM content_components WHERE name = 'archetype-taster-quiz';

DO $$
DECLARE tpl text; n int;
BEGIN
  SELECT html_template INTO STRICT tpl FROM content_components WHERE name='archetype-taster-quiz';
  IF position('<a class="atq-cta-primary" id="atq-result-cta" href="#">' in tpl) = 0 THEN
    RAISE EXCEPTION 'needle absent — anchor has drifted since 2026-07-25, re-derive before applying';
  END IF;

  UPDATE content_components SET html_template = replace(html_template,
      '<a class="atq-cta-primary" id="atq-result-cta" href="#">',
      '{{if .result_cta_url}}<a class="atq-cta-primary" id="atq-result-cta" href="{{.result_cta_url}}">'),
    updated_at = now()
  WHERE name = 'archetype-taster-quiz';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION 'anchor update touched % rows', n; END IF;

  -- close the gate after the anchor's own </a>, which is the first one after it
  UPDATE content_components SET html_template = replace(html_template,
      '          <span aria-hidden="true">→</span>
        </a>',
      '          <span aria-hidden="true">→</span>
        </a>{{end}}'),
    updated_at = now()
  WHERE name = 'archetype-taster-quiz';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION 'gate-close update touched % rows', n; END IF;

  UPDATE content_components SET input_schema = jsonb_set(
      CASE WHEN input_schema ? 'fields' THEN input_schema
           ELSE jsonb_set(COALESCE(input_schema,'{}'::jsonb), '{fields}', '{}'::jsonb) END,
      '{fields,result_cta_url}',
      '{"type":"url","source":"renderer","required":false,"on_missing":"skip_field","llm_guidance":"Never author this. The renderer resolves where the full assessment lives; without a resolved destination the button is not rendered at all (LNK-005)."}'::jsonb),
    updated_at = now()
  WHERE name = 'archetype-taster-quiz';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION 'schema update touched % rows', n; END IF;

  -- ── post-conditions ────────────────────────────────────────────────────────
  SELECT count(*) INTO n FROM content_components
   WHERE name='archetype-taster-quiz'
     AND html_template LIKE '%{{if .result_cta_url}}<a class="atq-cta-primary" id="atq-result-cta" href="{{.result_cta_url}}">%'
     AND html_template NOT LIKE '%<a class="atq-cta-primary" id="atq-result-cta" href="#">%'
     AND input_schema->'fields'->'result_cta_url'->>'source' = 'renderer'
     AND input_schema->'fields'->'result_cta_url'->>'required' = 'false';
  IF n <> 1 THEN RAISE EXCEPTION 'post-condition failed: gate or field not in place'; END IF;

  -- the template's {{if}}/{{range}}/{{with}} and {{end}} counts must still balance
  SELECT (SELECT count(*) FROM regexp_matches(html_template,'\{\{ *(if|range|with) ','g'))
       - (SELECT count(*) FROM regexp_matches(html_template,'\{\{ *end *\}\}','g'))
    INTO n FROM content_components WHERE name='archetype-taster-quiz';
  IF n <> 0 THEN RAISE EXCEPTION 'post-condition failed: template actions unbalanced by %', n; END IF;
END $$;

INSERT INTO schema_migrations (filename, notes)
VALUES ('212_cta_gate_archetype_taster_quiz.sql',
        'bugs_open/023: gate the last hardcoded placeholder CTA in the active library (archetype-taster-quiz "Explore the full assessment", href="#", 0 placements) and give it a renderer/optional result_cta_url. Found by scripts/check_cta_gates.py on its first run — the field parse behind migration 211 cannot see a literal-# anchor.');

COMMIT;

-- Post-apply: ./scripts/check_cta_gates.py  — PLACEHOLDER should drop to 1
-- (image-hover-card-grid, deliberately left: its anchor wraps the whole card, so
--  gating would delete the card's image and title, not just a control).
