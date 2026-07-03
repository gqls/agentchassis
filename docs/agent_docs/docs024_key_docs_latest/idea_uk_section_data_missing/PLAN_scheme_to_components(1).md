# PLAN — make a site's scheme reach its components

Design plan for the dedicated P0 thread. Companions: `REPORT_scheme_does_not_reach_components.md` (problem statement + investigations A–I + questions Q1–Q8), `HANDOFF_scheme_to_components.md` (cold-start brief), `RUNBOOK_scheme_to_components.md` (commands), `running_notes_scheme_to_components.md` (the journal). This file is the forward map: what is decided, what is open, and the order of work. No code or DB changes have been made in this thread — it is still the design phase.

---

## The problem in one paragraph

The chassis decides a site's scheme (light/dark) at composition — `design_intent.style_direction` drives the layout's `scheme` and the palette — and renders `styles.css` from it, but that scheme reaches only the stylesheet's colour variables, never the components that render each section, header and footer. The components come from a dark-oriented library by a one-active-component-per-function lookup and self-style with inline CSS that hardcodes a dark treatment, so a light-resolved site renders dark chrome over light content. The fix is to make the scheme reach the components as a **variable-value override consumed by one component** (palette + the existing `--section-*` mechanism), not by duplicating components into `*-light`/`*-dark`, splitting into new functions only where one is genuinely too structurally different to share. It must be designed before any code, then migrated across the library and back-filled to built sites without regressing dark sites.

## The thesis (the steer)

Scheme is a set of variable values supplied by the composition/renderer and consumed by de-hardcoded components — not a second copy of each component. A component declares its structure once and reads `--color-*` (palette) and `--section-*` (section text/surface/border context) for all colour and contrast; it hardcodes nothing and sets no `--section-*` itself. The override is the values those variables take: the site's base scheme sets the palette and base `--section-*`; a per-section contrast intent sets a contrasting `--section-*` on one section. New functions only for genuine structural divergence.

---

## Confirmed at code level this thread (the basis for the design)

These are established from the render path and schema in bundle 1, not assumed.

- **Scheme is computed then dropped at render.** `deriveSchemeFromDesignIntent(style_direction, suggested_style)` returns `light`/`dark`/`""`; `resolveLayoutByTags` uses it as a near-hard constraint to choose the layout. `buildResolvedCompositionSpec` records the layout/palette/typography ids, lineage and reasoning, but **not the scheme value** — downstream it survives only on `layouts.scheme` (constrained to `light`/`dark`/`neutral`).
- **`styles.css` is rendered in three appended parts** by `RenderCSSFromSpecAction`: (1) the layout `css_template` rendered with `{{palette}}`/`{{typo}}`/`{{token}}` FuncMap helpers over merged palette/typography/structure maps; (2) component CSS from the **`css_snippets`** table, where `applies_to` overlaps the site's component list; (3) a luminance-based `--section-*` defaults block from `buildSectionDefaults`, computed from the merged background/surface colours.
- **The renderer already has a section-context mechanism, in two halves.** `SectionStyles` is a per-component list `{Function, ClassName, IsDark}` passed into the layout template, with `ClassName = function + "-section"` and `IsDark` from `content_components.is_dark_section`. `buildSectionDefaults` appends global `--section-*` defaults chosen for readability against the palette's luminance. This is the existing `--section-*` vehicle the 003 contract describes.
- **The class-name contract is `{function}-section`, honoured on one side only.** The stylesheet side assumes components emit `hero-section`, `call-to-action-section`, etc.; the report's evidence is that the components actually emit `.hero`, `.cta-section`, `.tl-card`. So the layout's per-section rules and the section-context variables do not reach the components by class, and the components' own inline CSS wins.
- **Neither render path reads `layouts.scheme`.** The CSS loader's SELECT (`render_css_composition_loader.go`) pulls the layout's name, `css_template` and `structure_tokens` but not `scheme`; the component `RenderContext` (`component_library.go`) has palette-colour fields but no scheme field, and `contextToInterfaceMap` can only expose what the struct holds. So a component template gets colour values but no light/dark signal — which is why dark components hardcode their darkness.
- **The site→scheme recovery path exists.** `sites.style_collection_id → style_collections.css_theme_id → css_themes.layout_id → layouts.scheme` (the same join the CSS loader already uses to find the theme, extended by one column). The `resolved_composition` spec also carries `layout_id` directly.
- **There is a third CSS surface beyond the inline `<style>`.** `css_snippets` (`name`, `css_content`, `applies_to`) is appended to `styles.css` independently of `content_components.html_template`. Where the dark hero/CTA/footer treatment actually lives — inline `<style>`, `css_snippets`, or both — is not yet established and must be inventoried (it changes where de-hardcoding happens).
- **`is_dark_section` is the only per-section contrast signal today**, and it is weak: not used in section selection, unreliable (idea.uk's header is flagged false yet renders dark), and conflates "intrinsically dark" with "should contrast the page".

---

## Design questions — current status

**Q1 — Where does scheme live at render time, and how does it reach a component?**
Leaning: carry the site's base scheme as an explicit value into both render entry points — add `scheme` to the CSS loader's SELECT + a `Scheme` field on `themeComposition`, and a `Scheme` field on `RenderContext` (exposed via `contextToInterfaceMap`), both populated from `layouts.scheme` via the recovery path. The per-section dark/light treatment stays in the renderer's existing `SectionStyles` + `buildSectionDefaults` path. Mechanism choice (render-context value vs body class vs `--section-*` only) to finalise against an actual `css_template` (D). Status: largely answered, pending D.

**Q2 — Is a section's darkness a site, component, or per-placement property?**
Model: **site base scheme + per-section contrast intent**. The renderer already half-implements this — base scheme via the background luminance flags, per-section dark via `is_dark_section`. Open: where the per-section contrast intent is decided and stored (site plan? composition? auditor?) and how it reaches `--section-*`. Status: model agreed; storage/source open (E).

**Q3 — What is the override mechanism, exactly?**
Leaning: the renderer is the single point that sets each section's `--section-*` and background from (base scheme + contrast intent + palette luminance); components only read. Reuse `SectionStyles` + `buildSectionDefaults` rather than inventing a parallel mechanism. Status: leaning, pending D.

**Q4 — Reconcile the component class vocabulary with the layout class vocabulary (GATING).**
The concrete mismatch is now pinned: `buildCSSsectionStyles` emits `{function}-section`; components emit their own classes. Leaning **(a) components converge on the `{function}-section` contract**, because `SectionStyles` + `buildSectionDefaults` are already built for exactly that contract, so adopting it lets the stylesheet carry section structure + adaptive context and shrinks each component's inline `<style>` to true structure-only — the maximum reuse and the report's stated preference. Option (b) (keep own classes, become strictly `--color-*`/`--section-*`-driven) is a smaller per-component change but leaves two vocabularies and a partly-redundant `SectionStyles`. Decision gated on seeing an actual layout `css_template` (does its `SectionStyles` block emit structure+context for `{function}-section`, or only context?) — that is the investigation-D pull. Status: leaning (a), **decision pending D**.

**Q5 — What becomes of `is_dark_section`?**
Leaning: keep it as the per-section contrast signal but fix its reliability and stop it conflating two ideas; do not use a component's intrinsic darkness as the site's scheme. Possibly rename/replace with an explicit per-placement contrast intent once Q2's storage is settled. Status: open, tied to E.

**Q6 — Header/footer scheme + wiring.**
One adaptive header and one adaptive footer driven by scheme, wired via `layouts.default_header_component_id` / `default_footer_component_id` + running `update_site_defaults` in the composition/build path (today no layout declares defaults and the planner does not run it, so site_components keep the original dark header/footer). Status: direction clear; mechanics pending F.

**Q7 — Migration + backfill without breaking dark sites.**
Change shared component templates, keep existing dark sites dark, re-render every affected site. Reuse the `flag_page_image_rebuild` → `needs_page` → `page-build-handler` trigger for page rebuilds. Sequence (template change → which sites → page rebuild vs re-render) to be specified. Status: open (H).

**Q8 — A guard so it can't silently regress.**
A scheme-coherence check in the design auditor / improvement loop that flags "section scheme does not match site scheme / unintended contrast". Status: open (I).

---

## Provisional fix shape (to validate, not build yet) — annotated

1. Introduce an explicit **scheme signal** into both render entry points from the resolved composition. *Confirmed needed: neither path reads `layouts.scheme`.*
2. **De-hardcode** the dark components (`hero`, `call-to-action`, header, footer, plus any found in C): read `--color-*`, stop setting `--section-*`/dark backgrounds inline. One function each. *Pending C: also check `css_snippets`, not just inline `<style>`.*
3. Make the **renderer the single adaptation point** via `SectionStyles` + `buildSectionDefaults`. *Mechanism exists; reuse it.*
4. Resolve the **class-vocabulary** question (Q4) — leaning components adopt `{function}-section`. *Biggest structural decision; gated on D.*
5. **Header/footer**: one adaptive component each, wired via `layouts.default_*_component_id` + `update_site_defaults`. *Pending F.*
6. Keep **direct-function resolution**; treat `component_selector` scheme-awareness as a minor follow-up only. *Confirmed: selector is not on the current path.*
7. **Migrate + backfill** carefully. *Pending H.*
8. Add the **scheme-coherence audit**. *Pending I.*

Biggest risk: item 4. Biggest value: items 1–3.

---

## Investigation status (report §6)

- **A — render path:** mostly done. Component render context carries colours, no scheme; `styles.css` assembled from layout template + `css_snippets` + `buildSectionDefaults`. Remaining: read `buildSectionDefaults`/`isDarkHex` bodies (referenced, not yet pulled), and `rerender_single_page`/`CompilePageSectionsAction` assembly detail.
- **B — scheme signal:** done. Decided at composition, used for layout, dropped at both render entry points; recoverable via `layouts.scheme`.
- **G — `--section-*` mechanism:** done. `SectionStyles` ({function}-section, IsDark) + `buildSectionDefaults` luminance defaults; the dark components bypass it via class-name mismatch + inline hardcoding.
- **C — library inventory:** next. Active section + site components against scheme, plus an `html_template` and `css_snippets` scan for hardcoded hex / `--section-*` / legacy var names.
- **D — class-name contract:** next, and gates Q4. Actual layout `css_template`s vs the classes components emit.
- **E — section-contrast model:** after C/D. Where contrast intent is stored; how `is_dark_section` is set at component creation.
- **F — header/footer wiring:** after C/D. `update_site_defaults` + the original build's header choice.
- **H — migration/backfill:** with the design. Re-render fan-out; keep dark sites dark.
- **I — scheme-coherence guard:** with the design. Auditor check.

## Phasing (one bundle at a time)

- **Bundle 1 (done):** A + B + G → scheme-signal/render-path trace. Outcome above.
- **Bundle 2 (next):** C + D data → settle Q4. (SQL in the runbook.)
- **Bundle 3:** E + F → section-contrast storage + header/footer wiring.
- **Bundle 4:** H + I → migration/backfill plan + the guard.
- **Implementation** (each its own agent/workflow, not a sub-workflow): plumb the scheme signal → de-hardcode components → header/footer wiring → migrate + backfill → guard.

## Architecture + reuse constraints (carried)

Reuse the `--section-*` mechanism rather than inventing a parallel one. Fix the framework, not one page. Override, not proliferation. Scheme derivation and `--section-*` computation live in Go actions, not templates or SQL branching. Any new work (e.g. a component-migration pass) is its own agent with its own workflow. Every agent is an orchestrator.

## Out of scope here

Finishing idea.uk (its post-fix page rebuild, site review, VM cutover) and the parallel chassis backlog (P2) — tracked in the parent thread's `README_002_TODO_chassis_and_idea_uk.md` and `RUNBOOK_idea_uk_chassis_site_and_vm_deploy.md`.

---
## CLOSED (2026-07-03)
The P0 this plan opened with — the resolved scheme not reaching the components; light-resolved idea.uk deploying dark chrome and dark sections — is closed on deployed evidence (RUNBOOK §SCHEME CLOSE: all nine grep checks pass; the stale-section fossil `var(--accent-color` is gone). The mechanism that won: complete the existing paired-variable standard rather than restructure — one layout patched, ten templates de-hardcoded to consume the pairs/ink, the chrome repoint + forced re-render, then a full page-build-handler rebuild. Q1–Q8 statuses above are historical; the decision record lives in SPEC_scheme_to_components.md and the journey in running_notes/RUNBOOK. Follow-on work (imagery W7a/b/c, the Go batch, the library tail, hygiene) is tracked in RUNBOOK §SCHEME CLOSE → REMAINING WORK.
