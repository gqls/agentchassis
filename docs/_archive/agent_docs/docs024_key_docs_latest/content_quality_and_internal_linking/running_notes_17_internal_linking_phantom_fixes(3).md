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

## B4/B5 — done (ready for batch deploy)
- `section_index_for.go` — new `queryresolve` verb (new file in the package; only existing-code edit is one switch case). Shared-area lookup then URL-prefix fallback; returns the hub URL or `nil` (never `""`).
- `b4_b5_hub_links_schema.sql` — repoints `tool-list`/`game-list_pre_037`/`guide-list_pre_037` `cta_url` to `query.section_index_for:<type>`; drops `game-list`'s dead `/games/index.html` fallback. (Confirmed: that fallback never fired — the field had no `on_missing`, so it defaulted to `skip_field`, which ignores `fallback`; and the list templates render the Browse-All anchor ungated, hence `href=""`.)
- `b4_b5_hub_links_template_gate.sql` — gates the three Browse-All anchors on `{{if .cta_url}}` (correct-or-absent for hub-less sites).

## Step 3 — resolver core done; agent + wiring next
- Decision (confirmed): build-time resolution via `internal-link-resolver` spawned as a sub-agent of `page-content-writer` (like `research-agent`). No persistence; a post-deploy phantom → page rebuild re-runs resolution. `merge_with` is single-source, so the resolver augments each hero/`call-to-action` section's `resolved_data` in Go; the render loop is unchanged.
- `resolve_internal_links_action.go` — section-augmenting, component-aware (`hero`: `cta_url`/`secondary_cta_url`; `call-to-action`: `primary_cta_url`/`secondary_cta_url`), validates every target via `datahelpers.PageURLSet`, returns augmented `sections_ready` + `unresolved`. v1 rule: top content hubs by `nav_order`, excluding about/contact/legal and the page's own hub.

### Guideline audit (001/003) — fixes applied to the resolver action
- Was reading the `sections` array with the literal key via `ExtractNestedField` → would silently run on empty. Fixed: resolve the config PATH (`params.StepConfig.Config["sections"]`), then `ExtractNestedField`.
- `sections` was in `ActionInputSpec` → `current_page.sections` collision risk. Fixed: removed from the spec; read from the config path. Spec keeps scalars only (`site_id`, `page_type`, `page_name`).
- Own-hub exclusion was keyed on the section name (`"hero"`), never matched. Fixed: added `page_name` and excluded `hub.Name == page_name`.
- Logged the pattern in `016_debugging_guide_v2_45` (§9 + §0 #15).

### Agent-modeling facts (from `research-agent` row + 003)
- `agent_definitions` has NO `processing_mode` column; the workflow lives in `default_config.workflow`, with `processing_mode`/`timeout_seconds` inside `default_config`. `task_workflow`/`orchestrator_workflow` are null for called sub-agents.
- 003 requires `agent_category` (use `specialist`), `input_contract`, `output_contract`, `image_repository`/`image_tag`. Topics templated (`system.agent.{type}.process`, etc.).
- Rebuild trigger (003 `content_direction` + arch table): set `pages.build_status = 'needs_rebuild'`; the `page-rebuild` specialist picks it up. So the check's `page_component` finding should set `needs_rebuild`, not route to the resolver directly.

## Pre-deploy review (2026-06-11)
- Found the Step 1 `plan_sections_action.go` fix was MISSING from outputs while the current uploaded file still had the fabrication (line 334-335). Recreated as a full drop-in copy of the current upload + the one-block fix (returns `(nil,false)` + Info log instead of fabricating `/<path>.html`). Verified: one-block diff, balanced, `zap`/`r.logger`/`r.siteID` already present in the file.
- Verified `section_index_for.go` against the live `queryresolve` package: signature/parse/`req.SiteID uuid.UUID`/`*sql.DB` match the siblings, and the `status IN ('active','deployed')` filter is exactly `resolvePagesWhereType`'s proven filter.
- Verified the resolver action's seams: `datahelpers.ExtractNestedField` exists and is exported; `stringOrEmpty` is package `actions` (defined in `plan_sections_action.go`); no name collisions for the new helpers; the validation `PageURLSet` filter is identical to the deploy gate's (`NOT IN ('deleted','archived')`), and hub candidates (`IN ('active','deployed')`) are a subset — the resolver cannot emit a URL the gate flags.
- Verified my `render_site_components_action.go` / `validate_page_content.go` outputs are the current uploads + localized intended edits only.
- Noted intermediate state: layer1b SQL is applied but its Go isn't deployed, so header/footer phantoms PERSIST until the batch image rolls and site components re-render (old Go still injects `cta_url=/contact.html` and the hardcoded legal links into the now-gated/data-driven templates — values present, so gates pass). Expected; clears at RUNBOOK step 3.
- `check_phantom_internal_links.go` still routes `page_component` → `internal-link-resolver` (line 138); the agreed change to a rebuild route is pending `page-rebuild`'s input contract. Inert until the check is enabled, so not a batch blocker.


1. Write the `internal-link-resolver` `agent_definitions` row (model on `research-agent`: `default_config.workflow`, `specialist`, contracts, image fields, templated topics; thin workflow: `ensure_site_record → resolve_internal_links → complete`).
2. Wire `page-content-writer` (needs the CURRENT definition uploaded to edit the live JSON): `spawn_agent` (role `link-resolver`) near `spawn_research_agent`; `call_agent` after `build_render_context` mapping `site_id`/`page_type`/`page_name` + `sections: input_data.section_plan.sections_ready`; repoint the loop's `iterate_over` to the resolver's returned `sections_ready`.
3. `routeBySurface`: `page_component` → set `needs_rebuild` (page-rebuild) instead of `internal-link-resolver`. Confirm the exact work-item `item_type` `page-rebuild` consumes first.
4. `unresolved_cta` emission from the resolver's `unresolved` (non-blocking signal).
5. Before re-enabling `improvement-sweep`: enable `phantom_internal_links` with the handler path in place.
