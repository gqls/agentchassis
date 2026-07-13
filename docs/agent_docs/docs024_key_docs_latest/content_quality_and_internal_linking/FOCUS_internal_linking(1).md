# FOCUS — Internal Linking (agent-chassis / gamesdesign.co.uk)

**Status:** current as of 2026-06-10. Grounded in `plan_sections_action.go`, `validate_page_content.go`, `render_site_components_action.go`, `site_db_actions.go`, `fix_nav_link_templates_action.go`, `datahelpers/links.go`, and the `phantom_internal_links` discovery check. Companion to `FOCUS_content_quality.md` — the hero-CTA defect sits on the seam between the two.

---

## What "internal linking" covers
Every link from one built page to another: header/footer navigation, in-body anchor/cross links, list-hub cards, "Browse All X" buttons, and the destination half of hero/CTA buttons. A CTA's text is content; its destination is a link.

## Through-line (the structural rule)
Internal link targets resolve from the **real `pages` set**, never fabricated and never brochure-assumed. A target that resolves to no `pages` row is a phantom. Where a destination can't be resolved, render nothing (correct-or-absent) rather than a broken/empty link — and, at build time, emit a signal so the absence isn't silent.

## Policy (settled this round)
A missing/phantom internal link is **loud but non-blocking** — it does not stop a deploy; the improvement loop resolves it. So the deploy gate flags phantoms as warnings, not errors. (A missing link is not a page-stopper.)

---

## Decisive findings (the old open questions, now answered)
1. **Hero/CTA CTAs are resolvable fields, not template hardcoding.** `hero` and `call-to-action` are the **only** active components using a `pages.*` source. Both used `cta_url ← pages.contact` (fallback `/contact.html`) and `secondary_cta_url`/`primary_cta_url` ← `pages.services` (fallback `/services.html`); the button *text* is `source: llm`. So text and destination were decoupled by schema (LLM writes "Browse Tools", url hardwired to contact). **The deeper cause:** `sourceResolver.resolve` (`plan_sections_action.go`) `case "pages"` *fabricated* `"/" + path + ".html"` and returned `found=true` for any non-existent page — so `on_missing` never fired and the schema fallback was dead code. A `pages.*` source could never be "missing"; it always invented a phantom.
2. **`link_registry` only records, never validates.** It *has* a `target_page_id` column + FK to `pages`, but `syncLinksToDB` never populates it. And `extract_and_sync_links` is wired into **no live workflow**, so `link_registry` is empty. It is not a usable substrate today — the phantom check reads `rendered_html` directly.
3. **Nav is already real-page-derived; the brochure fallback was not the live path.** Header/footer nav is built by `GetNavItems`/`loadNavItems`/`loadFooterNavItems` from real `pages`. The header/footer phantoms (`/contact.html`, `/privacy.html`, `/terms.html`) came from **hardcoded `ContentData` in `render_site_components_action.go`** (`cta_url`, `legal_links`) — not from the `multipage_actions.go` 310–318 fallback nav, and not from the templates. (Correction to the earlier note that blamed 310–318.)
4. **`nav-link-fixer` cannot fix the header/footer phantoms.** It only find/replaces `#{{.slug}}`/`#{{.name}}` anchors inside `html_template` (`fix_nav_link_templates_action.go`). The B2/B3 literals were `ContentData` values (and, for the footer, literal `<a>` tags) — out of its reach. Hence B2/B3 fixed at source.
5. **Auditor/specialist reality:** `component-template-fixer` exists but **explicitly punts on CTAs** (`cta_improvement`/`cta` → `fixed:false, needs_review`). `identity-advisor` does **not** exist. `sites.approval_mode` does **not** exist. So those PLAN pieces are not available to rely on.

## Shared machinery (single source of truth)
- **`datahelpers/links.go`** (new) — canonical `ExtractHrefs`, `ClassifyLinkScope` (empty/internal-anchor/external/mailto/asset/page), `IsAssetPath`, `NormalizePagePath` (drop `#?`, lowercase, strip trailing `index.html` + slashes, root→`/`), `PageURLSet`. Replaces three previously-divergent normalisers (validator lowercased + appended `.html`; the audit stripped `index.html`/slashes; the inventory had no asset handling). Used by the gate and the audit so they agree by one implementation.
- **`validate_page_content.go` → `validateInternalLinks`** (deploy gate) — now extracts/classifies via `datahelpers`; flags `phantom_link` (no `pages` row at all) and `empty_internal_href`, both **non-blocking warnings**. Planned-but-unbuilt pages (a row exists, status not `deleted`/`archived`) are tolerated. `loadValidPagePaths` now returns `datahelpers.PageURLSet`.
- **`check_phantom_internal_links.go`** (post-deploy audit) — scans `page_components` + `site_components` `rendered_html` via the same `datahelpers` helpers; routes by surface to **distinct handlers**: `site_component` → `nav-link-fixer`, `page_component` → `internal-link-resolver`. Inert until `phantom_internal_links` is added to a discovery agent's `checks` array.
- **`GetNavItems(NavGroupLegal)`** — real legal pages, reused for footer legal links.

---

## Shipped this round
- **`resolve` fabrication fix** (`plan_sections_action.go`): `case "pages"` returns the real URL when the page exists, else `(nil, false)` — no more `/<path>.html` invention. Blast radius = `hero` + `call-to-action` only.
- **Step 1 / Layer 1a** (`hero` + `call-to-action` schema/template, SQL): CTA-url fields → `on_missing: skip_field`, phantom `fallback`s removed; templates gate each button on text AND url (dropping the `/contact.html` / `#features` literals). Result: an unresolved CTA renders no button rather than a phantom.
- **Layer 2** (`datahelpers/links.go`, `validate_page_content.go`, `check_phantom_internal_links.go`): consolidation + gate hardening + the audit check.
- **Layer 1b / B2-B3** (`render_site_components_action.go` + SQL): `legal_links` from `GetNavItems(NavGroupLegal)`; `cta_url` from the real contact page (reusing the `companyLinks` contact extraction); `header-bold-gradient` CTA gated on `cta_url`; `footer-4-column` legal links made data-driven (`{{range .legal_links}}`). After applying: force-rerender `site_components` (header/footer), then re-run the audit dry-run.

## Remaining
- **B4/B5 — empty-href hubs.** The three list slots (`tool-list`, `guide-list`, `game-list`) render "Browse All X" with `href=""` because the `*_index_url` site_specs are unpopulated and inconsistently sourced (`identity.tool_index_url`/`game_index_url` vs `navigation.*`/`blog.*`). Fix: resolve from the real hub pages (`tools-index` `/tools/index.html`, `guides-index`, `games-index`) from one consistent source, or omit. Same "resolve from real pages" principle. **Next.**
- **Step 3 — `internal-link-resolver` agent.** Its own orchestrator/workflow; responsibility = resolve/validate **any** intended internal link (not just hero/nav — guides link to tools, in-body prose links, related blocks) against the real pages, choosing intent-appropriate destinations (e.g. a games/tools site's header CTA → primary hub, a guide → its related tool) rather than a fixed contact page. Reuses `prepare_link_context`'s `available_pages` and the fixed `pages` resolver. Also owns the **build-time `unresolved_cta` signal** — flagging a section that has CTA text but no resolved url, so a correctly-dropped button is detectable (the deploy gate can't see an absence). This is the durable home for the intent-aware resolution that Step 1 / Layer 1b deliberately left as correct-or-absent.

## Operational note
`improvement-sweep` scheduled_task is **disabled** (`enabled = f`, last completed 2026-05-08), intentionally paused during core build. The detect→loop-fixes-shortly model depends on it. Before re-enabling: have the `phantom_internal_links` check enabled AND both handler agents in place (`nav-link-fixer` exists; `internal-link-resolver` is Step 3), so resuming the sweep clears findings rather than accumulating them.

## Files
- `plan_sections_action.go` — `sourceResolver.resolve` (`pages` case), field-source resolution loop.
- `datahelpers/links.go` — shared extractor/classifier/normaliser + `PageURLSet`.
- `validate_page_content.go` — `validateInternalLinks`, `loadValidPagePaths`.
- `check_phantom_internal_links.go` — post-deploy audit.
- `render_site_components_action.go` — header/footer `ContentData` (`cta_url`, `legal_links`); nav from `GetNavItems`.
- `fix_nav_link_templates_action.go` — `nav-link-fixer`'s action (`#slug`→`url` template fixes only).
- `site_db_actions.go` — `upsertPage`, `ExtractAndSyncLinksAction`/`link_registry` (unwired), `GetNavigationStructure`.
- content_components: `hero`, `call-to-action`, `header-bold-gradient`, `footer-4-column`.
