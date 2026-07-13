# Report — a site's colour scheme does not reach its components

**Status:** investigation scoping for a dedicated fix thread. This is a fundamental, framework-level gap. It is NOT to be fixed in passing — this report exists so a future thread can start from facts and a clear thesis rather than re-discovering them.

**How this surfaced:** idea.uk was re-resolved onto the `tool-portal-light` layout with a parchment palette, `styles.css` rendered and deployed correctly (verified — no drift), yet the deployed pages still render with a dark gradient header, a dark image hero, a dark CTA band and a dark footer. Only two sections (`tool-list`, `latest-news`) actually went parchment.

---

## 1. The problem in one paragraph

A site's **scheme** (light vs dark) is decided at the composition layer — `design_intent.style_direction` drives the layout's `scheme` and the palette. The palette reaches `styles.css` `:root`, so any component that *reads* the palette variables (`--color-*`) adapts to it. But scheme is only half-plumbed: the **colour values** flow to components, while the **light/dark structural treatment** (text colour, surface, image overlay, which sections contrast the page) does not. Components encode that treatment themselves — many hardcode a dark background and set a dark section context (`--section-*: white`) directly in their own CSS — so they ignore the site's scheme entirely. There are no light counterparts, and a section resolves to exactly one component per function, so there is nothing scheme-appropriate to fall back to. The result: a light composition renders dark chrome, because the components are dark by construction and the scheme decision never reaches them.

---

## 2. What we have confirmed (facts, not guesses)

These are established from code + live data this session; the next thread can rely on them.

- **`styles.css` is correct.** The deployed stylesheet is exactly `tool-portal-light` with the parchment palette in `:root` (`--color-background #EFE7D6`, `--color-accent #A8391A`, ink `#1A1816`). No LLM drift. So composition → stylesheet works end to end.
- **Pages render dark via the components' own inline CSS, not the layout.** The layout styles classes like `.hero-section`, `.tools-grid`, `.tool-card`; the components emit *different* classes (`.hero`, `.cta-section`, `.tl-card`) and carry their own `<style>`. So the layout's section rules largely do not apply. What the layout *does* contribute that reaches components is the `:root` palette variables and the generic element rules (`body`, `h1–h6`, `a`). That is precisely why `tool-list`/`latest-news` went parchment (their CSS reads `var(--color-background)`) while `hero`/`cta`/header/footer stayed dark (theirs hardcodes dark or sets `--section-*` white).
- **Section → component is a direct function lookup for all current sites.** `plan_sections` Path 1 looks the section name up directly against `content_components.function`; its own comment says "all current sites hit this path." The scoring selector (`component_selector`, Path 2) only runs for `section_type` names that are not real functions. So for idea.uk there is no selection step — each section maps to *the one active component* for that function (unique-active-function index).
- **`component_selector` ignores scheme.** It loads `is_dark_section` into the candidate but never scores on it; scoring is `suitable_site_types` / `suitable_page_types` / quality / usage.
- **No layout declares default header/footer.** `tool-portal-light`, `tool-portal-dark`, `brochure-formal` all have NULL `default_header_component_id` / `default_footer_component_id`; the new style collection inherited NULLs; `site-design-planner` does not run `update_site_defaults`; idea.uk's `site_components` still point at the original `site-header` / `site-footer` (both `is_active=false`, rendering dark via hardcoded inline CSS).
- **The library is dark-oriented.** Active `hero` and `call-to-action` are `is_dark_section=true`; there is no light `hero`, light `call-to-action`, or light footer. Light *headers* do exist (`header-with-categories`, `header-with-cart-or-nav`).
- **The hero var-name bug is already fixed in the library.** The active `hero` uses `var(--color-accent …)`; the inactive twin used `var(--accent-color …)` (which fell back to navy `#0f3460`). The deployed page is stale — a page rebuild picks up the correct (rust) button with no code change.
- **A renderer mechanism for adaptive sections already exists** (per `003 — Contracts & Standards`, "CSS Colour Inheritance Model"): the renderer is meant to append `--section-*` defaults based on the section background's luminance, and dark sections override `--section-*` on their container. The dark components **bypass** this by hardcoding `--section-*` and a dark background. So the adaptive vehicle exists; the dark components opt out of it.

---

## 3. Where we went wrong (the architectural diagnosis)

1. **Scheme was modelled as colour-only.** Light vs dark is also *structural*: text colour, surface, overlays, and which sections are meant to contrast the page. The colour half was plumbed (palette → `:root` → `--color-*`); the structural half never was.
2. **Components own the light/dark decision instead of receiving it.** A "dark hero" bakes its darkness in (dark background + `--section-*: white`). So the composition's scheme decision and the component's rendering are decoupled — nothing tells the component "this site is light."
3. **The library grew dark-first.** The most-used / active components are dark, so "default" is effectively dark and light is unrepresented. A light site has nothing light to render with.
4. **`is_dark_section` is weak.** It is (a) not used in selection, (b) unreliable (idea.uk's `site-header` is flagged `is_dark_section=false` yet renders a dark gradient), and (c) conflates two different ideas — "this component is intrinsically dark" vs "this section should contrast the surrounding page."
5. **Two parallel styling systems that barely overlap.** Layout `css_template` styles one class vocabulary; components self-style another. So the layout cannot re-skin the components even when we change the layout. This is the practical reason the scheme work didn't land on the page.

---

## 4. The thesis for the fix (built on your steer)

Your steer — *a component's scheme should be an **override**, not a whole new function, unless it is too structurally complex to share* — is the right shape, and it changes the fix substantially. Worked through, it implies:

**Scheme is a set of variable values (an override layer), supplied by the composition/renderer and consumed by de-hardcoded components — not a second copy of each component.**

Concretely, the likely shape (to be validated by the investigations, not assumed):

- A component declares its **structure once** and reads scheme-aware variables for all colour/contrast: `--color-*` for palette, `--section-*` for the section's text/surface/border context. It hardcodes nothing and sets no `--section-*` itself.
- The **scheme override is the values** those variables take. The site's base scheme (light/dark) sets the palette + base `--section-*`; a *per-section contrast intent* (this band should stand out) sets a contrasting `--section-*` on that one section. Both are data applied at render time, derived from scheme + intent + palette luminance — exactly what the existing `--section-*` mechanism is for.
- **New functions only for genuine structural divergence** (different HTML/layout between light and dark, which should be rare). Colour/scheme differences never justify a new function. This explicitly rejects a `hero-light` / `hero-dark` split.
- Because one component adapts, **scheme-aware *selection* becomes secondary.** Direct-function resolution (Path 1) is fine if the component adapts. So the fix is mostly about (a) the scheme signal reaching render, (b) de-hardcoding components to consume it, (c) a clean model for per-section contrast — and much less about the selector.

This also realigns with the standing principle that a component's inline `<style>` is an **override, not the main CSS**: today the dark components violate that by carrying a full dark treatment inline. The fix pushes colour/contrast back onto variables the composition controls, leaving the inline CSS as genuine structure-only overrides.

The hard part is not writing this — it is the **design questions in §5**. They must be answered before any code, because the wrong answer (e.g. duplicating components, or making every section blindly match the site scheme and losing intentional contrast) would be worse than today.

---

## 5. The core design questions the dedicated thread must resolve

1. **Where does "scheme" live at render time, and how does it reach a component?** Today components get the palette (via `:root`) but no explicit scheme flag. Options: a render-context value (`{{.scheme}}`), a body/section class (`.scheme-light`), or purely via `--section-*` values. Which is cheapest and fits "complexity in Go, simple templates"?
2. **Is a section's darkness a site property, a component property, or a per-placement design choice?** A dark hero on a light site is a *legitimate intentional* contrast — "make everything light" is wrong. We need a model: **site scheme (base) + per-section contrast intent**. Where does contrast intent get decided and stored (site plan? composition? auditor?) and how does it reach `--section-*`?
3. **What is the override mechanism, exactly?** Renderer sets `--section-*` + section background from (scheme + intent + luminance) and components read them? A modifier class with CSS rules? The layout owns section backgrounds and the component owns only structure? Each has consequences for the two-styling-systems problem (Q5).
4. **Reconcile the component class vocabulary with the layout class vocabulary.** Should components adopt the layout's class contract (so `styles.css` can style them, and the inline `<style>` shrinks to true overrides), or keep self-styling but become strictly palette-/section-variable-driven? This is the single biggest structural decision and gates everything else.
5. **What becomes of `is_dark_section`?** Keep, fix, or replace with a per-placement contrast intent? It is currently unreliable and unused in selection.
6. **Header/footer.** They are site-level (`site_components`), not section-resolved. Same override principle (one adaptive header/footer driven by scheme), plus the wiring gap (no layout default; `update_site_defaults` not run by the composition path). How do they become scheme-derived and adaptive?
7. **Migration + backfill without breaking dark sites.** Changing shared component templates must keep existing dark sites dark and re-render every affected site. What is the safe sequence (template change → which sites → page rebuild vs re-render)?
8. **A guard so it can't silently regress.** Can the design auditor (improvement loop) detect "section scheme does not match site scheme / unintended contrast" so this class of bug is caught automatically in future?

---

## 6. The investigations needed (with the artifacts each requires)

Each investigation is scoped to answer specific questions above. The artifacts are what to pull up front.

**A. Map the exact render path for a section and for a site component.**
Answers Q1, Q3, Q4. Read how stored components become final HTML and where any `--section-*` is injected.
Need: `rerender_single_page` action (assembly), `rerender_page_sections_action.go`, `render_site_components_action.go` + `inject_header_footer.go` (header/footer render), `RenderComponentAction` / `CompilePageSectionsAction` / `SavePageSectionsAction`, and the code that "appends `--section-*` based on luminance" (find it). Goal: identify the one place scheme/section-context could be injected.

**B. Trace the scheme signal end to end and find where it stops.**
Answers Q1. Where scheme is decided (`design_intent.style_direction`, `layouts.scheme`, `resolved_composition`) and what the render context actually carries.
Need: `BuildRenderContextAction` / `contextToInterfaceMap`, `render_css_from_spec_action.go` + `render_css_composition_loader.go` + `render_css_composition_helpers.go` (have these — re-read for what reaches a template), the `css_themes.css_content` shape. Goal: the cheapest insertion point for a scheme signal.

**C. Inventory the whole component library against scheme.**
Answers Q3, Q5, Q7. For every active section + site component: `is_dark_section`, whether the template hardcodes hex vs reads `--color-*`, whether it sets `--section-*` itself, and which classes it emits.
Need: `SELECT function, component_level, section_type, is_dark_section, suitable_site_types FROM content_components WHERE is_active = true ORDER BY component_level, function;` plus a scripted scan of `html_template` for hardcoded `#hex`, `--section-*` assignments, and legacy var names (`--accent-color`, `--primary-color`, `--color-white`). Goal: quantify how many components need de-hardcoding and which (if any) are genuinely too complex to share (and so justify a new function).

**D. Audit the css_template ↔ component class-name contract.**
Answers Q4 (the gating decision). Compare each active layout's `css_template` selectors against the classes components emit.
Need: `SELECT name, css_template FROM layouts WHERE is_active = true;` + the component templates from C. Goal: decide adopt-the-layout-contract vs variable-driven-self-styling.

**E. Clarify the section-contrast model.**
Answers Q2, Q5. How the site plan represents sections and whether it can carry a per-section emphasis/contrast flag; how `is_dark_section` is set when a component is created.
Need: `write_site_plan_action.go` (have it — re-read for the section shape), `StoreGeneratedComponentAction` (how `is_dark_section` is decided), `021_site_spec_and_classifier.md`, `026_component_regeneration_flow`. Goal: a coherent "base scheme + per-section contrast" model.

**F. Header/footer scheme + wiring.**
Answers Q6. How the original build chose idea.uk's `site-header`/`site-footer` given no layout default; what `update_site_defaults` does and where it runs.
Need: `update_site_defaults_action.go`, `render_site_components_action.go`, `WriteBuildItemsAction` / `apply_adoption_plan_action.go`, and the live wiring (`SELECT slot_name, component_id FROM site_components …`, `style_collections.{header,footer}_component_id`). Goal: scheme-derived, adaptive header/footer.

**G. The `--section-*` luminance mechanism (the key existing lever).**
Answers Q3. What it does today, why dark components bypass it, whether it can be the single adaptation point.
Need: the appender code from A; `003` CSS Colour Inheritance section (have it). Goal: reuse vs replace decision (prefer reuse, per guidelines).

**H. Migration + backfill safety.**
Answers Q7. How template changes fan out re-renders and keep dark sites dark.
Need: `026_component_regeneration_flow` (re-read), `StoreGeneratedComponentAction` (`needs_rerender` fan-out), `flag_page_image_rebuild_action.go` (the reusable page-rebuild trigger), `check_unresolved_sections.go`. Goal: a staged migration + per-site backfill plan.

**I. A scheme-coherence guard.**
Answers Q8. Whether the auditor can flag scheme mismatches.
Need: `004_improvement_loop.md`, the design-audit-agent definition, `023_llm_quality_testing.md`. Goal: an automated check so this does not regress.

---

## 7. Consolidated info / code / schema / data checklist for the fix thread

Pull these at the start of the dedicated thread.

**Agent definitions** (`SELECT type, jsonb_pretty(default_config) FROM agent_definitions WHERE type = …`): `page-build-handler` (have), `page-rerender` (have), `webdesign-agent` (have), `site-design-planner` (config known), the design-audit-agent, `component-creator`, `rerender-pages`, `image-build-handler`.

**Go actions to read:** `rerender_single_page` (assembly) + `rerender_page_sections_action.go`; `render_site_components_action.go` + `inject_header_footer.go`; `RenderComponentAction` / `CompilePageSectionsAction` / `SavePageSectionsAction`; the `--section-*` luminance appender (locate); `BuildRenderContextAction` / `contextToInterfaceMap`; `update_site_defaults_action.go`; `StoreGeneratedComponentAction`; `install_site_composition_action.go` (have). Already have: `plan_sections_action.go`, `component_selector.go`, `write_site_plan_action.go`, `render_css_*`, `flag_page_image_rebuild_action.go`, `check_unresolved_sections.go`, `nav_tables.go`.

**Docs:** `003` (have), `026 Component Regeneration Flow`, `027 / 026 Design Composition` (have), `021 Site Spec & Classifier`, `004 Improvement Loop`, `023 LLM Quality Testing`, `001 Development Guide` (the component contracts + checklist).

**Schemas (`\d`):** `content_components` (all columns — esp. `is_dark_section`, `section_type`, `component_level`, `suitable_site_types`, `html_template`), `page_components`, `site_components`, `style_collections`, `layouts` (esp. `scheme`, `default_*_component_id`, `css_template`), `css_themes`, `site_specs` (the `site_plan` shape).

**Data (queries):** the full active-component scan (C); all active layouts' `css_template` (D); a couple of dark sites' rendered pages for comparison (to confirm the same components render correctly dark — so the fix must preserve that); idea.uk's wiring (have).

**Reference site:** a known-good DARK site that uses the same components, so the fix can be checked against "dark still works" as well as "light now works." (Identify one in the thread.)

---

## 8. Provisional fix shape (to validate, not to build yet)

Stated so the thread has a hypothesis to test or reject:

1. Introduce an explicit **scheme signal** into the render context from the resolved composition (likely a single value + a per-section contrast intent).
2. **De-hardcode** the dark components (`hero`, `call-to-action`, the site header/footer, and any others found in C): read `--color-*`, stop setting `--section-*` and dark backgrounds inline. Keep one function each (override, not variant), per your steer.
3. Make the **renderer the single adaptation point**: it sets each section's `--section-*` and background from (site scheme + section contrast intent + palette luminance), reusing the existing `--section-*` mechanism.
4. Resolve the **two-styling-systems** question (D): preferably components converge on the layout's class contract so `styles.css` carries structure and the inline `<style>` shrinks to true overrides.
5. **Header/footer**: one adaptive component each; wire via `layouts.default_*_component_id` + run `update_site_defaults` in the build path.
6. Keep **direct-function resolution**; treat `component_selector` scheme-awareness as a minor follow-up only for genuinely distinct variants.
7. **Migrate + backfill** carefully (H): change templates, keep dark sites dark, re-render affected sites.
8. Add the **scheme-coherence audit** (I).

Each step is contingent on the investigations. The biggest risk is item 4; the biggest value is items 1–3.

---

## 9. idea.uk in the meantime (the agreed minimum)

No structural change now. The active `hero` already carries the corrected `--color-accent`, so a future page rebuild would render the rust button and let the var-reading sections pick up parchment; the dark header/hero/CTA/footer remain until the structural fix lands. idea.uk continues to run on the VM (the live £29 tool is untouched); the chassis build is staging only (DNS points at the VM), so none of this is user-visible. The VM cutover stays gated on this fix being done properly.

---

## 10. Alignment with guidelines and mission

- **Reuse before rebuild:** the `--section-*` luminance mechanism already exists — the fix should reuse it as the adaptation point rather than invent a parallel one.
- **Fix the framework, not one page:** this is explicitly a framework fix; idea.uk is just the symptom that exposed it.
- **Override, not proliferation** (your steer): scheme = variable-value overrides consumed by one component, new functions only for true structural divergence.
- **Simple workflows, complexity in Go:** scheme derivation + `--section-*` computation belong in Go actions, not in templates or SQL workflow branching.
- **Every agent an orchestrator / spawn sub-agents, not sub-workflows:** if new work is needed (e.g. a component-migration pass), it is its own agent with its own workflow, not a sub-workflow bolted onto an existing one.
- **Mission:** the platform is meant to build sophisticated, industry-appropriate sites that are "best for the users of the site." A coherent, intentional colour scheme — not a dark/light patchwork that falls out of which components happened to exist — is part of that. This gap directly undercuts the design quality the mission promises, which is why it is worth a dedicated thread.
