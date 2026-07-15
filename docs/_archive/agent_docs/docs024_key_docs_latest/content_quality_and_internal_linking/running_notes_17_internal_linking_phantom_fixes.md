# Running notes 17 — internal linking phantom fixes (2026-06-10)

Site: gamesdesign.co.uk. Work: eliminate phantom/broken internal links at source, add a detection net, leave intent-aware resolution for a dedicated agent. Companion: `FOCUS_internal_linking.md`, `FOCUS_content_quality.md`. Deploy runbook: `RUNBOOK_linking_phantom_fixes.md`. Remaining: `PLAN_b4_b5_hubs_and_link_resolver.md`.

## Policy settled
- Phantom/missing internal link = **loud but non-blocking**. Not a deploy stopper; the improvement loop resolves it. Deploy gate flags phantoms as warnings, not errors.
- Prevention is layered: stop producing them at each write path (Layer 1), detect at the deploy gate + post-deploy audit (Layer 2/3), restore intent-correct destinations via a dedicated agent (Step 3).

## Decisive findings (verified against code/data, not assumed)
- `sourceResolver.resolve` (`plan_sections_action.go`) `case "pages"` **fabricated** `/<path>.html` and returned `found=true` for any non-existent page → the phantom generator behind hero/CTA. `on_missing` never fired; schema fallbacks were dead code.
- `hero` + `call-to-action` are the **only** components using a `pages.*` source (`pages.contact`, `pages.services`). Tight blast radius.
- Header/footer phantoms (`/contact.html`, `/privacy.html`, `/terms.html`) were **hardcoded `ContentData`** in `render_site_components_action.go` (`cta_url`, `legal_links`) — NOT the `multipage_actions` 310–318 fallback nav (nav is already real-page-derived via `GetNavItems`), and NOT in the templates. So `nav-link-fixer` (which only rewrites `#slug` anchors in templates) can't reach them.
- `link_registry` has a `target_page_id` column + FK but `syncLinksToDB` never populates it; `extract_and_sync_links` is wired into no live workflow → `link_registry` empty. Audit reads `rendered_html`.
- `validate_page_content` already had `validateInternalLinks`, but emitted missing targets as one non-blocking warning (phantom + planned lumped) and never inspected `site_components`. Its `normalizePagePath` (lowercase + append `.html`) disagreed with the audit's normalisation.
- `component-template-fixer` exists but **punts on CTAs** (`cta_improvement`/`cta` → `needs_review`). `identity-advisor` and `sites.approval_mode` do **not** exist.
- B4/B5: the `*_index_url` specs are **absent** (the `identity` spec has tone/contact/services but no `*_index_url` keys); `game-list` even has a real fallback `/games/index.html` that still didn't apply (spec-path resolver / template gating to verify). Real hubs exist: `tools-index`, `guides-index`, `games-index`.
- Operational: `improvement-sweep` scheduled_task is **disabled** (`enabled=f`, last completed 2026-05-08), intentionally paused during core build.

## Shipped/written this session (files in outputs)
- `plan_sections_action.go` — `resolve` `pages` case: real URL or `(nil,false)`, no fabrication. (Patch.)
- `step1_hero_cta_phantom_fix.sql` — `hero`/`call-to-action`: `on_missing: skip_field`, fallbacks removed, templates gated on url.
- `datahelpers/links.go` — canonical `ExtractHrefs`/`ClassifyLinkScope`/`IsAssetPath`/`NormalizePagePath`/`PageURLSet`.
- `validate_page_content.go` — gate now uses datahelpers; `phantom_link` + `empty_internal_href`, non-blocking.
- `check_phantom_internal_links.go` — post-deploy audit on datahelpers; routes by surface (site_component→`nav-link-fixer`, page_component→`internal-link-resolver`). Inert until enabled.
- `render_site_components_action.go` + `layer1b_header_footer_phantom_fix.sql` — header/footer phantoms fixed at source; `legal_links` from `GetNavItems(NavGroupLegal)`, `cta_url` from real contact page, header CTA gated, footer legal data-driven.

## Next
1. B4/B5 — Browse-All hub links via a `query.section_index_for:<type>` verb (see PLAN). Smaller; finishes write-path phantoms.
2. Step 3 — `internal-link-resolver` agent: intent-aware destinations across ANY component + build-time `unresolved_cta` signal (see PLAN).
3. Before re-enabling `improvement-sweep`: enable `phantom_internal_links` and have both handler agents in place.
