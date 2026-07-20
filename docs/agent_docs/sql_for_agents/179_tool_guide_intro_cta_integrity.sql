-- 179_tool_guide_intro_cta_integrity.sql — 2026-07-20, cta_link_integrity (bugs_open/023, P2.2+P2.3)
--
-- tool-guide-intro shipped two of the four owner-reported broken buttons, by
-- construction, on every page that adopted it (leopardess ×2, finetuning.uk,
-- robot-hands.com — all four placements now removed; this fixes the COMPONENT
-- so the next adoption cannot re-ship the defects):
--
--   1. "Start the Guide" -> href="#guide-start" HARDCODED in the template; no
--      component anywhere emits that id (class D). There was no cta_primary_url
--      field in the schema at all.
--   2. "Visit the Tool"  -> cta_secondary_url was source:llm + required:true —
--      an instruction to author a URL the model cannot look up. It fabricated
--      leopardess.contactforsales.com (NXDOMAIN, an @->. transform of the site's
--      contact email) and finetuning.ai/services/... (different-TLD variant of
--      the site's own domain, RESOLVES to a page the owner does not control).
--      Class E. Both anchors were ungated (class C, violates LNK-005).
--
-- Fix, per bugs_open/023's candidates 2 & 4:
--   * add cta_primary_url   {type:url, source:renderer, optional, skip_field}
--   * cta_secondary_url  -> {type:url, source:renderer, optional, skip_field}
--     (source:renderer = resolver-owned; never LLM-authored. Today, with
--     tool-guide-intro outside ctaFieldNames, both stay unset and the gated
--     template renders NO button — LNK-005 satisfied. When the schema-derived
--     pairing lands (council trail 2525f980), the resolver can own both:
--     each *_url has its *_label sibling, and renderer is the immune class.)
--   * gate both anchors {{if .x_url}}...{{end}} and drop the #guide-start
--     hardcode. Precedent: this component already gates guide_image_url.
--
-- Zero live placements at apply time (verified below with a needle-gate), so
-- this changes no deployed page; it changes what the NEXT page gets.
-- ROLLBACK: restore from bak_tool_guide_intro_20260720 (full row snapshot).

\set ON_ERROR_STOP on
BEGIN;

-- snapshot (bak_ convention)
CREATE TABLE bak_tool_guide_intro_20260720 AS
SELECT * FROM content_components WHERE name = 'tool-guide-intro';

DO $$
DECLARE
  placements int;
  tpl text;
  changed int;
BEGIN
  -- needle-gate 0: the fix is prophylactic BY DESIGN — abort if adoption re-appeared
  SELECT count(*) INTO placements FROM page_components WHERE slot_name = 'tool-guide-intro';
  IF placements <> 0 THEN
    RAISE EXCEPTION 'expected 0 live placements of tool-guide-intro, found % — re-verify page impact before applying', placements;
  END IF;

  -- needle-gate 1: exactly one target row, carrying both needles we edit
  SELECT html_template INTO STRICT tpl FROM content_components WHERE name = 'tool-guide-intro';
  IF position('<a href="#guide-start" class="tgi-btn-primary">{{.cta_primary_label}}</a>' in tpl) = 0
     OR position('<a href="{{.cta_secondary_url}}" class="tgi-btn-secondary">{{.cta_secondary_label}}</a>' in tpl) = 0 THEN
    RAISE EXCEPTION 'template anchors not in expected form — component drifted since 2026-07-20; re-derive the replace() needles';
  END IF;

  -- schema: add cta_primary_url, replace cta_secondary_url
  UPDATE content_components SET input_schema =
    jsonb_set(
      jsonb_set(input_schema,
        '{fields,cta_primary_url}',
        '{"type":"url","source":"renderer","required":false,"on_missing":"skip_field",
          "llm_guidance":"Primary CTA destination. Resolver-owned (source:renderer) — never LLM-authored. Unset means the gated template renders no button (LNK-005)."}'::jsonb),
      '{fields,cta_secondary_url}',
      '{"type":"url","source":"renderer","required":false,"on_missing":"skip_field",
        "llm_guidance":"Secondary CTA destination. Resolver-owned (source:renderer) — never LLM-authored. The previous source:llm+required:true forced URL fabrication (bugs_open/023 class E). Unset means no button."}'::jsonb),
    updated_at = now()
  WHERE name = 'tool-guide-intro';
  GET DIAGNOSTICS changed = ROW_COUNT;
  IF changed <> 1 THEN RAISE EXCEPTION 'schema update touched % rows, expected 1', changed; END IF;

  -- template: gate both anchors, drop the dead fragment
  UPDATE content_components SET html_template = replace(replace(html_template,
      '<a href="#guide-start" class="tgi-btn-primary">{{.cta_primary_label}}</a>',
      '{{if .cta_primary_url}}<a href="{{.cta_primary_url}}" class="tgi-btn-primary">{{.cta_primary_label}}</a>{{end}}'),
      '<a href="{{.cta_secondary_url}}" class="tgi-btn-secondary">{{.cta_secondary_label}}</a>',
      '{{if .cta_secondary_url}}<a href="{{.cta_secondary_url}}" class="tgi-btn-secondary">{{.cta_secondary_label}}</a>{{end}}')
  WHERE name = 'tool-guide-intro';

  -- post-conditions: needles gone / gates present, schema sources flipped
  SELECT html_template INTO STRICT tpl FROM content_components WHERE name = 'tool-guide-intro';
  IF position('#guide-start' in tpl) <> 0 THEN
    RAISE EXCEPTION 'post-condition failed: #guide-start still present';
  END IF;
  IF position('{{if .cta_primary_url}}' in tpl) = 0 OR position('{{if .cta_secondary_url}}' in tpl) = 0 THEN
    RAISE EXCEPTION 'post-condition failed: gated anchors missing';
  END IF;
  IF EXISTS (
    SELECT 1 FROM content_components c, jsonb_each(c.input_schema->'fields') e(k,v)
    WHERE c.name='tool-guide-intro' AND e.k IN ('cta_primary_url','cta_secondary_url')
      AND (e.v->>'source' <> 'renderer' OR COALESCE(e.v->>'required','false') <> 'false')
  ) THEN
    RAISE EXCEPTION 'post-condition failed: cta url fields not renderer/optional';
  END IF;
END $$;

-- ledger row in the SAME transaction (bugs_open/007: applied and recorded in one sitting)
INSERT INTO schema_migrations (filename, notes)
VALUES ('179_tool_guide_intro_cta_integrity.sql',
        'bugs_open/023 P2.2+P2.3: tool-guide-intro — add cta_primary_url (renderer, optional), flip cta_secondary_url llm+required -> renderer+optional, gate both anchors, remove hardcoded #guide-start. Prophylactic: 0 live placements at apply time.');

COMMIT;

-- Post-apply (outside the transaction), for the record:
--   SELECT e.k, e.v->>'source', e.v->>'required' FROM content_components c,
--     jsonb_each(c.input_schema->'fields') e(k,v)
--   WHERE c.name='tool-guide-intro' AND e.k ~ 'cta';
--   SELECT position('#guide-start' in html_template) FROM content_components
--   WHERE name='tool-guide-intro';   -- must be 0
