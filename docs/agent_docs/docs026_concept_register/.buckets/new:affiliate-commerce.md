
<!-- SOURCE: U10_imagery.md -->
### Affiliate sites programme and the query.affiliate_products resolver gap
- **category:** NEW:affiliate-commerce
- **status-signal:** aspirational
- **status-evidence:** "This is not the active workstream right now — a holding doc" (2026-05-12); affiliate_products "Zero rows today"; resolver "a wired socket with no plug".
- **what:** The affiliate vision (boxing tickets, darts gear, lead-gen) with three vertical shapes (pure-product / event-ticket / lead-generation) and a layered build path (one product on one page → ingestion + editorial enrichment → imagery via illustrations → event/lead verticals). Substantial scaffolding exists — affiliate_products/affiliate_programs tables, five product components (product-card-with-cta declares `source: query.affiliate_products` with typed image_url; product-specs schema effectively empty), link_registry disclosure flags, the med-* scraper family as an ingestion model — but no program integration, no resolver populating the declared source, no editorial pipeline, no calendar/event infrastructure.
- **sources:** old/STATUS_affiliate_sites_2026-05-12.md, STATUS_imagery_2026-05-12.md#Component-audit-finding, FOCUS_imagery_assessment_1_.md#3.2
- **relations:** product illustration plugs in as a resolver precedence rule; link-management (doc 024); vet-med-pricing med-* agents as pattern.
- **verify-later:** affiliate_products row count; any resolver handling query.affiliate_products in queryresolve/sourceResolver.

<!-- SOURCE: U10_imagery.md -->
### Affiliate sites programme and the query.affiliate_products resolver gap
- **category:** NEW:affiliate-commerce
- **status-signal:** aspirational
- **status-evidence:** "This is not the active workstream right now — a holding doc" (2026-05-12); affiliate_products "Zero rows today"; resolver "a wired socket with no plug".
- **what:** The affiliate vision (boxing tickets, darts gear, lead-gen) with three vertical shapes (pure-product / event-ticket / lead-generation) and a layered build path (one product on one page → ingestion + editorial enrichment → imagery via illustrations → event/lead verticals). Substantial scaffolding exists — affiliate_products/affiliate_programs tables, five product components (product-card-with-cta declares `source: query.affiliate_products` with typed image_url; product-specs schema effectively empty), link_registry disclosure flags, the med-* scraper family as an ingestion model — but no program integration, no resolver populating the declared source, no editorial pipeline, no calendar/event infrastructure.
- **sources:** old/STATUS_affiliate_sites_2026-05-12.md, STATUS_imagery_2026-05-12.md#Component-audit-finding, FOCUS_imagery_assessment_1_.md#3.2
- **relations:** product illustration plugs in as a resolver precedence rule; link-management (doc 024); vet-med-pricing med-* agents as pattern.
- **verify-later:** affiliate_products row count; any resolver handling query.affiliate_products in queryresolve/sourceResolver.
