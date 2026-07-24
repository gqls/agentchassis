-- ============================================================================
-- 202_about_commercial_block_component.sql
-- COMPONENT: about-commercial-block — the quiet commercial signals block for
-- portfolio/storefront about pages (about_page_commercial workstream, Phase 1).
--
-- Carries up to three gated lines, ALL facts resolver-owned from the
-- site_specs 'commercial' aspect (source: site_specs.commercial.* — resolved in
-- plan_sections, merges LAST, so the copy LLM can never author or override
-- them; see resolve_internal_links_action.go:92-97 contract):
--   1. built-by  : class ∈ {portfolio, storefront}
--   2. for-sale  : class=portfolio AND for_sale_requested AND NOT advertising_active
--   3. advertise : class=portfolio AND inventory_open AND NOT for_sale_requested
-- Gates are computed IN-TEMPLATE from raw facts (never precomputed show_*
-- booleans — write-time derivation is the staleness/write-order bug the design
-- exists to prevent; PLAN D5/D6). Absent aspect => every gate false => renders
-- an empty shell only if planned onto the page at all (plan-level gating is the
-- primary fail-closed layer; these gates are defence-in-depth).
--
-- Copy is the LOCKED register (PLAN D8-D12): "available to acquire",
-- representation not adjectives, no prices, no "premium"/"serious offers".
-- Tier ("1"|"2"|"3" — STRING) picks the for-sale wording.
--
-- TEMPLATE RULES honoured (see RUNBOOK gotchas):
--   - Go text/template; custom eq/ne are STRING-comparing (call_agent.go:1160)
--     so {{if eq .tier "1"}} is safe whatever JSON type tier arrives as.
--   - Only funcs: if/else/range/with/and/or/not + eq/ne/default/lower/upper/
--     isset/safe. Anything else => Parse error => silent regex-renderer
--     fallback that mangles {{else if}} blocks.
--   - missingkey=zero => absent field is falsy => gates fail closed.
--   - Literal closing </section> kept (truncation guard requires it).
--   - One llm field (heading) so content_data is never empty (empty
--     content_data trips the rerender escalation, rerender:186-230).
-- ============================================================================

DO $$
DECLARE
  colliding int;
  candidates int;
BEGIN
  -- needle-gate 1: nothing already occupies the name/function
  SELECT count(*) INTO colliding
  FROM content_components
  WHERE function = 'about-commercial-block' OR name = 'about-commercial-block';
  IF colliding <> 0 THEN
    RAISE EXCEPTION 'about-commercial-block already exists (% row(s)) — another thread got here first; read about_page_commercial/NOTES before re-running', colliding;
  END IF;

  -- INSERT (named-column shape, mirrors 183)
  INSERT INTO content_components (
    name, display_name, description, function, section_type, component_level,
    render_mode, category, is_active, is_dark_section, forked_from,
    suitable_site_types, suitable_page_types, semantic_tags, visual_density,
    usage_count, created_from, input_schema, html_template
  ) VALUES (
    'about-commercial-block',
    'About Commercial Block',
    'Quiet commercial-signals block for the foot of a portfolio/storefront about page: built-by line, tier-worded domain-acquisition line, advertise line. All facts resolver-owned from site_specs.commercial.*; visibility gates computed in-template from raw facts (fail-closed). Built for the about_page_commercial workstream.',
    'about-commercial-block',
    'about-commercial-block',
    'section',
    'template',
    'content',
    true,
    false,
    NULL,
    '["portfolio", "storefront", "brochure", "interactive-platform", "authority-portal", "hub", "ecommerce"]'::jsonb,
    '["about", "content"]'::jsonb,
    '["about", "commercial", "footer-adjacent"]'::jsonb,
    'low',
    0,
    'manual',
    $schema$
{
  "fields": {
    "heading": {
      "type": "text",
      "source": "llm",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Short neutral heading for this quiet end-of-page block, e.g. 'About this site'. Two or three plain words. No hype, no marketing language."
    },
    "class": {
      "type": "text",
      "source": "site_specs.commercial.class",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Resolver-owned commercial class (keeper|storefront|portfolio). Never LLM-authored."
    },
    "tier": {
      "type": "text",
      "source": "site_specs.commercial.tier",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Resolver-owned domain-value tier string 1|2|3. Never LLM-authored."
    },
    "domain": {
      "type": "text",
      "source": "site_specs.commercial.domain",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Resolver-owned domain name. Never LLM-authored."
    },
    "for_sale_requested": {
      "type": "text",
      "source": "site_specs.commercial.for_sale_requested",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Resolver-owned boolean fact. Never LLM-authored."
    },
    "advertising_active": {
      "type": "text",
      "source": "site_specs.commercial.advertising_active",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Resolver-owned boolean fact. Never LLM-authored."
    },
    "inventory_open": {
      "type": "text",
      "source": "site_specs.commercial.inventory_open",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Resolver-owned boolean fact. Never LLM-authored."
    },
    "marketplace_url": {
      "type": "url",
      "source": "site_specs.commercial.marketplace_url",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Resolver-owned acquisition destination (Afternic). Never LLM-authored. Absent => the gated template renders no link."
    },
    "advertise_url": {
      "type": "url",
      "source": "site_specs.commercial.advertise_url",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Resolver-owned advertise destination (advertise.co.uk). Never LLM-authored. Absent => no link."
    },
    "built_by_url": {
      "type": "url",
      "source": "site_specs.commercial.built_by_url",
      "required": false,
      "on_missing": "skip_field",
      "llm_guidance": "Resolver-owned storefront URL (fundamentallyai.com). Never LLM-authored. Absent => built-by line renders without links."
    }
  }
}
$schema$::jsonb,
    $html$<style>
.about-commercial-block{border-top:1px solid var(--color-border,rgba(128,128,128,.25));padding:2.5rem 1.5rem;font-size:.9rem;line-height:1.6;color:var(--color-text-secondary,#6b7280)}
.about-commercial-block .acb-inner{max-width:var(--max-width,72rem);margin:0 auto;display:flex;flex-direction:column;gap:.55rem}
.about-commercial-block .acb-heading{font-size:1rem;font-weight:600;margin:0 0 .3rem;color:var(--color-text,inherit)}
.about-commercial-block p{margin:0}
.about-commercial-block a{color:var(--color-primary,inherit);text-decoration:underline;text-underline-offset:2px}
</style>
<section class="about-commercial-block" data-component="about-commercial-block">
  <div class="acb-inner">
    {{if or (eq .class "portfolio") (eq .class "storefront")}}
    {{if .heading}}<p class="acb-heading">{{.heading}}</p>{{end}}
    <p class="acb-builtby">Built by {{if .built_by_url}}<a href="{{.built_by_url}}" rel="nofollow">fundamentallyai.com</a>{{else}}fundamentallyai.com{{end}}. We design and build sites like this one{{if .built_by_url}} &mdash; <a href="{{.built_by_url}}" rel="nofollow">see how it's done</a>{{end}}.</p>
    {{if and (eq .class "portfolio") .for_sale_requested (not .advertising_active)}}
    <p class="acb-acquire">{{if eq .tier "1"}}The {{.domain}} name is available to acquire &mdash; acquisition enquiries via {{if .marketplace_url}}<a href="{{.marketplace_url}}" rel="nofollow">our domain team</a>{{else}}our domain team{{end}}.{{else if eq .tier "2"}}The {{.domain}} name is available to acquire{{if .marketplace_url}} &mdash; <a href="{{.marketplace_url}}" rel="nofollow">register your interest</a>{{end}}.{{else}}{{.domain}} is part of our portfolio and may be available to acquire{{if .marketplace_url}} &mdash; <a href="{{.marketplace_url}}" rel="nofollow">make an enquiry</a>{{end}}.{{end}}</p>
    {{end}}
    {{if and (eq .class "portfolio") .inventory_open (not .for_sale_requested)}}
    <p class="acb-advertise">Advertise on {{.domain}}. A small number of sponsored placements are available on this site &mdash; a flat monthly rate, set up in minutes{{if .advertise_url}}. <a href="{{.advertise_url}}" rel="nofollow">Advertise here</a>{{else}}.{{end}}</p>
    {{end}}
    {{end}}
  </div>
</section>$html$
  );

  -- needle-gate 2 (post): we are the SOLE selector candidate for this type
  SELECT count(*) INTO candidates
  FROM content_components
  WHERE section_type = 'about-commercial-block' AND component_level = 'section'
    AND is_active = true AND forked_from IS NULL;
  IF candidates <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 about-commercial-block selector candidate after insert, found %', candidates;
  END IF;

  RAISE NOTICE 'about-commercial-block inserted; sole selector candidate confirmed';
END $$;
