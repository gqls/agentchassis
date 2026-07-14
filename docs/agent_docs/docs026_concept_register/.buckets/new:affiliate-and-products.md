
<!-- SOURCE: U19_sql_tables_components.md -->
### Affiliate and products domain
- **category:** NEW:affiliate-and-products
- **status-signal:** partial
- **status-evidence:** Full schema in 004 (products, product_assets, affiliate_programs, affiliate_products, link_registry.affiliate_product_id + requires_disclosure); 043 (2026) still references "the affiliate_products resolver" as the source of product imagery, so the domain is alive but no seeds/operations appear in this unit.
- **what:** Commerce layer: first-party products (pricing incl. price_display "From £99", SEO fields, per-site slug uniqueness) with asset junctions; affiliate networks (tracking param templates, commission terms, API refs) and affiliate_products with cached network data + custom editorial overlay (pros/cons/verdict/rating, content_status cached→enhanced→reviewed) and availability checking. Link registry marks affiliate links and FTC/ASA disclosure requirements.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART7-11; docs/agent_docs/sql_for_tables/043_site_plan_imagery.sql#kind-comment
- **relations:** link-management (registry); product-card/product-grid library components; imagery (product images excluded from planner).
- **verify-later:** affiliate_products resolver code; any populated programs.

<!-- SOURCE: U19_sql_tables_components.md -->
### Affiliate and products domain
- **category:** NEW:affiliate-and-products
- **status-signal:** partial
- **status-evidence:** Full schema in 004 (products, product_assets, affiliate_programs, affiliate_products, link_registry.affiliate_product_id + requires_disclosure); 043 (2026) still references "the affiliate_products resolver" as the source of product imagery, so the domain is alive but no seeds/operations appear in this unit.
- **what:** Commerce layer: first-party products (pricing incl. price_display "From £99", SEO fields, per-site slug uniqueness) with asset junctions; affiliate networks (tracking param templates, commission terms, API refs) and affiliate_products with cached network data + custom editorial overlay (pros/cons/verdict/rating, content_status cached→enhanced→reviewed) and availability checking. Link registry marks affiliate links and FTC/ASA disclosure requirements.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART7-11; docs/agent_docs/sql_for_tables/043_site_plan_imagery.sql#kind-comment
- **relations:** link-management (registry); product-card/product-grid library components; imagery (product images excluded from planner).
- **verify-later:** affiliate_products resolver code; any populated programs.
