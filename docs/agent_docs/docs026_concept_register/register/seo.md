# Register — seo

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

1 concept, consolidated from 2 raw extractions across unit U21. (The cluster input file contained this category's raw block twice, back-to-back and byte-identical; merged into one entry below.)

### SEO-001 — SEO content agent
- **status:** aspirational
- **status-evidence:** docs017/019b "seo-content-agent | LLM for generation, algorithmic for validation | New — runs after page content is written"; seo-discovery-agent in maintenance Phase 0; slot exists in component-builder-v2 sketch.
- **what:** A post-content sweep owning meta titles/descriptions, structured data/JSON-LD, robots directives, canonical URLs and Open Graph across all pages, with algorithmic validation and LLM generation; complemented in maintenance by sitemap-sync, schema validation, and meta-freshness discovery plus sitemap-regenerator and schema-fixer fix agents. No dedicated SEO category exists in the current taxonomy despite recurring SEO responsibilities across eras.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#SEO-Content-Agent, #Fix-Agents
- **relations:** meta-manager (docs018/007); link technical types; site-finalizer sitemap generation; MKT-001 marketing as work items
- **verify-later:** any seo agent definitions; sitemap.xml generation code path
