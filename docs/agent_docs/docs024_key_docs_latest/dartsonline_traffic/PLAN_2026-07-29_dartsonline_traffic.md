# PLAN — dartsonline.com traffic & affiliate-readiness programme

**Started:** 2026-07-29. **Owner brief:** develop dartsonline.com into a leading site for darts
players — traffic-encouraging development, fresh news, analysis, rich imagery, tool-heavy — ready
for an affiliate-feed application. Every fix is backfilled into the framework generically so it
works for any similar domain (boxing, football, …).

Site: `5fe8785b-223d-41a3-88ee-c07187622381`, first site built end-to-end by the work-item relay
(bare on 2026-07-06 — see `../dartsonline.com_site_quality/RUNBOOK_site_quality.md`, this
programme's predecessor; that runbook's baseline is stale, its seven-leg framing still sound).

Full approved plan (session plan file, 2026-07-29): mirrored below in phase order. Evidence for
every claim came from three exploration passes + three design passes against live code/DB on
2026-07-29; the load-bearing citations are inline.

## Owner decisions (2026-07-29, all via direct question)

| # | decision |
|---|---|
| D1 | Affiliate = darts **equipment retail feed** (AWIN/Webgains-class product feeds). NOT betting. `content_direction.editorial.out_of_scope` will exclude betting/odds content. |
| D2 | News = **aggregated feed + our analysis**. Arm the live 6-site aggregation pipeline; original analysis pieces grounded/cited. Do NOT build NEWS-006 (feed-item → written-news articles). |
| D3 | Content cadence = **backfill burst + raised cap**: clear the 9 planned guides quickly (building already-planned pages is not budget-gated), then ~4–5 articles/week (`growth_config.weekly_blog_posts_max: 5`). |
| D4 | Identity = **UK online-only**: contact `darts@contactforsales.com` / `07934 524 911`, no address, no stock-relationship claims; reposition copy as specialist/curated site. Park shipping-returns. |
| D5 | **Tool-heavy: ~1 tool per 6 articles/guides.** Near-term ≈2 tools (setup-builder already planned; checkout calculator the natural second), +1 per ~6 further articles. Generic ratio policy: reuse the existing `evaluate_tools` mechanism if it fits (one open item for this site) before building anything. |

## Phases (dependencies stated; config lanes are live-immediate, code lanes council-gated)

- **P0 — workstream + preflight.** Standing five docs (this dir); `who-owns` on bugs
  083/107/114/117/118/122/125/141; verify bug 141 fix (`fc7c05c21`) in the running chassis pod;
  read the `evaluate_tools` discovery mechanism (D5 reuse check).
- **P1 — foundations (mostly data/config).** (1) Truth reset: supersede `identity` aspect
  (real contact both nested AND flat — bug 072 shape; delete Portland + brand-stocking claims),
  About rebuild via `needs_page` reason `identity_corrected`, copy rails below. (2) Nav
  reconciliation: `/shop.html` `/brands.html` `/guides.html` are STALE rows from superseded plans —
  archive/`in_header=false`; hubs `shop-index`/`brands-index`/`guides-index` into nav. (3) Repoint
  `site_plan_imagery.scope_ref` (shop→shop-index etc.) + INSERT rows for the 4 orphan
  `hero_guide_*` assets BEFORE builds. (4) Backfill `pages.sections`
  `[hero,article-body,call-to-action]` for the 9 blog-posts + hand-create blog-index. (5) Promote
  parked `needs_page` items 2–3 at a time (SKIP grip-styles — bug 107 lane; HOLD entity pages — no
  data); `reconcile_site_plan` for the rest; bug 125 watch (curl canonical AND wrong-path twin).
  (6) Chrome: promote `missing_structure` reassembly (15 pages); bug-118 pre-check before ANY
  chrome refresh; nav rebuild via `nav_drift`→nav-updater. (7) CTA rerenders after hubs 200.
  (8) WCAG hand-fix (our remit per OPEN_THREADS): palette `primary` repoint (verify with
  `platform/colour.AuditPalette`), stale-stylesheet regeneration (trigger UNVERIFIED — trace first),
  `scripts/render_audit.py` before/after. (9) Meta descriptions ×24 (pages + site_plan_pages) +
  `sites.company_name`/`tagline` + chrome `force_rerender` for og tags.
- **P2 — news.** Config: supersede `classification` with `content_features.news_feed`
  (recommended/separate_page/source_types [rss,news_search,api_news]/darts keywords); ORDER =
  spec → 6h tick auto-seeds search+api sources → THEN insert fetch-verified RSS rows (seeder is
  all-or-nothing, `seed_content_sources_action.go:92-111`); one-shot completeness-discovery →
  news-index page (VERIFY row = news-index//news/index.html/page_type exactly, bug 081) + homepage
  latest-news; ~12h two-pass latency is normal; RSS out only when healthy. Code (council):
  `matchVerticalNews` reads `industry_tags` token-exact len≥3 (the "ai" trap) + LLM fallback on map
  miss (autodetect design Phases 0+2; Phase 3 table DEFERRED as architecture-scope; Phase 4
  RSS-discovery agent follow-up). Code (council): role-default sections fallback in
  `load_page_sections_from_spec_action.go` + `check_sectionless_pages` extension (16 pages fleet-wide).
- **P3 — editorial cadence.** `growth_config` INSERT (blog 5/wk); `content_direction` + `editorial`
  key (MUST recompute `formatted` — page-content-writer reads only `.formatted`,
  `site_spec_actions.go:206-216`); blog-content-planner made news-aware + `blog-analysis-refresh`
  scheduled task (SCH-007/SCH-009); grounded-explainer for claim-dense pieces only.
- **P4 — imagery.** `imagery_style_guide` row FIRST (accent-first palette, 200-char cap, per-kind
  icon/content_hero, anchor hero_home after Reading the PNG; never hero-guides — 4-panel collage);
  council point fix `check_undeployed_assets.go:44,51` design/detected → build/triaged (then SQL-promote
  the 17 rows, image first); homepage emoji grid → ADOPT global `image-hover-card-grid` (do NOT touch
  `info-card-grid` — 23 placements/11 sites) + `image_landed` rerender; 6 fresh dark-ground card
  images (the 17 icons are pale-ground — wrong on #111520); article cards via
  `check_content_image_missing` once guides deploy.
- **P5 — SEO.** Discovery files via `scripts/site-discovery-files.py` (coordinate traffic_probe
  lane; Cloudflare zone toggles = owner action); JSON-LD Article via `PageInfo.PageType` (BOTH
  constructors; most careful council submission); bugs 132/116 owned elsewhere — contribute only.
- **P6 — affiliate (ships dark; armed on network acceptance).** Council seam #1: `queryresolve`
  `affiliate_products` case (custom→cached COALESCE, custom_image_id precedence, honest CTA, no fake
  ratings; fail-closed via suppressed_sections self-heal; 0 live consumers measured). Council seam #2:
  `affiliate_feed_actions.go` fetch/normalize/write (field_map config; credentials_ref = env var
  name; upsert never touches custom_*), feed-ingester route + sibling trigger/orchestrator +
  86400s task. Arming layers all default-OFF. Application checklist: guides live, nav 200s, privacy
  page, honest positioning, monitored contact, news live, SEO baseline; supersede `strategy` to
  record affiliate as near-term primary.
- **P-tools (D5, threads through P1/P3).** Build setup-builder via the TOOL pipeline (hold lifted
  07-24 — bugs_closed/020); darts checkout calculator as tool #2; generic ratio policy after
  reading `evaluate_tools` (candidate: `growth_config` ratio field + discovery check).

## Copy rails (D4, binding for every rebuilt page)

MAY say: specialist darts site; spec-first buying guides (tungsten %, barrel weight, flight
shapes); setup help; UK online-only; contact email/phone above.
MAY NOT say: "we stock/carry"; named brand relationships; any address; founding history;
shipping/warehouse claims. Verification grep on every rebuilt page:
`stock|carry|Portland|800\)|darts\.com` → 0 legitimate-context-free hits.

## Corrections to the originating brief (from verification — kept per working-docs rules)

- The 404 nav targets are NOT pages to build — they are stale rows from superseded plans; the
  current plan's hubs are the `*-index` pages. (Exploration first framed this as "build the missing
  landing pages".)
- The tool-imagery hold is LIFTED (020 closed 07-23, owner lifted 07-24) — earlier context said it
  gated setup-builder work.
- The fabricated identity lives primarily in the `identity` SPEC; the live About page uses the real
  contact but fabricates stock claims. Both layers need the fix, spec first.
- S3 credentials block NOTHING for dartsonline (that blocker is 3 other sites' hand-made bytes).
- All 33 asset FILES already serve 200 — the gap is references, not deployment.
- `cmd/contrastscan` was deleted (`ee1944f89`); the rendered witness is `scripts/render_audit.py`.
