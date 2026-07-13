# RUNNING NOTES — PART 2 (idea.uk chassis + scheme/design work)

Part 1 archived at running_notes.md (~5690 lines, checkpoints (a)…(kkk)). This file continues the journal. Memory is OFF; this doc is the journal. **Present this file at the END OF EVERY TURN.**

═══════════════════════════════════════════════════════════════════════
CARRY-OVER STATE (still valid as of 2026-06-26)
═══════════════════════════════════════════════════════════════════════

## Standing preferences (STRICT)
Go not Python; plain human language, no LLM-hype, no flattery; confirm live API/schema/product facts before asserting/coding (`0 rows` not decisive — check the query/state first); reuse before rebuild; fix the framework structurally over one-offs; honest caveats + pushback incl. correcting my own reads; British English; low risk appetite; reasonable steps; ≤1 question per reply; don't create summary docs unless asked; minimal formatting (prose not bullets); banned words "perfect"/"critical"/"excellent"; no `logger.Debug` (use `logger.Info`); don't call a fix "final/last". Keep the runbook + this journal fresh. SQL run as FILES via `kubectl … < file` (pasting mangles `\set`/`\echo`/blank lines).

## Project facts
- **idea.uk** — LIVE Go service selling £29 "verified AI product idea" reports. Single Go binary under systemd (`idea`) on a Hetzner VM, nginx + LetsEncrypt, 127.0.0.1:8080. Orders in /var/lib/idea/orders.json (no DB). Live Stripe webhook → https://idea.uk/stripe/webhook (money path). Reserved tool paths: /request /confirm /approve /decline /stripe/webhook /internal/* /order/*. DNS (Cloudflare) → the VM, so chassis B2 deploys are INVISIBLE to the live site (safe staging-in-place). UNCHANGED/earning.
- **Chassis website-builder** — multi-agent (Go/Kafka/Postgres in k8s). domain → multipage site → static → Backblaze B2 (github → GH Actions → B2). Rebuilding idea.uk's front site; go-live is a deliberate VM cutover (NOT done).
- k8s namespaces: `-n ai-persona-system` (app pods) + `-n kafka`. Cluster `personae-kafka-cluster`. Bootstrap `personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092`.
- DB access: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`. Run migration FILES via `< file`.
- agent_definitions: `type`+`version` UNIQUE; `processing_mode` is nested in `default_config` (NOT a top-level column). snapshot_agent backs up before edits.

## Chassis design model (three layers)
1. **Direction** — domain-research-classifier writes identity/classification/content_direction/design_intent (design_intent carries structured palette.reference_values + typography.reference_values + style_direction → light/dark scheme).
2. **Composition** — site-design-planner resolves layout (scheme-aware weighted matcher) / typography / palette, then install_site_composition writes css_themes + style_collections + sites.style_collection_id + a `resolved_composition` POINTER spec (palette_id/name/source, NOT colour values). Refuses to overwrite an installed composition.
3. **Execution** — webdesign-agent renders styles.css from the installed composition (+ overlay) and git-commits it (→ GH Actions → B2). Sole styles.css deployer.

## CURRENT STATE — idea.uk scheme fix
- Scheme-aware weighted layout matcher is LIVE in production (merged into fork_theme_composition.go + resolve_composition_layout_action.go). Added layouts.scheme column + a new tool-portal-light layout via migration_layouts_scheme_and_light_tool_portal.sql (applied).
- idea.uk (site **1244516d-014d-421c-88c6-090bb1e9552a**, domain idea.uk) RE-RESOLVED in place onto **tool-portal-light (scheme light)** + **parchment palette** (palette-idea-uk: text/primary/border #1A1816, accent #A8391A, surface #E8DFCC, secondary #4A4540, background #EFE7D6, text_muted #837C72). New chain: style_collection 6e7d98fb → css_theme 4734d51c (theme-idea-uk) → layout 278ae068 (tool-portal-light). No needs_new_layout_candidate.
- styles.css RENDERED + DEPLOYED via webdesign-agent (commit 05ef817). VERIFIED correct: exactly tool-portal-light, parchment :root, NO LLM drift.
- **OPEN — the page-layer gap (Step 7):** the deployed PAGES still render dark. Page HTML is unchanged from the 2026-06-21 build and bakes per-component inline `<style>` (dark gradient header `.site-header--gradient` #1A1816→#4A4540; dark stock-photo hero with primary button `var(--accent-color, #0f3460)` → falls back to NAVY because styles.css defines `--color-accent` not `--accent-color`; dark CTA `.cta-section` var(--color-primary)=#1A1816; dark footer #1A1816). Only `.tool-list-section` + `.latest-news-section` read `--color-*` vars → only they go parchment. Inline hardcoded colours beat the external stylesheet, so a styles.css swap can't flip the pages. The PAGE layer needs rebuilding — but only if page-build component selection is scheme-aware AND the components defer to the palette; else a rebuild reproduces the same dark chrome. INVESTIGATING how pages are built (this turn).

## Key runbook + files (/mnt/user-data/outputs/, mirrored by user at docs/agent_docs/docs024_key_docs_latest/idea.uk/)
RUNBOOK_idea_uk_chassis_site_and_vm_deploy.md (re-resolve Steps 1–7, Step 7 = page rebuild/component investigation = current). reresolve_idea_uk_0{1,1b,2,2b,4}*.sql + 03_trigger.sh (planner) + 05_render.sh (webdesign-agent). migration_layouts_scheme_and_light_tool_portal.sql. RUNBOOK_idea_uk_vm_cutover.md. 002_system_architecture.md (rewritten). 016_debugging_guide_v2_55.md.

## Backlog (post page-fix)
(b) build-standard classifier migration (proven, NOT applied). (c) dead-slot hardening Go (design_reference fingerprint fallback). (d) improver-not-rewriter overlay for webdesign-agent analyze_design. (e) review built site (CTAs→/request, reserved paths) → VM cutover. (f) populate remaining layouts.scheme. Known chassis hazard: a content rebuild de-tools a tool page (page-content-writer regenerates from plan_sections which doesn't know the interactive tool) — fix pending.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT 2026-06-26 (lll) — established HOW pages are built (from 027/003/page-build-handler def); idea.uk dark-page root cause is structural (component layer, not styles.css)
═══════════════════════════════════════════════════════════════════════
Read: page-build-handler agent_definition (inline), 027_design_and_site_planner_v2.md (= "026 Design Composition & Site Design Planner"), 003_contracts_and_standards_7_.md. 

HOW PAGES ARE BUILT:
- Components live in content_components (function = kebab-case identifier + html_template carrying its OWN scoped CSS). Variants = distinct functions (hero, hero-split, header-professional-dark, header-minimal-light, footer-4-column).
- SITE-LEVEL header/footer/head: chain = layouts.default_{header,footer}_component_id -> install_site_composition copies onto style_collections.{header,footer}_component_id -> update_site_defaults copies onto site_components.component_id -> renderAndStoreSiteComponent renders template -> site_components.rendered_html. If style_collections.header_component_id NULL or site_components unlinked -> RenderFallbackHeader (hardcoded generic). (003 Site Component Linkage Contract.)
- PER-PAGE sections: page-build-handler.plan_sections picks section functions from a GLOBAL component library (027 §10 + §312: layout-aware section selection is FUTURE work, "025 Phase 4"). page-content-writer writes prose per section -> save_page_sections persists to page_components (slot_name/component_id/rendered_html) -> page-rerender concatenates page_components.rendered_html in order, wraps with site header/footer, commits to git -> B2.
- COLOUR MODEL (003 CSS Colour Inheritance): styles.css body{color:var(--color-text)}; h1-6 use var(--section-heading,var(--color-primary)); p/li use var(--section-text,inherit). A DARK section sets --section-* on its OWN container (SANCTIONED override). Light sections inherit the palette automatically.

WHY idea.uk PAGES RENDER DARK despite tool-portal-light + correct styles.css:
The re-resolve changed composition + re-rendered styles.css, but did NOT touch page HTML or site header/footer wiring, and page assembly is NOT scheme-aware. Specifically:
1. Header/footer: pages still carry the original 2026-06-21 gradient DARK header + dark footer. The re-resolve installed a new style_collection but (a) tool-portal-light almost certainly declares NO default_{header,footer}_component_id (I only set css_template/structure_tokens/tags/scheme when I added it) -> new collection's ids likely NULL, and (b) site-design-planner does NOT run update_site_defaults -> site_components never re-pointed. So a re-render would keep the OLD header or drop to RenderFallbackHeader — NOT tool-portal-light's light header.
2. Hero + CTA: global `hero` (dark image overlay) + `call-to-action` (dark) components. plan_sections isn't layout-aware -> a re-render picks the SAME dark sections, even though tool-portal-light's own hero is light/typographic.
3. Hero component BUG: button background var(--accent-color, #0f3460) — system var is --color-accent (NOT --accent-color) -> falls back to NAVY. The stray blue is baked into the `hero` content_components template, NOT styles.css. (003 New-component checklist #3: use render-context vars not hardcoded; the var name is also just wrong.)

CONCLUSION: a plain page re-render will NOT make idea.uk light — it reproduces the dark chrome + navy button. Fix is STRUCTURAL at the component layer (matches "fix framework not patch one page"): (a) author + assign light header/footer components for tool-portal-light (+ ensure update_site_defaults runs to push them into site_components) and a light/typographic hero; OR (b) make section/component selection scheme-aware. The --accent-color->--color-accent hero fix lands regardless.

NEED FROM USER (queries + code) to pin the exact fix:
1. layouts: SELECT name, default_header_component_id, default_footer_component_id FROM layouts WHERE name IN ('tool-portal-light','tool-portal-dark','brochure-formal');
2. idea.uk wiring: SELECT header_component_id, footer_component_id FROM style_collections WHERE id='6e7d98fb-843c-4931-8626-1af5ffb237c0'; + SELECT slot_name, component_id FROM site_components WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a';
3. component inventory + buggy templates: SELECT function, is_active FROM content_components WHERE function LIKE 'header-%' OR function LIKE 'footer-%' ORDER BY function; + SELECT function, html_template FROM content_components WHERE function IN ('hero','call-to-action');
4. code: plan_sections (section selection) + update_site_defaults (header/footer -> site_components) + page-rerender agent definition. Have 026_component_regeneration_flow_2_.md uploaded — will read for plan_sections; action code/def would confirm.

NOTE other uploads now available for this work: install_site_composition_action.go, render_css_from_spec_action.go + render_css_composition_{loader,helpers}.go, emit_design_items_action.go, fork_theme_from_site_action.go, 026_component_regeneration_flow_2_.md, 030_phase1_plan_and_reconciler_5_.md, 021_site_spec_and_classifier.md, 003/002/001 latest. The 081*_trigger_rerender_*.sh scripts (gaswholesalers/robot-hands) are existing page-rerender trigger examples to reuse for idea.uk.
