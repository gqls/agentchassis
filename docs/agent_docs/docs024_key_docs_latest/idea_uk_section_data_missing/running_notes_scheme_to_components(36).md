# RUNNING NOTES — scheme reaches components (dedicated P0 thread)

The journal for the dedicated "make a site's scheme reach its components" thread. Memory is OFF; this doc is the journal. **Present this file at the END OF EVERY TURN.** Continues from the parent chat's `running_notes_2.md` checkpoint (qqq); the parent will resume its own notes once this fix is done. Companions: `PLAN_scheme_to_components.md`, `RUNBOOK_scheme_to_components.md`, and (source) `REPORT_scheme_does_not_reach_components.md` + `HANDOFF_scheme_to_components.md`.

═══════════════════════════════════════════════════════════════════════
CARRY-OVER STATE
═══════════════════════════════════════════════════════════════════════

## Standing preferences (STRICT)
Go not Python; plain human language, no LLM-hype, no flattery; confirm live API/schema/data facts before asserting or coding (`0 rows` not decisive until the query/state is checked); reuse and adapt before rebuild; fix the framework structurally over one-offs; honest caveats and pushback, including correcting my own reads; British English; low risk appetite; reasonable steps; ≤1 question per reply; don't create summary docs unless asked; minimal formatting (prose over bullets) in replies; banned words "perfect"/"critical"/"excellent"; no `logger.Debug` (it doesn't show — use `logger.Info`); don't call a fix "final"/"last"; no congratulation. Keep the runbook + this journal current. SQL run as FILES via `kubectl … < file`.

## Architecture conventions
Every agent is an orchestrator owning a workflow of steps that call actions. Keep workflows simple; put complexity in Go actions. Don't build sub-workflows in SQL — spawn sub-agents with their own workflows. Agents respond to the caller's (parent's) responses topic. Keep workflow variable names in sync with what the actions expect. Check DB schemas before writing SQL.

## Project facts
- **idea.uk** — LIVE Go service selling £29 reports; single binary under systemd on a Hetzner VM, nginx + LetsEncrypt, 127.0.0.1:8080; orders in a file (no DB); live Stripe webhook. Reserved tool paths: `/request /confirm /approve /decline /stripe/webhook /internal/* /order/*`. DNS (Cloudflare) → the VM, so chassis B2 deploys are invisible to the live site. UNCHANGED.
- **Chassis website-builder** — multi-agent Go/Kafka/Postgres in k8s; domain → multipage site → static → Backblaze B2 (github → GH Actions → B2).
- idea.uk chassis staging site_id `1244516d-014d-421c-88c6-090bb1e9552a`. (Note: the parent runbook's general-commands block also shows `97ed2f64-…`; treat as a separate/earlier row and confirm before relying on either.)
- DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`. k8s: `-n ai-persona-system`, `-n kafka`; cluster `personae-kafka-cluster`.

## The fix in one line
Make the site's scheme reach its components as a variable-value override consumed by one component (palette + `--section-*`), not `*-light`/`*-dark` duplication; framework-level; design before code; migrate + back-fill without regressing dark sites.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sa) — bundle 1 assembled and read; two setup issues
═══════════════════════════════════════════════════════════════════════
Built bundle 1 for investigations A (render path) + B (scheme signal) + G (`--section-*` mechanism), scoped to settle the gating Q4, via the `cmd/bundle` read-only composer (analyser + assembler + dbcontext). Two issues with the run, recorded so the next bundle is clean:
1. The two `-doc` paths (`docs/agent_docs/docs024_key_docs_latest/003…`, `…002…`) did not resolve — the bundle recorded `no such file or directory`. The standards are NOT in the bundle; read 002/003 from project knowledge instead. Fix the path for bundle 2 (`ls` the docs024 dir first).
2. The in-scope code came through as SIGNATURES ONLY, no function bodies — consistent with `-step framing` emitting structure rather than source (the parent example used `-step debug`); a wrong `-root` could also do it. Use `-step debug` next. The six schemas and the runtime evidence came through fully; the `render_css_*` bodies were the gap and were supplied separately afterward.
Runtime evidence (idea.uk/index): the only recent errors are 2026-06-21 `write_site_spec` "missing required fields: [spec_data]" on `persist_mission`/`persist_roadmap`/`persist_roadmap_brief` (a separate spec-write issue, not the scheme problem); the page's work-item history is three completed `page_rerender`/`needs_page` items from 2026-06-21 (the stale dark build).
No code/DB changes. idea.uk untouched (staging).

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sb) — investigation B settled: where the scheme signal stops
═══════════════════════════════════════════════════════════════════════
Traced scheme end to end from the bodies:
- `deriveSchemeFromDesignIntent(style_direction, suggested_style)` → `light`/`dark`/`""` (substring match: dark; or light/warm/editorial/paper/bright/soft → light).
- `resolveLayoutByTags` uses scheme as a near-hard constraint to pick the layout.
- `buildResolvedCompositionSpec` records `css_theme_id`/`palette_id`/`layout_id`/`typography_set_id` (+ names, lineage, reasoning, resolved_by/at) — **NOT the scheme value.** So scheme survives downstream only on `layouts.scheme` (schema-confirmed; check constraint `light`/`dark`/`neutral`; no scheme column on `style_collections`).
- It reaches `styles.css` (palette `:root` + the section mechanism, see Sc) but **stops before the component render context**: `RenderContext` (component_library.go) has palette-colour fields (`PrimaryColor`/`AccentColor`/`TextColor`/`BackgroundColor`) and content but **no scheme field and no `--section-*`**; `BuildRenderContextAction` merges sources into that struct, so a source carrying "scheme" has nowhere to land; `contextToInterfaceMap` exposes only struct fields. So a component template (`RenderTemplate(template, ctx)`) gets colour values but no light/dark signal — which is exactly why dark components hardcode their darkness inline.
Injection point (Q1/Q3 lean): carry the base scheme as an explicit value into the render context (add `RenderContext.Scheme`, populated from `layouts.scheme` via `sites.style_collection_id → style_collections.css_theme_id → css_themes.layout_id → layouts.scheme`) and into the CSS loader (add `l.scheme` to its SELECT + a `themeComposition.Scheme`), and let the renderer own per-section `--section-*`. Reuse, not a parallel mechanism.
Aside (standing-pref violation in existing code, noted not actioned): `BuildRenderContextAction` uses `logger.Debug` for its source-merge tracing — those lines won't show, which will hamper debugging the very plumbing we're about to add. Worth switching to `logger.Info` when we touch it.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sc) — investigation G settled: the CSS render path + the class-name contract
═══════════════════════════════════════════════════════════════════════
Read `render_css_from_spec_action.go` + loader + helpers in full. `RenderCSSFromSpecAction` assembles `styles.css` in three appended parts:
1. The layout `css_template` rendered as a Go `text/template` with FuncMap helpers `{{palette}}`/`{{typo}}`/`{{token}}` over merged maps. Merge rules (helpers): palette core slots (primary/secondary/accent/background/surface/text/text_muted/border) spec-wins; specialised slots theme-wins; typography spec-wins; structure layout-only. Theme composition loaded via `sites.style_collection_id → style_collections.css_theme_id → css_themes` joined to `palettes.colours` + `layouts.css_template`/`structure_tokens` + `typography_sets.fonts`/`scale`; hard-errors on NULL FKs.
2. **Component CSS from a `css_snippets` table** (`name`, `css_content`, `applies_to` jsonb), where `applies_to` overlaps the site's component list — appended to `styles.css`. This is a THIRD CSS surface distinct from `content_components.html_template` inline `<style>`. Where the dark hero/CTA/footer treatment actually lives (inline vs snippet vs both) is NOT yet known → inventory in C.
3. A luminance-based `--section-*` defaults block via `buildSectionDefaults(bg, surface, mergedPalette, backgroundIsDark, surfaceIsDark)` — renderer-owned, picks readable section colours from the palette. (`buildSectionDefaults`/`isDarkHex` are referenced but defined in a file not yet pulled — locate next.)
The section-context mechanism has two halves: `SectionStyles` (a `[]{Function, ClassName, IsDark}` passed to the layout template, with **`ClassName = function + "-section"`** and `IsDark` from `content_components.is_dark_section` via `queryDarkSectionsForCSS`) and the `buildSectionDefaults` global block. **The class contract is `{function}-section`** — the stylesheet assumes components emit `hero-section`/`call-to-action-section`, but the report's evidence is they emit `.hero`/`.cta-section`/`.tl-card`. That mismatch (+ the components' inline dark CSS winning) is the concrete, code-level form of Q4. Neither render path reads `layouts.scheme` (loader SELECT omits it; RenderContext has no scheme field) — so the section dark/light today keys off per-component `is_dark_section` and the palette's background luminance, never the site's resolved scheme.
Q4 leaning: **(a) components adopt `{function}-section`**, because `SectionStyles` + `buildSectionDefaults` are already built for that contract (maximum reuse, matches the report). Decision gated on an actual layout `css_template` (D) — does its `SectionStyles` block emit section STRUCTURE + context, or only context?

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sd) — docs written; next is bundle 2 (C + D → Q4)
═══════════════════════════════════════════════════════════════════════
Wrote this thread's three working docs (per request): `PLAN_scheme_to_components.md` (problem, thesis, Q1–Q8 status, fix shape, investigation status, phasing), `RUNBOOK_scheme_to_components.md` (corrected bundle command, the C and D SQL, verification queries, gated migration sketch, rollback/safety), and this journal. No code or DB changes this thread — still the design phase.
Status of the investigations: A mostly done (remaining: `buildSectionDefaults`/`isDarkHex` bodies; assembly detail), B done, G done. Open and gating: Q4, needing the C inventory (incl. `css_snippets` and the section-uniqueness check) and the D class-name pull. Then E (per-section contrast storage; `is_dark_section`'s role), F (header/footer wiring + `update_site_defaults`), H (migration/backfill), I (scheme-coherence guard).
NEXT: run bundle 2 — the C and D SQL in the runbook (against `clients_db`, as files) — and bring back the results to settle Q4. idea.uk unchanged; live £29 VM tool untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Se) — bundle 2 (C inventory + D component side) done; Q4 shape clear, two confirmations outstanding
═══════════════════════════════════════════════════════════════════════
Ran C1/C2/C3/C4/C5 + the verification joins. Findings:
- **C1/C2 (inventory):** ~40 active sections are `is_dark_section = t` and most self-declare `--section-*` inline (`sets_section_vars = t`) — that self-declaration is the bypass that overrides the renderer. C4: **no section function has >1 active component** (the "one active per function" premise holds for sections, though the UNIQUE index only covers `component_level='tool'`). The C2 legacy-var flag is COARSE — it catches benign `--color-white`; the genuinely buggy `--primary-color`/`--accent-color` (navy fallback) is in the INACTIVE hero twin. A few actives (social-proof, testimonials, portfolio-showcase, contact-form) trip the flag and need a body check before assuming they're buggy.
- **C3 (css_snippets):** cleared. None of the 21 snippets set `--section-*`; they are animation/button/card utilities + two news-grid templates (Latest News Grid, News Listing Page). So the dark-section treatment lives in the component `html_template` inline `<style>`, NOT `css_snippets` — de-hardcoding targets the inline `<style>`; blast radius narrowed.
- **C5 (D, component side):** the `{function}-section` contract is honoured INCONSISTENTLY. `footer-with-disclaimer` → class `footer-with-disclaimer-section` (matches) and reads `var(--section-*)` throughout — the MODEL; its only fault is self-declaring the dark `--section-*` values + `--color-footer-bg, #1a1a2e`. `hero` → `.hero`, `call-to-action` → `.cta-section` — neither matches the contract; both self-declare dark `--section-*`. Hero's darkness = full-bleed dark image overlay `rgba(0,0,0,.5–.6)` + hardcoded white text / dark gradient fallback → the structurally-distinct one (scheme-aware internal logic candidate; possible new-function case). CTA's darkness = deliberate contrast (primary bg, white text) → the per-section-intent case.
- **Assembly (`CompilePageSectionsAction`):** the compiler does NOT wrap sections — it concatenates each component's own HTML and optionally injects head/header/footer. So `{function}-section` is the COMPONENT's responsibility; the renderer provides no wrapper hook.
- **Verification:** recovery join works — idea.uk → `style_collection` → `css_theme` → `layout` = `tool-portal-light`, `scheme=light` ✓ (the lookup the fix adds is sound). No active layout sets `default_header`/`default_footer` (all NULL). idea.uk's `header`/`footer`/`head` slots are pinned via `site_components.component_id` to rows that are now `is_active = f` and DISTINCT from the active `site-header`/`site-footer` — idea.uk is not tracking the active library. Header/footer render via the separate `InjectHeader`/`InjectFooter` site-component path, so their darkness is their own inline `<style>`, independent of `is_dark_section`. → concrete evidence for the F bundle (leave for F).
- **Q4 lean:** (a) components adopt `{function}-section` AND become strictly variable-driven (consume `var(--section-*)`/`var(--color-*)`, self-declare nothing). `buildSectionDefaults` sets the base `--section-*` (site scheme, inherited by every section); a per-section contrast override is scoped by the stylesheet to `.{function}-section`. The footer is the proof-of-concept. New function only for genuine divergence (hero).
- **Outstanding to lock Q4:** D1 (`SELECT name, css_template FROM layouts WHERE is_active = true` — does the layout's `SectionStyles` loop emit per-`.{function}-section` rules, the home for the contrast override?) and the `buildSectionDefaults` body (does it set the base `--section-*` GLOBALLY, so a de-hardcoded component inherits and adapts?).
NEXT: run D1 (tool-portal-light/-dark `css_template`) + paste `buildSectionDefaults` (grep in the runbook) → lock Q4. Then E (per-section contrast storage; `is_dark_section`'s role) and F (header/footer wiring — idea.uk pinned to inactive components, layouts carry no defaults). No code/DB changes yet; idea.uk untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sf) — D1 + buildSectionDefaults read; Q4 SETTLED = option (a)
═══════════════════════════════════════════════════════════════════════
Read all 18 active layout `css_template`s + `buildSectionDefaults` (in `color_util.go`, with WCAG `isDarkHex`/`pickReadableOnBackground`). Decisive findings:
- **`--section-*` is a DARK-ONLY override; light is the fallback.** `buildSectionDefaults` returns "" unless bg or surface is dark. Whole-site-dark → a `body { --section-* }` block (palette-aware text); dark surface → same block on the 5 fixed surface classes. A fully light site gets NOTHING, and the layout element rules `var(--section-text, inherit)` / `var(--section-heading, var(--color-text))` fall back to `--color-*` (dark ink on light). So `tool-portal-light` renders correctly purely via fallback.
- **`SectionStyles` is DEAD for current sites.** None of the 18 active layouts reference `{{range .SectionStyles}}`, `.BackgroundIsDark`, `.SurfaceIsDark` or `.Components` — only `{{palette}}`/`{{typo}}`/`{{token}}` + the heading slot. So `buildCSSsectionStyles` is computed-but-unused. The live per-section mechanism = layout `:root` palette + element rules `var(--section-*, var(--color-*))` + `buildSectionDefaults` (body + 5 surface classes, Go-side) + each component's inline `<style>`. **The fix extends `buildSectionDefaults`; it does NOT wire `SectionStyles`.**
- **The `{function}-section` contract is REAL + operative, honoured unevenly.** Layouts style `.hero-section`, `.call-to-action-section`, `.testimonials-section`, `.features-section` etc.; layouts + `buildSectionDefaults` key the 5 surface classes by those exact names. Honoured by the 5 surface sections + `footer-with-disclaimer` (`.footer-with-disclaimer-section`); NOT by `hero` (`.hero`) or `call-to-action` (`.cta-section`) → layout structural rules + per-section overrides miss them, inline dark CSS wins.
- **Layouts DELIBERATELY leave hero/CTA/testimonials backgrounds component-owned** (Dark Section Variable Contract). So a component that paints itself dark on a LIGHT site gets no `--section-*` from `buildSectionDefaults` → naive de-hardcoding would give dark fallback text on a dark bg. Hence the renderer must set per-section `--section-*` for dark-by-intent sections, keyed on `.{function}-section` — which is WHY components must adopt that class (Q3 mechanism).
- **Phase 4.5 (025_migration) is the planned home:** move surface painting into components + switch the renderer from the hardcoded 5-class list to a `data-section-bg` attribute selector. The scheme fix generalises Phase 4.5 to cover scheme + hero/CTA.
- **Assumption corrected:** the scheme already reaches components IMPLICITLY via palette `:root` + luminance `buildSectionDefaults`; components defeat it by hardcoding. On `tool-portal-light` `--color-primary = #1a1a1a`, so today's hero/CTA (bg from primary) render dark; a de-hardcoded hero reading the vars renders light there automatically. So the CORE fix = de-hardcode + class convergence; explicit `RenderContext.Scheme` is SECONDARY (only structural branches like the hero image overlay, and per-section contrast intent). **This revises the Q1 emphasis in the PLAN.**
- **Q4 VERDICT = (a):** components adopt `{function}-section` AND become strictly variable-driven (consume `var(--section-*, var(--color-*))` + palette; self-declare nothing); the renderer (extended `buildSectionDefaults`) owns per-section `--section-*` for dark-by-intent sections; no `*-light`/`*-dark` variants. Hero is the one structural-divergence watch item.
NEXT: bundle 3 — E (where per-section contrast intent lives; `is_dark_section`'s fate; does the renderer set the section bg too, or only text?) and F (header/footer: `layouts.default_*_component_id` + `update_site_defaults`; idea.uk pinned to inactive `site-header`/`site-footer`/`head`). Then specify the fix. PLAN Q1/Q3/Q4 status to be revised. No code/DB changes yet; idea.uk untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sg) — bundle 3 scoped; F header/footer path read; REFERENCE doc written
═══════════════════════════════════════════════════════════════════════
Read the header/footer path bodies (component_library.go) to firm up F:
- **Selection (FINDING).** `RenderHeader`/`RenderFooter`: fetch StyleCollection (by site id, else by domain) → if `coll.HeaderComponentID != nil` use `GetComponentByID`; else `GetComponentByFunction("site-header")`; else `RenderFallbackHeader`. So the **operative store is `style_collections.header_component_id`/`footer_component_id`** — NOT `site_components`, NOT `sites.default_components`, NOT `layouts.default_*`.
- **Fallback is hardcoded DARK (FINDING).** `RenderFallbackHeader`/`Footer` emit `.site-header`/`.site-footer` with `background:<primary>` + white text (logo `#fff`, nav `rgba(255,255,255,.9)`), no scheme awareness, using the legacy `renderCtx.PrimaryColor`/`AccentColor`. primary defaults `#1a1a2e`. Any site reaching the fallback gets dark chrome.
- **`InjectHeader`/`InjectFooter`/`InjectHead` live in component_library.go** (1440/1491/1661) — no separate `inject_header_footer.go`. `InjectHeader` skips if the page already contains `class="site-header"`, refreshes nav from deployed pages, renders via `RenderHeader`, regex-replaces any existing header + trailing `<style>`/`<script>`, inserts after `<body>`.
- **`UpdateSiteDefaultsAction` (v3_site_actions.go:409) writes `sites.default_components` JSONB** — a store `RenderHeader` does NOT read on the path read so far. THEORY: intended-but-unwired default mechanism.
- **Four overlapping header/footer default stores confirmed to coexist** (the F tangle): style_collections.*_component_id (operative) / site_components slots (idea.uk pinned to INACTIVE rows) / sites.default_components JSONB / layouts.default_*_component_id (NULL). Untangling = F.

Artifacts this turn:
- Created `REFERENCE_styling_render_pipeline.md` — the full "how it works" model (pipeline, composition, 3-part styles.css, colour system, component model, header/footer path, §7 live-vs-dead inventory tagged FINDING/THEORY, §8 why light renders dark, §9 E/F pointers). This is the requested summary doc.
- Updated `RUNBOOK_scheme_to_components.md`: CURRENT POSITION (bundles 1–2 + D1 done, Q4 settled, SectionStyles-dead + implicit-scheme corrections); marked the D-section investigation resolved (its `{{range .SectionStyles}}` premise was false); added a **Bundle 3 — E + F** section (the `-scope`/`-doc`/`-schema-tables` command, the is_dark_section / update_site_defaults locate greps, and E1/E2 + F1–F4 data SQL).

NEXT: run bundle 3 + the E/F SQL → settle E (per-section intent storage; is_dark_section setter; does the renderer paint the section bg too) and F (untangle the four stores; does composition run update_site_defaults; scheme-aware fallback). Then specify the fix and revise PLAN Q1/Q3/Q4 (still reads "leaning, pending D"). No code/DB changes yet; idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sh) — bundle 3 read; E + F SETTLED
═══════════════════════════════════════════════════════════════════════
NB: bundle_e_f.md is a THIN SLICE (its runtime section says live data not included), so the F1–F4/E1–E2 live-data queries still need running separately. The schemas + code bodies + the 003 contract were enough to settle the design.

### E — section-contrast model
- **`is_dark_section` is a COMPONENT-level bool, LLM-authored.** `store_generated_component`'s `parseGeneratedTemplate` extracts `is_dark_section` from the component-creator's generated JSON (`data["is_dark_section"].(bool)`, plus a `"is_dark_section": true` string-contains fallback) and persists it on INSERT ($12) / UPDATE ($4). Not per-page, not per-section.
- **`pages.sections` is a JSONB array of section NAMES (text)** (`check_unresolved_sections` uses `jsonb_array_elements_text`). The normalised plan lives in `site_plans → site_plan_pages → site_plan_sections` (per `some_schemas`). A per-section intent, if added, belongs in `site_plan_sections`, not `pages.sections`.
- **`is_dark_section` has NO effect on the emitted stylesheet today.** `component_selector` carries it in `ComponentCandidate` but the score formula doesn't use it; `buildSectionDefaults` keys on palette LUMINANCE (`isDarkHex` on bg/surface) + the 5 hardcoded surface classes; `SectionStyles` (which could carry per-section dark) is dead. So `is_dark_section` is selection/metadata only.
- `CreateNeedsNewComponentItem` spec = section_type/site_type/page_context/description/design_direction — **no scheme/dark field**, so the LLM sets `is_dark_section` without the target site's scheme → the library skews dark independent of the site.
- **003 New-Component Checklist item 6 (Dark Section Contract): "if is_dark_section=true, template MUST set --section-* on container."** This is the EXISTING contract Q4 reverses → the fix AMENDS 003 item 6 to "the renderer owns --section-* keyed on `.{function}-section`; components self-declare nothing."

### F — header/footer wiring (grounded in 003 "Site Component Linkage Contract")
- **Intended chain:** `style_collections.header_component_id` (source of truth) → `update_site_defaults` COPIES the id into `site_components.component_id` → `renderAndStoreSiteComponent` (render_site_components_action.go) joins → `RenderTemplate` → `site_components.rendered_html`.
- **The page-compile path renders chrome independently:** `CompilePageSectionsAction` → `InjectHeader`/`InjectFooter` → `RenderHeader`/`RenderFooter` read `style_collections.header_component_id` directly and render fresh (do NOT read `site_components.rendered_html`). Both paths render the SAME component (the style_collections one), so output matches; both fall to `RenderFallbackHeader`/`Footer` if the chain breaks.
- **Operative source of truth = `style_collections.header_component_id`/`footer_component_id`.** Other stores: `site_components` (copy target + pre-render cache), `sites.default_components` JSONB (UpdateSiteDefaultsAction target — a tracking copy), `layouts.default_*_component_id` (FK, all NULL — intended layout default, never populated, nothing copies it to style_collections).
- **003 documented failure modes (lines 425–429) ARE idea.uk's case:** (1) update_site_defaults didn't run, or (2) style_collections.header_component_id NULL, or (3) site_components.component_id NULL → falls to `RenderFallbackHeader` = hardcoded DARK, no logo, stacked nav, search icon. Also the slot_name↔function mismatch (003:431–441): the fallback query `WHERE function='header'` never matches (function is `site-header`/`header-{variant}`), so it "fails silently."
- **Did composition run update_site_defaults?** Not confirmable from Go (it's a workflow step); the project SQL didn't show it; 003 lists "update_site_defaults skipped in workflow" as failure mode #1. THEORY (well-supported): the build/composition workflow doesn't reliably run it for idea.uk — confirm by dumping the planner/build workflow steps.

### Migration route (confirmed by 016 debugging guide)
- **Sections:** a template/source change needs a FULL rebuild via `page-build-handler` (`load_spec_sections → plan_sections → RenderComponent → save → deploy`), launched by a `site_work_items` insert (pipeline=build, handler_agent=page-build-handler, status=triaged), claimed by build-dispatch-loop. NOT `page-rebuild` (no plan_sections in its loop), NOT `needs_rerender` (reassembles stored page_components.rendered_html in place — does NOT re-render the component, so it would NOT pick up a template change).
- **Chrome:** rendered by `render_site_components` (separate path); a page rebuild won't refresh it. Backfill chrome = fix `style_collections` linkage → re-run `update_site_defaults` + `render_site_components` per site.
- `store_generated_component` on `regenerated` auto-creates one `needs_rerender` per affected site (003:348, deduped). BUT per the above, needs_rerender reassembles and may not re-render the changed template → a full page-build-handler rebuild is the reliable path. FLAG to confirm.

### Corrected facts
- **`site_specs` columns are `aspect` + `data` (jsonb)**, UNIQUE(site_id, aspect) WHERE is_current — NOT spec_type/spec_data (earlier runbook E2 guess — now corrected). The `write_site_spec` action's input field is `spec_data` (the 3 idea.uk runtime errors "missing required fields: [spec_data]" on persist_mission/roadmap/roadmap_brief, 2026-06-21, map into the `data` column) — incidental, separate from scheme.
- `sites.style_overrides` JSONB exists — candidate home for a site-level scheme/colour override.

### THE OPEN DESIGN DECISION (gates the fix spec)
On a LIGHT site, how is a component with `is_dark_section=true` treated?
- **A — scheme wins** (is_dark_section advisory): light site → all sections light. Simplest + fully coherent, but would lighten a hero-with-dark-image (white text on a photo must stay) — wrong for structural-dark.
- **B — is_dark_section wins** (honour literally): dark-by-intent section renders dark on every site. Preserves hero overlay + accents, but since the hero component is SHARED across sites it stays dark everywhere → the current patchwork.
- **C — scheme is the base + a per-section override** (`site_plan_sections`) for deliberate dark, and treat structurally-dark (image overlay) specially.
- **Recommendation C-lite:** de-hardcode ALL section + chrome components to follow the scheme palette vars; keep `is_dark_section` to mark genuinely structural-dark sections (hero image overlay), which the renderer paints dark + light-text via `.{function}-section`; CTA/footer become scheme-following. idea.uk goes light-coherent; structural dark preserved where real.

NEXT: confirm the rule (A/B/C), then write the consolidated fix spec — (1) components de-hardcode + adopt `{function}-section`/`.site-header`/`.site-footer` and consume `var(--section-*, var(--color-*))` + palette; (2) extend `buildSectionDefaults` to own per-section `--section-*` for dark-by-intent sections keyed on class; (3) chrome: repair the `style_collections → update_site_defaults → site_components` linkage so light collections get light chrome, + make `RenderFallbackHeader`/`Footer` scheme-aware; (4) amend 003 Dark Section Contract; (5) migrate via page-build-handler. Then revise PLAN Q1–Q8. No code/DB changes yet; idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Si) — PRIOR-ART AUDIT ("don't reinvent")
═══════════════════════════════════════════════════════════════════════
Read 025 (Phase 4.5), FOCUS-css_js_mechanisms, HANDOFF_2026-04-18, 002 §layout, 026; re-surfaced my OWN HANDOFF_scheme_to_components / REPORT_scheme_does_not_reach_components / RUNBOOK_idea_uk. Most of what bundles 1–3 "found" was already decided or framed. **Stop reconstructing; reuse.**

ALREADY DECIDED / DOCUMENTED:
- **Approach** (HANDOFF_scheme_to_components L9; REPORT L53): light/dark = a **variable-value override consumed by one component** (palette + the EXISTING `--section-*` mechanism), NOT `*-light`/`*-dark` variants; a component's inline `<style>` is an OVERRIDE, not the main CSS — de-hardcoding pushes colour/contrast onto composition-controlled vars. → my "Q4 option (a)" was restating this.
- **Light/dark at layout level** (002 §78/§90): scheme is a near-hard layout-match constraint; light/dark is handled by **paired layouts** (`tool-portal-light` is the hand-added light counterpart to `tool-portal-dark`) + the scheme-aware matcher; NOT runtime component flipping. `layouts.scheme` is the curated property.
- **Renderer-owns-section-vars** is already a contract + audited (003; HANDOFF_2026-04-18 §28/§32: "No --section-* declarations on section containers (renderer owns)" is one of the 7 audit points). The hero/CTA/chrome that self-declare dark are CONTRACT VIOLATIONS, not a design gap.
- **Phase 4.5** (025 §427–505) is the ALREADY-DESIGNED decouple: move surface-painting into components via a `data-section-bg="surface"` attribute; the renderer replaces its hardcoded 5-class list with a `[data-section-bg="surface"]` selector; **dual-write migration** (renderer first → components → delete layout block); contract edits to the Dark Section Variable Contract + Component Naming Contract. → my "extend buildSectionDefaults keyed on `.{function}-section`" should instead ALIGN with `data-section-bg` (attribute keying), the decided design.
- **F (header/footer)** was already root-caused in my own REPORT (L23/L34): no layout declares default header/footer; site-design-planner doesn't run `update_site_defaults`; idea.uk's `site_components` point at the original `is_active=false` site-header/site-footer (dark inline CSS); the library grew dark-first. Bundle 3 only added the 003 Site Component Linkage Contract chain + confirmed `style_collections` is the operative store. Not new.
- **E** was already FRAMED as the open question in REPORT (L62): "site scheme (base) + per-section contrast intent — where decided/stored?" Bundle 3 answered the storage half (no per-section contrast store today; `is_dark_section` is component-level + LLM-authored; `site_plan_sections` is the candidate home). C-lite is one way to realise the already-decided model without new storage.

EXISTING INFRASTRUCTURE — USE, don't reinvent:
- **`queryDarkSectionsForCSS` + `buildCSSsectionStyles` + `SectionStyles`** (render_css_from_spec_action.go): ALREADY query `is_dark_section` per site-component and build per-section entries keyed on `.{function}-section` with an `IsDark` flag (fallback dark list: hero / social-proof / call-to-action / testimonials). BUILT but DISCONNECTED — the 18 active layouts don't consume `{{range .SectionStyles}}`. So the per-section-dark-from-`is_dark_section` mechanism is ~80% built; **Phase 4.5 supersedes it with `data-section-bg`** → go with Phase 4.5, retire SectionStyles, do NOT reconnect the dead path.
- **`fix_forced_text_colours_action.go` → `processComponentCSS(html, function, isDarkSection, palette, minContrast) → cssFixResult`** (signature only in the bundle neighbourhood; body NOT on disk): an EXISTING post-processor for forced/hardcoded text colours, parameterised by `is_dark_section` + palette + contrast. The most directly relevant prior work to "components hardcode dark." **MUST READ before specifying anything** — it may already perform the de-hardcoding (as a render/rerender step or a fixer agent). 016 noted "the colour/CSS fixers run regex `UPDATE page_components SET rendered_html`."
- **Backfill** already documented (026 §156/§254/§379): regenerate component → one `needs_rerender` per affected site (handler `rerender-pages`; deduped `item_key=component_regen_rerender:<uuid>`); "re-render so the HTML reflects the new template." TENSION with my 016 finding (rerender reassembles in place, doesn't re-resolve) — reconcile whether rerender RE-RENDERS the component template (sufficient for a template change) or only reassembles stored HTML.

NEXT (revised — supersedes "write the fix spec"): (1) pull `fix_forced_text_colours_action.go` + its registration/callers to see whether the colour-fixer already does the de-hardcoding and whether it runs on the render/rerender path; (2) reconcile rerender-reflects-template vs patches-in-place; (3) only then specify the RESIDUAL work as "complete Phase 4.5 (`data-section-bg`) + repair the chrome linkage (`update_site_defaults` + scheme-aware fallback) + reuse the existing fixer", not hand-editing every component. No fix spec until the fixer is read. No code/DB changes; idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sj) — the existing fixers read; recommendation shifts
═══════════════════════════════════════════════════════════════════════
There IS established improvement-loop fixer infrastructure for "components hardcode colours":

- **color-variable-fixer agent** (dispatched by the improvement loop) runs TWO actions on a site:
  - `fix_hardcoded_colors`: within `<style>` blocks, dark background hex (first hex digit 0–4) → `var(--color-primary)`; 2-stop dark linear-gradient → `var(--color-primary), var(--color-secondary)`; **deliberately leaves `rgba(0,0,0,x)` overlay gradients alone** (hero image overlays). Fixes BOTH `content_components.html_template` AND `page_components.rendered_html`. Background only. Maps dark→primary, so it does NOT resolve light/dark — on a light palette whose primary is dark (idea.uk: primary `#1A1816`), the band stays dark.
  - `fix_forced_text_colors`: removes forced text colours from child text elements (h1-h6, p, li, blockquote, strong, em, cite) so they inherit `var(--section-*, var(--color-*))`; WCAG-validates the resulting text/bg against the site palette (min_contrast default 4.5), skips if too low; **for `is_dark_section` components MISSING the `--section-text` contract, INJECTS the `--section-*` contract (white text)**. Fixes template + rendered. **Keys on the component's `is_dark_section`, NOT the site scheme** → on a LIGHT site, an is_dark_section component is kept dark-with-white-text. This ENFORCES the CURRENT component-owns `--section-*` contract — the OPPOSITE of Phase 4.5 (renderer-owns).
- **nav-link-fixer agent** runs `fix_nav_link_templates`: regex href fixes on header/footer `content_components.html_template`; its header note confirms **"render_site_components (with force_rerender) must run afterwards to regenerate `site_components.rendered_html`"** — the chrome re-render mechanism + a `force_rerender` flag.
- **`fix_component_template`** (routes on `spec.fix_type`): `inject_nav_flex_css` (stacked nav → flex), `remove_element` (e.g. search icon), `align_slot_name`, `inject_responsive_css`, `repair_template_slots`. Modifies `rendered_html` directly for `site_components` (OK — re-rendered from templates); `page_components` only `slot_name`. These are SYMPTOM fixes for exactly the `RenderFallbackHeader` output (stacked nav + search icon) — patching the dark fallback's symptoms, not the linkage cause.

IMPLICATIONS:
- The fixers ALREADY de-hardcode (template + rendered, WCAG-aware), and the chrome re-render pattern (fix template → `render_site_components` force_rerender) already exists → this is the migration vehicle, not something to build.
- BUT as-aimed they implement the CURRENT contract, scheme-blind: `fix_hardcoded_colors`→primary (stays dark on a light site); `fix_forced_text_colors` injects `--section-*`/white for `is_dark_section` regardless of scheme. Running them as-is on idea.uk would ENTRENCH dark, not lighten it.
- They pull OPPOSITE to Phase 4.5 on the same files (inject `--section-*` vs remove it). So a fork:
  - **Lane 1 (structural, decided direction):** re-aim the fixers to apply the Phase 4.5 component contract — add `data-section-bg` + `background:var(--color-surface)`, STRIP self-declared `--section-*` and forced colours — and let the RENDERER own `--section-*`. Reuses the existing fixer/dispatch/rerender machinery as the Phase 4.5 migration tool.
  - **Lane 2 (remedial):** keep component-owns, extend the fixers to be SCHEME-aware (light site → light treatment), defer Phase 4.5.
- CHROME linkage repair (`style_collections → update_site_defaults → site_components` + scheme-aware `RenderFallbackHeader`/`Footer`) is COMMON to both and is the safest first piece; the `fix_component_template` symptom-fixes become moot once the proper light header renders.

TWO CHECKS before any spec:
1. Do the layouts expose scheme-flipping chrome/section vars (e.g. `--color-header-bg`/`--color-footer-bg`, or surface/background) that a LIGHT layout sets light — or only `--color-primary` (dark)? Determines what de-hardcoded chrome/sections should CONSUME so they go light on a light site. (Read a light vs dark layout `css_template`.)
2. Does `needs_rerender`/`rerender-pages` RE-RENDER the component template (sufficient for a template change) or only reassemble stored `rendered_html`? (026 says reflects new template; 016 says patches in place — reconcile.)

RECOMMENDATION: Lane 1 — re-aim the existing fixers as the Phase 4.5 migration vehicle + do the chrome-linkage repair first (lane-independent, safe). Decision needed. No code/DB changes; idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sk) — Check 1 results + user steer (initial-render; loop as backstop; minimal architecture change)
═══════════════════════════════════════════════════════════════════════
### Check 1 results (run on cluster 2026-07-01)
- **1a — idea.uk palette (`palette-idea-uk`)**: ONLY the 8 core slots `{text #1A1816, accent #A8391A, border #1A1816, primary #1A1816, surface #E8DFCC, secondary #4A4540, background #EFE7D6, text_muted #837C72}`. **No specialised chrome slots** (no header_bg/header_text/footer_bg/footer_text/cta_bg) → chrome vars fall through to the LAYOUT's fallback.
- **1b — layouts**:
  - `tool-portal-light` (light): `--color-header-bg` fallback **`#faf8f3` (LIGHT)**, `--color-footer-bg` **`#f1ede4` (LIGHT)**, `--color-cta-bg` **ABSENT (empty)**.
  - `tool-portal-dark` (dark): header `#121212`, footer `#0d0d0d`, cta `#1e1e1e` (all dark).

### Findings
- **Chrome vars ALREADY flip correctly.** On idea.uk (light layout, no palette specialised slots) `--color-header-bg` resolves `#faf8f3` and `--color-footer-bg` `#f1ede4` — both light. So **de-hardcoded chrome consuming `var(--color-header-bg)`/`var(--color-footer-bg)` renders light on idea.uk automatically** — no need to fall back to surface/background. The header/footer variable layer is complete and scheme-correct; the ONLY reason chrome renders dark is that the components (+ the dark `RenderFallbackHeader`) hardcode instead of consuming these vars.
- **Gap: `tool-portal-light` is MISSING `--color-cta-bg`** (the dark layout has it). A de-hardcoded CTA consuming `var(--color-cta-bg)` on the light layout gets no value → needs a one-line light `--color-cta-bg` added to `tool-portal-light`, or the CTA consumes `--color-accent`/`--color-surface`. Small, targeted — not architecture.
- **Sections already flip** via `buildSectionDefaults` on the merged palette's background/surface luminance (idea.uk light → no dark overrides → variable-reading sections render dark-ink-on-light). Works for a light site. The hardcoded 5-surface-class limit only bites DARK sites with surface sections outside the 5 — the case Phase 4.5 generalises.

### Reframing (given 1b)
The scheme-flipping variable layer is MORE complete than feared: chrome bg vars flip, section contrast flips. **The components are the sole thing defeating an already-working system by hardcoding dark.** So the CORE fix = **de-hardcode the components** (consume the already-correct vars; stop self-declaring `--section-*`), which is LANE-INDEPENDENT and belongs in the LIBRARY (initial render), not the improvement loop. Plus two small gaps: chrome linkage (composition must SET `style_collections.header/footer_component_id` + run `update_site_defaults` so a light site doesn't hit the dark fallback) and the missing `--color-cta-bg` on the light layout.

### User steer (this turn)
1. **Fix at INITIAL RENDER** (library + composition/renderer). The improvement-loop fixers stay AVAILABLE as a backstop but must NOT be *required* for new builds — matches how they're already wired (dispatched on audit findings). This effectively rules OUT "Lane 2 (make the fixers scheme-aware)" as the PRIMARY mechanism, since a loop fixer is post-hoc by nature; the fixers become the backstop, not the fix.
2. **Don't rewrite base architecture without good reason** → prefer the targeted de-hardcode + lean on the existing `buildSectionDefaults`; treat the full Phase 4.5 `data-section-bg` generalisation as deferrable (only needed for dark sites with surface sections outside the hardcoded 5).

### Lane — NOT settled (user unsure); but narrowed by the steer
The steer collapses "Lane 1 vs Lane 2" into a SCOPE question within the renderer/library fix:
- **(S-min)** De-hardcode components to consume the existing vars and STOP declaring `--section-*` (let `buildSectionDefaults` own it, as it already does for body + the 5 surface classes); add `--color-cta-bg` to `tool-portal-light`; fix the chrome linkage + scheme-aware fallback. No renderer/attribute rewrite. Fixes idea.uk at initial render. Leaves the dark-site-non-standard-surface gap unaddressed (acceptable until it bites).
- **(S-full)** Also do Phase 4.5 (`data-section-bg` attribute; renderer selector generalised beyond the 5 classes). Bigger, touches renderer + all layouts + component contract. Needed only for robustness on dark sites with surface sections outside the 5.
Common to both (the initial-render core): de-hardcode chrome + hero/CTA; chrome linkage repair; scheme-aware fallback; `--color-cta-bg` on the light layout.

### NEXT (proposed)
Read the ACTUAL templates to size the de-hardcode precisely — the `hero` component (image-overlay case), one surface section (e.g. `features`), and the `site-header`/`site-footer` components — and confirm composition's chrome-selection gap (does `install_site_composition` set header/footer component IDs, or leave them NULL?). No code/DB changes; idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sl) — chrome linkage + fallback sized; component templates need Check 2
═══════════════════════════════════════════════════════════════════════
- **`install_site_composition` sets `style_collections.header_component_id`/`footer_component_id` = NULL** (comment: "webdesign-agent populates these later"). BUT grep finds NO code that writes them non-NULL — only the NULL insert + reads; `UpdateSiteDefaultsAction` reads a header id from config/collection into `sites.default_components`, it does NOT write the collection. So **`header_component_id` is effectively a DEAD column — nothing populates it** → `RenderHeader` always skips it.
- **`RenderHeader` selection (component_library.go ~1300–1367):** `style_collections.header_component_id` (always NULL → skip) → `GetComponentByFunction("site-header")` (library-wide, `is_active=true`, `forked_from IS NULL`) → else `RenderFallbackHeader` (dark). Footer identical. So chrome = "the ONE library-active `site-header` for everyone" OR the dark fallback — **no scheme awareness; structurally identical to the sections' one-active-component-per-function**. (This is the page-compile path via `InjectHeader`; the `render_site_components → site_components.rendered_html` path is separate — which one produced idea.uk's deployed chrome is a per-site diagnosis, not needed to size the fix.)
- **`RenderFallbackHeader`/`Footer` hardcode dark** (read in full): header `.site-header{background:%s}` = `ctx.PrimaryColor` (default `#1a1a2e`; idea.uk `#1A1816`, dark); logo/nav text literal `#fff`/`rgba(255,255,255,.9)`. Footer `.site-footer{background:%s;color:rgba(255,255,255,.9)}` same. Uses `ctx.PrimaryColor`, NOT `var(--color-header-bg)`.
  - **Fallback fix (small, 2 funcs ~10 lines each):** swap `%s`(PrimaryColor) → `var(--color-header-bg)`/`var(--color-footer-bg)`; `#fff`/`rgba(255,255,255,…)` → `var(--color-header-text, var(--color-text))`/`var(--color-footer-text, var(--color-text))`. Nested fallback to `--color-text` covers the case where the light layout sets header *bg* but not header *text* (`--color-text` is reliably dark on a light palette).
- **Chrome fix (minimal, initial-render):** de-hardcode the ONE active `site-header`/`site-footer` library component to consume the chrome vars + make the fallback consume them. One var-consuming component then works for BOTH schemes (vars flip) — no need to revive `header_component_id`, no light/dark header variants. Plus ensure idea.uk resolves to an active `site-header` (its own is `is_active=false`) — a per-site step.
- **Component templates (hero/features/call-to-action/site-header/site-footer) are DB data** → Check 2 dumps them (2a) + chrome text vars (2b) + a blast-radius count (2c: active section/header/footer components that hardcode bg or self-declare `--section-*`) to size the library-wide de-hardcode and inform S-min vs S-full.

NEXT: run Check 2 (in runbook), then size + settle scope. No code/DB changes; idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sm) — Check 2 read; reframe; four alternatives; handoff for Claude Code
═══════════════════════════════════════════════════════════════════════
### Corrections owned (against Sl and earlier)
- "De-hardcode the one active site-header" was WRONG: the active `site-header` (generated) is ALREADY fully variable-driven (`var(--color-header-bg, var(--color-background))` / `var(--color-header-text, var(--color-text))`) — renders light on idea.uk, needs nothing. The library fix already happened for the header.
- "Components hardcode dark" was imprecise for hero/CTA: they carry NO hex backgrounds — bands come from palette vars — the defect is the ASSUMPTION (literal #fff text + unconditional dark `--section-*` block) that the vars resolve dark.
- "Q4 settled = strip all self-declarations" is SUPERSEDED: for self-painted dark bands the declaration is load-bearing — strip it and, on a light site (renderer emits no `--section-*`), CTA text becomes dark ink on a near-black band. Stripping is only safe if renderer (B) or layout (C) supplies the context first, or if the band itself lightens.

### Check 2 findings (full digest in runbook §CHECK 2 RESULTS)
- Active `site-footer` (generated) is HALF-migrated: light `--color-footer-bg` + self-declared white `--section-*` → white-on-cream unreadable on light sites; newsletter box invisible. Live library bug; also proves the CREATOR PROMPT currently emits half-migrated components (encode the chosen contract in the prompt + 003 or drift continues).
- Chrome text vars flip too (2b) — the full chrome var layer is scheme-correct; only `--color-cta-bg` missing on light, no cta-text pair anywhere.
- Blast radius (2c): sections 84 — 15 hex bg, 37 self-declare; header-level 4 active (only 1 is `site-header` — list others via 3d), 2 self-declare; footer 1.
- Deployed idea.uk dark header is therefore PROBABLY STALE build output (current active header would render light) — but per "don't jump to conclusions", verify with Check 3a/3b (component created_at vs last_built_at; grep deployed page's header class) before concluding.

### The reframe
The scheme→variable pipeline is verified correct end-to-end. The library answers "who supplies text context for a dark-by-design band" three conflicting ways at once (component self-declares; layout paired vars — proven by chrome; renderer luminance defaults). The decision is which single owner to standardise on. 025's Phase 4.5 is questioned, not assumed: it solves the 5-surface-class generalisation (a dark-site problem idea.uk never hits), its blanket "never self-declare" conflates surfaces (hazardous) with bands (load-bearing), renderer-ownership reintroduces component intent one hop away (is_dark_section/attribute), and even within renderer-owns the ~built SectionStyles append (Go-side, no layout edits, `[data-component]` selectors) is cheaper than 025's attribute migration.

### The alternatives (detail in HANDOFF_scheme_to_components_for_claude_code.md §Decision)
Alt 0 stale-build + point bugs | Alt A codify component-owned bands (declare iff you paint) | Alt B renderer-owned via connecting SectionStyles (strip after connecting; flag hygiene becomes the surface) | Alt C layout-owned pairs (generalise the chrome pattern; cta-bg/cta-text per layout) | Alt full-025 (Phase 4.5 as written; orthogonal extra; still needs B or C for bands).
GATING QUESTION (user): on a light site, are dark hero/CTA bands acceptable design, or must light mean fully light? Alt 0/A suffice if acceptable; B or C if light must lighten bands.

### Invariant no-regret set (safe under every alternative)
Footer text fix; fallback header/footer consume chrome vars; creator prompt + 003 alignment (and re-aim `fix_forced_text_colors` to match, so the backstop doesn't fight); `--color-cta-bg` on tool-portal-light; then rebuild idea.uk (page-build-handler, not needs_rerender) + `render_site_components force_rerender`.

### Artefacts this checkpoint
Runbook: newcomer problem paragraph at top; CURRENT POSITION rewritten (supersedes "Q4 settled"); CHECK 2 RESULTS recorded; CHECK 3 added (3a staleness, 3b provenance, 3c self-declarer split, 3d other header functions, 3e chrome vars across 18 layouts, 3f hex list). NEW: `HANDOFF_scheme_to_components_for_claude_code.md` (self-contained: problem, environment, verified mechanism, Check 2, decision framework, invariants, next actions, constraints, file map).
NEXT: user answers the gating question + runs Check 3 (likely in Claude Code via the handoff). No code/DB changes; idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sn) — Check 3 read; staleness refuted; design answer → paired-variable direction
═══════════════════════════════════════════════════════════════════════
### Correction owned
- Sm's leading hypothesis ("deployed dark chrome is stale build output") is REFUTED by 3a: good chrome active 2026-05-06; idea.uk pages built+deployed 2026-07-01 12:52–12:58. The 07-01 compile ran with the variable-driven site-header active. What it injected is OPEN → Check 4a (grep deployed HTML for site-header-section vs site-header--gradient vs fallback). If the gradient header is in the deployed pages, our compile-path mapping (InjectHeader → RenderHeader, ignores site_components) is wrong somewhere — verify against reality, not the code read.
- 3f substring bug owned (Postgres returns first capture group) — the 15-row list is right, displayed hex empty.

### Check 3 digest (full in runbook)
- 3b: site_components rows = stale dark renders pointing at INACTIVE components (site-header--gradient / footer-4-column / head). render_site_components never re-ran after deactivation.
- 3c split (per-template flags, per-element review needed at fix time): HAZARD ~18 (declare dark over surface vars or nothing: footer, site-head, header-docs, contact-block, platform-comparison, lobby-grid, brief-explanation, game-master-explanation, provocation-feed, gripper-payload-calculator, archetype-grid, archetype-taster-quiz, tool-agent-complexity-estimator + the 5 hero-* page variants that paint nothing) — white-on-light bugs TODAY. BAND ~19 (paint from primary/secondary/accent + white text: call-to-action, hero, social-proof, testimonials, system-stats, product-hero, tool-cta, gauntlet-*, several tool-*) — coherent today, block "fully light". Flag hygiene: 6 declarers have is_dark_section=f → never key styling on the LLM-authored flag.
- 3d: 5 other active header/footer functions (*_pre_037) unreachable on compile path — park.
- 3e: ALL 18 layouts set all four chrome vars (consumption safe library-wide). Only 3 layouts have scheme set — 15 empty; fallback VALUES for 16 layouts uncurated → Check 4b/4c sizes the layout work.

### USER DESIGN ANSWER (gating question)
"Fully flexible: a light scheme should be able to render fully light but it could carry dark hero bands." → dark/light of a band must be a CHOICE, not a component constant.

### Direction selected by that answer: PAIRED VARIABLES (Alt C) + existing renderer base
- Alt B (key styling on is_dark_section) FAILS the requirement: a component-level flag is the same for every site → cannot lighten CTA on site X while keeping it dark on site Y. Also flag hygiene is poor (6/37 contradictions).
- Alt C satisfies the flexibility ladder with existing machinery: per-SCHEME = layout css_template pair values ({{palette "cta_bg" fallback}} pattern — already in tool-portal-dark); per-SITE = palette specialised slots (merge rules: specialised = theme-wins → a light site's palette can opt a band dark); per-INSTANCE (same site, per page/section) = site_plan_directives scope=section, later, structure exists. The pair convention part-exists: --color-primary-text already consumed by the generated header.
- Components: consume pairs (band bg + band text) or surface/background + inherited text; NEVER declare --section-*; never literal #fff on a var-painted band. Renderer's buildSectionDefaults stays as the whole-palette-darkness base/safety net. is_dark_section demoted to selection/imagery metadata. SectionStyles stays retired. 025 Phase 4.5 (surface-class generalisation via data-section-bg) deferred as a separate dark-site concern.
- Hero: image branch keeps dark overlay + white text (correct on any scheme); gradient branch keeps primary/secondary/accent band but text becomes var(--color-primary-text, #fff) so a light-primary palette stays readable; a fully-light hero remains a component/plan choice later — don't over-engineer.

### Fix shape (to be specified next, after Check 4)
1) Layout curation: add --color-cta-bg to tool-portal-light + pair text vars (--color-cta-text; audit 4b/4c for the other 16 layouts' fallback values). 2) Hazard-class = straight bug fixes (drop dark declarations; text from inherited/pair vars) — footer first. 3) Band-class = pair conversion. 4) RenderFallbackHeader/Footer consume chrome vars. 5) Creator prompt + 003 rewrite (the contract: paint from pairs, declare nothing) + re-aim fix_forced_text_colors to enforce the SAME contract as backstop. 6) idea.uk: rebuild pages (page-build-handler) + render_site_components force_rerender (also clears the stale site_components rows). Sequencing safe: 1 before 2/3 (vars exist before consumers), 4/5 parallel, 6 last.

NEXT: user runs Check 4a (deployed grep) + 4b/4c (pair curation); then draft the fix specification against the paired-variable direction. No code/DB changes; idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (So) — Check 4 read: provenance ANSWERED; spec written
═══════════════════════════════════════════════════════════════════════
### The flow that produced the deployed page (now evidenced, not inferred)
- Deployed index.html chrome byte-matches the stale `site_components.rendered_html` rows (3b): `site-header--gradient` with `#1A1816→#4A4540` baked, `footer-4-column` `#1A1816` + white, old static head. Git commits: "Rerender: <page>" (~2 wks ago) + "Update stylesheet via webdesign-agent" (~1 wk). So the pages came from the RERENDER handler (reassembles `page_components`, injects `site_components.rendered_html`) — matching 016 — NOT the build path. The build path's InjectHeader→RenderHeader mapping stands; it just never ran for these pages.
- Deployed hero consumes legacy `var(--accent-color, #0f3460)` — the now-INACTIVE hero's variable naming → **sections are stale old-template renders too**. idea.uk = fossilised early renders + rerender loop. Full page-build-handler rebuild required; `needs_rerender` would re-fossilise. (This settles the 016-vs-026 rerender tension by direct evidence.)
- Timestamp puzzle (open, non-blocking): pages deployed/updated 2026-07-01 12:52–12:58, `last_deployed_at` 12:49, but no git commits that day → likely no-diff rerender or deploy-only touch. Identify what ran 07-01.
- 4b: all 18 layouts carry header/footer/cta fallbacks EXCEPT tool-portal-light (no `--color-cta-bg`); several light layouts curate DARK footer bands (affiliate-hub #18181b, brochure-formal #1a365d, comparison-aggregator/industry-hub #0f172a, ecommerce #111827, magazine-grid #1a1a1a, social-lobby #18181b) → "light site, dark band by choice" is already a curated model. cta_bg values are accent bands mostly.
- 4c: **pair convention is ALREADY the standard** — 18/18 `--color-primary-text`, 17/18 `--color-cta-text` (gap = tool-portal-light). Plus the deployed baked head CSS consumes `--color-cta-bg/-text` (`.section--cta`) and defines `--color-hero-title/-subtitle` → REUSE these names. Direction = COMPLETION of existing architecture, not restructure — the user's "don't rewrite base architecture" instinct is served.

### Artefacts
- `SPEC_scheme_to_components.md` CREATED: decision record; the contract (no literals; backgrounds from pairs/surface/palette-band-with-on-colour/image-overlay; band painters re-export `--section-*` AS REFERENCES via the pair, e.g. `--section-text: var(--color-cta-text,…)` + color-mix derivatives; non-painters declare nothing); W1 layouts (tool-portal-light cta pair + contrast sweep) → W2 hazard ~18 (footer/site-head first) → W3 band ~19 (CTA first; hero gradient text via `--color-hero-title`; NOTE dark-site CTA band shifts primary→cta_bg, visible+intended) → W4 chrome (renderAndStoreSiteComponent falls back to GetComponentByFunction on NULL/inactive + Info log; fallbacks consume chrome vars; idea.uk repoint + force_rerender; delete misleading install comment) → W5 003 + creator prompt + re-aim fixers (key on template's own painting, never is_dark_section) → W6 idea.uk sequence (layout patch → footer/site-head/CTA → repoint+force_rerender → FULL rebuild via page-build-handler → re-grep deployed HTML). Rollback: git revert + snapshot html_template before UPDATE.
- Runbook: CHECK 4 RESULTS + POSITION AFTER CHECK 4. REFERENCE: corrections block (two assembly paths; stale sections; pair standard). HANDOFF: status line updated (decision made; SPEC supersedes its §Decision).
- Open items: what ran 07-01; InjectHead selection (head vs site-head) before repointing head slot; planned pages (news-index/guides-index/tool-audience-check) linked in nav but unbuilt.

NEXT: implement W1 (add cta pair to tool-portal-light — smallest step, no components touched) when the user gives the word, here or in Claude Code via the handoff + spec. No code/DB changes yet; idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sp) — W1 prepared (files written; awaiting cluster execution)
═══════════════════════════════════════════════════════════════════════
- First mutating change of the thread → full treatment: shell backup of the whole css_template; `w1_00_before.sql` (\d layouts per schema-first rule + anchor-present/pair-absent + the 18-layout cta-pair sweep for the contrast check); `w1_01_add_cta_pair.sql` (regexp_replace anchored on the verified `--color-footer-text` line with `\1` backreference — no whitespace guessing; `NOT LIKE '%--color-cta-bg%'` guard = idempotent; first-occurrence-only; RETURNING shows the inserted lines); `w1_02_verify.sql` (occurrence counts expect 1/1 + insertion region); `w1_rollback.sql` (value-agnostic line removal).
- Values: **neutral light band `#e9e2d3` + ink `#1a1a1a`** (contrast ≈13.5) — mirrors tool-portal-dark's curation, whose CTA is a neutral elevated band (`#1e1e1e`), not an accent; and it demonstrates "fully light" end-to-end on idea.uk, this fix's acceptance case. Accent alternative offered as a two-value swap (`#9b4020`+`#faf8f3`, ≈6.1, soft-editorial family). Per-site override stays available via palette `cta_bg`/`cta_text` slots (specialised = theme-wins).
- **W1 is inert until W3**: no consumer of `--color-cta-bg` exists yet (CTA component paints `--color-primary`; legacy head `.section--cta` rule consumes the var but nothing carries that class). Zero visual change expected. styles.css re-render deferred to W6 step 1. `updated_at` untouched pending \d.
- AWAITING: execution output of 00/01/02 + the sweep rows → compute pair contrasts, flag any layout failing 4.5, then W2 (footer + site-head fixes) next.
No other code/DB changes; idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sq) — W1 COMPLETE; head slot resolved; W2a (footer) prepared
═══════════════════════════════════════════════════════════════════════
- **W1 complete + verified** (UPDATE 1; occurrences 1/1; pair sits between footer_text and code_bg). `layouts.updated_at` left stale (no trigger; optional one-liner in runbook). Contrast sweep: 13 pass; **5 seed layouts fail 4.5 with white text** (brochure-bold 2.8, affiliate-hub 3.6, media-grid 3.8, high-energy 3.9, docs-sidebar 4.1) — same-hue darker fallbacks proposed in runbook §W1 RESULTS; curation batch, user to approve, non-blocking. **No layout defines hero-title/subtitle vars** → W3 hero reuses `--color-primary-text` (no new vars in 18 layouts); value read added to w2_00 §0.3.
- **Head slot resolved from component_library.go**: `RenderHead` looks up function `"head"` (is_active filter) → only head component is INACTIVE → build path always uses `RenderFallbackHead`. `site-head` (section-level) is unreachable as chrome; w2_00 §0.2 checks whether any page places it as a section — its fix defers behind that. Incidental: `InjectHeader` SKIPS injection if incoming HTML already carries a `site-header` class (relevant to W6 verification); `InjectHead` uses `logger.Debug` (invisible) — tidy in W4.
- **W2a prepared** (footer de-hardcode): six exact-string nested replaces (needles byte-for-byte from Check 2a), alphas preserved (90/70/heading-full/5/20 via color-mix on the footer pair), bg fallback `#1a1a2e`→`var(--color-surface)`, `updated_at=now()`, guarded + idempotent; GATE = w2_00 needles all t AND white_rgba_count=4 (else re-derive from fresh dump). Dark-site parity via the pair (tool-portal-dark footer_text rgba(224,224,224,0.8) → slightly more muted than flat white; accepted). color-mix degrade path noted. **Inert until footer re-renders.** Backup one-liner + rollback file in runbook.
- AWAITING: w2_00/01/02 output → then site-head decision (§0.2), W3 CTA conversion (§0.3 values in hand), and the curation-batch decision on the 5 failing cta pairs.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sr) — W2a COMPLETE (footer follows the pair); verify bug owned; W2b trigger prepared
═══════════════════════════════════════════════════════════════════════
- Gate passed (needles 6×t, white_rgba_count=4; backup 8384B). **w2_01 UPDATE 1** — footer declarations now reference the footer pair (color-mix 90/70/heading/5/20), bg fallback → var(--color-surface); RETURNING f/t/t. **w2_02 errored: Postgres ARE quantifier bound max 255 → `.{0,420}` invalid** (w1_02's {0,230} was under the cap — why it worked). Corrected `w2_02_verify_fixed.sql` uses substr+position. SQL self-check lesson recorded.
- **site-head PARKED** — 0 refs in pages.sections AND page_components; unreachable as chrome (RenderHead looks up 'head'). Off the W2 queue.
- **Primary pairs read**: light `#1a1a1a/#ffffff`; dark portal `#00bcd4 (cyan!)/#0b0b0b`. → Today's white-text hero is a LATENT FAILURE on tool-portal-dark (white on cyan ≈2.3); the pair (≈8.6) fixes it. W3 design detail: no-image hero = (a) keep 3-stop gradient + primary_text vs (b) solid primary + primary_text — decide at W3 prep.
- **Blast radius: only ONE status='active' site (brochure-formal); idea.uk NOT status='active'** → read its status in w2b_00(a) before W6 assumes handlers pick it up.
- **W2b prepared** (user asked for auto-updating layouts.updated_at): w2b_00 reuse-check (existing trigger fns + which tables auto-update) → w2b_01 CREATE FUNCTION set_updated_at (errors on collision by design) + BEFORE UPDATE trigger on layouts + bump stale row through it. Convention note: codebase sets updated_at explicitly; trigger coexists. Layouts-only.
- AWAITING: w2_02_fixed + w2b outputs → then W3 prep (CTA conversion files + hero branch decision + the contrast curation batch decision).
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Ss) — W2b COMPLETE via reuse; sites.status resolved from code; W3 calls framed
═══════════════════════════════════════════════════════════════════════
- Footer verify clean (f/f; region as designed). **W2b: CREATE FUNCTION errored = gate firing (shared `set_updated_at` already exists, used by site_specs/site_plans/content_feed_items/training_runs); CREATE TRIGGER bound to the EXISTING function; bump proved it (11:50:17). Complete — the reuse path, no redo.** Future note: `\set ON_ERROR_STOP on` for mutation files whose later statements depend on earlier ones.
- **sites.status resolved** (UpdateSiteStatusAction, v3:323): vocabulary draft/building/review/published/deployed/archived/error; 'active' NOT in it (legacy value on one brochure-formal site); nothing in on-disk code filters sites by status → informational only; idea.uk='deployed' fine for W6. My 0.4 `status='active'` filter was a wrong borrowed assumption — corrected blast radius: 7 sites (brochure-formal 4, social-lobby 1, portal-dark 1, portal-light 1 = idea.uk).
- **W3 calls framed for the user** (runbook §W3 PENDING DECISIONS): hero no-image branch (a gradient+primary_text / b solid+primary_text / c single-hue mix; w3_00 imageless count informs — if 0, b is free) + the 5-layout contrast batch (zero live impact — none of the 7 sites on those layouts; seed hygiene only).
- `w3_00_before.sql` staged: CTA needles ×10 (+counts 4/4 expected; n2 hits root AND btn-secondary by design), imageless-hero counts for hero + hero-%, sites.status distinct.
- AWAITING: w3_00 output + the two calls → then w3 mutation files (CTA conversion; hero per call). idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (St) — calls received (hero=c, batch=yes); guide updated; W3a+W1b prepared
═══════════════════════════════════════════════════════════════════════
- w3_00: CTA gate PASSED (10×t, 4/4). **Imageless heroes are the COMMON case — hero 80/114, hero-* 26/26** (my "rare" assumption wrong at data level; (c) vindicated; hero-* variants confirmed live white-on-transparent hazards → next hazard batch after main hero). status distinct: deployed 7 / system 1 / active 1 ('system' also outside the vocabulary — second legacy value).
- **Guide updated → `016b_debugging_guide_7.md`** (v4 log + three §9 entries: light-site-dark-chrome two-paths pattern w/ provenance greps + legacy-var tell; SQL pitfalls incl. regex≤255 + capture-group + the needle-gate surgery pattern; sites.status vocabulary + never filter blast-radius on status='active' + set_updated_at reuse-gate).
- **HERO (c) design settled** (runbook §HERO (c) DESIGN): per-branch `--hero-ink` (image branch `#fff` under the structural-dark exception; else `var(--color-primary-text)`); layered `background: var(--color-primary);` + single-hue gradient mixing **15% toward the ink** (depth on both dark and light primaries; bounded contrast cost, worst observed ≈5.6; solid layer = color-mix-less fallback); section vars from ink at 95/80/·/10/30; btn-primary = inverse pair (ink bg, primary label — visual change noted; deployed button was off-palette navy via dead legacy var anyway); btn-secondary from ink mixes. Files W3b, next turn.
- **W3a prepared** (CTA conversion): ten exact replaces → pair + inverse buttons, no literal fallbacks (all 18 layouts define both pair vars); guard + updated_at + RETURNING f/f/f/t; rollback file; inert until re-render. **W1b prepared** (batch=yes): five anchored regexp_replace hex swaps, per-row guards, trigger bumps updated_at, zero live impact; verify + rollback files.
- AWAITING: w3a_01/02 + w1b_01/02 output → then W3b hero files → hero-* hazard batch → remaining W2 hazards → W4 chrome path.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Su) — W3a + W1b COMPLETE; W3b (hero ink model) prepared
═══════════════════════════════════════════════════════════════════════
- **W3a COMPLETE**: CTA converted (UPDATE 1, f/f/f/t; verify regions exactly as designed — pair bg/colour, color-mix context, inverse buttons). Inert until re-render.
- **W1b COMPLETE**: five layouts curated (#c2410c ×2, #dc2626, #c4001d, #0369a1); cta_text untouched; **W2b trigger observed working** (updated_at 15:52:36.33–.40 without explicit SET).
- **W3b prepared** (hero → ink model, option (c)): gate uses `position()>0` (needles contain literal `%` = LIKE wildcard — new pitfall class, sibling of the two in the guide); twelve replaces incl. FOUR multi-line E'…\n…' needles disambiguating the four `color: #fff;` sites (hero-content / btn-primary / btn-primary:hover / btn-secondary); counts gate 4/7/2 (color_fff / white_rgba / #0f3460). `--hero-ink:#fff` stays in the image branch by design. Expected diffs at rebuild: 80 imageless heroes go multi-hue→single-hue primary gradient; hero button accent→ink (deployed was off-palette navy via dead legacy var).
- Queue after W3b: hero-* variants hazard batch (26 live imageless renders; templates not yet dumped) → then DECIDE hand-needles vs re-aiming `fix_forced_text_colors` (W5) for the remaining ~12 surface-painters → W4 chrome (fallbacks, repoint, force_rerender) → W6 rebuild + grep.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sv) — W3b COMPLETE; gate false alarm owned; W3c read staged
═══════════════════════════════════════════════════════════════════════
- **W3b COMPLETE** (UPDATE 1; f/f/f/t/t; verify regions exactly as designed — ink in both inline branches, layered solid+single-hue color-mix gradient, five ink-referencing section vars, inverse primary button, ink secondary). Backup 2487B.
- **Gate false alarm OWNED**: counts 4/6/3 vs my stated 4/7/2 — BOTH mismatches were mis-derived expectations (heading is hex not rgba → white_rgba is 6; the gradient's accent stop is the third #0f3460, covered by g2), not drift. Booleans g1–g12 = the true coverage test, all held. Guide's needle-gate rule REFINED in 016b_7 (count expectations mechanically from the dump; mismatch = drift OR bad expectation; + the literal-% LIKE pitfall added). Lesson: derive counts by grep, never memory arithmetic.
- Fixed-template set: **site-footer, call-to-action, hero** — inert until re-render/rebuild.
- **W3c staged** (`w3c_00_before.sql`): five hero-* variant template dumps + **idea.uk per-page component inventory = the W6 gate** (hero-about 8 / hero-contact 7 renders across only 9 sites → idea.uk's about/contact very likely use them; unfixed → rebuild yields invisible white-on-parchment). Inventory defines the MINIMAL pre-rebuild fix set; the tail rides the hand-vs-fixer decision.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sw) — variants reclassified (regex gap owned); W3d one-update prepared; W6 gate list concrete
═══════════════════════════════════════════════════════════════════════
- **Reclassification owned + guide amended:** hero-* variants hardcode a legacy-palette dark HEX GRADIENT (`#1a1a2e→#16213e→#0f3460`) — 3c's `background:\s*#` regex misses gradient-embedded hexes (new pitfall bullet in 016b_7). Milder live severity (off-palette dark band, readable) but rule-1 violations; no image branch (explains 26 imageless renders).
- Five templates byte-identical bar class names → needles class-free → **W3d = ONE update for all five** (gradient→ink+layered pair; section vars→ink at 90/70/·/5/20; content colour→ink). Gate counts 1/4/1 per row, counted mechanically from the w3c dump. Files + backup cmd in runbook.
- **W6 gate from the idea.uk inventory:** fixed✓ hero/CTA/footer/header; W3d covers hero-about+hero-contact; REMAINING flagged = brief-explanation (index+tools; declarer+mixed painter) + about-content (hex bg) → `w3e_00_before.sql` dumps both + literal-scans contact-form/info-card-grid/differentiators/generic-text-block/latest-news/tool-list (incl. the gradient-hex test). differentiators = renderer surface class (layout-painted).
- **Critical-path insight:** the Go code batch (scheme-aware fallbacks, creator prompt, fixer re-aim, 003) is NOT on idea.uk's path — function lookups already resolve to fixed components — so W6 can follow W3e + repoint/force_rerender data-only; code ships as one deploy after.
- AWAITING: w3d results + w3e dumps → then W3e fixes → repoint+rerender → W6.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sx) — W3d COMPLETE; W3e specified (the real hazard found); gate closes
═══════════════════════════════════════════════════════════════════════
- **W3d COMPLETE**: UPDATE 5 one statement, f/f/f/t ×5, regions correct. Fixed set: footer, CTA, hero, 5× hero-*.
- **brief-explanation = the genuine white-on-parchment hazard** (generated; paints var(--color-background) + declares white --section-*; on idea.uk index AND tools). Plus: `rgba(var(--color-primary), 0.12)` is INVALID CSS (hex var) → the ::before glow has NEVER rendered — the color-mix fix makes it render for the first time; hardcoded violet ring → color-mix primary 25%. **Second generated half-migration exhibit → W5 creator-prompt case strengthens.**
- **about-content**: hardcoded light literals → core vars (6 needles; '#1a1a2e' ×2 both heading colours).
- **Fix rationale recorded**: page/surface painters whose internal consumers lack var() fallbacks get the AMBIENT PASS-THROUGH (--section-x: var(--color-x)) — scheme-correct both ways (core vars ARE the scheme), safer than deletion (fallback-less `border/background: var(--section-*)` would fall to currentColor/transparent). W5 prompt should mandate fallbacks so future components don't need it.
- **0.2 scan: six unflagged functions ALL clean** → the W6 gate CLOSES with W3e. Files: w3e_01_gate_and_fix (gate + 2 updates), w3e_02_verify, w3e_rollback + backup cmd.
- NEXT after w3e output: data-only chrome step — read render_site_components' join semantics (does force_rerender re-render rows whose component is inactive?), fetch the active header/footer component ids, repoint idea.uk's site_components rows, force_rerender — then W6 rebuild work items + the deployed grep.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sy) — W3e COMPLETE; W6 gate CLOSED; consolidated position written; W4b staged
═══════════════════════════════════════════════════════════════════════
- **W3e COMPLETE** (UPDATE 1 ×2; f/t and f/f/f/t; verify regions as designed; gate row absent from the paste — guards+RETURNING+verify confirm; stray kubectl-in-psql error harmless). **Ten templates fixed, seven verified clean, the W6 gate is CLOSED.**
- **Code fact settled locally** (render_site_components_action.go:345–430): pinned-component join has NO is_active filter; non-force runs SKIP non-empty slots (via logger.Debug, invisible) → **repoint BEFORE force_rerender** or the old dark chrome re-renders. Head row stays pinned (its template is variable-consuming and serviceable).
- **Runbook position REPLACED with a consolidated "WHERE WE ARE"** (done / next-in-order / open items) per the user's ask. `w4b_00_before.sql` staged: \d site_components + active chrome ids + idea.uk's rows + which agent invokes render_site_components (the operational trigger).
- AWAITING: w4b_00 output → then the repoint UPDATE + the force_rerender invocation, then W6 work items + the deployed grep.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sz) — W4b repoint prepared; trigger read staged
═══════════════════════════════════════════════════════════════════════
- w4b_00: targets header `f420f3fa-…` / footer `4238e467-…` (W2a timestamp ✓); idea.uk rows pinned to inactive components, last rendered 2026-06-21 17:15; **page-build-handler absent from the six render_site_components agents → 016's "rebuild won't refresh chrome" confirmed from data.** schema \d site_components read (unique (site_id,slot_name); locks columns present — none set on idea.uk's rows per earlier reads).
- **w4b_01_repoint.sql**: two UPDATEs guarded on the OLD ids (idempotent, single-site, single-slot); **rendered_html deliberately untouched** → no chrome-less window on the rerender path; post-check SELECT included. Rollback file restores old ids. Backup cmd dumps ids + full rendered_html first.
- **w4b_02_read_triggers.sql**: (2.1) substr-around config for each of the six agents — which passes force_rerender:true and how site_id arrives; (2.2) five most recent real rerender work items (item_type/handler/spec head) so our trigger item is crafted against the REAL spec shape. Decision next turn: force-passing agent vs targeted rendered_html=NULL + any-run.
- AWAITING: w4b_01 + w4b_02 output → craft the trigger (item insert or spawn per 016b's spawn+call entry) → verify site_components re-rendered light → W6 work items + deployed grep.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Ta) — repoint COMPLETE; trigger = rerender-pages v6; config read staged
═══════════════════════════════════════════════════════════════════════
- **Repoint COMPLETE** (UPDATE 1 ×2; header `f420f3fa-…` t, footer `4238e467-…` t, head pinned f, rendered_len unchanged — stale HTML serving until the forced re-render).
- **Trigger analysis from w4b_02**: rerender-pages v6 = the only visibly force_rerender:true render step (slots header/footer/head; input_fields site_id/domain); pageflow-builder + site-work-orchestrator use false → explains the fossilised head (full builds skip rendered chrome). Chrome step gated by `check_refresh_components` on **spec.refresh_site_components** (real items: false; ours: TRUE). Real spec shape: {reason, function, component_id, refresh_site_components}; item_key `component_regen_rerender:<uuid>`; handler rerender-pages; status flow → complete.
- **Ordering insight recorded**: W6 build renders chrome fresh, but any later automatic needs_rerender would re-inject site_components.rendered_html (items fired as recently as 07-01) → refresh stored chrome FIRST; also yields an intermediate visual checkpoint (light chrome over old sections).
- **w4b_03 staged**: full rerender-pages v6 config (how spec.function/component_id are consumed — a component_id page-filter would make a spec without one rerender nothing; must know) + complete real-item metadata (pipeline/source/severity/priority/created_by). The insert is crafted from those shapes next turn, dedup check-first per CreateNeedsNewComponentItem's pattern.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Tb) — rerender-pages v6 digested; trigger item + verify prepared
═══════════════════════════════════════════════════════════════════════
- **Workflow (v6) fully read:** gate `input_data.spec.refresh_site_components == true` (OR top-level) → FORCED chrome render (header/footer/head) → js snippets render+commit → blog listing → get_pages (deployed+active) → create_rerender_items (per-page, dispatch drains) → update_site_status deployed → complete. **spec.function/component_id consumed NOWHERE** → omitted from our spec. Timeout 300s (per-page items async beyond it).
- **Side-effects owned in the runbook:** the one item deploys the whole site as the intermediate checkpoint (light chrome + old sections) and marks it deployed. Inherent to the agent; acceptable and useful.
- **w4b_04_trigger_item.sql:** \d site_work_items first (never \d'd this thread); check-first dedup (CreateNeedsNewComponentItem pattern); INSERT mirrors real metadata (build/medium/99/rerender-pages/triaged) with TRUTHFUL provenance deviations (source manual, created_by w4b_chrome_refresh — noted in-file; real rows say component-creator/store_generated_component because they ARE regens). item_key `chrome_refresh_rerender:<site_id>`.
- **w4b_05_verify_chrome.sql:** item lifecycle; stored-chrome booleans (is_new_header/footer, footer color-mix, gradient absent, head updated_at bumped); per-page item drain by status. Then re-run the Check 4a grep on the fresh deployed index (expect site-header-section/site-footer-section; sections old until W6).
- NEXT after verification: W6 — full rebuild work items via page-build-handler for idea.uk's pages + the planned-pages decision (news-index/guides-index/tool-audience-check) + the post-rebuild grep. Then the Go code batch + the tail decision.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Tc) — chrome trigger inserted; W6 prepared (gated)
═══════════════════════════════════════════════════════════════════════
- **Trigger item inserted** (`0a30021b-…`, needs_rerender, refresh_site_components:true, triaged, prio 99). \d covered all NOT NULLs; learned: `idx_swi_dedup` also excludes 'unresolved' → W6's NOT EXISTS list matches the index exactly (w4b's list was one status short — index enforced regardless).
- **W6 shape from the real producer** (flag_page_image_rebuild_action.go): needs_page → page-build-handler; page_name-only spec ({"reason","page_name"}; handler re-derives context from the page record — the reconcile shape); item_key `page_rerender:<page>`; build/medium/99. w6_01 inserts SIX items (names FROM pages, page_id populated, truthful provenance) — **GATED until w4b_05 settles** (no interleaving rerender+rebuild on the same pages). w6_02 watches items+pages; the FINAL GREP checklist includes the legacy-var check (`var(--accent-color` count 0 = sections truly re-rendered, the fossil tell inverted) alongside header/footer/ink/color-mix/cta-pair expectations.
- **Planned-pages default stated:** news-index/guides-index/tool-audience-check stay unbuilt this pass (nav 404s pre-exist; no regression); build-or-unlink is a follow-up choice.
- AWAITING: w4b_05 settle output → run W6 → final grep → then the Go code batch (fallbacks/prompt/fixer re-aim/003 + Debug tidy) and the non-idea.uk tail decision.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Td) — chrome refresh VERIFIED; intermediate checkpoint LIVE; W6 gate open
═══════════════════════════════════════════════════════════════════════
- **W4b settled**: item complete 17:59; rows re-rendered 18:01 — header t/gradient f (3750→6258B), footer t + color-mix t (3596→7641B), head re-rendered (same template, 8009B). 5.3 drained to complete 10 (1 chrome + 9 pages — the 3 planned-status pages likely swept in via include_statuses deployed+active). Pages redeployed 18:03–18:59 → **intermediate checkpoint LIVE in B2 (light chrome + old sections)** — the first user-visible change of the whole fix.
- **0-rows rule applied to my own w6_02**: created_by filter matched nothing because w6_01 hadn't run; pages' 18:xx updates were the rerender's work. Key-family: rerender per-page items share `page_rerender:<name>` but are all complete → dedup index (non-terminal only) won't block the W6 inserts.
- **Gate OPEN**: next = w6_01 (INSERT 0 6) → w6_02 drain → final grep (fossil `var(--accent-color`=0; header/footer classes; --hero-ink; color-mix; cta pair; white rgba gone).
- Stray psql paste error again (prompt text into psql) — harmless, noted.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Te) — W6 rebuild COMPLETE; final verification staged
═══════════════════════════════════════════════════════════════════════
- **W6 drained**: INSERT 0 6 at 19:11; all complete; pages rebuilt+redeployed 19:18–22:45 (full build path heavier than rerender — plan_sections + render + deploy per page). **Three retried before completing** (about ×1, privacy ×2, report ×2 — one retry from failed; error text retained) → w6_03 §3.1 reads the failure class before calling retries healthy. Hygiene observation: site_work_items.updated_at frozen at insert through claim/retry/complete (layouts-pre-W2b family) — listed, not actioned.
- **w6_03_final_verify.sql staged**: §3.1 error heads; §3.2 rebuilt stored sections on index+about (hero_ink t, cta_pair t, legacy_fossil f, old_white f; note brief-explanation's pass-through uses plain core vars — its check is old_white=f, not color-mix). Then the deployed grep block (8 lines incl. the fossil check).
- On a clean grep the ORIGINAL SYMPTOM CLOSES. Remaining: Go code batch (one deploy), tail decision (hand vs re-aimed fixer for ~10 declarers + ~17 band-class), planned-pages choice, hygiene list (site_work_items.updated_at; the 'active'/'system' legacy statuses; what ran 07-01 — largely superseded by today's activity but still unexplained).
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Tf) — DB-side verify: templates IN; brief-explanation dropped from index (gating)
═══════════════════════════════════════════════════════════════════════
- **3.1**: all three retry errors = `Claim timed out — handler pod likely died` → dispatch-infra class; retries recovered; hygiene note (claim duration vs heavy builds; pod health 19:11–22:45).
- **3.2 PASS**: index/hero ink+mix t; index/CTA pair t; about/hero-about ink t; about-content all-f per its core-var fix; **legacy_fossil f + old_white f on EVERY row** — stored sections truly re-rendered from the fixed templates.
- **DISCREPANCY (uncocluded, gating the close)**: index rebuilt with FIVE components — **brief-explanation gone** (was six; tools also had it). Candidates: never in pages.sections (legacy) / plan_sections resolution failure (it has section_data history) / 016 drop-on-rebuild hazard. `w6_04` reads pages.sections (index+tools), the rebuilt tools+contact sets, and any items created in the rebuild window by other creators. Deployed grep gains `data-component="brief-explanation"` count. No conclusions until both are back.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Tg) — drop mechanism identified: needs_section_data escalation (illustration_url)
═══════════════════════════════════════════════════════════════════════
- 4.1 sections still list brief-explanation (index+tools) → not legacy. 4.2 tools dropped it too; contact intact. 4.3 **two needs_section_data items in needs_human_review** ("field 'illustration_url' from site_a…", tools 19:13 / index 21:34) → plan_sections escalates AND omits; the page built without the section. **Guide refinement queued: the rebuild drop is loud, not silent** — the fossil pages were hiding the unresolved dependency.
- Context: pre-thread deployed index ALSO lacked the section (Check 4 HTML) → no new user-visible loss vs weeks of state; the 07-01 `section_data_resolved` items prove it resolved once — why-not-now is the w6_05 read (full spec + suggested_action + resolution_path + input_schema source for illustration_url).
- Response options staged: imagery-supply (needs_imagery, section-scope) vs the STRUCTURAL fix (illustration_url optional in input_schema + `{{if .illustration_url}}` gate around the image column → section renders imageless; imagery arrives later, a rerender adds it). Leaning structural per standing preference, decide on evidence.
- **Deployed grep still pending = the scheme-close check.** brief-explanation count 0 expected+explained; the other seven lines decide.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Th) — imagery decision received; the pipeline already exists; w7_00 staged
═══════════════════════════════════════════════════════════════════════
- **User decision:** imagery must not block; a flow creates it dynamically from spec descriptions.
- **Reuse finding:** the flow EXISTS end-to-end — write_site_plan → site_plan_imagery (descriptions, section scope, Phase 2G) → emit_imagery_items (deployed 05-26, build-site-planner) → needs_imagery ≤98 → image-build-handler (generate/asset/deploy) → flag_page_image_rebuild → needs_page 99 → rebuild resolves via plan_sections' spi.key→assets.url join. Why it never fired for idea.uk = the only open question (no 2G row in an old plan vs row-without-emission) → w7_00 0.3/0.4.
- **w7_00 staged** (replaces the unreadable attachments): 0.1 input_schema + illustration region (gate needles); 0.2 escalations' suggested_action/resolution_path; 0.3 \d site_plan_imagery + idea.uk's rows; 0.4 idea.uk needs_imagery history + three real items for the spec shape.
- **Plan:** (A) optional+gate fix → rebuild index+tools imageless → close the two human-review items; (B) feed the existing pipeline (plan row or shaped item); (C) W5 batch: creator prompt + 003 make image fields optional-with-gate by default.
- **Attachments arrived unreadable** (stated plainly to the user); **scheme-close grep still pending** — nine numbers or index.html via file upload. Hygiene: full-build deploys commit as "Rerender: …" — message format no longer distinguishes paths.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Ti) — the section-imagery GAP found in code; W7 sequenced a/b/c
═══════════════════════════════════════════════════════════════════════
- **GAP (structural):** ensureAssets loads ONLY hero (page) + logo (site) + content_data fallback — NO section-scope lookup → section illustrations can NEVER resolve via site_assets today. The pipeline (spi → emit → needs_imagery → image-build-handler → asset → flag_page_image_rebuild → rebuild) is complete for hero/logo; the resolver's last inch was never wired for sections. content_data stuffing would work but is the patch class we avoid → the fix is a third ensureAssets query (section scope, scope_ref LIKE page||':%', r.assets[spi.key]=url), riding the Go batch.
- **on_missing: skip_field is the established optional pattern** (the phantom-link fix in the pages case) — W7a applies an existing mechanism.
- Sequenced: W7a data (schema on_missing + template gate → rebuild imageless → close the two items) → W7b data (spi rows: section scope_refs `index:brief-explanation`/`tools:brief-explanation`, kind=illustration, source=manual, prompt from site_plan_sections description + shaped needs_imagery item → asset lands NOW) → W7c Go (ensureAssets extension; key convention from 0.1's source path). Image appears after the Go deploy via the pipeline's own queued rebuild.
- **w7_00b staged (paste-as-text)**: 0.1 schema+markup, 0.2 advice, 0.3 existing spi rows, 0.4 item shapes, 0.5 \d site_plan_sections + the section description. **Scheme-close grep numbers still owed as text.** Attachments unreadable ×2 — pasted text is the channel.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Tj) — imagery half settled bar three reads; two assumptions corrected
═══════════════════════════════════════════════════════════════════════
- 0.3: 16 spi rows (5 heroes, 10 icons, logo) — **NO illustration/brief-explanation rows** → planner never requested it; emitter is gap-driven so nothing emitted. All 10 idea.uk needs_imagery completed 06-21 (deployed heroes explained).
- **Corrections owned:** scope_ref = page:ORDINAL (not page:name — my w7 plan text updated); 0.5 used section_name (column is component_name; NO description column → prompts are AUTHORED). Prompts drafted in the house voice (runbook, user to approve); final key awaits 0.1's declared site_assets.<KEY> path.
- 0.4: section ILLUSTRATIONS proven in the wild (illustration_game_master, complete) → generation works; the ensureAssets gap is the sole missing inch AND applies to the landed icon sets too (W7c benefits icons). Spec shape visible ({key,kind,scope,prompt,check:emit_imagery_items,…}); 0.6 fetches one FULL spec to copy.
- W7b final form: two spi inserts (page:ordinal scope_refs from corrected 0.5) + two hand-shaped items (spec per 0.6). No planner rerun.
- STILL OWED as text: 0.1 (schema key + gate needles), 0.2 (advice), corrected 0.5 (ordinals), 0.6 (full spec), **+ the nine grep numbers (scheme close)**.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Tk) — SCHEME CLOSED on deployed evidence
═══════════════════════════════════════════════════════════════════════
- **All nine grep checks PASS** (32/0/37/13/14/pair-consumed/0/0/0-expected). idea.uk: resolved light, renders light — header, footer, hero, CTA, sections, at initial render on the build path; fixers backstop only. **The thread's P0 is closed.** PLAN carries a CLOSED banner pointing at SPEC + runbook; runbook §SCHEME CLOSE holds the remaining-work map.
- Provenance nicety: the bare `var(--color-cta-bg)` = the legacy pinned head's `.section--cta` rule (class unused) → hygiene list (modernise head).
- Attachments unreadable ×3 (doc 30) — w7_00c OUTPUT still owed as pasted text (the message carried the queries, not the rows). W7a/W7b files write on receipt; prompts await approval; W7c rides the Go batch.
idea.uk live VM untouched throughout.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Tl) — schema was already non-blocking; the defer fork; W7a gate + W7b imagery prepared
═══════════════════════════════════════════════════════════════════════
- **Owned ×2:** (1) illustration_url was ALREADY required:false + skip_field — the planned W7a schema change was void; (2) my "OnMissing never consulted" was a grep-pattern error (local `onMissing` var). On-disk optional/skip_field branch is correct.
- **The defer fork:** deployed run dropped the section anyway → extraction misparse (unread block just above) OR deployed≠disk. Checks queued (extraction read; git log vs pod image age). Latent smell found: required-branch lacks skip_field → default-defer; case added to the batch list. **Quick patch (source→static) rejected** — revert debt, no user-visible need; section returns with the Go batch.
- **W7a prepared:** template gate ({{if .illustration_url}} around image-wrapper; needles from the pasted region; covers the src="" case whenever the section renders field-skipped). **W7b prepared:** two spi rows (index:1/tools:1; keys illustration_home/illustration_tools per house style — W7c maps kind→path like hero) + two items copying the 0.6 shape (check/source manual, priority 98, image-build-handler); prompts in-file, editable. Dedup mirrors the unique tuple + idx_swi_dedup.
- ORDINAL fragility noted for the record: existing icon rows' scope_ref index:4 vs current plan (latest-news at 4, info-card-grid at 5) — scope_ref-by-ordinal drifts when plans reorder; resolution is by key so cosmetic today; hygiene list.
- Go batch list consolidated (runbook §W7 FINDINGS). idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Tm) — W7a gate IN; W7b rows+items IN; fork resolved to deployed-drift
═══════════════════════════════════════════════════════════════════════
- W7a: UPDATE 1, gated t/t. W7b: 2 spi rows + 2 items (triaged 13:23); assets 0 = not-yet (0-rows rule: query fine, timing) → watch w7b_02 until complete + assets land; then expect needs_page for index+tools from the flag step, which will RE-DEFER brief-explanation under the current binary (expected; items dedupe).
- **Fork RESOLVED**: extraction sound (required parses false; on_missing defaults skip_field) → on-disk code cannot produce the July-2 escalation → deployed predates disk; **the skip_field fix exists and never shipped**. Batch item = deploy existing plan_sections + the required-branch skip_field case. Confirmatory: git log on the file vs pod image age.
- Remaining decisions: batch venue (here vs Claude Code); then the tail, planned pages, hygiene. idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Tn) — illustrations LANDED; batch venue = here; slice 1 spec'd
═══════════════════════════════════════════════════════════════════════
- W7b complete end-to-end: items complete 13:26/13:28; assets active at B2 13:25/13:28 (~3 min generation each). Flag-triggered rebuilds re-defer under current binary (expected); escalations will SELF-CLOSE post-deploy (closeResolvedDataRequest:1302).
- **Batch venue: HERE.** Slice plan: 1 plan_sections (spec'd — `gobatch_01_plan_sections.md`: Edit A required-branch skip_field case w/ logger-identifier check; Edit B ensureAssets section query mapping by key + kind-alias, hero-modelled) → 2 component_library fallbacks + Debug tidy → 3 fixer re-aim → 4 creator prompt + 003 (locate prompt: agent_definitions). `w8_01_post_deploy_rebuild.sql` ready (gated on the deploy): two needs_page + verification (section_back/has_image t; escalations complete).
- idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (To) — skip_field live post-deploy; image resolution pending w8 drain
═══════════════════════════════════════════════════════════════════════
- w8 items inserted (triaged); verify ran pre-drain → the t/f snapshot predates them. **Section BACK + both escalations self-closed = the deployed skip_field behaviour works** (some post-deploy rebuild produced it). has_image=f NOT concluded against Edit B — three candidates (pre-Edit-B render / query mismatch / code path), separated by: 2.1 the hand-run Edit-B query (data-side decisive), 2.2 render provenance, 2.3 post-drain recheck; then pod logs + image tag if needed.
- idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Tp) — data side proven (4 rows); two-deploys story; flag gap found
═══════════════════════════════════════════════════════════════════════
- 2.1 = 4 rows (illustration + 3 icons, active) → Edit B query/data sound; icons ride free once Edit B is live. 2.3 renders 14:04/15:42 = the FIRST w8 pair (post-deploy-1): section back (skip_field live) sans image.
- Timeline: deploy-1 (≤13:46, working-copy skip_field, Edit B uncertain) → w8 pair-1 complete → user's Debug→Info change (NOT in component_library — upload still 8 Debugs; likely plan_sections) + deploy-2 ~16:05–16:10 → claim released (claimed→triaged, W6 class) → w8 pair-2 (16:08) pending = THE TEST. has_image t → close brief-explanation + spec slice 2; f → verify running image contains Edit B (success path silent).
- **Flag gap banked:** zero flag-created needs_page in 30h → flag_page_image_rebuild misses section-scope landings (page-scope-only or scope_ref parse) → batch addition (parse page prefix from page:ordinal + emit).
- input_mapping.go = workflow input mapper (no site_assets) → second-path hypothesis ruled out.
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Tq) — Edit B LIVE (tools t); index-specific miss; slice 2 delivered
═══════════════════════════════════════════════════════════════════════
- **results.txt via file upload READ SUCCESSFULLY** — the reliable channel confirmed (inline attachments failed ×4).
- **2.3: tools t / index f, same wave, same binary** → Edit B works; the miss is index-specific. Leading hypothesis: applied-loop variance (first-row-only/early-exit) — index's rows sort icons first, tools' single row IS the illustration → the split fingerprints the variant exactly. Witness = the APPLIED block (paste or upload plan_sections_action.go). `w8_03_fingerprint.sql` staged: info-card-grid icon booleans (all/k1-only/none → three diagnoses) + the identical-microsecond 17:25 anomaly scope (two rows = component-targeted single UPDATE, fixer-like; all slots = site-wide pass) — identify the 17:25 toucher before trusting content diffs.
- **Slice 2 delivered** (gobatch_02): Edits C/D whole-function fallback header/footer → chrome pairs (+ compile note on removed locals; %%-escaping note); RenderFallbackHead unchanged (theme-color meta ≠ CSS var — ctx.PrimaryColor legitimate); Edit E = 8 Debug→Info lines. Independent of the index diagnosis; can share a deploy.
- Next: user runs w8_03 + pastes/uploads the applied Edit B (+ optional pod-log grep) → correct the loop if variant confirmed → rebuild index → icons + illustration land together. Then slices 3 (fixer re-aim) and 4 (creator prompt + 003).
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Tr) — 17:25 anomaly superseded; fingerprint assumption owned; current-state read staged
═══════════════════════════════════════════════════════════════════════
- 3.2: index rebuilt 17:30:07, tools 18:12:55 (sequential-ms build saves) — AFTER results.txt → the 17:25 identical-microsecond rows are overwritten; anomaly CLOSED as moot. New opens: provenance of the 17:30/18:12 builds (16:08 pair was already complete at results-capture — contradiction stated, not forced into a story) + has_image on the CURRENT renders (unchecked in the paste). User's "hasn't run yet" → 4.2 lists pending items.
- **3.1 assumption owned**: icon fingerprint presumed site_assets.icon_* sourcing for info-card-grid — schema unread → all-f non-decisive; 4.3 reads the real sources.
- Witness still owed: the APPLIED Edit-B block (paste/upload). w8_04 staged (4.1 has_image now; 4.2 provenance+pending; 4.3 icon sources).
idea.uk live VM untouched.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Ts) — split current (tools t / index f); two instruments corrected; witness or experiment
═══════════════════════════════════════════════════════════════════════
- 4.1: split persists on the 17:30/18:12 renders → Edit B live; index-specific. Hero resolves on index ⇒ pageName correct ⇒ applied-variant hypothesis is the survivor.
- Owned: 4.2 filter too narrow (needs_page only) → widened (w8_05); 4.3 discards the icon fingerprint (icon_svg LLM-sourced; no site_assets icon fields) — 3.1 was built on an unread-schema assumption.
- Third ask for the APPLIED Edit-B block (paste/upload). Fallback staged: w8_06 OPTIONAL reversible experiment (icon rows' scope_ref prefix toggle → index join = illustration only → rebuild → t ⇒ first-row-only confirmed; restore mandatory). Provenance oddity (17:25 vs 17:30/18:12 vs complete-at-capture) parked behind w8_05's wide read — moot for the split.
idea.uk live VM untouched.
