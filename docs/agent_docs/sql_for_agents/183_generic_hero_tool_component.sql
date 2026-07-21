-- 183_generic_hero_tool_component.sql — 2026-07-21, hero_tool_component_045 (bugs_open/045)
--
-- THE BUG (measured, not inferred — full evidence in bugs_open/045):
--   A page plan asks for a generic `hero-tool` section. The selector
--   (component_selector.go:queryCandidates) matches candidates by
--   `section_type = 'hero-tool' AND component_level='section' AND is_active
--    AND forked_from IS NULL`, scores them, and takes candidates[0]. There is
--   exactly ONE such row — `bayesian-ranking-hero-tool_pre_037` — and it is a
--   hero for a Bayesian ranking product, with 14 `source:static` fields whose
--   Bayesian fallbacks re-apply on every render and cannot be overridden by
--   content_data (plan_sections_action.go static branch). So every tool page
--   that requests `hero-tool` gets "Start Ranking Free" / "Calculate Rankings"
--   / "Try the Bayesian Ranker", whatever the page is actually about. This is
--   NOT a planner bug (the plan asked for the string "hero-tool") and NOT a
--   selector bug (it correctly resolves the only match). It is a LIBRARY GAP:
--   there is no neutral tool hero.
--
-- THE FIX (bugs_open/045 candidates 1 + 2, done atomically):
--   1. INSERT a generic `hero-tool` component (function=section_type=hero-tool):
--        * NO product-specific `source:static` fallbacks — every visible label
--          is `source:llm` (so it speaks the page's own vocabulary) or absent.
--        * CTA anchors GATED ({{if .x_url}}) and *_url fields source:renderer
--          (resolver-owned, never LLM-authored) — satisfies LNK-005 by
--          construction and derives cleanly under schema-derived CTA pairing
--          (bugs_open/023 class H). Same shape as tool-guide-intro after
--          migration 179.
--        * Trust stats are OPTIONAL and gated, with anti-fabrication guidance
--          (bugs_open/043): omit rather than invent a figure.
--        * No embedded product widget — the actual tool is a separate
--          `tool-<slug>` section on the page; the hero only introduces it.
--   2. RETIRE the Bayesian row from the hero-tool selector pool by moving its
--      section_type 'hero-tool' -> 'bayesian-ranking-hero-tool' (its own
--      function). It stays is_active=true and its function is UNCHANGED — this
--      is a supersede, NOT a delete (bugs_open/045 landmine + 023 R10: it is
--      the sole active row for its function; deleting it deletes the live
--      component). After this, the generic component is the SOLE candidate for
--      `hero-tool`, so selection is deterministic (SelectComponentByType has no
--      score threshold — candidates[0] wins).
--
-- BLAST RADIUS / SAFETY:
--   * DB config is live immediately; this changes what the NEXT plan/rebuild of
--     any `hero-tool` page resolves to. It touches NO deployed page: page_components
--     rows bake component_id + rendered_html, and nothing here rewrites them.
--   * Three pages currently request `hero-tool`:
--       - finetuning.uk/ai-agent-roi-estimator        (needs_rebuild, no placement) → rebuilds to generic ✔
--       - ai-agent-orchestration.com/agent-complexity-estimator (needs_rebuild, no placement) → rebuilds to generic ✔
--       - gamesdesign.co.uk/bayesian-ranking           (DEPLOYED, live Bayesian hero at position 1)
--     The gamesdesign page is untouched (deployed). On a FUTURE rebuild its hero
--     goes generic; its ranking FUNCTION is unaffected — that lives in the
--     separate `tool-bayesian-ranking` (component_level='tool') section at
--     position 3, verified present 2026-07-21. The embedded hero widget merely
--     duplicated it.
--
-- ROLLBACK: DELETE FROM content_components WHERE name='hero-tool';
--           UPDATE content_components SET section_type='hero-tool'
--             WHERE id='7cd0408b-5614-4d80-9fe9-9ecddd8ee110';
--   (snapshot of the Bayesian row taken below as bak_bayesian_hero_tool_20260721.)

\set ON_ERROR_STOP on
BEGIN;

-- snapshot the row we are about to re-point (bak_ convention)
CREATE TABLE bak_bayesian_hero_tool_20260721 AS
SELECT * FROM content_components WHERE id = '7cd0408b-5614-4d80-9fe9-9ecddd8ee110';

DO $mig$
DECLARE
  bayesian_section_type text;
  bayesian_active       boolean;
  colliding             int;
  hero_tool_candidates  int;
  new_tpl               text;
  new_id                uuid;
BEGIN
  -- ── needle-gate 0: the Bayesian row is present and still the hero-tool row ──
  SELECT section_type, is_active INTO bayesian_section_type, bayesian_active
  FROM content_components WHERE id = '7cd0408b-5614-4d80-9fe9-9ecddd8ee110';
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Bayesian hero-tool row 7cd0408b... not found — component drifted since 2026-07-21; re-derive before applying';
  END IF;
  IF bayesian_section_type <> 'hero-tool' THEN
    RAISE EXCEPTION 'Bayesian row section_type is % (expected hero-tool) — another thread may already have retired it; re-read bugs_open/045 before applying', bayesian_section_type;
  END IF;

  -- ── needle-gate 1: nothing already occupies function/name 'hero-tool' ──
  SELECT count(*) INTO colliding
  FROM content_components WHERE function = 'hero-tool' OR name = 'hero-tool';
  IF colliding <> 0 THEN
    RAISE EXCEPTION 'a component with function/name hero-tool already exists (% row(s)) — another thread may have built the generic hero; re-read bugs_open/045', colliding;
  END IF;

  -- ── needle-gate 2: the Bayesian row is the SOLE hero-tool selector candidate ──
  -- (mirrors component_selector.go queryCandidates exactly)
  SELECT count(*) INTO hero_tool_candidates
  FROM content_components
  WHERE section_type = 'hero-tool' AND component_level = 'section'
    AND is_active = true AND forked_from IS NULL;
  IF hero_tool_candidates <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 active hero-tool selector candidate, found % — library shifted; re-measure before applying', hero_tool_candidates;
  END IF;

  -- ── (1) INSERT the generic hero-tool component ──
  INSERT INTO content_components (
    name, display_name, description, function, section_type, component_level,
    render_mode, category, is_active, is_dark_section, forked_from,
    suitable_site_types, suitable_page_types, semantic_tags, visual_density,
    usage_count, created_from, input_schema, html_template
  ) VALUES (
    'hero-tool',
    'Tool Hero',
    'Generic hero panel for a tool page — badge, headline, subheadline, gated CTAs and up to three optional trust stats. No product-specific vocabulary; the tool itself is a separate section. Built for bugs_open/045 to replace the Bayesian-hardwired hero.',
    'hero-tool',
    'hero-tool',
    'section',
    'template',
    'hero',
    true,
    true,
    NULL,
    '["tool", "utility", "saas"]'::jsonb,
    '["tool"]'::jsonb,
    '["hero", "tool", "cta"]'::jsonb,
    'medium',
    0,
    'manual',
    $schema$
{
  "fields": {
    "badge_label": {
      "type": "text",
      "source": "llm",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Short badge above the headline, e.g. 'Free Tool' or 'Interactive'. Describe THIS tool in generic terms — never another product's name or vocabulary."
    },
    "hero_headline": {
      "type": "text",
      "source": "llm",
      "required": true,
      "llm_guidance": "Primary hero headline for THIS tool page. Communicate the tool's core value in the page's own terms. Keep it under 12 words."
    },
    "hero_subheadline": {
      "type": "text",
      "source": "llm",
      "required": true,
      "llm_guidance": "Two to three sentences explaining what this tool does and why it is useful. Accessible and specific to this tool."
    },
    "cta_primary_label": {
      "type": "text",
      "source": "llm",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Primary CTA button label for this tool, e.g. 'Try the Tool'. Only rendered when a destination URL is resolved (cta_primary_url is source:renderer)."
    },
    "cta_primary_url": {
      "type": "url",
      "source": "renderer",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Primary CTA destination. Resolver-owned (source:renderer) — never LLM-authored. Unset means the gated template renders no button (LNK-005)."
    },
    "cta_secondary_label": {
      "type": "text",
      "source": "llm",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Secondary CTA button label, e.g. 'How It Works'. Only rendered when a destination URL is resolved."
    },
    "cta_secondary_url": {
      "type": "url",
      "source": "renderer",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Secondary CTA destination. Resolver-owned (source:renderer) — never LLM-authored. Unset means no button."
    },
    "stat_one_value": {
      "type": "text",
      "source": "llm",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "First trust-stat value, e.g. '10K+' or 'Free'. Must be a real, defensible figure for THIS tool — OMIT it rather than invent one (bugs_open/043)."
    },
    "stat_one_label": {
      "type": "text",
      "source": "llm",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Label for the first trust stat, e.g. 'Calculations Run'. Omit if stat_one_value is omitted."
    },
    "stat_two_value": {
      "type": "text",
      "source": "llm",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Second trust-stat value. Real and defensible, or omitted (bugs_open/043)."
    },
    "stat_two_label": {
      "type": "text",
      "source": "llm",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Label for the second trust stat. Omit if stat_two_value is omitted."
    },
    "stat_three_value": {
      "type": "text",
      "source": "llm",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Third trust-stat value. Real and defensible, or omitted (bugs_open/043)."
    },
    "stat_three_label": {
      "type": "text",
      "source": "llm",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Label for the third trust stat. Omit if stat_three_value is omitted."
    }
  }
}
$schema$::jsonb,
    $html$<style>.hero-tool-section{--section-text:rgba(255,255,255,0.9);--section-text-muted:rgba(255,255,255,0.7);--section-heading:#ffffff;--section-surface:rgba(255,255,255,0.05);--section-border:rgba(255,255,255,0.2);background:var(--color-primary,#1a1f36);color:var(--section-text);padding:var(--spacing-section,5rem 2rem);font-family:inherit}.hero-tool-section *{box-sizing:border-box}.hero-tool-section .htl-container{max-width:var(--container-max-width,880px);margin:0 auto;padding:0 1rem;text-align:center;display:flex;flex-direction:column;align-items:center}.hero-tool-section .htl-badge{display:inline-block;background:var(--section-surface);border:1px solid var(--section-border);color:var(--section-text-muted);font-size:0.75rem;font-weight:600;letter-spacing:0.08em;text-transform:uppercase;padding:0.35rem 0.85rem;border-radius:2rem;margin-bottom:1.25rem}.hero-tool-section .htl-headline{font-size:clamp(2rem,4vw,3rem);font-weight:800;color:var(--section-heading);line-height:1.15;margin:0 0 1.25rem;max-width:20ch}.hero-tool-section .htl-subheadline{font-size:1.1rem;color:var(--section-text-muted);line-height:1.7;margin:0 0 2rem;max-width:60ch}.hero-tool-section .htl-cta-row{display:flex;flex-wrap:wrap;gap:0.75rem;align-items:center;justify-content:center}.hero-tool-section .htl-btn-primary{display:inline-flex;align-items:center;gap:0.5rem;background:var(--color-accent,#f59e0b);color:#000;font-weight:700;font-size:1rem;padding:0.85rem 1.75rem;border-radius:var(--border-radius,0.5rem);border:none;cursor:pointer;text-decoration:none;min-height:44px;transition:opacity 0.2s}.hero-tool-section .htl-btn-primary:hover{opacity:0.88}.hero-tool-section .htl-btn-secondary{display:inline-flex;align-items:center;gap:0.5rem;background:transparent;color:var(--section-text);font-weight:600;font-size:1rem;padding:0.85rem 1.75rem;border-radius:var(--border-radius,0.5rem);border:1px solid var(--section-border);cursor:pointer;text-decoration:none;min-height:44px;transition:background 0.2s}.hero-tool-section .htl-btn-secondary:hover{background:var(--section-surface)}.hero-tool-section .htl-trust-row{display:flex;flex-wrap:wrap;gap:1.5rem;margin-top:2.25rem;justify-content:center}.hero-tool-section .htl-trust-item{display:flex;flex-direction:column;gap:0.15rem}.hero-tool-section .htl-trust-value{font-size:1.5rem;font-weight:800;color:var(--section-heading)}.hero-tool-section .htl-trust-label{font-size:0.78rem;color:var(--section-text-muted);text-transform:uppercase;letter-spacing:0.06em}@media(max-width:768px){.hero-tool-section .htl-trust-row{gap:1rem}.hero-tool-section .htl-cta-row{flex-direction:column;align-items:stretch}}</style><section class="hero-tool-section" data-component="hero-tool"><div class="htl-container">{{if .badge_label}}<span class="htl-badge">{{.badge_label}}</span>{{end}}<h1 class="htl-headline">{{.hero_headline}}</h1><p class="htl-subheadline">{{.hero_subheadline}}</p>{{if or .cta_primary_url .cta_secondary_url}}<div class="htl-cta-row">{{if .cta_primary_url}}<a href="{{.cta_primary_url}}" class="htl-btn-primary">{{.cta_primary_label}}</a>{{end}}{{if .cta_secondary_url}}<a href="{{.cta_secondary_url}}" class="htl-btn-secondary">{{.cta_secondary_label}}</a>{{end}}</div>{{end}}{{if or .stat_one_value .stat_two_value .stat_three_value}}<div class="htl-trust-row">{{if .stat_one_value}}<div class="htl-trust-item"><span class="htl-trust-value">{{.stat_one_value}}</span><span class="htl-trust-label">{{.stat_one_label}}</span></div>{{end}}{{if .stat_two_value}}<div class="htl-trust-item"><span class="htl-trust-value">{{.stat_two_value}}</span><span class="htl-trust-label">{{.stat_two_label}}</span></div>{{end}}{{if .stat_three_value}}<div class="htl-trust-item"><span class="htl-trust-value">{{.stat_three_value}}</span><span class="htl-trust-label">{{.stat_three_label}}</span></div>{{end}}</div>{{end}}</div></section>$html$
  )
  RETURNING id INTO new_id;

  -- ── (2) RETIRE the Bayesian row from the hero-tool pool (supersede, not delete) ──
  UPDATE content_components
  SET section_type = 'bayesian-ranking-hero-tool', updated_at = now()
  WHERE id = '7cd0408b-5614-4d80-9fe9-9ecddd8ee110';

  -- ── post-conditions ──
  -- P1: exactly one active hero-tool selector candidate, and it is the new generic one
  SELECT count(*) INTO hero_tool_candidates
  FROM content_components
  WHERE section_type = 'hero-tool' AND component_level = 'section'
    AND is_active = true AND forked_from IS NULL;
  IF hero_tool_candidates <> 1 THEN
    RAISE EXCEPTION 'post-condition P1 failed: % hero-tool candidates after apply (expected 1)', hero_tool_candidates;
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM content_components
    WHERE section_type = 'hero-tool' AND component_level='section' AND is_active AND forked_from IS NULL
      AND function = 'hero-tool' AND name = 'hero-tool'
  ) THEN
    RAISE EXCEPTION 'post-condition P1b failed: the sole hero-tool candidate is not the new generic component';
  END IF;

  -- P2: the Bayesian row is still active, function unchanged, out of the hero-tool pool
  SELECT section_type, is_active INTO bayesian_section_type, bayesian_active
  FROM content_components WHERE id = '7cd0408b-5614-4d80-9fe9-9ecddd8ee110';
  IF bayesian_section_type <> 'bayesian-ranking-hero-tool' OR bayesian_active IS NOT TRUE THEN
    RAISE EXCEPTION 'post-condition P2 failed: Bayesian row section_type=% is_active=% (expected bayesian-ranking-hero-tool / true)', bayesian_section_type, bayesian_active;
  END IF;

  -- P3: the new template is well-formed, carries its data-component, and holds NO Bayesian vocabulary
  SELECT html_template INTO STRICT new_tpl FROM content_components WHERE id = new_id;
  IF position('</section>' in new_tpl) = 0 THEN
    RAISE EXCEPTION 'post-condition P3 failed: new template missing </section> (would trip the truncation guard)';
  END IF;
  IF position('data-component="hero-tool"' in new_tpl) = 0 THEN
    RAISE EXCEPTION 'post-condition P3 failed: new template missing data-component="hero-tool"';
  END IF;
  IF new_tpl ~* '(Bayesian|Ranking Free|Calculate Rankings|Prior belief|Add Item)' THEN
    RAISE EXCEPTION 'post-condition P3 failed: new template contains Bayesian vocabulary';
  END IF;

  -- P4: no field in the new component is source:static (the whole point)
  IF EXISTS (
    SELECT 1 FROM content_components c, jsonb_each(c.input_schema->'fields') e(k,v)
    WHERE c.id = new_id AND e.v->>'source' = 'static'
  ) THEN
    RAISE EXCEPTION 'post-condition P4 failed: new component has a source:static field';
  END IF;

  -- P5: every *_url field is source:renderer + optional (LNK-005 by construction)
  IF EXISTS (
    SELECT 1 FROM content_components c, jsonb_each(c.input_schema->'fields') e(k,v)
    WHERE c.id = new_id AND e.k LIKE '%\_url'
      AND (e.v->>'source' <> 'renderer' OR COALESCE(e.v->>'required','false') <> 'false')
  ) THEN
    RAISE EXCEPTION 'post-condition P5 failed: a *_url field is not renderer/optional';
  END IF;

  RAISE NOTICE 'hero-tool generic component created: %; Bayesian row retired to section_type=bayesian-ranking-hero-tool', new_id;
END $mig$;

-- ledger row in the SAME transaction (bugs_open/007: apply and record in one sitting)
INSERT INTO schema_migrations (filename, notes)
VALUES ('183_generic_hero_tool_component.sql',
        'bugs_open/045: add generic hero-tool component (llm labels, gated renderer CTAs, optional anti-fabrication trust stats, no embedded widget) and retire the Bayesian hero-tool row from the selector pool (section_type -> bayesian-ranking-hero-tool, kept is_active). Deterministic: generic is now the sole hero-tool candidate. Touches no deployed page.');

COMMIT;

-- Post-apply, for the record (run outside the transaction):
--   SELECT name, function, section_type, is_active FROM content_components
--     WHERE section_type IN ('hero-tool','bayesian-ranking-hero-tool') ORDER BY section_type;
--   -- must show: hero-tool | hero-tool | hero-tool | t   AND
--   --            bayesian-ranking-hero-tool_pre_037 | bayesian-ranking-hero-tool | bayesian-ranking-hero-tool | t
