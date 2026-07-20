-- 181_class_e_live_cta_url_integrity.sql — 2026-07-20, cta_link_integrity (bugs_open/023, class C+E)
--
-- Migration 179 closed classes C/D/E on tool-guide-intro, which had ZERO live
-- placements — prophylactic by design. This closes the same classes on the
-- components that are actually on live pages. Measured 2026-07-20:
--
--   content-block-about  cta_url            llm+required, no fallback   13 placements / 5 sites
--   tool-cta             primary_cta_url    llm+required, no fallback    3 placements / 2 sites
--                        secondary_cta_url  llm+required, no fallback
--   platform-comparison  cta_url            llm+required, no fallback    1 placement  / 1 site
--
-- All four anchors are UNGATED (class C). Every one of these components pairs
-- that LLM-authored URL with a label that ALWAYS renders — content-block-about's
-- cta_label is llm+fallback, and platform-comparison's + tool-cta's labels are
-- source:static with fallbacks, which plan_sections_action.go:1210-1218 writes
-- unconditionally, bypassing required/on_missing. So the button always has text
-- and its destination is invented. That is 023's root cause, live, on 17 placements.
--
-- source:llm + required:true on a URL is an instruction to author a value the
-- model cannot look up. It is what produced leopardess.contactforsales.com
-- (an @->. transform of the site's own contact email) and finetuning.ai.
--
-- ── The three components are NOT treated alike, on purpose ──────────────────
--
-- (a) content-block-about -> flip llm to renderer + gate.
--     SAFE because the resolver already owns this field: content-block-about
--     {cta_url} is in ctaFieldNames (resolve_internal_links_action.go), so
--     chooseCTATargets populates it on every render. Flipping the declared
--     source to renderer makes the schema tell the truth about who writes it.
--
-- (b) tool-cta -> flip both to renderer + gate.
--     tool-cta is NOT in ctaFieldNames, so after the flip nothing populates
--     these and the gated template renders NO button. That is the intended
--     LNK-005 outcome and it is strictly an improvement here, because the
--     values it renders today are broken anyway — verified live 2026-07-20:
--       finetuning.uk        /tools                       404
--       finetuning.uk        /tools/llm-cost-calculator   404
--       leopardess           /tools                       404
--     Removing a 404 button beats keeping it. (These are also 3 of the 312
--     live broken links catalogued in bugs_open/049.)
--
-- (c) platform-comparison -> GATE ONLY. Schema deliberately left alone.
--     Its single live value, vonc.com /tools/gauntlet/index.html, is a REAL
--     working page (200, verified 2026-07-20). It is not in ctaFieldNames
--     either, so flipping it to renderer would leave nothing to populate it
--     and would DELETE A WORKING BUTTON from a live page. The structurally
--     correct flip is therefore deferred until the field has an owner —
--     i.e. until the schema-derived pairing lands (council trail 2525f980),
--     which would make the resolver cover it. Gating alone is still worth
--     doing: it removes the href="" failure mode without removing the button.
--     Residual recorded in bugs_open/023.
--
-- Being a config change this is LIVE IMMEDIATELY — no image roll. It takes
-- effect per page at that page's next render; it does not rewrite deployed HTML.
--
-- ROLLBACK: bak_class_e_components_20260720 holds the three full rows.

\set ON_ERROR_STOP on
BEGIN;

CREATE TABLE bak_class_e_components_20260720 AS
SELECT * FROM content_components
WHERE name IN ('content-block-about','tool-cta','platform-comparison');

DO $$
DECLARE
  tpl text;
  changed int;
  n int;
BEGIN
  -- ── needle-gate: all four anchors must be present in their expected, UNGATED form.
  -- If any has drifted since 2026-07-20, abort rather than half-apply.
  SELECT html_template INTO STRICT tpl FROM content_components WHERE name='content-block-about';
  IF position('<a href="{{.cta_url}}" class="about-cta">{{.cta_label}}</a>' in tpl) = 0 THEN
    RAISE EXCEPTION 'content-block-about anchor not in expected form — re-derive the needle';
  END IF;

  SELECT html_template INTO STRICT tpl FROM content_components WHERE name='tool-cta';
  IF position('<a href="{{.primary_cta_url}}" class="tool-cta-btn-primary">{{.primary_cta_label}}</a>' in tpl) = 0
     OR position('<a href="{{.secondary_cta_url}}" class="tool-cta-btn-secondary">{{.secondary_cta_label}}</a>' in tpl) = 0 THEN
    RAISE EXCEPTION 'tool-cta anchors not in expected form — re-derive the needles';
  END IF;

  SELECT html_template INTO STRICT tpl FROM content_components WHERE name='platform-comparison';
  IF position('<a href="{{.cta_url}}" class="pc-cta">{{.cta_label}}</a>' in tpl) = 0 THEN
    RAISE EXCEPTION 'platform-comparison anchor not in expected form — re-derive the needle';
  END IF;

  -- ── needle-gate: the four fields must still be the llm+required shape we measured.
  SELECT count(*) INTO n
  FROM content_components c, jsonb_each(c.input_schema->'fields') e(k,v)
  WHERE (c.name,e.k) IN (('content-block-about','cta_url'),('tool-cta','primary_cta_url'),
                         ('tool-cta','secondary_cta_url'),('platform-comparison','cta_url'))
    AND v->>'source'='llm' AND v->>'required'='true';
  IF n <> 4 THEN
    RAISE EXCEPTION 'expected 4 llm+required url fields, found % — schema drifted, re-measure', n;
  END IF;

  -- ── (a) content-block-about: source flip + gate ────────────────────────────
  UPDATE content_components SET input_schema = jsonb_set(input_schema,
      '{fields,cta_url}',
      '{"type":"url","source":"renderer","required":false,"on_missing":"skip_field",
        "llm_guidance":"CTA destination. Resolver-owned (source:renderer) — never LLM-authored; this component is in ctaFieldNames and chooseCTATargets writes it each render. Unset means the gated template renders no button (LNK-005). Was llm+required:true, which forced URL fabrication — bugs_open/023 class E."}'::jsonb),
    html_template = replace(html_template,
      '<a href="{{.cta_url}}" class="about-cta">{{.cta_label}}</a>',
      '{{if .cta_url}}<a href="{{.cta_url}}" class="about-cta">{{.cta_label}}</a>{{end}}'),
    updated_at = now()
  WHERE name='content-block-about';
  GET DIAGNOSTICS changed = ROW_COUNT;
  IF changed <> 1 THEN RAISE EXCEPTION 'content-block-about update touched % rows', changed; END IF;

  -- ── (b) tool-cta: both sources flipped + both anchors gated ────────────────
  UPDATE content_components SET input_schema = jsonb_set(jsonb_set(input_schema,
      '{fields,primary_cta_url}',
      '{"type":"url","source":"renderer","required":false,"on_missing":"skip_field",
        "llm_guidance":"Primary CTA destination. Resolver-owned — never LLM-authored. This component is NOT in ctaFieldNames today, so this stays unset and the gated template renders no button; that is correct (LNK-005) and better than the 404s it emitted while llm+required. bugs_open/023 class E."}'::jsonb),
      '{fields,secondary_cta_url}',
      '{"type":"url","source":"renderer","required":false,"on_missing":"skip_field",
        "llm_guidance":"Secondary CTA destination. Resolver-owned — never LLM-authored. See primary_cta_url. bugs_open/023 class E."}'::jsonb),
    html_template = replace(replace(html_template,
      '<a href="{{.primary_cta_url}}" class="tool-cta-btn-primary">{{.primary_cta_label}}</a>',
      '{{if .primary_cta_url}}<a href="{{.primary_cta_url}}" class="tool-cta-btn-primary">{{.primary_cta_label}}</a>{{end}}'),
      '<a href="{{.secondary_cta_url}}" class="tool-cta-btn-secondary">{{.secondary_cta_label}}</a>',
      '{{if .secondary_cta_url}}<a href="{{.secondary_cta_url}}" class="tool-cta-btn-secondary">{{.secondary_cta_label}}</a>{{end}}'),
    updated_at = now()
  WHERE name='tool-cta';
  GET DIAGNOSTICS changed = ROW_COUNT;
  IF changed <> 1 THEN RAISE EXCEPTION 'tool-cta update touched % rows', changed; END IF;

  -- ── (c) platform-comparison: GATE ONLY — schema intentionally untouched ────
  UPDATE content_components SET html_template = replace(html_template,
      '<a href="{{.cta_url}}" class="pc-cta">{{.cta_label}}</a>',
      '{{if .cta_url}}<a href="{{.cta_url}}" class="pc-cta">{{.cta_label}}</a>{{end}}'),
    updated_at = now()
  WHERE name='platform-comparison';
  GET DIAGNOSTICS changed = ROW_COUNT;
  IF changed <> 1 THEN RAISE EXCEPTION 'platform-comparison update touched % rows', changed; END IF;

  -- ── post-conditions ────────────────────────────────────────────────────────
  -- every one of the four anchors is now gated
  SELECT count(*) INTO n FROM content_components
  WHERE (name='content-block-about' AND html_template LIKE '%{{if .cta_url}}<a href="{{.cta_url}}" class="about-cta">%')
     OR (name='tool-cta' AND html_template LIKE '%{{if .primary_cta_url}}%' AND html_template LIKE '%{{if .secondary_cta_url}}%')
     OR (name='platform-comparison' AND html_template LIKE '%{{if .cta_url}}<a href="{{.cta_url}}" class="pc-cta">%');
  IF n <> 3 THEN RAISE EXCEPTION 'post-condition failed: % of 3 components gated', n; END IF;

  -- the four ungated needles are gone (checked by exact string, NOT by a
  -- "no ungated url anchor anywhere" regex — tool-cta also contains
  --   {{range .items}}<a href="{{.url}}" class="tool-cta-card">
  -- which is a RANGE-SCOPED item link from a query-provided list, not a CTA
  -- label/url pair. It is a different class with a different fix and is
  -- deliberately out of scope here; a blanket regex would fail on it and roll
  -- back the whole migration. 15 such anchors exist across 11 components.)
  SELECT count(*) INTO n FROM content_components
  WHERE (name='content-block-about' AND position('<a href="{{.cta_url}}" class="about-cta">' in html_template) > 0
         AND position('{{if .cta_url}}<a href="{{.cta_url}}" class="about-cta">' in html_template) = 0)
     OR (name='tool-cta' AND position('{{if .primary_cta_url}}' in html_template) = 0)
     OR (name='tool-cta' AND position('{{if .secondary_cta_url}}' in html_template) = 0)
     OR (name='platform-comparison' AND position('<a href="{{.cta_url}}" class="pc-cta">' in html_template) > 0
         AND position('{{if .cta_url}}<a href="{{.cta_url}}" class="pc-cta">' in html_template) = 0);
  IF n <> 0 THEN RAISE EXCEPTION 'post-condition failed: % ungated CTA anchor(s) remain', n; END IF;

  -- the three flipped fields are renderer/optional; platform-comparison is deliberately still llm
  SELECT count(*) INTO n
  FROM content_components c, jsonb_each(c.input_schema->'fields') e(k,v)
  WHERE (c.name,e.k) IN (('content-block-about','cta_url'),('tool-cta','primary_cta_url'),('tool-cta','secondary_cta_url'))
    AND v->>'source'='renderer' AND v->>'required'='false';
  IF n <> 3 THEN RAISE EXCEPTION 'post-condition failed: % of 3 url fields are renderer/optional', n; END IF;

  SELECT count(*) INTO n
  FROM content_components c, jsonb_each(c.input_schema->'fields') e(k,v)
  WHERE c.name='platform-comparison' AND e.k='cta_url' AND v->>'source'='llm';
  IF n <> 1 THEN RAISE EXCEPTION 'post-condition failed: platform-comparison.cta_url should still be llm (deliberate, see header)'; END IF;
END $$;

INSERT INTO schema_migrations (filename, notes)
VALUES ('181_class_e_live_cta_url_integrity.sql',
        'bugs_open/023 classes C+E on the LIVE components (17 placements): gate all four CTA anchors in content-block-about / tool-cta / platform-comparison; flip cta_url, primary_cta_url, secondary_cta_url from source:llm+required:true to renderer+optional. platform-comparison.cta_url deliberately NOT flipped — its sole live value is a working page and nothing would repopulate it; deferred to schema-derived pairing (council trail 2525f980).');

COMMIT;

-- Post-apply, for the record:
--   SELECT c.name, e.k, e.v->>'source', e.v->>'required'
--     FROM content_components c, jsonb_each(c.input_schema->'fields') e(k,v)
--    WHERE c.name IN ('content-block-about','tool-cta','platform-comparison') AND e.k ~ 'url';
--   -- and re-run the real parse (RUNBOOK R9b): ungated anchors should fall 171 -> 167.
