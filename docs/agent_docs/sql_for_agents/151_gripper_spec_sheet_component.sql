-- 151: gripper-spec-sheet content_component + query.products resolver
--
-- Context (2026-07-14, robot-hands Phase 4 — category-wrong product pages):
-- robot-hands.com is a spec/comparison site; its two product pages served
-- Add-to-Cart/Buy-Now e-commerce furniture regardless of whether the
-- (missing) data ever got filled — category-wrong for a site that sells
-- nothing. Owner decision: replace with a spec-sheet component sourced from
-- REAL, verified gripper manufacturer data (no fabrication).
--
-- Feasibility note: the original handoff's "dartsonline's 14 real product
-- cards prove the pipeline is sound" does not hold up — dartsonline's
-- product-grid sources `query.products`, but queryresolve.go's Resolve()
-- only implemented pages_where_type/pages_under_section before this
-- migration. dartsonline's cards are frozen rendered_html, not a live
-- mechanism. This migration adds the missing live resolver
-- (resolveProducts, in queryresolve.go) — so gripper-spec-sheet does NOT
-- inherit dartsonline's staleness risk, and dartsonline's product-grid
-- becomes render-safe again as a side effect (same query name, same table).
--
-- Requires the chassis image with resolveProducts registered (case
-- "products" in queryresolve.Resolve).
--
-- Companion: 152_gripper_products_seed.sql (the real product rows),
-- 153_gripper_detail_page_swap.sql (removes the old e-commerce components,
-- installs this one on gripper-detail / product-detail).
--
-- Verify after applying:
--   SELECT name, function, category FROM content_components
--   WHERE name = 'gripper-spec-sheet';

BEGIN;

INSERT INTO content_components (
    name, display_name, description, html_template, input_schema,
    function, category, section_type, suitable_page_types,
    content_shape, visual_density, data_sources, component_level,
    render_mode, is_active
) VALUES (
    'gripper-spec-sheet',
    'Gripper Spec Sheet',
    'Real, sourced gripper product specifications — no cart/buy furniture. Every field optional (manufacturer spec pages disclose different subsets); each card carries a source URL and verified date.',
    $TPL$<style>
.gripper-spec-sheet-section {
  padding: var(--spacing-section, 4rem 1.5rem);
}
.gripper-spec-sheet-section .gss-container {
  max-width: var(--container-max-width, 1200px);
  margin: 0 auto;
}
.gripper-spec-sheet-section .gss-title {
  color: var(--section-heading, var(--color-heading, #0f172a));
  font-size: 1.5rem;
  font-weight: 700;
  margin: 0 0 1.5rem;
}
.gripper-spec-sheet-section .gss-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1.25rem;
}
.gripper-spec-sheet-section .gss-card {
  background: var(--color-card-bg, #ffffff);
  border: 1px solid var(--color-border, #e2e8f0);
  border-radius: var(--border-radius, 0.5rem);
  padding: 1.25rem;
}
.gripper-spec-sheet-section .gss-card-manufacturer {
  color: var(--color-text-muted, #64748b);
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  margin: 0 0 0.25rem;
}
.gripper-spec-sheet-section .gss-card-name {
  color: var(--section-heading, var(--color-heading, #0f172a));
  font-size: 1.0625rem;
  font-weight: 700;
  margin: 0 0 0.25rem;
  line-height: 1.3;
}
.gripper-spec-sheet-section .gss-card-subcategory {
  color: var(--color-text-muted, #64748b);
  font-size: 0.8125rem;
  margin: 0 0 0.875rem;
}
.gripper-spec-sheet-section .gss-spec-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
  margin-bottom: 0.875rem;
}
.gripper-spec-sheet-section .gss-spec-table tr {
  border-bottom: 1px solid var(--color-border, #f1f5f9);
}
.gripper-spec-sheet-section .gss-spec-table tr:last-child {
  border-bottom: none;
}
.gripper-spec-sheet-section .gss-spec-table td {
  padding: 0.375rem 0;
}
.gripper-spec-sheet-section .gss-spec-label {
  color: var(--color-text-muted, #64748b);
  width: 45%;
}
.gripper-spec-sheet-section .gss-spec-value {
  color: var(--section-heading, var(--color-heading, #0f172a));
  font-weight: 600;
  text-align: right;
}
.gripper-spec-sheet-section .gss-card-price {
  font-size: 0.8125rem;
  color: var(--color-text-muted, #64748b);
  font-style: italic;
  margin-bottom: 0.75rem;
}
.gripper-spec-sheet-section .gss-card-source {
  font-size: 0.6875rem;
  color: var(--color-text-muted, #94a3b8);
  border-top: 1px solid var(--color-border, #f1f5f9);
  padding-top: 0.625rem;
}
.gripper-spec-sheet-section .gss-card-source a {
  color: inherit;
}
@media (max-width: 768px) {
  .gripper-spec-sheet-section .gss-grid { grid-template-columns: 1fr; }
}
</style>
<section class="gripper-spec-sheet-section" data-component="gripper-spec-sheet">
  <div class="gss-container">
    {{if .headline}}<h2 class="gss-title">{{.headline}}</h2>{{end}}
    <div class="gss-grid">
      {{range .products}}
      <div class="gss-card">
        {{if .manufacturer}}<p class="gss-card-manufacturer">{{.manufacturer}}</p>{{end}}
        <h3 class="gss-card-name">{{.name}}</h3>
        {{if .subcategory}}<p class="gss-card-subcategory">{{.subcategory}}</p>{{end}}
        <table class="gss-spec-table">
          {{if .stroke}}<tr><td class="gss-spec-label">Stroke</td><td class="gss-spec-value">{{.stroke}}</td></tr>{{end}}
          {{if .gripping_force}}<tr><td class="gss-spec-label">Gripping force</td><td class="gss-spec-value">{{.gripping_force}}</td></tr>{{end}}
          {{if .payload}}<tr><td class="gss-spec-label">Payload</td><td class="gss-spec-value">{{.payload}}</td></tr>{{end}}
          {{if .weight}}<tr><td class="gss-spec-label">Weight</td><td class="gss-spec-value">{{.weight}}</td></tr>{{end}}
          {{if .ip_rating}}<tr><td class="gss-spec-label">IP rating</td><td class="gss-spec-value">{{.ip_rating}}</td></tr>{{end}}
          {{if .interface}}<tr><td class="gss-spec-label">Interface</td><td class="gss-spec-value">{{.interface}}</td></tr>{{end}}
          {{if .voltage}}<tr><td class="gss-spec-label">Voltage</td><td class="gss-spec-value">{{.voltage}}</td></tr>{{end}}
        </table>
        <p class="gss-card-price">{{if .price_display}}{{.price_display}}{{else}}Contact manufacturer for pricing{{end}}</p>
        {{if .source_url}}<p class="gss-card-source">Source: <a href="{{.source_url}}" target="_blank" rel="noopener nofollow">manufacturer datasheet</a>{{if .verified_date}} · Verified {{.verified_date}}{{end}}</p>{{end}}
      </div>
      {{end}}
    </div>
  </div>
</section>
$TPL$,
    '{
      "fields": {
        "headline": {
          "type": "text",
          "source": "llm",
          "required": false,
          "on_missing": "skip_field",
          "llm_guidance": "Optional section heading, e.g. Gripper Specifications or Comparable Models"
        },
        "products": {
          "type": "array",
          "items": {
            "name": {"type": "text"},
            "manufacturer": {"type": "text"},
            "subcategory": {"type": "text"},
            "stroke": {"type": "text"},
            "gripping_force": {"type": "text"},
            "payload": {"type": "text"},
            "weight": {"type": "text"},
            "ip_rating": {"type": "text"},
            "interface": {"type": "text"},
            "voltage": {"type": "text"},
            "price_display": {"type": "text"},
            "source_url": {"type": "text"},
            "verified_date": {"type": "text"}
          },
          "source": "query.products:gripper",
          "required": true,
          "min_items": 1,
          "on_missing": "skip_section"
        }
      }
    }'::jsonb,
    'gripper-spec-sheet',
    'content',
    'spec-sheet',
    '["entity-page", "product-listing", "category"]'::jsonb,
    'structured_card',
    'medium',
    ARRAY['products'],
    'section',
    'template',
    true
);

COMMIT;
