# TODO — chassis scheme work + idea.uk + carried backlog

Consolidated from this chat. Companion docs: `REPORT_scheme_does_not_reach_components.md` (the P0 investigation plan), `RUNBOOK_idea_uk_chassis_site_and_vm_deploy.md` (idea.uk steps 1–8), `running_notes_2.md` (the journal). idea.uk site_id `1244516d-014d-421c-88c6-090bb1e9552a`.

---

## Done this chat (context)
- [x] Scheme-aware **weighted layout matcher** merged into `fork_theme_composition.go` + `resolve_composition_layout_action.go` and live in production.
- [x] Migration applied: `layouts.scheme` column + new `tool-portal-light` layout (`migration_layouts_scheme_and_light_tool_portal.sql`).
- [x] idea.uk **re-resolved in place** onto `tool-portal-light` + parchment palette (steps 1–4). Composition correct; no `needs_new_layout_candidate`.
- [x] **styles.css rendered + deployed** via `webdesign-agent` (commit `05ef817`) and **verified** — exactly `tool-portal-light`, parchment, no LLM drift.
- [x] **Established exactly how pages are built** (page-build-handler → plan_sections → content-writer → save_page_sections → page-rerender; header/footer via site_components; component_selector scoring) and **root-caused** why idea.uk renders dark.
- [x] **Wrote the structural-gap report** for a dedicated fix thread.

---

## P0 — FUNDAMENTAL: a site's scheme does not reach its components  (dedicated thread)
The big one. idea.uk renders dark despite a correct light composition/stylesheet because the component library is dark-oriented and the scheme signal stops at `styles.css :root`. Full plan in `REPORT_scheme_does_not_reach_components.md`. **Do not fix in passing.**
- [ ] Open a dedicated thread; cold-start from the report §7 checklist + `running_notes_2.md`.
- [ ] Run the **9 investigations** (report §6): (A) exact render path for a section + a site component; (B) trace the scheme signal end-to-end; (C) inventory the whole component library against scheme (hardcoded vs variable-driven); (D) audit the css_template ↔ component class-name contract; (E) the section-contrast model; (F) header/footer scheme + `update_site_defaults` wiring; (G) the existing `--section-*` luminance mechanism; (H) migration + backfill safety; (I) a scheme-coherence audit guard.
- [ ] Answer the **8 design questions** (report §5) before any code — especially Q4 (should components adopt the layout's class vocabulary so `styles.css` styles them?).
- [ ] Validate or revise the **provisional fix shape** (report §8) against the findings.
- [ ] Hold to the steer: scheme = a **variable-value override** consumed by one component (palette + `--section-*`); **new functions only** where a component is genuinely too structurally complex to share. No `*-light`/`*-dark` duplication.

---

## P1 — idea.uk  (minimum now; finish after the P0 fix)
- [x] **Minimum now: no change.** Active `hero` already uses `--color-accent`, so a later rebuild gets the rust button for free; dark chrome stays until P0.
- [ ] **After P0:** rebuild idea.uk pages (reuse `flag_page_image_rebuild` → `needs_page` → `page-build-handler`) so they pick up the light components + fixed hero.
- [ ] Re-verify the deployed pages read light/parchment with no stray blue.
- [ ] **Review the built site before cutover:** CTAs resolve to `/request` (not a phantom `/contact`); no collision with reserved tool paths (`/request /confirm /approve /decline /stripe/webhook /internal/* /order/*`); check the differentiators/pricing/contact-form gaps.
- [ ] **VM cutover** (`RUNBOOK_idea_uk_vm_cutover.md`) — gated on P0 + the site review. nginx front door: static for general pages, reverse-proxy reserved paths to `127.0.0.1:8080`; **prove the Stripe webhook through nginx before cutover**; DNS unchanged; rollback = restore one nginx block. The live £29 tool stays untouched throughout.

---

## P2 — carried chassis backlog  (independent of P0; can proceed any time)
- [ ] **Apply + test the build-standard classifier migration** (`migration_classifier_build_standard.sql` — proven by simulation, NOT applied). Adds a quality+fit block to the classifier's `classify_and_extract` prompt. Test on a fresh build first; confirm an adopted rebuild (e.g. gamesdesign) stays faithful.
- [ ] **Deploy dead-slot hardening** (3 Go files: `resolve_composition_reference_helpers.go` design_reference-fingerprint fallback after design_intent, + the two `extractPaletteSignal`/`extractTypographySignal` swaps). Chassis image rebuild + roll `site-design-planner`.
- [ ] **improver-not-rewriter overlay** for `webdesign-agent.analyze_design` (show the established palette + a diff + audit; check for an existing audit/log table first). Note: targets `webdesign-agent`, not the separate `site-scraper`.
- [ ] **Populate the remaining `layouts.scheme`** (~15 rows) from each layout's real background (query is in the layouts migration).

---

## P3 — known hazards / smaller
- [ ] **Content rebuild de-tools a tool page** (confirmed hazard): a `needs_page`/`link_resolution_rebuild` for a tool/game page routes to `page-build-handler` → page-content-writer, which regenerates from `plan_sections` (which doesn't know the interactive tool that lives as a section's `rendered_html`), silently dropping the tool. Fix pending: route link maintenance through a preserve-sections re-render; stamp `source_item_id`; add an interactivity-aware save guard. **Relevant to P1** — if idea.uk gains interactive tools, its page rebuild could hit this. See 005 / 016b / 020 / 026.
- [ ] Runbook backlog (idea.uk): differentiators empty cards; unresolved CTAs; dead contact form; thin nav/footer + empty meta; a full fresh rebuild as an end-to-end test. **Parked:** rewrite the £29 report language/format.

---

## Quick status line
P0 (scheme reaches components) is the blocker for a genuinely-light idea.uk and for the VM cutover. Everything in P2 is independent and can move in parallel. idea.uk is safe as-is (staging only; live tool on the VM untouched).
