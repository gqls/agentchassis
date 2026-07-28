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


### SEO-002 — Site discovery files generator (`scripts/site-discovery-files.py`)
- **status:** built and exercised — run against 3 sites 2026-07-28; relojistas.com's three files are LIVE
- **status-evidence:** `scripts/site-discovery-files.py`; relojistas.com serves `/sitemap.xml` (18 urls), `/llms.txt` (18 pages), `/robots.txt` (own file, Cloudflare block confirmed off per-agent). Dry-run verified against oufe.com and robot-hands.com.
- **what:** Generates `robots.txt` (Content-Signal + Sitemap pointer), `sitemap.xml` (with lastmod) and `llms.txt` for ANY site, from the `pages` table. Dry-run by default; never commits or deploys. Encodes three rules learned expensively: (1) **probe every URL and list only 200s** — a sitemap advertising a 404 is worse than none, and it caught a dead page on oufe.com on its first fleet run; (2) **llms.txt is built FROM the pages, not written ABOUT them** — each entry is the page's own `<h1>` and own first sentence, so nothing invents a claim about a site; (3) **it detects whether Cloudflare's managed robots.txt is being merged in**, and names the agents currently disallowed — because Cloudflare PREPENDS its file to the origin's rather than yielding, so shipping your own changes nothing until the dashboard setting is off.
- **sources:** docs024_key_docs_latest/traffic_probe/EVIDENCE_2026-07-28_crawl_budget_and_the_dead_forum.md; docs024_key_docs_latest/FLEET_GUIDANCE_discoverability.md
- **relations:** answers SEO-001's `verify-later: sitemap.xml generation code path` — **there was none**; this is it, at script tier. SEO-001's agent remains aspirational. Related: `bugs_open/131` (og:image 404 on 11 of 14 sites), and the dormant `process_html` → `AddStructuredData` path (registered, reachable, emits nothing on any site).
- **verify-later:** whether this should become a Go action beside `render_rss_feed` (same shape: read DB rows, emit a file artefact, gate on `deploy_config`) so the files stay fresh as pages change, rather than a script someone must remember to re-run.
